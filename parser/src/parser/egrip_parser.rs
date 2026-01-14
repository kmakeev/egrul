//! Парсер ЕГРИП (индивидуальные предприниматели)

use quick_xml::events::{BytesStart, Event};
use quick_xml::Reader;
use tracing::{debug, warn};

use crate::error::{Error, Result};
use crate::models::{
    EgripRecord, Address, Activity, Person,
    HistoryRecord, RegistrationAuthority, TaxAuthority,
    PensionFund, SocialInsurance,
};
use crate::models::egrip::{Gender, CitizenshipInfo, CitizenshipType, IpRegistrationInfo};
use super::attributes::{AttributeExt, attr_names, tag_matches, normalize_string};

/// Парсер XML для ЕГРИП
pub struct EgripXmlParser;

impl EgripXmlParser {
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
    pub fn parse(&self, content: &str) -> Result<Vec<EgripRecord>> {
        let mut reader = Reader::from_str(content);
        reader.config_mut().trim_text(true);

        let mut records = Vec::new();
        let mut buf = Vec::new();

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // Ищем элементы СвИП
                    if tag_matches(tag, "СвИП".as_bytes()) {
                        match self.parse_sv_ip(&mut reader, e) {
                            Ok(record) => {
                                if record.is_valid() {
                                    records.push(record);
                                }
                            }
                            Err(e) => {
                                warn!("Ошибка парсинга СвИП: {}", e);
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

        debug!("Распарсено {} записей ЕГРИП", records.len());
        Ok(records)
    }

    /// Парсинг элемента СвИП
    fn parse_sv_ip(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<EgripRecord> {
        let mut record = EgripRecord::new();

        // Атрибуты корневого элемента СвИП
        record.ogrnip = start.get_attr("ОГРНИП".as_bytes())
            .or_else(|| start.get_attr(attr_names::OGRNIP))
            .unwrap_or_default();
        record.ogrnip_date = start.get_date_attr("ДатаОГРНИП".as_bytes())
            .or_else(|| start.get_date_attr("ДатаРег".as_bytes()));
        record.inn = start.get_attr("ИННФЛ".as_bytes())
            .or_else(|| start.get_attr(attr_names::INN))
            .or_else(|| start.get_attr("ИНН".as_bytes()))
            .unwrap_or_default();
        record.extract_date = start.get_date_attr("ДатаВып".as_bytes())
            .or_else(|| start.get_date_attr(attr_names::DATE_VYP));
        
        // Статус: на уровне парсера сохраняем только исходный код/реквизиты из XML.
        // Никаких попыток интерпретации кода или текста статуса здесь не делаем.
        if let Some(status_str) = start.get_attr("Статус".as_bytes())
            .or_else(|| start.get_attr(attr_names::STATUS))
        {
            record.status_code = Some(status_str);
        }

        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // СвФЛ - содержит вложенный ФИОРус и атрибут Пол
                    if tag_matches(tag, "СвФЛ".as_bytes()) {
                        // Считываем пол из атрибута СвФЛ
                        if let Some(gender_code) = e.get_attr("Пол".as_bytes()) {
                            record.gender = Some(Gender::from_code(&gender_code));
                        }
                        // Внутри СвФЛ ищем ФИОРус
                        self.parse_sv_fl(reader, &mut record)?;
                        depth -= 1;
                    }
                    // ФИО напрямую (если без вложенности)
                    else if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИОИП".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        record.person = self.parse_person(e);
                        self.skip_element(reader)?;
                        depth -= 1;
                    }
                    // ОГРНИП в элементе СвРегИП
                    else if tag_matches(tag, "СвРегИП".as_bytes()) {
                        if let Some(ogrnip) = e.get_attr("ОГРНИП".as_bytes()) {
                            record.ogrnip = ogrnip;
                        }
                        if let Some(date) = e.get_date_attr("ДатаОГРНИП".as_bytes()) {
                            record.ogrnip_date = Some(date);
                        }
                    }
                    // Пол
                    else if tag_matches(tag, "Пол".as_bytes()) {
                        if let Some(gender_code) = e.get_attr("КодПол".as_bytes())
                            .or_else(|| e.get_attr("Пол".as_bytes())) {
                            record.gender = Some(Gender::from_code(&gender_code));
                        }
                    }
                    // Гражданство
                    else if tag_matches(tag, "СвГражд".as_bytes()) || tag_matches(tag, "Гражданство".as_bytes()) {
                        record.citizenship = Some(self.parse_citizenship(e));
                    }
                    // Статус ИП (элемент СвСтатус)
                    else if tag_matches(tag, "СвСтатус".as_bytes()) {
                        let code = e.get_attr("КодСтатус".as_bytes());
                        let term_date = e.get_date_attr("ДатаПрекращ".as_bytes());

                        if let Some(c) = code {
                            record.status_code = Some(c);
                        }
                        if let Some(d) = term_date {
                            record.termination_date = Some(d);
                        }
                    }
                    // Адрес (регион места жительства)
                    else if tag_matches(tag, "СвАдрМЖ".as_bytes()) || tag_matches(tag, "СвАдрес".as_bytes()) {
                        record.address = Some(self.parse_address(reader, e)?);
                        depth -= 1;
                    }
                    // ОКВЭД
                    else if tag_matches(tag, "СвОКВЭД".as_bytes()) {
                        self.parse_okvad(reader, e, &mut record)?;
                        depth -= 1;
                    }
                    // Регистрация
                    else if tag_matches(tag, "СвРегОрг".as_bytes()) || tag_matches(tag, "СвГосРег".as_bytes()) {
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
                        
                        // Сохраняем адрес регистрирующего органа как строку
                        if let Some(adro) = e.get_attr("АдрРО".as_bytes()) {
                            // Если у нас еще нет адреса, создаем новый с полным адресом из АдрРО
                            if record.address.is_none() {
                                let mut address = Address::default();
                                address.full_address = Some(adro);
                                record.address = Some(address);
                            } else if let Some(ref mut address) = record.address {
                                // Дополняем существующий адрес полным адресом из АдрРО
                                if address.full_address.is_none() {
                                    address.full_address = Some(adro);
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
                    else if tag_matches(tag, "СвЗапись".as_bytes()) || tag_matches(tag, "СвЗапЕГРИП".as_bytes()) {
                        if let Ok(history_record) = self.parse_history_record(reader, e) {
                            record.history.push(history_record);
                        }
                        depth -= 1;
                    }
                    // Email
                    else if tag_matches(tag, "СведКонт".as_bytes()) || tag_matches(tag, "Контакт".as_bytes()) {
                        if let Some(email) = e.get_attr("E-mail".as_bytes())
                            .or_else(|| e.get_attr(attr_names::EMAIL)) {
                            record.email = Some(email);
                        }
                    }
                    // Прекращение
                    else if tag_matches(tag, "СвПрекр".as_bytes()) {
                        record.termination_date = e.get_date_attr("ДатаПрекр".as_bytes())
                            .or_else(|| e.get_date_attr("Дата".as_bytes()));
                        record.termination_method = e.get_attr("СпосПрекр".as_bytes());
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    // ФИО (empty element)
                    if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИОИП".as_bytes()) || tag_matches(tag, "СвФЛ".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        record.person = self.parse_person(e);
                    }
                    // ОГРНИП в элементе СвРегИП (empty element)
                    else if tag_matches(tag, "СвРегИП".as_bytes()) {
                        if let Some(ogrnip) = e.get_attr("ОГРНИП".as_bytes()) {
                            record.ogrnip = ogrnip;
                        }
                        if let Some(date) = e.get_date_attr("ДатаОГРНИП".as_bytes()) {
                            record.ogrnip_date = Some(date);
                        }
                    }
                    // Пол
                    else if tag_matches(tag, "Пол".as_bytes()) {
                        if let Some(gender_code) = e.get_attr("КодПол".as_bytes())
                            .or_else(|| e.get_attr("Пол".as_bytes())) {
                            record.gender = Some(Gender::from_code(&gender_code));
                        }
                    }
                    // Гражданство
                    else if tag_matches(tag, "СвГражд".as_bytes()) {
                        record.citizenship = Some(self.parse_citizenship(e));
                    }
                    // Статус ИП (СвСтатус как empty element)
                    else if tag_matches(tag, "СвСтатус".as_bytes()) {
                        let code = e.get_attr("КодСтатус".as_bytes());
                        let term_date = e.get_date_attr("ДатаПрекращ".as_bytes());

                        if let Some(c) = code {
                            record.status_code = Some(c);
                        }
                        if let Some(d) = term_date {
                            record.termination_date = Some(d);
                        }
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
                    else if tag_matches(tag, "СведКонт".as_bytes()) {
                        if let Some(email) = e.get_attr("E-mail".as_bytes()) {
                            record.email = Some(email);
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

        // Устанавливаем дату регистрации из ОГРНИП даты если не задана
        if record.registration_date.is_none() {
            record.registration_date = record.ogrnip_date;
        }

        Ok(record)
    }

    /// Парсинг СвФЛ (сведения о физлице)
    fn parse_sv_fl(&self, reader: &mut Reader<&[u8]>, record: &mut EgripRecord) -> Result<()> {
        let mut buf = Vec::new();
        let mut depth = 1;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        record.person = self.parse_person(e);
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let name = e.name();
                    let tag = name.as_ref();
                    
                    if tag_matches(tag, "ФИОРус".as_bytes()) || tag_matches(tag, "ФИО".as_bytes()) {
                        record.person = self.parse_person(e);
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

    /// Парсинг гражданства
    fn parse_citizenship(&self, e: &BytesStart) -> CitizenshipInfo {
        let citizenship_type = e.get_attr("ВидГражд".as_bytes())
            .or_else(|| e.get_attr(attr_names::VID_GRAZD))
            .map(|s| CitizenshipType::from_code(&s))
            .unwrap_or_default();
        
        CitizenshipInfo {
            citizenship_type,
            country_code: e.get_attr("ОКСМ".as_bytes())
                .or_else(|| e.get_attr(attr_names::KOD_OKSM)),
            country_name: e.get_attr("НаимСтран".as_bytes())
                .or_else(|| e.get_attr(attr_names::NAIM_STRAN)),
        }
    }

    /// Парсинг адреса
    fn parse_address(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<Address> {
        let mut address = Address::default();
        let mut buf = Vec::new();
        let mut depth = 1;

        // Из корневого элемента
        address.region_code = start.get_attr("КодРегион".as_bytes())
            .or_else(|| start.get_attr(attr_names::KOD_REGION));

        if let Some(full) = start.get_attr("Адрес".as_bytes()) {
            address.full_address = Some(normalize_string(&full));
        }

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    depth += 1;
                    self.process_address_element_ip(e, e.name().as_ref(), &mut address);
                }
                Ok(Event::Empty(ref e)) => {
                    self.process_address_element_ip(e, e.name().as_ref(), &mut address);
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

    /// Обработка элемента адреса для ИП
    fn process_address_element_ip(&self, e: &BytesStart, tag: &[u8], address: &mut Address) {
        if tag_matches(tag, "АдресРФ".as_bytes()) {
            address.postal_code = e.get_attr("Индекс".as_bytes());
            address.region_code = address.region_code.take().or_else(|| e.get_attr("КодРегион".as_bytes()));
            address.kladr_code = e.get_attr("КодАдрКладр".as_bytes())
                .or_else(|| e.get_attr("КодКладр".as_bytes()));
        }
        else if tag_matches(tag, "Регион".as_bytes()) {
            address.region = e.get_attr("НаимРегион".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
            address.region_code = address.region_code.take().or_else(|| e.get_attr("Код".as_bytes()));
        }
        // МуниципРайон - муниципальный район (ФИАС формат)
        else if tag_matches(tag, "МуниципРайон".as_bytes()) {
            if address.district.is_none() {
                address.district = e.get_attr("Наим".as_bytes())
                    .or_else(|| e.get_attr("Наименов".as_bytes()));
            }
        }
        else if tag_matches(tag, "Район".as_bytes()) {
            address.district = e.get_attr("НаимРайон".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        else if tag_matches(tag, "Город".as_bytes()) {
            address.city = e.get_attr("НаимГород".as_bytes())
                .or_else(|| e.get_attr("НаименГор".as_bytes()))
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        // НаселенПункт - добавлен атрибут Наим для ФИАС формата
        else if tag_matches(tag, "НаселенПункт".as_bytes()) || tag_matches(tag, "НаселПункт".as_bytes()) {
            address.locality = e.get_attr("НаименНП".as_bytes())
                .or_else(|| e.get_attr("Наим".as_bytes()))
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr("Наименование".as_bytes()));
        }
        else if tag_matches(tag, "Улица".as_bytes()) {
            let typ = e.get_attr("ТипУлица".as_bytes())
                .or_else(|| e.get_attr("Тип".as_bytes()));
            let name = e.get_attr("НаимУлица".as_bytes())
                .or_else(|| e.get_attr("Наименов".as_bytes()))
                .or_else(|| e.get_attr("Наименование".as_bytes()));

            address.street = match (typ, name) {
                (Some(t), Some(n)) => Some(format!("{} {}", t, n)),
                (None, Some(n)) => Some(n),
                (Some(t), None) => Some(t),
                (None, None) => None,
            };
        }
        // ЭлУлДорСети - улица в ФИАС формате (Тип + Наим)
        else if tag_matches(tag, "ЭлУлДорСети".as_bytes()) {
            if address.street.is_none() {
                let typ = e.get_attr("Тип".as_bytes());
                let name = e.get_attr("Наим".as_bytes());

                address.street = match (typ, name) {
                    (Some(t), Some(n)) => Some(format!("{} {}", t, n)),
                    (None, Some(n)) => Some(n),
                    (Some(t), None) => Some(t),
                    (None, None) => None,
                };
            }
        }
        // Здание - дом в ФИАС формате (Тип + Номер)
        else if tag_matches(tag, "Здание".as_bytes()) {
            if address.house.is_none() {
                let typ = e.get_attr("Тип".as_bytes());
                let num = e.get_attr("Номер".as_bytes());

                address.house = match (typ, num) {
                    (Some(t), Some(n)) => Some(format!("{}{}", t, n)),
                    (None, Some(n)) => Some(n),
                    (Some(t), None) => Some(t),
                    (None, None) => None,
                };
            }
        }
        // ПомещЗдания - помещение в ФИАС формате
        else if tag_matches(tag, "ПомещЗдания".as_bytes()) {
            if address.flat.is_none() {
                let typ = e.get_attr("Тип".as_bytes());
                let num = e.get_attr("Номер".as_bytes());

                address.flat = match (typ, num) {
                    (Some(t), Some(n)) => Some(format!("{}{}", t, n)),
                    (None, Some(n)) => Some(n),
                    (Some(t), None) => Some(t),
                    (None, None) => None,
                };
            }
        }
        // ПомещКвартиры - комната/офис в ФИАС формате
        else if tag_matches(tag, "ПомещКвартиры".as_bytes()) {
            let typ = e.get_attr("Тип".as_bytes());
            let num = e.get_attr("Номер".as_bytes());

            let room = match (typ, num) {
                (Some(t), Some(n)) => Some(format!("{}{}", t, n)),
                (None, Some(n)) => Some(n),
                (Some(t), None) => Some(t),
                (None, None) => None,
            };

            if let Some(r) = room {
                address.flat = match address.flat.take() {
                    Some(existing) => Some(format!("{}, {}", existing, r)),
                    None => Some(r),
                };
            }
        }
        // Дом, корпус, квартира - из атрибутов (классический формат)
        address.house = address.house.take().or_else(|| e.get_attr("Дом".as_bytes()));
        address.building = address.building.take().or_else(|| e.get_attr("Корп".as_bytes())
            .or_else(|| e.get_attr("Корпус".as_bytes())));
        address.flat = address.flat.take().or_else(|| e.get_attr("Кварт".as_bytes())
            .or_else(|| e.get_attr("Квартира".as_bytes())));
        address.fias_id = address.fias_id.take().or_else(|| e.get_attr("ИдНом".as_bytes())
            .or_else(|| e.get_attr("ФИАС".as_bytes())));
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
    fn parse_okvad(&self, reader: &mut Reader<&[u8]>, _start: &BytesStart, record: &mut EgripRecord) -> Result<()> {
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

    /// Парсинг регистрации
    fn parse_registration(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<IpRegistrationInfo> {
        let mut reg = IpRegistrationInfo::default();
        
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
        let tax = TaxAuthority {
            code: start.get_attr("КодНО".as_bytes())
                .or_else(|| start.get_attr("СвНалОрг".as_bytes()))
                .unwrap_or_default(),
            name: start.get_attr("НаимНО".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes()))
                .unwrap_or_default(),
            registration_date: start.get_date_attr("ДатаПост".as_bytes())
                .or_else(|| start.get_date_attr("ДатаРег".as_bytes())),
        };

        self.skip_element(reader)?;
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
        let fss = SocialInsurance {
            reg_number: start.get_attr("РегНомФСС".as_bytes())
                .or_else(|| start.get_attr("РегНомер".as_bytes()))
                .unwrap_or_default(),
            registration_date: start.get_date_attr("ДатаРег".as_bytes()),
            authority_name: start.get_attr("НаимФСС".as_bytes())
                .or_else(|| start.get_attr("НаимОрг".as_bytes())),
        };

        self.skip_element(reader)?;
        Ok(fss)
    }

    /// Парсинг записи истории
    fn parse_history_record(&self, reader: &mut Reader<&[u8]>, start: &BytesStart) -> Result<HistoryRecord> {
        let mut record = HistoryRecord::default();
        
        // ГРНИП в EGRIP вместо ГРН
        record.grn = start.get_attr("ГРНИП".as_bytes())
            .or_else(|| start.get_attr("ГРН".as_bytes()))
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

impl Default for EgripXmlParser {
    fn default() -> Self {
        Self::new()
    }
}

