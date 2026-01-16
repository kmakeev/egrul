#!/bin/bash
# ==============================================================================
# PostgreSQL Initialization Script - Metadata Schema
# ==============================================================================
# Создание схемы метаданных для системных данных:
# - import_sessions: сессии импорта данных
# - users: пользователи для будущей аутентификации
# - api_keys: API ключи для доступа к системе
# ==============================================================================

set -e

echo "======================================"
echo "  PostgreSQL Initialization"
echo "======================================"
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_info "Инициализация PostgreSQL..."

# Создание схемы metadata
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  -- ===========================================================================
  -- Расширения
  -- ===========================================================================
  CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
  CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Для полнотекстового поиска

  log_info "Расширения установлены"

  -- ===========================================================================
  -- Схема для системных данных
  -- ===========================================================================
  CREATE SCHEMA IF NOT EXISTS metadata;

  log_info "Схема metadata создана"

  -- ===========================================================================
  -- Таблица сессий импорта
  -- ===========================================================================
  CREATE TABLE IF NOT EXISTS metadata.import_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_type VARCHAR(50) NOT NULL,  -- 'egrul', 'egrip', 'full'
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'running',  -- 'running', 'completed', 'failed'
    records_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}'::jsonb,  -- Дополнительные данные
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
  );

  -- ===========================================================================
  -- Таблица пользователей (для будущей аутентификации)
  -- ===========================================================================
  CREATE TABLE IF NOT EXISTS metadata.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',  -- 'admin', 'user', 'readonly'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );

  -- ===========================================================================
  -- Таблица API ключей
  -- ===========================================================================
  CREATE TABLE IF NOT EXISTS metadata.api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES metadata.users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,  -- Описательное имя ключа
    permissions JSONB DEFAULT '[]'::jsonb,  -- ['read', 'write', 'admin']
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP
  );

  log_info "Таблицы созданы"

  -- ===========================================================================
  -- Триггеры для updated_at
  -- ===========================================================================
  CREATE OR REPLACE FUNCTION metadata.update_updated_at_column()
  RETURNS TRIGGER AS \$\$
  BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
  END;
  \$\$ language 'plpgsql';

  CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON metadata.users
    FOR EACH ROW
    EXECUTE FUNCTION metadata.update_updated_at_column();

  log_info "Триггеры созданы"

  -- ===========================================================================
  -- Роль для read-only доступа
  -- ===========================================================================
  DO \$\$
  BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'egrul_reader') THEN
      CREATE ROLE egrul_reader LOGIN PASSWORD 'readonly123';
      log_info "Роль egrul_reader создана"
    END IF;
  END
  \$\$;

  -- Предоставление прав
  GRANT CONNECT ON DATABASE $POSTGRES_DB TO egrul_reader;
  GRANT USAGE ON SCHEMA metadata TO egrul_reader;
  GRANT SELECT ON ALL TABLES IN SCHEMA metadata TO egrul_reader;
  ALTER DEFAULT PRIVILEGES IN SCHEMA metadata GRANT SELECT ON TABLES TO egrul_reader;

  -- ===========================================================================
  -- Индексы для оптимизации запросов
  -- ===========================================================================
  CREATE INDEX IF NOT EXISTS idx_import_sessions_status
    ON metadata.import_sessions(status);

  CREATE INDEX IF NOT EXISTS idx_import_sessions_started_at
    ON metadata.import_sessions(started_at DESC);

  CREATE INDEX IF NOT EXISTS idx_import_sessions_type
    ON metadata.import_sessions(session_type);

  CREATE INDEX IF NOT EXISTS idx_users_email
    ON metadata.users(email);

  CREATE INDEX IF NOT EXISTS idx_users_username
    ON metadata.users(username);

  CREATE INDEX IF NOT EXISTS idx_users_is_active
    ON metadata.users(is_active);

  CREATE INDEX IF NOT EXISTS idx_api_keys_user_id
    ON metadata.api_keys(user_id);

  CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at
    ON metadata.api_keys(expires_at) WHERE expires_at IS NOT NULL;

  log_info "Индексы созданы"

  -- ===========================================================================
  -- Комментарии к таблицам
  -- ===========================================================================
  COMMENT ON SCHEMA metadata IS 'Системные метаданные для ЕГРЮЛ/ЕГРИП системы';

  COMMENT ON TABLE metadata.import_sessions IS
    'История сессий импорта данных из Parquet файлов';

  COMMENT ON TABLE metadata.users IS
    'Пользователи системы для аутентификации и авторизации';

  COMMENT ON TABLE metadata.api_keys IS
    'API ключи для доступа к системе через REST/GraphQL API';

EOSQL

echo ""
log_success "PostgreSQL инициализирован успешно"
log_success "Создана схема metadata с таблицами:"
log_success "  - import_sessions (история импорта)"
log_success "  - users (пользователи)"
log_success "  - api_keys (API ключи)"
log_success "  - egrul_reader роль (read-only)"
echo ""
