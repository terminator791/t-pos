-- Add sync optimization indexes for improved query performance
-- These indexes are critical for sync operations that filter by updated_at and license_id

-- Indexes for updated_at columns (critical for delta sync)
CREATE INDEX IF NOT EXISTS idx_carts_updated_at ON carts(updated_at);
CREATE INDEX IF NOT EXISTS idx_categories_updated_at ON categories(updated_at);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at);
CREATE INDEX IF NOT EXISTS idx_transactions_updated_at ON transactions(updated_at);
CREATE INDEX IF NOT EXISTS idx_payments_updated_at ON payments(updated_at);
CREATE INDEX IF NOT EXISTS idx_expenses_updated_at ON expenses(updated_at);
CREATE INDEX IF NOT EXISTS idx_receipts_updated_at ON receipts(updated_at);
CREATE INDEX IF NOT EXISTS idx_histories_updated_at ON histories(updated_at);
CREATE INDEX IF NOT EXISTS idx_stock_histories_updated_at ON stock_histories(updated_at);
CREATE INDEX IF NOT EXISTS idx_transaction_products_updated_at ON transaction_products(updated_at);
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users(updated_at);

-- Composite indexes for efficient sync queries (license_id + updated_at)
-- These optimize the common sync pattern: WHERE shops.license_id = ? AND entity.updated_at > ?

-- For entities that directly have license_id through shops relationship
CREATE INDEX IF NOT EXISTS idx_carts_shop_updated ON carts(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_categories_shop_updated ON categories(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_products_shop_updated ON products(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_transactions_shop_updated ON transactions(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_payments_shop_updated ON payments(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_expenses_shop_updated ON expenses(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_receipts_shop_updated ON receipts(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_histories_shop_updated ON histories(shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_stock_histories_shop_updated ON stock_histories(shop_id, updated_at);

-- For entities with product_id relationship
CREATE INDEX IF NOT EXISTS idx_transaction_products_product_updated ON transaction_products(product_id, updated_at);

-- For users (direct license_id relationship)
CREATE INDEX IF NOT EXISTS idx_users_license_updated ON users(license_id, updated_at);

-- Additional indexes for sync metadata and conflict resolution
CREATE INDEX IF NOT EXISTS idx_shops_license_updated ON shops(license_id, updated_at);

-- Indexes for common filter operations during sync
CREATE INDEX IF NOT EXISTS idx_carts_user_shop_updated ON carts(user_id, shop_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_transactions_cashier_updated ON transactions(cashier_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_updated ON payments(transaction_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_receipts_payment_updated ON receipts(payments_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_histories_transaction_updated ON histories(transaction_id, updated_at);

-- Comment explaining the optimization strategy
COMMENT ON INDEX idx_carts_updated_at IS 'Optimizes sync queries filtering by updated_at timestamp';
COMMENT ON INDEX idx_carts_shop_updated IS 'Optimizes sync queries with shop_id and updated_at filters';
COMMENT ON INDEX idx_users_license_updated IS 'Optimizes user sync queries with license_id and updated_at filters';