# Ecosystem Engine

A robust, high-performance backend service engine built with Go, designed for managing complex ecosystem services with scalability and security at its core.

## 🚀 Overview

Ecosystem Engine is a production-ready RESTful API developed using the Go programming language and the Gin web framework. It strictly adheres to **Clean Architecture** principles, ensuring a modular codebase that is easy to maintain, test, and extend. The project leverages Redis for high-speed caching and real-time features, alongside PostgreSQL for persistent data storage.

## ✨ Key Features

- **🔐 Security & Session Management**: Robust authentication system using Redis-backed sessions for secure, stateless-feeling but server-controlled access.
- **🚦 Traffic Control**: Advanced rate limiting implemented with the Redis Fixed Window algorithm to prevent API abuse and ensure fair usage.
- **📍 Discovery & Geospatial Indexing**: High-performance "nearby" search functionality utilizing Redis GEO commands and PostgreSQL geospatial data.
- **🏆 Real-time Leaderboards**: Dynamic ranking system using Redis Sorted Sets to track and display user points and standings in real-time.
- **🛠 Service Management**: Comprehensive CRUD operations for managing ecosystem partners and services.
- **📈 Infrastructure**: Standardized error handling, request binding, and logging for consistent API responses.

## 🛠 Tech Stack

- **Language**: [Go](https://golang.org/) (1.26.2)
- **Web Framework**: [Gin Gonic](https://gin-gonic.com/)
- **Database**: [PostgreSQL](https://www.postgresql.org/) with [pgx](https://github.com/jackc/pgx)
- **Caching & Real-time**: [Redis](https://redis.io/) with [go-redis](https://github.com/redis/go-redis)
- **Migrations**: [Goose](https://github.com/pressly/goose)
- **Live Reload**: [Air](https://github.com/cosmtrek/air)
- **Validation**: [Validator v10](https://github.com/go-playground/validator)
- **Environment Management**: [Godotenv](https://github.com/joho/godotenv)

## 🏗 Project Structure

The project follows a feature-based Clean Architecture structure:

```text
.
├── cmd/                # Application entry points
│   └── api/            # Main API server
├── internal/           # Private application logic
│   ├── apperror/       # Standardized error definitions
│   ├── features/       # Business domains (Feature Slices)
│   │   ├── auth/       # Authentication & Session logic
│   │   ├── discovery/  # Geospatial indexing & location search
│   │   ├── leaderboard/# Ranking & point system
│   │   └── services/   # Partner/Service management
│   └── platform/       # Infrastructure & Shared Libraries
│       ├── cache/      # Redis client & utilities
│       ├── config/     # Environment configuration
│       ├── database/   # PostgreSQL connection (pgxpool)
│       └── http/       # Server setup & custom middleware
├── migrations/         # SQL database migrations (Goose)
├── postman/            # API Documentation & Collections
├── scripts/            # Automation and helper scripts
└── Makefile            # Task runner for development
```

## 🏁 Getting Started

### Prerequisites

- Go 1.26.2+
- Docker & Docker Compose (optional, for infrastructure)
- PostgreSQL 15+
- Redis 7+

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/taqiyyaghazi/ecosystem-engine.git
   cd ecosystem-engine
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Setup environment**:
   ```bash
   cp .env.example .env
   # Edit .env to match your local/docker setup
   ```

4. **Spin up Infrastructure (Docker)**:
   ```bash
   docker-compose up -d
   ```

5. **Run Migrations**:
   ```bash
   make migrate-up
   ```

### Running the Application

- **Development Mode** (with Air hot-reload):
  ```bash
  make dev
  ```
- **Production Mode**:
  ```bash
  make run
  ```

## ⚙️ Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Port for the API server | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/db` |
| `REDIS_URL` | Redis connection address | `localhost:6379` |
| `REDIS_PASSWORD`| Redis password (if any) | `""` |
| `REDIS_DB` | Redis database index | `0` |

## 🧪 Testing & Quality

- **Run all tests**: `make test`
- **Lint check**: `make lint` (requires golangci-lint)

## 📖 API Documentation

A Postman collection is available in the `postman/` directory. 
1. Import `EcoSystemEngine.postman_collection.json` into Postman.
2. Import `EcoSystemEngine.postman_environment.json` for environment variables.

## 📜 Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the API binary |
| `make run` | Build and run the application |
| `make dev` | Run with live reloading (Air) |
| `make test` | Execute unit and integration tests |
| `make migrate-up` | Apply all database migrations |
| `make migrate-down` | Rollback the last migration |
| `make migrate-status`| Show current migration status |
| `make migrate-create`| Create a new migration file |

## 🤝 Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feat/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add some AmazingFeature'`)
4. Push to the Branch (`git push origin feat/AmazingFeature`)
5. Open a Pull Request
