-- Миграция 002: Таблица пользователей и миграция на user_id
-- Описание: Создает таблицу users, таблицу favorites, и мигрирует
--           entity_subscriptions с user_email на user_id

-- ============================================================================
-- Таблица пользователей
-- ============================================================================

CREATE TABLE subscriptions.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt hash
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE NOT NULL,
    email_verification_token VARCHAR(255),
    email_verification_expires_at TIMESTAMP WITH TIME ZONE,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Индексы для users
CREATE INDEX idx_users_email ON subscriptions.users(email);
CREATE INDEX idx_users_verification_token ON subscriptions.users(email_verification_token)
    WHERE email_verification_token IS NOT NULL;
CREATE INDEX idx_users_reset_token ON subscriptions.users(password_reset_token)
    WHERE password_reset_token IS NOT NULL;

-- Триггер для updated_at
CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON subscriptions.users
    FOR EACH ROW
    EXECUTE FUNCTION subscriptions.update_updated_at_column();

-- Комментарии
COMMENT ON TABLE subscriptions.users IS 'Пользователи системы с аутентификацией';
COMMENT ON COLUMN subscriptions.users.password_hash IS 'Bcrypt hash пароля';
COMMENT ON COLUMN subscriptions.users.email_verified IS 'Подтвержден ли email';
COMMENT ON COLUMN subscriptions.users.email_verification_token IS 'Токен для подтверждения email (UUID)';
COMMENT ON COLUMN subscriptions.users.last_login_at IS 'Время последнего входа пользователя';

-- ============================================================================
-- Миграция существующих подписок на user_id
-- ============================================================================

-- Добавить колонку user_id в entity_subscriptions
ALTER TABLE subscriptions.entity_subscriptions
ADD COLUMN user_id UUID REFERENCES subscriptions.users(id) ON DELETE CASCADE;

-- Создать временных пользователей из существующих email в подписках
INSERT INTO subscriptions.users (email, password_hash, first_name, last_name, email_verified)
SELECT DISTINCT
    user_email,
    '$2a$10$PLACEHOLDER',  -- Временный hash, пользователь должен сбросить пароль
    'User',  -- Placeholder
    split_part(user_email, '@', 1),  -- Фамилия из email
    FALSE  -- Email не подтвержден
FROM subscriptions.entity_subscriptions
WHERE user_email IS NOT NULL
ON CONFLICT (email) DO NOTHING;

-- Связать подписки с пользователями
UPDATE subscriptions.entity_subscriptions
SET user_id = (
    SELECT id FROM subscriptions.users
    WHERE users.email = entity_subscriptions.user_email
);

-- Сделать user_id обязательным (после миграции данных)
ALTER TABLE subscriptions.entity_subscriptions
ALTER COLUMN user_id SET NOT NULL;

-- Обновить индексы
CREATE INDEX idx_subscriptions_user_id ON subscriptions.entity_subscriptions(user_id)
    WHERE is_active = TRUE;

-- Комментарий
COMMENT ON COLUMN subscriptions.entity_subscriptions.user_id IS 'ID пользователя (FK на users)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.user_email IS 'Email пользователя (deprecated, use user_id)';

-- ============================================================================
-- Таблица избранного
-- ============================================================================

CREATE TABLE subscriptions.favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES subscriptions.users(id) ON DELETE CASCADE,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('company', 'entrepreneur')),
    entity_id VARCHAR(15) NOT NULL,
    entity_name TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    CONSTRAINT unique_favorite_entity UNIQUE (user_id, entity_type, entity_id)
);

-- Индексы для favorites
CREATE INDEX idx_favorites_user_id ON subscriptions.favorites(user_id);
CREATE INDEX idx_favorites_entity ON subscriptions.favorites(entity_type, entity_id);
CREATE INDEX idx_favorites_created_at ON subscriptions.favorites(created_at DESC);

-- Комментарии
COMMENT ON TABLE subscriptions.favorites IS 'Избранные компании и ИП пользователей';
COMMENT ON COLUMN subscriptions.favorites.user_id IS 'ID пользователя (FK на users)';
COMMENT ON COLUMN subscriptions.favorites.notes IS 'Пользовательские заметки о компании/ИП';

-- ============================================================================
-- Вывод информации о миграции
-- ============================================================================

DO $$
DECLARE
    users_count INT;
    subscriptions_count INT;
BEGIN
    SELECT COUNT(*) INTO users_count FROM subscriptions.users;
    SELECT COUNT(*) INTO subscriptions_count FROM subscriptions.entity_subscriptions;

    RAISE NOTICE 'Миграция 002 завершена успешно!';
    RAISE NOTICE 'Создано пользователей: %', users_count;
    RAISE NOTICE 'Подписок с user_id: %', subscriptions_count;
    RAISE NOTICE 'Таблица favorites создана';
END $$;
