//! Parquet писатель для ЕГРЮЛ/ЕГРИП с поддержкой разбивки файлов

use std::path::Path;
use std::sync::Arc;
use std::fs::File;

use arrow::array::*;
use arrow::datatypes::{DataType, Field, Schema};
use arrow::record_batch::RecordBatch;
use parquet::arrow::ArrowWriter;
use parquet::basic::Compression;
use parquet::file::properties::WriterProperties;
use tracing::info;

use crate::error::Result;
use crate::models::{EgrulRecord, EgripRecord};

/// Parquet писатель с поддержкой разбивки файлов
pub struct ParquetOutputWriter {
    base_path: std::path::PathBuf,
    egrul_batches: Vec<RecordBatch>,
    egrip_batches: Vec<RecordBatch>,
    egrul_records_count: usize,
    egrip_records_count: usize,
    egrul_file_index: usize,
    egrip_file_index: usize,
    max_file_size_mb: Option<usize>,
    max_records_per_file: Option<usize>,
}

impl ParquetOutputWriter {
    /// Создание нового писателя
    pub fn new(path: &Path) -> Result<Self> {
        Ok(Self {
            base_path: path.to_path_buf(),
            egrul_batches: Vec::new(),
            egrip_batches: Vec::new(),
            egrul_records_count: 0,
            egrip_records_count: 0,
            egrul_file_index: 0,
            egrip_file_index: 0,
            max_file_size_mb: None,
            max_records_per_file: None,
        })
    }

