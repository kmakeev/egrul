#!/bin/bash
set -e

echo "🚀 Настройка проекта ЕГРЮЛ/ЕГРИП"
echo "================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция проверки команды
check_command() {
    if command -v "$1" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 установлен"
        return 0
    else
        echo -e "${RED}✗${NC} $1 не найден"
        return 1
    fi
}

echo ""
echo "📋 Проверка зависимостей..."
echo ""

# Проверка обязательных зависимостей
MISSING_DEPS=0

check_command "docker" || MISSING_DEPS=1
check_command "docker-compose" || check_command "docker compose" || MISSING_DEPS=1
check_command "pnpm" || MISSING_DEPS=1
check_command "cargo" || MISSING_DEPS=1
check_command "go" || MISSING_DEPS=1

if [ $MISSING_DEPS -eq 1 ]; then
    echo ""
    echo -e "${YELLOW}⚠️  Некоторые зависимости не установлены${NC}"
    echo "Установите недостающие инструменты и запустите скрипт снова"
    exit 1
fi

echo ""
echo "📦 Установка зависимостей Node.js..."
pnpm install

echo ""
echo "🦀 Проверка Rust проекта..."
cd parser
cargo check
cd ..

echo ""
echo "🐹 Проверка Go модулей..."
cd services/api-gateway
go mod tidy
cd ../search-service
go mod tidy
cd ../..

echo ""
echo "🐳 Сборка Docker образов..."
docker compose build

echo ""
echo -e "${GREEN}✅ Настройка завершена!${NC}"
echo ""
echo "Доступные команды:"
echo "  make dev          - запуск в режиме разработки"
echo "  make build        - сборка всех сервисов"
echo "  make test         - запуск тестов"
echo "  make docker-up    - запуск Docker контейнеров"
echo ""

