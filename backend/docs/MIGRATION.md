# Database Migration Guide

This guide explains how to manage database migrations in the T-POS backend using GORM.

## Migration Commands

The project includes a migration CLI tool that provides Laravel Artisan-like functionality for database management.

### Available Commands

| Command               | Description                           | Laravel Equivalent                   |
| --------------------- | ------------------------------------- | ------------------------------------ |
| `make migrate-up`     | Run pending migrations                | `php artisan migrate`                |
| `make migrate-down`   | Drop all tables (rollback)            | `php artisan migrate:rollback --all` |
| `make migrate-fresh`  | Drop all tables and re-run migrations | `php artisan migrate:fresh`          |
| `make migrate-status` | Check migration status                | `php artisan migrate:status`         |

### Manual Commands

You can also run commands directly:

```bash
# Navigate to backend directory
cd backend

# Run migrations
go run cmd/migrate/main.go up

# Drop all tables
go run cmd/migrate/main.go down

# Fresh migration (drop + recreate)
go run cmd/migrate/main.go fresh

# Check status
go run cmd/migrate/main.go status
```

## UUID Implementation Status

All tables except `logs` and `license_logs` now use UUID as primary keys:

### ✅ Tables with UUID Primary Keys:

- `licenses` - UUID (original)
- `users` - UUID (original)
- `transactions` - UUID (original)
- `shops` - UUID ✅
- `categories` - UUID ✅
- `products` - UUID ✅
- `carts` - UUID ✅
- `transaction_products` - UUID ✅
- `payments` - UUID ✅
- `receipts` - UUID ✅
- `histories` - UUID ✅
- `stock_histories` - UUID ✅
- `expenses` - UUID ✅

### ❌ Tables with BIGSERIAL Primary Keys (By Design):

- `logs` - BIGSERIAL/uint (for performance)
- `license_logs` - BIGSERIAL/uint (for compatibility)

## GORM Auto Migration

The project uses GORM's `AutoMigrate` feature which:

1. **Creates tables** if they don't exist
2. **Adds missing columns** when you add new fields to entities
3. **Creates indexes** defined in struct tags
4. **Does NOT drop columns** (safe operation)
5. **Does NOT modify column types** (safe operation)

### Entity Structure

All entities follow this pattern:

```go
type EntityName struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    // ... other fields
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### Foreign Key Relationships

All foreign keys are properly defined with constraints:

```go
ShopID    uuid.UUID `gorm:"type:uuid;not null" json:"shop_id"`
Shop      Shop      `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
```

## Usage Examples

### 1. First Time Setup

```bash
# Make sure PostgreSQL is running and database exists
createdb tpos_db

# Run initial migration
make migrate-fresh
```

### 2. After Adding New Entity Fields

```bash
# Just run migrate up (GORM will add new columns)
make migrate-up
```

### 3. Complete Database Reset

```bash
# Warning: This will delete ALL data
make migrate-fresh
```

### 4. Check Current Status

```bash
make migrate-status
```

## Database Connection

The migration tool uses the same configuration as the main application:

- Database config from `.env` file
- PostgreSQL with UUID extension
- Connection pooling via GORM

## Safety Notes

1. **`migrate-fresh` deletes all data** - use with caution in production
2. **`migrate-down` drops all tables** - backup first in production
3. **GORM AutoMigrate is safe** - it only adds, never removes
4. **Foreign key constraints** are properly handled during drops

## Troubleshooting

### Connection Issues

```bash
# Check if PostgreSQL is running
pg_ctl status

# Check if database exists
psql -l | grep tpos
```

### Permission Issues

```bash
# Make sure user has proper permissions
GRANT ALL PRIVILEGES ON DATABASE tpos_db TO your_user;
```

### UUID Extension Issues

```bash
# Connect to database and enable UUID extension
psql tpos_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```
