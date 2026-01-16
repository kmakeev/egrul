// Integration тесты с реальными XML файлами
use std::fs;
use std::path::PathBuf;

#[test]
fn test_parse_real_xml_files_from_test3() {
    // Используем реальные XML файлы из директории test3/
    let test3_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .unwrap()
        .join("test3");

    // Проверяем что директория существует
    if !test3_dir.exists() {
        println!("⚠️ Directory test3/ not found, skipping integration test");
        return;
    }

    let xml_files: Vec<_> = fs::read_dir(&test3_dir)
        .expect("Failed to read test3 directory")
        .filter_map(|entry| entry.ok())
        .filter(|entry| {
            entry
                .path()
                .extension()
                .and_then(|ext| ext.to_str())
                .map(|ext| ext == "xml")
                .unwrap_or(false)
        })
        .collect();

    if xml_files.is_empty() {
        println!("⚠️ No XML files found in test3/, skipping test");
        return;
    }

    let mut total_records = 0;
    let mut files_parsed = 0;

    for file in xml_files {
        let path = file.path();
        println!("Testing file: {:?}", path);

        // Читаем файл с правильной кодировкой (windows-1251)
        let content = fs::read(&path).expect(&format!("Failed to read file {:?}", path));
        let (decoded, _, had_errors) = encoding_rs::WINDOWS_1251.decode(&content);

        if had_errors {
            println!("⚠️ Warning: Had decoding errors in file {:?}", path);
        }

        // Определяем парсер по имени файла
        if path.to_str().unwrap().contains("RIGFO") || path.to_str().unwrap().contains("RUGFO") {
            // ЕГРЮЛ файл
            let parser = egrul_parser::EgrulXmlParser::new();
            let result = parser.parse(&decoded);

            assert!(
                result.is_ok(),
                "Failed to parse EGRUL file {:?}: {:?}",
                path,
                result.err()
            );

            let records = result.unwrap();
            println!("  ✓ Parsed {} EGRUL records", records.len());

            // Проверяем что записи валидны
            for record in &records {
                assert!(!record.ogrn.is_empty(), "OGRN should not be empty");
                assert!(!record.inn.is_empty(), "INN should not be empty");
                assert!(!record.full_name.is_empty(), "Full name should not be empty");
            }

            total_records += records.len();
            files_parsed += 1;
        }
    }

    assert!(
        files_parsed > 0,
        "Expected to parse at least one file, but parsed 0"
    );
    println!(
        "\n✅ Successfully parsed {} files with {} total records",
        files_parsed, total_records
    );
}

#[test]
fn test_parse_specific_real_file() {
    // Тестируем конкретный известный файл если он существует
    let test_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .unwrap()
        .join("test3/VO_RUGFO_0000_9965_20251102_d17fdf23-7079-4b06-a02b-f74405c08baa.xml");

    if !test_file.exists() {
        println!("⚠️ Specific test file not found, skipping test");
        return;
    }

    let content = fs::read(&test_file).expect("Failed to read test file");
    let (decoded, _, _) = encoding_rs::WINDOWS_1251.decode(&content);

    let parser = egrul_parser::EgrulXmlParser::new();
    let result = parser.parse(&decoded);

    assert!(result.is_ok(), "Failed to parse specific test file");
    let records = result.unwrap();

    assert!(!records.is_empty(), "Expected at least one record");

    let first_record = &records[0];
    println!("First record parsed:");
    println!("  OGRN: {}", first_record.ogrn);
    println!("  INN: {}", first_record.inn);
    println!("  Name: {}", first_record.full_name);

    // Базовые проверки валидности
    assert!(
        first_record.ogrn.len() == 13 || first_record.ogrn.len() == 15,
        "OGRN should be 13 or 15 characters"
    );
    assert!(
        first_record.inn.len() == 10 || first_record.inn.len() == 12,
        "INN should be 10 or 12 characters"
    );
}

#[test]
fn test_windows1251_encoding_handling() {
    // Проверяем что мы корректно обрабатываем windows-1251 кодировку
    let xml_with_cyrillic = r#"<?xml version="1.0" encoding="windows-1251"?>
<Файл>
<Документ>
<СвЮЛ ОГРН="1234567890123" ИНН="7707083893">
    <СвНаимЮЛ НаимЮЛПолн="Тест кириллицы: АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"/>
</СвЮЛ>
</Документ>
</Файл>"#;

    // Эмулируем windows-1251 кодировку
    let encoded = encoding_rs::WINDOWS_1251
        .encode(xml_with_cyrillic)
        .0
        .into_owned();
    let (decoded, _, _) = encoding_rs::WINDOWS_1251.decode(&encoded);

    let parser = egrul_parser::EgrulXmlParser::new();
    let result = parser.parse(&decoded);

    assert!(result.is_ok());
    let records = result.unwrap();
    assert_eq!(records.len(), 1);

    let full_name = &records[0].full_name;
    assert!(
        full_name.contains("кириллицы"),
        "Should correctly decode cyrillic: {}",
        full_name
    );
    assert!(
        full_name.contains("АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"),
        "Should contain full cyrillic alphabet"
    );
}
