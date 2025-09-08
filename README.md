# T-POS (Terminal Point of Sale)

A modern Point of Sale system with a Golang backend and React frontend, designed with Clean Architecture principles for scalability and maintainability.

## Project Structure

```
t-pos/
├── backend/           # Go backend with Clean Architecture
├── frontend/          # React frontend with Vite
├── Makefile          # Development workflow commands
└── README.md         # This file
```

## Quick Start

### Prerequisites
- Go 1.24+
- Node.js 18+
- Yarn
- PostgreSQL 12+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd t-pos
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Setup database**
   ```bash
   createdb tpos_db
   ```

4. **Configure environment**
   ```bash
   cp backend/.env.example backend/.env
   # Edit backend/.env with your database credentials
   ```

5. **Run the application**
   ```bash
   make run    # Runs both backend and frontend
   ```

## Available Commands

- `make run` - Run both backend and frontend concurrently
- `make run-backend` - Run only the backend server
- `make run-frontend` - Run only the frontend development server
- `make build-backend` - Build the backend application
- `make test-backend` - Run backend tests
- `make deps` - Install all dependencies
- `make clean` - Clean build artifacts

## Architecture

### Backend (Go + Clean Architecture)
- **Domain Layer**: Business entities and interfaces
- **Application Layer**: Use cases and business logic
- **Infrastructure Layer**: Database, repositories, external services
- **Interface Layer**: HTTP handlers, middleware, routes

### Frontend (React + Vite)
- Modern React application with TypeScript
- Tailwind CSS for styling
- Vite for fast development and building

## API Documentation

The backend provides a RESTful API with the following main endpoints:

- **Products**: `/api/v1/products/*`
- **Orders**: `/api/v1/orders/*`
- **Health Check**: `/health`

For detailed API documentation, see [backend/docs/BACKEND.md](backend/docs/BACKEND.md)

## Technology Stack

### Backend
- **Language**: Go 1.24+
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL

### Frontend
- **Framework**: React 18
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Package Manager**: Yarn

## Documentation

- [Backend Documentation](backend/docs/BACKEND.md)
- [Frontend Documentation](frontend/docs/FRONTEND.md)

## Development

### Database Schema

The system manages:
- Users (cashiers, managers, admins)
- Products and Categories
- Customers
- Orders and Order Items
- Payments

### Key Features

- Product management with inventory tracking
- Order processing with multiple payment methods
- Customer management
- Low stock alerts
- Sales reporting
- User role management

## Contributing

1. Follow Clean Architecture principles for backend changes
2. Write tests for new features
3. Update documentation for API changes
4. Follow Go and React best practices

## License

[Add your license information here]