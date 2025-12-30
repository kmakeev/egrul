//! Конфигурация парсера
//!
//! Поддерживает загрузку из:
//! - TOML файла (по умолчанию ~/.config/egrul-parser/config.toml)
//! - Переменных окружения с префиксом EGRUL_
//! - Значений по умолчанию

use std::path::{Path, PathBuf};
use serde::{Deserialize, Serialize};
use tracing::{debug, info, warn};

/// Конфигурация приложения
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    /// Настройки парсера
    #[serde(default)]
    pub parser: ParserSettings,
    
    /// Настройки вывода
    #[serde(default)]
    pub output: OutputSettings,
    
    /// Настройки логирования
    #[serde(default)]
    pub logging: LoggingSettings,
    
    /// Настройки параллельной обработки
    #[serde(default)]
    pub parallel: ParallelSettings,
}

/// Настройки парсера
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParserSettings {
    /// Продолжать при ошибках
    #[serde(default = "default_continue_on_error")]
    pub continue_on_error: bool,
    
    /// Размер батча для записи
    #[serde(default = "default_batch_size")]
    pub batch_size: usize,
    
    /// Размер буфера канала
    #[serde(default = "default_channel_buffer_size")]
    pub channel_buffer_size: usize,
    
    /// Показывать прогресс
    #[serde(default = "default_show_progress")]
    pub show_progress: bool,
}

/// Настройки вывода
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OutputSettings {
    /// Формат вывода по умолчанию
    #[serde(default = "default_format")]
    pub format: String,
    
    /// Директория вывода по умолчанию
    #[serde(default = "default_output_dir")]
    pub output_dir: PathBuf,
    
    /// Сжатие для Parquet
    #[serde(default = "default_compression")]
    pub compression: String,
    
    /// Максимальный размер файла в МБ (0 = без ограничений)
    #[serde(default)]
    pub max_file_size_mb: usize,
    
    /// Максимальное количество записей в файле (0 = без ограничений)
    #[serde(default)]
    pub max_records_per_file: usize,
}

/// Настройки логирования
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoggingSettings {
    /// Уровень логирования
    #[serde(default = "default_log_level")]
    pub level: String,
    
    /// Путь к файлу лога (опционально)
    pub file: Option<PathBuf>,
    
    /// Формат логов (text, json)
    #[serde(default = "default_log_format")]
    pub format: String,
}

/// Настройки параллельной обработки
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParallelSettings {
    /// Количество воркеров (0 = авто)
    #[serde(default)]
    pub workers: usize,
    
    /// Максимальный размер очереди на воркер
    #[serde(default = "default_queue_size")]
    pub queue_size: usize,
    
    /// Распределение файлов по размеру
    #[serde(default = "default_distribute_by_size")]
    pub distribute_by_size: bool,
}

// Значения по умолчанию
fn default_continue_on_error() -> bool { true }
fn default_batch_size() -> usize { 5000 }
fn default_channel_buffer_size() -> usize { 10000 }
fn default_show_progress() -> bool { true }
fn default_format() -> String { "parquet".to_string() }
fn default_output_dir() -> PathBuf { PathBuf::from("./output") }
fn default_compression() -> String { "snappy".to_string() }
fn default_log_level() -> String { "info".to_string() }
fn default_log_format() -> String { "text".to_string() }
fn default_queue_size() -> usize { 1000 }
fn default_distribute_by_size() -> bool { true }

impl Default for ParserSettings {
    fn default() -> Self {
        Self {
            continue_on_error: default_continue_on_error(),
            batch_size: default_batch_size(),
            channel_buffer_size: default_channel_buffer_size(),
            show_progress: default_show_progress(),
        }
    }
}

impl Default for OutputSettings {
    fn default() -> Self {
        Self {
            format: default_format(),
            output_dir: default_output_dir(),
            compression: default_compression(),
            max_file_size_mb: 0, // Без ограничений по умолчанию
            max_records_per_file: 0, // Без ограничений по умолчанию
        }
    }
}

impl Default for LoggingSettings {
    fn default() -> Self {
        Self {
            level: default_log_level(),
            file: None,
            format: default_log_format(),
        }
    }
}

impl Default for ParallelSettings {
    fn default() -> Self {
        Self {
            workers: 0, // Авто
            queue_size: default_queue_size(),
            distribute_by_size: default_distribute_by_size(),
        }
    }
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            parser: ParserSettings::default(),
            output: OutputSettings::default(),
            logging: LoggingSettings::default(),
            parallel: ParallelSettings::default(),
        }
    }
}

impl AppConfig {
    /// Загрузка конфигурации
    pub fn load() -> Self {
        Self::load_from_paths(&[
            Self::default_config_path(),
            Some(PathBuf::from("./config.toml")),
            Some(PathBuf::from("./egrul-parser.toml")),
        ])
    }
    
    /// Загрузка из указанного файла
    pub fn load_from_file(path: &Path) -> Result<Self, ConfigError> {
        let content = std::fs::read_to_string(path)
            .map_err(|e| ConfigError::FileRead(path.to_path_buf(), e.to_string()))?;
        
        let mut config: AppConfig = toml::from_str(&content)
            .map_err(|e| ConfigError::Parse(e.to_string()))?;
        
        // Применяем переменные окружения поверх файла
        config.apply_env_overrides();
        
        Ok(config)
    }
    
