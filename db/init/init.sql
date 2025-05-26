-- Расширения
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================
-- ТАБЛИЦЫ
-- =====================

-- Пользователи
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
    email_verified BOOLEAN DEFAULT FALSE,
    is_superuser BOOLEAN NOT NULL DEFAULT FALSE 
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

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID    NOT NULL,
    component_id TEXT   NOT NULL,
    shop_id     INT     NOT NULL,
    old_price   NUMERIC NOT NULL,
    new_price   NUMERIC NOT NULL,
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
    );
CREATE INDEX ON notifications(user_id, is_read);

CREATE TABLE IF NOT EXISTS subscriptions (
                                             id           SERIAL PRIMARY KEY,
                                             user_id      UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    component_id TEXT    NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(user_id, component_id)
    );

-- индекс для быстрого поиска подписчиков по компоненту
CREATE INDEX idx_subscriptions_by_component
    ON subscriptions(component_id);
-- для быстрого поиска всех подписок пользователя
CREATE INDEX IF NOT EXISTS idx_subscriptions_by_user
    ON subscriptions (user_id);

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


INSERT INTO shops (code, name, base_url, api_endpoint)
VALUES
    ('DNS', 'DNS Shop', 'https://www.dns-shop.ru', NULL)
;
INSERT INTO shops (code, name, base_url, api_endpoint)
VALUES
    ('Citilink', 'Citilink Shop', 'https://www.citilink.ru/', NULL)
;
INSERT INTO shops (code, name, base_url, api_endpoint)
VALUES
    ('Regard', 'Regard Shop', 'https://www.regard.ru/', NULL)
;
INSERT INTO shops (code, name, base_url, api_endpoint)
VALUES
    ('Nix', 'Nix Shop', 'https://www.nix.ru/', NULL)
;

-- CPUs: добавляем power_draw равный TDP
UPDATE components
SET specs = specs || '{"power_draw": 65}'::jsonb
WHERE id = 1;

UPDATE components
SET specs = specs || '{"power_draw": 65}'::jsonb
WHERE id = 2;

UPDATE components
SET specs = specs || '{"power_draw": 105}'::jsonb
WHERE id = 3;

-- GPUs: добавляем интерфейс и высоту карты
UPDATE components
SET specs = specs || '{"interface": "PCIe 4.0", "height_mm": 40}'::jsonb
WHERE id = 4;

UPDATE components
SET specs = specs || '{"interface": "PCIe 4.0", "height_mm": 43}'::jsonb
WHERE id = 5;

-- Motherboards: max память, слоты, PCIe, M.2 и SATA
UPDATE components
SET specs = specs || '{
  "max_memory_gb": 128,
  "memory_slots": 4,
  "pcie_version": "PCIe 4.0",
  "m2_slots": 2,
  "sata_ports": 6
}'::jsonb
WHERE id = 6;

UPDATE components
SET specs = specs || '{
  "max_memory_gb": 128,
  "memory_slots": 4,
  "pcie_version": "PCIe 5.0",
  "m2_slots": 2,
  "sata_ports": 6
}'::jsonb
WHERE id = 7;

-- RAM: количество модулей и напряжение
UPDATE components
SET specs = specs || '{"modules": 2, "voltage": 1.35}'::jsonb
WHERE id IN (8, 9);

UPDATE components
SET specs = specs || '{"modules": 2, "voltage": 1.25}'::jsonb
WHERE id = 10;

-- SSD: ёмкость, ключ M.2 и пропускная способность
UPDATE components
SET specs = specs || '{"capacity_gb": 1000, "m2_key": "M", "max_throughput": 7000}'::jsonb
WHERE id = 11;

UPDATE components
SET specs = specs || '{"capacity_gb": 1000, "m2_key": "M", "max_throughput": 5000}'::jsonb
WHERE id = 12;

UPDATE components
SET specs = specs || '{"capacity_gb": 500,  "m2_key": "M", "max_throughput": 3400}'::jsonb
WHERE id = 13;

UPDATE components
SET specs = specs || '{"capacity_gb": 500,  "m2_key": "M", "max_throughput": 2200}'::jsonb
WHERE id = 14;

-- HDD: ёмкость
UPDATE components
SET specs = specs || '{"capacity_gb": 4000}'::jsonb
WHERE id = 15;

UPDATE components
SET specs = specs || '{"capacity_gb": 4000}'::jsonb
WHERE id = 16;

-- PSUs: эффективность, модульность и форм-фактор
UPDATE components
SET specs = specs || '{"efficiency": "80 Plus Gold", "modular": true,  "form_factor": "ATX"}'::jsonb
WHERE id IN (17, 18, 19);

