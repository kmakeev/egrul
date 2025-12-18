//! Общие структуры данных

use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

/// Физическое лицо (ФИО)
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Person {
    /// Фамилия
    pub last_name: String,
    /// Имя
    pub first_name: String,
    /// Отчество
    pub middle_name: Option<String>,
    /// ИНН физического лица
    pub inn: Option<String>,
}

impl Person {
    /// Полное ФИО
    pub fn full_name(&self) -> String {
        match &self.middle_name {
            Some(middle) if !middle.is_empty() => {
                format!("{} {} {}", self.last_name, self.first_name, middle)
            }
            _ => format!("{} {}", self.last_name, self.first_name),
        }
    }

    /// Краткое ФИО (Фамилия И.О.)
    pub fn short_name(&self) -> String {
        let first_initial = self.first_name.chars().next().unwrap_or(' ');
        match &self.middle_name {
            Some(middle) if !middle.is_empty() => {
                let middle_initial = middle.chars().next().unwrap_or(' ');
                format!("{} {}.{}.", self.last_name, first_initial, middle_initial)
            }
            _ => format!("{} {}.", self.last_name, first_initial),
        }
    }

    pub fn is_empty(&self) -> bool {
        self.last_name.is_empty() && self.first_name.is_empty()
    }
}

/// Адрес
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Address {
    /// Индекс
    pub postal_code: Option<String>,
    /// Код региона
    pub region_code: Option<String>,
    /// Наименование региона
    pub region: Option<String>,
    /// Район
    pub district: Option<String>,
    /// Город
    pub city: Option<String>,
    /// Населенный пункт
    pub locality: Option<String>,
    /// Улица
    pub street: Option<String>,
    /// Дом
    pub house: Option<String>,
    /// Корпус
    pub building: Option<String>,
    /// Квартира/Офис
    pub flat: Option<String>,
    /// Полный адрес одной строкой
    pub full_address: Option<String>,
    /// ФИАС код
    pub fias_id: Option<String>,
}

impl Address {
    /// Формирование полного адреса из компонентов
    pub fn build_full_address(&self) -> String {
        let mut parts = Vec::new();

        if let Some(ref code) = self.postal_code {
            if !code.is_empty() {
                parts.push(code.clone());
            }
        }

        if let Some(ref region) = self.region {
            if !region.is_empty() {
                parts.push(region.clone());
            }
        }

        if let Some(ref district) = self.district {
            if !district.is_empty() {
                parts.push(format!("р-н {}", district));
            }
        }

        if let Some(ref city) = self.city {
            if !city.is_empty() {
                parts.push(format!("г. {}", city));
            }
        }

        if let Some(ref locality) = self.locality {
            if !locality.is_empty() && self.city.as_ref() != Some(locality) {
                parts.push(locality.clone());
            }
        }

        if let Some(ref street) = self.street {
            if !street.is_empty() {
                parts.push(format!("ул. {}", street));
            }
        }

        if let Some(ref house) = self.house {
            if !house.is_empty() {
                parts.push(format!("д. {}", house));
            }
        }

        if let Some(ref building) = self.building {
            if !building.is_empty() {
                parts.push(format!("корп. {}", building));
            }
        }

        if let Some(ref flat) = self.flat {
            if !flat.is_empty() {
                parts.push(format!("кв. {}", flat));
            }
        }

        parts.join(", ")
    }

    pub fn is_empty(&self) -> bool {
        self.postal_code.is_none()
            && self.region.is_none()
            && self.city.is_none()
            && self.street.is_none()
            && self.full_address.is_none()
    }
}

/// Вид экономической деятельности (ОКВЭД)
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Activity {
    /// Код ОКВЭД
    pub code: String,
    /// Наименование вида деятельности
    pub name: String,
    /// Версия ОКВЭД (2001, 2014)
    pub version: Option<String>,
    /// Признак основного вида деятельности
    pub is_main: bool,
}

/// Уставный капитал
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Capital {
    /// Размер капитала
    pub amount: f64,
    /// Валюта
    pub currency: String,
}

impl Capital {
    pub fn new(amount: f64, currency: impl Into<String>) -> Self {
        Self {
            amount,
            currency: currency.into(),
        }
    }
}

/// Доля в уставном капитале
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Share {
    /// Номинальная стоимость
    pub nominal_value: Option<f64>,
    /// Размер доли (числитель)
    pub numerator: Option<i64>,
    /// Размер доли (знаменатель)
    pub denominator: Option<i64>,
    /// Процент доли
    pub percent: Option<f64>,
}

