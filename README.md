# Ecosystem Engine

A robust backend service engine built with Go, designed for managing ecosystem services with performance and scalability in mind.

## 🚀 Overview

Ecosystem Engine is a RESTful API built using the Go programming language and the Gin web framework. It follows Clean Architecture principles to ensure maintainability, testability, and clear separation of concerns.

## 🛠 Tech Stack

- **Language**: [Go](https://golang.org/) (1.26.2+)
- **Web Framework**: [Gin Gonic](https://gin-gonic.com/)
- **Database**: [PostgreSQL](https://www.postgresql.org/) with [pgx](https://github.com/jackc/pgx)
- **Caching**: [Redis](https://redis.io/) with [go-redis](https://github.com/redis/go-redis)
- **Migrations**: [Goose](https://github.com/pressly/goose)
- **Live Reload**: [Air](https://github.com/cosmtrek/air)
- **Validation**: [Validator v10](https://github.com/go-playground/validator)

## 🏗 Project Structure

```text
.
├── cmd/                # Application entry points
│   └── api/            # Main API server
├── internal/           # Private application and library code
│   ├── apperror/       # Custom error handling
│   ├── features/       # Business logic (Domain driven)
│   │   └── services/   # Service management feature
│   │       ├── delivery/   # HTTP handlers
│   │       ├── dto/        # Data Transfer Objects
│   │       ├── entity/     # Domain models
│   │       ├── repository/ # Database operations
│   │       └── usecase/    # Business rules
│   └── platform/       # Infrastructure and shared libraries
│       ├── cache/      # Redis integration
│       ├── config/     # Configuration management
│       ├── database/   # Database connection handling
│       └── http/       # HTTP server setup
├── migrations/         # SQL migration files
├── scripts/            # Helper scripts
└── Makefile            # Task automation
```

## 🏁 Getting Started

### Prerequisites

- Go 1.26.2 or later
- PostgreSQL
- Redis
- [Goose](https://github.com/pressly/goose) (for migrations)
- [Air](https://github.com/cosmtrek/air) (optional, for development)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/taqiyyaghazi/ecosystem-engine.git
   cd ecosystem-engine
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Setup environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

### Database Migrations

Run migrations to setup your database schema:

```bash
make migrate-up
```

## 🏃 Running the Application

### Development Mode (with Hot Reload)

```bash
make dev
```

### Production Mode (Build and Run)

```bash
make run
```

## 🧪 Testing & Linting

- **Run tests**: `make test`
- **Run linter**: `make lint`

## 📜 Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the API binary |
| `make run` | Build and run the binary |
| `make dev` | Run with live reloading using Air |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make migrate-up` | Run database migrations up |
| `make migrate-down` | Rollback migrations |
| `make migrate-status` | Check migration status |
| `make migrate-create` | Create a new migration file |

## 🤝 Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feat/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat(scope): add some AmazingFeature'`)
4. Push to the Branch (`git push origin feat/AmazingFeature`)
5. Open a Pull Request
