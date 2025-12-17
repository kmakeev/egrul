//! Воркер для параллельной обработки файлов

use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread::JoinHandle;

use crossbeam_channel::{Sender, Receiver, bounded, TrySendError};
use tracing::{debug, warn, error};

use crate::error::{Error, Result};
use crate::models::{EgrulRecord, EgripRecord, RegistryRecord};
use crate::parser::Parser;

/// Статус воркера
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum WorkerStatus {
    /// Ожидание
    Idle,
    /// Обработка
    Processing,
    /// Завершен
    Finished,
    /// Ошибка
    Error,
}

/// Статистика воркера
#[derive(Debug, Default)]
pub struct WorkerStats {
    pub files_processed: AtomicUsize,
    pub files_failed: AtomicUsize,
    pub egrul_records: AtomicUsize,
    pub egrip_records: AtomicUsize,
    pub current_file: std::sync::RwLock<Option<PathBuf>>,
}

impl WorkerStats {
    pub fn new() -> Self {
        Self::default()
    }
    
    pub fn total_records(&self) -> usize {
        self.egrul_records.load(Ordering::Relaxed) + 
        self.egrip_records.load(Ordering::Relaxed)
    }
    
    pub fn set_current_file(&self, path: Option<PathBuf>) {
        if let Ok(mut current) = self.current_file.write() {
            *current = path;
        }
    }
    
    pub fn get_current_file(&self) -> Option<PathBuf> {
        self.current_file.read().ok().and_then(|f| f.clone())
    }
}

/// Задание для воркера
#[derive(Debug)]
pub struct WorkerTask {
    pub file_path: PathBuf,
    pub file_size: u64,
}

/// Воркер для обработки файлов
pub struct Worker {
    id: usize,
    parser: Parser,
    stats: Arc<WorkerStats>,
    running: Arc<AtomicBool>,
}

impl Worker {
    /// Создание нового воркера
    pub fn new(id: usize, parser: Parser) -> Self {
        Self {
            id,
            parser,
            stats: Arc::new(WorkerStats::new()),
            running: Arc::new(AtomicBool::new(false)),
        }
    }
    
    /// ID воркера
    pub fn id(&self) -> usize {
        self.id
    }
    
    /// Статистика воркера
    pub fn stats(&self) -> Arc<WorkerStats> {
        Arc::clone(&self.stats)
    }
    
    /// Проверка, запущен ли воркер
    pub fn is_running(&self) -> bool {
        self.running.load(Ordering::Relaxed)
    }
    
    /// Обработка одного файла
    pub fn process_file(&self, task: &WorkerTask) -> Result<Vec<RegistryRecord>> {
        debug!("Worker {} обрабатывает {:?}", self.id, task.file_path);
        self.stats.set_current_file(Some(task.file_path.clone()));
        
        let result = self.parser.parse_file(&task.file_path);
        
        self.stats.set_current_file(None);
        
        match &result {
            Ok(records) => {
                let mut egrul = 0;
                let mut egrip = 0;
                for record in records {
                    match record {
                        RegistryRecord::Egrul(_) => egrul += 1,
                        RegistryRecord::Egrip(_) => egrip += 1,
                    }
                }
                self.stats.egrul_records.fetch_add(egrul, Ordering::Relaxed);
                self.stats.egrip_records.fetch_add(egrip, Ordering::Relaxed);
                self.stats.files_processed.fetch_add(1, Ordering::Relaxed);
            }
            Err(_) => {
                self.stats.files_failed.fetch_add(1, Ordering::Relaxed);
            }
        }
        
        result
    }
    