impl Share {
    /// Расчет процента из дроби
    pub fn calculate_percent(&self) -> Option<f64> {
        if let (Some(num), Some(den)) = (self.numerator, self.denominator) {
            if den != 0 {
                return Some((num as f64 / den as f64) * 100.0);
            }
        }
        self.percent
    }
}

/// Учредитель (физическое или юридическое лицо)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Founder {
    /// Физическое лицо
    Person {
        person: Person,
        share: Option<Share>,
        #[serde(skip_serializing_if = "Option::is_none")]
        citizenship: Option<String>,
    },
    /// Российское юридическое лицо
    RussianLegalEntity {
        ogrn: String,
        inn: String,
        name: String,
        share: Option<Share>,
    },
    /// Иностранное юридическое лицо
    ForeignLegalEntity {
        name: String,
        country: Option<String>,
        reg_number: Option<String>,
        share: Option<Share>,
    },
    /// Публично-правовое образование (РФ, субъект РФ, муниципальное образование)
    PublicEntity {
        name: String,
        share: Option<Share>,
    },
    /// ПИФ (паевой инвестиционный фонд)
    MutualFund {
        name: String,
        share: Option<Share>,
    },
}

impl Founder {
    pub fn share(&self) -> Option<&Share> {
        match self {
            Founder::Person { share, .. } => share.as_ref(),
            Founder::RussianLegalEntity { share, .. } => share.as_ref(),
            Founder::ForeignLegalEntity { share, .. } => share.as_ref(),
            Founder::PublicEntity { share, .. } => share.as_ref(),
            Founder::MutualFund { share, .. } => share.as_ref(),
        }
    }

    pub fn name(&self) -> String {
        match self {
            Founder::Person { person, .. } => person.full_name(),
            Founder::RussianLegalEntity { name, .. } => name.clone(),
            Founder::ForeignLegalEntity { name, .. } => name.clone(),
            Founder::PublicEntity { name, .. } => name.clone(),
            Founder::MutualFund { name, .. } => name.clone(),
        }
    }
}

/// Запись в истории изменений (ГРН)
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct HistoryRecord {
    /// ГРН (государственный регистрационный номер записи)
    pub grn: String,
    /// Дата записи
    pub date: Option<NaiveDate>,
    /// Код причины внесения записи
    pub reason_code: Option<String>,
    /// Описание причины
    pub reason_description: Option<String>,
    /// Регистрирующий орган (код)
    pub authority_code: Option<String>,
    /// Регистрирующий орган (наименование)
    pub authority_name: Option<String>,
    /// Серия свидетельства
    pub certificate_series: Option<String>,
    /// Номер свидетельства
    pub certificate_number: Option<String>,
    /// Дата выдачи свидетельства
    pub certificate_date: Option<NaiveDate>,
}

/// Сведения о регистрирующем органе
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct RegistrationAuthority {
    /// Код органа
    pub code: String,
    /// Наименование
    pub name: String,
    /// Адрес
    pub address: Option<String>,
}

/// Сведения о налоговом органе
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct TaxAuthority {
    /// Код налогового органа
    pub code: String,
    /// Наименование
    pub name: String,
    /// Дата постановки на учет
    pub registration_date: Option<NaiveDate>,
}

/// Сведения о ПФР
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PensionFund {
    /// Регистрационный номер в ПФР
    pub reg_number: String,
    /// Дата регистрации
    pub registration_date: Option<NaiveDate>,
    /// Наименование органа ПФР
    pub authority_name: Option<String>,
}

/// Сведения о ФСС
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SocialInsurance {
    /// Регистрационный номер в ФСС
    pub reg_number: String,
    /// Дата регистрации
    pub registration_date: Option<NaiveDate>,
    /// Наименование органа ФСС
    pub authority_name: Option<String>,
}

/// Статус юридического лица / ИП
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum EntityStatus {
    /// Действующее
    Active,
    /// Ликвидировано
    Liquidated,
    /// В процессе ликвидации
    Liquidating,
    /// Реорганизовано
    Reorganized,
    /// В процессе реорганизации
    Reorganizing,
    /// Исключено из реестра
    Excluded,
    /// Банкрот
    Bankrupt,
    /// Прекращена деятельность
    Ceased,
    /// Неизвестный статус
    Unknown,
}

