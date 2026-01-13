-- Добавление детализированных полей адреса для ИП
-- Миграция для приведения схемы ИП в соответствие с компаниями

-- Добавляем недостающие поля адреса в таблицу entrepreneurs
ALTER TABLE egrul.entrepreneurs 
ADD COLUMN IF NOT EXISTS street Nullable(String) COMMENT 'Улица' AFTER locality;

ALTER TABLE egrul.entrepreneurs 
ADD COLUMN IF NOT EXISTS house Nullable(String) COMMENT 'Дом' AFTER street;

ALTER TABLE egrul.entrepreneurs 
ADD COLUMN IF NOT EXISTS building Nullable(String) COMMENT 'Корпус/Строение' AFTER house;

ALTER TABLE egrul.entrepreneurs 
ADD COLUMN IF NOT EXISTS flat Nullable(String) COMMENT 'Квартира' AFTER building;

ALTER TABLE egrul.entrepreneurs 
ADD COLUMN IF NOT EXISTS office Nullable(String) COMMENT 'Офис' AFTER flat;

-- Комментарий к миграции
-- Эти поля необходимы для полного отображения адреса ИП в UI
-- и обеспечения единообразия с отображением адресов компаний