    /// Загрузка из нескольких путей (первый найденный)
    pub fn load_from_paths(paths: &[Option<PathBuf>]) -> Self {
        // Загружаем .env если есть
        let _ = dotenvy::dotenv();
        
        for path in paths.iter().flatten() {
            if path.exists() {
                debug!("Загрузка конфигурации из {:?}", path);
                match Self::load_from_file(path) {
                    Ok(config) => {
                        info!("Конфигурация загружена из {:?}", path);
                        return config;
                    }
                    Err(e) => {
                        warn!("Ошибка загрузки конфигурации из {:?}: {}", path, e);
                    }
                }
            }
        }
        
        debug!("Используются настройки по умолчанию");
        let mut config = Self::default();
        config.apply_env_overrides();
        config
    }
    
    /// Путь к конфигурации по умолчанию
    pub fn default_config_path() -> Option<PathBuf> {
        dirs::config_dir().map(|p| p.join("egrul-parser").join("config.toml"))
    }
    
    /// Применение переменных окружения
    fn apply_env_overrides(&mut self) {
        // Parser settings
        if let Ok(val) = std::env::var("EGRUL_BATCH_SIZE") {
            if let Ok(size) = val.parse() {
                self.parser.batch_size = size;
            }
        }
        
        if let Ok(val) = std::env::var("EGRUL_CONTINUE_ON_ERROR") {
            self.parser.continue_on_error = val.to_lowercase() == "true" || val == "1";
        }
        
        if let Ok(val) = std::env::var("EGRUL_SHOW_PROGRESS") {
            self.parser.show_progress = val.to_lowercase() == "true" || val == "1";
        }
        
        if let Ok(val) = std::env::var("EGRUL_CHANNEL_BUFFER_SIZE") {
            if let Ok(size) = val.parse() {
                self.parser.channel_buffer_size = size;
            }
        }
        
        // Output settings
        if let Ok(val) = std::env::var("EGRUL_OUTPUT_FORMAT") {
            self.output.format = val;
        }
        
        if let Ok(val) = std::env::var("EGRUL_OUTPUT_DIR") {
            self.output.output_dir = PathBuf::from(val);
        }
        
        if let Ok(val) = std::env::var("EGRUL_COMPRESSION") {
            self.output.compression = val;
        }
        
        if let Ok(val) = std::env::var("EGRUL_MAX_FILE_SIZE_MB") {
            if let Ok(size) = val.parse() {
                self.output.max_file_size_mb = size;
            }
        }
        
        if let Ok(val) = std::env::var("EGRUL_MAX_RECORDS_PER_FILE") {
            if let Ok(records) = val.parse() {
                self.output.max_records_per_file = records;
            }
        }
        
        // Logging settings
        if let Ok(val) = std::env::var("EGRUL_LOG_LEVEL") {
            self.logging.level = val;
        }
        
        if let Ok(val) = std::env::var("EGRUL_LOG_FILE") {
            self.logging.file = Some(PathBuf::from(val));
        }
        
        // Parallel settings
        if let Ok(val) = std::env::var("EGRUL_WORKERS") {
            if let Ok(workers) = val.parse() {
                self.parallel.workers = workers;
            }
        }
        
        if let Ok(val) = std::env::var("EGRUL_QUEUE_SIZE") {
            if let Ok(size) = val.parse() {
                self.parallel.queue_size = size;
            }
        }
    }
    
    /// Сохранение конфигурации в файл
    pub fn save_to_file(&self, path: &Path) -> Result<(), ConfigError> {
        let content = toml::to_string_pretty(self)
            .map_err(|e| ConfigError::Serialize(e.to_string()))?;
        
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent)
                .map_err(|e| ConfigError::FileWrite(path.to_path_buf(), e.to_string()))?;
        }
        
        std::fs::write(path, content)
            .map_err(|e| ConfigError::FileWrite(path.to_path_buf(), e.to_string()))?;
        
        Ok(())
    }
    
    /// Количество воркеров (с учётом авто-определения)
    pub fn num_workers(&self) -> usize {
        if self.parallel.workers == 0 {
            num_cpus::get()
        } else {
            self.parallel.workers
        }
    }
}

/// Ошибки конфигурации
#[derive(Debug)]
pub enum ConfigError {
    FileRead(PathBuf, String),
    FileWrite(PathBuf, String),
    Parse(String),
    Serialize(String),
}

impl std::fmt::Display for ConfigError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ConfigError::FileRead(path, err) => {
                write!(f, "Ошибка чтения файла {:?}: {}", path, err)
            }
            ConfigError::FileWrite(path, err) => {
                write!(f, "Ошибка записи файла {:?}: {}", path, err)
            }
            ConfigError::Parse(err) => write!(f, "Ошибка парсинга конфигурации: {}", err),
            ConfigError::Serialize(err) => write!(f, "Ошибка сериализации конфигурации: {}", err),
        }
    }
}

impl std::error::Error for ConfigError {}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_default_config() {
        let config = AppConfig::default();
        assert_eq!(config.parser.batch_size, 5000);
        assert!(config.parser.continue_on_error);
        assert_eq!(config.output.format, "parquet");
    }
    
    #[test]
    fn test_toml_serialization() {
        let config = AppConfig::default();
        let toml_str = toml::to_string_pretty(&config).unwrap();
        assert!(toml_str.contains("[parser]"));
        assert!(toml_str.contains("[output]"));
    }
}

