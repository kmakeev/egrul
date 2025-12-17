//! ЕГРЮЛ/ЕГРИП XML Parser
//!
//! Высокопроизводительный парсер XML файлов реестров ФНС
//!
//! # Использование
//!
//! ```bash
//! # Парсинг директории
//! egrul-parser parse -i ./data -o ./output -f parquet
//!
//! # Параллельный парсинг с 8 воркерами
//! egrul-parser parse -i ./data -o ./output -w 8 --batch-size 10000
//!
//! # Валидация файла
//! egrul-parser validate -i ./data/file.xml
//!
//! # Статистика
//! egrul-parser stats -i ./output
//!
//! # Информация о файле
//! egrul-parser info -i ./data/file.xml
//!
//! # Управление конфигурацией
//! egrul-parser config show
//! egrul-parser config init
//! egrul-parser config path
//! ```

use clap::Parser as ClapParser;
use tracing::Level;
use tracing_subscriber::{fmt, EnvFilter};

use egrul_parser::cli::{Cli, Commands};
use egrul_parser::config::AppConfig;
use egrul_parser::Result;

fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // Загружаем конфигурацию
    let config = if let Some(ref config_path) = cli.config {
        AppConfig::load_from_file(config_path)
            .unwrap_or_else(|e| {
                eprintln!("Ошибка загрузки конфигурации: {}", e);
                AppConfig::default()
            })
    } else {
        AppConfig::load()
    };
    
    // Уровень логирования
    let log_level = if cli.quiet {
        Level::ERROR
    } else {
        match cli.log_level.to_lowercase().as_str() {
            "error" => Level::ERROR,
            "warn" => Level::WARN,
            "info" => Level::INFO,
            "debug" => Level::DEBUG,
            "trace" => Level::TRACE,
            _ => match config.logging.level.to_lowercase().as_str() {
                "error" => Level::ERROR,
                "warn" => Level::WARN,
                "debug" => Level::DEBUG,
                "trace" => Level::TRACE,
                _ => Level::INFO,
            }
        }
    };
    
    // Инициализация логирования
    let filter = EnvFilter::from_default_env()
        .add_directive(format!("egrul_parser={}", log_level).parse().unwrap());
    
    fmt()
        .with_env_filter(filter)
        .with_target(false)
        .with_thread_ids(false)
        .init();
    
    // Выполняем команду
    match cli.command {
        Commands::Parse(args) => {
            egrul_parser::cli::execute_parse(args, &config)
        }
        Commands::Validate(args) => {
            egrul_parser::cli::execute_validate(args, &config)
        }
        Commands::Stats(args) => {
            egrul_parser::cli::execute_stats(args, &config)
        }
        Commands::Info(args) => {
            egrul_parser::cli::execute_info(args, &config)
        }
        Commands::Config(args) => {
            egrul_parser::cli::execute_config(args.action, &config)
        }
    }
}
