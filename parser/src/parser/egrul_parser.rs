//! Парсер ЕГРЮЛ (юридические лица)

use quick_xml::events::{BytesStart, Event};
use quick_xml::Reader;
use tracing::{debug, warn};

use crate::error::{Error, Result};
use crate::models::{
    EgrulRecord, Address, Capital, Activity, Person, Founder, Share,
    HistoryRecord, RegistrationAuthority, TaxAuthority, 
    PensionFund, SocialInsurance,
};
use crate::models::egrul::{HeadInfo, RegistrationInfo};
use super::attributes::{AttributeExt, attr_names, tag_names, tag_matches, normalize_string};

/// Парсер XML для ЕГРЮЛ
pub struct EgrulXmlParser;

impl EgrulXmlParser {
    pub fn new() -> Self {
        Self
    }

    /// Извлечение кода региона из кода налогового органа (первые 2 цифры)
    fn extract_region_code_from_tax_code(tax_code: &str) -> Option<String> {
        if tax_code.len() >= 2 {
            let region_part = &tax_code[0..2];
            // Проверяем, что это действительно цифры
            if region_part.chars().all(|c| c.is_ascii_digit()) {
                return Some(region_part.to_string());
            }
        }
        None
    }

    /// Парсинг XML контента
    pub fn parse(&self, content: &str) -> Result<Vec<EgrulRecord>> {
        let mut reader = Reader::from_str(content);
        reader.config_mut().trim_text(true);

        let mut records = Vec::new();
        let mut buf = Vec::new();

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Ищем элементы СвЮЛ
                    if tag_matches(tag, tag_names::SV_UL) || tag_matches(tag, "СвЮЛ".as_bytes()) {
                        match self.parse_sv_ul(&mut reader, e) {
                            Ok(record) => {
                                if record.is_valid() {
                                    records.push(record);
                                }
                            }
                            Err(e) => {
                                warn!("Ошибка парсинга СвЮЛ: {}", e);
                            }
                        }
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => {
                    warn!("Ошибка чтения XML: {}", e);
                    break;
                }
                _ => {}
            }
            buf.clear();
        }

