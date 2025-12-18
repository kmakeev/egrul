//! Parquet писатель для ЕГРЮЛ/ЕГРИП

use std::path::Path;
use std::sync::Arc;
use std::fs::File;

use arrow::array::*;
use arrow::datatypes::{DataType, Field, Schema};
use arrow::record_batch::RecordBatch;
use parquet::arrow::ArrowWriter;
use parquet::basic::Compression;
use parquet::file::properties::WriterProperties;
use tracing::debug;

use crate::error::Result;
use crate::models::{EgrulRecord, EgripRecord};

/// Parquet писатель
pub struct ParquetOutputWriter {
    path: std::path::PathBuf,
    egrul_batches: Vec<RecordBatch>,
    egrip_batches: Vec<RecordBatch>,
}

impl ParquetOutputWriter {
    /// Создание нового писателя
    pub fn new(path: &Path) -> Result<Self> {
        Ok(Self {
            path: path.to_path_buf(),
            egrul_batches: Vec::new(),
            egrip_batches: Vec::new(),
        })
    }

    /// Схема для ЕГРЮЛ
    fn egrul_schema() -> Schema {
        Schema::new(vec![
            Field::new("ogrn", DataType::Utf8, false),
            Field::new("ogrn_date", DataType::Utf8, true),
            Field::new("inn", DataType::Utf8, false),
            Field::new("kpp", DataType::Utf8, true),
            Field::new("full_name", DataType::Utf8, false),
            Field::new("short_name", DataType::Utf8, true),
            Field::new("status", DataType::Utf8, true),
            Field::new("status_code", DataType::Utf8, true),
            Field::new("registration_date", DataType::Utf8, true),
            Field::new("termination_date", DataType::Utf8, true),
            Field::new("address", DataType::Utf8, true),
            Field::new("region_code", DataType::Utf8, true),
            Field::new("region", DataType::Utf8, true),
            Field::new("capital_amount", DataType::Float64, true),
            Field::new("capital_currency", DataType::Utf8, true),
            Field::new("head_name", DataType::Utf8, true),
            Field::new("head_position", DataType::Utf8, true),
            Field::new("main_activity_code", DataType::Utf8, true),
            Field::new("main_activity_name", DataType::Utf8, true),
            Field::new("additional_activities", DataType::Utf8, true), // JSON массив
            Field::new("email", DataType::Utf8, true),
            Field::new("founders_count", DataType::Int32, true),
            Field::new("extract_date", DataType::Utf8, true),
        ])
    }

    /// Схема для ЕГРИП
    fn egrip_schema() -> Schema {
        Schema::new(vec![
            Field::new("ogrnip", DataType::Utf8, false),
            Field::new("ogrnip_date", DataType::Utf8, true),
            Field::new("inn", DataType::Utf8, false),
            Field::new("last_name", DataType::Utf8, false),
            Field::new("first_name", DataType::Utf8, false),
            Field::new("middle_name", DataType::Utf8, true),
            Field::new("full_name", DataType::Utf8, false),
            Field::new("gender", DataType::Utf8, true),
            Field::new("citizenship", DataType::Utf8, true),
            Field::new("status", DataType::Utf8, true),
            Field::new("status_code", DataType::Utf8, true),
            Field::new("registration_date", DataType::Utf8, true),
            Field::new("termination_date", DataType::Utf8, true),
            Field::new("region_code", DataType::Utf8, true),
            Field::new("region", DataType::Utf8, true),
            Field::new("main_activity_code", DataType::Utf8, true),
            Field::new("main_activity_name", DataType::Utf8, true),
            Field::new("additional_activities", DataType::Utf8, true), // JSON массив
            Field::new("email", DataType::Utf8, true),
            Field::new("extract_date", DataType::Utf8, true),
        ])
    }

