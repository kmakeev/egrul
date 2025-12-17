//! Модели данных ЕГРЮЛ (юридические лица)

use chrono::NaiveDate;
use serde::{Deserialize, Serialize};

use super::common::*;

/// Запись ЕГРЮЛ — юридическое лицо
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct EgrulRecord {
    // === Основные идентификаторы ===
    /// ОГРН — основной государственный регистрационный номер
    pub ogrn: String,
    /// Дата присвоения ОГРН
    pub ogrn_date: Option<NaiveDate>,
    /// ИНН — идентификационный номер налогоплательщика
    pub inn: String,
    /// КПП — код причины постановки на учет
    pub kpp: Option<String>,

    // === Наименование ===
    /// Полное наименование
    pub full_name: String,
    /// Сокращенное наименование
    pub short_name: Option<String>,
    /// Фирменное наименование
    pub brand_name: Option<String>,

    // === Организационно-правовая форма ===
    /// Код ОПФ
    pub opf_code: Option<String>,
    /// Наименование ОПФ
    pub opf_name: Option<String>,

    // === Статус и даты ===
    /// Статус юридического лица
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
    /// Юридический адрес
    pub address: Option<Address>,
    /// Адрес электронной почты
    pub email: Option<String>,

    // === Уставный капитал ===
    /// Уставный капитал
    pub capital: Option<Capital>,

    // === Руководитель ===
    /// Сведения о руководителе
    pub head: Option<HeadInfo>,

    // === Учредители ===
    /// Учредители
    pub founders: Vec<Founder>,

    // === Виды деятельности ===
    /// Основной вид деятельности
    pub main_activity: Option<Activity>,
    /// Дополнительные виды деятельности
    pub additional_activities: Vec<Activity>,

    // === Регистрация ===
    /// Сведения о регистрации
    pub registration: Option<RegistrationInfo>,
    /// Сведения о налоговом органе
    pub tax_authority: Option<TaxAuthority>,
    /// Сведения о ПФР
    pub pension_fund: Option<PensionFund>,
    /// Сведения о ФСС
    pub social_insurance: Option<SocialInsurance>,

    // === Дополнительные сведения ===
    /// Сведения о лицензиях
    pub licenses: Vec<License>,
    /// Сведения о филиалах и представительствах
    pub branches: Vec<Branch>,
    /// История изменений
    pub history: Vec<HistoryRecord>,

    // === Специальные статусы ===
    /// Сведения о банкротстве
    pub bankruptcy: Option<BankruptcyInfo>,
    /// Сведения о реорганизации
    pub reorganization: Option<ReorganizationInfo>,
    /// Сведения о ликвидации
    pub liquidation: Option<LiquidationInfo>,

    // === Метаданные ===
    /// Идентификатор документа в выписке
    pub document_id: Option<String>,
    /// Источник данных (имя файла)
    pub source_file: Option<String>,
}

/// Сведения о руководителе
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct HeadInfo {
    /// ФИО руководителя
    pub person: Person,
    /// Должность
    pub position: Option<String>,
    /// Код должности
    pub position_code: Option<String>,
    /// Дата вступления в должность
    pub start_date: Option<NaiveDate>,
    /// ГРН записи
    pub grn: Option<String>,
    /// Дата ГРН
    pub grn_date: Option<NaiveDate>,
    /// Признак дисквалификации
    pub is_disqualified: bool,
}

/// Сведения о регистрации
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct RegistrationInfo {
    /// Способ образования
    pub formation_method: Option<String>,
    /// Код способа образования
    pub formation_code: Option<String>,
    /// Регистрирующий орган
    pub authority: Option<RegistrationAuthority>,
    /// Регистрационный номер до 01.07.2002
    pub old_reg_number: Option<String>,
    /// Дата регистрации до 01.07.2002
    pub old_reg_date: Option<NaiveDate>,
    /// Наименование органа, зарегистрировавшего до 01.07.2002
    pub old_authority: Option<String>,
    /// Серия и номер свидетельства
    pub certificate_series: Option<String>,
    pub certificate_number: Option<String>,
    pub certificate_date: Option<NaiveDate>,
}

/// Филиал или представительство
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Branch {
    /// Тип: филиал или представительство
    pub branch_type: BranchType,
    /// Наименование
    pub name: Option<String>,
    /// Адрес
    pub address: Option<Address>,
    /// КПП
    pub kpp: Option<String>,
    /// ГРН записи
    pub grn: Option<String>,
    /// Дата ГРН
    pub grn_date: Option<NaiveDate>,
}

/// Тип структурного подразделения
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Serialize, Deserialize)]
pub enum BranchType {
    #[default]
    Branch,         // Филиал
    Representative, // Представительство
}

/// Сведения о банкротстве
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct BankruptcyInfo {
    /// Стадия банкротства
    pub stage: Option<String>,
    /// Код стадии
    pub stage_code: Option<String>,
    /// Дата введения стадии
    pub stage_date: Option<NaiveDate>,
    /// Арбитражный управляющий
    pub manager: Option<Person>,
    /// Наименование СРО арбитражных управляющих
    pub sro_name: Option<String>,
}

/// Сведения о реорганизации
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ReorganizationInfo {
    /// Форма реорганизации
    pub form: Option<String>,
    /// Код формы
    pub form_code: Option<String>,
    /// Правопреемники
    pub successors: Vec<SuccessorInfo>,
    /// Правопредшественники
    pub predecessors: Vec<PredecessorInfo>,
}

/// Сведения о правопреемнике
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SuccessorInfo {
    pub ogrn: String,
    pub inn: Option<String>,
    pub name: Option<String>,
}

/// Сведения о правопредшественнике
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PredecessorInfo {
    pub ogrn: String,
    pub inn: Option<String>,
    pub name: Option<String>,
}

/// Сведения о ликвидации
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct LiquidationInfo {
    /// Способ прекращения
    pub method: Option<String>,
    /// Код способа
    pub method_code: Option<String>,
    /// Дата принятия решения о ликвидации
    pub decision_date: Option<NaiveDate>,
    /// Ликвидационная комиссия / ликвидатор
    pub liquidator: Option<Person>,
    /// Наименование ликвидатора (если юр. лицо)
    pub liquidator_org: Option<String>,
}

impl EgrulRecord {
    pub fn new() -> Self {
        Self::default()
    }

    /// Проверка валидности записи
    pub fn is_valid(&self) -> bool {
        !self.ogrn.is_empty() && !self.inn.is_empty() && !self.full_name.is_empty()
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

    /// Проверка активности юр. лица
    pub fn is_active(&self) -> bool {
        matches!(self.status, EntityStatus::Active)
    }
}

