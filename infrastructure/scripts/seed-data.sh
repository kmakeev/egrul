#!/bin/bash
# ==============================================================================
# Seed Data Script - Загрузка тестовых данных
# ==============================================================================
# Парсит XML файлы из test/ директорий и импортирует их в ClickHouse
# ==============================================================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${CYAN}==>${NC} $1"
}

# Конфигурация
TEST_DIRS=("./test" "./test2" "./test3")
OUTPUT_DIR="./output/seed"
PARSER_BIN="./target/release/egrul-parser"

echo ""
echo "======================================"
echo "  Загрузка тестовых данных"
echo "======================================"
echo ""

# ==============================================================================
# 1. Проверка парсера
# ==============================================================================
log_step "Проверка наличия парсера..."

if [ ! -f "$PARSER_BIN" ]; then
  log_error "Парсер не найден: $PARSER_BIN"
  log_info "Запустите: make parser-build"
  exit 1
fi

log_success "Парсер найден: $PARSER_BIN"

# ==============================================================================
# 2. Создание выходной директории
# ==============================================================================
log_step "Подготовка выходной директории..."

mkdir -p "$OUTPUT_DIR"
log_info "Выходная директория: $OUTPUT_DIR"

# ==============================================================================
# 3. Поиск XML файлов
# ==============================================================================
log_step "Поиск XML файлов в test директориях..."

xml_files=()
xml_count=0

for dir in "${TEST_DIRS[@]}"; do
  if [ -d "$dir" ]; then
    log_info "Сканирование: $dir"

    # Поиск XML файлов
    while IFS= read -r file; do
      if [ -n "$file" ]; then
        xml_files+=("$file")
        ((xml_count++))
      fi
    done < <(find "$dir" -name "*.xml" 2>/dev/null || true)

    if [ ${#xml_files[@]} -gt 0 ]; then
      log_info "  → Найдено файлов: $xml_count"
    fi
  else
    log_warning "Директория не найдена: $dir"
  fi
done

if [ ${#xml_files[@]} -eq 0 ]; then
  log_error "XML файлы не найдены в test директориях"
  log_info "Проверьте наличие файлов в: ${TEST_DIRS[*]}"
  exit 1
fi

log_success "Всего найдено XML файлов: ${#xml_files[@]}"
echo ""

# ==============================================================================
# 4. Парсинг файлов
# ==============================================================================
log_step "Запуск парсера..."

# Получаем первый найденный файл для определения директории
first_file="${xml_files[0]}"
input_dir=$(dirname "$first_file")

log_info "Входная директория: $input_dir"
log_info "Выходная директория: $OUTPUT_DIR"
echo ""

# Запуск парсера
"$PARSER_BIN" \
  --input "$input_dir" \
  --output "$OUTPUT_DIR" \
  --workers 4 \
  --compression snappy

if [ $? -ne 0 ]; then
  log_error "Ошибка парсинга"
  exit 1
fi

log_success "Парсинг завершен"
echo ""

# Показываем созданные Parquet файлы
log_info "Созданные Parquet файлы:"
find "$OUTPUT_DIR" -name "*.parquet" -exec ls -lh {} \; | awk '{print "  →", $9, "("$5")"}'
echo ""

# ==============================================================================
# 5. Импорт в ClickHouse
# ==============================================================================
log_step "Импорт данных в ClickHouse..."

# Экспортируем переменные для import-data.sh
export DATA_DIR="$OUTPUT_DIR"
export HISTORY_MAX_MEMORY=${HISTORY_MAX_MEMORY:-2000000000}
export HISTORY_BUCKETS=${HISTORY_BUCKETS:-10}

log_info "Настройки импорта:"
log_info "  → DATA_DIR: $DATA_DIR"
log_info "  → HISTORY_MAX_MEMORY: $HISTORY_MAX_MEMORY"
log_info "  → HISTORY_BUCKETS: $HISTORY_BUCKETS"
echo ""

# Запуск импорта
if [ -f "./infrastructure/scripts/import-data.sh" ]; then
  ./infrastructure/scripts/import-data.sh
else
  log_error "Скрипт импорта не найден: ./infrastructure/scripts/import-data.sh"
  exit 1
fi

if [ $? -ne 0 ]; then
  log_error "Ошибка импорта"
  exit 1
fi

echo ""
echo "======================================"
log_success "Тестовые данные загружены успешно!"
echo "======================================"
echo ""

# Показываем статистику
log_info "Статистика загрузки:"
log_info "  → XML файлов обработано: ${#xml_files[@]}"
log_info "  → Parquet файлов создано: $(find "$OUTPUT_DIR" -name "*.parquet" | wc -l)"
log_info "  → Выходная директория: $OUTPUT_DIR"
echo ""

log_info "Для просмотра данных используйте:"
log_info "  make ch-shell"
log_info "  SELECT count() FROM companies;"
log_info "  SELECT count() FROM entrepreneurs;"
echo ""
