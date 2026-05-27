INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING;

UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Аксессуары' LIMIT 1)
WHERE category_id IN (
  SELECT id FROM categories WHERE name IN ('Глайды/грипсы', 'Рукава')
);

UPDATE products SET category_id = NULL
WHERE category_id IN (
  SELECT id FROM categories
  WHERE name NOT IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары')
);

DELETE FROM categories
WHERE name NOT IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары');

INSERT INTO categories (name) VALUES ('Мыши') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Коврики') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING;
