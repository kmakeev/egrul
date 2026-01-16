//! Модуль чтения XML с поддержкой windows-1251

use std::fs::File;
use std::io::{BufReader, Read};
use std::path::Path;

use encoding_rs::WINDOWS_1251;
use memmap2::Mmap;
use tracing::debug;

use crate::error::{Error, Result};

/// Читатель XML файлов с автоопределением кодировки
pub struct XmlReader {
    data: Vec<u8>,
    encoding: XmlEncoding,
}

/// Кодировка XML
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum XmlEncoding {
    Utf8,
    Windows1251,
    Unknown,
}

impl XmlReader {
    /// Создание читателя из файла
    pub fn from_file(path: &Path) -> Result<Self> {
        let file = File::open(path)?;
        let metadata = file.metadata()?;
        let file_size = metadata.len() as usize;

        // Для больших файлов используем memory-mapped I/O
        let data = if file_size > 10 * 1024 * 1024 {
            // > 10 MB
            debug!("Использование mmap для файла {} ({} MB)", 
                path.display(), file_size / 1024 / 1024);
            let mmap = unsafe { Mmap::map(&file)? };
            mmap.to_vec()
        } else {
            let mut reader = BufReader::new(file);
            let mut data = Vec::with_capacity(file_size);
            reader.read_to_end(&mut data)?;
            data
        };

        let encoding = Self::detect_encoding(&data);
        debug!("Определена кодировка: {:?}", encoding);

        Ok(Self { data, encoding })
    }

    /// Создание читателя из байтов
    pub fn from_bytes(data: Vec<u8>) -> Self {
        let encoding = Self::detect_encoding(&data);
        Self { data, encoding }
    }

    /// Определение кодировки XML
    fn detect_encoding(data: &[u8]) -> XmlEncoding {
        // Проверяем BOM
        if data.starts_with(&[0xEF, 0xBB, 0xBF]) {
            return XmlEncoding::Utf8;
        }

        // Ищем encoding в XML declaration
        if let Some(decl_end) = data.iter().position(|&b| b == b'>') {
            let decl = &data[..decl_end.min(200)];
            let decl_lower: Vec<u8> = decl.iter().map(|b| b.to_ascii_lowercase()).collect();

            if decl_lower.windows(12).any(|w| w == b"windows-1251") {
                return XmlEncoding::Windows1251;
            }
            if decl_lower.windows(5).any(|w| w == b"utf-8") {
                return XmlEncoding::Utf8;
            }
        }

        // Эвристика: проверяем наличие кириллицы в windows-1251
        // В windows-1251 кириллица: 0xC0-0xFF
        let cyrillic_win1251 = data.iter().filter(|&&b| b >= 0xC0).count();
        
        // В UTF-8 многобайтовые символы начинаются с 0xD0-0xD1 для кириллицы
        let utf8_multibyte = data.windows(2)
            .filter(|w| w[0] == 0xD0 || w[0] == 0xD1)
            .count();

        if cyrillic_win1251 > utf8_multibyte * 2 {
            XmlEncoding::Windows1251
        } else if utf8_multibyte > 0 {
            XmlEncoding::Utf8
        } else {
            XmlEncoding::Unknown
        }
    }

    /// Конвертация в UTF-8 строку
    pub fn read_to_string(&self) -> Result<String> {
        match self.encoding {
            XmlEncoding::Utf8 => {
                String::from_utf8(self.data.clone())
                    .map_err(|e| Error::encoding(format!("Invalid UTF-8: {}", e)))
            }
            XmlEncoding::Windows1251 | XmlEncoding::Unknown => {
                let (result, _, had_errors) = WINDOWS_1251.decode(&self.data);
                if had_errors {
                    debug!("Были ошибки при декодировании windows-1251");
                }
                
                // Заменяем encoding в XML declaration на UTF-8
                let mut content = result.into_owned();
                content = content.replace("windows-1251", "UTF-8");
                content = content.replace("WINDOWS-1251", "UTF-8");
                content = content.replace("Windows-1251", "UTF-8");
                
                Ok(content)
            }
        }
    }

    /// Получение сырых данных
    pub fn raw_data(&self) -> &[u8] {
        &self.data
    }

    /// Получение кодировки
    pub fn encoding(&self) -> XmlEncoding {
        self.encoding
    }
}

/// Итератор по XML событиям с потоковым чтением
pub struct StreamingXmlReader<R: Read> {
    reader: quick_xml::Reader<BufReader<R>>,
    buffer: Vec<u8>,
}

impl<R: Read> StreamingXmlReader<R> {
    pub fn new(reader: R) -> Self {
        let buf_reader = BufReader::new(reader);
        let mut xml_reader = quick_xml::Reader::from_reader(buf_reader);
        xml_reader.config_mut().trim_text(true);

        Self {
            reader: xml_reader,
            buffer: Vec::with_capacity(4096),
        }
    }

    /// Чтение следующего события
    pub fn read_event(&mut self) -> Result<quick_xml::events::Event<'_>> {
        self.buffer.clear();
        Ok(self.reader.read_event_into(&mut self.buffer)?)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encoding_detection_utf8() {
        let data = b"<?xml version=\"1.0\" encoding=\"UTF-8\"?><root></root>";
        let encoding = XmlReader::detect_encoding(data);
        assert_eq!(encoding, XmlEncoding::Utf8);
    }

    #[test]
    fn test_encoding_detection_windows1251() {
        let data = b"<?xml version=\"1.0\" encoding=\"windows-1251\"?><root></root>";
        let encoding = XmlReader::detect_encoding(data);
        assert_eq!(encoding, XmlEncoding::Windows1251);
    }

    #[test]
    fn test_encoding_detection_bom() {
        let data = b"\xEF\xBB\xBF<?xml version=\"1.0\"?><root></root>";
        let encoding = XmlReader::detect_encoding(data);
        assert_eq!(encoding, XmlEncoding::Utf8);
    }
}

