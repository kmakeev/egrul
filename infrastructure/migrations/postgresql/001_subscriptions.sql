-- Миграция 001: Схема для подписок на изменения контрагентов
-- Описание: Создает схему subscriptions с таблицами для управления подписками
--           пользователей и логированием отправленных уведомлений

-- Создание схемы
CREATE SCHEMA IF NOT EXISTS subscriptions;

-- Включение расширения для UUID (если еще не включено)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Комментарий к схеме
COMMENT ON SCHEMA subscriptions IS 'Схема для управления подписками на изменения в данных ЕГРЮЛ/ЕГРИП';

-- ============================================================================
-- Таблица: subscriptions.entity_subscriptions
-- Описание: Подписки пользователей на отслеживание изменений в компаниях и ИП
-- ============================================================================

CREATE TABLE subscriptions.entity_subscriptions (
    -- Основные поля
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_email VARCHAR(255) NOT NULL,  -- Email пользователя (временное решение вместо user_id)

    -- Отслеживаемая сущность
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('company', 'entrepreneur')),
    entity_id VARCHAR(15) NOT NULL,    -- OGRN (13 цифр) или ОГРНИП (15 цифр)
    entity_name TEXT NOT NULL,         -- Название компании или ФИО ИП (для отображения)

    -- Фильтры изменений (JSON)
    change_filters JSONB DEFAULT '{"status": true, "director": true, "founders": true, "address": true, "capital": true, "activities": true}'::jsonb NOT NULL,
    -- Пример структуры:
    -- {
    --   "status": true,      // Статус (ликвидация, реорганизация)
    --   "director": true,    // Смена руководителя
    --   "founders": true,    // Изменение учредителей
    --   "address": true,     // Изменение адреса
    --   "capital": true,     // Изменение уставного капитала
    --   "activities": true   // Изменение видов деятельности (ОКВЭД)
    -- }

    -- Каналы уведомлений (JSON)
    notification_channels JSONB DEFAULT '{"email": true}'::jsonb NOT NULL,
    -- Пример структуры (на будущее можно добавить telegram, websocket):
    -- {
    --   "email": true,
    --   "telegram": false,
    --   "websocket": true
    -- }

    -- Настройки уведомлений
    is_active BOOLEAN DEFAULT TRUE NOT NULL,                    -- Активна ли подписка
    notify_immediately BOOLEAN DEFAULT TRUE NOT NULL,            -- Отправлять уведомления сразу или батчем
    batch_frequency INTERVAL DEFAULT INTERVAL '1 day',           -- Частота батч-уведомлений (если notify_immediately = false)

    -- Метаданные
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    last_notified_at TIMESTAMP WITH TIME ZONE,                   -- Последнее отправленное уведомление

    -- Уникальность: один пользователь не может дважды подписаться на одну сущность
    CONSTRAINT unique_user_entity UNIQUE (user_email, entity_type, entity_id)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_subscriptions_user_email
    ON subscriptions.entity_subscriptions(user_email)
    WHERE is_active = TRUE;

CREATE INDEX idx_subscriptions_entity
    ON subscriptions.entity_subscriptions(entity_type, entity_id)
    WHERE is_active = TRUE;

CREATE INDEX idx_subscriptions_created_at
    ON subscriptions.entity_subscriptions(created_at DESC);

-- GIN индекс для поиска по JSON полям
CREATE INDEX idx_subscriptions_change_filters
    ON subscriptions.entity_subscriptions USING GIN (change_filters);

CREATE INDEX idx_subscriptions_channels
    ON subscriptions.entity_subscriptions USING GIN (notification_channels);

-- Комментарии
COMMENT ON TABLE subscriptions.entity_subscriptions IS 'Подписки пользователей на отслеживание изменений в компаниях и ИП';
COMMENT ON COLUMN subscriptions.entity_subscriptions.user_email IS 'Email пользователя (временное решение без полной аутентификации)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.entity_type IS 'Тип сущности: company (компания) или entrepreneur (ИП)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.entity_id IS 'ОГРН для компаний (13 цифр) или ОГРНИП для ИП (15 цифр)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.change_filters IS 'JSON с настройками типов изменений для отслеживания';
COMMENT ON COLUMN subscriptions.entity_subscriptions.notification_channels IS 'JSON с каналами для отправки уведомлений (email, telegram, websocket)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.is_active IS 'Активна ли подписка (можно временно отключить без удаления)';
COMMENT ON COLUMN subscriptions.entity_subscriptions.notify_immediately IS 'Отправлять уведомления сразу при обнаружении изменения или накапливать в батч';
COMMENT ON COLUMN subscriptions.entity_subscriptions.batch_frequency IS 'Частота отправки батч-уведомлений (если notify_immediately = false)';

-- ============================================================================
-- Таблица: subscriptions.notification_log
-- Описание: История отправленных уведомлений для аудита и отладки
-- ============================================================================

