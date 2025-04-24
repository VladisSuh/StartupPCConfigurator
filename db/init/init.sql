CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
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

CREATE TABLE IF NOT EXISTS components (
                                          id SERIAL PRIMARY KEY,
                                          name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,   -- "cpu", "gpu", "motherboard" и т.д.
    brand VARCHAR(100),               -- производитель, напр. "Intel", "AMD", "NVIDIA"
    specs JSONB,                      -- дополнительные характеристики в формате JSON (кол-во ядер, частота, etc.)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

-- Индекс для быстрого поиска по категории
CREATE INDEX IF NOT EXISTS idx_components_category
    ON components (category);


CREATE INDEX IF NOT EXISTS idx_components_specs 
    ON components USING GIN (specs);


-- (Опционально) Индекс для поиска по имени/бренду:
CREATE INDEX IF NOT EXISTS idx_components_name_brand
    ON components (name, brand);

CREATE TABLE IF NOT EXISTS configurations (
                                              id SERIAL PRIMARY KEY,
                                              user_id UUID NOT NULL,            -- или TEXT, если не хотите строго проверять формат
                                              name VARCHAR(255) NOT NULL,       -- название сборки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

-- Индекс для быстрого поиска конфигураций конкретного пользователя:
CREATE INDEX IF NOT EXISTS idx_configurations_user
    ON configurations (user_id);

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

-- Индекс для быстрого доступа по config_id:
CREATE INDEX IF NOT EXISTS idx_configuration_components_config
    ON configuration_components (config_id);

-- Индекс по component_id (опционально):
CREATE INDEX IF NOT EXISTS idx_configuration_components_component
    ON configuration_components (component_id);

-- ---------- shops ----------
CREATE TABLE IF NOT EXISTS shops (
                                     id           SERIAL PRIMARY KEY,
                                     code         TEXT UNIQUE NOT NULL,
                                     name         TEXT NOT NULL,
                                     base_url     TEXT,
                                     api_endpoint TEXT,
                                     is_active    BOOLEAN DEFAULT TRUE,
                                     created_at   TIMESTAMP DEFAULT now(),
    updated_at   TIMESTAMP DEFAULT now()
    );

-- ---------- offers ----------
CREATE TABLE IF NOT EXISTS offers (
                                      id            SERIAL PRIMARY KEY,
                                      component_id  INT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    shop_id       INT  NOT NULL REFERENCES shops(id)      ON DELETE CASCADE,
    price         NUMERIC(12,2) NOT NULL,
    currency      CHAR(3)  NOT NULL DEFAULT 'USD',
    availability  TEXT,
    url           TEXT,
    fetched_at    TIMESTAMP NOT NULL,
    UNIQUE (component_id, shop_id)
    );
CREATE INDEX IF NOT EXISTS idx_offers_component
    ON offers (component_id);
CREATE INDEX IF NOT EXISTS idx_offers_price
    ON offers (price);

-- ---------- price_history ----------
CREATE TABLE IF NOT EXISTS price_history (
                                             id           BIGSERIAL PRIMARY KEY,
                                             component_id TEXT NOT NULL,
                                             shop_id      INT  NOT NULL,
                                             price        NUMERIC(12,2) NOT NULL,
    currency     CHAR(3) NOT NULL DEFAULT 'USD',
    captured_at  TIMESTAMP NOT NULL DEFAULT now()
    );
CREATE INDEX IF NOT EXISTS idx_price_history_component_time
    ON price_history (component_id, captured_at DESC);

-- ---------- update_jobs (optional) ----------
CREATE TABLE IF NOT EXISTS update_jobs (
                                           id           BIGSERIAL PRIMARY KEY,
                                           shop_id      INT REFERENCES shops(id),
    status       TEXT NOT NULL,            -- queued | running | done | failed
    started_at   TIMESTAMP,
    finished_at  TIMESTAMP,
    message      TEXT
    );
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
ALTER TABLE users DROP COLUMN id;

ALTER TABLE users
    ADD COLUMN id uuid PRIMARY KEY DEFAULT gen_random_uuid();

-- CPU
INSERT INTO components (name, category, brand, specs)
VALUES
    ('AMD Ryzen 5 5600X', 'cpu', 'AMD', '{"cores": 6, "threads": 12, "socket": "AM4"}'),
    ('Intel Core i5-12400', 'cpu', 'Intel', '{"cores": 6, "threads": 12, "socket": "LGA1700"}');

-- GPU
INSERT INTO components (name, category, brand, specs)
VALUES
    ('NVIDIA GeForce RTX 3060', 'gpu', 'NVIDIA', '{"memory": "12GB", "base_clock": "1320MHz"}'),
    ('AMD Radeon RX 6700 XT', 'gpu', 'AMD', '{"memory": "12GB", "base_clock": "2321MHz"}');

-- Motherboard
INSERT INTO components (name, category, brand, specs)
VALUES
    ('MSI B550 Tomahawk', 'motherboard', 'MSI', '{"socket": "AM4", "form_factor": "ATX"}'),
    ('ASUS PRIME B660-PLUS', 'motherboard', 'ASUS', '{"socket": "LGA1700", "form_factor": "ATX"}');

-- RAM
INSERT INTO components (name, category, brand, specs)
VALUES
    ('Corsair Vengeance LPX 16GB DDR4-3200', 'ram', 'Corsair', '{"capacity": "16GB", "speed": "3200MHz"}'),
    ('Kingston Fury 32GB DDR4-3600', 'ram', 'Kingston', '{"capacity": "32GB", "speed": "3600MHz"}');

-- SSD
INSERT INTO components (name, category, brand, specs)
VALUES
    ('Samsung 970 EVO Plus 1TB', 'ssd', 'Samsung', '{"form_factor": "M.2", "interface": "PCIe 3.0"}'),
    ('WD Blue 1TB SSD', 'ssd', 'WD', '{"form_factor": "2.5", "interface": "SATA III"}');

-- PSU
INSERT INTO components (name, category, brand, specs)
VALUES
    ('Seasonic Focus GX-650', 'psu', 'Seasonic', '{"wattage": 650, "modular": true}'),
    ('Corsair RM750', 'psu', 'Corsair', '{"wattage": 750, "modular": true}');

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('AMD Ryzen 5 5600X', 'cpu', 'AMD', '{"socket": "AM4", "tdp": 65}', NOW(), NOW()),
    ('Intel Core i5-12400F', 'cpu', 'Intel', '{"socket": "LGA1700", "tdp": 65}', NOW(), NOW());

-- Motherboard
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('MSI B550 Tomahawk', 'motherboard', 'MSI', '{"socket": "AM4", "ram_type": "DDR4", "form_factor": "ATX"}', NOW(), NOW()),
    ('ASUS PRIME B660M', 'motherboard', 'ASUS', '{"socket": "LGA1700", "ram_type": "DDR4", "form_factor": "mATX"}', NOW(), NOW());

-- GPU
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('NVIDIA RTX 3060', 'gpu', 'NVIDIA', '{"length_mm": 242, "power_draw": 170}', NOW(), NOW()),
    ('AMD RX 6600 XT', 'gpu', 'AMD', '{"length_mm": 270, "power_draw": 160}', NOW(), NOW());

-- RAM
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('Corsair Vengeance 16GB DDR4-3200', 'ram', 'Corsair', '{"ram_type": "DDR4", "frequency": 3200, "capacity": 16}', NOW(), NOW()),
    ('Kingston Fury 32GB DDR4-3600', 'ram', 'Kingston', '{"ram_type": "DDR4", "frequency": 3600, "capacity": 32}', NOW(), NOW());

-- PSU
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('Seasonic Focus 650W', 'psu', 'Seasonic', '{"power": 650}', NOW(), NOW());

-- Case
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
    ('NZXT H510', 'case', 'NZXT', '{"form_factor_support": "ATX", "gpu_max_length": 325, "cooler_max_height": 165}', NOW(), NOW());


INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
           'ASRock X670E Taichi',
           'motherboard',
           'ASRock',
           '{"socket": "AM5", "ram_type": "DDR5", "form_factor": "ATX"}',
           NOW(), NOW()
       );

-- CPU с сокетом AM5
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
    ('AMD Ryzen 7 7700X', 'cpu', 'AMD', '{"socket": "AM5", "tdp": 105}', NOW(), NOW());

-- DDR5 память
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
    ('G.Skill Trident Z5 RGB 32GB DDR5-6000', 'ram', 'G.Skill', '{"ram_type": "DDR5", "frequency": 6000, "capacity": 32}', NOW(), NOW());

-- Совместимая материнская плата под AM5 + DDR5
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
    ('ASRock X670E Taichi', 'motherboard', 'ASRock', '{"socket": "AM5", "ram_type": "DDR5", "form_factor": "ATX"}', NOW(), NOW());

-- Доп. совместимый блок питания
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
    ('be quiet! Pure Power 12 M 850W', 'psu', 'be quiet!', '{"power": 850}', NOW(), NOW());



-- 1) Сверхдлинная видеокарта для проверки ограничения корпуса
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Test GPU XXL', 'gpu', 'TestBrand',
   '{"length_mm":400,"power_draw":300}', NOW(), NOW());

