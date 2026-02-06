-- Миграция 003: Добавление is_read статуса для уведомлений
-- Цель: Поддержка отслеживания прочитанных/непрочитанных уведомлений для real-time системы

-- Добавить колонки is_read и read_at в notification_log
ALTER TABLE subscriptions.notification_log
ADD COLUMN IF NOT EXISTS is_read BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN IF NOT EXISTS read_at TIMESTAMP WITH TIME ZONE;

-- Создать индекс для быстрого поиска непрочитанных уведомлений пользователя
CREATE INDEX IF NOT EXISTS idx_notification_log_read_status
ON subscriptions.notification_log(recipient, is_read, created_at DESC);

-- Создать представление для непрочитанных уведомлений
CREATE OR REPLACE VIEW subscriptions.unread_notifications AS
SELECT *
FROM subscriptions.notification_log
WHERE is_read = FALSE
ORDER BY created_at DESC;

-- Добавить комментарии
COMMENT ON COLUMN subscriptions.notification_log.is_read IS 'Флаг прочитанности уведомления';
COMMENT ON COLUMN subscriptions.notification_log.read_at IS 'Время когда уведомление было отмечено как прочитанное';
COMMENT ON INDEX subscriptions.idx_notification_log_read_status IS 'Индекс для оптимизации запросов непрочитанных уведомлений';
COMMENT ON VIEW subscriptions.unread_notifications IS 'Представление непрочитанных уведомлений';
