-- Инициализация базы данных ЕГРЮЛ/ЕГРИП

-- Расширения
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Схема для основных данных
CREATE SCHEMA IF NOT EXISTS egrul;

-- ==================== Юридические лица ====================

CREATE TABLE IF NOT EXISTS egrul.legal_entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ogrn VARCHAR(13) UNIQUE NOT NULL,
    inn VARCHAR(10) UNIQUE NOT NULL,
    kpp VARCHAR(9),
    full_name TEXT NOT NULL,
    short_name TEXT,
    registration_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для юр. лиц
CREATE INDEX IF NOT EXISTS idx_legal_entities_inn ON egrul.legal_entities(inn);
CREATE INDEX IF NOT EXISTS idx_legal_entities_ogrn ON egrul.legal_entities(ogrn);
CREATE INDEX IF NOT EXISTS idx_legal_entities_status ON egrul.legal_entities(status);
CREATE INDEX IF NOT EXISTS idx_legal_entities_full_name_trgm ON egrul.legal_entities USING gin(full_name gin_trgm_ops);

-- ==================== Адреса ====================

CREATE TABLE IF NOT EXISTS egrul.addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    entity_type VARCHAR(20) NOT NULL, -- legal_entity, entrepreneur
    postal_code VARCHAR(6),
    region VARCHAR(100),
    city VARCHAR(100),
    street VARCHAR(200),
    house VARCHAR(50),
    office VARCHAR(50),
    full_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_addresses_entity ON egrul.addresses(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_addresses_full_trgm ON egrul.addresses USING gin(full_address gin_trgm_ops);

-- ==================== Виды деятельности ====================

CREATE TABLE IF NOT EXISTS egrul.activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    entity_type VARCHAR(20) NOT NULL,
    code VARCHAR(10) NOT NULL,
    name TEXT NOT NULL,
    is_main BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activities_entity ON egrul.activities(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_activities_code ON egrul.activities(code);

-- ==================== Уставный капитал ====================

CREATE TABLE IF NOT EXISTS egrul.capitals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL UNIQUE,
    amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'RUB',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ==================== Руководители ====================

CREATE TABLE IF NOT EXISTS egrul.heads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    inn VARCHAR(12),
    position VARCHAR(200),
    start_date DATE,
    end_date DATE,
    is_current BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_heads_entity ON egrul.heads(entity_id);
CREATE INDEX IF NOT EXISTS idx_heads_inn ON egrul.heads(inn);

-- ==================== Учредители ====================

CREATE TABLE IF NOT EXISTS egrul.founders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    founder_type VARCHAR(20) NOT NULL, -- person, legal_entity
    -- Для физ. лиц
    last_name VARCHAR(100),
    first_name VARCHAR(100),
    middle_name VARCHAR(100),
    person_inn VARCHAR(12),
    -- Для юр. лиц
    org_name TEXT,
    org_ogrn VARCHAR(13),
    org_inn VARCHAR(10),
    -- Доля
    share_nominal DECIMAL(20, 2),
    share_percent DECIMAL(5, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_founders_entity ON egrul.founders(entity_id);

-- ==================== Индивидуальные предприниматели ====================

CREATE TABLE IF NOT EXISTS egrul.entrepreneurs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ogrnip VARCHAR(15) UNIQUE NOT NULL,
    inn VARCHAR(12) UNIQUE NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    registration_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entrepreneurs_inn ON egrul.entrepreneurs(inn);
CREATE INDEX IF NOT EXISTS idx_entrepreneurs_ogrnip ON egrul.entrepreneurs(ogrnip);
CREATE INDEX IF NOT EXISTS idx_entrepreneurs_status ON egrul.entrepreneurs(status);
CREATE INDEX IF NOT EXISTS idx_entrepreneurs_name_trgm ON egrul.entrepreneurs 
    USING gin((last_name || ' ' || first_name || ' ' || COALESCE(middle_name, '')) gin_trgm_ops);

-- ==================== Триггеры для updated_at ====================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_legal_entities_updated_at
    BEFORE UPDATE ON egrul.legal_entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entrepreneurs_updated_at
    BEFORE UPDATE ON egrul.entrepreneurs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_addresses_updated_at
    BEFORE UPDATE ON egrul.addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_capitals_updated_at
    BEFORE UPDATE ON egrul.capitals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_heads_updated_at
    BEFORE UPDATE ON egrul.heads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_founders_updated_at
    BEFORE UPDATE ON egrul.founders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================== Вьюхи ====================

CREATE OR REPLACE VIEW egrul.v_legal_entities_full AS
SELECT 
    le.id,
    le.ogrn,
    le.inn,
    le.kpp,
    le.full_name,
    le.short_name,
    le.registration_date,
    le.status,
    a.full_address AS address,
    c.amount AS capital_amount,
    c.currency AS capital_currency,
    h.last_name || ' ' || h.first_name || ' ' || COALESCE(h.middle_name, '') AS head_name,
    h.position AS head_position,
    le.created_at,
    le.updated_at
FROM egrul.legal_entities le
LEFT JOIN egrul.addresses a ON a.entity_id = le.id AND a.entity_type = 'legal_entity'
LEFT JOIN egrul.capitals c ON c.entity_id = le.id
LEFT JOIN egrul.heads h ON h.entity_id = le.id AND h.is_current = TRUE;

CREATE OR REPLACE VIEW egrul.v_entrepreneurs_full AS
SELECT 
    e.id,
    e.ogrnip,
    e.inn,
    e.last_name,
    e.first_name,
    e.middle_name,
    e.last_name || ' ' || e.first_name || ' ' || COALESCE(e.middle_name, '') AS full_name,
    e.registration_date,
    e.status,
    a.full_address AS address,
    e.created_at,
    e.updated_at
FROM egrul.entrepreneurs e
LEFT JOIN egrul.addresses a ON a.entity_id = e.id AND a.entity_type = 'entrepreneur';

COMMENT ON SCHEMA egrul IS 'Схема для данных ЕГРЮЛ/ЕГРИП';
COMMENT ON TABLE egrul.legal_entities IS 'Юридические лица';
COMMENT ON TABLE egrul.entrepreneurs IS 'Индивидуальные предприниматели';

