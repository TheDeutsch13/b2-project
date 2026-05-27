ALTER TABLE products
    ADD COLUMN IF NOT EXISTS rating_avg NUMERIC(3, 2) NOT NULL DEFAULT 0 CHECK (rating_avg >= 0 AND rating_avg <= 5),
    ADD COLUMN IF NOT EXISTS rating_count INTEGER NOT NULL DEFAULT 0 CHECK (rating_count >= 0);

-- Перенос товаров из старых категорий в «Аксессуары»
INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING;

UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Аксессуары' LIMIT 1)
WHERE category_id IN (
  SELECT id FROM categories WHERE name IN ('Глайды/грипсы', 'Рукава')
);

DELETE FROM categories
WHERE name NOT IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары');

INSERT INTO categories (name) VALUES ('Мыши') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Коврики') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING;
