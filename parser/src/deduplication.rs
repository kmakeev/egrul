//! Модуль дедупликации для автоматического устранения дублей при импорте
//! 
//! ВРЕМЕННО ОТКЛЮЧЕН - требует дополнительных зависимостей

/*
use std::collections::{HashMap, HashSet};
use std::path::Path;
use sha2::{Digest, Sha256};
use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

use crate::models::common::HistoryRecord;
use crate::Result;
*/

// Заглушка для совместимости
pub struct DeduplicationManager;

impl DeduplicationManager {
    pub fn new() -> Self {
        Self
    }
}

impl Default for DeduplicationManager {
    fn default() -> Self {
        Self::new()
    }
}

/*
/// Менеджер дедупликации для записей истории
#[derive(Debug, Default)]
pub struct DeduplicationManager {
    /// Кеш обработанных файлов (хеш файла -> метаданные)
    processed_files: HashMap<String, FileMetadata>,
    /// Кеш уникальных записей истории (ключ дедупликации -> запись)
    unique_records: HashMap<DeduplicationKey, HistoryRecordWithSources>,
    /// Статистика дедупликации
    stats: DeduplicationStats,
}

/// Метаданные обработанного файла
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileMetadata {
    pub file_path: String,
    pub file_hash: String,
    pub extract_date: Option<NaiveDate>,
    pub processed_at: chrono::DateTime<chrono::Utc>,
    pub records_count: usize,
}

/// Ключ для дедупликации записей истории
#[derive(Debug, Clone, Hash, PartialEq, Eq)]
pub struct DeduplicationKey {
    pub entity_type: String,
    pub entity_id: String,
    pub grn: String,
}

/// Запись истории с информацией об источниках
#[derive(Debug, Clone)]
pub struct HistoryRecordWithSources {
    pub record: HistoryRecord,
    pub source_files: Vec<String>,
    pub file_hashes: Vec<String>,
    pub extract_dates: Vec<Option<NaiveDate>>,
    pub first_seen: chrono::DateTime<chrono::Utc>,
    pub last_updated: chrono::DateTime<chrono::Utc>,
}

/// Статистика дедупликации
#[derive(Debug, Default, Clone, Serialize, Deserialize)]
pub struct DeduplicationStats {
    pub total_files_processed: usize,
    pub total_records_processed: usize,
    pub unique_records: usize,
    pub duplicates_removed: usize,
    pub deduplication_ratio: f64,
    pub files_skipped: usize,
}

impl DeduplicationManager {
    /// Создает новый менеджер дедупликации
    pub fn new() -> Self {
        Self::default()
    }

    /// Вычисляет хеш файла
    pub fn calculate_file_hash<P: AsRef<Path>>(file_path: P) -> Result<String> {
        let content = std::fs::read(file_path)?;
        let mut hasher = Sha256::new();
        hasher.update(&content);
        let hash = hasher.finalize();
        Ok(format!("{:x}", hash))
    }

    /// Проверяет, был ли файл уже обработан
    pub fn is_file_processed(&self, file_hash: &str) -> bool {
        self.processed_files.contains_key(file_hash)
    }

    /// Регистрирует обработанный файл
    pub fn register_processed_file(&mut self, metadata: FileMetadata) {
        self.processed_files.insert(metadata.file_hash.clone(), metadata);
        self.stats.total_files_processed += 1;
    }

    /// Добавляет запись истории с автоматической дедупликацией
    pub fn add_history_record(
        &mut self,
        entity_type: String,
        entity_id: String,
        record: HistoryRecord,
        source_file: String,
        file_hash: String,
        extract_date: Option<NaiveDate>,
    ) {
        let key = DeduplicationKey {
            entity_type,
            entity_id,
            grn: record.grn.clone(),
        };

        let now = chrono::Utc::now();

        match self.unique_records.get_mut(&key) {
            Some(existing) => {
                // Запись уже существует - добавляем новый источник
                if !existing.source_files.contains(&source_file) {
                    existing.source_files.push(source_file);
                    existing.file_hashes.push(file_hash);
                    existing.extract_dates.push(extract_date);
                    existing.last_updated = now;
                    self.stats.duplicates_removed += 1;
                }
            }
            None => {
                // Новая уникальная запись
                let record_with_sources = HistoryRecordWithSources {
                    record,
                    source_files: vec![source_file],
                    file_hashes: vec![file_hash],
                    extract_dates: vec![extract_date],
                    first_seen: now,
                    last_updated: now,
                };
                self.unique_records.insert(key, record_with_sources);
                self.stats.unique_records += 1;
            }
        }

        self.stats.total_records_processed += 1;
        self.update_deduplication_ratio();
    }

    /// Получает все уникальные записи для экспорта
    pub fn get_unique_records(&self) -> Vec<&HistoryRecordWithSources> {
        self.unique_records.values().collect()
    }

    /// Получает записи для конкретной сущности
    pub fn get_records_for_entity(&self, entity_type: &str, entity_id: &str) -> Vec<&HistoryRecordWithSources> {
        self.unique_records
            .iter()
            .filter(|(key, _)| key.entity_type == entity_type && key.entity_id == entity_id)
            .map(|(_, record)| record)
            .collect()
    }

    /// Обновляет коэффициент дедупликации
    fn update_deduplication_ratio(&mut self) {
        if self.stats.total_records_processed > 0 {
            self.stats.deduplication_ratio = 
                (self.stats.duplicates_removed as f64 / self.stats.total_records_processed as f64) * 100.0;
        }
    }

    /// Получает статистику дедупликации
    pub fn get_stats(&self) -> &DeduplicationStats {
        &self.stats
    }

    /// Очищает кеш (для освобождения памяти)
    pub fn clear_cache(&mut self) {
        self.unique_records.clear();
        self.processed_files.clear();
    }

    /// Экспортирует уникальные записи в формате для ClickHouse
    pub fn export_for_clickhouse(&self) -> Vec<ClickHouseHistoryRecord> {
        self.unique_records
            .values()
            .map(|record_with_sources| ClickHouseHistoryRecord {
                id: uuid::Uuid::new_v4().to_string(),
                entity_type: "company".to_string(), // TODO: получать из ключа
                entity_id: "".to_string(), // TODO: получать из ключа
                inn: None,
                grn: record_with_sources.record.grn.clone(),
                grn_date: record_with_sources.record.date,
                reason_code: record_with_sources.record.reason_code.clone(),
                reason_description: record_with_sources.record.reason_description.clone(),
                authority_code: record_with_sources.record.authority_code.clone(),
                authority_name: record_with_sources.record.authority_name.clone(),
                certificate_series: record_with_sources.record.certificate_series.clone(),
                certificate_number: record_with_sources.record.certificate_number.clone(),
                certificate_date: record_with_sources.record.certificate_date,
                snapshot_full_name: None,
                snapshot_status: None,
                snapshot_address: None,
                snapshot_json: None,
                source_files: record_with_sources.source_files.clone(),
                extract_date: record_with_sources.extract_dates.first().copied().flatten(),
                file_hash: record_with_sources.file_hashes.first().cloned().unwrap_or_default(),
                created_at: record_with_sources.first_seen,
                updated_at: record_with_sources.last_updated,
            })
            .collect()
    }

    /// Сохраняет статистику в файл
    pub fn save_stats_to_file<P: AsRef<Path>>(&self, path: P) -> Result<()> {
        let json = serde_json::to_string_pretty(&self.stats)?;
        std::fs::write(path, json)?;
        Ok(())
    }

    /// Загружает статистику из файла
    pub fn load_stats_from_file<P: AsRef<Path>>(&mut self, path: P) -> Result<()> {
        let content = std::fs::read_to_string(path)?;
        self.stats = serde_json::from_str(&content)?;
        Ok(())
    }
}

/// Запись истории в формате для ClickHouse
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClickHouseHistoryRecord {
    pub id: String,
    pub entity_type: String,
    pub entity_id: String,
    pub inn: Option<String>,
    pub grn: String,
    pub grn_date: Option<NaiveDate>,
    pub reason_code: Option<String>,
    pub reason_description: Option<String>,
    pub authority_code: Option<String>,
    pub authority_name: Option<String>,
    pub certificate_series: Option<String>,
    pub certificate_number: Option<String>,
    pub certificate_date: Option<NaiveDate>,
    pub snapshot_full_name: Option<String>,
    pub snapshot_status: Option<String>,
    pub snapshot_address: Option<String>,
    pub snapshot_json: Option<String>,
    pub source_files: Vec<String>,
    pub extract_date: Option<NaiveDate>,
    pub file_hash: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

/// Утилиты для работы с дедупликацией
pub mod utils {
    use super::*;

    /// Создает ключ дедупликации из записи
    pub fn create_deduplication_key(
        entity_type: &str,
        entity_id: &str,
        grn: &str,
    ) -> DeduplicationKey {
        DeduplicationKey {
            entity_type: entity_type.to_string(),
            entity_id: entity_id.to_string(),
            grn: grn.to_string(),
        }
    }

    /// Проверяет, являются ли две записи дублями
    pub fn are_records_duplicates(
        record1: &HistoryRecord,
        record2: &HistoryRecord,
    ) -> bool {
        record1.grn == record2.grn
    }

    /// Объединяет информацию из нескольких записей-дублей
    pub fn merge_duplicate_records(
        base: &HistoryRecord,
        duplicate: &HistoryRecord,
    ) -> HistoryRecord {
        let mut merged = base.clone();

        // Заполняем пустые поля из дубля
        if merged.reason_code.is_none() && duplicate.reason_code.is_some() {
            merged.reason_code = duplicate.reason_code.clone();
        }
        if merged.reason_description.is_none() && duplicate.reason_description.is_some() {
            merged.reason_description = duplicate.reason_description.clone();
        }
        if merged.authority_code.is_none() && duplicate.authority_code.is_some() {
            merged.authority_code = duplicate.authority_code.clone();
        }
        if merged.authority_name.is_none() && duplicate.authority_name.is_some() {
            merged.authority_name = duplicate.authority_name.clone();
        }

        merged
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deduplication_manager() {
        let mut manager = DeduplicationManager::new();

        // Создаем тестовую запись
        let record = HistoryRecord {
            grn: "1067425004635".to_string(),
            date: Some(chrono::NaiveDate::from_ymd_opt(2006, 9, 11).unwrap()),
            reason_code: Some("11201".to_string()),
            reason_description: Some("Создание юридического лица".to_string()),
            authority_code: None,
            authority_name: None,
            certificate_series: None,
            certificate_number: None,
            certificate_date: None,
        };

        // Добавляем запись из первого файла
        manager.add_history_record(
            "company".to_string(),
            "1067425004635".to_string(),
            record.clone(),
            "file1.xml".to_string(),
            "hash1".to_string(),
            Some(chrono::NaiveDate::from_ymd_opt(2024, 8, 1).unwrap()),
        );

        // Добавляем ту же запись из второго файла (должна быть дедуплицирована)
        manager.add_history_record(
            "company".to_string(),
            "1067425004635".to_string(),
            record.clone(),
            "file2.xml".to_string(),
            "hash2".to_string(),
            Some(chrono::NaiveDate::from_ymd_opt(2024, 8, 1).unwrap()),
        );

        let stats = manager.get_stats();
        assert_eq!(stats.total_records_processed, 2);
        assert_eq!(stats.unique_records, 1);
        assert_eq!(stats.duplicates_removed, 1);
        assert_eq!(stats.deduplication_ratio, 50.0);

        // Проверяем, что запись содержит оба источника
        let unique_records = manager.get_unique_records();
        assert_eq!(unique_records.len(), 1);
        assert_eq!(unique_records[0].source_files.len(), 2);
        assert!(unique_records[0].source_files.contains(&"file1.xml".to_string()));
        assert!(unique_records[0].source_files.contains(&"file2.xml".to_string()));
    }

    #[test]
    fn test_file_hash_calculation() {
        // Создаем временный файл для тестирования
        let temp_file = std::env::temp_dir().join("test_file.txt");
        std::fs::write(&temp_file, "test content").unwrap();

        let hash1 = DeduplicationManager::calculate_file_hash(&temp_file).unwrap();
        let hash2 = DeduplicationManager::calculate_file_hash(&temp_file).unwrap();

        assert_eq!(hash1, hash2);
        assert!(!hash1.is_empty());

        // Очищаем
        std::fs::remove_file(&temp_file).unwrap();
    }
}
*/