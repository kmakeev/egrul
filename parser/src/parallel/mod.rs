//! Модуль параллельной обработки файлов
//!
//! Обеспечивает:
//! - Параллельный парсинг нескольких XML файлов
//! - Распределение файлов по воркерам
//! - Объединение результатов
//! - Прогресс-бар и статистика

mod worker;

pub use worker::*;

use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicUsize, AtomicU64, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use crossbeam_channel::{bounded, Sender, Receiver};
use indicatif::{MultiProgress, ProgressBar, ProgressStyle};
use rayon::prelude::*;
use tracing::{info, warn, error};
use walkdir::WalkDir;

use crate::error::{Error, Result};
use crate::models::{EgrulRecord, EgripRecord, RegistryRecord};
use crate::output::{OutputFormat, OutputWriter};
use crate::parser::{Parser, ParserConfig};

/// Параллельный обработчик файлов
pub struct ParallelProcessor {
    config: ParserConfig,
    stats: Arc<ProcessingStats>,
}

/// Статистика обработки
#[derive(Debug, Default)]
pub struct ProcessingStats {
    /// Обработано файлов
    pub files_processed: AtomicUsize,
    /// Файлов с ошибками
    pub files_failed: AtomicUsize,
    /// Записей ЕГРЮЛ
    pub egrul_records: AtomicUsize,
    /// Записей ЕГРИП
    pub egrip_records: AtomicUsize,
    /// Обработано байт
    pub bytes_processed: AtomicU64,
    /// Ошибок парсинга
    pub parse_errors: AtomicUsize,
    /// Время начала
    start_time: std::sync::RwLock<Option<Instant>>,
}

impl ProcessingStats {
    pub fn new() -> Self {
        Self::default()
    }
    
    pub fn start(&self) {
        let mut start = self.start_time.write().unwrap();
        *start = Some(Instant::now());
    }
    
    pub fn elapsed(&self) -> Duration {
        self.start_time.read().unwrap()
            .map(|t| t.elapsed())
            .unwrap_or_default()
    }
    
    pub fn total_records(&self) -> usize {
        self.egrul_records.load(Ordering::Relaxed) + 
        self.egrip_records.load(Ordering::Relaxed)
    }
    
    pub fn records_per_second(&self) -> f64 {
        let elapsed = self.elapsed().as_secs_f64();
        if elapsed > 0.0 {
            self.total_records() as f64 / elapsed
        } else {
            0.0
        }
    }
    
    pub fn print_summary(&self) {
        info!("═══════════════════════════════════════════════════════════");
        info!("                  ИТОГОВАЯ СТАТИСТИКА                       ");
        info!("═══════════════════════════════════════════════════════════");
        info!("");
        info!("  Файлов обработано:  {}", self.files_processed.load(Ordering::Relaxed));
        info!("  Файлов с ошибками:  {}", self.files_failed.load(Ordering::Relaxed));
        info!("");
        info!("  Записей ЕГРЮЛ:      {}", self.egrul_records.load(Ordering::Relaxed));
        info!("  Записей ЕГРИП:      {}", self.egrip_records.load(Ordering::Relaxed));
        info!("  Всего записей:      {}", self.total_records());
        info!("");
        info!("  Ошибок парсинга:    {}", self.parse_errors.load(Ordering::Relaxed));
        info!("");
        
        let elapsed = self.elapsed();
        info!("  Время:              {:.2} сек", elapsed.as_secs_f64());
        
        let rate = self.records_per_second();
        if rate > 0.0 {
            info!("  Скорость:           {:.0} записей/сек ({:.0} записей/мин)", 
                rate, rate * 60.0);
        }
        
        info!("═══════════════════════════════════════════════════════════");
    }
}

/// Информация о файле для обработки
#[derive(Debug, Clone)]
pub struct FileInfo {
    pub path: PathBuf,
    pub size: u64,
}

impl ParallelProcessor {
    /// Создание нового процессора
    pub fn new(config: ParserConfig) -> Self {
        // Настройка rayon
        rayon::ThreadPoolBuilder::new()
            .num_threads(config.num_threads)
            .build_global()
            .ok();
        
        Self {
            config,
            stats: Arc::new(ProcessingStats::new()),
        }
    }
    
    /// Получение статистики
    pub fn stats(&self) -> &ProcessingStats {
        &self.stats
    }
    
