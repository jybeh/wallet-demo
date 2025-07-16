
# Wallet Service

A Go-based wallet service that provides account management, transaction processing, and peer-to-peer transfers.

## Features

- Account management with balance tracking
- Transaction history and querying
- Peer-to-peer transfers
- Deposit and withdrawal operations
- RESTful API with JSON responses
- PostgreSQL database integration
- Comprehensive test coverage

**Note**: This demo focuses on wallet operations. User management and user-account associations are not implemented in this version.

## Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM ORM
- **Validation**: go-playground/validator
- **Testing**: testify
- **Mocking**: mockery

## Project Structure

The application follows a 3-layer architecture pattern:

```
wallet/
├── cmd/wallet/          # Application entry point
├── server/              # Server setup and configuration
├── handler/             # HTTP handlers and routing (Route Layer)
├── logic/transfer/      # Business logic for transfers (Logic Layer)
├── storage/             # Data access layer (Storage Layer)
├── dto/                 # Data transfer objects
├── util/                # Utility functions
├── db/                  # Database initialization scripts
└── mocks/               # Generated mocks for testing
```

### Architecture Design

- **Route Layer** (`handler/`): Handles HTTP requests, validation, and response formatting. Can customize and orchestrate multiple logic layers as needed
- **Logic Layer** (`logic/`): Contains business rules and transaction processing logic. Optional layer that can be bypassed for simple CRUD operations
- **Storage Layer** (`storage/`): Defines database objects, schemas, and DAO operations. Contains no business logic, only data access methods

## Key Design Decisions

### Framework Choices
- **Gin Framework**: Chosen for its lightweight nature and excellent performance, making it ideal for this assessment
- **GORM ORM**: Widely used Go ORM with built-in security features like SQL injection prevention and convenient transaction handling

### Design Patterns
- **Singleton Pattern**: Used for database connection and service initialization
- **Repository Pattern**: Implemented in storage layer for clean data access abstraction
- **Dependency Injection**: Used throughout for better testability and modularity

## How to Review This Code

### Recommended Review Order
1. **Project Structure**: Start with the 3-layer architecture overview
2. **API Endpoints**: Review each endpoint and its usage (detailed in API section below)
3. **Key Implementation**: Focus on `logic/transfer/transfer.go` - demonstrates core transaction logic
4. **Database Design**: Check `db/init.sql` for schema structure
5. **Testing**: Review test files to understand coverage and mocking approach

### Key File: `logic/transfer/transfer.go`
This file demonstrates the core transaction processing logic:
- Each transfer creates **1 transfer record** and **2 transaction entries** (one credit, one debit)
- Balance calculations are performed atomically
- All operations use **GORM transactions** to ensure data consistency
- Implements retry logic for handling concurrent operations

### API Usage
For exact request/response body formats, refer to the Postman collection at `postman/Wallet Demo.postman_collection.json`.

## API Endpoints

### Accounts
- `POST /v1/accounts/query` - Get account details and current balance
- `POST /v1/accounts/transactions/query` - Get paginated account transaction history
- `POST /v1/accounts/deposits` - Create deposit (from holding account to user account)
- `POST /v1/accounts/withdrawals` - Create withdrawal (from user account to holding account)

### Transfers
- `POST /v1/payment/transfers` - Create peer-to-peer transfer between user accounts

## Getting Started

### Prerequisites
- Go 1.24+
- PostgreSQL
- Docker (optional)

### Database Setup

1. Create a PostgreSQL database named `wallet`
2. Run the initialization script:
```bash
psql -d wallet -f db/init.sql
```

### Installation

1. Clone the repository
2. Install dependencies:
```bash
go mod download
```

3. Update database connection in `server/serve.go` if needed:
```go
dsn := "host=localhost dbname=wallet port=5432 sslmode=disable TimeZone=Asia/Kuala_Lumpur"
```

### Running the Application

```bash
go run cmd/wallet/main.go
```

The server will start on the default port (8080).

## Quick Demo

This is a demo wallet application. Follow these steps to try it out:

### 1. Database Setup

