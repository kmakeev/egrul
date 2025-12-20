//! Утилиты для работы с XML атрибутами

use chrono::NaiveDate;
use quick_xml::events::BytesStart;
use quick_xml::name::QName;

use crate::error::Result;

/// Расширение для работы с атрибутами XML элемента
pub trait AttributeExt {
    /// Получение строкового значения атрибута
    fn get_attr(&self, name: &[u8]) -> Option<String>;

    /// Получение обязательного атрибута
    fn require_attr(&self, name: &[u8]) -> Result<String>;

    /// Получение атрибута как даты
    fn get_date_attr(&self, name: &[u8]) -> Option<NaiveDate>;

    /// Получение атрибута как числа
    fn get_f64_attr(&self, name: &[u8]) -> Option<f64>;

    /// Получение атрибута как целого числа
    fn get_i64_attr(&self, name: &[u8]) -> Option<i64>;

    /// Получение булевого атрибута
    fn get_bool_attr(&self, name: &[u8]) -> Option<bool>;
}

impl<'a> AttributeExt for BytesStart<'a> {
    fn get_attr(&self, name: &[u8]) -> Option<String> {
        self.attributes()
            .filter_map(|a| a.ok())
            .find(|a| a.key == QName(name))
            .and_then(|a| {
                String::from_utf8(a.value.to_vec()).ok()
            })
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
    }

    fn require_attr(&self, name: &[u8]) -> Result<String> {
        self.get_attr(name)
            .ok_or_else(|| {
                let attr_name = String::from_utf8_lossy(name);
                crate::error::Error::missing_field(attr_name.to_string())
            })
    }

    fn get_date_attr(&self, name: &[u8]) -> Option<NaiveDate> {
        self.get_attr(name)
            .and_then(|s| parse_date(&s))
    }

    fn get_f64_attr(&self, name: &[u8]) -> Option<f64> {
        self.get_attr(name)
            .and_then(|s| {
                // Заменяем запятую на точку для русского формата чисел
                let normalized = s.replace(',', ".");
                normalized.parse().ok()
            })
    }

    fn get_i64_attr(&self, name: &[u8]) -> Option<i64> {
        self.get_attr(name)
            .and_then(|s| s.parse().ok())
    }

    fn get_bool_attr(&self, name: &[u8]) -> Option<bool> {
        self.get_attr(name)
            .map(|s| {
                let lower = s.to_lowercase();
                lower == "1" || lower == "true" || lower == "да" || lower == "yes"
            })
    }
}

/// Парсинг даты в различных форматах
pub fn parse_date(s: &str) -> Option<NaiveDate> {
    let s = s.trim();
    if s.is_empty() {
        return None;
    }

    // Формат YYYY-MM-DD
    if let Ok(date) = NaiveDate::parse_from_str(s, "%Y-%m-%d") {
        return Some(date);
    }

    // Формат DD.MM.YYYY
    if let Ok(date) = NaiveDate::parse_from_str(s, "%d.%m.%Y") {
        return Some(date);
    }

    // Формат YYYY.MM.DD
    if let Ok(date) = NaiveDate::parse_from_str(s, "%Y.%m.%d") {
        return Some(date);
    }

    // Формат DD/MM/YYYY
    if let Ok(date) = NaiveDate::parse_from_str(s, "%d/%m/%Y") {
        return Some(date);
    }

    // Формат YYYYMMDD
    if s.len() == 8 {
        if let Ok(date) = NaiveDate::parse_from_str(s, "%Y%m%d") {
            return Some(date);
        }
    }

    None
}

/// Нормализация строки (удаление лишних пробелов)
pub fn normalize_string(s: &str) -> String {
    s.split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
}

/// Извлечение имени тега без namespace
pub fn tag_name(qname: &[u8]) -> &[u8] {
    // Убираем namespace prefix если есть
    if let Some(pos) = qname.iter().position(|&b| b == b':') {
        &qname[pos + 1..]
    } else {
        qname
    }
}

/// Сравнение имен тегов (без учета namespace и регистра для латиницы)
pub fn tag_matches(tag: &[u8], expected: &[u8]) -> bool {
    let tag = tag_name(tag);
    tag == expected
}