UPDATE components
SET specs = specs || '{"efficiency": "80 Plus Gold", "modular": false, "form_factor": "SFX"}'::jsonb
WHERE id = 23;

-- Cases: поддержка форм-факторов плат, длина блока питания, и корзины для дисков
UPDATE components
SET specs = specs || '{
  "max_motherboard_form_factors": ["ATX", "Micro-ATX", "Mini-ITX"],
  "max_psu_length": 200,
  "psu_form_factor": "ATX",
  "drive_bays_2_5": 2,
  "drive_bays_3_5": 2
}'::jsonb
WHERE id IN (20, 22);

UPDATE components
SET specs = specs || '{
  "max_motherboard_form_factors": ["Mini-ITX"],
  "max_psu_length": 130,
  "psu_form_factor": "SFX",
  "drive_bays_2_5": 1,
  "drive_bays_3_5": 0
}'::jsonb
WHERE id = 21;


-- пример вставки вручную
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  (
    'BeQuiet Shadow Rock 3',
    'cooler',
    'BeQuiet',
    '{"socket":"AM4","height_mm":160}',
    NOW(), NOW()
  ),
  (
    'Noctua NH-D15',
    'cooler',
    'Noctua',
    '{"socket":"AM4","height_mm":165}',
    NOW(), NOW()
  );

-- =====================
-- ДОБАВЛЕНИЕ СОВРЕМЕННЫХ КОМПОНЕНТОВ (РАСШИРЕНО ДО 15+ КАЖДОЙ КАТЕГОРИИ)
-- =====================

-- CPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Ryzen 5 5600X', 'cpu', 'AMD', '{"cores":6,"threads":12,"socket":"AM4","tdp":65,"power_draw":65,"cooler_height":158}'),
  ('Intel Core i5-12400', 'cpu', 'Intel', '{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"power_draw":65,"cooler_height":145}'),
  ('AMD Ryzen 7 7700X', 'cpu', 'AMD', '{"cores":8,"threads":16,"socket":"AM5","tdp":105,"power_draw":105,"cooler_height":160}'),
  ('Intel Core i7-14700K', 'cpu', 'Intel', '{"cores":20,"threads":28,"socket":"LGA1700","tdp":125,"power_draw":125,"cooler_height":160}'),
  ('AMD Ryzen 9 7950X', 'cpu', 'AMD', '{"cores":16,"threads":32,"socket":"AM5","tdp":170,"power_draw":170,"cooler_height":165}');

-- GPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 3060', 'gpu', 'NVIDIA', '{"length_mm":242,"power_draw":170,"memory_gb":12,"interface":"PCIe 4.0","height_mm":40}'),
  ('AMD Radeon RX 6700 XT', 'gpu', 'AMD', '{"length_mm":267,"power_draw":230,"memory_gb":12,"interface":"PCIe 4.0","height_mm":43}'),
  ('NVIDIA GeForce RTX 4070 SUPER', 'gpu', 'NVIDIA', '{"length_mm":267,"power_draw":220,"memory_gb":12,"interface":"PCIe 4.0","height_mm":45}'),
  ('AMD Radeon RX 7800 XT', 'gpu', 'AMD', '{"length_mm":276,"power_draw":263,"memory_gb":16,"interface":"PCIe 4.0","height_mm":48}'),
  ('NVIDIA GeForce RTX 4090', 'gpu', 'NVIDIA', '{"length_mm":304,"power_draw":450,"memory_gb":24,"interface":"PCIe 4.0","height_mm":60}');

-- Motherboards
INSERT INTO components (name, category, brand, specs) VALUES
  ('MSI B550 Tomahawk', 'motherboard', 'MSI', '{"socket":"AM4","ram_type":"DDR4","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":6}'),
  ('ASUS PRIME B660-PLUS', 'motherboard', 'ASUS', '{"socket":"LGA1700","ram_type":"DDR4","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":6}'),
  ('ASRock B650M Pro RS WiFi', 'motherboard', 'ASRock', '{"socket":"AM5","ram_type":"DDR5","form_factor":"Micro-ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":4}'),
  ('Gigabyte Z790 AORUS Elite AX', 'motherboard', 'Gigabyte', '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":3,"sata_ports":6}'),
  ('MSI PRO B760M-A DDR5', 'motherboard', 'MSI', '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"Micro-ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":2,"sata_ports":4}');