Create a PostgreSQL database and run the migration scripts:

```bash
# Create database
createdb wallet

# Run schema migration
psql -d wallet -f db/init.sql

# Load demo data (optional)
psql -d wallet -f db/seed.sql
```

The seed data creates:
- Holding Account (`1000000001`) with RM 1,000,000,000.00
- Demo Wallet 1 (`12345678`) with RM 1,000.00
- Demo Wallet 2 (`87654321`) with RM 1,000.00

### 2. Start the Application

```bash
go run cmd/wallet/main.go
```

Server starts on `http://localhost:8080`

### 3. Try the APIs

Import the Postman collection from `postman/Wallet Demo.postman_collection.json` and fire the requests to explore the wallet functionality.

### 4. Demo Scenarios

Try these scenarios to explore the wallet functionality:

1. **Transfer between accounts**: Move money from Demo Wallet 1 to Demo Wallet 2
2. **Withdrawal**: Withdraw money from a wallet (goes to holding account)
3. **Deposit**: Add money to a wallet (comes from holding account)
4. **Check balances**: Verify account balances after transactions
5. **Transaction history**: View all transactions for an account

**Note**: All amounts are in minor units (e.g., 100 = RM 1.00 for MYR)

## Testing

Run all tests:
```bash
go test ./...
```

Generate test coverage:
```bash
go test -cover ./...
```

### Generating Mocks

This project uses mockery for generating mocks. To regenerate mocks:

```bash
mockery
```

Configuration is in `.mockery.yml`.

## Features Not Implemented

The following features were intentionally excluded from this submission:

- **User Management**: Authentication and authorization systems (focus was on transaction processing)
- **User-Account Association**: User registration and account linking
- **Real-time Notifications**: Push notifications for transactions
- **Caching Layer**: Requires careful cache invalidation handling, considered as enhancement
- **Rate Limiting**: API throttling mechanisms
- **Advanced Error Handling**: Comprehensive error categorization and recovery

## Areas for Improvement

### Architecture & Scalability
- **Microservices Migration**: Current monolithic architecture can be decoupled into microservices for better scalability and maintainability:
    - **Account Service**: Handle account management and balance operations
    - **Transaction Service**: Process and store transaction records
    - **Transfer Service**: Manage peer-to-peer transfers and business logic
    - **Notification Service**: Handle transfer notifications and webhooks
    - **API Gateway**: Route requests and handle cross-cutting concerns

### Configuration Management
- **Move to configuration files**: Replace hardcoded values with a proper configuration system using [Viper](https://github.com/spf13/viper)
    - Database connection strings
    - Server port and host settings
    - Environment-specific configurations
    - Logging levels and output formats

### Security Enhancements
- Implement API rate limiting
- Secure database connection with proper SSL/TLS

### Observability
- Add structured logging with correlation IDs

### Performance & Scalability
- Implement database connection pooling optimization
- Add caching layer for frequently accessed data
- Consider implementing async processing for transfers
- Add database read replicas support

### Error Handling
- Standardize error response format across all endpoints
- Add proper error codes and categorization
- Implement circuit breaker pattern for external dependencies

### Testing
- Increase integration test coverage
- Add end-to-end API tests
- Performance and load testing
- Database migration testing

## Future Enhancements (Not included)

If additional time were available, the following improvements would be prioritized:

1. **End-to-End Testing**: Comprehensive API testing with real database interactions
2. **User-Account Association**: Complete user management system with proper account linking
3. **Queue Implementation**: Message queue system (Redis/RabbitMQ) for handling higher transaction loads
4. **Caching Layer**: Redis-based caching for frequently accessed account and transaction data
5. **Advanced Monitoring**: Metrics collection and alerting systems

## Development

### Adding New Features
1. Define interfaces in the appropriate package
2. Implement business logic in `logic/` packages
3. Add data access methods in `storage/`
4. Create HTTP handlers in `handler/`
5. Add tests with mocks
6. Update API documentation

### Database Migrations
Database schema changes should be added to `db/init.sql` or separate migration files.