    /// Создание писателя с ограничениями на размер файлов
    pub fn with_limits(
        path: &Path,
        max_file_size_mb: Option<usize>,
        max_records_per_file: Option<usize>,
    ) -> Result<Self> {
        Ok(Self {
            base_path: path.to_path_buf(),
            egrul_batches: Vec::new(),
            egrip_batches: Vec::new(),
            egrul_records_count: 0,
            egrip_records_count: 0,
            egrul_file_index: 0,
            egrip_file_index: 0,
            max_file_size_mb,
            max_records_per_file,
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
            // Адрес - все поля
            Field::new("postal_code", DataType::Utf8, true),
            Field::new("region_code", DataType::Utf8, true),
            Field::new("region", DataType::Utf8, true),
            Field::new("district", DataType::Utf8, true),
            Field::new("city", DataType::Utf8, true),
            Field::new("locality", DataType::Utf8, true),
            Field::new("street", DataType::Utf8, true),
            Field::new("house", DataType::Utf8, true),
            Field::new("building", DataType::Utf8, true),
            Field::new("flat", DataType::Utf8, true),
            Field::new("full_address", DataType::Utf8, true),
            Field::new("fias_id", DataType::Utf8, true),
            Field::new("kladr_code", DataType::Utf8, true),
            Field::new("capital_amount", DataType::Float64, true),
            Field::new("capital_currency", DataType::Utf8, true),
            Field::new("head_name", DataType::Utf8, true),
            Field::new("head_inn", DataType::Utf8, true),
            Field::new("head_middle_name", DataType::Utf8, true),
            Field::new("head_position", DataType::Utf8, true),
            Field::new("main_activity_code", DataType::Utf8, true),
            Field::new("main_activity_name", DataType::Utf8, true),
            Field::new("additional_activities", DataType::Utf8, true), // JSON массив
            Field::new("email", DataType::Utf8, true),
            Field::new("founders_count", DataType::Int32, true),
            Field::new("founders", DataType::Utf8, true), // JSON массив учредителей
            Field::new("history", DataType::Utf8, true), // JSON массив истории изменений
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
            // Адрес - все поля
            Field::new("postal_code", DataType::Utf8, true),
            Field::new("region_code", DataType::Utf8, true),
            Field::new("region", DataType::Utf8, true),
            Field::new("district", DataType::Utf8, true),
            Field::new("city", DataType::Utf8, true),
            Field::new("locality", DataType::Utf8, true),
            Field::new("street", DataType::Utf8, true),
            Field::new("house", DataType::Utf8, true),
            Field::new("building", DataType::Utf8, true),
            Field::new("flat", DataType::Utf8, true),
            Field::new("full_address", DataType::Utf8, true),
            Field::new("fias_id", DataType::Utf8, true),
            Field::new("kladr_code", DataType::Utf8, true),
            Field::new("main_activity_code", DataType::Utf8, true),
            Field::new("main_activity_name", DataType::Utf8, true),
            Field::new("additional_activities", DataType::Utf8, true), // JSON массив
            Field::new("email", DataType::Utf8, true),
            Field::new("history", DataType::Utf8, true), // JSON массив истории изменений
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
        
        // Все поля адреса
        let postal_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.postal_code.as_deref()))
            .collect();
        let region_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region_code.as_deref()))
            .collect();
        let region: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region.as_deref()))
            .collect();
        let district: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.district.as_deref()))
            .collect();
        let city: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.city.as_deref()))
            .collect();
        let locality: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.locality.as_deref()))
            .collect();
        let street: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.street.as_deref()))
            .collect();
        let house: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.house.as_deref()))
            .collect();
        let building: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.building.as_deref()))
            .collect();
        let flat: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.flat.as_deref()))
            .collect();
        let full_address: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.full_address.as_deref()))
            .collect();
        let fias_id: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.fias_id.as_deref()))
            .collect();
        let kladr_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.kladr_code.as_deref()))
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
        let head_inn: StringArray = records.iter()
            .map(|r| r.head.as_ref().and_then(|h| h.person.inn.as_deref()))
            .collect();
        let head_middle_name: StringArray = records.iter()
            .map(|r| r.head.as_ref().and_then(|h| h.person.middle_name.as_deref()))
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
        let founders: StringArray = records.iter()
            .map(|r| {
                if r.founders.is_empty() {
                    None
                } else {
                    serde_json::to_string(&r.founders).ok()
                }
            })
            .collect();
        let history: StringArray = records.iter()
            .map(|r| {
                if r.history.is_empty() {
                    None
                } else {
                    serde_json::to_string(&r.history).ok()
                }
            })
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
                // Все поля адреса
                Arc::new(postal_code),
                Arc::new(region_code),
                Arc::new(region),
                Arc::new(district),
                Arc::new(city),
                Arc::new(locality),
                Arc::new(street),
                Arc::new(house),
                Arc::new(building),
                Arc::new(flat),
                Arc::new(full_address),
                Arc::new(fias_id),
                Arc::new(kladr_code),
                Arc::new(capital_amount),
                Arc::new(capital_currency),
                Arc::new(head_name),
                Arc::new(head_inn),
                Arc::new(head_middle_name),
                Arc::new(head_position),
                Arc::new(main_activity_code),
                Arc::new(main_activity_name),
                Arc::new(additional_activities),
                Arc::new(email),
                Arc::new(founders_count),
                Arc::new(founders),
                Arc::new(history),
                Arc::new(extract_date),
            ],
        )?;

        self.egrul_batches.push(batch);
        self.egrul_records_count += records.len();

        // Проверяем, нужно ли сохранить файл
        if self.should_flush_egrul() {
            self.flush_egrul_file()?;
        }

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
        
        // Все поля адреса
        let postal_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.postal_code.as_deref()))
            .collect();
        let region_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region_code.as_deref()))
            .collect();
        let region: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.region.as_deref()))
            .collect();
        let district: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.district.as_deref()))
            .collect();
        let city: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.city.as_deref()))
            .collect();
        let locality: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.locality.as_deref()))
            .collect();
        let street: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.street.as_deref()))
            .collect();
        let house: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.house.as_deref()))
            .collect();
        let building: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.building.as_deref()))
            .collect();
        let flat: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.flat.as_deref()))
            .collect();
        let full_address: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.full_address.as_deref()))
            .collect();
        let fias_id: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.fias_id.as_deref()))
            .collect();
        let kladr_code: StringArray = records.iter()
            .map(|r| r.address.as_ref().and_then(|a| a.kladr_code.as_deref()))
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
        let history: StringArray = records.iter()
            .map(|r| {
                if r.history.is_empty() {
                    None
                } else {
                    serde_json::to_string(&r.history).ok()
                }
            })
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
                // Все поля адреса
                Arc::new(postal_code),
                Arc::new(region_code),
                Arc::new(region),
                Arc::new(district),
                Arc::new(city),
                Arc::new(locality),
                Arc::new(street),
                Arc::new(house),
                Arc::new(building),
                Arc::new(flat),
                Arc::new(full_address),
                Arc::new(fias_id),
                Arc::new(kladr_code),
                Arc::new(main_activity_code),
                Arc::new(main_activity_name),
                Arc::new(additional_activities),
                Arc::new(email),
                Arc::new(history),
                Arc::new(extract_date),
            ],
        )?;

        self.egrip_batches.push(batch);
        self.egrip_records_count += records.len();

        // Проверяем, нужно ли сохранить файл
        if self.should_flush_egrip() {
            self.flush_egrip_file()?;
        }

        Ok(())
    }

    /// Проверка, нужно ли сохранить файл ЕГРЮЛ
    fn should_flush_egrul(&self) -> bool {
        if let Some(max_records) = self.max_records_per_file {
            if self.egrul_records_count >= max_records {
                return true;
            }
        }

        if let Some(max_size_mb) = self.max_file_size_mb {
            // Приблизительная оценка размера в памяти
            let estimated_size_mb = self.estimate_egrul_size_mb();
            if estimated_size_mb >= max_size_mb {
                return true;
            }
        }

        false
    }

    /// Проверка, нужно ли сохранить файл ЕГРИП
    fn should_flush_egrip(&self) -> bool {
        if let Some(max_records) = self.max_records_per_file {
            if self.egrip_records_count >= max_records {
                return true;
            }
        }

        if let Some(max_size_mb) = self.max_file_size_mb {
            // Приблизительная оценка размера в памяти
            let estimated_size_mb = self.estimate_egrip_size_mb();
            if estimated_size_mb >= max_size_mb {
                return true;
            }
        }

        false
    }

    /// Приблизительная оценка размера ЕГРЮЛ в МБ
    fn estimate_egrul_size_mb(&self) -> usize {
        // Скорректированная оценка с учетом сжатия Parquet Snappy
        // Реальные тесты показали: ~20KB на запись в несжатом виде, ~4KB после сжатия
        // Используем 18KB для более точной оценки (учитывая overhead и вариативность данных)
        let estimated_bytes = self.egrul_records_count * 18432; // ~18KB на запись
        estimated_bytes / (1024 * 1024)
    }

    /// Приблизительная оценка размера ЕГРИП в МБ
    fn estimate_egrip_size_mb(&self) -> usize {
        // Скорректированная оценка с учетом сжатия Parquet Snappy
        // ЕГРИП более компактный, используем 8KB на запись
        let estimated_bytes = self.egrip_records_count * 8192; // ~8KB на запись
        estimated_bytes / (1024 * 1024)
    }

    /// Сохранение файла ЕГРЮЛ
    fn flush_egrul_file(&mut self) -> Result<()> {
        if self.egrul_batches.is_empty() {
            return Ok(());
        }

        let file_path = if self.egrul_file_index == 0 {
            self.base_path.with_file_name(
                format!("{}_egrul.parquet", 
                    self.base_path.file_stem().unwrap_or_default().to_string_lossy())
            )
        } else {
            self.base_path.with_file_name(
                format!("{}_egrul_part_{:03}.parquet", 
                    self.base_path.file_stem().unwrap_or_default().to_string_lossy(),
                    self.egrul_file_index + 1)
            )
        };

        self.write_parquet_file(&file_path, &self.egrul_batches)?;
        
        // Показываем информацию только при создании дополнительных файлов
        if self.egrul_file_index > 0 {
            info!("Создан файл ЕГРЮЛ: {:?} ({} записей)", 
                  file_path.file_name().unwrap_or_default(), 
                  self.egrul_records_count);
        }

        // Очищаем батчи и увеличиваем индекс
        self.egrul_batches.clear();
        self.egrul_records_count = 0;
        self.egrul_file_index += 1;

        Ok(())
    }

    /// Сохранение файла ЕГРИП
    fn flush_egrip_file(&mut self) -> Result<()> {
        if self.egrip_batches.is_empty() {
            return Ok(());
        }

        let file_path = if self.egrip_file_index == 0 {
            self.base_path.with_file_name(
                format!("{}_egrip.parquet",
                    self.base_path.file_stem().unwrap_or_default().to_string_lossy())
            )
        } else {
            self.base_path.with_file_name(
                format!("{}_egrip_part_{:03}.parquet",
                    self.base_path.file_stem().unwrap_or_default().to_string_lossy(),
                    self.egrip_file_index + 1)
            )
        };

        self.write_parquet_file(&file_path, &self.egrip_batches)?;
        
        // Показываем информацию только при создании дополнительных файлов
        if self.egrip_file_index > 0 {
            info!("Создан файл ЕГРИП: {:?} ({} записей)", 
                  file_path.file_name().unwrap_or_default(), 
                  self.egrip_records_count);
        }

        // Очищаем батчи и увеличиваем индекс
        self.egrip_batches.clear();
        self.egrip_records_count = 0;
        self.egrip_file_index += 1;

        Ok(())
    }

    /// Завершение записи
    pub fn finish(mut self) -> Result<()> {
        // Записываем оставшиеся данные ЕГРЮЛ если есть
        if !self.egrul_batches.is_empty() {
            self.flush_egrul_file()?;
        }

        // Записываем оставшиеся данные ЕГРИП если есть
        if !self.egrip_batches.is_empty() {
            self.flush_egrip_file()?;
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

