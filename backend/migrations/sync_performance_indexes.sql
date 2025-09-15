-- Performance optimization indexes for T-POS Sync Service
-- These indexes are specifically designed to optimize sync queries

-- Index for shops table: license_id lookup (used in bulk validation)
CREATE INDEX IF NOT EXISTS idx_shops_license_id ON shops(license_id);

-- Index for products table: shop_id + updated_at (used in sync pull operations)
CREATE INDEX IF NOT EXISTS idx_products_shop_updated ON products(shop_id, updated_at);

-- Index for products table: id lookup for bulk operations
CREATE INDEX IF NOT EXISTS idx_products_id ON products(id);

-- Index for transactions table: shop_id + updated_at (used in sync pull operations)  
CREATE INDEX IF NOT EXISTS idx_transactions_shop_updated ON transactions(shop_id, updated_at);

-- Index for transaction_products table: transaction_id + updated_at
CREATE INDEX IF NOT EXISTS idx_transaction_products_transaction_updated ON transaction_products(transaction_id, updated_at);

-- Index for carts table: shop_id + updated_at
CREATE INDEX IF NOT EXISTS idx_carts_shop_updated ON carts(shop_id, updated_at);

-- Index for categories table: shop_id + updated_at
CREATE INDEX IF NOT EXISTS idx_categories_shop_updated ON categories(shop_id, updated_at);

-- Index for expenses table: shop_id + updated_at
CREATE INDEX IF NOT EXISTS idx_expenses_shop_updated ON expenses(shop_id, updated_at);

-- Index for payments table: transaction_id + updated_at
CREATE INDEX IF NOT EXISTS idx_payments_transaction_updated ON payments(transaction_id, updated_at);

-- Index for receipts table: transaction_id + updated_at
CREATE INDEX IF NOT EXISTS idx_receipts_transaction_updated ON receipts(transaction_id, updated_at);

-- Index for histories table: shop_id + updated_at
CREATE INDEX IF NOT EXISTS idx_histories_shop_updated ON histories(shop_id, updated_at);

-- Index for stock_histories table: product_id + updated_at (critical for stock history sync)
CREATE INDEX IF NOT EXISTS idx_stock_histories_product_updated ON stock_histories(product_id, updated_at);

-- Index for stock_histories table: shop_id + updated_at (for shop-level filtering)
CREATE INDEX IF NOT EXISTS idx_stock_histories_shop_updated ON stock_histories(shop_id, updated_at);

-- Index for users table: license_id + updated_at
CREATE INDEX IF NOT EXISTS idx_users_license_updated ON users(license_id, updated_at);

-- Composite index for license-based entity filtering (covers multiple tables)
-- This is particularly useful for role-based access control queries

-- Index for shops table: license_id + id (for shop ID validation)
CREATE INDEX IF NOT EXISTS idx_shops_license_id_composite ON shops(license_id, id);

-- Performance monitoring: Create index on commonly used foreign key relationships
-- These help with JOIN operations during sync validation

-- Index for products table: shop_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_products_shop_id ON products(shop_id);

-- Index for transactions table: shop_id (foreign key index)  
CREATE INDEX IF NOT EXISTS idx_transactions_shop_id ON transactions(shop_id);

-- Index for stock_histories table: product_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_stock_histories_product_id ON stock_histories(product_id);

-- Index for transaction_products table: transaction_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_transaction_products_transaction_id ON transaction_products(transaction_id);

-- Index for transaction_products table: product_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_transaction_products_product_id ON transaction_products(product_id);

-- Index for payments table: transaction_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);

-- Index for receipts table: transaction_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_receipts_transaction_id ON receipts(transaction_id);

-- Index for carts table: user_id + shop_id (for user-shop cart filtering)
CREATE INDEX IF NOT EXISTS idx_carts_user_shop ON carts(user_id, shop_id);

-- Index for carts table: product_id (foreign key index)
CREATE INDEX IF NOT EXISTS idx_carts_product_id ON carts(product_id);

-- Special indexes for performance-critical sync queries

-- Index for last sync timestamp queries (covers all sync entities)
-- This helps with "updated_at > last_sync" queries across all tables

-- Index for entity counting queries (used in memory estimation)
CREATE INDEX IF NOT EXISTS idx_products_count ON products(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_count ON transactions(shop_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_carts_count ON carts(shop_id) WHERE deleted_at IS NULL;

-- Partial indexes for active/non-deleted records only (improves query performance)
CREATE INDEX IF NOT EXISTS idx_products_active_shop_updated ON products(shop_id, updated_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_active_shop_updated ON transactions(shop_id, updated_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_carts_active_shop_updated ON carts(shop_id, updated_at) WHERE deleted_at IS NULL;

-- Covering indexes for common SELECT operations (includes frequently accessed columns)
-- These indexes include the columns needed for the query, avoiding table lookups

-- Covering index for product sync queries (includes essential product fields)
CREATE INDEX IF NOT EXISTS idx_products_sync_covering ON products(shop_id, updated_at) 
    INCLUDE (id, name, price_buy, price_sell, stock, barcode, deleted_at);

-- Covering index for transaction sync queries (includes essential transaction fields)  
CREATE INDEX IF NOT EXISTS idx_transactions_sync_covering ON transactions(shop_id, updated_at)
    INCLUDE (id, total_amount, status, discount_percentage, deleted_at);

-- Statistics and monitoring indexes
-- These help with sync performance monitoring and analysis

-- Index for sync operation timing analysis
CREATE INDEX IF NOT EXISTS idx_products_created_updated ON products(created_at, updated_at);
CREATE INDEX IF NOT EXISTS idx_transactions_created_updated ON transactions(created_at, updated_at);

-- Index for entity modification frequency analysis
CREATE INDEX IF NOT EXISTS idx_stock_histories_updated_at ON stock_histories(updated_at);
CREATE INDEX IF NOT EXISTS idx_payments_updated_at ON payments(updated_at);

-- Comments for maintenance and monitoring
COMMENT ON INDEX idx_shops_license_id IS 'Performance index for sync license validation queries';
COMMENT ON INDEX idx_products_shop_updated IS 'Core index for product sync pull operations';
COMMENT ON INDEX idx_transactions_shop_updated IS 'Core index for transaction sync pull operations';
COMMENT ON INDEX idx_stock_histories_product_updated IS 'Critical index for stock history sync operations';
COMMENT ON INDEX idx_products_sync_covering IS 'Covering index to avoid table lookups in product sync';
COMMENT ON INDEX idx_transactions_sync_covering IS 'Covering index to avoid table lookups in transaction sync';