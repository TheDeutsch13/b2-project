ALTER TABLE users
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS nickname,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS updated_at;