-- 2) CPU с очень высоким кулером (для проверки ограничения высоты)
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Test CPU TallCooler', 'cpu', 'TestBrand',
   '{"socket":"AM4","tdp":95,"cooler_height":200}', NOW(), NOW());

-- 3) Мини-корпус форм-фактора Mini-ITX с очень жёсткими габаритами
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Mini-ITX TinyCase', 'case', 'TestBrand',
   '{"form_factor_support":"Mini-ITX","gpu_max_length":170,"cooler_max_height":120}', NOW(), NOW());

-- 4) Маломощный блок питания (для проверки на слабую PSU)
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Test PSU 300W', 'psu', 'TestBrand',
   '{"power":300}', NOW(), NOW());

-- 5) RAM с другим типом (DDR3) для проверки несовместимости
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Test RAM DDR3 8GB', 'ram', 'TestBrand',
   '{"ram_type":"DDR3","frequency":1600,"capacity":8}', NOW(), NOW());

-- 6) Материнка под Mini-ITX, но без нужного слота под сверхвысокий кулер
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('Test MB Mini-ITX', 'motherboard', 'TestBrand',
   '{"socket":"AM4","ram_type":"DDR4","form_factor":"Mini-ITX"}', NOW(), NOW());



-- CPU
UPDATE components
SET specs = specs || '{"cooler_height":158}'
WHERE name = 'AMD Ryzen 5 5600X';

