# T-POS Authentication & Authorization Guide

## Overview

This T-POS system now includes comprehensive JWT authentication and Casbin-based authorization with RBAC + domain/tenant support.

## Quick Start

### 1. Environment Setup

Create a `.env` file in the backend directory:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=tpos_db
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRY_HOUR=24

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
```

### 2. Database Setup

```bash
# Create database
createdb tpos_db

# Run migrations (creates all tables including auth tables)
cd backend
go run cmd/migrate/main.go up
```

### 3. Start Server

```bash
cd backend
go run cmd/main.go
```

The server will automatically:
- Create auth tables (roles, user_roles, policies, casbin_rule)
- Seed default roles: super_admin, admin, manager, cashier, user
- Seed default policies for each role

## API Usage

### Authentication Endpoints

#### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "username",
    "name": "Full Name",
    "password": "securepassword",
    "domain": "shop1"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "domain": "shop1"
  }'
```

Response includes JWT token:
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {...},
    "roles": ["user"],
    "domain": "shop1"
  }
}
```

#### Access Protected Endpoints
```bash
curl -X GET http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### User Management

#### Get Profile
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get User Permissions
```bash
curl -X GET http://localhost:8080/api/v1/auth/permissions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Role-Based Permissions

### Default Roles

1. **super_admin**: Full system access
   - All API endpoints
   - All HTTP methods
   - All domains

2. **admin**: Domain administrative access
   - Full CRUD on products, transactions, users
   - Domain-specific access

3. **manager**: Shop management
   - Product management (create, read, update)
   - Transaction viewing
   - User viewing

4. **cashier**: Point of sale operations
   - Product reading and searching
   - Checkout operations
   - Transaction details

5. **user**: Basic access
   - Product viewing
   - Profile management
   - Authentication operations

### Domain/Tenant Support

Users can have different roles in different domains:
- A user might be an "admin" in "shop1" but a "user" in "shop2"
- Permissions are checked per domain
- Default domain is "*" (global)

### URL Pattern Matching

Casbin uses KeyMatch2 for flexible URL patterns:
- `/api/v1/products/*` matches `/api/v1/products/123`
- `/api/v1/*` matches all v1 API endpoints
- Regex support for HTTP methods: `GET|POST|PUT|DELETE`

## Security Features

### JWT Tokens
- HS256 signing (configurable to RS256)
- Configurable expiry time
- Secure token validation
- Refresh token capability

### Password Security
- Bcrypt hashing with secure defaults
- No plain text password storage
- Password verification on login

### Authorization
- Request-level permission checking
- Domain/tenant isolation
- URL pattern matching
- Database-backed policy storage

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific auth tests
go test ./internal/infrastructure/auth -v

# Run response utility tests
go test ./pkg/response -v
```

### Adding New Roles
1. Add role to seeder: `internal/infrastructure/seeders/auth_seeder.go`
2. Add policies for the role
3. Restart server or reload policies

### Adding New Policies
```go
// Add to auth_seeder.go
{"new_role", "*", "/api/v1/new-endpoint/*", "GET|POST"}
```

### Custom Middleware
The system provides both authentication and authorization middleware:
- `authMiddleware.RequireAuth()` - JWT validation
- `authzMiddleware.RequirePermission()` - Casbin authorization
- `authzMiddleware.RequireRole("admin")` - Role-specific access

## Troubleshooting

### Common Issues

1. **Database Connection**: Ensure PostgreSQL is running and credentials are correct
2. **JWT Secret**: Use a strong, random secret key
3. **Migration Errors**: Drop and recreate database if schema changes
4. **Permission Denied**: Check user roles and policies in database

### Debug Queries

```sql
-- Check user roles
SELECT u.email, r.name as role, ur.domain 
FROM users u 
JOIN user_roles ur ON u.id = ur.user_id 
JOIN roles r ON ur.role_id = r.id;

-- Check policies
SELECT p.subject, p.domain, p.object, p.action 
FROM policies p 
WHERE p.is_active = true;

-- Check Casbin rules
SELECT * FROM casbin_rule;
```

## API Response Format

All API endpoints use consistent response format:

### Success Response
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": { ... }
}
```

### Error Response
```json
{
  "status": "failed",
  "message": "Error description",
  "errors": { ... }
}
```

This ensures consistent client-side handling across the entire API.