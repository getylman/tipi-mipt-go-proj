-- Справочник продуктов (компонентов машины)
CREATE TABLE IF NOT EXISTS products (
    id             VARCHAR(100)   PRIMARY KEY,
    name           VARCHAR(255)   NOT NULL,
    price_per_unit NUMERIC(10,4)  NOT NULL,
    unit           VARCHAR(50)    NOT NULL,
    updated_at     TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- Пользователи (только uid, никаких лишних данных)
CREATE TABLE IF NOT EXISTS users (
    user_id    UUID           PRIMARY KEY,
    created_at TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- История потребления — одна строка на каждый product_id в запросе
CREATE TABLE IF NOT EXISTS consumption (
    id             SERIAL         PRIMARY KEY,
    user_id        UUID           NOT NULL REFERENCES users(user_id),
    product_id     VARCHAR(100)   NOT NULL REFERENCES products(id),
    quantity       NUMERIC(12,4)  NOT NULL,
    unit_price     NUMERIC(10,4)  NOT NULL,
    total_price    NUMERIC(12,4)  NOT NULL,
    calculated_at  TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consumption_user_id ON consumption(user_id);
CREATE INDEX IF NOT EXISTS idx_consumption_calculated_at ON consumption(calculated_at);

-- Начальные данные: компоненты машины
INSERT INTO products (id, name, price_per_unit, unit) VALUES
    ('vcpu',        'Virtual CPU',        2.50,  'core'),
    ('ram_gb',      'RAM',                0.80,  'gb'),
    ('disk_gb',     'SSD Disk',           0.15,  'gb'),
    ('network_mbps','Network bandwidth',  0.05,  'mbps')
ON CONFLICT (id) DO NOTHING;
