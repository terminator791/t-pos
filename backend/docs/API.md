# T-POS API Documentation

This document provides comprehensive documentation for the T-POS (Terminal Point of Sale) REST API, including license, customer, and user management endpoints.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All API endpoints require JWT authentication with domain-based authorization using Casbin. The JWT token must contain:
- `user_id`: The authenticated user's UUID
- `domain`: The license serial number for multi-tenant access control

## Response Format
All API responses follow a standardized format:

### Success Response
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    // Response data object
  }
}
```

### Error Response
```json
{
  "status": "failed",
  "message": "Error description",
  "errors": {
    // Error details or validation errors
  }
}
```

## HTTP Status Codes
- `200 OK` - Successful GET/PUT/DELETE operations
- `201 Created` - Successful POST operations
- `400 Bad Request` - Invalid request data or validation errors
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server-side errors

---

## License Management

### Get All Licenses
Retrieve all licenses with pagination support.

**Endpoint:** `GET /licenses`

**Query Parameters:**
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Number of records to skip (default: 0)

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/licenses?limit=20&offset=0" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Licenses retrieved successfully",
  "data": {
    "licenses": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "serial_number": "LICENSE-2024-001",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "count": 1,
    "limit": 20,
    "offset": 0
  }
}
```

### Get License by ID
Retrieve a specific license by its UUID.

**Endpoint:** `GET /licenses/{id}`

**Path Parameters:**
- `id` (required): License UUID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/licenses/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "License retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "serial_number": "LICENSE-2024-001",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Create License
Create a new license with automatic license log entry.

**Endpoint:** `POST /licenses`

**Request Body:**
```json
{
  "serial_number": "LICENSE-2024-002"
}
```

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/licenses" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "serial_number": "LICENSE-2024-002"
  }'
```

**Example Response:**
```json
{
  "status": "success",
  "message": "License created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "serial_number": "LICENSE-2024-002",
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

**Notes:**
- Creates a corresponding entry in the `license_logs` table
- Operation is wrapped in a database transaction for data integrity

### Delete License
Delete a license and all associated license logs.

**Endpoint:** `DELETE /licenses/{id}`

**Path Parameters:**
- `id` (required): License UUID

**Example Request:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/licenses/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "License deleted successfully",
  "data": null
}
```

**Notes:**
- Deletes all associated license logs
- Operation is wrapped in a database transaction

---

## Customer Management

Customer endpoints manage users with `cashier` or `owner_business` roles only.

### Get All Customers
Retrieve all customers (users with cashier or owner_business roles).

**Endpoint:** `GET /customers`

**Query Parameters:**
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Number of records to skip (default: 0)

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/customers?limit=10&offset=0" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Customers retrieved successfully",
  "data": {
    "customers": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "username": "cashier_user",
        "role_id": "cashier",
        "serial_number": "LICENSE-2024-001",
        "created_at": "2024-01-15T09:00:00Z",
        "updated_at": "2024-01-15T09:00:00Z"
      }
    ],
    "count": 1,
    "limit": 10,
    "offset": 0
  }
}
```

### Get Customer by ID
Retrieve a specific customer by UUID.

**Endpoint:** `GET /customers/{id}`

**Path Parameters:**
- `id` (required): Customer UUID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/customers/660e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Customer retrieved successfully",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "username": "cashier_user",
    "role_id": "cashier",
    "serial_number": "LICENSE-2024-001",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T09:00:00Z"
  }
}
```

### Create Customer
Create a new customer user.

**Endpoint:** `POST /customers`

**Request Body:**
```json
{
  "username": "new_cashier",
  "serial_number": "LICENSE-2024-001",
  "role_id": "cashier",
  "pin": "123456"
}
```

**Valid Role IDs:**
- `cashier`
- `owner_business`

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/customers" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_cashier",
    "serial_number": "LICENSE-2024-001",
    "role_id": "cashier",
    "pin": "123456"
  }'
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Customer created successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "username": "new_cashier",
    "role_id": "cashier",
    "serial_number": "LICENSE-2024-001",
    "created_at": "2024-01-15T12:00:00Z",
    "updated_at": "2024-01-15T12:00:00Z"
  }
}
```

**Notes:**
- PIN is securely hashed using bcrypt
- Validates that the license exists
- Only allows `cashier` or `owner_business` roles

### Delete Customer
Remove a customer.

**Endpoint:** `DELETE /customers/{id}`

**Path Parameters:**
- `id` (required): Customer UUID

**Example Request:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/customers/660e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Customer deleted successfully",
  "data": null
}
```

---

## User Management

User management endpoints handle users with `admin` or `super_admin` roles only.

### Get All Users
Retrieve all admin users.

**Endpoint:** `GET /users`

**Query Parameters:**
- `limit` (optional): Number of results per page (default: 10)
- `offset` (optional): Number of records to skip (default: 0)

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/users?limit=10&offset=0" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "Users retrieved successfully",
  "data": {
    "users": [
      {
        "id": "880e8400-e29b-41d4-a716-446655440000",
        "username": "admin_user",
        "role_id": "admin",
        "serial_number": "LICENSE-2024-001",
        "created_at": "2024-01-15T08:00:00Z",
        "updated_at": "2024-01-15T08:00:00Z"
      }
    ],
    "count": 1,
    "limit": 10,
    "offset": 0
  }
}
```