    /// Запись батча ЕГРЮЛ
    pub fn write_egrul_batch(&mut self, records: &[EgrulRecord]) -> Result<()> {
        if records.is_empty() {
            return Ok(());
        }

        let schema = Arc::new(Self::egrul_schema());

        // Создаем массивы для каждого поля
        let ogrn: StringArray = records.iter().map(|r| Some(r.ogrn.as_str())).collect();
        let ogrn_date: StringArray = records.iter()
            .map(|r| r.ogrn_date.map(|d| d.to_string()))
            .collect();
        let inn: StringArray = records.iter().map(|r| Some(r.inn.as_str())).collect();
        let kpp: StringArray = records.iter()
            .map(|r| r.kpp.as_deref())
            .collect();
        let full_name: StringArray = records.iter().map(|r| Some(r.full_name.as_str())).collect();
        let short_name: StringArray = records.iter()
            .map(|r| r.short_name.as_deref())
            .collect();
        let status: StringArray = records.iter()
            .map(|r| Some(r.status.as_code()))
            .collect();
        let status_code: StringArray = records.iter()
            .map(|r| r.status_code.as_deref())
            .collect();
        let registration_date: StringArray = records.iter()
            .map(|r| r.registration_date.map(|d| d.to_string()))
            .collect();
        let termination_date: StringArray = records.iter()
            .map(|r| r.termination_date.map(|d| d.to_string()))
            .collect();
        let address: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.full_address.as_deref()))
            .collect();
        let region_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region_code.as_deref()))
            .collect();
        let region: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region.as_deref()))
            .collect();
        let capital_amount: Float64Array = records.iter()
            .map(|r| r.capital.as_ref().map(|c| c.amount))
            .collect();
        let capital_currency: StringArray = records.iter()
            .map(|r| r.capital.as_ref().map(|c| c.currency.as_str()))
            .collect();
        let head_name: StringArray = records.iter()
            .map(|r| r.head.as_ref().map(|h| h.person.full_name()))
            .collect();
        let head_position: StringArray = records.iter()
            .map(|r| r.head.as_ref().and_then(|h| h.position.as_deref()))
            .collect();
        let main_activity_code: StringArray = records.iter()
            .map(|r| r.main_activity.as_ref().map(|a| a.code.as_str()))
            .collect();
        let main_activity_name: StringArray = records.iter()
            .map(|r| r.main_activity.as_ref().map(|a| a.name.as_str()))
            .collect();
        let additional_activities: StringArray = records.iter()
            .map(|r| {
                if r.additional_activities.is_empty() {
                    None
                } else {
                    serde_json::to_string(&r.additional_activities).ok()
                }
            })
            .collect();
        let email: StringArray = records.iter()
            .map(|r| r.email.as_deref())
            .collect();
        let founders_count: Int32Array = records.iter()
            .map(|r| Some(r.founders.len() as i32))
            .collect();
        let extract_date: StringArray = records.iter()
            .map(|r| r.extract_date.map(|d| d.to_string()))
            .collect();

        let batch = RecordBatch::try_new(
            schema,
            vec![
                Arc::new(ogrn),
                Arc::new(ogrn_date),
                Arc::new(inn),
                Arc::new(kpp),
                Arc::new(full_name),
                Arc::new(short_name),
                Arc::new(status),
                Arc::new(status_code),
                Arc::new(registration_date),
                Arc::new(termination_date),
                Arc::new(address),
                Arc::new(region_code),
                Arc::new(region),
                Arc::new(capital_amount),
                Arc::new(capital_currency),
                Arc::new(head_name),
                Arc::new(head_position),
                Arc::new(main_activity_code),
                Arc::new(main_activity_name),
                Arc::new(additional_activities),
                Arc::new(email),
                Arc::new(founders_count),
                Arc::new(extract_date),
            ],
        )?;

        self.egrul_batches.push(batch);
        Ok(())
    }

    /// Запись батча ЕГРИП
    pub fn write_egrip_batch(&mut self, records: &[EgripRecord]) -> Result<()> {
        if records.is_empty() {
            return Ok(());
        }

        let schema = Arc::new(Self::egrip_schema());

        // Создаем массивы для каждого поля
        let ogrnip: StringArray = records.iter().map(|r| Some(r.ogrnip.as_str())).collect();
        let ogrnip_date: StringArray = records.iter()
            .map(|r| r.ogrnip_date.map(|d| d.to_string()))
            .collect();
        let inn: StringArray = records.iter().map(|r| Some(r.inn.as_str())).collect();
        let last_name: StringArray = records.iter()
            .map(|r| Some(r.person.last_name.as_str()))
            .collect();
        let first_name: StringArray = records.iter()
            .map(|r| Some(r.person.first_name.as_str()))
            .collect();
        let middle_name: StringArray = records.iter()
            .map(|r| r.person.middle_name.as_deref())
            .collect();
        let full_name: StringArray = records.iter()
            .map(|r| Some(r.full_name()))
            .collect();
        let gender: StringArray = records.iter()
            .map(|r| r.gender.map(|g| g.to_string()))
            .collect();
        let citizenship: StringArray = records.iter()
            .map(|r| r.citizenship.as_ref().and_then(|c| c.country_name.as_deref()))
            .collect();
        let status: StringArray = records.iter()
            .map(|r| Some(r.status.as_code()))
            .collect();
        let status_code: StringArray = records.iter()
            .map(|r| r.status_code.as_deref())
            .collect();
        let registration_date: StringArray = records.iter()
            .map(|r| r.registration_date.map(|d| d.to_string()))
            .collect();
        let termination_date: StringArray = records.iter()
            .map(|r| r.termination_date.map(|d| d.to_string()))
            .collect();
        let region_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region_code.as_deref()))
            .collect();
        let region: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region.as_deref()))
            .collect();
        let main_activity_code: StringArray = records.iter()
            .map(|r| r.main_activity.as_ref().map(|a| a.code.as_str()))
            .collect();
        let main_activity_name: StringArray = records.iter()
            .map(|r| r.main_activity.as_ref().map(|a| a.name.as_str()))
            .collect();
        let additional_activities: StringArray = records.iter()
            .map(|r| {
                if r.additional_activities.is_empty() {
                    None
                } else {
                    serde_json::to_string(&r.additional_activities).ok()
                }
            })
            .collect();
        let email: StringArray = records.iter()
            .map(|r| r.email.as_deref())
            .collect();
        let extract_date: StringArray = records.iter()
            .map(|r| r.extract_date.map(|d| d.to_string()))
            .collect();

        let batch = RecordBatch::try_new(
            schema,
            vec![
                Arc::new(ogrnip),
                Arc::new(ogrnip_date),
                Arc::new(inn),
                Arc::new(last_name),
                Arc::new(first_name),
                Arc::new(middle_name),
                Arc::new(full_name),
                Arc::new(gender),
                Arc::new(citizenship),
                Arc::new(status),
                Arc::new(status_code),
                Arc::new(registration_date),
                Arc::new(termination_date),
                 Arc::new(region_code),
                Arc::new(region),
                Arc::new(main_activity_code),
                Arc::new(main_activity_name),
                Arc::new(additional_activities),
                Arc::new(email),
                Arc::new(extract_date),
            ],
        )?;

        self.egrip_batches.push(batch);
        Ok(())
    }

    /// Завершение записи
    pub fn finish(self) -> Result<()> {
        // Записываем ЕГРЮЛ если есть данные
        if !self.egrul_batches.is_empty() {
            let egrul_path = self.path.with_file_name(
                format!("{}_egrul.parquet", 
                    self.path.file_stem().unwrap_or_default().to_string_lossy())
            );
            self.write_parquet_file(&egrul_path, &self.egrul_batches)?;
            debug!("Записан файл ЕГРЮЛ: {:?}", egrul_path);
        }

        // Записываем ЕГРИП если есть данные
        if !self.egrip_batches.is_empty() {
            let egrip_path = self.path.with_file_name(
                format!("{}_egrip.parquet",
                    self.path.file_stem().unwrap_or_default().to_string_lossy())
            );
            self.write_parquet_file(&egrip_path, &self.egrip_batches)?;
            debug!("Записан файл ЕГРИП: {:?}", egrip_path);
        }

        Ok(())
    }

    /// Запись Parquet файла
    fn write_parquet_file(&self, path: &Path, batches: &[RecordBatch]) -> Result<()> {
        if batches.is_empty() {
            return Ok(());
        }

        let file = File::create(path)?;
        
        let props = WriterProperties::builder()
            .set_compression(Compression::SNAPPY)
            .set_writer_version(parquet::file::properties::WriterVersion::PARQUET_2_0)
            .build();

        let schema = batches[0].schema();
        let mut writer = ArrowWriter::try_new(file, schema, Some(props))?;

        for batch in batches {
            writer.write(batch)?;
        }

        writer.close()?;
        Ok(())
    }
}

