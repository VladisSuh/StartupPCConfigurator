-- Расширения
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================
-- ТАБЛИЦЫ
-- =====================

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    refresh_token TEXT,
    refresh_token_expires_at TIMESTAMP,
    reset_token TEXT,
    reset_token_expires_at TIMESTAMP,
    verification_code TEXT,
    email_verified BOOLEAN DEFAULT FALSE
);

-- Компоненты
CREATE TABLE IF NOT EXISTS components (
                                          id SERIAL PRIMARY KEY,
                                          name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,   -- "cpu", "gpu", "motherboard" и т.д.
    brand VARCHAR(100),               -- производитель, напр. "Intel", "AMD", "NVIDIA"
    specs JSONB,                      -- дополнительные характеристики в формате JSON (кол-во ядер, частота, etc.)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Конфигурации пользователей
CREATE TABLE IF NOT EXISTS configurations (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Связующие таблицы «конфигурация — компонент»
CREATE TABLE IF NOT EXISTS configuration_components (
                                                        id SERIAL PRIMARY KEY,
                                                        config_id INT NOT NULL,
                                                        component_id INT NOT NULL,
                                                        category VARCHAR(100),            -- продублировано или получаем из components
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_config
    FOREIGN KEY (config_id)
    REFERENCES configurations(id)
    ON DELETE CASCADE,

    CONSTRAINT fk_component
    FOREIGN KEY (component_id)
    REFERENCES components(id)
    ON DELETE RESTRICT
    );

-- Магазины
CREATE TABLE IF NOT EXISTS shops (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    base_url TEXT,
    api_endpoint TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Предложения
CREATE TABLE IF NOT EXISTS offers (
    id SERIAL PRIMARY KEY,
    component_id INT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    availability TEXT,
    url TEXT,
    fetched_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (component_id, shop_id)
);

-- История цен
CREATE TABLE IF NOT EXISTS price_history (
    id BIGSERIAL PRIMARY KEY,
    component_id INT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    captured_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Задания на обновление
CREATE TABLE IF NOT EXISTS update_jobs (
    id BIGSERIAL PRIMARY KEY,
    shop_id INT REFERENCES shops(id),
    status TEXT NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    message TEXT
);

-- Сценарии использования
CREATE TABLE IF NOT EXISTS usecases (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- Индекс для быстрого поиска по категории
CREATE INDEX IF NOT EXISTS idx_components_category
    ON components (category);


CREATE INDEX IF NOT EXISTS idx_components_specs 
    ON components USING GIN (specs);


-- (Опционально) Индекс для поиска по имени/бренду:
CREATE INDEX IF NOT EXISTS idx_components_name_brand
    ON components (name, brand);



-- Индекс для быстрого поиска конфигураций конкретного пользователя:
CREATE INDEX IF NOT EXISTS idx_configurations_user
    ON configurations (user_id);

-- Индекс для быстрого доступа по config_id:
CREATE INDEX IF NOT EXISTS idx_configuration_components_config
    ON configuration_components (config_id);

-- Индекс по component_id (опционально):
CREATE INDEX IF NOT EXISTS idx_configuration_components_component
    ON configuration_components (component_id);



CREATE INDEX IF NOT EXISTS idx_offers_component
    ON offers (component_id);
CREATE INDEX IF NOT EXISTS idx_offers_price
    ON offers (price);


CREATE INDEX IF NOT EXISTS idx_price_history_component_time
    ON price_history (component_id, captured_at DESC);



-- =====================
-- СИД-ДАННЫЕ
-- =====================

-- Use Cases
INSERT INTO usecases (name, description) VALUES
  ('office',   'Офисный ПК для документов и почты'),
  ('htpc',     'Домашний медиаплеер в мини-корпусе'),
  ('gaming',   'Игровой ПК среднего–высокого уровня'),
  ('streamer', 'Станция для стриминга и игр'),
  ('design',   'Графический дизайн и фото-обработка'),
  ('video',    'Видеомонтаж и рендеринг'),
  ('cad',      '3D-моделирование и CAD'),
  ('dev',      'ПК для разработки и виртуализации'),
  ('enthusiast','High-End система с разгоном и RGB'),
  ('nas',      'Домашний сервер / NAS');

-- CPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Ryzen 5 5600X', 'cpu', 'AMD', '{"cores":6,"threads":12,"socket":"AM4","tdp":65,"cooler_height":158}'),
  ('Intel Core i5-12400', 'cpu', 'Intel', '{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"cooler_height":145}'),
  ('AMD Ryzen 7 7700X', 'cpu', 'AMD', '{"socket":"AM5","tdp":105,"cooler_height":160}');

-- GPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 3060', 'gpu', 'NVIDIA', '{"length_mm":242,"power_draw":170}'),
  ('AMD Radeon RX 6700 XT', 'gpu', 'AMD', '{"length_mm":267,"power_draw":230}');

-- Motherboards
INSERT INTO components (name, category, brand, specs) VALUES
  ('MSI B550 Tomahawk', 'motherboard', 'MSI', '{"socket":"AM4","ram_type":"DDR4","form_factor":"ATX"}'),
  ('ASUS PRIME B660-PLUS', 'motherboard', 'ASUS', '{"socket":"LGA1700","ram_type":"DDR4","form_factor":"ATX"}');

-- RAM
INSERT INTO components (name, category, brand, specs) VALUES
  ('Corsair Vengeance LPX 16GB DDR4-3200', 'ram', 'Corsair', '{"ram_type":"DDR4","frequency":3200,"capacity":16}'),
  ('Kingston Fury 32GB DDR4-3600', 'ram', 'Kingston', '{"ram_type":"DDR4","frequency":3600,"capacity":32}'),
  ('G.Skill Trident Z5 RGB 32GB DDR5-6000', 'ram', 'G.Skill', '{"ram_type":"DDR5","frequency":6000,"capacity":32}');

-- SSD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Samsung 980 Pro 1TB', 'ssd', 'Samsung', '{"form_factor":"M.2","interface":"PCIe 4.0"}'),
  ('WD Black SN770 1TB', 'ssd', 'WD', '{"form_factor":"M.2","interface":"PCIe 4.0"}'),
  ('Crucial P5 500GB', 'ssd', 'Crucial', '{"form_factor":"M.2","interface":"PCIe 3.0"}'),
  ('Kingston A2000 500GB', 'ssd', 'Kingston', '{"form_factor":"M.2","interface":"PCIe 3.0"}');

-- HDD
INSERT INTO components (name, category, brand, specs) VALUES
  ('WD Red Plus 4TB', 'hdd', 'Western Digital', '{"form_factor":"3.5","interface":"SATA III","rpm":5400}'),
  ('Seagate IronWolf 4TB', 'hdd', 'Seagate', '{"form_factor":"3.5","interface":"SATA III","rpm":5900}');

-- PSU
INSERT INTO components (name, category, brand, specs) VALUES
  ('Seasonic Focus GX-650', 'psu', 'Seasonic', '{"power":650}'),
  ('Corsair RM750', 'psu', 'Corsair', '{"power":750}'),
  ('be quiet! Pure Power 12 M 850W', 'psu', 'be quiet!', '{"power":850}');

-- Cases
INSERT INTO components (name, category, brand, specs) VALUES
  ('NZXT H510', 'case', 'NZXT', '{"gpu_max_length":325,"cooler_max_height":165}'),
  ('Mini-ITX TinyCase', 'case', 'TestBrand', '{"form_factor_support":"Mini-ITX","gpu_max_length":170,"cooler_max_height":120}');


-- 1) Добавляем memory_gb для GPU
UPDATE components
SET specs = '{"length_mm":242,"power_draw":170,"memory_gb":12}'
WHERE name = 'NVIDIA GeForce RTX 3060';

UPDATE components
SET specs = '{"length_mm":267,"power_draw":230,"memory_gb":12}'
WHERE name = 'AMD Radeon RX 6700 XT';

-- 2) Переименовываем form_factor_support → form_factor в корпусах
UPDATE components
SET specs = '{"form_factor":"ATX","gpu_max_length":325,"cooler_max_height":165}'
WHERE name = 'NZXT H510';

UPDATE components
SET specs = '{"form_factor":"Mini-ITX","gpu_max_length":170,"cooler_max_height":120}'
WHERE name = 'Mini-ITX TinyCase';


INSERT INTO components (name, category, brand, specs)
VALUES ('NZXT H510', 'case', 'NZXT',
        '{"form_factor":"ATX","gpu_max_length":325,"cooler_max_height":165}');
-- Пример базовой «офисной» модели на 450 Вт
INSERT INTO components (name, category, brand, specs)
VALUES ('Seasonic Core GM-450', 'psu', 'Seasonic', '{"power": 450}');
