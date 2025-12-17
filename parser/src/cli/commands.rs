//! Реализация команд CLI

use std::path::Path;
use std::time::Instant;

use console::style;
use indicatif::{HumanBytes, HumanDuration};
use tracing::info;
use walkdir::WalkDir;

use crate::config::AppConfig;
use crate::parallel::ParallelProcessor;
use crate::{Parser, ParserConfig, Result, Error, OutputFormat};
use crate::parser::XmlReader;
use crate::models::RegistryType;

use super::{ParseArgs, ValidateArgs, StatsArgs, InfoArgs, ConfigAction};

/// Выполнение команды parse
pub fn execute_parse(args: ParseArgs, config: &AppConfig) -> Result<()> {
    print_banner();
    
    info!("Входной путь:    {:?}", args.input);
    info!("Выходной путь:   {:?}", args.output);
    info!("Формат вывода:   {}", args.format);
    info!("");
    
    let num_workers = args.workers.unwrap_or_else(|| config.num_workers());
    let batch_size = if args.batch_size != 5000 { 
        args.batch_size 
    } else { 
        config.parser.batch_size 
    };
    
    let parser_config = ParserConfig {
        num_threads: num_workers,
        channel_buffer_size: config.parser.channel_buffer_size,
        batch_size,
        show_progress: !args.no_progress && config.parser.show_progress,
        continue_on_error: args.continue_on_error && config.parser.continue_on_error,
    };
    
    info!("Воркеров:        {}", num_workers);
    info!("Размер батча:    {}", batch_size);
    info!("");
    
    let start = Instant::now();
    let format: OutputFormat = args.format.into();
    
    if args.input.is_file() {
        // Парсинг одного файла
        execute_parse_file(&args.input, &args.output, format, &parser_config)?;
    } else if args.input.is_dir() {
        // Параллельный парсинг директории
        let processor = ParallelProcessor::new(parser_config.clone());
        processor.process_directory(&args.input, &args.output, format)?;
    } else {
        return Err(Error::config(format!("Путь не существует: {:?}", args.input)));
    }
    
    let elapsed = start.elapsed();
    info!("");
    info!("═══════════════════════════════════════════════════════════");
    info!("Время выполнения: {}", HumanDuration(elapsed));
    info!("═══════════════════════════════════════════════════════════");
    
    Ok(())
}

/// Парсинг одного файла
fn execute_parse_file(
    input: &Path,
    output: &Path,
    format: OutputFormat,
    config: &ParserConfig,
) -> Result<()> {
    use crate::{OutputWriter, RegistryRecord};
    
    info!("Парсинг файла...");
    let parser = Parser::with_config(config.clone());
    let records = parser.parse_file(input)?;
    
    std::fs::create_dir_all(output)?;
    
    let egrul_path = output.join(format!("egrul.{}", format.extension()));
    let egrip_path = output.join(format!("egrip.{}", format.extension()));
    
    let mut egrul_writer = OutputWriter::new(&egrul_path, format)?;
    let mut egrip_writer = OutputWriter::new(&egrip_path, format)?;
    
    let mut egrul_records = Vec::new();
    let mut egrip_records = Vec::new();
    
    for record in records {
        match record {
            RegistryRecord::Egrul(r) => egrul_records.push(r),
            RegistryRecord::Egrip(r) => egrip_records.push(r),
        }
    }
    
    if !egrul_records.is_empty() {
        egrul_writer.write_egrul_batch(&egrul_records)?;
        egrul_writer.finish()?;
        info!("Записано {} записей ЕГРЮЛ в {:?}", egrul_records.len(), egrul_path);
    }
    
    if !egrip_records.is_empty() {
        egrip_writer.write_egrip_batch(&egrip_records)?;
        egrip_writer.finish()?;
        info!("Записано {} записей ЕГРИП в {:?}", egrip_records.len(), egrip_path);
    }
    
    Ok(())
}

