//! # ЕГРЮЛ/ЕГРИП Parser
//!
//! Высокопроизводительный парсер XML файлов реестров ЕГРЮЛ и ЕГРИП.
//!
//! ## Возможности
//!
//! - Потоковый парсинг XML без загрузки всего файла в память
//! - Поддержка кодировки windows-1251
//! - Параллельная обработка файлов через rayon
//! - Вывод в формат Apache Parquet, JSON, JSONL
//! - Обработка 500K+ записей/мин
//! - Конфигурация через TOML и переменные окружения
//! - CLI интерфейс с прогресс-баром и статистикой
//!
//! ## Пример использования
//!
//! ```rust,no_run
//! use egrul_parser::{Parser, OutputFormat};
//! use std::path::Path;
//!
//! fn main() -> egrul_parser::Result<()> {
//!     let parser = Parser::new();
//!     
//!     // Парсинг одного файла
//!     let records = parser.parse_file(Path::new("./data/file.xml"))?;
//!     println!("Найдено {} записей", records.len());
//!     
//!     // Парсинг директории
//!     parser.parse_directory(
//!         Path::new("./input"),
//!         Path::new("./output"),
//!         OutputFormat::Parquet
//!     )?;
//!     
//!     Ok(())
//! }
//! ```
//!
//! ## Структура модулей
//!
//! - `models` - структуры данных для ЕГРЮЛ и ЕГРИП
//! - `parser` - парсеры XML файлов
//! - `output` - писатели в различные форматы
//! - `error` - типы ошибок
//! - `config` - конфигурация приложения
//! - `parallel` - параллельная обработка
//! - `cli` - интерфейс командной строки

pub mod models;
pub mod parser;
pub mod output;
pub mod error;
pub mod config;
pub mod parallel;
pub mod cli;
pub mod deduplication;

// Re-export основных типов
pub use error::{Error, Result};
pub use models::{
    EgrulRecord, 
    EgripRecord, 
    RegistryType, 
    RegistryRecord,
    // Common types
    Person,
    Address,
    Activity,
    Capital,
    Share,
    Founder,
    HistoryRecord,
    EntityStatus,
    TaxAuthority,
    PensionFund,
    SocialInsurance,
    RegistrationAuthority,
    License,
    Document,
    // EGRUL specific
    egrul::HeadInfo,
    egrul::RegistrationInfo,
    egrul::Branch,
    egrul::BranchType,
    egrul::BankruptcyInfo,
    egrul::ReorganizationInfo,
    egrul::LiquidationInfo,
    // EGRIP specific
    egrip::Gender,
    egrip::CitizenshipInfo,
    egrip::CitizenshipType,
    egrip::IpRegistrationInfo,
    egrip::IpBankruptcyInfo,
};
pub use parser::{Parser, ParserConfig, XmlReader, EgrulXmlParser, EgripXmlParser};
pub use output::{OutputFormat, OutputWriter};
pub use config::AppConfig;
pub use parallel::{ParallelProcessor, ProcessingStats};

/// Версия библиотеки
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

/// Информация о библиотеке
pub fn info() -> String {
    format!(
        "egrul-parser v{}\nВысокопроизводительный парсер XML файлов ЕГРЮЛ/ЕГРИП",
        VERSION
    )
}
