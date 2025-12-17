//! CLI интерфейс для парсера ЕГРЮЛ/ЕГРИП
//!
//! Поддерживает команды:
//! - `parse` - парсинг XML файлов
//! - `validate` - валидация XML файлов
//! - `stats` - статистика по Parquet файлам
//! - `info` - информация о файле
//! - `config` - управление конфигурацией

mod commands;

pub use commands::*;

use std::path::PathBuf;
use clap::{Parser as ClapParser, Subcommand, ValueEnum, Args};

use crate::output::OutputFormat;

/// ЕГРЮЛ/ЕГРИП XML Parser CLI
#[derive(ClapParser)]
#[command(
    name = "egrul-parser",
    author = "EGRUL System",
    version,
    about = "Высокопроизводительный парсер XML файлов ЕГРЮЛ/ЕГРИП",
    long_about = r#"
Парсер XML файлов реестров ЕГРЮЛ и ЕГРИП от ФНС России.

Поддерживает:
  - Потоковый парсинг без загрузки всего файла в память
  - Кодировку windows-1251
  - Параллельную обработку файлов
  - Вывод в форматы Parquet, JSON, JSONL

Примеры:
  # Парсинг директории
  egrul-parser parse -i ./data -o ./output -f parquet

  # Параллельный парсинг с 8 воркерами
  egrul-parser parse -i ./data -o ./output -w 8 --batch-size 10000

  # Валидация файла
  egrul-parser validate -i ./data/file.xml

  # Статистика по результатам
  egrul-parser stats -i ./output

  # Информация о файле
  egrul-parser info -i ./data/file.xml

Конфигурация:
  Файл: ~/.config/egrul-parser/config.toml
  Переменные окружения: EGRUL_* (например EGRUL_BATCH_SIZE=10000)
"#
)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,

    /// Уровень логирования (error, warn, info, debug, trace)
    #[arg(short, long, default_value = "info", global = true)]
    pub log_level: String,

    /// Путь к файлу конфигурации
    #[arg(short, long, global = true)]
    pub config: Option<PathBuf>,
    
    /// Тихий режим (только ошибки)
    #[arg(short, long, global = true)]
    pub quiet: bool,
}

/// Команды CLI
#[derive(Subcommand)]
pub enum Commands {
    /// Парсинг XML файлов
    Parse(ParseArgs),

    /// Валидация XML файлов
    Validate(ValidateArgs),

    /// Статистика по результатам парсинга
    Stats(StatsArgs),

    /// Информация о XML файле
    Info(InfoArgs),

    /// Управление конфигурацией
    Config(ConfigArgs),
}

/// Аргументы команды parse
#[derive(Args)]
pub struct ParseArgs {
    /// Входной файл или директория
    #[arg(short, long)]
    pub input: PathBuf,

    /// Выходная директория
    #[arg(short, long, default_value = "./output")]
    pub output: PathBuf,

    /// Формат вывода
    #[arg(short, long, value_enum, default_value = "parquet")]
    pub format: OutputFormatArg,

    /// Количество воркеров (по умолчанию: число CPU)
    #[arg(short, long)]
    pub workers: Option<usize>,

    /// Размер батча для записи
    #[arg(short, long, default_value = "5000")]
    pub batch_size: usize,

    /// Продолжать при ошибках
    #[arg(long, default_value = "true")]
    pub continue_on_error: bool,

    /// Скрыть прогресс-бар
    #[arg(long)]
    pub no_progress: bool,
    
    /// Рекурсивный обход директорий
    #[arg(short, long, default_value = "true")]
    pub recursive: bool,
    
    /// Маска файлов (например *.xml)
    #[arg(long, default_value = "*.xml")]
    pub pattern: String,
    
    /// Пропустить существующие выходные файлы
    #[arg(long)]
    pub skip_existing: bool,
}

/// Аргументы команды validate
#[derive(Args)]
pub struct ValidateArgs {
    /// Входной файл или директория
    #[arg(short, long)]
    pub input: PathBuf,
    
    /// Показывать только ошибки
    #[arg(long)]
    pub errors_only: bool,
    
    /// Вывод в формате JSON
    #[arg(long)]
    pub json: bool,
}

/// Аргументы команды stats
#[derive(Args)]
pub struct StatsArgs {
    /// Входная директория с Parquet/JSON файлами
    #[arg(short, long)]
    pub input: PathBuf,
    
    /// Детальная статистика
    #[arg(short, long)]
    pub detailed: bool,
    
    /// Вывод в формате JSON
    #[arg(long)]
    pub json: bool,
}

/// Аргументы команды info
#[derive(Args)]
pub struct InfoArgs {
    /// Входной файл
    #[arg(short, long)]
    pub input: PathBuf,
    
    /// Показать примеры записей
    #[arg(long)]
    pub samples: bool,
    
    /// Количество примеров
    #[arg(long, default_value = "5")]
    pub sample_count: usize,
}

/// Аргументы команды config
#[derive(Args)]
pub struct ConfigArgs {
    #[command(subcommand)]
    pub action: ConfigAction,
}

/// Действия с конфигурацией
#[derive(Subcommand)]
pub enum ConfigAction {
    /// Показать текущую конфигурацию
    Show,
    
    /// Сгенерировать файл конфигурации по умолчанию
    Init {
        /// Путь для сохранения
        #[arg(short, long)]
        output: Option<PathBuf>,
        
        /// Перезаписать существующий файл
        #[arg(long)]
        force: bool,
    },
    
    /// Показать путь к конфигурации
    Path,
}

/// Формат вывода (CLI)
#[derive(Clone, Copy, ValueEnum)]
pub enum OutputFormatArg {
    Parquet,
    Json,
    Jsonl,
}

impl From<OutputFormatArg> for OutputFormat {
    fn from(arg: OutputFormatArg) -> Self {
        match arg {
            OutputFormatArg::Parquet => OutputFormat::Parquet,
            OutputFormatArg::Json => OutputFormat::Json,
            OutputFormatArg::Jsonl => OutputFormat::JsonLines,
        }
    }
}

impl std::fmt::Display for OutputFormatArg {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            OutputFormatArg::Parquet => write!(f, "parquet"),
            OutputFormatArg::Json => write!(f, "json"),
            OutputFormatArg::Jsonl => write!(f, "jsonl"),
        }
    }
}

