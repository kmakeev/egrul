//! Модуль обработки ошибок

use thiserror::Error;

/// Тип результата для библиотеки
pub type Result<T> = std::result::Result<T, Error>;

/// Ошибки парсера ЕГРЮЛ/ЕГРИП
#[derive(Error, Debug)]
pub enum Error {
    /// Ошибка чтения файла
    #[error("Ошибка чтения файла: {0}")]
    Io(#[from] std::io::Error),

    /// Ошибка парсинга XML
    #[error("Ошибка парсинга XML: {0}")]
    Xml(#[from] quick_xml::Error),

    /// Ошибка декодирования
    #[error("Ошибка декодирования: {0}")]
    Encoding(String),

    /// Ошибка сериализации JSON
    #[error("Ошибка JSON: {0}")]
    Json(#[from] serde_json::Error),

    /// Ошибка записи Parquet
    #[error("Ошибка Parquet: {0}")]
    Parquet(#[from] parquet::errors::ParquetError),

    /// Ошибка Arrow
    #[error("Ошибка Arrow: {0}")]
    Arrow(#[from] arrow::error::ArrowError),

    /// Неизвестный тип реестра
    #[error("Неизвестный тип реестра: {0}")]
    UnknownRegistryType(String),

    /// Отсутствует обязательное поле
    #[error("Отсутствует обязательное поле: {0}")]
    MissingField(String),

    /// Ошибка парсинга даты
    #[error("Ошибка парсинга даты: {0}")]
    DateParse(String),

    /// Ошибка конфигурации
    #[error("Ошибка конфигурации: {0}")]
    Config(String),

    /// Общая ошибка
    #[error("{0}")]
    Other(String),
}

impl Error {
    pub fn encoding(msg: impl Into<String>) -> Self {
        Error::Encoding(msg.into())
    }

    pub fn missing_field(field: impl Into<String>) -> Self {
        Error::MissingField(field.into())
    }

    pub fn date_parse(msg: impl Into<String>) -> Self {
        Error::DateParse(msg.into())
    }

    pub fn config(msg: impl Into<String>) -> Self {
        Error::Config(msg.into())
    }

    pub fn other(msg: impl Into<String>) -> Self {
        Error::Other(msg.into())
    }
}

