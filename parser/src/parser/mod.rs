//! Модули парсинга XML

mod xml_reader;
mod egrul_parser;
mod egrip_parser;
mod attributes;

pub use xml_reader::*;
pub use egrul_parser::*;
pub use egrip_parser::*;
pub use attributes::*;

use std::path::Path;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;

use crossbeam_channel::{bounded, Sender, Receiver};
use rayon::prelude::*;
use tracing::{info, warn, error, debug};
use walkdir::WalkDir;
use indicatif::{ProgressBar, ProgressStyle};

use crate::error::{Error, Result};
use crate::models::{RegistryRecord, RegistryType, EgrulRecord, EgripRecord};
use crate::output::{OutputFormat, OutputWriter};

/// Конфигурация парсера
#[derive(Debug, Clone)]
pub struct ParserConfig {
    /// Количество потоков для параллельной обработки
    pub num_threads: usize,
    /// Размер буфера канала
    pub channel_buffer_size: usize,
    /// Размер батча для записи
    pub batch_size: usize,
    /// Показывать прогресс
    pub show_progress: bool,
    /// Продолжать при ошибках
    pub continue_on_error: bool,
}

impl Default for ParserConfig {
    fn default() -> Self {
        Self {
            num_threads: num_cpus::get(),
            channel_buffer_size: 10000,
            batch_size: 5000,
            show_progress: true,
            continue_on_error: true,
        }
    }
}

/// Основной парсер
pub struct Parser {
    config: ParserConfig,
    stats: Arc<ParserStats>,
}

/// Статистика парсинга
#[derive(Debug, Default)]
pub struct ParserStats {
    pub files_processed: AtomicUsize,
    pub files_failed: AtomicUsize,
    pub egrul_records: AtomicUsize,
    pub egrip_records: AtomicUsize,
    pub errors: AtomicUsize,
}

impl ParserStats {
    pub fn total_records(&self) -> usize {
        self.egrul_records.load(Ordering::Relaxed) + self.egrip_records.load(Ordering::Relaxed)
    }

    pub fn print_summary(&self) {
        info!("=== Статистика парсинга ===");
        info!("Файлов обработано: {}", self.files_processed.load(Ordering::Relaxed));
        info!("Файлов с ошибками: {}", self.files_failed.load(Ordering::Relaxed));
        info!("Записей ЕГРЮЛ: {}", self.egrul_records.load(Ordering::Relaxed));
        info!("Записей ЕГРИП: {}", self.egrip_records.load(Ordering::Relaxed));
        info!("Всего записей: {}", self.total_records());
        info!("Ошибок: {}", self.errors.load(Ordering::Relaxed));
    }
}

impl Parser {
    /// Создание нового парсера с настройками по умолчанию
    pub fn new() -> Self {
        Self::with_config(ParserConfig::default())
    }

    /// Создание парсера с заданной конфигурацией
    pub fn with_config(config: ParserConfig) -> Self {
        // Настройка пула потоков rayon
        rayon::ThreadPoolBuilder::new()
            .num_threads(config.num_threads)
            .build_global()
            .ok();

        Self {
            config,
            stats: Arc::new(ParserStats::default()),
        }
    }

    /// Получение статистики
    pub fn stats(&self) -> &ParserStats {
        &self.stats
    }

    /// Парсинг одного файла
    pub fn parse_file(&self, path: &Path) -> Result<Vec<RegistryRecord>> {
        debug!("Парсинг файла: {:?}", path);

        let reader = XmlReader::from_file(path)?;
        let content = reader.read_to_string()?;

        // Определяем тип реестра
        let registry_type = RegistryType::from_content(&content)
            .or_else(|| RegistryType::from_filename(path.file_name()?.to_str()?))
            .ok_or_else(|| Error::UnknownRegistryType(path.display().to_string()))?;

        let records = match registry_type {
            RegistryType::Egrul => {
                let parser = EgrulXmlParser::new();
                parser.parse(&content)?
                    .into_iter()
                    .map(RegistryRecord::Egrul)
                    .collect()
            }
            RegistryType::Egrip => {
                let parser = EgripXmlParser::new();
                parser.parse(&content)?
                    .into_iter()
                    .map(RegistryRecord::Egrip)
                    .collect()
            }
        };

        Ok(records)
    }