/// Список русских названий тегов с соответствующими английскими именами полей
pub mod tag_names {
    // Корневые элементы
    pub const FILE: &[u8] = "ФАЙЛ".as_bytes();
    pub const DOCUMENT: &[u8] = "Документ".as_bytes();
    
    // ЕГРЮЛ
    pub const SV_UL: &[u8] = "СвЮЛ".as_bytes();
    pub const SV_NAIM: &[u8] = "СвНаим".as_bytes();
    pub const SV_NAIM_POLN: &[u8] = "СвНаимПолн".as_bytes();
    pub const SV_ADR: &[u8] = "СвАдрес".as_bytes();
    pub const SV_ADR_MESTO: &[u8] = "СвАдресМН".as_bytes();
    pub const ADR_RF: &[u8] = "АдресРФ".as_bytes();
    pub const SV_UST_KAP: &[u8] = "СвУстКап".as_bytes();
    pub const SV_OKVAD: &[u8] = "СвОКВЭД".as_bytes();
    pub const SV_OKVAD_OSN: &[u8] = "СвОКВЭДОсн".as_bytes();
    pub const SV_OKVAD_DOP: &[u8] = "СвОКВЭДДоп".as_bytes();
    pub const SV_UCHRED: &[u8] = "СвУчредит".as_bytes();
    pub const SV_UCHRED_FL: &[u8] = "УчрФЛ".as_bytes();
    pub const SV_UCHRED_UL: &[u8] = "УчрЮЛРос".as_bytes();
    pub const SV_UCHRED_IN: &[u8] = "УчрЮЛИн".as_bytes();
    pub const SV_LIC_EIO: &[u8] = "СвЛицЕИО".as_bytes();
    pub const GRN_DATA: &[u8] = "ГРНДата".as_bytes();
    pub const SV_REG: &[u8] = "СвРегОрг".as_bytes();
    pub const SV_NAL: &[u8] = "СвНалУч".as_bytes();
    pub const SV_PFR: &[u8] = "СвРегПФ".as_bytes();
    pub const SV_FSS: &[u8] = "СвРегФСС".as_bytes();
    pub const SV_ZAPIS: &[u8] = "СвЗапись".as_bytes();
    pub const SV_STATUS: &[u8] = "СвСтатус".as_bytes();
    pub const SV_PREKR: &[u8] = "СвПрекрЮЛ".as_bytes();
    pub const SV_LIKV: &[u8] = "СвЛиквЮЛ".as_bytes();
    pub const SV_REORG: &[u8] = "СвРеорг".as_bytes();
    pub const DOLYA: &[u8] = "ДоляУстКап".as_bytes();
    
    // ЕГРИП
    pub const SV_IP: &[u8] = "СвИП".as_bytes();
    pub const SV_FL: &[u8] = "СвФЛ".as_bytes();
    pub const FIO: &[u8] = "ФИОРус".as_bytes();
    pub const SV_GRAZD: &[u8] = "СвГражд".as_bytes();
    pub const SV_REGIP: &[u8] = "СвРегИП".as_bytes();
    
    // Общие
    pub const REGION: &[u8] = "Регион".as_bytes();
    pub const GOROD: &[u8] = "Город".as_bytes();
    pub const NASEL_PUNKT: &[u8] = "НаселПункт".as_bytes();
    pub const ULICA: &[u8] = "Улица".as_bytes();
    pub const KONTAKT: &[u8] = "Контакт".as_bytes();
}

/// Атрибуты XML элементов
pub mod attr_names {
    // Основные идентификаторы
    pub const OGRN: &[u8] = "ОГРН".as_bytes();
    pub const OGRN_DATE: &[u8] = "ДатаОГРН".as_bytes();
    pub const OGRNIP: &[u8] = "ОГРНИП".as_bytes();
    pub const INN: &[u8] = "ИНН".as_bytes();
    pub const KPP: &[u8] = "КПП".as_bytes();
    pub const GRN: &[u8] = "ГРН".as_bytes();
    pub const DATE: &[u8] = "Дата".as_bytes();
    pub const DATE_REG: &[u8] = "ДатаРег".as_bytes();
    pub const DATE_VYP: &[u8] = "ДатаВып".as_bytes();
    
