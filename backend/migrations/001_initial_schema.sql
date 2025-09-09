-- Initial database schema for t-pos (Point of Sale System)
-- Based on canonical schema with licenses, shops, and comprehensive POS functionality

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Licenses table - Controls shop access and usage
CREATE TABLE licenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users table (owner, admin, cashier, client)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    license_id UUID REFERENCES licenses(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE,
    email_verified_at TIMESTAMP,
    username VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    pin VARCHAR(255),
    info_device VARCHAR(255),
    fcm_token VARCHAR(255),
    remember_token VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- License logs - Track license generation and assignment
CREATE TABLE license_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    license_id UUID REFERENCES licenses(id) ON DELETE CASCADE ON UPDATE CASCADE,
    generate_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Shops table - Merchant shops operating under licenses
CREATE TABLE shops (
    id BIGSERIAL PRIMARY KEY,
    license_id UUID NOT NULL REFERENCES licenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    photo VARCHAR(255),
    address TEXT,
    slogan VARCHAR(255),
    profit_calculate BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Categories table - Product categories per shop
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table - Sellable items
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    cat_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    photo VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    barcode VARCHAR(255),
    unit VARCHAR(50),
    ppn DECIMAL(5,2), -- tax percentage
    sale DECIMAL(10,2) NOT NULL,
    buy DECIMAL(10,2) NOT NULL,
    profit DECIMAL(10,2),
    stock INTEGER DEFAULT 0,
    is_schedule BOOLEAN DEFAULT FALSE,
    schedule JSON,
    qty INTEGER,
    is_have_stock BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Carts table - User shopping carts (pre-transaction basket)
CREATE TABLE carts (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Transactions table - Sales transaction header
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- customer
    discount DECIMAL(10,2) DEFAULT 0,
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    additional_cost DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled', 'failed')),
    total_price DECIMAL(10,2) NOT NULL,
    profit_transaction DECIMAL(10,2),
    cashier_name VARCHAR(255),
    change DECIMAL(10,2),
    amount BIGINT DEFAULT 0, -- amount paid
    initial_payment_status VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Transaction products table - Line items per transaction
CREATE TABLE transaction_products (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payments table - Payments linked to transactions
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    total DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Receipts table - Receipt records pointing to payments
CREATE TABLE receipts (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    payments_id BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Histories table - Lightweight link between shop and transaction for history views
CREATE TABLE histories (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Stock histories table - Append-only changes to product stock
CREATE TABLE stock_histories (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    stock INTEGER NOT NULL,
    last_stock INTEGER NOT NULL,
    stocked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Expenses table - Shop expenses/outflows
CREATE TABLE expenses (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    nominal DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    date DATE NOT NULL,
    label VARCHAR(255),
    desc TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Logs table - Audit trail of user actions and model changes
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(255) NOT NULL,
    model VARCHAR(255),
    model_id UUID,
    old_values JSON,
    new_values JSON,
    ip_address VARCHAR(255),
    user_agent TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_license_logs_license_id ON license_logs(license_id);
CREATE INDEX idx_license_logs_user_id ON license_logs(user_id);
CREATE INDEX idx_shops_license_id ON shops(license_id);
CREATE INDEX idx_shops_user_id ON shops(user_id);
CREATE INDEX idx_categories_shop_id ON categories(shop_id);
CREATE INDEX idx_products_cat_id ON products(cat_id);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_barcode ON products(barcode);
CREATE INDEX idx_products_created_at ON products(created_at);
CREATE INDEX idx_products_shop_cat ON products(shop_id, cat_id);
CREATE INDEX idx_carts_product_id ON carts(product_id);
CREATE INDEX idx_carts_shop_id ON carts(shop_id);
CREATE INDEX idx_carts_user_created ON carts(user_id, created_at);
CREATE INDEX idx_transactions_shop_id ON transactions(shop_id);
CREATE INDEX idx_transactions_cashier_id ON transactions(cashier_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transaction_products_transaction_id ON transaction_products(transaction_id);
CREATE INDEX idx_transaction_products_product_id ON transaction_products(product_id);
CREATE INDEX idx_payments_shop_id ON payments(shop_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX idx_receipts_shop_id ON receipts(shop_id);
CREATE INDEX idx_receipts_payments_id ON receipts(payments_id);
CREATE INDEX idx_histories_shop_id ON histories(shop_id);
CREATE INDEX idx_histories_transaction_id ON histories(transaction_id);
CREATE INDEX idx_stock_histories_product_created ON stock_histories(product_id, created_at);
CREATE INDEX idx_stock_histories_stocked_at ON stock_histories(stocked_at);
CREATE INDEX idx_expenses_shop_id ON expenses(shop_id);
CREATE INDEX idx_logs_created_at ON logs(created_at);
CREATE INDEX idx_logs_user_created ON logs(user_id, created_at);
CREATE INDEX idx_logs_model ON logs(model, model_id);