    /// Запуск воркера с получением задач из канала
    pub fn run_with_channel(
        self,
        task_rx: Receiver<WorkerTask>,
        egrul_tx: Sender<EgrulRecord>,
        egrip_tx: Sender<EgripRecord>,
        continue_on_error: bool,
    ) -> JoinHandle<Result<()>> {
        let running = Arc::clone(&self.running);
        
        std::thread::spawn(move || {
            running.store(true, Ordering::Relaxed);
            
            for task in task_rx {
                match self.process_file(&task) {
                    Ok(records) => {
                        for record in records {
                            match record {
                                RegistryRecord::Egrul(r) => {
                                    if let Err(e) = egrul_tx.send(r) {
                                        warn!("Worker {} не смог отправить ЕГРЮЛ: {}", self.id, e);
                                    }
                                }
                                RegistryRecord::Egrip(r) => {
                                    if let Err(e) = egrip_tx.send(r) {
                                        warn!("Worker {} не смог отправить ЕГРИП: {}", self.id, e);
                                    }
                                }
                            }
                        }
                    }
                    Err(e) => {
                        if continue_on_error {
                            warn!("Worker {} пропустил {:?}: {}", self.id, task.file_path, e);
                        } else {
                            error!("Worker {} ошибка {:?}: {}", self.id, task.file_path, e);
                            running.store(false, Ordering::Relaxed);
                            return Err(e);
                        }
                    }
                }
            }
            
            running.store(false, Ordering::Relaxed);
            Ok(())
        })
    }
}

/// Пул воркеров
pub struct WorkerPool {
    workers: Vec<Worker>,
    task_tx: Option<Sender<WorkerTask>>,
    handles: Vec<JoinHandle<Result<()>>>,
}

impl WorkerPool {
    /// Создание пула воркеров
    pub fn new(
        num_workers: usize,
        queue_size: usize,
        parser: Parser,
        egrul_tx: Sender<EgrulRecord>,
        egrip_tx: Sender<EgripRecord>,
        continue_on_error: bool,
    ) -> Self {
        let (task_tx, task_rx) = bounded(queue_size);
        
        let mut handles = Vec::with_capacity(num_workers);
        let mut workers = Vec::with_capacity(num_workers);
        
        for id in 0..num_workers {
            let worker = Worker::new(id, parser.clone());
            workers.push(Worker::new(id, parser.clone())); // Для статистики
            
            let handle = worker.run_with_channel(
                task_rx.clone(),
                egrul_tx.clone(),
                egrip_tx.clone(),
                continue_on_error,
            );
            handles.push(handle);
        }
        
        Self {
            workers,
            task_tx: Some(task_tx),
            handles,
        }
    }
    
    /// Отправка задания в пул
    pub fn submit(&self, task: WorkerTask) -> Result<()> {
        if let Some(ref tx) = self.task_tx {
            tx.send(task).map_err(|e| Error::other(format!("Канал закрыт: {}", e)))
        } else {
            Err(Error::other("Пул закрыт"))
        }
    }
    
    /// Попытка отправки без блокировки
    pub fn try_submit(&self, task: WorkerTask) -> Result<bool> {
        if let Some(ref tx) = self.task_tx {
            match tx.try_send(task) {
                Ok(()) => Ok(true),
                Err(TrySendError::Full(_)) => Ok(false),
                Err(TrySendError::Disconnected(_)) => Err(Error::other("Канал закрыт")),
            }
        } else {
            Err(Error::other("Пул закрыт"))
        }
    }
    
    /// Закрытие канала задач и ожидание завершения
    pub fn shutdown(mut self) -> Result<()> {
        // Закрываем канал задач
        self.task_tx.take();
        
        // Ждем завершения всех воркеров
        for handle in self.handles {
            handle.join()
                .map_err(|_| Error::other("Ошибка потока воркера"))??;
        }
        
        Ok(())
    }
    
    /// Получение статистики воркеров
    pub fn stats(&self) -> Vec<(usize, Arc<WorkerStats>)> {
        self.workers.iter()
            .map(|w| (w.id(), w.stats()))
            .collect()
    }
}

impl Clone for Parser {
    fn clone(&self) -> Self {
        Parser::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_worker_stats() {
        let stats = WorkerStats::new();
        
        stats.files_processed.fetch_add(1, Ordering::Relaxed);
        stats.egrul_records.fetch_add(10, Ordering::Relaxed);
        stats.egrip_records.fetch_add(5, Ordering::Relaxed);
        
        assert_eq!(stats.files_processed.load(Ordering::Relaxed), 1);
        assert_eq!(stats.total_records(), 15);
    }
}

