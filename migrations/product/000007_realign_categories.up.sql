UPDATE products SET category_id = NULL
WHERE category_id IN (
  SELECT id FROM categories
  WHERE name NOT IN ('Мыши', 'Коврики', 'Глайды/грипсы', 'Рукава')
);

DELETE FROM categories
WHERE name NOT IN ('Мыши', 'Коврики', 'Глайды/грипсы', 'Рукава');

INSERT INTO categories (name) VALUES ('Мыши') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Коврики') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Глайды/грипсы') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Рукава') ON CONFLICT (name) DO NOTHING;