### Get User by ID
Retrieve a specific admin user by UUID.

**Endpoint:** `GET /users/{id}`

**Path Parameters:**
- `id` (required): User UUID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/users/880e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "User retrieved successfully",
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "username": "admin_user",
    "role_id": "admin",
    "serial_number": "LICENSE-2024-001",
    "created_at": "2024-01-15T08:00:00Z",
    "updated_at": "2024-01-15T08:00:00Z"
  }
}
```

### Create User
Create a new admin user.

**Endpoint:** `POST /users`

**Request Body:**
```json
{
  "username": "new_admin",
  "serial_number": "LICENSE-2024-001",
  "role_id": "admin",
  "pin": "admin123"
}
```

**Valid Role IDs:**
- `admin`
- `super_admin`

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_admin",
    "serial_number": "LICENSE-2024-001",
    "role_id": "admin",
    "pin": "admin123"
  }'
```

**Example Response:**
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": "990e8400-e29b-41d4-a716-446655440000",
    "username": "new_admin",
    "role_id": "admin",
    "serial_number": "LICENSE-2024-001",
    "created_at": "2024-01-15T13:00:00Z",
    "updated_at": "2024-01-15T13:00:00Z"
  }
}
```

**Notes:**
- PIN is securely hashed using bcrypt
- Validates that the license exists
- Only allows `admin` or `super_admin` roles

### Update User Password
Update an admin user's password.

**Endpoint:** `PUT /users/{id}`

**Path Parameters:**
- `id` (required): User UUID

**Request Body:**
```json
{
  "password": "new_admin_password"
}
```

**Example Request:**
```bash
curl -X PUT "http://localhost:8080/api/v1/users/880e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "new_admin_password"
  }'
```

**Example Response:**
```json
{
  "status": "success",
  "message": "User password updated successfully",
  "data": null
}
```

**Notes:**
- Password is securely hashed using bcrypt
- Only works for users with `admin` or `super_admin` roles

### Delete User
Remove an admin user.

**Endpoint:** `DELETE /users/{id}`

**Path Parameters:**
- `id` (required): User UUID

**Example Request:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/users/880e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer {jwt_token}"
```

**Example Response:**
```json
{
  "status": "success",
  "message": "User deleted successfully",
  "data": null
}
```

---

## Domain-Based Authorization

The API implements Casbin domain-based authorization where:

- **Domain = License Serial Number**: Each JWT token contains the license serial number as the domain
- **Multi-Tenant Access Control**: Users can only access resources within their licensed domain
- **Role-Based Permissions**: Different roles have different access levels within their domain

### Domain Mapping
```
Domain (JWT) → License Serial Number → License ID → Resource Access
```

### Authorization Rules
1. All requests must include a valid JWT token with domain information
2. Users can only access resources associated with their license domain
3. Role-based access control restricts operations based on user roles:
   - `super_admin`: Full system access
   - `admin`: Admin operations within domain
   - `owner_business`: Business operations within domain
   - `cashier`: Point-of-sale operations within domain

---

## Error Handling

### Common Error Responses

**Invalid Request Body (400)**
```json
{
  "status": "failed",
  "message": "Invalid request body",
  "errors": "Key: 'CreateLicenseRequest.SerialNumber' Error:Tag Required"
}
```

**Unauthorized Access (401)**
```json
{
  "status": "failed",
  "message": "User not authenticated",
  "errors": "User ID not found in context"
}
```

**Resource Not Found (404)**
```json
{
  "status": "failed",
  "message": "License not found",
  "errors": "record not found"
}
```

**Internal Server Error (500)**
```json
{
  "status": "failed",
  "message": "Failed to retrieve licenses",
  "errors": "database connection error"
}
```

### Validation Rules

1. **UUID Parameters**: All ID parameters must be valid UUIDs
2. **Required Fields**: All required fields in request bodies must be provided
3. **Role Validation**: Role IDs must match predefined values
4. **License Validation**: Serial numbers must reference existing licenses
5. **Authentication**: All endpoints require valid JWT authentication

---

## Testing

### Manual Testing
All endpoints have been manually tested to verify:
- Proper response format compliance
- Error handling for invalid inputs
- Domain-based authorization functionality
- Transaction integrity for operations involving multiple tables

### Example Test Scenarios

1. **License Creation**: Verify license and license log are created atomically
2. **Role Filtering**: Confirm customers and users are properly segregated by role
3. **Domain Authorization**: Test that users cannot access resources outside their domain
4. **Error Responses**: Validate error format and appropriate HTTP status codes

---

## Implementation Notes

### Architecture
- **Clean Architecture**: Follows domain, application, infrastructure, and interface layers
- **Repository Pattern**: Consistent data access with interface abstractions
- **Service Layer**: Business logic encapsulation with transaction management
- **Standardized Responses**: All endpoints use the centralized response package

### Security Features
- **JWT Authentication**: Token-based authentication with domain claims
- **PIN Hashing**: Secure bcrypt hashing for user PINs/passwords
- **Transaction Safety**: Database transactions ensure data consistency
- **Role-Based Access**: Strict role validation prevents unauthorized operations

### Performance Considerations
- **Pagination**: All list endpoints support limit/offset pagination
- **Efficient Queries**: Repository layer optimizes database queries
- **Connection Pooling**: Database connection management for scalability