/// Выполнение команды validate
pub fn execute_validate(args: ValidateArgs, _config: &AppConfig) -> Result<()> {
    info!("Валидация: {:?}", args.input);
    info!("");
    
    let files: Vec<_> = if args.input.is_file() {
        vec![args.input.clone()]
    } else {
        WalkDir::new(&args.input)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| {
                e.path()
                    .extension()
                    .map(|ext| ext.eq_ignore_ascii_case("xml"))
                    .unwrap_or(false)
            })
            .map(|e| e.path().to_path_buf())
            .collect()
    };
    
    info!("Файлов для проверки: {}", files.len());
    info!("");
    
    let parser = Parser::new();
    let mut valid = 0;
    let mut invalid = 0;
    let mut results = Vec::new();
    
    for file in &files {
        match parser.parse_file(file) {
            Ok(records) => {
                if !args.errors_only {
                    let msg = format!("✓ {:?} ({} записей)", 
                        file.file_name().unwrap_or_default(), 
                        records.len()
                    );
                    if args.json {
                        results.push(serde_json::json!({
                            "file": file.display().to_string(),
                            "status": "valid",
                            "records": records.len()
                        }));
                    } else {
                        println!("{}", style(&msg).green());
                    }
                }
                valid += 1;
            }
            Err(e) => {
                let msg = format!("✗ {:?}: {}", 
                    file.file_name().unwrap_or_default(), 
                    e
                );
                if args.json {
                    results.push(serde_json::json!({
                        "file": file.display().to_string(),
                        "status": "invalid",
                        "error": e.to_string()
                    }));
                } else {
                    println!("{}", style(&msg).red());
                }
                invalid += 1;
            }
        }
    }
    
    if args.json {
        let summary = serde_json::json!({
            "total": files.len(),
            "valid": valid,
            "invalid": invalid,
            "files": results
        });
        println!("{}", serde_json::to_string_pretty(&summary)?);
    } else {
        info!("");
        info!("Результат: {} валидных, {} с ошибками", valid, invalid);
    }
    
    if invalid > 0 {
        std::process::exit(1);
    }
    
    Ok(())
}

/// Выполнение команды stats
pub fn execute_stats(args: StatsArgs, _config: &AppConfig) -> Result<()> {
    info!("Анализ директории: {:?}", args.input);
    info!("");
    
    let mut stats = OutputStats::default();
    
    // Ищем все поддерживаемые файлы
    for entry in WalkDir::new(&args.input)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().is_file())
    {
        let path = entry.path();
        let ext = path.extension()
            .and_then(|e| e.to_str())
            .unwrap_or("");
        
        let metadata = std::fs::metadata(path)?;
        let size = metadata.len();
        
        match ext.to_lowercase().as_str() {
            "parquet" => {
                stats.parquet_files += 1;
                stats.parquet_size += size;
                
                if args.detailed {
                    if let Ok(info) = get_parquet_info(path) {
                        stats.total_records += info.num_rows;
                    }
                }
            }
            "json" | "jsonl" => {
                stats.json_files += 1;
                stats.json_size += size;
                
                if args.detailed {
                    if let Ok(count) = count_json_records(path) {
                        stats.total_records += count;
                    }
                }
            }
            _ => {}
        }
    }
    
    if args.json {
        let json_stats = serde_json::json!({
            "parquet_files": stats.parquet_files,
            "parquet_size": stats.parquet_size,
            "json_files": stats.json_files,
            "json_size": stats.json_size,
            "total_records": stats.total_records,
            "total_size": stats.parquet_size + stats.json_size
        });
        println!("{}", serde_json::to_string_pretty(&json_stats)?);
    } else {
        println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
        println!("{}", style("                     СТАТИСТИКА                            ").cyan().bold());
        println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
        println!("");
        
        println!("{}  Parquet файлов:  {}", style("📁").cyan(), stats.parquet_files);
        println!("{}  Размер Parquet:  {}", style("💾").cyan(), HumanBytes(stats.parquet_size));
        println!("");
        println!("{}  JSON файлов:     {}", style("📁").cyan(), stats.json_files);
        println!("{}  Размер JSON:     {}", style("💾").cyan(), HumanBytes(stats.json_size));
        println!("");
        println!("{}  Всего файлов:    {}", style("📊").cyan(), stats.parquet_files + stats.json_files);
        println!("{}  Общий размер:    {}", style("💿").cyan(), HumanBytes(stats.parquet_size + stats.json_size));
        
        if args.detailed && stats.total_records > 0 {
            println!("");
            println!("{}  Всего записей:   {}", style("📝").cyan(), stats.total_records);
        }
        
        println!("");
        println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
    }
    
    Ok(())
}

