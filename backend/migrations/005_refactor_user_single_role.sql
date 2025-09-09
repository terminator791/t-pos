-- Migration: Refactor User Single Role System
-- Each user has only one role but can access multiple domains

-- First, backup existing data
CREATE TABLE IF NOT EXISTS user_roles_backup AS 
SELECT * FROM user_roles;

-- Add role_id directly to users table
ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles(id);

-- Migrate existing data (take the first role for each user)
UPDATE users 
SET role_id = (
    SELECT ur.role_id 
    FROM user_roles ur 
    WHERE ur.user_id = users.id 
    LIMIT 1
);

-- Create user_domains table for domain access control
CREATE TABLE IF NOT EXISTS user_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, domain)
);

-- Migrate domain data from user_roles to user_domains
INSERT INTO user_domains (user_id, domain, created_at, updated_at)
SELECT DISTINCT user_id, domain, created_at, updated_at
FROM user_roles
WHERE deleted_at IS NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_domains_user_id ON user_domains(user_id);
CREATE INDEX IF NOT EXISTS idx_user_domains_domain ON user_domains(domain);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);

-- Note: We'll keep user_roles table for now as backup
-- DROP TABLE user_roles; -- Uncomment this after testing
