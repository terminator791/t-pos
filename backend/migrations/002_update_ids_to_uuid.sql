-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Migration to change primary keys to UUID except logs and license_logs
-- Run this after backing up the database

-- Update shops table
ALTER TABLE shops DROP CONSTRAINT shops_pkey;
ALTER TABLE shops ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE shops ADD PRIMARY KEY (id);

-- Update categories table
ALTER TABLE categories DROP CONSTRAINT categories_pkey;
ALTER TABLE categories ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE categories ADD PRIMARY KEY (id);

-- Update products table
ALTER TABLE products DROP CONSTRAINT products_pkey;
ALTER TABLE products ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE products ADD PRIMARY KEY (id);

-- Update carts table
ALTER TABLE carts DROP CONSTRAINT carts_pkey;
ALTER TABLE carts ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE carts ADD PRIMARY KEY (id);

-- Update transaction_products table
ALTER TABLE transaction_products DROP CONSTRAINT transaction_products_pkey;
ALTER TABLE transaction_products ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE transaction_products ADD PRIMARY KEY (id);

-- Update payments table
ALTER TABLE payments DROP CONSTRAINT payments_pkey;
ALTER TABLE payments ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE payments ADD PRIMARY KEY (id);

-- Update receipts table
ALTER TABLE receipts DROP CONSTRAINT receipts_pkey;
ALTER TABLE receipts ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE receipts ADD PRIMARY KEY (id);

-- Update histories table
ALTER TABLE histories DROP CONSTRAINT histories_pkey;
ALTER TABLE histories ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE histories ADD PRIMARY KEY (id);

-- Update stock_histories table
ALTER TABLE stock_histories DROP CONSTRAINT stock_histories_pkey;
ALTER TABLE stock_histories ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE stock_histories ADD PRIMARY KEY (id);

-- Update expenses table
ALTER TABLE expenses DROP CONSTRAINT expenses_pkey;
ALTER TABLE expenses ALTER COLUMN id TYPE UUID USING uuid_generate_v4();
ALTER TABLE expenses ADD PRIMARY KEY (id);

-- Update foreign keys
ALTER TABLE categories DROP CONSTRAINT categories_shop_id_fkey;
ALTER TABLE categories ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE categories ADD CONSTRAINT categories_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE products DROP CONSTRAINT products_shop_id_fkey;
ALTER TABLE products ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE products ADD CONSTRAINT products_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE products DROP CONSTRAINT products_cat_id_fkey;
ALTER TABLE products ALTER COLUMN cat_id TYPE UUID;
ALTER TABLE products ADD CONSTRAINT products_cat_id_fkey FOREIGN KEY (cat_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE carts DROP CONSTRAINT carts_shop_id_fkey;
ALTER TABLE carts ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE carts ADD CONSTRAINT carts_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE carts DROP CONSTRAINT carts_product_id_fkey;
ALTER TABLE carts ALTER COLUMN product_id TYPE UUID;
ALTER TABLE carts ADD CONSTRAINT carts_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE transactions DROP CONSTRAINT transactions_shop_id_fkey;
ALTER TABLE transactions ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE transactions ADD CONSTRAINT transactions_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE transaction_products DROP CONSTRAINT transaction_products_product_id_fkey;
ALTER TABLE transaction_products ALTER COLUMN product_id TYPE UUID;
ALTER TABLE transaction_products ADD CONSTRAINT transaction_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE payments DROP CONSTRAINT payments_shop_id_fkey;
ALTER TABLE payments ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE payments ADD CONSTRAINT payments_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE receipts DROP CONSTRAINT receipts_shop_id_fkey;
ALTER TABLE receipts ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE receipts ADD CONSTRAINT receipts_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE receipts DROP CONSTRAINT receipts_payments_id_fkey;
ALTER TABLE receipts ALTER COLUMN payments_id TYPE UUID;
ALTER TABLE receipts ADD CONSTRAINT receipts_payments_id_fkey FOREIGN KEY (payments_id) REFERENCES payments(id) ON DELETE CASCADE;

ALTER TABLE histories DROP CONSTRAINT histories_shop_id_fkey;
ALTER TABLE histories ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE histories ADD CONSTRAINT histories_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

ALTER TABLE stock_histories DROP CONSTRAINT stock_histories_product_id_fkey;
ALTER TABLE stock_histories ALTER COLUMN product_id TYPE UUID;
ALTER TABLE stock_histories ADD CONSTRAINT stock_histories_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE expenses DROP CONSTRAINT expenses_shop_id_fkey;
ALTER TABLE expenses ALTER COLUMN shop_id TYPE UUID;
ALTER TABLE expenses ADD CONSTRAINT expenses_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE;

-- Update indexes that reference changed columns
DROP INDEX idx_categories_shop_id;
CREATE INDEX idx_categories_shop_id ON categories(shop_id);

DROP INDEX idx_products_cat_id;
CREATE INDEX idx_products_cat_id ON products(cat_id);

DROP INDEX idx_products_shop_cat;
CREATE INDEX idx_products_shop_cat ON products(shop_id, cat_id);

DROP INDEX idx_carts_product_id;
CREATE INDEX idx_carts_product_id ON carts(product_id);

DROP INDEX idx_carts_shop_id;
CREATE INDEX idx_carts_shop_id ON carts(shop_id);

DROP INDEX idx_transactions_shop_id;
CREATE INDEX idx_transactions_shop_id ON transactions(shop_id);

DROP INDEX idx_transaction_products_product_id;
CREATE INDEX idx_transaction_products_product_id ON transaction_products(product_id);

DROP INDEX idx_payments_shop_id;
CREATE INDEX idx_payments_shop_id ON payments(shop_id);

DROP INDEX idx_receipts_shop_id;
CREATE INDEX idx_receipts_shop_id ON receipts(shop_id);

DROP INDEX idx_receipts_payments_id;
CREATE INDEX idx_receipts_payments_id ON receipts(payments_id);

DROP INDEX idx_histories_shop_id;
CREATE INDEX idx_histories_shop_id ON histories(shop_id);

DROP INDEX idx_stock_histories_product_created;
CREATE INDEX idx_stock_histories_product_created ON stock_histories(product_id, created_at);

DROP INDEX idx_expenses_shop_id;
CREATE INDEX idx_expenses_shop_id ON expenses(shop_id);