/// Выполнение команды info
pub fn execute_info(args: InfoArgs, _config: &AppConfig) -> Result<()> {
    info!("Анализ файла: {:?}", args.input);
    info!("");
    
    let reader = XmlReader::from_file(&args.input)?;
    let content = reader.read_to_string()?;
    
    let registry_type = RegistryType::from_content(&content)
        .or_else(|| RegistryType::from_filename(args.input.file_name()?.to_str()?));
    
    let metadata = std::fs::metadata(&args.input)?;
    
    println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
    println!("{}", style("                 ИНФОРМАЦИЯ О ФАЙЛЕ                        ").cyan().bold());
    println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
    println!("");
    
    println!("{}  Файл:           {:?}", style("📄").cyan(), args.input.file_name().unwrap_or_default());
    println!("{}  Размер:         {}", style("💾").cyan(), HumanBytes(metadata.len()));
    println!("{}  Кодировка:      {:?}", style("🔤").cyan(), reader.encoding());
    
    if let Some(rt) = registry_type {
        println!("{}  Тип реестра:    {}", style("📋").cyan(), rt);
    } else {
        println!("{}  Тип реестра:    Не определен", style("⚠️").yellow());
    }
    
    let egrul_count = content.matches("<СвЮЛ").count();
    let egrip_count = content.matches("<СвИП").count();
    
    println!("");
    println!("{}  Записей ЕГРЮЛ:  {}", style("🏢").cyan(), egrul_count);
    println!("{}  Записей ЕГРИП:  {}", style("👤").cyan(), egrip_count);
    println!("{}  Всего записей:  {}", style("📊").cyan(), egrul_count + egrip_count);
    
    if args.samples && (egrul_count > 0 || egrip_count > 0) {
        println!("");
        println!("{}", style("─── Примеры записей ───").dim());
        
        let parser = Parser::new();
        if let Ok(records) = parser.parse_file(&args.input) {
            for (i, record) in records.iter().take(args.sample_count).enumerate() {
                println!("");
                println!("{}  Запись #{}", style("•").cyan(), i + 1);
                match record {
                    crate::models::RegistryRecord::Egrul(r) => {
                        println!("   ОГРН: {}", r.ogrn);
                        println!("   Название: {}", r.full_name);
                        println!("   ИНН: {}", r.inn);
                    }
                    crate::models::RegistryRecord::Egrip(r) => {
                        println!("   ОГРНИП: {}", r.ogrnip);
                        println!("   ФИО: {}", r.person.full_name());
                        println!("   ИНН: {}", r.inn);
                    }
                }
            }
        }
    }
    
    println!("");
    println!("{}", style("═══════════════════════════════════════════════════════════").cyan());
    
    Ok(())
}

/// Выполнение команды config
pub fn execute_config(action: ConfigAction, config: &AppConfig) -> Result<()> {
    match action {
        ConfigAction::Show => {
            let toml_str = toml::to_string_pretty(config)
                .map_err(|e| Error::config(e.to_string()))?;
            println!("{}", toml_str);
        }
        
        ConfigAction::Init { output, force } => {
            let path = output.unwrap_or_else(|| {
                AppConfig::default_config_path()
                    .unwrap_or_else(|| std::path::PathBuf::from("./config.toml"))
            });
            
            if path.exists() && !force {
                return Err(Error::config(format!(
                    "Файл {:?} уже существует. Используйте --force для перезаписи.",
                    path
                )));
            }
            
            let default_config = AppConfig::default();
            default_config.save_to_file(&path)
                .map_err(|e| Error::config(e.to_string()))?;
            
            println!("{} Конфигурация сохранена в {:?}", style("✓").green(), path);
        }
        
        ConfigAction::Path => {
            if let Some(path) = AppConfig::default_config_path() {
                println!("{}", path.display());
            } else {
                println!("Не удалось определить путь к конфигурации");
            }
        }
    }
    
    Ok(())
}

/// Вывод баннера
fn print_banner() {
    info!("╔════════════════════════════════════════════════════════════╗");
    info!("║       ЕГРЮЛ/ЕГРИП XML Parser v{}                     ║", env!("CARGO_PKG_VERSION"));
    info!("╚════════════════════════════════════════════════════════════╝");
    info!("");
}

/// Статистика вывода
#[derive(Default)]
struct OutputStats {
    parquet_files: usize,
    parquet_size: u64,
    json_files: usize,
    json_size: u64,
    total_records: usize,
}

/// Информация о Parquet файле
struct ParquetInfo {
    num_rows: usize,
}

/// Получение информации о Parquet файле
fn get_parquet_info(path: &Path) -> Result<ParquetInfo> {
    use parquet::file::reader::{FileReader, SerializedFileReader};
    use std::fs::File;
    
    let file = File::open(path)?;
    let reader = SerializedFileReader::new(file)
        .map_err(|e| Error::other(e.to_string()))?;
    
    let metadata = reader.metadata();
    let num_rows = metadata.file_metadata().num_rows() as usize;
    
    Ok(ParquetInfo { num_rows })
}

/// Подсчёт записей в JSON/JSONL файле
fn count_json_records(path: &Path) -> Result<usize> {
    use std::io::{BufRead, BufReader};
    use std::fs::File;
    
    let ext = path.extension()
        .and_then(|e| e.to_str())
        .unwrap_or("");
    
    if ext.to_lowercase() == "jsonl" {
        // JSONL - каждая строка = запись
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        Ok(reader.lines().count())
    } else {
        // JSON массив
        let content = std::fs::read_to_string(path)?;
        let value: serde_json::Value = serde_json::from_str(&content)?;
        if let Some(arr) = value.as_array() {
            Ok(arr.len())
        } else {
            Ok(1)
        }
    }
}