UPDATE components
SET specs = specs || '{"cooler_height":145}'
WHERE name = 'Intel Core i5-12400';

UPDATE components
SET specs = specs || '{"cooler_height":160}'
WHERE name = 'AMD Ryzen 7 7700X';  -- для AM5-модели

-- RAM
UPDATE components
SET specs = specs || '{"ram_type":"DDR4"}'
WHERE name = 'Corsair Vengeance LPX 16GB DDR4-3200';

UPDATE components
SET specs = specs || '{"ram_type":"DDR4"}'
WHERE name = 'Kingston Fury 32GB DDR4-3600';

UPDATE components
SET specs = specs || '{"ram_type":"DDR5"}'
WHERE name = 'G.Skill Trident Z5 RGB 32GB DDR5-6000';

-- GPU
UPDATE components
SET specs = specs || '{"length_mm":242,"power_draw":170}'
WHERE name = 'NVIDIA GeForce RTX 3060';

UPDATE components
SET specs = specs || '{"length_mm":267,"power_draw":230}'
WHERE name = 'AMD Radeon RX 6700 XT';

-- PSU
UPDATE components
SET specs = specs || '{"power":650}'
WHERE name = 'Seasonic Focus GX-650';

UPDATE components
SET specs = specs || '{"power":750}'
WHERE name = 'Corsair RM750';

