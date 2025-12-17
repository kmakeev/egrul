//! Модели данных для ЕГРЮЛ/ЕГРИП

mod common;
pub mod egrul;
pub mod egrip;

pub use common::*;
pub use egrul::*;
pub use egrip::*;

/// Тип реестра
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RegistryType {
    /// ЕГРЮЛ — Единый государственный реестр юридических лиц
    Egrul,
    /// ЕГРИП — Единый государственный реестр индивидуальных предпринимателей
    Egrip,
}

impl RegistryType {
    /// Определение типа реестра по имени файла
    pub fn from_filename(filename: &str) -> Option<Self> {
        let upper = filename.to_uppercase();
        // RUGFO = Реестр Юридических лиц ФНС = ЕГРЮЛ
        // RIGFO = Реестр Индивидуальных предпринимателей ФНС = ЕГРИП
        if upper.contains("RUGFO") || upper.contains("EGRUL") {
            Some(RegistryType::Egrul)
        } else if upper.contains("RIGFO") || upper.contains("EGRIP") {
            Some(RegistryType::Egrip)
        } else {
            None
        }
    }

    /// Определение типа реестра по содержимому XML
    pub fn from_content(content: &str) -> Option<Self> {
        if content.contains("СвЮЛ") || content.contains("RUGFO") {
            Some(RegistryType::Egrul)
        } else if content.contains("СвИП") || content.contains("RIGFO") {
            Some(RegistryType::Egrip)
        } else {
            None
        }
    }
}

impl std::fmt::Display for RegistryType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            RegistryType::Egrul => write!(f, "ЕГРЮЛ"),
            RegistryType::Egrip => write!(f, "ЕГРИП"),
        }
    }
}

/// Унифицированная запись реестра
#[derive(Debug, Clone)]
pub enum RegistryRecord {
    Egrul(EgrulRecord),
    Egrip(EgripRecord),
}

impl RegistryRecord {
    pub fn registry_type(&self) -> RegistryType {
        match self {
            RegistryRecord::Egrul(_) => RegistryType::Egrul,
            RegistryRecord::Egrip(_) => RegistryType::Egrip,
        }
    }

    pub fn inn(&self) -> &str {
        match self {
            RegistryRecord::Egrul(r) => &r.inn,
            RegistryRecord::Egrip(r) => &r.inn,
        }
    }

    pub fn ogrn(&self) -> &str {
        match self {
            RegistryRecord::Egrul(r) => &r.ogrn,
            RegistryRecord::Egrip(r) => &r.ogrnip,
        }
    }
}