-- RAM
INSERT INTO components (name, category, brand, specs) VALUES
  ('Corsair Vengeance LPX 16GB DDR4-3200', 'ram', 'Corsair', '{"ram_type":"DDR4","frequency":3200,"capacity":16,"modules":2,"voltage":1.35}'),
  ('Kingston Fury 32GB DDR4-3600', 'ram', 'Kingston', '{"ram_type":"DDR4","frequency":3600,"capacity":32,"modules":2,"voltage":1.35}'),
  ('G.Skill Trident Z5 RGB 32GB DDR5-6000', 'ram', 'G.Skill', '{"ram_type":"DDR5","frequency":6000,"capacity":32,"modules":2,"voltage":1.25}'),
  ('Corsair Vengeance 32GB DDR5-6000', 'ram', 'Corsair', '{"ram_type":"DDR5","frequency":6000,"capacity":32,"modules":2,"voltage":1.25}'),
  ('Kingston Fury Renegade 64GB DDR5-6400', 'ram', 'Kingston', '{"ram_type":"DDR5","frequency":6400,"capacity":64,"modules":2,"voltage":1.35}');

-- SSD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Samsung 980 Pro 1TB', 'ssd', 'Samsung', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":7000}'),
  ('WD Black SN770 1TB', 'ssd', 'WD', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":5000}'),
  ('Crucial P5 500GB', 'ssd', 'Crucial', '{"form_factor":"M.2","interface":"PCIe 3.0","capacity_gb":500,"m2_key":"M","max_throughput":3400}'),
  ('Kingston A2000 500GB', 'ssd', 'Kingston', '{"form_factor":"M.2","interface":"PCIe 3.0","capacity_gb":500,"m2_key":"M","max_throughput":2200}'),
  ('Samsung 990 Pro 2TB', 'ssd', 'Samsung', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":2000,"m2_key":"M","max_throughput":7400}');

-- PSU
INSERT INTO components (name, category, brand, specs) VALUES
  ('Seasonic Focus GX-650', 'psu', 'Seasonic', '{"power":650,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Corsair RM750', 'psu', 'Corsair', '{"power":750,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('be quiet! Pure Power 12 M 850W', 'psu', 'be quiet!', '{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Corsair RM850x SHIFT', 'psu', 'Corsair', '{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Cooler Master V850 SFX Gold', 'psu', 'Cooler Master', '{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"SFX"}');

-- Cases
INSERT INTO components (name, category, brand, specs) VALUES
  ('NZXT H510', 'case', 'NZXT', '{"form_factor":"ATX","gpu_max_length":325,"cooler_max_height":165,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],"max_psu_length":200,"psu_form_factor":"ATX","drive_bays_2_5":2,"drive_bays_3_5":2}'),
  ('Fractal Design North', 'case', 'Fractal Design', '{"form_factor":"ATX","gpu_max_length":355,"cooler_max_height":170,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],"max_psu_length":250,"psu_form_factor":"ATX","drive_bays_2_5":3,"drive_bays_3_5":2}'),
  ('Lian Li A4-H2O', 'case', 'Lian Li', '{"form_factor":"Mini-ITX","gpu_max_length":322,"cooler_max_height":55,"max_motherboard_form_factors":["Mini-ITX"],"max_psu_length":130,"psu_form_factor":"SFX","drive_bays_2_5":2,"drive_bays_3_5":0}');

-- Coolers
INSERT INTO components (name, category, brand, specs) VALUES
  ('BeQuiet Shadow Rock 3', 'cooler', 'BeQuiet', '{"socket":"AM4","height_mm":160}'),
  ('Noctua NH-D15', 'cooler', 'Noctua', '{"socket":"AM4","height_mm":165}'),
  ('Thermalright Peerless Assassin 120 SE', 'cooler', 'Thermalright', '{"socket":"AM5","height_mm":155}'),
  ('Noctua NH-U12A chromax.black', 'cooler', 'Noctua', '{"socket":"LGA1700","height_mm":158}');



-- CPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Ryzen 9 7900X', 'cpu', 'AMD', '{"cores":12,"threads":24,"socket":"AM5","tdp":170,"power_draw":170,"cooler_height":165}'),
  ('Intel Core i9-13900K', 'cpu', 'Intel', '{"cores":24,"threads":32,"socket":"LGA1700","tdp":125,"power_draw":125,"cooler_height":163}'),
  ('AMD Ryzen 5 8600G', 'cpu', 'AMD', '{"cores":6,"threads":12,"socket":"AM5","tdp":65,"power_draw":65,"cooler_height":155}');

-- GPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 4080 SUPER', 'gpu', 'NVIDIA', '{"length_mm":304,"power_draw":320,"memory_gb":16,"interface":"PCIe 4.0","height_mm":61}'),
  ('AMD Radeon RX 7900 XTX', 'gpu', 'AMD', '{"length_mm":287,"power_draw":355,"memory_gb":24,"interface":"PCIe 4.0","height_mm":50}'),
  ('Intel Arc A770', 'gpu', 'Intel', '{"length_mm":267,"power_draw":225,"memory_gb":16,"interface":"PCIe 4.0","height_mm":45}');

-- Motherboards
INSERT INTO components (name, category, brand, specs) VALUES
  ('MSI MAG X670E Tomahawk WiFi', 'motherboard', 'MSI', '{"socket":"AM5","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":3,"sata_ports":6}'),
  ('ASUS ROG STRIX Z790-F', 'motherboard', 'ASUS', '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":4,"sata_ports":6}');

-- RAM
INSERT INTO components (name, category, brand, specs) VALUES
  ('G.Skill Ripjaws S5 64GB DDR5-6000', 'ram', 'G.Skill', '{"ram_type":"DDR5","frequency":6000,"capacity":64,"modules":2,"voltage":1.35}');

-- SSD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Sabrent Rocket 4 Plus 2TB', 'ssd', 'Sabrent', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":2000,"m2_key":"M","max_throughput":7100}'),
  ('ADATA XPG Gammix S70 Blade 1TB', 'ssd', 'ADATA', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":7400}');

-- HDD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Toshiba N300 6TB', 'hdd', 'Toshiba', '{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":6000}');

-- PSU
INSERT INTO components (name, category, brand, specs) VALUES
  ('Seasonic PRIME TX-1000', 'psu', 'Seasonic', '{"power":1000,"efficiency":"80 Plus Titanium","modular":true,"form_factor":"ATX"}');

-- Case
INSERT INTO components (name, category, brand, specs) VALUES
  ('Phanteks Eclipse G500A', 'case', 'Phanteks', '{"form_factor":"ATX","gpu_max_length":435,"cooler_max_height":185,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],"max_psu_length":270,"psu_form_factor":"ATX","drive_bays_2_5":4,"drive_bays_3_5":3}');

-- Cooler
INSERT INTO components (name, category, brand, specs) VALUES
  ('DeepCool AK620 Digital', 'cooler', 'DeepCool', '{"socket":"AM5","height_mm":162}');

-- Ещё CPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('Intel Core i3-13100F', 'cpu', 'Intel', '{"cores":4,"threads":8,"socket":"LGA1700","tdp":58,"power_draw":58,"cooler_height":130}'),
  ('AMD Ryzen 3 4100', 'cpu', 'AMD', '{"cores":4,"threads":8,"socket":"AM4","tdp":65,"power_draw":65,"cooler_height":120}');

-- Ещё GPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce GTX 1660 SUPER', 'gpu', 'NVIDIA', '{"length_mm":229,"power_draw":125,"memory_gb":6,"interface":"PCIe 3.0","height_mm":38}'),
  ('AMD Radeon RX 6600', 'gpu', 'AMD', '{"length_mm":190,"power_draw":132,"memory_gb":8,"interface":"PCIe 4.0","height_mm":40}');

-- Ещё RAM
INSERT INTO components (name, category, brand, specs) VALUES
  ('TeamGroup T-Force Delta RGB 32GB DDR5-6400', 'ram', 'TeamGroup', '{"ram_type":"DDR5","frequency":6400,"capacity":32,"modules":2,"voltage":1.35}'),
  ('Patriot Viper Steel 16GB DDR4-3200', 'ram', 'Patriot', '{"ram_type":"DDR4","frequency":3200,"capacity":16,"modules":2,"voltage":1.35}');

-- Ещё SSD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Lexar NM790 1TB', 'ssd', 'Lexar', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":7400}'),
  ('Silicon Power UD90 1TB', 'ssd', 'Silicon Power', '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":5000}');

-- Ещё HDD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Seagate Barracuda 2TB', 'hdd', 'Seagate', '{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":2000}'),
  ('WD Blue 1TB', 'hdd', 'Western Digital', '{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":1000}');

-- Ещё PSU
INSERT INTO components (name, category, brand, specs) VALUES
  ('Thermaltake Toughpower GF3 1200W', 'psu', 'Thermaltake', '{"power":1200,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('MSI MPG A850G', 'psu', 'MSI', '{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}');

-- Ещё Case
INSERT INTO components (name, category, brand, specs) VALUES
  ('Cooler Master NR200P', 'case', 'Cooler Master', '{"form_factor":"Mini-ITX","gpu_max_length":330,"cooler_max_height":155,"max_motherboard_form_factors":["Mini-ITX"],"max_psu_length":160,"psu_form_factor":"SFX","drive_bays_2_5":3,"drive_bays_3_5":1}'),
  ('Phanteks Eclipse P400A', 'case', 'Phanteks', '{"form_factor":"ATX","gpu_max_length":420,"cooler_max_height":160,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],"max_psu_length":270,"psu_form_factor":"ATX","drive_bays_2_5":2,"drive_bays_3_5":2}');

-- Ещё Cooler
INSERT INTO components (name, category, brand, specs) VALUES
  ('Scythe Fuma 2 Rev.B', 'cooler', 'Scythe', '{"socket":"AM4","height_mm":155}'),
  ('ARCTIC Freezer 34 eSports DUO', 'cooler', 'ARCTIC', '{"socket":"LGA1700","height_mm":157}');
  -- =====================
-- ДОБАВЛЕНИЕ КОМПОНЕНТОВ ПОД ВСЕ СЦЕНАРИИ
-- =====================

-- CPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Ryzen 5 5600G', 'cpu', 'AMD', '{"cores":6,"threads":12,"socket":"AM4","tdp":65,"power_draw":65,"cooler_height":135}'),
  ('Intel Core i3-12100', 'cpu', 'Intel', '{"cores":4,"threads":8,"socket":"LGA1700","tdp":60,"power_draw":60,"cooler_height":130}'),
  ('AMD Ryzen 9 7950X3D', 'cpu', 'AMD', '{"cores":16,"threads":32,"socket":"AM5","tdp":120,"power_draw":120,"cooler_height":165}'),
  ('Intel Core i9-14900K', 'cpu', 'Intel', '{"cores":24,"threads":32,"socket":"LGA1700","tdp":125,"power_draw":125,"cooler_height":165}');

-- RAM
INSERT INTO components (name, category, brand, specs) VALUES
  ('TeamGroup Elite 8GB DDR4-2666', 'ram', 'TeamGroup', '{"ram_type":"DDR4","frequency":2666,"capacity":8,"modules":1,"voltage":1.2}'),
  ('Corsair Dominator Platinum 64GB DDR5-6200', 'ram', 'Corsair', '{"ram_type":"DDR5","frequency":6200,"capacity":64,"modules":2,"voltage":1.35}');

-- GPU
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce GTX 1650', 'gpu', 'NVIDIA', '{"length_mm":170,"power_draw":75,"memory_gb":4,"interface":"PCIe 3.0","height_mm":40}'),
  ('NVIDIA RTX 4080', 'gpu', 'NVIDIA', '{"length_mm":320,"power_draw":320,"memory_gb":16,"interface":"PCIe 4.0","height_mm":60}');

-- PSU
INSERT INTO components (name, category, brand, specs) VALUES
  ('FSP HV PRO 500W', 'psu', 'FSP', '{"power":500,"efficiency":"80 Plus","modular":false,"form_factor":"ATX"}'),
  ('Corsair HX1200i', 'psu', 'Corsair', '{"power":1200,"efficiency":"80 Plus Platinum","modular":true,"form_factor":"ATX"}');

-- SSD
INSERT INTO components (name, category, brand, specs) VALUES
  ('Kingston NV2 250GB', 'ssd', 'Kingston', '{"form_factor":"M.2","interface":"PCIe 3.0","capacity_gb":250,"m2_key":"M","max_throughput":2100}'),
  ('Sabrent Rocket 5.0 2TB', 'ssd', 'Sabrent', '{"form_factor":"M.2","interface":"PCIe 5.0","capacity_gb":2000,"m2_key":"M","max_throughput":14000}');

-- Case
INSERT INTO components (name, category, brand, specs) VALUES
  ('Cooler Master NR200P', 'case', 'Cooler Master', '{"form_factor":"Mini-ITX","gpu_max_length":330,"cooler_max_height":155,"max_motherboard_form_factors":["Mini-ITX"],"max_psu_length":160,"psu_form_factor":"SFX","drive_bays_2_5":2,"drive_bays_3_5":1}'),
  ('Thermaltake Core V21', 'case', 'Thermaltake', '{"form_factor":"Micro-ATX","gpu_max_length":350,"cooler_max_height":185,"max_motherboard_form_factors":["Micro-ATX","Mini-ITX"],"max_psu_length":200,"psu_form_factor":"ATX","drive_bays_2_5":3,"drive_bays_3_5":3}');

-- Cooler
INSERT INTO components (name, category, brand, specs) VALUES
  ('Arctic Freezer 34 eSports DUO', 'cooler', 'Arctic', '{"socket":"AM4","height_mm":157}'),
  ('Scythe Fuma 2 Rev.B', 'cooler', 'Scythe', '{"socket":"LGA1700","height_mm":154}');