UPDATE components
SET specs = specs || '{"power":850}'
WHERE name = 'be quiet! Pure Power 12 M 850W';

-- Case
UPDATE components
SET specs = specs || '{"gpu_max_length":325,"cooler_max_height":165}'
WHERE name = 'NZXT H510';

UPDATE components
SET specs = specs || '{"gpu_max_length":170,"cooler_max_height":120}'
WHERE name = 'Mini-ITX TinyCase';  -- если тестовый корпус добавлен


INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('NZXT H510', 'case', 'NZXT',
   '{"gpu_max_length":325,"cooler_max_height":165}', NOW(), NOW()),
  ('Mini-ITX TinyCase', 'case', 'TestBrand',
   '{"gpu_max_length":170,"cooler_max_height":120}', NOW(), NOW());




-- Таблица сценариев использования (Use Cases)
CREATE TABLE IF NOT EXISTS usecases (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,  -- например: office, gaming, designer и т.д.
    description TEXT                               -- краткое описание сценария
);

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


-- 1. Новые CPU
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Intel Core i9-12900K',   'cpu', 'Intel', '{"socket":"LGA1700","tdp":125,"cooler_height":165}', NOW(), NOW()),
  ('Intel Core i7-12700K',   'cpu', 'Intel', '{"socket":"LGA1700","tdp":125,"cooler_height":160}', NOW(), NOW()),
  ('AMD Ryzen 9 7950X',      'cpu', 'AMD',   '{"socket":"AM5","tdp":170,"cooler_height":165}', NOW(), NOW()),
  ('AMD Ryzen 5 7600',       'cpu', 'AMD',   '{"socket":"AM5","tdp":65,"cooler_height":158}', NOW(), NOW()),
  ('Intel Core i3-12100',    'cpu', 'Intel', '{"socket":"LGA1700","tdp":60,"cooler_height":140}', NOW(), NOW()),
  ('AMD Athlon 3000G',       'cpu', 'AMD',   '{"socket":"AM4","tdp":35,"cooler_height":120}', NOW(), NOW());

-- 2. Новые материнские платы
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Asus ROG Strix Z690-A',   'motherboard', 'ASUS',     '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"ATX"}', NOW(), NOW()),
  ('Gigabyte X670 Aorus Elite','motherboard','Gigabyte', '{"socket":"AM5","ram_type":"DDR5","form_factor":"ATX"}', NOW(), NOW()),
  ('MSI MAG B650M Mortar',    'motherboard', 'MSI',      '{"socket":"AM5","ram_type":"DDR5","form_factor":"mATX"}', NOW(), NOW()),
  ('Gigabyte B660M DS3H',     'motherboard', 'Gigabyte', '{"socket":"LGA1700","ram_type":"DDR4","form_factor":"mATX"}', NOW(), NOW()),
  ('NZXT N7 B550',            'motherboard', 'NZXT',     '{"socket":"AM4","ram_type":"DDR4","form_factor":"ATX"}', NOW(), NOW()),
  ('Asus ROG Strix B660-I',   'motherboard', 'ASUS',     '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"Mini-ITX"}', NOW(), NOW());

-- 3. Новые модули RAM
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Corsair Dominator Platinum 32GB DDR5-5600', 'ram','Corsair','{"ram_type":"DDR5","frequency":5600,"capacity":32}', NOW(), NOW()),
  ('G.Skill Ripjaws S5 16GB DDR5-5200',         'ram','G.Skill','{"ram_type":"DDR5","frequency":5200,"capacity":16}', NOW(), NOW()),
  ('Crucial Ballistix 32GB DDR4-3600',          'ram','Crucial','{"ram_type":"DDR4","frequency":3600,"capacity":32}', NOW(), NOW()),
  ('Kingston HyperX Fury 16GB DDR4-2666',       'ram','Kingston','{"ram_type":"DDR4","frequency":2666,"capacity":16}', NOW(), NOW()),
  ('Patriot Viper Steel 8GB DDR4-3200',         'ram','Patriot','{"ram_type":"DDR4","frequency":3200,"capacity":8}', NOW(), NOW()),
  ('Corsair Vengeance LPX 8GB DDR4-2400',       'ram','Corsair','{"ram_type":"DDR4","frequency":2400,"capacity":8}', NOW(), NOW());

