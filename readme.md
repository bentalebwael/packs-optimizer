# Pack Optimization

## Project Overview

This application solves the problem of determining the optimal combination of product packs to fulfill customer orders. The key business rules are:

1. **Only whole packs can be sent.** Packs cannot be broken open.
2. **Send the minimum number of items** to fulfill the order (while respecting rule #1).
3. **Use as few packs as possible** to fulfill each order (while respecting rules #1 and #2).

Note that rule #2 takes precedence over rule #3, meaning we prioritize minimizing total items over minimizing the number of packs.

## Example Orders and Solutions

### Example of available pack sizes:
- 250 Items
- 500 Items
- 1000 Items
- 2000 Items
- 5000 Items

| Items ordered | Correct number of packs | Explanation |
|---------------|-------------------------|-------------|
| 1             | 1 x 250                 | Smallest available pack |
| 250           | 1 x 250                 | Exact match with available pack |
| 251           | 1 x 500                 | Next available pack size that can fulfill the order |
| 501           | 1 x 500 + 1 x 250       | Combination that minimizes total items |
| 12001         | 2 x 5000 + 1 x 2000 + 1 x 250 | Optimal combination for large order |

### Examples of incorrect Pack Combinations

| Items ordered | Incorrect solutions | Reason |
|---------------|---------------------|--------|
| 1             | 1 x 500             | More items than necessary |
| 250           | 1 x 500             | More items than necessary |
| 251           | 2 x 250             | More packs than necessary |
| 501           | 1 x 1000            | More items than necessary |
| 501           | 3 x 250             | More packs than necessary |
| 12001         | 3 x 5000            | More items than necessary |

## Features

- RESTful API for calculating optimal pack combinations
- Dynamic pack size configuration with database persistence
- Interactive Swagger API documentation
- Request/Response logging with structured logging (zap)
- Rate limiting protection
- Database migrations for schema management
- Simple UI for interacting with the API
- Comprehensive test coverage
- Clean architecture with domain-driven design

## Technical Stack

- Backend: Go (Golang)
- Database: PostgreSQL
- Frontend: HTML, CSS, JavaScript
- Documentation: OpenAPI/Swagger
- Logging: Uber's zap logger

## Project Structure

```
pack_calculator/
├── api/                    # API layer
│   ├── middleware/        # Request middleware (logging, rate limiting)
│   ├── router.go         # Route definitions
│   └── swagger.yaml      # API documentation
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── config/               # Configuration management
│   └── config.go
├── internal/             # Internal packages
│   ├── order_calculations/    # Order calculation domain
│   │   ├── entity.go
│   │   ├── handler.go
│   │   ├── repository.go
│   │   └── service.go
│   └── pack_configurations/   # Pack configuration domain
│       ├── entity.go
│       ├── handler.go
│       ├── repository.go
│       └── service.go
├── migrations/           # Database migrations
├── static/              # Frontend assets
│   ├── index.html
│   ├── style.css
│   └── app.js
└── README.md
```

## API Documentation

API documentation is available through Swagger UI at `/swagger/index.html` when running the application. The following endpoints are available:

### Endpoints

- `GET /api/packs`: Get active pack configuration
- `POST /api/packs`: Update pack sizes configuration
- `POST /api/calculate`: Calculate optimal packs for an order

For detailed request/response schemas and examples, refer to the Swagger documentation.

## Configuration

The application uses environment variables for configuration:

```bash
# Database
DATABASE_URL=postgres://username:password@localhost:5432/dbname?sslmode=disable

# Server
PORT=8080

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
```

## Setup and Running

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/pack-calculator.git
   cd pack-calculator
   ```

2. Run docker compose:
   ```bash
   docker-compose up -d
   ```

3. Access:
   - Web UI: http://localhost:8080
   - API Documentation: http://localhost:8080/swagger/index.html

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```
