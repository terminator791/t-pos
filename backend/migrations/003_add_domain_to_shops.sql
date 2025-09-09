-- Migration: Add domain column to shops table
-- This adds a domain column for multi-tenancy support

-- Add domain column to shops table with default value first
ALTER TABLE shops 
ADD COLUMN domain VARCHAR(100) NOT NULL DEFAULT '';

-- Update existing shops with default domain pattern using their ID
-- This is for backward compatibility
UPDATE shops SET domain = 'shop-' || REPLACE(id::text, '-', '') WHERE domain = '';

-- Create unique index for domain
CREATE UNIQUE INDEX IF NOT EXISTS idx_shops_domain ON shops(domain);

-- Add comment for documentation
COMMENT ON COLUMN shops.domain IS 'Unique domain identifier for multi-tenant shop access';
