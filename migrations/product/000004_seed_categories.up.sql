INSERT INTO categories (name) VALUES ('Мыши') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Коврики') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING;
INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING;
