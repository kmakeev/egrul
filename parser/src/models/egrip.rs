//! Модели данных ЕГРИП (индивидуальные предприниматели)

use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

use super::common::*;

/// Запись ЕГРИП — индивидуальный предприниматель
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct EgripRecord {
    // === Основные идентификаторы ===
    /// ОГРНИП — основной государственный регистрационный номер ИП
    pub ogrnip: String,
    /// Дата присвоения ОГРНИП
    pub ogrnip_date: Option<NaiveDate>,
    /// ИНН
    pub inn: String,

    // === Сведения о физическом лице ===
    /// ФИО
    pub person: Person,
    /// Пол
    pub gender: Option<Gender>,
    /// Гражданство
    pub citizenship: Option<CitizenshipInfo>,

    // === Статус и даты ===
    /// Статус ИП
    pub status: EntityStatus,
    /// Код статуса
    pub status_code: Option<String>,
    /// Способ прекращения деятельности
    pub termination_method: Option<String>,
    /// Дата регистрации
    pub registration_date: Option<NaiveDate>,
    /// Дата прекращения деятельности
    pub termination_date: Option<NaiveDate>,
    /// Дата выписки
    pub extract_date: Option<NaiveDate>,

    // === Адрес ===
    /// Адрес места жительства (регион)
    pub address: Option<Address>,
    /// Адрес электронной почты
    pub email: Option<String>,

    // === Виды деятельности ===
    /// Основной вид деятельности
    pub main_activity: Option<Activity>,
    /// Дополнительные виды деятельности
    pub additional_activities: Vec<Activity>,

    // === Регистрация ===
    /// Сведения о регистрации
    pub registration: Option<IpRegistrationInfo>,
    /// Сведения о налоговом органе
    pub tax_authority: Option<TaxAuthority>,
    /// Сведения о ПФР
    pub pension_fund: Option<PensionFund>,
    /// Сведения о ФСС
    pub social_insurance: Option<SocialInsurance>,

    // === Дополнительные сведения ===
    /// Сведения о лицензиях
    pub licenses: Vec<License>,
    /// История изменений
    pub history: Vec<HistoryRecord>,

    // === Специальные статусы ===
    /// Сведения о банкротстве
    pub bankruptcy: Option<IpBankruptcyInfo>,

    // === Метаданные ===
    /// Идентификатор документа в выписке
    pub document_id: Option<String>,
    /// Источник данных (имя файла)
    pub source_file: Option<String>,
}

/// Пол
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Gender {
    Male,
    Female,
    Unknown,
}

impl Gender {
    pub fn from_code(code: &str) -> Self {
        match code {
            "1" | "м" | "М" | "male" => Gender::Male,
            "2" | "ж" | "Ж" | "female" => Gender::Female,
            _ => Gender::Unknown,
        }
    }
}

impl std::fmt::Display for Gender {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Gender::Male => write!(f, "Мужской"),
            Gender::Female => write!(f, "Женский"),
            Gender::Unknown => write!(f, "Не указан"),
        }
    }
}

/// Сведения о гражданстве
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct CitizenshipInfo {
    /// Тип гражданства
    pub citizenship_type: CitizenshipType,
    /// Код страны (ОКСМ)
    pub country_code: Option<String>,
    /// Наименование страны
    pub country_name: Option<String>,
}

/// Тип гражданства
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Serialize, Deserialize)]
pub enum CitizenshipType {
    /// Гражданин РФ
    #[default]
    Russian,
    /// Иностранный гражданин
    Foreign,
    /// Лицо без гражданства
    Stateless,
}

impl CitizenshipType {
    pub fn from_code(code: &str) -> Self {
        match code {
            "1" => CitizenshipType::Russian,
            "2" => CitizenshipType::Foreign,
            "3" => CitizenshipType::Stateless,
            _ => CitizenshipType::Russian,
        }
    }
}

/// Сведения о регистрации ИП
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct IpRegistrationInfo {
    /// Способ образования
    pub formation_method: Option<String>,
    /// Код способа образования
    pub formation_code: Option<String>,
    /// Регистрирующий орган
    pub authority: Option<RegistrationAuthority>,
    /// Серия свидетельства
    pub certificate_series: Option<String>,
    /// Номер свидетельства
    pub certificate_number: Option<String>,
    /// Дата выдачи свидетельства
    pub certificate_date: Option<NaiveDate>,
}

/// Сведения о банкротстве ИП
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct IpBankruptcyInfo {
    /// Признак банкротства
    pub is_bankrupt: bool,
    /// Дата признания банкротом
    pub bankruptcy_date: Option<NaiveDate>,
    /// Номер дела о банкротстве
    pub case_number: Option<String>,
    /// Арбитражный суд
    pub court: Option<String>,
    /// Арбитражный управляющий
    pub manager: Option<Person>,
}

impl EgripRecord {
    pub fn new() -> Self {
        Self::default()
    }

    /// Проверка валидности записи
    pub fn is_valid(&self) -> bool {
        !self.ogrnip.is_empty() && !self.inn.is_empty() && !self.person.is_empty()
    }

    /// Полное имя ИП
    pub fn full_name(&self) -> String {
        format!("ИП {}", self.person.full_name())
    }

    /// Краткое имя ИП
    pub fn short_name(&self) -> String {
        format!("ИП {}", self.person.short_name())
    }

    /// Получение всех видов деятельности
    pub fn all_activities(&self) -> Vec<&Activity> {
        let mut activities = Vec::new();
        if let Some(ref main) = self.main_activity {
            activities.push(main);
        }
        activities.extend(self.additional_activities.iter());
        activities
    }

    /// Проверка активности ИП
    pub fn is_active(&self) -> bool {
        matches!(self.status, EntityStatus::Active)
    }
}