        debug!("Распарсено {} записей ЕГРЮЛ", records.len());
        Ok(records)
    }

    /// Парсинг элемента СвЮЛ
    fn parse_sv_ul(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<EgrulRecord> {
        let mut record = EgrulRecord::new();

        // Атрибуты корневого элемента СвЮЛ
        record.ogrn = start.get_attr(attr_names::OGRN)
            .or_else(|| start.get_attr("ОГРН".as_bytes()))
            .unwrap_or_default();
        record.ogrn_date = start.get_date_attr(attr_names::OGRN_DATE)
            .or_else(|| start.get_date_attr("ДатаОГРН".as_bytes()));
        record.inn = start.get_attr(attr_names::INN)
            .or_else(|| start.get_attr("ИНН".as_bytes()))
            .unwrap_or_default();
        record.kpp = start.get_attr(attr_names::KPP)
            .or_else(|| start.get_attr("КПП".as_bytes()));
        record.extract_date = start.get_date_attr(attr_names::DATE_VYP)
            .or_else(|| start.get_date_attr("ДатаВып".as_bytes()));
        
        // Статус: на уровне парсера сохраняем только исходное значение из XML.
        // Любая интерпретация (активно/ликвидировано и т.п.) должна выполняться выше по конвейеру.
        if let Some(status_str) = start.get_attr(attr_names::STATUS_UL)
            .or_else(|| start.get_attr("СтатусЮЛ".as_bytes()))
            .or_else(|| start.get_attr(attr_names::STATUS)) 
        {
            record.status_code = Some(status_str);
        }

        // Организационно-правовая форма (ОПФ)
        record.opf_code = start.get_attr("КодОПФ".as_bytes())
            .or_else(|| start.get_attr(attr_names::KOD_OPF));
        record.opf_name = start.get_attr("ПолнНаимОПФ".as_bytes())
            .or_else(|| start.get_attr("НаимОПФ".as_bytes()))
            .or_else(|| start.get_attr(attr_names::NAIM_OPF));

        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Наименование
                    if tag_matches(tag, "СвНаимЮЛ".as_bytes()) || tag_matches(tag, "СвНаим".as_bytes()) {
                        self.parse_sv_naim(reader, e, &mut record)?;
                        depth -= 1;
                    }
                    // Адрес
                    else if tag_matches(tag, "СвАдресЮЛ".as_bytes()) || tag_matches(tag, "СвАдрес".as_bytes()) {
                        record.address = Some(self.parse_address(reader, e)?);
                        depth -= 1;
                    }
                    // Уставный капитал
                    else if tag_matches(tag, "СвУстКап".as_bytes()) {
                        record.capital = self.parse_capital(e);
                    }
                    // ОКВЭД
                    else if tag_matches(tag, "СвОКВЭД".as_bytes()) {
                        self.parse_okvad(reader, e, &mut record)?;
                        depth -= 1;
                    }
                    // Руководитель (ЕИО)
                    else if tag_matches(tag, "СведДолжнФЛ".as_bytes()) || tag_matches(tag, "СвЛицЕИО".as_bytes()) || tag_matches(tag, "СведДол662".as_bytes()) {
                        record.head = Some(self.parse_head(reader, e)?);
                        depth -= 1;
                    }
                    // Учредители
                    else if tag_matches(tag, "СвУчредит".as_bytes()) {
                        self.parse_founders(reader, e, &mut record)?;
                        depth -= 1;
                    }
                    // Регистрация
                    else if tag_matches(tag, "СвРегОрг".as_bytes()) {
                        // Извлекаем код региона из КодНО как fallback
                        if let Some(tax_code) = e.get_attr("КодНО".as_bytes())
                            .or_else(|| e.get_attr("КодОрг".as_bytes())) {
                            if let Some(region_code) = Self::extract_region_code_from_tax_code(&tax_code) {
                                // Устанавливаем код региона только если его нет в адресе
                                if record.address.is_none() {
                                    record.address = Some(Address::default());
                                }
                                if let Some(ref mut address) = record.address {
                                    // Используем код из налогового органа только если в адресе нет кода региона
                                    if address.region_code.is_none() {
                                        address.region_code = Some(region_code);
                                    }
                                }
                            }
                        }
                        record.registration = Some(self.parse_registration(reader, e)?);
                        depth -= 1;
                    }
                    // Налоговый орган
                    else if tag_matches(tag, "СвУчетНО".as_bytes()) || tag_matches(tag, "СвНалУч".as_bytes()) {
                        record.tax_authority = Some(self.parse_tax_authority(reader, e)?);
                        depth -= 1;
                    }
                    // ПФР
                    else if tag_matches(tag, "СвРегПФ".as_bytes()) {
                        record.pension_fund = Some(self.parse_pension_fund(reader, e)?);
                        depth -= 1;
                    }
                    // ФСС
                    else if tag_matches(tag, "СвРегФСС".as_bytes()) {
                        record.social_insurance = Some(self.parse_social_insurance(reader, e)?);
                        depth -= 1;
                    }
                    // История записей
                    else if tag_matches(tag, "СвЗапись".as_bytes()) || tag_matches(tag, "СвЗапЕГРЮЛ".as_bytes()) {
                        if let Ok(history_record) = self.parse_history_record(reader, e) {
                            record.history.push(history_record);
                        }
                        depth -= 1;
                    }
                    // Email
                    else if tag_matches(tag, "СвКонт".as_bytes()) || tag_matches(tag, "СведКонт".as_bytes()) {
                        if let Some(email) = e.get_attr("E-mail".as_bytes())
                            .or_else(|| e.get_attr(attr_names::EMAIL)) {
                            record.email = Some(email);
                        }
                    }
                    // Прекращение
                    else if tag_matches(tag, "СвПрекрЮЛ".as_bytes()) {
                        record.termination_date = e.get_date_attr("ДатаПрекращ".as_bytes())
                            .or_else(|| e.get_date_attr("Дата".as_bytes()));
                        record.termination_method = e.get_attr("СпосПрекращ".as_bytes());
                    }
                    // Статус ЮЛ (вложенный блок СвСтатус / СвСтатусЮЛ)
                    else if tag_matches(tag, "СвСтатус".as_bytes()) || tag_matches(tag, "СвСтатусЮЛ".as_bytes()) {
                        let code = e.get_attr("КодСтатусЮЛ".as_bytes())
                            .or_else(|| e.get_attr("КодСтатус".as_bytes()));

                        if let Some(c) = code {
                            record.status_code = Some(c);
                        }
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Наименование (empty element)
                    if tag_matches(tag, "СвНаимЮЛ".as_bytes()) || tag_matches(tag, "СвНаим".as_bytes()) {
                        record.full_name = e.get_attr("НаимЮЛПолн".as_bytes())
                            .or_else(|| e.get_attr("НаимПолн".as_bytes()))
                            .or_else(|| e.get_attr(attr_names::NAIM_POLN))
                            .map(|s| normalize_string(&s))
                            .unwrap_or_default();
                        record.short_name = e.get_attr("НаимСокр".as_bytes())
                            .or_else(|| e.get_attr(attr_names::NAIM_SOKR))
                            .map(|s| normalize_string(&s));
                    }
                    // Уставный капитал
                    else if tag_matches(tag, "СвУстКап".as_bytes()) {
                        record.capital = self.parse_capital(e);
                    }
                    // ОКВЭД основной
                    else if tag_matches(tag, "СвОКВЭДОсн".as_bytes()) {
                        record.main_activity = self.parse_activity(e);
                    }
                    // ОКВЭД дополнительный
                    else if tag_matches(tag, "СвОКВЭДДоп".as_bytes()) {
                        if let Some(activity) = self.parse_activity(e) {
                            record.additional_activities.push(activity);
                        }
                    }
                    // Email
                    else if tag_matches(tag, "СвКонт".as_bytes()) {
                        if let Some(email) = e.get_attr("E-mail".as_bytes()) {
                            record.email = Some(email);
                        }
                    }
                    // Статус ЮЛ (СвСтатус / СвСтатусЮЛ как empty element)
                    else if tag_matches(tag, "СвСтатус".as_bytes()) || tag_matches(tag, "СвСтатусЮЛ".as_bytes()) {
                        let code = e.get_attr("КодСтатусЮЛ".as_bytes())
                            .or_else(|| e.get_attr("КодСтатус".as_bytes()));

                        if let Some(c) = code {
                            record.status_code = Some(c);
                        }
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        // Устанавливаем дату регистрации из ОГРН даты если не задана
        if record.registration_date.is_none() {
            record.registration_date = record.ogrn_date;
        }

        // Копируем данные о первом свидетельстве из истории в registration
        if let Some(ref mut registration) = record.registration {
            if registration.certificate_series.is_none() {
                // Ищем первую запись истории со свидетельством
                if let Some(history_with_cert) = record.history.iter()
                    .find(|h| h.certificate_series.is_some() || h.certificate_number.is_some())
                {
                    registration.certificate_series = history_with_cert.certificate_series.clone();
                    registration.certificate_number = history_with_cert.certificate_number.clone();
                    registration.certificate_date = history_with_cert.certificate_date;
                }
            }
        }

        Ok(record)
    }

    /// Парсинг наименования (СвНаимЮЛ)
    fn parse_sv_naim(&self, reader: &mut Reader<&[u8]>, start: &BytesStart, record: &mut EgrulRecord) -> Result<()> {
        // Ищем полное наименование на корневом элементе
        record.full_name = start.get_attr("НаимЮЛПолн".as_bytes())
            .or_else(|| start.get_attr("НаимПолн".as_bytes()))
            .or_else(|| start.get_attr(attr_names::NAIM_POLN))
            .map(|s| normalize_string(&s))
            .unwrap_or_default();
        
        // Ищем сокращенное наименование на корневом элементе (может быть здесь)
        record.short_name = start.get_attr("НаимСокр".as_bytes())
            .or_else(|| start.get_attr(attr_names::NAIM_SOKR))
            .map(|s| normalize_string(&s));

        // Парсим вложенные элементы для поиска сокращенного наименования
        let mut buf = Vec::new();
        let mut depth = 1;
        
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Вложенный СвНаимЮЛСокр содержит сокращённое название
                    if tag_matches(tag, "СвНаимЮЛСокр".as_bytes()) {
                        if let Some(short) = e.get_attr("НаимСокр".as_bytes()) {
                            record.short_name = Some(normalize_string(&short));
                        }
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвНаимЮЛСокр".as_bytes()) {
                        if let Some(short) = e.get_attr("НаимСокр".as_bytes()) {
                            record.short_name = Some(normalize_string(&short));
                        }
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }
        
        Ok(())
    }

    /// Парсинг адреса
    fn parse_address(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<Address> {
        let mut address = Address::default();
        let mut buf = Vec::new();
        let mut depth = 1;

        // Проверяем атрибут Адрес на корневом элементе
        if let Some(full) = start.get_attr("Адрес".as_bytes())
            .or_else(|| start.get_attr(attr_names::ADRES)) {
            address.full_address = Some(normalize_string(&full));
        }

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();

                    self.process_address_element(e, tag, &mut address);
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();

                    self.process_address_element(e, tag, &mut address);
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        // Если полный адрес не был задан, собираем из компонентов
        if address.full_address.is_none() && !address.is_empty() {
            address.full_address = Some(address.build_full_address());
        }

        Ok(address)
    }

    /// Обработка элемента адреса
    fn process_address_element(&self, e: &BytesStart, tag: &[u8], address: &mut Address) {
        if tag_matches(tag, "АдресРФ".as_bytes()) || tag_matches(tag, "Адрес".as_bytes()) {
            address.postal_code = e.get_attr("Индекс".as_bytes())
                .or_else(|| e.get_attr(attr_names::INDEX));
            address.region_code = e.get_attr("КодРегион".as_bytes())
                .or_else(|| e.get_attr(attr_names::KOD_REGION));
        }
        else if tag_matches(tag, "Регион".as_bytes()) {
            // Парсим только код региона, название не сохраняем
            // Код региона уже извлекается из АдресРФ выше
        }
        else if tag_matches(tag, "Район".as_bytes()) {
            address.district = e.get_attr("НаименРай".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr(attr_names::NAIMENOV))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        else if tag_matches(tag, "Город".as_bytes()) {
            address.city = e.get_attr("НаименГор".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr(attr_names::NAIMENOV))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        else if tag_matches(tag, "НаселПункт".as_bytes()) {
            address.locality = e.get_attr("НаименНП".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr(attr_names::NAIMENOV))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        else if tag_matches(tag, "Улица".as_bytes()) {
            // Собираем улицу из типа и названия: "МКР. 7-Й"
            let typ = e.get_attr("ТипУлица".as_bytes())
                .or_else(|| e.get_attr("Тип".as_bytes()));
            let name = e.get_attr("НаимУлица".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr(attr_names::NAIMENOV));
            
            address.street = match (typ, name) {
                (Some(t), Some(n)) => Some(format!("{} {}", t, n)),
                (None, Some(n)) => Some(n),
                (Some(t), None) => Some(t),
                (None, None) => None,
            };
        }
        // Дом, корпус, квартира - из атрибутов адреса
        address.house = address.house.take().or_else(|| e.get_attr("Дом".as_bytes()));
        address.building = address.building.take().or_else(|| e.get_attr("Корп".as_bytes()));
        address.flat = address.flat.take().or_else(|| e.get_attr("Кварт".as_bytes())
            .or_else(|| e.get_attr("Офис".as_bytes())));
        address.fias_id = address.fias_id.take().or_else(|| e.get_attr("ИдНом".as_bytes())
            .or_else(|| e.get_attr("ФИАС".as_bytes())));
    }

    /// Парсинг уставного капитала
    fn parse_capital(&self, e: &BytesStart) -> Option<Capital> {
        let amount = e.get_f64_attr("СумКап".as_bytes())
            .or_else(|| e.get_f64_attr(attr_names::SUM_KAP))?;
        
        let currency = e.get_attr("НаимВал".as_bytes())
            .or_else(|| e.get_attr(attr_names::NAIM_VAL))
            .unwrap_or_else(|| "рубль".to_string());

        Some(Capital::new(amount, currency))
    }

    /// Парсинг вида деятельности
    fn parse_activity(&self, e: &BytesStart) -> Option<Activity> {
        let code = e.get_attr("КодОКВЭД".as_bytes())
            .or_else(|| e.get_attr(attr_names::KOD_OKVAD))?;
        
        let name = e.get_attr("НаимОКВЭД".as_bytes())
            .or_else(|| e.get_attr(attr_names::NAIM_OKVAD))
            .unwrap_or_default();
        
        let version = e.get_attr("ВерсОКВЭД".as_bytes())
            .or_else(|| e.get_attr(attr_names::VER_OKVAD));

        Some(Activity {
            code,
            name,
            version,
            is_main: false,
        })
    }

    /// Парсинг блока ОКВЭД
    fn parse_okvad(&self, reader: &mut Reader<&[u8]>, _start: &BytesStart, record: &mut EgrulRecord) -> Result<()> {
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОКВЭДОсн".as_bytes()) {
                        if let Some(mut activity) = self.parse_activity(e) {
                            activity.is_main = true;
                            record.main_activity = Some(activity);
                        }
                    }
                    else if tag_matches(tag, "СвОКВЭДДоп".as_bytes()) {
                        if let Some(activity) = self.parse_activity(e) {
                            record.additional_activities.push(activity);
                        }
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОКВЭДОсн".as_bytes()) {
                        if let Some(mut activity) = self.parse_activity(e) {
                            activity.is_main = true;
                            record.main_activity = Some(activity);
                        }
                    }
                    else if tag_matches(tag, "СвОКВЭДДоп".as_bytes()) {
                        if let Some(activity) = self.parse_activity(e) {
                            record.additional_activities.push(activity);
                        }
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(())
    }

    /// Парсинг руководителя
    fn parse_head(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<HeadInfo> {
        let mut head = HeadInfo::default();
        
        // Атрибуты корневого элемента (если есть)
        head.position = start.get_attr("НаимДолжн".as_bytes())
            .or_else(|| start.get_attr(attr_names::NAIM_DOLZHN));
        head.position_code = start.get_attr("ВидДолжн".as_bytes())
            .or_else(|| start.get_attr(attr_names::VID_DOLZHN));

        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // ФИО в СвФЛ или ФИОРус
                    if tag_matches(tag, "СвФЛ".as_bytes()) || tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        head.person = self.parse_person(e);
                    }
                    // Должность в СвДолжн
                    else if tag_matches(tag, "СвДолжн".as_bytes()) {
                        head.position = e.get_attr("НаимДолжн".as_bytes())
                            .or_else(|| e.get_attr("НаимВидДолжн".as_bytes()));
                        head.position_code = e.get_attr("ВидДолжн".as_bytes());
                    }
                    // ГРН дата
                    else if tag_matches(tag, "ГРНДата".as_bytes()) || tag_matches(tag, "ГРНДатаПерв".as_bytes()) {
                        head.grn = e.get_attr("ГРН".as_bytes());
                        head.grn_date = e.get_date_attr("ДатаЗаписи".as_bytes())
                            .or_else(|| e.get_date_attr("Дата".as_bytes()));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвФЛ".as_bytes()) || tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        head.person = self.parse_person(e);
                    }
                    else if tag_matches(tag, "СвДолжн".as_bytes()) {
                        head.position = e.get_attr("НаимДолжн".as_bytes())
                            .or_else(|| e.get_attr("НаимВидДолжн".as_bytes()));
                        head.position_code = e.get_attr("ВидДолжн".as_bytes());
                    }
                    else if tag_matches(tag, "ГРНДата".as_bytes()) || tag_matches(tag, "ГРНДатаПерв".as_bytes()) {
                        head.grn = e.get_attr("ГРН".as_bytes());
                        head.grn_date = e.get_date_attr("ДатаЗаписи".as_bytes())
                            .or_else(|| e.get_date_attr("Дата".as_bytes()));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(head)
    }

    /// Парсинг ФИО
    fn parse_person(&self, e: &BytesStart) -> Person {
        Person {
            last_name: e.get_attr("Фамилия".as_bytes())
                .or_else(|| e.get_attr(attr_names::FAMILIA))
                .unwrap_or_default(),
            first_name: e.get_attr("Имя".as_bytes())
                .or_else(|| e.get_attr(attr_names::IMYA))
                .unwrap_or_default(),
            middle_name: e.get_attr("Отчество".as_bytes())
                .or_else(|| e.get_attr(attr_names::OTCHESTVO)),
            inn: e.get_attr("ИННФЛ".as_bytes())
                .or_else(|| e.get_attr("ИНН".as_bytes())),
        }
    }

    /// Парсинг учредителей
    fn parse_founders(&self, reader: &mut Reader<&[u8]>, _start: &BytesStart, record: &mut EgrulRecord) -> Result<()> {
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Учредитель - физическое лицо
                    if tag_matches(tag, "УчрФЛ".as_bytes()) {
                        if let Ok(founder) = self.parse_founder_fl(reader, e) {
                            record.founders.push(founder);
                            depth -= 1;
                        }
                    }
                    // Учредитель - российское юр. лицо
                    else if tag_matches(tag, "УчрЮЛРос".as_bytes()) {
                        if let Ok(founder) = self.parse_founder_ul_ros(reader, e) {
                            record.founders.push(founder);
                            depth -= 1;
                        }
                    }
                    // Учредитель - иностранное юр. лицо
                    else if tag_matches(tag, "УчрЮЛИн".as_bytes()) {
                        if let Ok(founder) = self.parse_founder_ul_in(reader, e) {
                            record.founders.push(founder);
                            depth -= 1;
                        }
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(())
    }

    /// Парсинг учредителя - физического лица
    fn parse_founder_fl(&self, reader: &mut Reader<&[u8]>, _start: &BytesStart) -> Result<Founder> {
        let mut person = Person::default();
        let mut share: Option<Share> = None;
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "СвФЛ".as_bytes()) {
                        person = self.parse_person(e);
                    }
                    else if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "СвФЛ".as_bytes()) {
                        person = self.parse_person(e);
                    }
                    else if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(Founder::Person {
            person,
            share,
            citizenship: None,
        })
    }

    /// Парсинг учредителя - российского юр. лица
    fn parse_founder_ul_ros(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<Founder> {
        let ogrn = start.get_attr("ОГРН".as_bytes()).unwrap_or_default();
        let inn = start.get_attr("ИНН".as_bytes()).unwrap_or_default();
        let name = start.get_attr("НаимЮЛПолн".as_bytes())
            .or_else(|| start.get_attr("Наименование".as_bytes()))
            .unwrap_or_default();
        
        let mut share: Option<Share> = None;
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(Founder::RussianLegalEntity {
            ogrn,
            inn,
            name,
            share,
        })
    }

    /// Парсинг учредителя - иностранного юр. лица
    fn parse_founder_ul_in(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<Founder> {
        let name = start.get_attr("НаимЮЛПолн".as_bytes())
            .or_else(|| start.get_attr("Наименование".as_bytes()))
            .unwrap_or_default();
        let country = start.get_attr("НаимСтран".as_bytes());
        let reg_number = start.get_attr("РегНомер".as_bytes());
        
        let mut share: Option<Share> = None;
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();

                    if tag_matches(tag, "ДоляУстКап".as_bytes()) {
                        share = Some(self.parse_share(e));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(Founder::ForeignLegalEntity {
            name,
            country,
            reg_number,
            share,
        })
    }

    /// Парсинг доли
    fn parse_share(&self, e: &BytesStart) -> Share {
        Share {
            nominal_value: e.get_f64_attr("НоминСтоим".as_bytes())
                .or_else(|| e.get_f64_attr("НоминСтоим".as_bytes())),
            numerator: e.get_i64_attr("Числит".as_bytes()),
            denominator: e.get_i64_attr("Знамен".as_bytes()),
            percent: e.get_f64_attr("Процент".as_bytes()),
        }
    }

    /// Парсинг регистрации
    fn parse_registration(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<RegistrationInfo> {
        let mut reg = RegistrationInfo::default();
        
        let tax_code = start.get_attr("КодНО".as_bytes())
            .or_else(|| start.get_attr("КодОрг".as_bytes()))
            .unwrap_or_default();
        
        reg.authority = Some(RegistrationAuthority {
            code: tax_code.clone(),
            name: start.get_attr("НаимНО".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes()))
                .unwrap_or_default(),
            address: start.get_attr("АдрРО".as_bytes())
                .or_else(|| start.get_attr("АдрОрг".as_bytes())),
        });

        self.skip_element(reader)?;
        Ok(reg)
    }

    /// Парсинг налогового органа
    fn parse_tax_authority(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<TaxAuthority> {
        let mut tax = TaxAuthority {
            code: start.get_attr("КодНО".as_bytes())
                .or_else(|| start.get_attr("СвНалОрг".as_bytes()))
                .unwrap_or_default(),
            name: start.get_attr("НаимНО".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes()))
                .unwrap_or_default(),
            registration_date: start.get_date_attr("ДатаПостУч".as_bytes())
                .or_else(|| start.get_date_attr("ДатаПост".as_bytes()))
                .or_else(|| start.get_date_attr("ДатаРег".as_bytes())),
        };

        // Парсим вложенные элементы для получения СвНО
        let mut buf = Vec::new();
        let mut depth = 1;
        
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвНО".as_bytes()) {
                        if let Some(code) = e.get_attr("КодНО".as_bytes()) {
                            tax.code = code;
                        }
                        if let Some(name) = e.get_attr("НаимНО".as_bytes()) {
                            tax.name = name;
                        }
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвНО".as_bytes()) {
                        if let Some(code) = e.get_attr("КодНО".as_bytes()) {
                            tax.code = code;
                        }
                        if let Some(name_val) = e.get_attr("НаимНО".as_bytes()) {
                            tax.name = name_val;
                        }
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(tax)
    }

    /// Парсинг ПФР
    fn parse_pension_fund(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<PensionFund> {
        let mut pf = PensionFund {
            reg_number: start.get_attr("РегНомПФ".as_bytes())
                .or_else(|| start.get_attr("РегНомер".as_bytes()))
                .unwrap_or_default(),
            registration_date: start.get_date_attr("ДатаРег".as_bytes()),
            authority_name: start.get_attr("НаимПФ".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes())),
        };

        // Парсим вложенные элементы для получения СвОргПФ
        let mut buf = Vec::new();
        let mut depth = 1;
        
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОргПФ".as_bytes()) {
                        pf.authority_name = e.get_attr("НаимПФ".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОргПФ".as_bytes()) {
                        pf.authority_name = e.get_attr("НаимПФ".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(pf)
    }

    /// Парсинг ФСС
    fn parse_social_insurance(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<SocialInsurance> {
        let mut fss = SocialInsurance {
            reg_number: start.get_attr("РегНомФСС".as_bytes())
                .or_else(|| start.get_attr("РегНомер".as_bytes()))
                .unwrap_or_default(),
            registration_date: start.get_date_attr("ДатаРег".as_bytes()),
            authority_name: start.get_attr("НаимФСС".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes())),
        };

        // Парсим вложенные элементы для получения СвОргФСС
        let mut buf = Vec::new();
        let mut depth = 1;
        
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОргФСС".as_bytes()) {
                        fss.authority_name = e.get_attr("НаимФСС".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "СвОргФСС".as_bytes()) {
                        fss.authority_name = e.get_attr("НаимФСС".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(fss)
    }

    /// Парсинг записи истории
    fn parse_history_record(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<HistoryRecord> {
        let mut record = HistoryRecord::default();
        
        record.grn = start.get_attr("ГРН".as_bytes())
            .or_else(|| start.get_attr("ИдЗап".as_bytes()))
            .unwrap_or_default();
        record.date = start.get_date_attr("ДатаЗап".as_bytes())
            .or_else(|| start.get_date_attr("Дата".as_bytes()));
        
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "ВидЗап".as_bytes()) {
                        record.reason_code = e.get_attr("КодСПВЗ".as_bytes());
                        record.reason_description = e.get_attr("НаимВидЗап".as_bytes())
                            .or_else(|| e.get_attr("НаимВид".as_bytes()))
                            .or_else(|| e.get_attr("Наименование".as_bytes()));
                    }
                    else if tag_matches(tag, "СвРегОрг".as_bytes()) || tag_matches(tag, "РегОрг".as_bytes()) {
                        record.authority_code = e.get_attr("КодНО".as_bytes())
                            .or_else(|| e.get_attr("КодОрг".as_bytes()));
                        record.authority_name = e.get_attr("НаимНО".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "ВидЗап".as_bytes()) {
                        record.reason_code = e.get_attr("КодСПВЗ".as_bytes());
                        record.reason_description = e.get_attr("НаимВидЗап".as_bytes())
                            .or_else(|| e.get_attr("НаимВид".as_bytes()))
                            .or_else(|| e.get_attr("Наименование".as_bytes()));
                    }
                    else if tag_matches(tag, "СвРегОрг".as_bytes()) || tag_matches(tag, "РегОрг".as_bytes()) {
                        record.authority_code = e.get_attr("КодНО".as_bytes())
                            .or_else(|| e.get_attr("КодОрг".as_bytes()));
                        record.authority_name = e.get_attr("НаимНО".as_bytes())
                            .or_else(|| e.get_attr("НаимОрг".as_bytes()));
                    }
                    else if tag_matches(tag, "СвСвид".as_bytes()) {
                        record.certificate_series = e.get_attr("Серия".as_bytes());
                        record.certificate_number = e.get_attr("Номер".as_bytes());
                        record.certificate_date = e.get_date_attr("ДатаВыдСвид".as_bytes())
                            .or_else(|| e.get_date_attr("ДатаВыд".as_bytes()));
                    }
                }
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(record)
    }

    /// Пропуск элемента до его закрывающего тега
    fn skip_element(&self, reader: &mut Reader<&[u8]>) -> Result<()> {
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(_)) => depth += 1,
                Ok(Event::End(_)) => {
                    depth -= 1;
                    if depth == 0 {
                        break;
                    }
                }
                Ok(Event::Eof) => break,
                Err(e) => return Err(Error::Xml(e)),
                _ => {}
            }
            buf.clear();
        }

        Ok(())
    }
}

impl Default for EgrulXmlParser {
    fn default() -> Self {
        Self::new()
    }
}

