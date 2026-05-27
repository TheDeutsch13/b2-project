ALTER TABLE products
    DROP COLUMN IF EXISTS rating_avg,
    DROP COLUMN IF EXISTS rating_count;