-- 4. Новые видеокарты
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('NVIDIA GeForce RTX 3080',  'gpu','NVIDIA','{"length_mm":285,"power_draw":320}', NOW(), NOW()),
  ('NVIDIA GeForce RTX 4090',  'gpu','NVIDIA','{"length_mm":304,"power_draw":450}', NOW(), NOW()),
  ('AMD Radeon RX 7900 XT',    'gpu','AMD',   '{"length_mm":267,"power_draw":300}', NOW(), NOW()),
  ('NVIDIA Quadro P2200',      'gpu','NVIDIA','{"length_mm":172,"power_draw":75}', NOW(), NOW()),
  ('AMD Radeon Pro W5500',     'gpu','AMD',   '{"length_mm":182,"power_draw":50}', NOW(), NOW()),
  ('NVIDIA GeForce GTX 1650',  'gpu','NVIDIA','{"length_mm":175,"power_draw":75}', NOW(), NOW());

-- 5. Новые блоки питания
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Corsair RM550X',               'psu','Corsair','{"power":550}',  NOW(), NOW()),
  ('EVGA SuperNOVA 1000 G+',       'psu','EVGA',   '{"power":1000}', NOW(), NOW()),
  ('Seasonic Prime TX-1000',       'psu','Seasonic','{"power":1000}', NOW(), NOW()),
  ('be quiet! Straight Power 11',  'psu','be quiet!','{"power":550}',NOW(), NOW()),
  ('Thermaltake Toughpower GF1',   'psu','Thermaltake','{"power":750}',NOW(), NOW());

-- 6. Новые корпуса
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Cooler Master NR200',         'case','Cooler Master','{"gpu_max_length":330,"cooler_max_height":155}', NOW(), NOW()),
  ('Fractal Design Node 304',     'case','Fractal','{"gpu_max_length":315,"cooler_max_height":56}', NOW(), NOW()),
  ('Corsair 4000D',               'case','Corsair','{"gpu_max_length":360,"cooler_max_height":170}', NOW(), NOW()),
  ('Lian Li O11 Dynamic',         'case','Lian Li','{"gpu_max_length":420,"cooler_max_height":165}', NOW(), NOW()),
  ('Phanteks Enthoo Pro',         'case','Phanteks','{"gpu_max_length":420,"cooler_max_height":190}', NOW(), NOW());

-- 7. Новые SSD
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Samsung 980 Pro 1TB',   'ssd','Samsung','{"form_factor":"M.2","interface":"PCIe 4.0"}', NOW(), NOW()),
  ('WD Black SN770 1TB',    'ssd','WD',     '{"form_factor":"M.2","interface":"PCIe 4.0"}', NOW(), NOW()),
  ('Crucial P5 500GB',      'ssd','Crucial','{"form_factor":"M.2","interface":"PCIe 3.0"}', NOW(), NOW()),
  ('Kingston A2000 500GB',  'ssd','Kingston','{"form_factor":"M.2","interface":"PCIe 3.0"}', NOW(), NOW());

-- 8. Новые HDD
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('WD Red Plus 4TB',       'hdd','Western Digital','{"form_factor":"3.5","interface":"SATA III","rpm":5400}', NOW(), NOW()),
  ('Seagate IronWolf 4TB',  'hdd','Seagate','{"form_factor":"3.5","interface":"SATA III","rpm":5900}', NOW(), NOW());
