//! Модуль вывода данных

mod parquet_writer;

pub use parquet_writer::*;

use std::path::Path;

use crate::error::{Error, Result};
use crate::models::{EgrulRecord, EgripRecord};

/// Формат вывода
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OutputFormat {
    /// Apache Parquet
    Parquet,
    /// JSON Lines (один JSON объект на строку)
    JsonLines,
    /// Обычный JSON массив
    Json,
}

impl OutputFormat {
    /// Расширение файла
    pub fn extension(&self) -> &'static str {
        match self {
            OutputFormat::Parquet => "parquet",
            OutputFormat::JsonLines => "jsonl",
            OutputFormat::Json => "json",
        }
    }

    /// Определение формата по расширению
    pub fn from_extension(ext: &str) -> Option<Self> {
        match ext.to_lowercase().as_str() {
            "parquet" | "pq" => Some(OutputFormat::Parquet),
            "jsonl" | "ndjson" => Some(OutputFormat::JsonLines),
            "json" => Some(OutputFormat::Json),
            _ => None,
        }
    }
}

impl std::str::FromStr for OutputFormat {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self> {
        match s.to_lowercase().as_str() {
            "parquet" | "pq" => Ok(OutputFormat::Parquet),
            "jsonl" | "jsonlines" | "ndjson" => Ok(OutputFormat::JsonLines),
            "json" => Ok(OutputFormat::Json),
            _ => Err(Error::config(format!("Неизвестный формат вывода: {}", s))),
        }
    }
}

/// Писатель выходных данных
pub struct OutputWriter {
    inner: WriterInner,
}

enum WriterInner {
    Parquet(ParquetOutputWriter),
    JsonLines(JsonLinesWriter),
    Json(JsonArrayWriter),
}

impl OutputWriter {
    /// Создание нового писателя
    pub fn new(path: &Path, format: OutputFormat) -> Result<Self> {
        let inner = match format {
            OutputFormat::Parquet => {
                WriterInner::Parquet(ParquetOutputWriter::new(path)?)
            }
            OutputFormat::JsonLines => {
                WriterInner::JsonLines(JsonLinesWriter::new(path)?)
            }
            OutputFormat::Json => {
                WriterInner::Json(JsonArrayWriter::new(path)?)
            }
        };

        Ok(Self { inner })
    }

    /// Запись батча ЕГРЮЛ
    pub fn write_egrul_batch(&mut self, records: &[EgrulRecord]) -> Result<()> {
        match &mut self.inner {
            WriterInner::Parquet(w) => w.write_egrul_batch(records),
            WriterInner::JsonLines(w) => w.write_egrul_batch(records),
            WriterInner::Json(w) => w.write_egrul_batch(records),
        }
    }

    /// Запись батча ЕГРИП
    pub fn write_egrip_batch(&mut self, records: &[EgripRecord]) -> Result<()> {
        match &mut self.inner {
            WriterInner::Parquet(w) => w.write_egrip_batch(records),
            WriterInner::JsonLines(w) => w.write_egrip_batch(records),
            WriterInner::Json(w) => w.write_egrip_batch(records),
        }
    }

    /// Завершение записи
    pub fn finish(self) -> Result<()> {
        match self.inner {
            WriterInner::Parquet(w) => w.finish(),
            WriterInner::JsonLines(w) => w.finish(),
            WriterInner::Json(w) => w.finish(),
        }
    }
}

/// JSON Lines писатель
pub struct JsonLinesWriter {
    file: std::io::BufWriter<std::fs::File>,
}

impl JsonLinesWriter {
    pub fn new(path: &Path) -> Result<Self> {
        let file = std::fs::File::create(path)?;
        let file = std::io::BufWriter::new(file);
        Ok(Self { file })
    }

    pub fn write_egrul_batch(&mut self, records: &[EgrulRecord]) -> Result<()> {
        use std::io::Write;
        for record in records {
            serde_json::to_writer(&mut self.file, record)?;
            writeln!(self.file)?;
        }
        Ok(())
    }

    pub fn write_egrip_batch(&mut self, records: &[EgripRecord]) -> Result<()> {
        use std::io::Write;
        for record in records {
            serde_json::to_writer(&mut self.file, record)?;
            writeln!(self.file)?;
        }
        Ok(())
    }

    pub fn finish(mut self) -> Result<()> {
        use std::io::Write;
        self.file.flush()?;
        Ok(())
    }
}

/// JSON Array писатель
pub struct JsonArrayWriter {
    file: std::io::BufWriter<std::fs::File>,
    first: bool,
}

impl JsonArrayWriter {
    pub fn new(path: &Path) -> Result<Self> {
        use std::io::Write;
        let file = std::fs::File::create(path)?;
        let mut file = std::io::BufWriter::new(file);
        write!(file, "[")?;
        Ok(Self { file, first: true })
    }

    pub fn write_egrul_batch(&mut self, records: &[EgrulRecord]) -> Result<()> {
        use std::io::Write;
        for record in records {
            if !self.first {
                write!(self.file, ",")?;
            }
            self.first = false;
            writeln!(self.file)?;
            serde_json::to_writer(&mut self.file, record)?;
        }
        Ok(())
    }

    pub fn write_egrip_batch(&mut self, records: &[EgripRecord]) -> Result<()> {
        use std::io::Write;
        for record in records {
            if !self.first {
                write!(self.file, ",")?;
            }
            self.first = false;
            writeln!(self.file)?;
            serde_json::to_writer(&mut self.file, record)?;
        }
        Ok(())
    }

    pub fn finish(mut self) -> Result<()> {
        use std::io::Write;
        writeln!(self.file)?;
        writeln!(self.file, "]")?;
        self.file.flush()?;
        Ok(())
    }
}

