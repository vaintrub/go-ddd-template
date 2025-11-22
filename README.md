# Go DDD Template

A production-ready Domain-Driven Design (DDD) template demonstrating **Hexagonal Architecture** (Ports & Adapters) with **Bounded Contexts**, **CQRS**, and **Event-Driven Design** in Go.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Bounded Contexts](#bounded-contexts)
  - [Hexagonal Architecture](#hexagonal-architecture)
  - [CQRS Pattern](#cqrs-pattern)
  - [Error Handling](#error-handling)
- [Project Structure](#project-structure)
- [Conventions](#conventions)
- [Quick Start](#quick-start)
- [Development](#development)
- [Technology Stack](#technology-stack)

## Overview

This project implements a **modular monolith** using DDD tactical patterns with three bounded contexts:

- **Trainer**: Manages trainer availability (hours scheduling)
- **Trainings**: Handles training sessions lifecycle (schedule, cancel, reschedule)
- **Users**: User account and authentication management

Each context is completely isolated with explicit contracts through gRPC and domain events.

## Architecture

### Bounded Contexts

Each bounded context is a **separate module** with its own:
- **Domain models** - Business entities and rules
- **Application layer** - Use cases (commands & queries)
- **Ports** - Interfaces for external communication
- **Adapters** - Implementations (HTTP, gRPC, PostgreSQL)

**Inter-context communication** happens only through:
1. **gRPC** for synchronous calls (e.g., Trainings → Users)
2. **Domain events** for asynchronous notifications (future)

```
┌─────────────────────────────────────────────────────────┐
│                    Modular Monolith                     │
├─────────────────┬─────────────────┬─────────────────────┤
│  Trainer        │  Trainings      │  Users              │
│  Context        │  Context        │  Context            │
│                 │                 │                     │
│  ├─ domain      │  ├─ domain      │  ├─ domain         │
│  ├─ app         │  ├─ app         │  ├─ app            │
│  ├─ ports       │  ├─ ports       │  ├─ ports          │
│  └─ adapters    │  └─ adapters    │  └─ adapters       │
└─────────────────┴─────────────────┴─────────────────────┘
         │                 │                 │
         └────────gRPC─────┴────────gRPC─────┘
```

### Hexagonal Architecture

Each bounded context follows **Hexagonal Architecture** (Ports & Adapters):

```
┌─────────────────────────────────────────────────────────┐
│                    Bounded Context                      │
│                                                         │
│   ┌─────────────────────────────────────────────┐     │
│   │            Domain Layer (Core)              │     │
│   │  • Entities, Value Objects, Aggregates      │     │
│   │  • Business rules and invariants            │     │
│   │  • Domain events                            │     │
│   │  • Pure Go errors (no dependencies)         │     │
│   └─────────────────────────────────────────────┘     │
│                        ▲                               │
│   ┌────────────────────┼────────────────────────┐     │
│   │       Application Layer (Use Cases)         │     │
│   │  • Commands (write operations)              │     │
│   │  • Queries (read operations)                │     │
│   │  • Wraps domain errors in SlugError         │     │
│   │  • Transaction orchestration                │     │
│   └─────────────────────────────────────────────┘     │
│            ▲                            ▲              │
│   ┌────────┴────────┐         ┌────────┴────────┐     │
│   │     Ports       │         │     Ports       │     │
│   │  (Interfaces)   │         │  (Interfaces)   │     │
│   └────────┬────────┘         └────────┬────────┘     │
│            ▼                            ▼              │
│   ┌────────────────┐          ┌────────────────┐      │
│   │   Adapters     │          │   Adapters     │      │
│   │  • HTTP/gRPC   │          │  • PostgreSQL  │      │
│   │  • Validation  │          │  • Repository  │      │
│   └────────────────┘          └────────────────┘      │
└─────────────────────────────────────────────────────────┘
```

**Key principles:**

1. **Domain layer** has NO dependencies on infrastructure
2. **Application layer** depends only on domain + port interfaces
3. **Adapters** implement ports and depend on external libraries
4. Dependencies point **inward** (Dependency Inversion Principle)

### Layer Responsibilities

#### 1. Domain Layer (`domain/`)

**Purpose**: Core business logic and rules

**Contents**:
- **Entities**: Objects with identity (e.g., `Hour`, `Training`)
- **Value Objects**: Immutable objects without identity (e.g., `Availability`, `UserType`)
- **Aggregates**: Consistency boundaries (e.g., `Training` aggregate)
- **Domain Services**: Business logic that doesn't belong to entities
- **Domain Events**: Things that happened (future)
- **Factories**: Complex object creation with validation

**Rules**:
- ✅ Pure Go with standard library only
- ✅ Returns standard Go errors (`errors.New`, `fmt.Errorf`)
- ✅ Encapsulates business invariants
- ❌ NO database, HTTP, or external dependencies
- ❌ NO error conversion (app layer handles that)

**Example** (`internal/trainer/domain/hour/hour.go`):
```go
// Factory creates domain objects with validation
type Factory struct {
    maxWeeksInFuture int
    minUtcHour       int
    maxUtcHour       int
}

// Returns standard Go error
func (f Factory) NewAvailableHour(time time.Time) (*Hour, error) {
    if err := f.validateTime(time); err != nil {
        return nil, err // Domain error
    }
    return &Hour{time: time, availability: Available}, nil
}
```

#### 2. Application Layer (`app/`)

**Purpose**: Use cases and orchestration

**Structure**:
```
app/
├── command/              # Write operations (mutations)
│   ├── schedule_training.go
│   ├── cancel_training.go
│   └── ...
└── query/               # Read operations (projections)
    ├── available_hours.go
    └── ...
```

**CQRS Pattern**:
- **Commands** modify state, return errors
- **Queries** read data, return DTOs

**Responsibilities**:
- Execute use cases
- Wrap domain errors in `SlugError` (see Error Handling)
- Orchestrate transactions
- Call external services (via ports)

**Example** (`internal/trainings/app/command/schedule_training.go`):
```go
type ScheduleTrainingHandler struct {
    repo         Repository
    userService  UserService  // Port interface
    trainerService TrainerService
}

func (h ScheduleTrainingHandler) Handle(ctx context.Context, cmd ScheduleTraining) error {
    // Domain logic returns Go error
    tr, err := training.NewTraining(cmd.UUID, cmd.UserUUID, cmd.UserName, cmd.Time, cmd.Notes)
    if err != nil {
        // Wrap domain error
        return errors.NewIncorrectInputError(err.Error(), "invalid-training")
    }

    // Infrastructure error
    if err := h.repo.AddTraining(ctx, tr); err != nil {
        // Wrap infrastructure error
        return errors.NewSlugError(err.Error(), "repository-error")
    }

    return nil
}
```

#### 3. Ports Layer (`ports/`)

**Purpose**: Interface definitions for adapters

**Types**:
- **Primary Ports** (driving): HTTP/gRPC handlers
- **Secondary Ports** (driven): Repository, external services

**Example**:
```go
// ports/repositories.go
type HourRepository interface {
    GetHour(ctx context.Context, time time.Time) (*hour.Hour, error)
    UpdateHour(ctx context.Context, time time.Time, updateFn func(*hour.Hour) (*hour.Hour, error)) error
}

// ports/user_service.go
type UserService interface {
    UpdateTrainingBalance(ctx context.Context, userUUID string, amountChange int) error
}
```

#### 4. Adapters Layer (`adapters/`)

**Purpose**: Implementations of ports

**Types**:
- **HTTP adapter**: REST API (chi router, OpenAPI generated)
- **gRPC adapter**: gRPC service implementations
- **PostgreSQL adapter**: Repository implementations (SQLC generated)

**Example** (`internal/trainer/adapters/hour_postgres_repository.go`):
```go
type HourPostgresRepository struct {
    pool    *pgxpool.Pool
    factory hour.Factory
}

// Implements HourRepository port
func (r *HourPostgresRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(*hour.Hour) (*hour.Hour, error),
) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    queries := sqlc_trainer.New(tx)

    // SELECT FOR UPDATE ensures transaction isolation
    dbHour, err := queries.GetHourByTime(ctx, hourTime)
    if err != nil {
        return db.TranslatePgError(err)
    }

    // Convert to domain object
    availability, _ := hour.NewAvailabilityFromString(dbHour.Availability)
    domainHour, _ := r.factory.UnmarshalHourFromDatabase(dbHour.HourTime, availability)

    // Apply domain logic
    updatedHour, err := updateFn(domainHour)
    if err != nil {
        return err // Domain error propagates
    }

    // Persist changes
    err = queries.UpdateHourAvailability(ctx, dbHour.ID, updatedHour.Availability().String())
    if err != nil {
        return db.TranslatePgError(err)
    }

    return tx.Commit(ctx)
}
```

### CQRS Pattern

**Command Query Responsibility Segregation** separates reads from writes:

```
Commands (Write)              Queries (Read)
     │                             │
     ▼                             ▼
┌─────────────┐            ┌──────────────┐
│  Command    │            │   Query      │
│  Handler    │            │   Handler    │
└──────┬──────┘            └──────┬───────┘
       │                          │
       ▼                          ▼
┌─────────────┐            ┌──────────────┐
│  Domain     │            │  Read Model  │
│  Repository │            │  (Optimized) │
└─────────────┘            └──────────────┘
```

**Why CQRS?**
- **Commands** enforce business rules (domain layer)
- **Queries** optimize for reads (denormalized views)
- **Scales independently** (different databases for read/write)
- **Clearer intent** in code

**Example**:
```go
// Command - modifies state
type ScheduleTraining struct {
    UUID      string
    UserUUID  string
    Time      time.Time
    Notes     string
}

// Query - returns DTO
type AvailableHoursHandler struct {
    readModel AvailableHoursReadModel
}

func (h AvailableHoursHandler) Handle(ctx context.Context, from, to time.Time) ([]Date, error) {
    // Optimized read, bypasses domain layer
    return h.readModel.AvailableHours(ctx, from, to)
}
```

### Error Handling

The project uses a **layered error handling** strategy:

```
┌─────────────────────────────────────────────────────────┐
│  HTTP/gRPC Layer                                        │
│  Maps SlugError.ErrorType → HTTP Status                │
│  • Authorization      → 401 Unauthorized                │
│  • IncorrectInput    → 400 Bad Request                  │
│  • Unknown           → 500 Internal Server Error        │
└─────────────────────────────────────────────────────────┘
                          ▲
┌─────────────────────────┴───────────────────────────────┐
│  Application Layer                                      │
│  Wraps errors in SlugError with ErrorType               │
│  • Domain errors     → IncorrectInput                   │
│  • Infrastructure    → Unknown                          │
│  • Authorization     → Authorization                    │
└─────────────────────────────────────────────────────────┘
                          ▲
┌─────────────────────────┴───────────────────────────────┐
│  Domain Layer                                           │
│  Returns standard Go errors                             │
│  • errors.New("past hour")                             │
│  • fmt.Errorf("invalid: %w", err)                      │
└─────────────────────────────────────────────────────────┘
```

**SlugError** (`internal/common/errors/errors.go`):

```go
type SlugError struct {
    error     string      // Human-readable message
    slug      string      // Machine-readable identifier (e.g., "training-not-found")
    errorType ErrorType   // Maps to HTTP status
}

// Factory functions
func NewIncorrectInputError(error, slug string) SlugError
func NewAuthorizationError(error, slug string) SlugError
func NewSlugError(error, slug string) SlugError  // Unknown type
```

**Usage pattern**:

```go
// Domain layer - standard Go error
func (t *Training) Cancel() error {
    if t.IsCanceled() {
        return errors.New("training is already canceled")
    }
    t.canceled = true
    return nil
}

// Application layer - wrap in SlugError
func (h CancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
    tr, err := h.repo.GetTraining(ctx, cmd.TrainingUUID)
    if err != nil {
        return errors.NewSlugError(err.Error(), "training-not-found")
    }

    if err := tr.Cancel(); err != nil {
        // Domain error → IncorrectInput
        return errors.NewIncorrectInputError(err.Error(), "cancel-failed")
    }

    if err := h.repo.UpdateTraining(ctx, tr); err != nil {
        // Infrastructure error → Unknown
        return errors.NewSlugError(err.Error(), "repository-error")
    }

    return nil
}

// HTTP adapter - map to status code
func (h HttpServer) CancelTraining(w http.ResponseWriter, r *http.Request) {
    err := h.app.Commands.CancelTraining.Handle(r.Context(), cmd)
    if err != nil {
        slugError := err.(errors.SlugError)
        switch slugError.ErrorType() {
        case errors.ErrorTypeIncorrectInput:
            httperr.BadRequest("cancel-failed", err, w, r)
        case errors.ErrorTypeAuthorization:
            httperr.Unauthorized("unauthorized", err, w, r)
        default:
            httperr.InternalError("server-error", err, w, r)
        }
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

**Benefits**:
- Domain remains **pure** (no dependencies)
- **Consistent** error responses across contexts
- **Type-safe** error categorization
- **Machine-readable** slugs for client handling

## Project Structure

```
.
├── api/                           # API contracts
│   ├── openapi/                   # OpenAPI/Swagger specs
│   └── protobuf/                  # gRPC protobuf definitions
│
├── internal/                      # Private application code
│   ├── common/                    # Shared utilities
│   │   ├── auth/                  # JWT authentication
│   │   ├── client/                # gRPC client helpers
│   │   ├── config/                # Centralized configuration
│   │   ├── context/               # Context utilities
│   │   ├── db/                    # Database utilities (connection, errors)
│   │   ├── decorator/             # Logging/metrics decorators
│   │   ├── errors/                # SlugError definitions
│   │   ├── logs/                  # Structured logging (slog)
│   │   ├── server/                # HTTP/gRPC server setup
│   │   └── tests/                 # Test helpers
│   │
│   ├── trainer/                   # Trainer bounded context
│   │   ├── domain/                # Business logic
│   │   │   └── hour/              # Hour aggregate
│   │   │       ├── hour.go        # Entity
│   │   │       ├── availability.go # Value object
│   │   │       └── factory.go     # Factory with validation
│   │   ├── app/                   # Application services
│   │   │   ├── command/           # Write operations
│   │   │   └── query/             # Read operations
│   │   ├── ports/                 # Interface definitions
│   │   ├── adapters/              # Implementations
│   │   │   ├── hour_postgres_repository.go
│   │   │   └── sqlc/              # Generated SQLC code
│   │   └── service/               # Service composition
│   │       └── application.go     # Wires everything together
│   │
│   ├── trainings/                 # Trainings bounded context
│   │   ├── domain/
│   │   │   └── training/          # Training aggregate
│   │   ├── app/
│   │   │   ├── command/
│   │   │   └── query/
│   │   ├── ports/
│   │   ├── adapters/
│   │   └── service/
│   │
│   └── users/                     # Users bounded context
│       ├── adapters/
│       │   └── sqlc/
│       └── ...
│
├── migrations/                    # Database migrations (golang-migrate)
│   ├── 000001_create_trainer_hours.up.sql
│   ├── 000001_create_trainer_hours.down.sql
│   └── ...
│
├── sql/                          # SQLC query definitions
│   └── queries/
│       ├── trainer.sql           # Trainer context queries
│       ├── trainings.sql         # Trainings context queries
│       └── users.sql             # Users context queries
│
├── specs/                        # Feature specifications
│   ├── 001-postgres-sqlc-migration/
│   └── 002-slog-migration/
│
├── scripts/                      # Build/deployment scripts
│
├── Makefile                      # Development automation
├── sqlc.yaml                     # SQLC configuration
├── go.work                       # Go workspace
└── docker-compose.yml            # Local development environment
```

## Conventions

### Naming Conventions

#### Files and Directories
- **Package names**: Singular, lowercase (e.g., `hour`, `training`)
- **Domain files**: `{entity}.go`, `{value_object}.go`, `factory.go`
- **Command handlers**: `{action}_{entity}.go` (e.g., `schedule_training.go`)
- **Repository**: `{entity}_postgres_repository.go`

#### Code
- **Interfaces**: Port interfaces (e.g., `HourRepository`, `UserService`)
- **Implementations**: Concrete types (e.g., `HourPostgresRepository`)
- **Commands**: Imperative (e.g., `ScheduleTraining`, `CancelTraining`)
- **Queries**: Noun phrase (e.g., `AvailableHours`, `TrainingsList`)
- **Domain events**: Past tense (e.g., `TrainingScheduled`, `HourMadeAvailable`)

### Database Conventions

#### Table Naming
- Format: `{context}_{entity_plural}` (e.g., `trainer_hours`, `trainings_trainings`)
- Primary keys: `id` (UUID)
- Foreign keys: `{entity}_id` or `{entity}_uuid`
- Timestamps: `created_at`, `updated_at` (both `TIMESTAMPTZ NOT NULL`)

#### SQLC Queries
- Location: `sql/queries/{context}.sql`
- Comments: `-- name: {FunctionName} :{mode}`
  - Modes: `:one`, `:many`, `:exec`
- Generate: `make sqlc-generate`

**Example** (`sql/queries/trainer.sql`):
```sql
-- name: GetHourByTime :one
SELECT * FROM trainer_hours
WHERE hour_time = $1
FOR UPDATE;  -- Row-level locking for concurrency

-- name: ListHoursByTimeRange :many
SELECT * FROM trainer_hours
WHERE hour_time BETWEEN $1 AND $2
ORDER BY hour_time;
```

### Transaction Patterns

**Use transactions for**:
1. **Multi-step operations** (e.g., schedule training + update balance)
2. **Concurrent updates** (use `SELECT FOR UPDATE` to prevent race conditions)
3. **Aggregate consistency** (all aggregate changes in one transaction)

**Example**:
```go
func (r *HourPostgresRepository) UpdateHour(
    ctx context.Context,
    hourTime time.Time,
    updateFn func(*hour.Hour) (*hour.Hour, error),
) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)  // Safe to call after commit

    queries := sqlc_trainer.New(tx)

    // SELECT FOR UPDATE locks the row
    dbHour, err := queries.GetHourByTime(ctx, hourTime)
    if err != nil {
        return db.TranslatePgError(err)
    }

    // Domain logic
    updatedHour, err := updateFn(domainHour)
    if err != nil {
        return err  // Transaction rolls back
    }

    // Persist
    err = queries.UpdateHourAvailability(ctx, dbHour.ID, updatedHour.Availability().String())
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

### Testing Conventions

#### Test Organization
```
internal/trainer/
├── domain/
│   └── hour/
│       ├── hour.go
│       └── hour_test.go              # Unit tests (domain logic)
├── adapters/
│   ├── hour_postgres_repository.go
│   └── hour_repository_test.go       # Integration tests (database)
└── service/
    └── component_test.go              # Component tests (HTTP/gRPC)
```

#### Test Types
1. **Unit tests**: Domain logic only (fast, no dependencies)
2. **Integration tests**: Repository + real PostgreSQL (testcontainers)
3. **Component tests**: Full HTTP/gRPC stack + database

**Example**:
```go
// Unit test - pure domain logic
func TestNewAvailableHour_past_date(t *testing.T) {
    t.Parallel()
    pastHour := time.Now().Add(-time.Hour)
    _, err := factory.NewAvailableHour(pastHour)
    assert.Equal(t, hour.ErrPastHour, err)
}

// Integration test - database operations
func TestRepository_UpdateHour_parallel(t *testing.T) {
    t.Parallel()
    repo := setupPostgresRepo(t)  // testcontainers

    // 20 concurrent workers try to schedule same hour
    // Only 1 should succeed due to SELECT FOR UPDATE
    var wg sync.WaitGroup
    successChan := make(chan int, 20)

    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(workerNum int) {
            defer wg.Done()
            err := repo.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
                return h.MakeUnavailable()
            })
            if err == nil {
                successChan <- workerNum
            }
        }(i)
    }

    wg.Wait()
    close(successChan)

    succeeded := []int{}
    for workerNum := range successChan {
        succeeded = append(succeeded, workerNum)
    }

    require.Len(t, succeeded, 1, "only one worker should succeed")
}
```

### Configuration Management

**Centralized config** (`internal/common/config/config.go`):

```go
type Config struct {
    Environment string
    Server      ServerConfig
    Logging     LoggingConfig
    Database    DatabaseConfig
    GRPC        GRPCConfig
}

// Loads config with precedence: defaults → .env → env vars
func MustLoad(ctx context.Context) Config {
    // Fails fast if required fields missing
}
```

**Usage in services**:
```go
func main() {
    ctx := context.Background()
    cfg := config.MustLoad(ctx)

    logger := logs.Init(cfg.Logging)
    pool := db.NewPool(ctx, cfg.Database)

    app := service.NewApplication(ctx, cfg)
    server.RunHTTPServer(cfg.Server, logger, app.HTTPHandler())
}
```

## Quick Start

### Prerequisites

- **Go**: 1.21 or higher
- **PostgreSQL**: 14 or higher
- **Docker** (optional, for local PostgreSQL)
- **Make**: For build automation

### Development Tools

```bash
# Install golang-migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golangci-lint
brew install golangci-lint  # macOS
# OR
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### 1. Start PostgreSQL

**Option A: Docker (Recommended)**
```bash
docker run --name postgres-dev \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=go_ddd_template_dev \
  -p 5432:5432 \
  -d postgres:14-alpine
```

**Option B: Local Installation**
```bash
# macOS
brew install postgresql@14
brew services start postgresql@14

# Linux (Ubuntu/Debian)
sudo apt install postgresql-14
sudo systemctl start postgresql
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your settings:
```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/go_ddd_template_dev?sslmode=disable
ENV=development
LOG_LEVEL=INFO
```

### 3. Run Migrations

```bash
# Apply all migrations
make migrate-up

# Check status
make migrate-status
```

### 4. Generate SQLC Code

```bash
make sqlc-generate
```

### 5. Run Tests

```bash
# All tests (unit + integration + component)
make test

# Unit tests only (fast)
go test ./internal/... -short

# With coverage
go test -cover ./internal/...
```

### 6. Build Services

```bash
# Format code
make fmt

# Run linters
make lint

# Build (TODO: add build targets)
go build -o bin/trainer ./cmd/trainer
go build -o bin/trainings ./cmd/trainings
go build -o bin/users ./cmd/users
```

## Development

### Make Commands

```bash
make fmt              # Format code (gofmt, goimports)
make lint             # Run golangci-lint
make test             # Run all tests
make sqlc-generate    # Generate SQLC code
make migrate-up       # Apply migrations
make migrate-down     # Rollback last migration
make migrate-status   # Check migration status
make migrate-create NAME=add_feature  # Create new migration
```

### Adding a New Feature

1. **Create specification** in `specs/XXX-feature-name/`
2. **Write tests first** (TDD)
   - Unit tests for domain logic
   - Integration tests for repository
   - Component tests for HTTP/gRPC
3. **Implement domain layer**
   - Entities, value objects, aggregates
   - Business rules and validation
   - Return standard Go errors
4. **Implement application layer**
   - Command/query handlers
   - Wrap errors in SlugError
   - Orchestrate transactions
5. **Implement adapters**
   - Create SQLC queries in `sql/queries/`
   - Run `make sqlc-generate`
   - Implement repository
6. **Wire in service layer**
   - Update `service/application.go`
   - Add HTTP/gRPC handlers in `ports/`
7. **Update documentation**

### Example: Adding a New Command

```bash
# 1. Create SQL query
cat >> sql/queries/trainings.sql <<EOF
-- name: ApproveReschedule :exec
UPDATE trainings_trainings
SET training_time = proposed_new_time,
    proposed_new_time = NULL,
    move_proposed_by = NULL,
    updated_at = NOW()
WHERE id = $1;
EOF

# 2. Generate SQLC code
make sqlc-generate

# 3. Create command handler
cat > internal/trainings/app/command/approve_reschedule.go <<EOF
package command

import (
    "context"
    "github.com/vaintrub/go-ddd-template/internal/common/errors"
)

type ApproveReschedule struct {
    TrainingUUID string
    UserUUID     string
}

type ApproveRescheduleHandler struct {
    repo Repository
}

func (h ApproveRescheduleHandler) Handle(ctx context.Context, cmd ApproveReschedule) error {
    tr, err := h.repo.GetTraining(ctx, cmd.TrainingUUID)
    if err != nil {
        return errors.NewSlugError(err.Error(), "training-not-found")
    }

    if err := tr.ApproveReschedule(); err != nil {
        return errors.NewIncorrectInputError(err.Error(), "approve-failed")
    }

    if err := h.repo.UpdateTraining(ctx, tr); err != nil {
        return errors.NewSlugError(err.Error(), "repository-error")
    }

    return nil
}
EOF

# 4. Wire in application
# Edit internal/trainings/service/application.go
# Add handler to Commands struct

# 5. Add HTTP endpoint
# Edit internal/trainings/ports/http.go
# Implement OpenAPI handler

# 6. Run tests
make test
```

## Technology Stack

### Core
- **Language**: Go 1.21+
- **Architecture**: Hexagonal (Ports & Adapters)
- **Patterns**: DDD, CQRS, Event-Driven

### Database
- **Database**: PostgreSQL 14+
- **Driver**: pgx/v5 (connection pooling, context support)
- **Query Builder**: SQLC (type-safe SQL)
- **Migrations**: golang-migrate

### API
- **HTTP Router**: chi/v5
- **HTTP API**: OpenAPI 3.0 (generated handlers)
- **RPC**: gRPC + Protocol Buffers

### Observability
- **Logging**: slog (structured logging)
- **Metrics**: Prometheus (future)
- **Tracing**: OpenTelemetry (future)

### Testing
- **Framework**: testify (assertions, mocking)
- **Containers**: testcontainers-go (PostgreSQL)
- **Isolation**: Parallel tests with `t.Parallel()`

### Configuration
- **Format**: .env files + environment variables
- **Loader**: gotenv (precedence: defaults → .env → env vars)
- **Validation**: Fail-fast on missing required fields

## Contributing

1. Create a feature branch from `main`
2. Write tests (TDD approach)
3. Implement feature following conventions
4. Run `make lint` and `make test`
5. Create pull request with specification

## License

See [LICENSE](LICENSE) file.

## Documentation

- [Database Schema](migrations/)
- [API Contracts](api/)
