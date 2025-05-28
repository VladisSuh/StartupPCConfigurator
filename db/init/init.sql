-- ===================================================================
-- PostgreSQL Setup Script
-- ===================================================================

-- =====================
-- 1. EXTENSIONS
-- =====================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- =====================
-- 2. TABLE DEFINITIONS
-- =====================

-- 2.1 Users
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

-- 2.2 Components
CREATE TABLE IF NOT EXISTS components (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,        -- e.g. 'cpu', 'gpu'
    brand VARCHAR(100),                    -- e.g. 'Intel', 'AMD'
    specs JSONB,                           -- e.g. {"cores":6, ...}
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.3 Configurations
CREATE TABLE IF NOT EXISTS configurations (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.4 Configuration ⇆ Component Link
CREATE TABLE IF NOT EXISTS configuration_components (
    id SERIAL PRIMARY KEY,
    config_id INT NOT NULL
        REFERENCES configurations(id) ON DELETE CASCADE,
    component_id INT NOT NULL
        REFERENCES components(id) ON DELETE RESTRICT,
    category VARCHAR(100),                  -- optional duplicate or lookup
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.5 Shops
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

-- 2.6 Offers
CREATE TABLE IF NOT EXISTS offers (
    id SERIAL PRIMARY KEY,
    component_id INT NOT NULL
        REFERENCES components(id) ON DELETE CASCADE,
    shop_id INT NOT NULL
        REFERENCES shops(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    availability TEXT,
    url TEXT,
    fetched_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (component_id, shop_id)
);

-- 2.7 Price History
CREATE TABLE IF NOT EXISTS price_history (
    id BIGSERIAL PRIMARY KEY,
    component_id INT NOT NULL
        REFERENCES components(id) ON DELETE CASCADE,
    shop_id INT NOT NULL
        REFERENCES shops(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    captured_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.8 Update Jobs
CREATE TABLE IF NOT EXISTS update_jobs (
    id BIGSERIAL PRIMARY KEY,
    shop_id INT REFERENCES shops(id),
    status TEXT NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    message TEXT
);

-- 2.9 Use Cases
CREATE TABLE IF NOT EXISTS usecases (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- 2.10 Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    component_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, component_id)
);

-- 2.11 Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    component_id TEXT NOT NULL,
    shop_id INT NOT NULL,
    old_price NUMERIC NOT NULL,
    new_price NUMERIC NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);



-- =====================
-- 4. INDEXES

-- Components
CREATE INDEX IF NOT EXISTS idx_components_category ON components(category);
CREATE INDEX IF NOT EXISTS idx_components_specs    ON components USING GIN (specs);
CREATE INDEX IF NOT EXISTS idx_components_name_brand ON components(name, brand);

-- Configurations
CREATE INDEX IF NOT EXISTS idx_configurations_user                   ON configurations(user_id);
CREATE INDEX IF NOT EXISTS idx_configuration_components_config       ON configuration_components(config_id);
CREATE INDEX IF NOT EXISTS idx_configuration_components_component    ON configuration_components(component_id);

-- Offers & Price History
CREATE INDEX IF NOT EXISTS idx_offers_component      ON offers(component_id);
CREATE INDEX IF NOT EXISTS idx_offers_price          ON offers(price);
CREATE INDEX IF NOT EXISTS idx_price_history_comp_time ON price_history(component_id, captured_at DESC);

-- Subscriptions & Notifications
CREATE INDEX IF NOT EXISTS idx_subscriptions_by_user      ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_by_component ON subscriptions(component_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read    ON notifications(user_id, is_read);


-- =====================
-- 4. SEED DATA
-- =====================

-- 4.1 Use Cases
INSERT INTO usecases(name, description) VALUES
  ('office',    'Офисный ПК для документов и почты'),
  ('htpc',      'Домашний медиаплеер в мини-корпусе'),
  ('gaming',    'Игровой ПК среднего–высокого уровня'),
  ('streamer',  'Станция для стриминга и игр'),
  ('design',    'Графический дизайн и фото-обработка'),
  ('video',     'Видеомонтаж и рендеринг'),
  ('cad',       '3D-моделирование и CAD'),
  ('dev',       'ПК для разработки и виртуализации'),
  ('enthusiast','High-End система с разгоном и RGB'),
  ('nas',       'Домашний сервер / NAS');

-- 4.2 Components by Category
-- 4.2.1 CPUs
INSERT INTO components(name, category, brand, specs) VALUES
  ('AMD Ryzen 5 5600X', 'cpu', 'AMD', '{"cores":6,"threads":12,"socket":"AM4","tdp":65,"power_draw":65,"cooler_height":158}'),
  ('Intel Core i5-12400','cpu','Intel','{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"power_draw":65,"cooler_height":145}'),
  ('AMD Ryzen 7 7700X', 'cpu', 'AMD', '{"cores":8,"threads":16,"socket":"AM5","tdp":105,"power_draw":105,"cooler_height":160}'),
  ('Intel Core i7-14700K','cpu','Intel','{"cores":20,"threads":28,"socket":"LGA1700","tdp":125,"power_draw":125,"cooler_height":160}'),
  ('AMD Ryzen 9 7950X','cpu','AMD','{"cores":16,"threads":32,"socket":"AM5","tdp":170,"power_draw":170,"cooler_height":165}');

-- 4.2.2 GPUs
INSERT INTO components(name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 3060','gpu','NVIDIA','{"length_mm":242,"power_draw":170,"memory_gb":12,"interface":"PCIe 4.0","height_mm":40}'),
  ('AMD Radeon RX 6700 XT','gpu','AMD','{"length_mm":267,"power_draw":230,"memory_gb":12,"interface":"PCIe 4.0","height_mm":43}'),
  ('NVIDIA GeForce RTX 4070 SUPER','gpu','NVIDIA','{"length_mm":267,"power_draw":220,"memory_gb":12,"interface":"PCIe 4.0","height_mm":45}'),
  ('AMD Radeon RX 7800 XT','gpu','AMD','{"length_mm":276,"power_draw":263,"memory_gb":16,"interface":"PCIe 4.0","height_mm":48}'),
  ('NVIDIA GeForce RTX 4090','gpu','NVIDIA','{"length_mm":304,"power_draw":450,"memory_gb":24,"interface":"PCIe 4.0","height_mm":60}');

-- 4.2.3 Motherboards
INSERT INTO components(name, category, brand, specs) VALUES
  ('MSI B550 Tomahawk','motherboard','MSI','{"socket":"AM4","ram_type":"DDR4","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":6}'),
  ('ASUS PRIME B660-PLUS','motherboard','ASUS','{"socket":"LGA1700","ram_type":"DDR4","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":6}'),
  ('ASRock B650M Pro RS WiFi','motherboard','ASRock','{"socket":"AM5","ram_type":"DDR5","form_factor":"Micro-ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 4.0","m2_slots":2,"sata_ports":4}'),
  ('Gigabyte Z790 AORUS Elite AX','motherboard','Gigabyte','{"socket":"LGA1700","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":3,"sata_ports":6}'),
  ('MSI PRO B760M-A DDR5','motherboard','MSI','{"socket":"LGA1700","ram_type":"DDR5","form_factor":"Micro-ATX","max_memory_gb":128,"memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":2,"sata_ports":4}');

-- 4.2.4 RAM
INSERT INTO components(name, category, brand, specs) VALUES
  ('Corsair Vengeance LPX 16GB DDR4-3200','ram','Corsair','{"ram_type":"DDR4","frequency":3200,"capacity":16,"modules":2,"voltage":1.35}'),
  ('Kingston Fury 32GB DDR4-3600','ram','Kingston','{"ram_type":"DDR4","frequency":3600,"capacity":32,"modules":2,"voltage":1.35}'),
  ('G.Skill Trident Z5 RGB 32GB DDR5-6000','ram','G.Skill','{"ram_type":"DDR5","frequency":6000,"capacity":32,"modules":2,"voltage":1.25}'),
  ('Corsair Vengeance 32GB DDR5-6000','ram','Corsair','{"ram_type":"DDR5","frequency":6000,"capacity":32,"modules":2,"voltage":1.25}'),
  ('Kingston Fury Renegade 64GB DDR5-6400','ram','Kingston','{"ram_type":"DDR5","frequency":6400,"capacity":64,"modules":2,"voltage":1.35}');

-- 4.2.5 Storage (SSD & HDD)
INSERT INTO components(name, category, brand, specs) VALUES
  ('Samsung 980 Pro 1TB','ssd','Samsung','{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":7000}'),
  ('WD Black SN770 1TB','ssd','WD','{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":5000}'),
  ('Crucial P5 500GB','ssd','Crucial','{"form_factor":"M.2","interface":"PCIe 3.0","capacity_gb":500,"m2_key":"M","max_throughput":3400}'),
  ('Kingston A2000 500GB','ssd','Kingston','{"form_factor":"M.2","interface":"PCIe 3.0","capacity_gb":500,"m2_key":"M","max_throughput":2200}'),
  ('Sabrent Rocket 4 Plus 2TB','ssd','Sabrent','{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":2000,"m2_key":"M","max_throughput":7100}'),
  ('ADATA XPG Gammix S70 Blade 1TB','ssd','ADATA','{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":1000,"m2_key":"M","max_throughput":7400}'),
  ('WD Red Plus 4TB','hdd','Western Digital','{"form_factor":"3.5","interface":"SATA III","rpm":5400,"capacity_gb":4000}'),
  ('Seagate IronWolf 4TB','hdd','Seagate','{"form_factor":"3.5","interface":"SATA III","rpm":5900,"capacity_gb":4000}'),
  ('Toshiba N300 6TB','hdd','Toshiba','{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":6000}'),
  ('Seagate Barracuda 2TB','hdd','Seagate','{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":2000}'),
  ('WD Blue 1TB','hdd','Western Digital','{"form_factor":"3.5","interface":"SATA III","rpm":7200,"capacity_gb":1000}');

-- 4.2.6 PSUs
INSERT INTO components(name, category, brand, specs) VALUES
  ('Seasonic Focus GX-650','psu','Seasonic','{"power":650,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Corsair RM750','psu','Corsair','{"power":750,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('be quiet! Pure Power 12 M 850W','psu','be quiet!','{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Corsair RM850x SHIFT','psu','Corsair','{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('Cooler Master V850 SFX Gold','psu','Cooler Master','{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"SFX"}'),
  ('Seasonic PRIME TX-1000','psu','Seasonic','{"power":1000,"efficiency":"80 Plus Titanium","modular":true,"form_factor":"ATX"}'),
  ('Thermaltake Toughpower GF3 1200W','psu','Thermaltake','{"power":1200,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}'),
  ('MSI MPG A850G','psu','MSI','{"power":850,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}');

-- 4.2.7 Cases
INSERT INTO components(name, category, brand, specs) VALUES
  ('NZXT H510','case','NZXT','{"form_factor":"ATX","gpu_max_length":325,"cooler_max_height":165,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],
                                   "max_psu_length":200,"psu_form_factor":"ATX","drive_bays_2_5":2,"drive_bays_3_5":2}'),
  ('Fractal Design North','case','Fractal Design','{"form_factor":"ATX","gpu_max_length":355,"cooler_max_height":170,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],
                                           "max_psu_length":250,"psu_form_factor":"ATX","drive_bays_2_5":3,"drive_bays_3_5":2}'),
  ('Lian Li A4-H2O','case','Lian Li','{"form_factor":"Mini-ITX","gpu_max_length":322,"cooler_max_height":55,
                                     "max_motherboard_form_factors":["Mini-ITX"],"max_psu_length":130,"psu_form_factor":"SFX","drive_bays_2_5":2,"drive_bays_3_5":0}'),
  ('Cooler Master NR200P','case','Cooler Master','{"form_factor":"Mini-ITX","gpu_max_length":330,"cooler_max_height":155,"max_motherboard_form_factors":["Mini-ITX"],
                                            "max_psu_length":160,"psu_form_factor":"SFX","drive_bays_2_5":3,"drive_bays_3_5":1}'),
  ('Phanteks Eclipse P400A','case','Phanteks','{"form_factor":"ATX","gpu_max_length":420,"cooler_max_height":160,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],
                                       "max_psu_length":270,"psu_form_factor":"ATX","drive_bays_2_5":2,"drive_bays_3_5":2}'),
  ('Phanteks Eclipse G500A','case','Phanteks','{"form_factor":"ATX","gpu_max_length":435,"cooler_max_height":185,"max_motherboard_form_factors":["ATX","Micro-ATX","Mini-ITX"],
                                       "max_psu_length":270,"psu_form_factor":"ATX","drive_bays_2_5":4,"drive_bays_3_5":3}');

-- 4.2.8 Coolers
INSERT INTO components(name, category, brand, specs) VALUES
  ('BeQuiet Shadow Rock 3','cooler','BeQuiet','{"socket":"AM4","height_mm":160}'),
  ('Noctua NH-D15','cooler','Noctua','{"socket":"AM4","height_mm":165}'),
  ('Thermalright Peerless Assassin 120 SE','cooler','Thermalright','{"socket":"AM5","height_mm":155}'),
  ('Noctua NH-U12A chromax.black','cooler','Noctua','{"socket":"LGA1700","height_mm":158}'),
  ('DeepCool AK620 Digital','cooler','DeepCool','{"socket":"AM5","height_mm":162}'),
  ('Scythe Fuma 2 Rev.B','cooler','Scythe','{"socket":"AM4","height_mm":155}'),
  ('ARCTIC Freezer 34 eSports DUO','cooler','ARCTIC','{"socket":"LGA1700","height_mm":157}');


-- 4.2.x  GPU c 10 GB VRAM (для минимального ранга design)
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Radeon RX 6700 10GB', 'gpu', 'AMD',
   '{"length_mm":255,"height_mm":42,"memory_gb":10,"power_draw":175,"interface":"PCIe 4.0"}');

-- 4.2.x  Дополнительная ATX-плата AM5 + DDR5 (опционально для разнообразия)
INSERT INTO components (name, category, brand, specs) VALUES
  ('ASUS TUF Gaming B650-PLUS WIFI', 'motherboard', 'ASUS',
   '{"socket":"AM5","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,"memory_slots":4,
     "pcie_version":"PCIe 4.0","m2_slots":3,"sata_ports":4}');

 -- 10-гигабайтная карта с высоким TBP для ранга 0
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 3080 10GB','gpu','NVIDIA',
   '{"length_mm":285,"height_mm":52,"memory_gb":10,
     "power_draw":320,"interface":"PCIe 4.0"}');
INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Radeon RX 6700 10GB','gpu','AMD',
   '{"length_mm":255,"height_mm":42,"memory_gb":10,
     "power_draw":175,"interface":"PCIe 4.0"}');
   
 /* 2. GPUs – промежуточная и топовая */
INSERT INTO components (name, category, brand, specs) VALUES
  ('NVIDIA GeForce RTX 4080 SUPER 16GB', 'gpu', 'NVIDIA',
   '{"length_mm":304,"height_mm":50,"memory_gb":16,"power_draw":320,"interface":"PCIe 4.0"}'),
  ('AMD Radeon RX 7900 XTX 24GB',        'gpu', 'AMD',
   '{"length_mm":305,"height_mm":55,"memory_gb":24,"power_draw":355,"interface":"PCIe 4.0"}');

/* 3. PCIe 5.0 SSD для максимальной сборки */
INSERT INTO components (name, category, brand, specs) VALUES
  ('Crucial T700 2TB', 'ssd', 'Crucial',
   '{"form_factor":"M.2","interface":"PCIe 5.0","capacity_gb":2000,
     "m2_key":"M","max_throughput":12400}');

/* 4. 1-кВт блок питания с высоким КПД */
INSERT INTO components (name, category, brand, specs) VALUES
  ('Corsair HX1000i 1000W', 'psu', 'Corsair',
   '{"power":1000,"efficiency":"80 Plus Platinum","modular":true,"form_factor":"ATX"}');
-- 4.3 Shops
INSERT INTO shops(code, name, base_url, api_endpoint) VALUES
  ('DNS','DNS Shop','https://www.dns-shop.ru',NULL),
  ('Citilink','Citilink Shop','https://www.citilink.ru/',NULL),
  ('Regard','Regard Shop','https://www.regard.ru/',NULL),
  ('Nix','Nix Shop','https://www.nix.ru/',NULL);


-- CPU
INSERT INTO components(name, category, brand, specs) VALUES
 ('AMD Ryzen 7 7800X3D','cpu','AMD',
  '{"cores":8,"threads":16,"socket":"AM5","tdp":120,"power_draw":120,"cooler_height":160}'),
 ('AMD Ryzen 9 7950X3D','cpu','AMD',
  '{"cores":16,"threads":32,"socket":"AM5","tdp":120,"power_draw":120,"cooler_height":165}'),
 ('Intel Core i9-14900K','cpu','Intel',
  '{"cores":24,"threads":32,"socket":"LGA1700","tdp":125,"power_draw":125,"cooler_height":170}');

-- Motherboards (ATX, DDR5)
INSERT INTO components(name, category, brand, specs) VALUES
 ('ASUS ROG Strix X670E-E Gaming WiFi','motherboard','ASUS',
  '{"socket":"AM5","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,
    "memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":4,"sata_ports":6}'),
 ('Gigabyte Z790 AORUS Master','motherboard','Gigabyte',
  '{"socket":"LGA1700","ram_type":"DDR5","form_factor":"ATX","max_memory_gb":128,
    "memory_slots":4,"pcie_version":"PCIe 5.0","m2_slots":5,"sata_ports":6}');

-- GPUs 16–20 GB
INSERT INTO components(name, category, brand, specs) VALUES
 ('NVIDIA GeForce RTX 4070 Ti SUPER 16GB','gpu','NVIDIA',
  '{"memory_gb":16,"interface":"PCIe 4.0","length_mm":305,"height_mm":50,"power_draw":285}'),
 ('AMD Radeon RX 7900 XT 20GB','gpu','AMD',
  '{"memory_gb":20,"interface":"PCIe 4.0","length_mm":276,"height_mm":52,"power_draw":315}');

-- High-capacity / mid-speed SSD
INSERT INTO components(name, category, brand, specs) VALUES
 ('Samsung 990 Pro 4TB','ssd','Samsung',
  '{"form_factor":"M.2","interface":"PCIe 4.0","capacity_gb":4000,
    "m2_key":"M","max_throughput":7450}');

-- 128 GB RAM
INSERT INTO components(name, category, brand, specs) VALUES
 ('G.Skill Trident Z5 128GB DDR5-6000','ram','G.Skill',
  '{"ram_type":"DDR5","frequency":6000,"capacity":128,
    "modules":2,"voltage":1.40}');

-- 1500 W PSU
INSERT INTO components(name, category, brand, specs) VALUES
 ('Corsair HX1500i','psu','Corsair',
  '{"power":1500,"efficiency":"80 Plus Platinum","modular":true,"form_factor":"ATX"}');

-- Premium ATX case
INSERT INTO components(name, category, brand, specs) VALUES
 ('Lian Li O11D EVO','case','Lian Li',
  '{"form_factor":"ATX","gpu_max_length":426,"cooler_max_height":167,
    "max_motherboard_form_factors":["E-ATX","ATX","Micro-ATX","Mini-ITX"],
    "max_psu_length":220,"psu_form_factor":"ATX","drive_bays_2_5":6,"drive_bays_3_5":4}');

INSERT INTO components (name, category, brand, specs) VALUES
  -- AM5
  ('AMD Ryzen 5 7600',   'cpu', 'AMD',
   '{"cores":6,"threads":12,"socket":"AM5","tdp":65,"power_draw":65,"cooler_height":158}'),
  ('AMD Ryzen 5 7500F',  'cpu', 'AMD',
   '{"cores":6,"threads":12,"socket":"AM5","tdp":65,"power_draw":65,"cooler_height":158}'),

  -- LGA1700
  ('Intel Core i5-13400',  'cpu', 'Intel',
   '{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"power_draw":65,"cooler_height":145}'),
  ('Intel Core i5-13400F', 'cpu', 'Intel',
   '{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"power_draw":65,"cooler_height":145}'),
  ('Intel Core i5-14400',  'cpu', 'Intel',
   '{"cores":6,"threads":12,"socket":"LGA1700","tdp":65,"power_draw":65,"cooler_height":145}');

INSERT INTO components (name, category, brand, specs) VALUES
  ('AMD Radeon RX 7600 8GB', 'gpu', 'AMD',
   '{"length_mm":212,"height_mm":42,"power_draw":165,"memory_gb":8,"interface":"PCIe 4.0"}'),
  ('NVIDIA GeForce RTX 4060 8GB','gpu','NVIDIA',
   '{"length_mm":242,"height_mm":40,"power_draw":115,"memory_gb":8,"interface":"PCIe 4.0"}');


INSERT INTO components (name, category, brand, specs) VALUES
  ('Corsair RM750e','psu','Corsair',
   '{"power":750,"efficiency":"80 Plus Gold","modular":true,"form_factor":"ATX"}');
-- === CPU =========================================================
INSERT INTO components ( name, category, brand, specs, created_at, updated_at) VALUES
( 'AMD Ryzen 7 5700X', 'cpu', 'AMD',
 '{"tdp":65,"cores":8,"socket":"AM4","threads":16,"power_draw":65,"cooler_height":158}',
 NOW(), NOW()),
( 'AMD Ryzen 9 5900',  'cpu', 'AMD',
 '{"tdp":65,"cores":12,"socket":"AM4","threads":24,"power_draw":65,"cooler_height":158}',
 NOW(), NOW()),
( 'Intel Core i7-13700', 'cpu', 'Intel',
 '{"tdp":65,"cores":16,"socket":"LGA1700","threads":24,"power_draw":65,"cooler_height":160}',
 NOW(), NOW()),
( 'Intel Core i9-13900', 'cpu', 'Intel',
 '{"tdp":65,"cores":24,"socket":"LGA1700","threads":32,"power_draw":65,"cooler_height":170}',
 NOW(), NOW());

-- === GPU =========================================================
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
('AMD Radeon RX 6400 4GB', 'gpu', 'AMD',
 '{"height_mm":40,"interface":"PCIe 4.0","length_mm":180,"memory_gb":4,"power_draw":53}',
 NOW(), NOW()),
('NVIDIA GeForce GTX 1660 6GB', 'gpu', 'NVIDIA',
 '{"height_mm":38,"interface":"PCIe 3.0","length_mm":229,"memory_gb":6,"power_draw":120}',
 NOW(), NOW());

-- === SSD / SATA & NVMe PCIe 3.0 ==================================
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
 ('Kingston A400 480GB', 'ssd', 'Kingston',
 '{"interface":"SATA III","capacity_gb":480,"form_factor":"2.5","max_throughput":550}',
 NOW(), NOW()),
( 'Crucial P3 1TB',      'ssd', 'Crucial',
 '{"m2_key":"M","interface":"PCIe 3.0","capacity_gb":1000,"form_factor":"M.2","max_throughput":3500}',
 NOW(), NOW()),
( 'WD Blue SN570 1TB',   'ssd', 'WD',
 '{"m2_key":"M","interface":"PCIe 3.0","capacity_gb":1000,"form_factor":"M.2","max_throughput":3500}',
 NOW(), NOW());

-- === RAM =========================================================
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
( 'Corsair Vengeance RGB Pro 64GB DDR4-3600', 'ram', 'Corsair',
 '{"modules":2,"voltage":1.35,"capacity":64,"ram_type":"DDR4","frequency":3600}',
 NOW(), NOW());

-- === PSU =========================================================
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
( 'EVGA 550 BR',  'psu', 'EVGA',
 '{"power":550,"modular":false,"efficiency":"80 Plus Bronze","form_factor":"ATX"}',
 NOW(), NOW()),
( 'Seasonic PRIME PX-800', 'psu', 'Seasonic',
 '{"power":800,"modular":true,"efficiency":"80 Plus Platinum","form_factor":"ATX"}',
 NOW(), NOW());

 -- БЮДЖЕТНЫЕ ПРОЦЕССОРЫ
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
('AMD Ryzen 3 3100',
 'cpu',
 'AMD',
 '{"tdp":65,"cores":4,"socket":"AM4","threads":8,"power_draw":65,"cooler_height":158}',
 now(), now()),
('Intel Core i3-12100F',
 'cpu',
 'Intel',
 '{"tdp":58,"cores":4,"socket":"LGA1700","threads":8,"power_draw":58,"cooler_height":145}',
 now(), now());

-- ДЕШЁВЫЕ ПЛАТЫ
INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
('ASUS PRIME A320M-K',
 'motherboard',
 'ASUS',
 '{"socket":"AM4","m2_slots":1,"ram_type":"DDR4","sata_ports":4,
   "form_factor":"Micro-ATX","memory_slots":2,"pcie_version":"PCIe 3.0",
   "max_memory_gb":64}',
 now(), now()),
('Gigabyte H610M K',
 'motherboard',
 'Gigabyte',
 '{"socket":"LGA1700","m2_slots":1,"ram_type":"DDR4","sata_ports":4,
   "form_factor":"Micro-ATX","memory_slots":2,"pcie_version":"PCIe 4.0",
   "max_memory_gb":64}',
 now(), now());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
('Kingston NV2 500GB',
 'ssd',
 'Kingston',
 '{"m2_key":"M","interface":"PCIe 3.0",
   "capacity_gb":500,"form_factor":"M.2","max_throughput":2100}',
 now(), now()),
('WD Blue SN550 500GB',
 'ssd',
 'WD',
 '{"m2_key":"M","interface":"PCIe 3.0",
   "capacity_gb":500,"form_factor":"M.2","max_throughput":2400}',
 now(), now());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
('DeepCool Matrexx 30',
 'case',
 'DeepCool',
 '{"form_factor":"Micro-ATX","drive_bays_2_5":2,"drive_bays_3_5":1,
   "gpu_max_length":250,"max_psu_length":150,"psu_form_factor":"ATX",
   "cooler_max_height":151,
   "max_motherboard_form_factors":["Micro-ATX","Mini-ITX"]}',
 now(), now());
INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  ('AMD Ryzen 3 5300G', 'cpu', 'AMD',
   '{"socket":"AM5","tdp":95,"cores":4,"threads":8,"power_draw":95,"cooler_height":155}',
   NOW(), NOW());


INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  (
    'ASRock B550 Phantom Gaming-ITX/ax',
    'motherboard',
    'ASRock',
    '{"socket":"AM4","m2_slots":2,"ram_type":"DDR4","sata_ports":4,"form_factor":"Mini-ITX","memory_slots":2,"pcie_version":"PCIe 4.0","max_memory_gb":64}',
    NOW(), NOW()
  );

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES
  (
    'Corsair CX450M',
    'psu',
    'Corsair',
    '{"power":450,"modular":true,"efficiency":"80 Plus Bronze","form_factor":"ATX"}',
    NOW(), NOW()
  );

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
  'SilverStone SX450-G',
  'psu',
  'SilverStone',
  '{"power":450, "modular":true, "efficiency":"80 Plus Gold", "form_factor":"SFX"}',
  NOW(), NOW()
);

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
  'Corsair SF500',
  'psu',
  'Corsair',
  '{"power":500, "modular":true, "efficiency":"80 Plus Gold", "form_factor":"SFX"}',
  NOW(), NOW()
);

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Intel Core i3-12100T',      'cpu', 'Intel',
   '{"tdp":35,"cores":4,"threads":8,"socket":"LGA1700","power_draw":35,"cooler_height":47}',
   NOW(), NOW()),
  ('Intel Pentium Gold G7400T', 'cpu', 'Intel',
   '{"tdp":35,"cores":2,"threads":4,"socket":"LGA1700","power_draw":35,"cooler_height":45}',
   NOW(), NOW()),
  ('AMD Ryzen 5 5600GE',        'cpu', 'AMD',
   '{"tdp":35,"cores":6,"threads":12,"socket":"AM4","power_draw":35,"cooler_height":54}',
   NOW(), NOW()),
  ('AMD Athlon 3000G',          'cpu', 'AMD',
   '{"tdp":35,"cores":2,"threads":4,"socket":"AM4","power_draw":35,"cooler_height":46}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Seasonic SSP-300ET',        'psu', 'Seasonic',
   '{"power":300,"modular":false,"efficiency":"80 Plus Bronze","form_factor":"ATX"}',
   NOW(), NOW()),
  ('SilverStone ST30SF V2.0',   'psu', 'SilverStone',
   '{"power":300,"modular":false,"efficiency":"80 Plus Bronze","form_factor":"SFX"}',
   NOW(), NOW()),
  ('FSP Dagger 450W',           'psu', 'FSP',
   '{"power":450,"modular":true,"efficiency":"80 Plus Gold","form_factor":"SFX"}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Corsair SF500',             'psu', 'Corsair',
   '{"power":500,"modular":true,"efficiency":"80 Plus Gold","form_factor":"SFX"}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('SilverStone SG13',          'case', 'SilverStone',
   '{"form_factor":"Mini-ITX","drive_bays_2_5":3,"drive_bays_3_5":1,
     "gpu_max_length":270,"max_psu_length":150,"psu_form_factor":"SFX",
     "cooler_max_height":61,
     "max_motherboard_form_factors":["Mini-ITX"]}',
   NOW(), NOW()),
  ('Fractal Design Node 202',   'case', 'Fractal Design',
   '{"form_factor":"Mini-ITX","drive_bays_2_5":2,"drive_bays_3_5":0,
     "gpu_max_length":310,"max_psu_length":130,"psu_form_factor":"SFX",
     "cooler_max_height":56,
     "max_motherboard_form_factors":["Mini-ITX"]}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Kingston ValueRAM 8GB DDR4-2666', 'ram', 'Kingston',
   '{"modules":1,"capacity":8,"ram_type":"DDR4","frequency":2666,"voltage":1.20}',
   NOW(), NOW()),
  ('ADATA SU650 240GB',               'ssd', 'ADATA',
   '{"interface":"SATA III","capacity_gb":240,"form_factor":"2.5","max_throughput":520}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('Intel Core i7-13700T', 'cpu', 'Intel',
   '{"tdp":35,"cores":16,"threads":24,"socket":"LGA1700","power_draw":35,"cooler_height":47}',
   NOW(), NOW()),
  ('Intel Core i5-13600T', 'cpu', 'Intel',
   '{"tdp":35,"cores":14,"threads":20,"socket":"LGA1700","power_draw":35,"cooler_height":47}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('ASUS ROG Strix B660-I Gaming WiFi', 'motherboard', 'ASUS',
   '{"socket":"LGA1700","m2_slots":2,"ram_type":"DDR4","sata_ports":4,
     "form_factor":"Mini-ITX","memory_slots":2,"pcie_version":"PCIe 4.0",
     "max_memory_gb":64}', NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('WD Black SN850X 1TB', 'ssd', 'WD',
   '{"m2_key":"M","interface":"PCIe 4.0","capacity_gb":1000,
     "form_factor":"M.2","max_throughput":7300}', NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at) VALUES
  ('SilverStone SX500-G', 'psu', 'SilverStone',
   '{"power":500,"modular":true,"efficiency":"80 Plus Gold","form_factor":"SFX"}',
   NOW(), NOW());

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
  'Intel Pentium Gold G7400T',
  'cpu',
  'Intel',
  '{"tdp":35,"cores":2,"threads":4,"socket":"LGA1700",
    "power_draw":35,"cooler_height":45}',
  NOW(), NOW()
);

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
  'Kingston ValueRAM 4GB DDR4-2400',
  'ram',
  'Kingston',
  '{"modules":1,"voltage":1.20,"capacity":4,
    "ram_type":"DDR4","frequency":2400}',
  NOW(), NOW()
);

INSERT INTO components (name, category, brand, specs, created_at, updated_at)
VALUES (
  'FSP HV PRO 200W',
  'psu',
  'FSP',
  '{"power":200,"modular":false,
    "efficiency":"80 Plus","form_factor":"ATX"}',
  NOW(), NOW()
);

-- ===================================================================
-- End of Script
-- ===================================================================