CREATE TABLE subscriptions.notification_log (
    -- Основные поля
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL,                          -- Ссылка на подписку
    change_event_id VARCHAR(100) NOT NULL,                  -- ID события изменения из ClickHouse

    -- Информация о сущности
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('company', 'entrepreneur')),
    entity_id VARCHAR(15) NOT NULL,
    entity_name TEXT NOT NULL,

    -- Информация об изменении
    change_type VARCHAR(50) NOT NULL,                       -- Тип изменения (status, director, founder_added, etc.)
    field_name VARCHAR(100),                                -- Название поля
    old_value TEXT,                                         -- Старое значение (JSON string)
    new_value TEXT,                                         -- Новое значение (JSON string)
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL,          -- Время детектирования изменения

    -- Информация об отправке уведомления
    channel VARCHAR(20) NOT NULL,                           -- Канал: email, telegram, websocket
    recipient VARCHAR(255) NOT NULL,                        -- Email или Telegram chat_id
    status VARCHAR(20) DEFAULT 'pending' NOT NULL CHECK (status IN ('pending', 'sent', 'failed')),

    -- Метаданные отправки
    sent_at TIMESTAMP WITH TIME ZONE,                       -- Время успешной отправки
    retry_count INTEGER DEFAULT 0 NOT NULL,                 -- Количество попыток отправки
    error_message TEXT,                                     -- Сообщение об ошибке (если failed)

    -- Метаданные записи
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    -- Внешние ключи
    CONSTRAINT fk_subscription
        FOREIGN KEY (subscription_id)
        REFERENCES subscriptions.entity_subscriptions(id)
        ON DELETE CASCADE,

    -- Уникальность: одно изменение = одно уведомление на подписку (для idempotency)
    CONSTRAINT unique_notification_per_change
        UNIQUE (subscription_id, change_event_id)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_notification_log_subscription
    ON subscriptions.notification_log(subscription_id, created_at DESC);

CREATE INDEX idx_notification_log_status
    ON subscriptions.notification_log(status, created_at DESC)
    WHERE status IN ('pending', 'failed');

CREATE INDEX idx_notification_log_entity
    ON subscriptions.notification_log(entity_type, entity_id, created_at DESC);

CREATE INDEX idx_notification_log_change_event
    ON subscriptions.notification_log(change_event_id);

CREATE INDEX idx_notification_log_sent_at
    ON subscriptions.notification_log(sent_at DESC NULLS LAST);

-- Партиционирование по дате (для больших объемов логов в будущем)
-- Раскомментировать если потребуется:
-- CREATE TABLE subscriptions.notification_log_2024_01
--     PARTITION OF subscriptions.notification_log
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Комментарии
COMMENT ON TABLE subscriptions.notification_log IS 'История отправленных уведомлений с детальным логированием';
COMMENT ON COLUMN subscriptions.notification_log.subscription_id IS 'ID подписки, для которой отправлено уведомление';
COMMENT ON COLUMN subscriptions.notification_log.change_event_id IS 'Уникальный ID события изменения из ClickHouse (для idempotency)';
COMMENT ON COLUMN subscriptions.notification_log.change_type IS 'Тип изменения: status, director, founder_added, address, capital, activity';
COMMENT ON COLUMN subscriptions.notification_log.channel IS 'Канал отправки: email, telegram, websocket';
COMMENT ON COLUMN subscriptions.notification_log.status IS 'Статус отправки: pending (ожидает), sent (отправлено), failed (ошибка)';
COMMENT ON COLUMN subscriptions.notification_log.retry_count IS 'Количество попыток отправки (для retry механизма)';
COMMENT ON COLUMN subscriptions.notification_log.error_message IS 'Текст ошибки при failed статусе';

-- ============================================================================
-- Триггеры для автоматического обновления updated_at
-- ============================================================================

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION subscriptions.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для entity_subscriptions
CREATE TRIGGER trigger_update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions.entity_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION subscriptions.update_updated_at_column();

-- Триггер для notification_log
CREATE TRIGGER trigger_update_notification_log_updated_at
    BEFORE UPDATE ON subscriptions.notification_log
    FOR EACH ROW
    EXECUTE FUNCTION subscriptions.update_updated_at_column();

-- ============================================================================
-- Функции для работы с подписками
-- ============================================================================