impl Default for EntityStatus {
    fn default() -> Self {
        EntityStatus::Unknown
    }
}

impl EntityStatus {
    /// Парсинг статуса из строки
    pub fn from_str(s: &str) -> Self {
        let lower = s.to_lowercase();
        if lower.contains("актив") || lower.contains("действ") {
            EntityStatus::Active
        } else if lower.contains("ликвидир") && lower.contains("процесс") {
            EntityStatus::Liquidating
        } else if lower.contains("ликвидир") {
            EntityStatus::Liquidated
        } else if lower.contains("реорганиз") && lower.contains("процесс") {
            EntityStatus::Reorganizing
        } else if lower.contains("реорганиз") {
            EntityStatus::Reorganized
        } else if lower.contains("исключ") {
            EntityStatus::Excluded
        } else if lower.contains("банкрот") || lower.contains("несостоят") {
            EntityStatus::Bankrupt
        } else if lower.contains("прекращ") {
            EntityStatus::Ceased
        } else {
            EntityStatus::Unknown
        }
    }

    /// Статус из кода
    pub fn from_code(code: &str) -> Self {
        let c = code.trim();
        match c {
            // Обобщённые строковые коды
            "актив" | "1" => EntityStatus::Active,
            "ликвид" | "2" => EntityStatus::Liquidated,
            "реорг" | "3" => EntityStatus::Reorganized,
            "исключ" | "4" => EntityStatus::Excluded,
            "банкрот" | "5" => EntityStatus::Bankrupt,

            // Типичные числовые коды ФНС для ИП (ЕГРИП)
            // 101, 102, 103... — активные/зарегистрированные
            "101" | "102" | "103" | "104" => EntityStatus::Active,
            // 201, 202, 203... — прекращение деятельности ИП
            "201" | "202" | "203" | "211" | "212" | "221" => EntityStatus::Ceased,

            // Часть кодов для ЮЛ (ЕГРЮЛ) по банкротству / ликвидации
            // 114 — наблюдение в деле о банкротстве
            "114" => EntityStatus::Bankrupt,

            _ => EntityStatus::Unknown,
        }
    }

    /// Нормализованный код статуса для БД / Parquet (совпадает с GraphQL/ClickHouse)
    pub fn as_code(self) -> &'static str {
        match self {
            EntityStatus::Active => "active",
            EntityStatus::Liquidated => "liquidated",
            EntityStatus::Liquidating => "liquidating",
            EntityStatus::Reorganized => "reorganized",
            EntityStatus::Reorganizing => "reorganizing",
            EntityStatus::Excluded => "excluded",
            EntityStatus::Bankrupt => "bankrupt",
            EntityStatus::Ceased => "ceased",
            EntityStatus::Unknown => "unknown",
        }
    }
}

impl std::fmt::Display for EntityStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EntityStatus::Active => write!(f, "Действующее"),
            EntityStatus::Liquidated => write!(f, "Ликвидировано"),
            EntityStatus::Liquidating => write!(f, "В процессе ликвидации"),
            EntityStatus::Reorganized => write!(f, "Реорганизовано"),
            EntityStatus::Reorganizing => write!(f, "В процессе реорганизации"),
            EntityStatus::Excluded => write!(f, "Исключено из реестра"),
            EntityStatus::Bankrupt => write!(f, "Банкрот"),
            EntityStatus::Ceased => write!(f, "Прекращена деятельность"),
            EntityStatus::Unknown => write!(f, "Неизвестно"),
        }
    }
}

/// Информация о документе (свидетельство, устав и т.д.)
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Document {
    /// Тип документа
    pub doc_type: String,
    /// Серия
    pub series: Option<String>,
    /// Номер
    pub number: Option<String>,
    /// Дата выдачи
    pub issue_date: Option<NaiveDate>,
    /// Кем выдан
    pub issued_by: Option<String>,
}

/// Лицензия
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct License {
    /// Номер лицензии
    pub number: String,
    /// Серия
    pub series: Option<String>,
    /// Вид деятельности
    pub activity: Option<String>,
    /// Дата начала действия
    pub start_date: Option<NaiveDate>,
    /// Дата окончания действия
    pub end_date: Option<NaiveDate>,
    /// Лицензирующий орган
    pub authority: Option<String>,
    /// Статус лицензии
    pub status: Option<String>,
}

