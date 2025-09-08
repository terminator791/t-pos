# t-pos - Point of Sale System

A modern, full-stack Point of Sale (POS) system built with Go backend and React frontend, designed for retail businesses of all sizes.

## 🚀 Features

- **User Management**: Multi-role user system (Admin, Manager, Cashier)
- **Customer Management**: Complete customer database with purchase history
- **Product Catalog**: Hierarchical categories with comprehensive product management
- **Inventory Tracking**: Real-time stock management with low-stock alerts
- **Order Processing**: Full order lifecycle management
- **Payment Processing**: Support for multiple payment methods
- **Sales Analytics**: Comprehensive reporting and dashboard
- **Clean Architecture**: Scalable, maintainable codebase

## 🏗️ Architecture

### Backend (Go)
- **Clean Architecture** implementation
- **RESTful API** with Gin framework
- **PostgreSQL** database with GORM ORM
- **JWT Authentication**
- **Comprehensive logging**

### Frontend (React)
- **React 18** with modern hooks
- **Vite** for fast development
- **Tailwind CSS** for styling
- **Redux Toolkit** for state management

## 📋 Prerequisites

- **Go 1.24+**
- **Node.js 16+** and **Yarn**
- **PostgreSQL 12+**
- **Git**

## 🚀 Quick Start

### 1. Clone the Repository
```bash
git clone <repository-url>
cd t-pos
```

### 2. Backend Setup
```bash
# Navigate to backend directory
cd backend

# Copy environment file and configure
cp .env.example .env
# Edit .env with your database credentials

# Install dependencies
go mod tidy

# Set up database (ensure PostgreSQL is running)
createdb tpos

# Run migrations and seed data
make db-migrate
make db-seed
```

### 3. Frontend Setup
```bash
# Navigate to frontend directory
cd ../frontend

# Install dependencies
yarn install
```

### 4. Run the Application

From the project root directory:

```bash
# Run both backend and frontend
make run

# Or run individually:
make run-backend    # Backend only
make run-frontend   # Frontend only
```

The application will be available at:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## 📝 Available Commands

### Root Level Commands
```bash
make help           # Show all available commands
make run            # Run both backend and frontend
make run-backend    # Run backend only
make run-frontend   # Run frontend only
make build-backend  # Build backend
make build          # Build both backend and frontend
make test-backend   # Run backend tests
make clean          # Clean build artifacts
make setup          # Setup development environment
make db-migrate     # Run database migrations
make db-seed        # Seed database with sample data
```

### Backend Commands
```bash
cd backend
go run cmd/main.go           # Run main application
go run cmd/migrate/main.go   # Run migrations
go run cmd/seed/main.go      # Seed database
go test ./...                # Run tests
go build -o bin/t-pos cmd/main.go  # Build binary
```

### Frontend Commands
```bash
cd frontend
yarn dev        # Development server
yarn build      # Production build
yarn preview    # Preview production build
```

## 🔧 Configuration

### Backend Configuration

Copy `backend/.env.example` to `backend/.env` and configure:

```env
# Server
SERVER_HOST=localhost
SERVER_PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=tpos

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24
```

### Database Setup

1. **Install PostgreSQL** and ensure it's running
2. **Create database**:
   ```bash
   createdb tpos
   ```
3. **Run migrations**:
   ```bash
   make db-migrate
   ```
4. **Seed initial data**:
   ```bash
   make db-seed
   ```

## 📖 API Documentation

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration

### Core Resources
- **Users**: `/api/v1/users`
- **Customers**: `/api/v1/customers`
- **Categories**: `/api/v1/categories`
- **Products**: `/api/v1/products`
- **Orders**: `/api/v1/orders`
- **Payments**: `/api/v1/payments`
- **Analytics**: `/api/v1/analytics`

### Default Login Credentials
After seeding the database:
- **Email**: admin@tpos.com
- **Password**: password

## 📁 Project Structure

```
t-pos/
├── backend/                 # Go backend
│   ├── cmd/                # Application entry points
│   ├── internal/           # Private application code
│   │   ├── domain/         # Domain layer (entities, interfaces)
│   │   ├── delivery/       # Delivery layer (HTTP handlers)
│   │   └── infrastructure/ # Infrastructure layer (DB, config)
│   ├── pkg/                # Public packages
│   ├── docs/               # Backend documentation
│   └── migrations/         # Database migrations
├── frontend/               # React frontend
│   ├── src/                # Source code
│   │   ├── components/     # Reusable components
│   │   ├── pages/          # Page components
│   │   ├── store/          # Redux store
│   │   └── hooks/          # Custom hooks
│   └── docs/               # Frontend documentation
├── Makefile                # Development commands
└── README.md               # This file
```

## 🧪 Testing

### Backend Tests
```bash
make test-backend
# Or
cd backend && go test ./...
```

### API Testing
Use the health endpoint to verify the backend is running:
```bash
curl http://localhost:8080/health
```

## 🚀 Deployment

### Building for Production
```bash
# Build backend
make build-backend

# Build frontend
cd frontend && yarn build
```

### Environment Variables
Ensure all production environment variables are properly set, especially:
- `GIN_MODE=release`
- `JWT_SECRET` (use a strong, unique secret)
- Database connection parameters

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add some amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines
- Follow Clean Architecture principles
- Write tests for new features
- Update documentation
- Use conventional commit messages

## 📚 Documentation

- **Backend Documentation**: [`backend/docs/BACKEND.md`](backend/docs/BACKEND.md)
- **Frontend Documentation**: [`frontend/docs/FRONTEND.md`](frontend/docs/FRONTEND.md)

## 🛠️ Technology Stack

### Backend
- **Go 1.24+**
- **Gin** - HTTP web framework
- **GORM** - ORM library
- **PostgreSQL** - Database
- **JWT** - Authentication
- **Godotenv** - Environment configuration

### Frontend
- **React 18**
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **Redux Toolkit** - State management
- **React Router** - Routing

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
1. Check the documentation in the `docs/` folders
2. Create an issue in the repository
3. Contact the development team

## 🔄 Development Workflow

1. **Setup**: `make setup`
2. **Develop**: `make run` (runs both backend and frontend)
3. **Test**: `make test-backend`
4. **Build**: `make build`
5. **Deploy**: Use production builds

---

**Happy coding! 🎉**