# t-pos Backend API

A modern Point of Sale (POS) system backend built with Go, implementing Clean Architecture principles.

## Overview

The t-pos backend is a RESTful API service that provides comprehensive functionality for managing a Point of Sale system. It supports user management, product catalog, inventory tracking, order processing, payment handling, and business analytics.

## Architecture

This project follows **Clean Architecture** principles with the following layers:

### 1. Domain Layer (`internal/domain/`)
- **Entities** (`entity/`): Core business entities (User, Product, Order, etc.)
- **Repository Interfaces** (`repository/`): Data access abstractions
- **Use Case Interfaces** (`usecase/`): Business logic abstractions

### 2. Infrastructure Layer (`internal/infrastructure/`)
- **Database** (`database/`): Database connection and migrations
- **Config** (`config/`): Configuration management
- **Logger** (`logger/`): Logging functionality

### 3. Delivery Layer (`internal/delivery/`)
- **HTTP** (`http/`): REST API handlers, middleware, and routing
- **gRPC** (`grpc/`): gRPC service implementations (future)

### 4. Shared Packages (`pkg/`)
- **Utils** (`utils/`): Common utility functions
- **Constants** (`constants/`): Application constants
- **Errors** (`errors/`): Custom error types

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Configuration**: Environment variables with godotenv
- **Logging**: Custom structured logging

## Features

### Core Functionality
- ✅ User authentication and authorization
- ✅ Customer management
- ✅ Product catalog with categories
- ✅ Inventory management
- ✅ Order processing
- ✅ Payment processing (multiple methods)
- ✅ Sales analytics and reporting

### API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration

#### Users
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

#### Customers
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `GET /api/v1/customers/:id` - Get customer by ID
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

#### Categories
- `GET /api/v1/categories` - List categories
- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories/:id` - Get category by ID
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

#### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:id` - Get product by ID
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product
- `GET /api/v1/products/search` - Search products
- `GET /api/v1/products/low-stock` - Get low stock products

#### Orders
- `GET /api/v1/orders` - List orders
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/:id` - Get order by ID
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Cancel order
- `POST /api/v1/orders/:id/items` - Add order item
- `DELETE /api/v1/orders/:id/items/:item_id` - Remove order item

#### Payments
- `GET /api/v1/payments` - List payments
- `POST /api/v1/payments` - Process payment
- `GET /api/v1/payments/:id` - Get payment by ID
- `POST /api/v1/payments/:id/refund` - Refund payment

#### Analytics
- `GET /api/v1/analytics/dashboard` - Dashboard statistics
- `GET /api/v1/analytics/sales` - Sales reports
- `GET /api/v1/analytics/products` - Product reports
- `GET /api/v1/analytics/customers` - Customer reports

## Database Schema

The system uses the following main entities:

- **Users**: System users (admin, manager, cashier)
- **Customers**: Customer information
- **Categories**: Product categories (hierarchical)
- **Products**: Product catalog with inventory
- **Orders**: Sales transactions
- **OrderItems**: Individual items in orders
- **Payments**: Payment transactions

## Getting Started

### Prerequisites
- Go 1.24 or higher
- PostgreSQL 12+
- Git

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd t-pos/backend
```

2. **Install dependencies**
```bash
go mod tidy
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Set up database**
```bash
# Create PostgreSQL database named 'tpos'
createdb tpos

# Run migrations
make db-migrate

# Seed initial data
make db-seed
```

5. **Run the application**
```bash
# Development mode
make run-backend

# Or directly with go
go run cmd/main.go
```

### Available Make Commands

From the project root directory:

```bash
# Run backend only
make run-backend

# Run both backend and frontend
make run

# Build backend
make build-backend

# Run tests
make test-backend

# Database operations
make db-migrate    # Run migrations
make db-seed      # Seed initial data

# Development helpers
make dev-backend  # Run with hot reload
make clean        # Clean build artifacts
```

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options.

### Key Configuration Options

- **Server**: Host, port, timeouts
- **Database**: Connection parameters
- **JWT**: Secret key and expiration times
- **Logging**: Level and file configuration

## Development

### Project Structure
```
backend/
├── cmd/                    # Application entry points
│   ├── main.go            # Main application
│   ├── migrate/           # Database migration command
│   └── seed/              # Database seeding command
├── internal/              # Private application code
│   ├── domain/            # Domain layer
│   │   ├── entity/        # Business entities
│   │   ├── repository/    # Repository interfaces
│   │   └── usecase/       # Use case interfaces
│   ├── delivery/          # Delivery layer
│   │   └── http/          # HTTP delivery
│   │       ├── handler/   # HTTP handlers
│   │       ├── middleware/# HTTP middleware
│   │       └── router/    # HTTP routing
│   └── infrastructure/    # Infrastructure layer
│       ├── config/        # Configuration
│       ├── database/      # Database connection
│       └── logger/        # Logging
├── pkg/                   # Public packages
│   ├── constants/         # Application constants
│   ├── errors/           # Custom errors
│   └── utils/            # Utility functions
├── docs/                  # Documentation
├── logs/                  # Log files
└── migrations/            # Database migrations
```

### Adding New Features

1. **Define entities** in `internal/domain/entity/`
2. **Create repository interfaces** in `internal/domain/repository/`
3. **Define use cases** in `internal/domain/usecase/`
4. **Implement repositories** in `internal/infrastructure/database/`
5. **Create HTTP handlers** in `internal/delivery/http/handler/`
6. **Add routes** in `internal/delivery/http/router/`

### Testing

```bash
# Run all tests
make test-backend

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/entity/
```

## API Documentation

### Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Response Format

All API responses follow this format:

```json
{
  "status": "success|error",
  "message": "Description of the result",
  "data": {}, // Response data (if applicable)
  "errors": [] // Error details (if applicable)
}
```

### Error Handling

The API returns appropriate HTTP status codes:

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

## Deployment

### Environment Variables

Ensure all required environment variables are set in production:

```bash
# Production settings
GIN_MODE=release
JWT_SECRET=your-production-secret
DB_HOST=your-production-db-host
# ... other production settings
```

### Database Setup

1. Create production database
2. Run migrations: `make db-migrate`
3. Optionally seed initial data: `make db-seed`

### Building for Production

```bash
# Build the application
make build-backend

# The binary will be created in backend/bin/t-pos
```

## Contributing

1. Follow Clean Architecture principles
2. Write tests for new features
3. Update documentation
4. Use conventional commit messages
5. Ensure all tests pass before submitting

## License

This project is licensed under the MIT License.

## Support

For support and questions, please create an issue in the repository or contact the development team.