    /// Парсинг директории с файлами
    pub fn parse_directory(
        &self,
        input_dir: &Path,
        output_dir: &Path,
        format: OutputFormat,
    ) -> Result<()> {
        info!("Начало парсинга директории: {:?}", input_dir);

        // Собираем список XML файлов
        let files: Vec<_> = WalkDir::new(input_dir)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| {
                e.path()
                    .extension()
                    .map(|ext| ext.eq_ignore_ascii_case("xml"))
                    .unwrap_or(false)
            })
            .map(|e| e.path().to_path_buf())
            .collect();

        info!("Найдено {} XML файлов", files.len());

        if files.is_empty() {
            warn!("XML файлы не найдены в {:?}", input_dir);
            return Ok(());
        }

        // Создаем выходную директорию
        std::fs::create_dir_all(output_dir)?;

        // Настраиваем прогресс-бар
        let progress = if self.config.show_progress {
            let pb = ProgressBar::new(files.len() as u64);
            pb.set_style(
                ProgressStyle::default_bar()
                    .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {pos}/{len} ({eta}) {msg}")
                    .unwrap()
                    .progress_chars("#>-"),
            );
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
            let format = format.clone();
            let batch_size = config.batch_size;
            move || -> Result<()> {
                let output_path = output_dir.join(format!("egrul.{}", format.extension()));
                let mut writer = OutputWriter::new(&output_path, format)?;
                let mut batch = Vec::with_capacity(batch_size);

                for record in egrul_rx {
                    batch.push(record);
                    if batch.len() >= batch_size {
                        writer.write_egrul_batch(&batch)?;
                        batch.clear();
                    }
                }

                // Записываем остаток
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
            let format = format.clone();
            let batch_size = config.batch_size;
            move || -> Result<()> {
                let output_path = output_dir.join(format!("egrip.{}", format.extension()));
                let mut writer = OutputWriter::new(&output_path, format)?;
                let mut batch = Vec::with_capacity(batch_size);

                for record in egrip_rx {
                    batch.push(record);
                    if batch.len() >= batch_size {
                        writer.write_egrip_batch(&batch)?;
                        batch.clear();
                    }
                }

                // Записываем остаток
                if !batch.is_empty() {
                    writer.write_egrip_batch(&batch)?;
                }

                writer.finish()?;
                Ok(())
            }
        });

        // Параллельный парсинг файлов
        files.par_iter().for_each(|file_path| {
            match self.parse_file(file_path) {
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
                }
                Err(e) => {
                    stats.files_failed.fetch_add(1, Ordering::Relaxed);
                    stats.errors.fetch_add(1, Ordering::Relaxed);
                    if !config.continue_on_error {
                        error!("Ошибка парсинга {:?}: {}", file_path, e);
                    } else {
                        warn!("Пропуск файла {:?}: {}", file_path, e);
                    }
                }
            }

            if let Some(ref pb) = progress {
                pb.inc(1);
                pb.set_message(format!(
                    "ЕГРЮЛ: {} | ЕГРИП: {}",
                    stats.egrul_records.load(Ordering::Relaxed),
                    stats.egrip_records.load(Ordering::Relaxed)
                ));
            }
        });

        // Закрываем каналы
        drop(egrul_tx);
        drop(egrip_tx);

        // Ждем завершения записи
        egrul_writer_handle.join().map_err(|_| Error::other("Ошибка потока записи ЕГРЮЛ"))??;
        egrip_writer_handle.join().map_err(|_| Error::other("Ошибка потока записи ЕГРИП"))??;

        if let Some(pb) = progress {
            pb.finish_with_message("Завершено");
        }

        self.stats.print_summary();
        Ok(())
    }
}

impl Default for Parser {
    fn default() -> Self {
        Self::new()
    }
}

/// Получение числа CPU
mod num_cpus {
    pub fn get() -> usize {
        std::thread::available_parallelism()
            .map(|p| p.get())
            .unwrap_or(4)
    }
}