-- Функция: Получить активные подписки для конкретной сущности
CREATE OR REPLACE FUNCTION subscriptions.get_active_subscriptions_for_entity(
    p_entity_type VARCHAR,
    p_entity_id VARCHAR
)
RETURNS TABLE (
    subscription_id UUID,
    user_email VARCHAR,
    change_filters JSONB,
    notification_channels JSONB,
    last_notified_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        id,
        entity_subscriptions.user_email,
        entity_subscriptions.change_filters,
        entity_subscriptions.notification_channels,
        entity_subscriptions.last_notified_at
    FROM subscriptions.entity_subscriptions
    WHERE entity_subscriptions.entity_type = p_entity_type
      AND entity_subscriptions.entity_id = p_entity_id
      AND entity_subscriptions.is_active = TRUE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION subscriptions.get_active_subscriptions_for_entity IS
    'Возвращает все активные подписки для конкретной компании или ИП';

-- Функция: Проверить нужно ли отправлять уведомление о типе изменения
CREATE OR REPLACE FUNCTION subscriptions.should_notify_for_change(
    p_change_filters JSONB,
    p_change_type VARCHAR
)
RETURNS BOOLEAN AS $$
DECLARE
    filter_key VARCHAR;
BEGIN
    -- Маппинг типов изменений на ключи в change_filters
    filter_key := CASE
        WHEN p_change_type IN ('status_change', 'liquidation', 'reorganization') THEN 'status'
        WHEN p_change_type IN ('director_change', 'director_added', 'director_removed') THEN 'director'
        WHEN p_change_type IN ('founder_added', 'founder_removed', 'founder_share_change') THEN 'founders'
        WHEN p_change_type IN ('address_change') THEN 'address'
        WHEN p_change_type IN ('capital_change') THEN 'capital'
        WHEN p_change_type IN ('activity_added', 'activity_removed', 'main_activity_change') THEN 'activities'
        ELSE NULL
    END;

    -- Если тип не найден, не отправляем уведомление
    IF filter_key IS NULL THEN
        RETURN FALSE;
    END IF;

    -- Проверяем включен ли фильтр
    RETURN COALESCE((p_change_filters->>filter_key)::BOOLEAN, FALSE);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION subscriptions.should_notify_for_change IS
    'Проверяет нужно ли отправлять уведомление о данном типе изменения согласно фильтрам подписки';

-- ============================================================================
-- Вспомогательные представления (Views)
-- ============================================================================

-- Представление: Статистика по подпискам
CREATE OR REPLACE VIEW subscriptions.subscription_stats AS
SELECT
    entity_type,
    COUNT(*) AS total_subscriptions,
    COUNT(*) FILTER (WHERE is_active = TRUE) AS active_subscriptions,
    COUNT(*) FILTER (WHERE is_active = FALSE) AS inactive_subscriptions,
    COUNT(DISTINCT user_email) AS unique_users,
    MIN(created_at) AS first_subscription_at,
    MAX(created_at) AS last_subscription_at
FROM subscriptions.entity_subscriptions
GROUP BY entity_type;

COMMENT ON VIEW subscriptions.subscription_stats IS 'Статистика по подпискам (компании, ИП, активные/неактивные)';

-- Представление: Статистика по уведомлениям
CREATE OR REPLACE VIEW subscriptions.notification_stats AS
SELECT
    DATE(created_at) AS notification_date,
    channel,
    status,
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE retry_count > 0) AS retried_count,
    AVG(EXTRACT(EPOCH FROM (sent_at - created_at))) AS avg_delivery_time_seconds
FROM subscriptions.notification_log
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at), channel, status
ORDER BY notification_date DESC, channel, status;

COMMENT ON VIEW subscriptions.notification_stats IS 'Статистика по уведомлениям за последние 30 дней';

-- ============================================================================
-- Начальные данные и тестовые подписки (опционально)
-- ============================================================================

-- Для тестирования можно добавить тестовую подписку:
-- INSERT INTO subscriptions.entity_subscriptions (
--     user_email,
--     entity_type,
--     entity_id,
--     entity_name
-- ) VALUES (
--     'test@example.com',
--     'company',
--     '1234567890123',
--     'ООО ТЕСТОВАЯ КОМПАНИЯ'
-- );

-- ============================================================================
-- Права доступа (если нужно ограничить доступ)
-- ============================================================================

-- Для production окружения можно создать отдельные роли:
-- CREATE ROLE egrul_app_role;
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA subscriptions TO egrul_app_role;
-- GRANT USAGE ON SCHEMA subscriptions TO egrul_app_role;

-- CREATE ROLE egrul_notification_service_role;
-- GRANT SELECT, INSERT, UPDATE ON subscriptions.notification_log TO egrul_notification_service_role;
-- GRANT SELECT ON subscriptions.entity_subscriptions TO egrul_notification_service_role;
-- GRANT USAGE ON SCHEMA subscriptions TO egrul_notification_service_role;

-- ============================================================================
-- Информация о миграции
-- ============================================================================

COMMENT ON SCHEMA subscriptions IS
'Версия: 001
Дата: 2026-01-22
Автор: Claude Code
Описание: Создание базовой схемы для системы подписок на изменения в ЕГРЮЛ/ЕГРИП данных.
          Включает таблицы для подписок пользователей и логирования уведомлений.';
