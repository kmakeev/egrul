//! Бенчмарки для парсера ЕГРЮЛ/ЕГРИП

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use std::path::Path;

use egrul_parser::{Parser, XmlReader};
use egrul_parser::parser::{EgrulXmlParser, EgripXmlParser};

/// Тестовый XML для бенчмарков
const TEST_EGRUL_XML: &str = r#"<?xml version="1.0" encoding="UTF-8"?>
<ФАЙЛ>
  <Документ>
    <СвЮЛ ДатаВып="2025-01-01" ОГРН="1234567890123" ДатаОГРН="2020-01-01" 
          ИНН="1234567890" КПП="123456789" СтатусЮЛ="актив">
      <СвНаим НаимПолн="ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ ТЕСТ" НаимСокр="ООО ТЕСТ"/>
      <СвАдресЮЛ>
        <АдресРФ Индекс="123456" КодРегион="77">
          <Регион Наименов="Москва"/>
          <Город НаименГор="Москва"/>
          <Улица НаименУлица="Тестовая"/>
        </АдресРФ>
      </СвАдресЮЛ>
      <СвУстКап СумКап="10000" НаимВал="рубль"/>
      <СвОКВЭД>
        <СвОКВЭДОсн КодОКВЭД="62.01" НаимОКВЭД="Разработка программного обеспечения"/>
        <СвОКВЭДДоп КодОКВЭД="62.02" НаимОКВЭД="Консультирование по IT"/>
      </СвОКВЭД>
    </СвЮЛ>
  </Документ>
</ФАЙЛ>"#;

const TEST_EGRIP_XML: &str = r#"<?xml version="1.0" encoding="UTF-8"?>
<ФАЙЛ>
  <Документ>
    <СвИП ДатаВып="2025-01-01" ОГРНИП="123456789012345" ИНН="123456789012" Статус="актив">
      <ФИОИП Фамилия="Иванов" Имя="Иван" Отчество="Иванович"/>
      <СвГражд ВидГражд="1" НаимСтран="Россия"/>
      <СвОКВЭД>
        <СвОКВЭДОсн КодОКВЭД="47.11" НаимОКВЭД="Торговля розничная"/>
      </СвОКВЭД>
    </СвИП>
  </Документ>
</ФАЙЛ>"#;

fn benchmark_egrul_parsing(c: &mut Criterion) {
    let mut group = c.benchmark_group("egrul_parsing");
    
    // Бенчмарк парсинга одной записи
    group.throughput(Throughput::Elements(1));
    group.bench_function("parse_single_record", |b| {
        let parser = EgrulXmlParser::new();
        b.iter(|| {
            black_box(parser.parse(TEST_EGRUL_XML).unwrap())
        })
    });

    // Бенчмарк с 100 записями
    let large_xml = generate_large_egrul_xml(100);
    group.throughput(Throughput::Elements(100));
    group.bench_function("parse_100_records", |b| {
        let parser = EgrulXmlParser::new();
        b.iter(|| {
            black_box(parser.parse(&large_xml).unwrap())
        })
    });

    group.finish();
}

fn benchmark_egrip_parsing(c: &mut Criterion) {
    let mut group = c.benchmark_group("egrip_parsing");
    
    group.throughput(Throughput::Elements(1));
    group.bench_function("parse_single_record", |b| {
        let parser = EgripXmlParser::new();
        b.iter(|| {
            black_box(parser.parse(TEST_EGRIP_XML).unwrap())
        })
    });

    group.finish();
}

fn benchmark_xml_reader(c: &mut Criterion) {
    let mut group = c.benchmark_group("xml_reader");
    
    // Бенчмарк декодирования
    let win1251_bytes = encoding_rs::WINDOWS_1251
        .encode(TEST_EGRUL_XML)
        .0
        .to_vec();
    
    group.bench_function("decode_windows1251", |b| {
        b.iter(|| {
            let reader = XmlReader::from_bytes(black_box(win1251_bytes.clone()));
            black_box(reader.read_to_string().unwrap())
        })
    });

    group.finish();
}

/// Генерация большого XML для тестов
fn generate_large_egrul_xml(count: usize) -> String {
    let mut xml = String::from(r#"<?xml version="1.0" encoding="UTF-8"?><ФАЙЛ>"#);
    
    for i in 0..count {
        xml.push_str(&format!(r#"
  <Документ>
    <СвЮЛ ДатаВып="2025-01-01" ОГРН="{:013}" ДатаОГРН="2020-01-01" 
          ИНН="{:010}" КПП="123456789" СтатусЮЛ="актив">
      <СвНаим НаимПолн="ТЕСТОВАЯ ОРГАНИЗАЦИЯ {}" НаимСокр="ТЕСТ {}"/>
      <СвОКВЭД>
        <СвОКВЭДОсн КодОКВЭД="62.01" НаимОКВЭД="Разработка ПО"/>
      </СвОКВЭД>
    </СвЮЛ>
  </Документ>"#, 
            1000000000000 + i, 
            1000000000 + i,
            i,
            i
        ));
    }
    
    xml.push_str("</ФАЙЛ>");
    xml
}

criterion_group!(
    benches,
    benchmark_egrul_parsing,
    benchmark_egrip_parsing,
    benchmark_xml_reader,
);

criterion_main!(benches);

