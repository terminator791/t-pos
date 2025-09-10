-- Remove user_id column from transactions table
-- Since user_id is redundant with cashier_id

-- Drop the foreign key constraint first
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_user_id_fkey;

-- Drop the index on user_id
DROP INDEX IF EXISTS idx_transactions_user_id;

-- Drop the user_id column
ALTER TABLE transactions DROP COLUMN IF EXISTS user_id;
