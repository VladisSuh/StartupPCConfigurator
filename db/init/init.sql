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
