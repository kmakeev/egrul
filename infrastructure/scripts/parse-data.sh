#!/bin/bash
set -e

# Скрипт для парсинга XML файлов ЕГРЮЛ/ЕГРИП

INPUT_DIR="${1:-./data/input}"
OUTPUT_DIR="${2:-./data/output}"

echo "🔄 Запуск парсера ЕГРЮЛ/ЕГРИП"
echo "Входная директория: $INPUT_DIR"
echo "Выходная директория: $OUTPUT_DIR"

# Проверка входной директории
if [ ! -d "$INPUT_DIR" ]; then
    echo "❌ Директория $INPUT_DIR не существует"
    exit 1
fi

# Создание выходной директории
mkdir -p "$OUTPUT_DIR"

# Подсчет файлов
FILE_COUNT=$(find "$INPUT_DIR" -name "*.xml" | wc -l | tr -d ' ')
echo "📁 Найдено XML файлов: $FILE_COUNT"

if [ "$FILE_COUNT" -eq 0 ]; then
    echo "⚠️  XML файлы не найдены в $INPUT_DIR"
    exit 0
fi

# Запуск парсера
echo "🚀 Запуск парсинга..."
cargo run --release --package egrul-parser -- \
    --input "$INPUT_DIR" \
    --output "$OUTPUT_DIR" \
    --format json

echo "✅ Парсинг завершен"
echo "📊 Результаты сохранены в $OUTPUT_DIR"

