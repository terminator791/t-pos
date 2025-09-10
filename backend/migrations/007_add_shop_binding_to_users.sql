-- Migration: Add shop binding for cashiers
-- Allow cashiers to be bound to a specific shop while owner business can access all shops under their license

-- Add shop_id to users table for cashiers
ALTER TABLE users ADD COLUMN shop_id UUID REFERENCES shops(id) ON DELETE SET NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_users_shop_id ON users(shop_id);

-- Add comment for clarity
COMMENT ON COLUMN users.shop_id IS 'Shop assignment for cashiers. NULL for owner business and other roles that can access multiple shops';