    // Наименование
    pub const NAIM_POLN: &[u8] = "НаимПолн".as_bytes();
    pub const NAIM_SOKR: &[u8] = "НаимСокр".as_bytes();
    pub const NAIMENOVANIE: &[u8] = "Наименование".as_bytes();
    
    // Статус
    pub const STATUS: &[u8] = "Статус".as_bytes();
    pub const STATUS_UL: &[u8] = "СтатусЮЛ".as_bytes();
    pub const KOD_STATUS: &[u8] = "КодСтатус".as_bytes();
    
    // Адрес
    pub const INDEX: &[u8] = "Индекс".as_bytes();
    pub const KOD_REGION: &[u8] = "КодРегион".as_bytes();
    pub const NAIMENOV: &[u8] = "Наименов".as_bytes();
    pub const TIP: &[u8] = "Тип".as_bytes();
    pub const DOM: &[u8] = "Дом".as_bytes();
    pub const KORP: &[u8] = "Корп".as_bytes();
    pub const KVAR: &[u8] = "Кварт".as_bytes();
    pub const ADRES: &[u8] = "Адрес".as_bytes();
    pub const FIAS: &[u8] = "ФИАС".as_bytes();
    
    // Уставный капитал
    pub const SUM_KAP: &[u8] = "СумКап".as_bytes();
    pub const NAIM_VAL: &[u8] = "НаимВал".as_bytes();
    
    // ОКВЭД
    pub const KOD_OKVAD: &[u8] = "КодОКВЭД".as_bytes();
    pub const NAIM_OKVAD: &[u8] = "НаимОКВЭД".as_bytes();
    pub const VER_OKVAD: &[u8] = "ВерсОКВЭД".as_bytes();
    
    // ФИО
    pub const FAMILIA: &[u8] = "Фамилия".as_bytes();
    pub const IMYA: &[u8] = "Имя".as_bytes();
    pub const OTCHESTVO: &[u8] = "Отчество".as_bytes();
    pub const POL: &[u8] = "Пол".as_bytes();
    
    // Руководитель
    pub const DOLZHN: &[u8] = "Должн".as_bytes();
    pub const NAIM_DOLZHN: &[u8] = "НаимДолжн".as_bytes();
    pub const VID_DOLZHN: &[u8] = "ВидДолжн".as_bytes();
    
    // Доля
    pub const NOMIN_STOIMOST: &[u8] = "НоминСтоим".as_bytes();
    pub const RAZM_DOLI: &[u8] = "РазмерДоли".as_bytes();
    pub const CHISLITEL: &[u8] = "Числит".as_bytes();
    pub const ZNAMENATEL: &[u8] = "Знамен".as_bytes();
    pub const PROCENT: &[u8] = "Процент".as_bytes();
    
    // Организационно-правовая форма
    pub const KOD_OPF: &[u8] = "КодОПФ".as_bytes();
    pub const NAIM_OPF: &[u8] = "ПолнНаимОПФ".as_bytes();
    
    // Регистрация
    pub const KOD_ORG: &[u8] = "КодОрг".as_bytes();
    pub const NAIM_ORG: &[u8] = "НаимОрг".as_bytes();
    pub const REG_NOMER: &[u8] = "РегНомер".as_bytes();
    
    // Гражданство
    pub const VID_GRAZD: &[u8] = "ВидГражд".as_bytes();
    pub const KOD_OKSM: &[u8] = "ОКСМ".as_bytes();
    pub const NAIM_STRAN: &[u8] = "НаимСтран".as_bytes();
    
    // Email
    pub const EMAIL: &[u8] = "E-mail".as_bytes();
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_date_iso() {
        assert_eq!(
            parse_date("2023-05-15"),
            Some(NaiveDate::from_ymd_opt(2023, 5, 15).unwrap())
        );
    }

    #[test]
    fn test_parse_date_russian() {
        assert_eq!(
            parse_date("15.05.2023"),
            Some(NaiveDate::from_ymd_opt(2023, 5, 15).unwrap())
        );
    }

    #[test]
    fn test_parse_date_compact() {
        assert_eq!(
            parse_date("20230515"),
            Some(NaiveDate::from_ymd_opt(2023, 5, 15).unwrap())
        );
    }

    #[test]
    fn test_normalize_string() {
        assert_eq!(
            normalize_string("  hello   world  "),
            "hello world"
        );
    }
}