    /// Сброс статистики
    pub fn reset_stats(&self) {
        self.stats.files_processed.store(0, Ordering::Relaxed);
        self.stats.files_failed.store(0, Ordering::Relaxed);
        self.stats.egrul_records.store(0, Ordering::Relaxed);
        self.stats.egrip_records.store(0, Ordering::Relaxed);
        self.stats.bytes_processed.store(0, Ordering::Relaxed);
        self.stats.parse_errors.store(0, Ordering::Relaxed);
    }
    
    /// Сбор списка файлов для обработки
    pub fn collect_files(&self, input_dir: &Path) -> Result<Vec<FileInfo>> {
        let files: Vec<FileInfo> = WalkDir::new(input_dir)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| {
                e.path()
                    .extension()
                    .map(|ext| ext.eq_ignore_ascii_case("xml"))
                    .unwrap_or(false)
            })
            .filter_map(|e| {
                let path = e.path().to_path_buf();
                std::fs::metadata(&path).ok().map(|m| FileInfo {
                    path,
                    size: m.len(),
                })
            })
            .collect();
        
        Ok(files)
    }
    
    /// Распределение файлов по воркерам (по размеру)
    pub fn distribute_files(&self, mut files: Vec<FileInfo>, num_workers: usize) -> Vec<Vec<FileInfo>> {
        // Сортируем по размеру (большие первыми)
        files.sort_by(|a, b| b.size.cmp(&a.size));
        
        // Распределяем по воркерам методом "наименьшей загрузки"
        let mut worker_loads: Vec<(usize, u64, Vec<FileInfo>)> = (0..num_workers)
            .map(|i| (i, 0u64, Vec::new()))
            .collect();
        
        for file in files {
            // Находим наименее загруженного воркера
            worker_loads.sort_by_key(|w| w.1);
            worker_loads[0].1 += file.size;
            worker_loads[0].2.push(file);
        }
        
        worker_loads.into_iter().map(|w| w.2).collect()
    }
    
    /// Обработка директории
    pub fn process_directory(
        &self,
        input_dir: &Path,
        output_dir: &Path,
        format: OutputFormat,
    ) -> Result<()> {
        info!("Начало параллельной обработки: {:?}", input_dir);
        
        // Собираем файлы
        let files = self.collect_files(input_dir)?;
        let total_files = files.len();
        let total_size: u64 = files.iter().map(|f| f.size).sum();
        
        info!("Найдено {} XML файлов ({} МБ)", total_files, total_size / 1024 / 1024);
        
        if files.is_empty() {
            warn!("XML файлы не найдены в {:?}", input_dir);
            return Ok(());
        }
        
        // Создаем выходную директорию
        std::fs::create_dir_all(output_dir)?;
        
        // Запускаем таймер
        self.stats.start();
        
        // Настраиваем прогресс-бар
        let multi = MultiProgress::new();
        
        let main_progress = if self.config.show_progress {
            let pb = multi.add(ProgressBar::new(total_files as u64));
            pb.set_style(
                ProgressStyle::default_bar()
                    .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {pos}/{len} ({percent}%) {msg}")
                    .unwrap()
                    .progress_chars("█▓░"),
            );
            pb.set_message("Обработка файлов...");
            Some(pb)
        } else {
            None
        };
        
        // Создаем каналы для передачи данных
        let (egrul_tx, egrul_rx): (Sender<EgrulRecord>, Receiver<EgrulRecord>) =
            bounded(self.config.channel_buffer_size);
        let (egrip_tx, egrip_rx): (Sender<EgripRecord>, Receiver<EgripRecord>) =
            bounded(self.config.channel_buffer_size);
        
        let stats = Arc::clone(&self.stats);
        let config = self.config.clone();
        let output_dir_clone = output_dir.to_path_buf();
        
        // Поток записи ЕГРЮЛ
        let egrul_writer_handle = std::thread::spawn({
            let output_dir = output_dir_clone.clone();
            let batch_size = config.batch_size;
            let max_file_size_mb = config.max_file_size_mb;
            let max_records_per_file = config.max_records_per_file;
            move || -> Result<()> {
                let output_path = output_dir.join(format!("egrul.{}", format.extension()));
                let mut writer = OutputWriter::with_limits(
                    &output_path,
                    format,
                    max_file_size_mb,
                    max_records_per_file,
                )?;
                let mut batch = Vec::with_capacity(batch_size);
                
                for record in egrul_rx {
                    batch.push(record);
                    if batch.len() >= batch_size {
                        writer.write_egrul_batch(&batch)?;
                        batch.clear();
                    }
                }
                
                if !batch.is_empty() {
                    writer.write_egrul_batch(&batch)?;
                }
                
                writer.finish()?;
                Ok(())
            }
        });
        
        // Поток записи ЕГРИП
        let egrip_writer_handle = std::thread::spawn({
            let output_dir = output_dir_clone;
            let batch_size = config.batch_size;
            let max_file_size_mb = config.max_file_size_mb;
            let max_records_per_file = config.max_records_per_file;
            move || -> Result<()> {
                let output_path = output_dir.join(format!("egrip.{}", format.extension()));
                let mut writer = OutputWriter::with_limits(
                    &output_path,
                    format,
                    max_file_size_mb,
                    max_records_per_file,
                )?;
                let mut batch = Vec::with_capacity(batch_size);
                
                for record in egrip_rx {
                    batch.push(record);
                    if batch.len() >= batch_size {
                        writer.write_egrip_batch(&batch)?;
                        batch.clear();
                    }
                }
                
                if !batch.is_empty() {
                    writer.write_egrip_batch(&batch)?;
                }
                
                writer.finish()?;
                Ok(())
            }
        });
        
        // Параллельная обработка файлов
        let parser = Parser::with_config(config.clone());
        
        files.par_iter().for_each(|file_info| {
            match parser.parse_file(&file_info.path) {
                Ok(records) => {
                    for record in records {
                        match record {
                            RegistryRecord::Egrul(r) => {
                                stats.egrul_records.fetch_add(1, Ordering::Relaxed);
                                let _ = egrul_tx.send(r);
                            }
                            RegistryRecord::Egrip(r) => {
                                stats.egrip_records.fetch_add(1, Ordering::Relaxed);
                                let _ = egrip_tx.send(r);
                            }
                        }
                    }
                    stats.files_processed.fetch_add(1, Ordering::Relaxed);
                    stats.bytes_processed.fetch_add(file_info.size, Ordering::Relaxed);
                }
                Err(e) => {
                    stats.files_failed.fetch_add(1, Ordering::Relaxed);
                    stats.parse_errors.fetch_add(1, Ordering::Relaxed);
                    
                    if config.continue_on_error {
                        warn!("Пропуск файла {:?}: {}", file_info.path, e);
                    } else {
                        error!("Ошибка парсинга {:?}: {}", file_info.path, e);
                    }
                }
            }
            
            if let Some(ref pb) = main_progress {
                pb.inc(1);
                let processed = stats.files_processed.load(Ordering::Relaxed);
                let egrul = stats.egrul_records.load(Ordering::Relaxed);
                let egrip = stats.egrip_records.load(Ordering::Relaxed);
                pb.set_message(format!(
                    "{}/{} | ЕГРЮЛ: {} | ЕГРИП: {}",
                    processed, total_files, egrul, egrip
                ));
            }
        });
        
        // Закрываем каналы
        drop(egrul_tx);
        drop(egrip_tx);
        
        // Ждем завершения записи
        egrul_writer_handle.join()
            .map_err(|_| Error::other("Ошибка потока записи ЕГРЮЛ"))??;
        egrip_writer_handle.join()
            .map_err(|_| Error::other("Ошибка потока записи ЕГРИП"))??;
        
        if let Some(pb) = main_progress {
            pb.finish_with_message("Завершено");
        }
        
        // Выводим статистику
        self.stats.print_summary();
        
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_distribute_files() {
        let config = ParserConfig::default();
        let processor = ParallelProcessor::new(config);
        
        let files = vec![
            FileInfo { path: PathBuf::from("a.xml"), size: 100 },
            FileInfo { path: PathBuf::from("b.xml"), size: 200 },
            FileInfo { path: PathBuf::from("c.xml"), size: 50 },
            FileInfo { path: PathBuf::from("d.xml"), size: 150 },
        ];
        
        let distributed = processor.distribute_files(files, 2);
        assert_eq!(distributed.len(), 2);
        
        // Проверяем что файлы распределены
        let total: usize = distributed.iter().map(|w| w.len()).sum();
        assert_eq!(total, 4);
    }
}

