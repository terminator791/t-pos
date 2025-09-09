# T-POS Backend Documentation

## Overview

The T-POS (Terminal Point of Sale) backend is a robust REST API built with Go, following Clean Architecture principles. It provides APIs for managing products, orders, customers, and users in a point-of-sale system.

## Architecture

This project follows Clean Architecture with clear separation of concerns:

```
backend/
├── cmd/                    # Application entry points
│   └── main.go            # Main application
├── config/                # Configuration management
├── internal/              # Private application code
│   ├── domain/           # Business logic layer
│   │   ├── entities/     # Business entities/models
│   │   ├── repositories/ # Repository interfaces
│   │   └── usecases/     # Business use cases
│   ├── infrastructure/   # External concerns
│   │   ├── database/     # Database connection
│   │   ├── repositories/ # Repository implementations
│   │   └── web/          # HTTP server setup
│   └── interfaces/       # External interfaces
│       ├── http/         # HTTP handlers and routes
│       └── cli/          # CLI interfaces
├── migrations/           # Database migrations
├── docs/                # Documentation
└── pkg/                 # Public libraries
```

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Configuration**: Environment variables with godotenv

## Database Schema

The system manages the following entities:

### Core Entities
- **Users**: System users (cashiers, managers, admins)
- **Categories**: Product categories
- **Products**: Items for sale
- **Customers**: Customer information
- **Orders**: Sales transactions
- **Order Items**: Individual items within an order
- **Payments**: Payment records for orders

### Key Relationships
- Products belong to Categories
- Orders belong to Users (cashier who processed the sale)
- Orders may belong to Customers (optional)
- Orders have many Order Items
- Orders have many Payments (split payments support)

## API Endpoints

### Products
- `GET /api/v1/products` - List products
- `GET /api/v1/products/:id` - Get product by ID
- `GET /api/v1/products/sku/:sku` - Get product by SKU
- `GET /api/v1/products/search?q=query` - Search products
- `GET /api/v1/products/low-stock` - Get low stock products
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

### Orders
- `GET /api/v1/orders` - List orders
- `GET /api/v1/orders/:id` - Get order by ID
- `GET /api/v1/orders/number/:orderNumber` - Get order by number
- `GET /api/v1/orders/today` - Get today's orders
- `POST /api/v1/orders` - Create order

### Health Check
- `GET /health` - Service health status

## Getting Started

### Prerequisites
- Go 1.24 or higher
- PostgreSQL 12 or higher
- Make (optional, for using Makefile)

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd t-pos/backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Setup environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

4. **Setup database**
   ```bash
   createdb tpos_db
   ```

5. **Run the application**
   ```bash
   go run cmd/main.go
   ```

### Using Make Commands

From the root directory:
```bash
make run-backend    # Run backend only
make run           # Run both backend and frontend
make build-backend # Build backend binary
make test-backend  # Run tests
```

## Development Guidelines

### Adding New Features

1. **Domain Layer**: Start by defining entities and repository interfaces
2. **Use Cases**: Implement business logic in use cases
3. **Infrastructure**: Implement repository interfaces
4. **Interface**: Add HTTP handlers and routes

### Code Organization

- Keep business logic in the domain layer
- External dependencies should only be in infrastructure layer
- Use dependency injection for loose coupling
- Follow Go naming conventions and best practices

### Database Migrations

GORM AutoMigrate is used for development. For production, consider using proper migration tools.

### Error Handling

- Use proper HTTP status codes
- Return meaningful error messages
- Log errors appropriately

## Testing

```bash
go test ./...
```

Tests should be organized by layer:
- Unit tests for use cases
- Integration tests for repositories
- End-to-end tests for HTTP handlers

## Configuration

All configuration is handled through environment variables. See `.env.example` for available options.

## Production Deployment

1. Build the binary: `make build-backend`
2. Set production environment variables
3. Run database migrations
4. Deploy the binary with proper process management

## Contributing

1. Follow Clean Architecture principles
2. Write tests for new features
3. Document API changes
4. Follow Go best practices and formatting

## Support

For questions or issues, please refer to the project repository or contact the development team.