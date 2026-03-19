# N-API Template - Code Generation Reference

This document serves as a comprehensive reference for generating API code following the N-API Template standards. Use this guide to ensure consistency and adherence to established patterns.

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Main Application Entry Point](#main-application-entry-point)
3. [Bootstrap Configuration](#bootstrap-configuration)
4. [Go Module Setup](#go-module-setup)
5. [Configuration Files](#configuration-files)
6. [Port Layer (Request/Response Interfaces)](#port-layer-requestresponse-interfaces)
7. [Domain Model Pattern](#domain-model-pattern)
8. [Repository Pattern](#repository-pattern)
9. [Handler Pattern](#handler-pattern)
10. [Request DTO Pattern](#request-dto-pattern)
11. [Response DTO Pattern](#response-dto-pattern)
12. [Routing Pattern](#routing-pattern)
13. [Validation Pattern](#validation-pattern)
14. [Database Schema](#database-schema)
15. [Database Query Optimization Patterns](#database-query-optimization-patterns)
16. [pgx.Batch Patterns](#pgxbatch-patterns)
17. [Temporal Workflow Integration](#temporal-workflow-integration)
18. [Workflow State Optimization](#workflow-state-optimization)
19. [Naming Conventions](#naming-conventions)
20. [Error Handling](#error-handling)
21. [Complete Example Workflow](#complete-example-workflow)
22. [Development Workflow](#development-workflow)

---

## Project Structure

When creating a new resource, follow this structure:

```
n-api-template/
├── main.go                        # Application entry point
├── go.mod                         # Go module dependencies
├── go.sum                         # Dependency checksums
├── configs/                       # Configuration files
│   ├── config.yaml                # Base configuration
│   ├── config.dev.yaml            # Development environment
│   ├── config.sit.yaml            # System Integration Test
│   ├── config.staging.yaml        # Staging environment
│   ├── config.training.yaml       # Training environment
│   ├── config.test.yaml           # Test environment
│   └── config.prod.yaml           # Production environment
├── bootstrap/
│   └── bootstrapper.go            # Dependency injection modules
├── core/
│   ├── domain/
│   │   └── {resource}.go          # Domain model
│   └── port/
│       ├── request.go             # Common request structures
│       └── response.go            # Common response structures
├── handler/
│   ├── {resource}.go              # Handler with routes
│   ├── request.go                 # Request DTOs (add new structs here)
│   ├── request_*_validator.go     # Auto-generated validators
│   └── response/
│       └── {resource}.go          # Response DTOs
├── repo/
│   └── postgres/
│       └── {resource}.go          # Repository/data access
├── workflows/                     # Temporal workflows (for long-running processes)
│   ├── {resource}_workflow.go     # Workflow definitions
│   └── activities/
│       └── {resource}_activities.go  # Activity implementations
├── db/
│   └── {resource}.sql             # Database schema
└── docs/                          # Swagger documentation (auto-generated)
```

**Note on `workflows/` Directory**:
- **When to use**: For endpoints requiring long-running processes (>1 minute), human approvals, external waits, or complex multi-step orchestration
- **Structure**:
  - `workflows/{resource}_workflow.go` - Workflow definitions (orchestration logic)
  - `workflows/activities/{resource}_activities.go` - Activity implementations (actual work, calls repositories)
- **Examples**: Death claim processing (with investigation), maturity claims (with approval), premium collection (recurring)
- **See**: Section 17 (Temporal Workflow Integration) and Section 18 (Workflow State Optimization) for patterns

**Directory Usage by Endpoint Type**:

| Endpoint Type | Uses | Directory Structure |
|--------------|------|---------------------|
| Simple CRUD | Handler → Repository | handler/, repo/ only |
| Complex business logic | Handler → Repository | handler/, repo/ only |
| Long-running process | Handler → Workflow → Activities → Repository | handler/, workflows/, workflows/activities/, repo/ |

---

## Main Application Entry Point

**Location**: `main.go`

**Purpose**: Application entry point that initializes and starts the server with all dependencies.

**Pattern**:
```go
package main

import (
    "context"
    "{project}/bootstrap"

    bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

func main() {
    app := bootstrapper.New().Options(
        // Add your FX modules here
        bootstrap.FxHandler,  // Register all handlers
        bootstrap.FxRepo,     // Register all repositories
        // bootstrap.Fxvalidator, // Optional: custom validators
    )
    app.WithContext(context.Background()).Run()
}
```

**Rules**:
- Import your project's bootstrap package
- Import n-api-bootstrapper for application initialization
- Register all FX modules in `.Options()` call
- Pass `context.Background()` to `WithContext()`
- Call `.Run()` to start the application
- The bootstrapper automatically handles:
  - Configuration loading
  - Database connection
  - Server initialization
  - Graceful shutdown
  - Signal handling
  - Dependency injection

**Example**:
```go
package main

import (
    "context"
    "pisapi/bootstrap"

    bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

func main() {
    app := bootstrapper.New().Options(
        bootstrap.FxHandler,
        bootstrap.FxRepo,
    )
    app.WithContext(context.Background()).Run()
}
```

---

## Bootstrap Configuration

**Location**: `bootstrap/bootstrapper.go`

**Purpose**: Defines Uber FX dependency injection modules for automatic wiring of dependencies.

**Complete Pattern**:
```go
package bootstrap

import (
    "go.uber.org/fx"
    serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
    handler "{project}/handler"
    repo "{project}/repo/postgres"
)

// FxRepo module provides all repository implementations
var FxRepo = fx.Module(
    "Repomodule",
    fx.Provide(
        repo.New{Resource1}Repository,
        repo.New{Resource2}Repository,
        // Add more repository constructors here
        // repo.New{Resource3}Repository,
    ),
)

// FxHandler module provides all HTTP handlers
var FxHandler = fx.Module(
    "Handlermodule",
    fx.Provide(
        // Each handler must be annotated to implement serverHandler.Handler interface
        fx.Annotate(
            handler.New{Resource1}Handler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
        fx.Annotate(
            handler.New{Resource2}Handler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
        // Add more handler constructors here
        // fx.Annotate(
        //     handler.New{Resource3}Handler,
        //     fx.As(new(serverHandler.Handler)),
        //     fx.ResultTags(serverHandler.ServerControllersGroupTag),
        // ),
    ),
)

// Optional: Custom validator module (if using custom validators)
// var Fxvalidator = fx.Module(
//     "Validatormodule",
//     fx.Provide(
//         // Add custom validator providers here
//     ),
// )
```

**Rules**:
- Create separate FX modules for different concerns (Repo, Handler, Validator, etc.)
- Module naming convention: `Fx{ModuleName}` (e.g., FxRepo, FxHandler)
- Module string name: `"{ModuleName}module"` (e.g., "Repomodule", "Handlermodule")
- Use `fx.Provide()` to register constructors
- Handlers MUST be wrapped with `fx.Annotate()` with:
  - `fx.As(new(serverHandler.Handler))` - Converts to Handler interface
  - `fx.ResultTags(serverHandler.ServerControllersGroupTag)` - Groups handlers
- Repositories are provided directly without annotation
- Add comments to indicate where new resources should be added
- Dependencies are automatically injected based on constructor parameters
- Order of registration doesn't matter (FX resolves dependency graph)

**Dependency Injection Flow**:
1. Bootstrapper creates database connection (*dblib.DB)
2. Bootstrapper loads configuration (*config.Config)
3. FxRepo provides repositories (injecting db and config)
4. FxHandler provides handlers (injecting repositories)
5. Server automatically discovers and registers all handlers

**Example with Multiple Resources**:
```go
package bootstrap

import (
    handler "pisapi/handler"
    repo "pisapi/repo/postgres"

    serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
    "go.uber.org/fx"
)

var FxRepo = fx.Module(
    "Repomodule",
    fx.Provide(
        repo.NewUserRepository,
        repo.NewProductRepository,
        repo.NewOrderRepository,
    ),
)

var FxHandler = fx.Module(
    "Handlermodule",
    fx.Provide(
        fx.Annotate(
            handler.NewUserHandler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
        fx.Annotate(
            handler.NewProductHandler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
        fx.Annotate(
            handler.NewOrderHandler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
    ),
)
```

**Temporal Workflow Module** (Optional - only if using workflows):

```go
package bootstrap

import (
    "go.uber.org/fx"
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    
    "{project}/workflows"
    "{project}/workflows/activities"
)

// FxTemporal module provides Temporal client and worker
var FxTemporal = fx.Module(
    "Temporalmodule",
    fx.Provide(
        // Provide Temporal client
        func() (client.Client, error) {
            return client.NewClient(client.Options{
                HostPort: "localhost:7233", // Or from config
            })
        },
        
        // Provide activity structs
        activities.NewClaimActivities,
        // Add more activity structs here
        // activities.NewPolicyActivities,
    ),
    
    fx.Invoke(
        // Register workflows and activities with worker
        func(c client.Client, claimActivities *activities.ClaimActivities) error {
            w := worker.New(c, "claim-processing-queue", worker.Options{})
            
            // Register workflows
            w.RegisterWorkflow(workflows.DeathClaimWorkflow)
            w.RegisterWorkflow(workflows.MaturityClaimWorkflow)
            // Add more workflow registrations here
            
            // Register activities
            w.RegisterActivity(claimActivities.ValidatePolicyActivity)
            w.RegisterActivity(claimActivities.RegisterDeathClaimActivity)
            w.RegisterActivity(claimActivities.ProcessPaymentActivity)
            // Add more activity registrations here
            
            return w.Start()
        },
    ),
)
```

**Usage in main.go** (when using Temporal):
```go
func main() {
    app := bootstrapper.New().Options(
        bootstrap.FxHandler,
        bootstrap.FxRepo,
        bootstrap.FxTemporal,  // Add Temporal module
    )
    app.WithContext(context.Background()).Run()
}
```

**Rules for Temporal Module**:
- Only add FxTemporal if endpoints require long-running workflows
- Register all workflows and activities in fx.Invoke
- Activities must be methods on structs (for dependency injection)
- Worker task queue name should match workflow options in handlers
- Multiple workers can be created for different task queues

---

## Go Module Setup

**Location**: `go.mod`

**Purpose**: Defines Go module and manages dependencies.

**Pattern**:
```go
module {project}

go 1.25.0

require (
    github.com/Masterminds/squirrel v1.5.4
    github.com/jackc/pgx/v5 v5.7.6
    gitlab.cept.gov.in/it-2.0-common/api-config v0.0.17
    gitlab.cept.gov.in/it-2.0-common/n-api-db v1.0.32
    gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper v0.0.14
    gitlab.cept.gov.in/it-2.0-common/n-api-log v0.0.1
    gitlab.cept.gov.in/it-2.0-common/n-api-server v0.0.17
    gitlab.cept.gov.in/it-2.0-common/n-api-validation v0.0.3
    go.uber.org/fx v1.24.0
    
    // Temporal dependencies (only if using workflows)
    go.temporal.io/sdk v1.25.1
)
```

**Note**: Only add Temporal dependencies if your project uses long-running workflows. For simple CRUD APIs, omit these dependencies.
```

**Core Dependencies**:
- `github.com/Masterminds/squirrel` - SQL query builder
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `gitlab.cept.gov.in/it-2.0-common/api-config` - Configuration management
- `gitlab.cept.gov.in/it-2.0-common/n-api-db` - Database utilities
- `gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper` - Application bootstrapper
- `gitlab.cept.gov.in/it-2.0-common/n-api-log` - Logging utilities
- `gitlab.cept.gov.in/it-2.0-common/n-api-server` - Server framework
- `gitlab.cept.gov.in/it-2.0-common/n-api-validation` - Validation framework
- `go.uber.org/fx` - Dependency injection framework

**Optional Dependencies** (for long-running workflows):
- `go.temporal.io/sdk` - Temporal workflow orchestration

**Commands**:
```bash
# Initialize new module
go mod init {project}

# Add dependency
go get package@version

# Update all dependencies
go get -u ./...

# Tidy up (remove unused, add missing)
go mod tidy

# Download dependencies
go mod download

# Verify checksums
go mod verify
```

**Rules**:
- Use semantic versioning for your module
- Pin exact versions in production
- Run `go mod tidy` after adding/removing imports
- Commit both go.mod and go.sum
- Use private GitLab registry for internal packages

---

## Configuration Files

**Location**: `configs/config.yaml` (and environment-specific variants)

**Purpose**: Application configuration for different environments.

**Base Configuration Pattern** (`configs/config.yaml`):
```yaml
# Application name
appname: "{project-name}"

# Tracing configuration (OpenTelemetry)
trace:
  enabled: false  # Enable/disable distributed tracing
  processor: 
    type: "otlp-grpc"  # Export format: otlp-grpc or otlp-http
    options:
      host: "localhost:4317"  # OpenTelemetry collector endpoint
  sampler:
    type: always-on  # Sampling strategy: always-on, always-off, parent-based-trace-id-ratio
    options: 
      ratio: 0.1  # Sample 10% of traces (if using ratio sampler)

# Cache configuration (Redis + Local)
cache:
  # Redis settings
  redisserver: "10.20.30.33:6379"
  redispassword: ""
  redisdbindex: 1
  redisexpirationtime: 20m
  
  # Local cache settings
  lccapacity: 10000              # Maximum number of entries
  lcnumshards: 20                # Number of shards for concurrent access
  lcttl: 2m                      # Time to live for cache entries
  lcevictionpercentage: 10       # Percentage to evict when full
  lcminrefreshdelay: 15m         # Minimum delay before refresh
  lcmaxrefreshdelay: 30m         # Maximum delay before refresh
  lcretrybasedelay: 1s           # Base delay for retries
  lcbatchsize: 10                # Batch size for operations
  lcbatchbuffertimeout: 30s      # Batch buffer timeout
  
  # Enable/disable cache layers
  isredisenabled: true
  islocalcacheenabled: false

# Database configuration (PostgreSQL)
db: 
  username: "postgres"
  password: "your-password"
  host: "localhost"
  port: "5432"
  database: "your-database"
  schema: "public"
  
  # Connection pool settings
  maxconns: 10                # Maximum connections
  minconns: 1                 # Minimum connections
  maxconnlifetime: 30         # Max connection lifetime (minutes)
  maxconnidletime: 10         # Max idle time (minutes)
  healthcheckperiod: 5        # Health check interval (minutes)
  
  # Query timeouts
  QueryTimeoutLow: 2s         # Simple queries
  QueryTimeoutMed: 5s         # Complex queries/aggregations

# Application info (for Swagger)
info:
  name: "{project-name}"
  version: "1.0.0"
```

**Environment-Specific Files**:
- `config.dev.yaml` - Development environment
- `config.test.yaml` - Test environment
- `config.sit.yaml` - System Integration Test
- `config.training.yaml` - Training environment
- `config.staging.yaml` - Staging environment
- `config.prod.yaml` - Production environment

**Environment Override Example** (`config.prod.yaml`):
```yaml
# Production overrides (only specify what changes)
db:
  host: "prod-db-server.example.com"
  password: "${DB_PASSWORD}"  # Use environment variable
  maxconns: 50                # Higher for production
  minconns: 10

trace:
  enabled: true  # Enable tracing in production
  sampler:
    type: parent-based-trace-id-ratio
    options:
      ratio: 0.1  # Sample 10% in production

cache:
  redisserver: "prod-redis.example.com:6379"
  redispassword: "${REDIS_PASSWORD}"
  isredisenabled: true
  islocalcacheenabled: true  # Enable both layers in production
```

**Configuration Access in Code**:
```go
// In repository or service
timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()

// Get string value
appName := cfg.GetString("appname")

// Get int value
maxConns := cfg.GetInt("db.maxconns")

// Get bool value
traceEnabled := cfg.GetBool("trace.enabled")
```

**Environment Selection**:
```bash
# Set via environment variable
export ENV=prod
go run main.go

# Or via command line flag
go run main.go -env=prod
```

**Rules**:
- Base config (config.yaml) contains all keys with defaults
- Environment configs only override specific values
- Use environment variables for secrets (${VAR_NAME})
- Never commit passwords/secrets to git
- Use duration format: 2s, 5m, 1h, etc.
- Database password should use environment variable in production
- Always have separate configs for each environment
- Cache and tracing can be disabled per environment

---

## Port Layer (Request/Response Interfaces)

**Location**: `core/port/request.go` and `core/port/response.go`

**Purpose**: Defines common request/response structures and interfaces used across handlers.

### Request Structures (`core/port/request.go`)

```go
package port

// MetadataRequest provides common pagination and sorting parameters
// Embed this in list/search request structs
type MetadataRequest struct {
    Skip     uint64 `form:"skip,default=0" validate:"omitempty"`
    Limit    uint64 `form:"limit,default=10" validate:"omitempty"`
    OrderBy  string `form:"orderBy" validate:"omitempty"`
    SortType string `form:"sortType" validate:"omitempty"`
}
```

**Usage in Handlers**:
```go
// In handler/request.go
type ListUsersParams struct {
    port.MetadataRequest
    // Add additional filters here if needed
    Status string `form:"status" validate:"omitempty"`
}
```

### Response Structures (`core/port/response.go`)

```go
package port

import "io"

// Standard status messages for all operations
var (
    ListSuccess   StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "list retrieved successfully", Success: true}
    FetchSuccess  StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "data retrieved successfully", Success: true}
    CreateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 201, Message: "resource created successfully", Success: true}
    UpdateSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "resource updated successfully", Success: true}
    DeleteSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "resource deleted successfully", Success: true}
    CustomEnv     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "This is environment specific", Success: true}
)

// OTP-related status constants
var (
    OTPSuccess     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "OTP generated successfully", Success: true}
    OTPAuthSuccess StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 200, Message: "OTP authenticated successfully", Success: true}
)

// StatusCodeAndMessage is embedded in all response structs
// Provides consistent status code, success flag, and message
type StatusCodeAndMessage struct {
    StatusCode int    `json:"status_code"`
    Success    bool   `json:"success"`
    Message    string `json:"message"`
}

// Status returns HTTP status code (interface compliance)
func (s StatusCodeAndMessage) Status() int {
    return s.StatusCode
}

func (s StatusCodeAndMessage) ResponseType() string {
    return "standard"
}

func (s StatusCodeAndMessage) GetContentType() string {
    return "application/json"
}

func (s StatusCodeAndMessage) GetContentDisposition() string {
    return ""
}

func (s StatusCodeAndMessage) Object() []byte {
    return nil
}

// FileResponse for file downloads/uploads
type FileResponse struct {
    ContentDisposition string
    ContentType        string
    Data               []byte        // Memory-based payload
    Reader             io.ReadCloser // Optional streaming source
}

func (s FileResponse) GetContentType() string {
    return s.ContentType
}

func (s FileResponse) GetContentDisposition() string {
    return s.ContentDisposition
}

func (s FileResponse) ResponseType() string {
    return "file"
}

func (s FileResponse) Status() int {
    return 200
}

func (s FileResponse) Object() []byte {
    return s.Data
}

// Stream copies Reader to w if available; else writes Data
func (s FileResponse) Stream(w io.Writer) error {
    if s.Reader == nil {
        if len(s.Data) > 0 {
            _, err := w.Write(s.Data)
            return err
        }
        return nil
    }
    defer s.Reader.Close()
    _, err := io.Copy(w, s.Reader)
    return err
}

// MetaDataResponse provides pagination metadata
// Embed this in list response structs
type MetaDataResponse struct {
    Skip                 uint64 `json:"skip,default=0"`
    Limit                uint64 `json:"limit,default=10"`
    OrderBy              string `json:"order_by,omitempty"`
    SortType             string `json:"sort_type,omitempty"`
    TotalRecordsCount    int    `json:"total_records_count,omitempty"`
    ReturnedRecordsCount uint64 `json:"returned_records_count"`
}

// Helper function to create metadata response
func NewMetaDataResponse(skip, limit, total uint64) MetaDataResponse {
    return MetaDataResponse{
        Skip:                 skip,
        Limit:                limit,
        TotalRecordsCount:    int(total),
        ReturnedRecordsCount: limit,
    }
}
```

**Usage in Response DTOs**:
```go
// In handler/response/user.go
type UsersListResponse struct {
    port.StatusCodeAndMessage `json:",inline"`  // Adds status_code, success, message
    port.MetaDataResponse     `json:",inline"`  // Adds pagination metadata
    Data                      []UserResponse `json:"data"`
}

type UserCreateResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      UserResponse `json:"data"`
}
```

**Rules**:
- Use predefined `StatusCodeAndMessage` constants (ListSuccess, CreateSuccess, etc.)
- Embed `port.MetadataRequest` for list endpoints (provides pagination)
- Embed `port.StatusCodeAndMessage` in all response structs
- Embed `port.MetaDataResponse` for list responses
- Use `json:",inline"` to flatten embedded structs
- FileResponse for serving files/downloads
- Never modify port layer structs directly (they're shared)

---

## Domain Model Pattern

**Location**: `core/domain/{resource}.go`

**Purpose**: Represents the business entity with database mapping.

**Pattern**:
```go
package domain

import "time"

type {Resource} struct {
    ID        int64     `json:"id" db:"id"`
    Field1    string    `json:"field1" db:"field1"`
    Field2    string    `json:"field2" db:"field2"`
    Field3    int       `json:"field3" db:"field3"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

**Rules**:
- Use `snake_case` for JSON and DB tags
- Always include `ID`, `CreatedAt`, `UpdatedAt`
- Match DB column names exactly in `db:` tags
- Export all fields (capitalize first letter)
- Use appropriate Go types (int64 for IDs, time.Time for timestamps)

**Example**:
```go
package domain

import "time"

type Product struct {
    ID          int64     `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Price       float64   `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    CategoryID  int64     `json:"category_id" db:"category_id"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

---

## Repository Pattern

**Location**: `repo/postgres/{resource}.go`

**Purpose**: Handles all database operations for the resource.

**Pattern**:
```go
package repo

import (
    "context"
    "time"

    sq "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
    config "gitlab.cept.gov.in/it-2.0-common/api-config"
    dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
    "pisapi/core/domain"
)

type {Resource}Repository struct {
    db  *dblib.DB
    cfg *config.Config
}

func New{Resource}Repository(db *dblib.DB, cfg *config.Config) *{Resource}Repository {
    return &{Resource}Repository{
        db:  db,
        cfg: cfg,
    }
}

const {resource}Table = "{resources}"

// Create inserts a new {resource}
func (r *{Resource}Repository) Create(ctx context.Context, data domain.{Resource}) (domain.{Resource}, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Insert({resource}Table).
        Columns("field1", "field2", "field3").
        Values(data.Field1, data.Field2, data.Field3).
        Suffix("RETURNING id, field1, field2, field3, created_at, updated_at").
        PlaceholderFormat(sq.Dollar)

    var result domain.{Resource}
    err := dblib.Insert(ctx, r.db, query, &result)
    return result, err
}

// FindByID retrieves a {resource} by ID
func (r *{Resource}Repository) FindByID(ctx context.Context, id int64) (domain.{Resource}, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Select("id", "field1", "field2", "field3", "created_at", "updated_at").
        From({resource}Table).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    var result domain.{Resource}
    err := dblib.SelectOne(ctx, r.db, query, &result)
    if err != nil {
        if err == pgx.ErrNoRows {
            return result, err
        }
        return result, err
    }
    return result, nil
}

// List retrieves all {resources} with pagination
func (r *{Resource}Repository) List(ctx context.Context, skip, limit int64, orderBy, sortType string) ([]domain.{Resource}, int64, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
    defer cancel()

    // Count query
    countQuery := sq.Select("COUNT(*)").
        From({resource}Table).
        PlaceholderFormat(sq.Dollar)

    var totalCount int64
    err := dblib.SelectOne(ctx, r.db, countQuery, &totalCount)
    if err != nil {
        return nil, 0, err
    }

    // Data query
    query := sq.Select("id", "field1", "field2", "field3", "created_at", "updated_at").
        From({resource}Table).
        OrderBy(orderBy + " " + sortType).
        Limit(uint64(limit)).
        Offset(uint64(skip)).
        PlaceholderFormat(sq.Dollar)

    var results []domain.{Resource}
    err = dblib.SelectRows(ctx, r.db, query, &results)
    if err != nil {
        return nil, 0, err
    }

    return results, totalCount, nil
}

// Update updates a {resource} by ID
func (r *{Resource}Repository) Update(ctx context.Context, id int64, field1, field2 *string, field3 *int) (domain.{Resource}, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Update({resource}Table).
        Set("updated_at", time.Now()).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    // Only update non-nil fields
    if field1 != nil {
        query = query.Set("field1", *field1)
    }
    if field2 != nil {
        query = query.Set("field2", *field2)
    }
    if field3 != nil {
        query = query.Set("field3", *field3)
    }

    query = query.Suffix("RETURNING id, field1, field2, field3, created_at, updated_at")

    var result domain.{Resource}
    err := dblib.Update(ctx, r.db, query, &result)
    return result, err
}

// Delete deletes a {resource} by ID
func (r *{Resource}Repository) Delete(ctx context.Context, id int64) error {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Delete({resource}Table).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    return dblib.Delete(ctx, r.db, query)
}
```

**Rules**:
- Always inject `*dblib.DB` and `*config.Config`
- Use context with timeout for all queries
- Use Squirrel query builder (alias `sq`)
- Always use `.PlaceholderFormat(sq.Dollar)` for PostgreSQL
- Use `dblib.Insert()`, `dblib.SelectOne()`, `dblib.SelectRows()`, `dblib.Update()`, `dblib.Delete()`
- Handle `pgx.ErrNoRows` for not found errors
- For updates: use pointers for optional fields, only update non-nil fields
- Always set `updated_at` in update queries
- Return domain models, not DTOs

---

## Handler Pattern

**Location**: `handler/{resource}.go`

**Purpose**: Defines HTTP routes and handles HTTP requests.

**Pattern**:
```go
package handler

import (
    "github.com/jackc/pgx/v5"
    log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
    serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
    serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
    "pisapi/core/port"
    resp "pisapi/handler/response"
    repo "pisapi/repo/postgres"
)

type {Resource}Handler struct {
    *serverHandler.Base
    svc *repo.{Resource}Repository
}

func New{Resource}Handler(svc *repo.{Resource}Repository) *{Resource}Handler {
    base := serverHandler.New("{Resources}").
        SetPrefix("/v1").
        AddPrefix("")
    return &{Resource}Handler{
        Base: base,
        svc:  svc,
    }
}

// Routes defines all routes for this handler
func (h *{Resource}Handler) Routes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.POST("/{resources}", h.Create{Resource}).Name("Create {Resource}"),
        serverRoute.GET("/{resources}", h.List{Resources}).Name("List {Resources}"),
        serverRoute.GET("/{resources}/:id", h.Get{Resource}ByID).Name("Get {Resource} By ID"),
        serverRoute.PUT("/{resources}/:id", h.Update{Resource}ByID).Name("Update {Resource} By ID"),
        serverRoute.DELETE("/{resources}/:id", h.Delete{Resource}ByID).Name("Delete {Resource} By ID"),
    }
}

// Create{Resource} creates a new {resource}
func (h *{Resource}Handler) Create{Resource}(sctx *serverRoute.Context, req Create{Resource}Request) (*resp.{Resource}CreateResponse, error) {
    // Convert request to domain model
    data := req.ToDomain()

    // Call repository
    result, err := h.svc.Create(sctx.Ctx, data)
    if err != nil {
        log.Error(sctx.Ctx, "Error creating {resource}: %v", err)
        return nil, err
    }

    log.Info(sctx.Ctx, "{Resource} created with ID: %d", result.ID)
    // Convert to response
    r := &resp.{Resource}CreateResponse{
        StatusCodeAndMessage: port.CreateSuccess,
        Data:                 resp.New{Resource}Response(result),
    }
    return r, nil
}

// List{Resources} retrieves all {resources}
func (h *{Resource}Handler) List{Resources}(sctx *serverRoute.Context, req List{Resources}Params) (*resp.{Resources}ListResponse, error) {
    // Call repository
    results, totalCount, err := h.svc.List(sctx.Ctx, req.Skip, req.Limit, req.OrderBy, req.SortType)
    if err != nil {
        log.Error(sctx.Ctx, "Error fetching {resources}: %v", err)
        return nil, err
    }

    // Convert to response
    r := &resp.{Resources}ListResponse{
        StatusCodeAndMessage: port.ListSuccess,
        MetaDataResponse: port.MetaDataResponse{
            TotalCount: totalCount,
            Count:      int64(len(results)),
            Skip:       req.Skip,
            Limit:      req.Limit,
        },
        Data: resp.New{Resources}Response(results),
    }
    return r, nil
}

// Get{Resource}ByID retrieves a {resource} by ID
func (h *{Resource}Handler) Get{Resource}ByID(sctx *serverRoute.Context, req {Resource}IDUri) (*resp.{Resource}FetchResponse, error) {
    // Call repository
    result, err := h.svc.FindByID(sctx.Ctx, req.ID)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Error(sctx.Ctx, "{Resource} not found with ID: %d", req.ID)
            return nil, err
        }
        log.Error(sctx.Ctx, "Error fetching {resource} by ID: %v", err)
        return nil, err
    }

    // Convert to response
    r := &resp.{Resource}FetchResponse{
        StatusCodeAndMessage: port.FetchSuccess,
        Data:                 resp.New{Resource}Response(result),
    }
    return r, nil
}

// Update{Resource}ByID updates a {resource} by ID
func (h *{Resource}Handler) Update{Resource}ByID(sctx *serverRoute.Context, req Update{Resource}Request) (*resp.{Resource}UpdateResponse, error) {
    // Convert non-empty fields to pointers
    var field1, field2 *string
    var field3 *int

    if req.Field1 != "" {
        field1 = &req.Field1
    }
    if req.Field2 != "" {
        field2 = &req.Field2
    }
    if req.Field3 != 0 {
        field3 = &req.Field3
    }

    // Call repository
    result, err := h.svc.Update(sctx.Ctx, req.ID, field1, field2, field3)
    if err != nil {
        log.Error(sctx.Ctx, "Error updating {resource} by ID: %v", err)
        return nil, err
    }

    // Convert to response
    r := &resp.{Resource}UpdateResponse{
        StatusCodeAndMessage: port.UpdateSuccess,
        Data:                 resp.New{Resource}Response(result),
    }
    return r, nil
}

// Delete{Resource}ByID deletes a {resource} by ID
func (h *{Resource}Handler) Delete{Resource}ByID(sctx *serverRoute.Context, req {Resource}IDUri) (*resp.{Resource}DeleteResponse, error) {
    // Call repository
    err := h.svc.Delete(sctx.Ctx, req.ID)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Error(sctx.Ctx, "{Resource} not found with ID: %d", req.ID)
            return nil, err
        }
        log.Error(sctx.Ctx, "Error deleting {resource} by ID: %v", err)
        return nil, err
    }

    // Return success response
    r := &resp.{Resource}DeleteResponse{
        StatusCodeAndMessage: port.DeleteSuccess,
    }
    return r, nil
}
```

**Rules**:
- Embed `*serverHandler.Base`
- Inject repository as `svc` with correct type (e.g., `*repo.UserRepository`)
- Import repository package as `repo "pisapi/repo/postgres"`
- Import log package as `log "gitlab.cept.gov.in/it-2.0-common/n-api-log"`
- Use `serverHandler.New()` with resource name (plural, capitalized)
- Set prefix to `/v1` for API versioning
- Handler signature: `(sctx *serverRoute.Context, req RequestType) (*ResponseType, error)`
- Always log errors before returning using `log.Error(sctx.Ctx, "message: %v", err)`
- Use `log.Info(sctx.Ctx, "message: %v", value)` for info logging
- Logging format: `log.Error(sctx.Ctx, "Error description: %v", err)` with printf-style formatting
- Check for `pgx.ErrNoRows` for 404 errors in repository errors
- For updates: no need to check existence first, handle error from Update
- For deletes: no need to check existence first, handle error from Delete (returns pgx.ErrNoRows if not found)
- Use `sctx.Ctx` for context parameter
- Always create response in intermediate variable `r`, then return `r, nil` (not inline return)

---

## Request DTO Pattern

**Location**: `handler/request.go`

**Purpose**: Defines request data transfer objects with validation.

**Pattern**:
```go
package handler

import "pisapi/core/domain"

// Create{Resource}Request represents the request body for creating a {resource}
type Create{Resource}Request struct {
    Field1 string `json:"field1" validate:"required"`
    Field2 string `json:"field2" validate:"required"`
    Field3 int    `json:"field3" validate:"required"`
}

func (r Create{Resource}Request) ToDomain() domain.{Resource} {
    return domain.{Resource}{
        Field1: r.Field1,
        Field2: r.Field2,
        Field3: r.Field3,
    }
}

// Update{Resource}Request represents the request body for updating a {resource}
type Update{Resource}Request struct {
    ID     int64  `uri:"id" validate:"required"`
    Field1 string `json:"field1" validate:"omitempty"`
    Field2 string `json:"field2" validate:"omitempty"`
    Field3 int    `json:"field3" validate:"omitempty"`
}

// {Resource}IDUri represents the URI parameter for {resource} ID
type {Resource}IDUri struct {
    ID int64 `uri:"id" validate:"required"`
}

// List{Resources}Params represents query parameters for listing {resources}
type List{Resources}Params struct {
    port.MetadataRequest
}
```

**Rules**:
- Add all request structs to `handler/request.go`
- Use `validate:"required"` for mandatory fields
- Use `validate:"omitempty"` for optional fields (updates)
- Use `uri:` tag for URL parameters
- Use `json:` tag for JSON body fields
- Use `form:` tag for form data
- Embed `port.MetadataRequest` for list endpoints (provides Skip, Limit, OrderBy, SortType)
- Include `ToDomain()` method for create requests
- Use `snake_case` for JSON field names

**Validation Tags**:
- `required` - Field must not be empty
- `omitempty` - Field is optional
- `email` - Must be valid email format
- `min=N` - Minimum value/length
- `max=N` - Maximum value/length
- `oneof=val1 val2` - Must be one of specified values

---

## Response DTO Pattern

**Location**: `handler/response/{resource}.go`

**Purpose**: Defines response data transfer objects.

**Pattern**:
```go
package response

import (
    "pisapi/core/domain"
    "pisapi/core/port"
)

// {Resource}Response represents a {resource} in API responses
type {Resource}Response struct {
    ID        int64  `json:"id"`
    Field1    string `json:"field1"`
    Field2    string `json:"field2"`
    Field3    int    `json:"field3"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

// New{Resource}Response converts domain model to response DTO
func New{Resource}Response(d domain.{Resource}) {Resource}Response {
    return {Resource}Response{
        ID:        d.ID,
        Field1:    d.Field1,
        Field2:    d.Field2,
        Field3:    d.Field3,
        CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt: d.UpdatedAt.Format("2006-01-02 15:04:05"),
    }
}

// New{Resources}Response converts slice of domain models to response DTOs
func New{Resources}Response(data []domain.{Resource}) []{Resource}Response {
    res := make([]{Resource}Response, 0, len(data))
    for _, d := range data {
        res = append(res, New{Resource}Response(d))
    }
    return res
}

// {Resource}CreateResponse represents the response for creating a {resource}
type {Resource}CreateResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      {Resource}Response `json:"data"`
}

// {Resource}FetchResponse represents the response for fetching a single {resource}
type {Resource}FetchResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      {Resource}Response `json:"data"`
}

// {Resources}ListResponse represents the response for listing {resources}
type {Resources}ListResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    port.MetaDataResponse     `json:",inline"`
    Data                      []{Resource}Response `json:"data"`
}

// {Resource}UpdateResponse represents the response for updating a {resource}
type {Resource}UpdateResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      {Resource}Response `json:"data"`
}

// {Resource}DeleteResponse represents the response for deleting a {resource}
type {Resource}DeleteResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
}
```

**Rules**:
- Create separate response structs for each operation (Create, Fetch, List, Update, Delete)
- Embed `port.StatusCodeAndMessage` for status info
- Embed `port.MetaDataResponse` for list responses (pagination)
- Use `json:",inline"` for embedded structs
- Provide conversion functions: `New{Resource}Response()` and `New{Resources}Response()`
- Format timestamps as strings: `"2006-01-02 15:04:05"`
- Use `snake_case` for JSON field names

**Standard Response Structures**:
- Create: `{StatusCodeAndMessage, Data: {Resource}Response}`
- Fetch: `{StatusCodeAndMessage, Data: {Resource}Response}`
- List: `{StatusCodeAndMessage, MetaDataResponse, Data: []{Resource}Response}`
- Update: `{StatusCodeAndMessage, Data: {Resource}Response}`
- Delete: `{StatusCodeAndMessage}` (no data)

---

## Routing Pattern

**Routes Definition**:
```go
func (h *{Resource}Handler) Routes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.POST("/{resources}", h.Create{Resource}).Name("Create {Resource}"),
        serverRoute.GET("/{resources}", h.List{Resources}).Name("List {Resources}"),
        serverRoute.GET("/{resources}/:id", h.Get{Resource}ByID).Name("Get {Resource} By ID"),
        serverRoute.PUT("/{resources}/:id", h.Update{Resource}ByID).Name("Update {Resource} By ID"),
        serverRoute.DELETE("/{resources}/:id", h.Delete{Resource}ByID).Name("Delete {Resource} By ID"),
    }
}
```

**RESTful Conventions**:
| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| POST | `/{resources}` | `Create{Resource}` | Create new resource |
| GET | `/{resources}` | `List{Resources}` | List all resources |
| GET | `/{resources}/:id` | `Get{Resource}ByID` | Get single resource |
| PUT | `/{resources}/:id` | `Update{Resource}ByID` | Update resource |
| DELETE | `/{resources}/:id` | `Delete{Resource}ByID` | Delete resource |

**Rules**:
- Use plural for collection endpoints (`/users`)
- Use `:id` for path parameters
- Use `.Name()` for Swagger documentation
- Prefix is set in handler constructor (`/v1`)
- Final URL: `/v1/{resources}` or `/v1/{resources}/:id`

---

## Validation Pattern

**Auto-generated Validators**:
- Run `govalid` tool to generate validators
- Generated files: `handler/request_*_validator.go`
- Implements `Validator` interface with `Validate()` method

**Manual Generation**:
```bash
# Navigate to handler directory
cd handler

# Run govalid
govalid
```

**Generated Validator Example**:
```go
// Auto-generated by govalid
func (r Create{Resource}Request) Validate() error {
    var validationErrors []ValidationError

    if r.Field1 == "" {
        validationErrors = append(validationErrors, ValidationError{
            Reason: "Field1 is required",
            Path:   "field1",
            Type:   "required",
            Value:  r.Field1,
        })
    }

    if len(validationErrors) > 0 {
        return ValidationErrors(validationErrors)
    }
    return nil
}
```

**Rules**:
- Validators are auto-generated, do not modify
- Add validation tags to request structs
- Re-run `govalid` after modifying request structs
- Framework automatically calls `Validate()` before handler execution

---

## Database Schema

**Location**: `db/{resource}.sql`

**Pattern**:
```sql
CREATE TABLE IF NOT EXISTS {resources} (
    id SERIAL PRIMARY KEY,
    field1 VARCHAR(255) NOT NULL,
    field2 VARCHAR(255) NOT NULL,
    field3 INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add indexes for frequently queried fields
CREATE INDEX IF NOT EXISTS idx_{resources}_field1 ON {resources}(field1);

-- Add unique constraints if needed
ALTER TABLE {resources} ADD CONSTRAINT unique_{resources}_field1 UNIQUE (field1);
```

**Rules**:
- Use `SERIAL` for auto-incrementing IDs
- Always include `created_at` and `updated_at` with `DEFAULT NOW()`
- Use appropriate data types:
  - `VARCHAR(N)` for strings
  - `INTEGER` for whole numbers
  - `DECIMAL(P,S)` for money/decimals
  - `TIMESTAMP` for dates/times
  - `BOOLEAN` for true/false
  - `TEXT` for large text
- Add indexes for foreign keys and frequently queried fields
- Add unique constraints where applicable
- Use `IF NOT EXISTS` to make migrations idempotent

---

## Database Query Optimization Patterns

**Location**: Repository layer implementation

**Purpose**: Minimize database round trips through intelligent query consolidation and PostgreSQL-specific features using dblib/Squirrel patterns.

### Decision Tree: Query Optimization Strategy

```
Need to perform operation?
│
├─ Single simple query (SELECT/INSERT/UPDATE)?
│  └─ Use dblib.Psql with appropriate method
│
├─ 2+ queries needed?
│  │
│  ├─ Same transaction scope?
│  │  └─ Use pgx.Batch (see next section)
│  │
│  └─ Independent queries?
│     └─ Execute separately with dblib.Psql
│
├─ Insert data derived from existing table?
│  └─ Use raw SQL INSERT...SELECT (Squirrel doesn't support well)
│
├─ Update based on query results?
│  └─ Use raw SQL UPDATE...FROM (Squirrel doesn't support FROM clause)
│
├─ Complex multi-step with dependencies?
│  └─ Use raw SQL WITH (CTE) - Squirrel doesn't support CTEs
│
└─ Simple single-table operations?
   └─ Use dblib.Psql.Select/Insert/Update/Delete (Squirrel builder)
```

### When to Use Squirrel vs Raw SQL

| Operation Type | Use | Reason |
|----------------|-----|--------|
| Simple SELECT/INSERT/UPDATE | dblib.Psql with Squirrel | Clean, type-safe, composable |
| INSERT...SELECT | Raw SQL with r.db.Exec | Squirrel doesn't support SELECT in INSERT |
| UPDATE...FROM | Raw SQL with r.db.Exec | Squirrel doesn't support FROM in UPDATE |
| WITH (CTE) | Raw SQL with r.db.Exec | Squirrel doesn't support CTEs |
| Batch operations | pgx.Batch with Squirrel queries | Best of both worlds |

### Pattern 1: Simple Operations with dblib.Psql (Squirrel)

**When to Use**: Single-table SELECT/INSERT/UPDATE/DELETE

**✅ Use Squirrel Builder**:
```go
func (r *PolicyRepository) GetPolicyByID(ctx context.Context, policyID string) (*domain.Policy, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    // Use Squirrel query builder
    query := dblib.Psql.Select(
        "id", "policy_number", "sum_assured", "status", "maturity_date",
    ).From("policies").
        Where(sq.Eq{"id": policyID})

    sql, args, _ := query.ToSql()
    
    var policy domain.Policy
    err := r.db.QueryRow(cCtx, sql, args...).Scan(
        &policy.ID, &policy.PolicyNumber, &policy.SumAssured, 
        &policy.Status, &policy.MaturityDate,
    )
    
    return &policy, err
}

func (r *ClaimRepository) UpdateClaimStatus(ctx context.Context, claimID, status string) error {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    // Use Squirrel for simple update
    query := dblib.Psql.Update("claims").
        Set("status", status).
        Set("updated_at", time.Now()).
        Where(sq.Eq{"id": claimID})

    sql, args, _ := query.ToSql()
    _, err := r.db.Exec(cCtx, sql, args...)
    
    return err
}
```

### Pattern 2: INSERT...SELECT (Raw SQL - Squirrel Limitation)

**When to Use**: Inserting data derived from existing tables

**❌ Avoid This** (2 round trips with Squirrel):
```go
// Round trip 1: SELECT with Squirrel
query := dblib.Psql.Select("id", "sum_assured").
    From("policies").
    Where(sq.Eq{"status": "ACTIVE"})

sql, args, _ := query.ToSql()
rows, _ := r.db.Query(ctx, sql, args...)

// Round trip 2: INSERT each row
for rows.Next() {
    var policyID string
    var sumAssured decimal.Decimal
    rows.Scan(&policyID, &sumAssured)
    
    insertQuery := dblib.Psql.Insert("maturity_claims").
        Columns("policy_id", "sum_assured", "status").
        Values(policyID, sumAssured, "PENDING")
    // ... execute
}
```

**✅ Use Raw SQL Instead** (1 round trip):
```go
func (r *ClaimRepository) GenerateMaturityClaims(ctx context.Context, targetDate time.Time) error {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    // Use raw SQL for INSERT...SELECT (Squirrel doesn't support this well)
    query := `
        INSERT INTO maturity_claims (
            policy_id, 
            claim_date, 
            sum_assured, 
            bonus_amount,
            status
        )
        SELECT 
            p.id,
            $1,
            p.sum_assured,
            p.accumulated_bonus,
            'PENDING'
        FROM policies p
        WHERE p.maturity_date = $1
          AND p.status = 'ACTIVE'
          AND NOT EXISTS (
              SELECT 1 FROM maturity_claims mc 
              WHERE mc.policy_id = p.id
          )
    `
    
    _, err := r.db.Exec(cCtx, query, targetDate)
    return err
}
```

### Pattern 3: UPDATE...FROM (Raw SQL - Squirrel Limitation)

**When to Use**: Update based on data from another table

**❌ Avoid This** (N+1 round trips):
```go
// SELECT approvals
selectQuery := dblib.Psql.Select("claim_id", "approved_amount").
    From("approvals").
    Where(sq.Eq{"status": "APPROVED"})

sql, args, _ := selectQuery.ToSql()
rows, _ := r.db.Query(ctx, sql, args...)

// UPDATE each claim (N queries!)
for rows.Next() {
    var claimID string
    var amount decimal.Decimal
    rows.Scan(&claimID, &amount)
    
    updateQuery := dblib.Psql.Update("claims").
        Set("amount", amount).
        Set("status", "APPROVED").
        Where(sq.Eq{"id": claimID})
    
    sql, args, _ := updateQuery.ToSql()
    r.db.Exec(ctx, sql, args...)
}
```

**✅ Use Raw SQL UPDATE...FROM** (1 round trip):
```go
func (r *ClaimRepository) UpdateClaimsFromApprovals(ctx context.Context) error {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    // Use raw SQL for UPDATE...FROM (Squirrel doesn't support FROM clause)
    query := `
        UPDATE claims c
        SET 
            amount = a.approved_amount,
            status = 'APPROVED',
            approved_at = a.approved_at,
            approved_by = a.approved_by,
            updated_at = NOW()
        FROM approvals a
        WHERE c.id = a.claim_id
          AND a.status = 'APPROVED'
          AND c.status = 'PENDING_APPROVAL'
    `
    
    _, err := r.db.Exec(cCtx, query)
    return err
}
```

### Pattern 4: WITH (CTE) for Complex Multi-Step (Raw SQL)

**When to Use**: Multiple dependent queries that can be consolidated

**❌ Avoid This** (3 round trips):
```go
// Step 1: Get policy data with Squirrel
query1 := dblib.Psql.Select("sum_assured", "accumulated_bonus", "loan_outstanding").
    From("policies").
    Where(sq.Eq{"id": policyID})

var sumAssured, bonus, loan decimal.Decimal
r.db.QueryRow(ctx, sql1, args1...).Scan(&sumAssured, &bonus, &loan)

// Step 2: Insert claim
claimAmount := sumAssured + bonus - loan
query2 := dblib.Psql.Insert("claims").
    Columns("policy_id", "claim_amount", "status").
    Values(policyID, claimAmount, "REGISTERED").
    Suffix("RETURNING id")

var claimID string
r.db.QueryRow(ctx, sql2, args2...).Scan(&claimID)

// Step 3: Update policy
query3 := dblib.Psql.Update("policies").
    Set("status", "CLAIMED").
    Where(sq.Eq{"id": policyID})

r.db.Exec(ctx, sql3, args3...)
```

**✅ Use Raw SQL WITH (CTE)** (1 round trip):
```go
func (r *ClaimRepository) CreateDeathClaim(ctx context.Context, req domain.DeathClaimRequest) (*domain.ClaimID, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    // Use raw SQL for CTE (Squirrel doesn't support CTEs)
    query := `
        WITH claim_calculation AS (
            -- Step 1: Calculate claim amount
            SELECT 
                id as policy_id,
                sum_assured + accumulated_bonus - loan_outstanding as claim_amount
            FROM policies
            WHERE id = $1
              AND status = 'ACTIVE'
        ),
        new_claim AS (
            -- Step 2: Insert claim
            INSERT INTO claims (
                policy_id,
                claim_type,
                claim_amount,
                claimant_name,
                death_date,
                status
            )
            SELECT 
                cc.policy_id,
                'DEATH',
                cc.claim_amount,
                $2,
                $3,
                'REGISTERED'
            FROM claim_calculation cc
            RETURNING id, policy_id
        )
        -- Step 3: Update policy status
        UPDATE policies p
        SET 
            status = 'CLAIMED',
            claim_date = NOW(),
            updated_at = NOW()
        FROM new_claim nc
        WHERE p.id = nc.policy_id
        RETURNING nc.id
    `
    
    var claimID domain.ClaimID
    err := r.db.QueryRow(cCtx, query, 
        req.PolicyID, 
        req.ClaimantName, 
        req.DeathDate,
    ).Scan(&claimID.ID)
    
    return &claimID, err
}
```

### Pattern 5: Hybrid - Squirrel in pgx.Batch for Complex Operations

**When to Use**: Multiple operations where some can use Squirrel, some need raw SQL

**✅ Best of Both Worlds**:
```go
func (r *ClaimRepository) ProcessClaimWithInvestigation(ctx context.Context, req domain.DeathClaimRequest) (*domain.Claim, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    batch := &pgx.Batch{}

    // Query 1: Simple select - use Squirrel
    query1 := dblib.Psql.Select(
        "id", "sum_assured", "accumulated_bonus", "status",
    ).From("policies").
        Where(sq.Eq{"id": req.PolicyID})

    var policy domain.Policy
    dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.Policy], &policy)

    // Query 2: Simple insert - use Squirrel
    query2 := dblib.Psql.Insert("claims").
        Columns("policy_id", "claim_type", "status").
        Values(req.PolicyID, "DEATH", "REGISTERED").
        Suffix("RETURNING id")

    var claimID domain.ClaimID
    dblib.QueueReturnRow(batch, query2, pgx.RowToStructByNameLax[domain.ClaimID], &claimID)

    // Query 3: Complex CTE - use raw SQL (wrap in Squirrel for batch)
    rawCTE := `
        WITH new_investigation AS (
            INSERT INTO investigations (claim_id, investigation_type, due_date)
            SELECT $1, 
                   CASE WHEN EXTRACT(YEAR FROM AGE($2, p.issue_date)) < 3 
                        THEN 'MANDATORY' ELSE 'DISCRETIONARY' END,
                   NOW() + INTERVAL '21 days'
            FROM policies p WHERE p.id = $3
            RETURNING id
        )
        INSERT INTO investigation_tasks (investigation_id, task_name, status)
        SELECT id, task_name, 'PENDING'
        FROM new_investigation
        CROSS JOIN (VALUES ('Verify Death Certificate'), ('Field Visit')) AS tasks(task_name)
    `
    
    // Use Expr to wrap raw SQL in Squirrel
    query3 := sq.Expr(rawCTE, req.PolicyID, req.DeathDate, req.PolicyID)
    dblib.QueueExecRow(batch, query3)

    // Execute batch
    err := r.db.SendBatch(cCtx, batch).Close()
    if err != nil {
        return nil, err
    }

    return &domain.Claim{
        ID:       claimID.ID,
        PolicyID: req.PolicyID,
        Status:   "REGISTERED",
    }, nil
}
```

### Pattern 6: Dynamic Query Building with Squirrel

**When to Use**: Queries with optional filters/conditions

**✅ Squirrel Excels Here**:
```go
func (r *ClaimRepository) SearchClaims(ctx context.Context, filters domain.ClaimFilters) ([]domain.Claim, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    // Start with base query
    query := dblib.Psql.Select(
        "id", "policy_id", "claim_type", "status", "claim_amount",
    ).From("claims")

    // Add filters conditionally
    if filters.PolicyID != "" {
        query = query.Where(sq.Eq{"policy_id": filters.PolicyID})
    }
    
    if filters.ClaimType != "" {
        query = query.Where(sq.Eq{"claim_type": filters.ClaimType})
    }
    
    if len(filters.Statuses) > 0 {
        query = query.Where(sq.Eq{"status": filters.Statuses})
    }
    
    if !filters.FromDate.IsZero() {
        query = query.Where(sq.GtOrEq{"claim_date": filters.FromDate})
    }
    
    if !filters.ToDate.IsZero() {
        query = query.Where(sq.LtOrEq{"claim_date": filters.ToDate})
    }

    // Add ordering
    query = query.OrderBy("claim_date DESC").Limit(filters.Limit)

    // Execute
    sql, args, _ := query.ToSql()
    rows, err := r.db.Query(cCtx, sql, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var claims []domain.Claim
    for rows.Next() {
        var claim domain.Claim
        err := rows.Scan(&claim.ID, &claim.PolicyID, &claim.ClaimType, 
                        &claim.Status, &claim.ClaimAmount)
        if err != nil {
            return nil, err
        }
        claims = append(claims, claim)
    }

    return claims, nil
}
```

### Key Principles

1. **Use Squirrel for Simple Operations**: SELECT/INSERT/UPDATE/DELETE on single tables
2. **Use Raw SQL for Complex Operations**: INSERT...SELECT, UPDATE...FROM, CTEs
3. **Combine Both in pgx.Batch**: Squirrel queries alongside raw SQL in same batch
4. **Think in Sets**: Avoid row-by-row processing
5. **Minimize Round Trips**: Consolidate operations using CTEs or batch
6. **Dynamic Queries**: Leverage Squirrel's builder pattern for conditional filters

---

## pgx.Batch Patterns (Replace Transactions)

**Location**: Repository layer when multiple queries needed in same transaction

**Purpose**: Execute multiple queries in a single round trip with implicit transaction semantics.

**CRITICAL RULE**: ❌ NEVER use explicit transactions (BEGIN/COMMIT). ✅ ALWAYS use pgx.Batch.

### Why pgx.Batch Instead of Transactions?

| Feature | Traditional Transaction | pgx.Batch |
|---------|------------------------|-----------|
| Round Trips | Multiple (BEGIN, query1, query2, COMMIT) | Single round trip |
| Network Latency | High (4 round trips minimum) | Low (1 round trip) |
| Transaction Semantics | Explicit | Implicit |
| Performance | Slower | Faster |
| Code Complexity | Higher | Lower |

### Decision Tree: When to Use pgx.Batch?

```
How many queries do you need to execute?
│
├─ 1 query?
│  └─ Use db.QueryRow() or db.Exec() directly
│  
├─ 2+ queries?
│  │
│  ├─ Need to return value(s)?
│  │  │
│  │  ├─ Return single struct?
│  │  │  └─ Use QueueReturnRow()
│  │  │
│  │  └─ Return multiple rows?
│  │     └─ Use QueueReturn()
│  │
│  └─ No return value needed?
│     └─ Use QueueExecRow()
│
└─ Queries are independent (no transaction needed)?
   └─ Execute separately, don't use Batch
```

### Available Batch Functions

From `dblib` package:

1. **QueueReturnRow(batch, query, rowMapper, destination)**
   - Use when: Query returns single row and you need the result
   - Returns: Single struct/value
   - Example: INSERT...RETURNING id

2. **QueueReturn(batch, query, rowMapper)**
   - Use when: Query returns multiple rows
   - Returns: Slice of structs
   - Example: SELECT multiple records

3. **QueueExecRow(batch, query)**
   - Use when: Query doesn't return data (INSERT, UPDATE, DELETE without RETURNING)
   - Returns: Nothing (command tag only)
   - Example: UPDATE, DELETE

### Pattern 1: Basic Batch (Insert + Update)

**Scenario**: Create order and update inventory

```go
func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order, items []domain.OrderItem) (*domain.OrderID, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    batch := &pgx.Batch{}

    // Query 1: Insert order and return ID
    query1 := dblib.Psql.Insert("orders").
        Columns("user_id", "total_amount", "shipping_address", "payment_method", "status").
        Values(order.UserID, order.TotalAmount, order.ShippingAddress, order.PaymentMethod, "PENDING").
        Suffix("RETURNING id")

    var orderID domain.OrderID
    dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.OrderID], &orderID)

    // Query 2: Insert order items (no return value needed)
    query2 := dblib.Psql.Insert("order_items").
        Columns("order_id", "product_id", "quantity", "unit_price")

    for _, item := range items {
        // Note: orderID will be populated after batch execution
        query2 = query2.Values(item.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
    }
    dblib.QueueExecRow(batch, query2)

    // Query 3: Update product stock
    for _, item := range items {
        query3 := dblib.Psql.Update("products").
            Set("stock_quantity", sq.Expr("stock_quantity - ?", item.Quantity)).
            Where(sq.Eq{"id": item.ProductID})
        
        dblib.QueueExecRow(batch, query3)
    }

    // Execute all queries in single round trip
    err := r.db.SendBatch(cCtx, batch).Close()
    if err != nil {
        return nil, err
    }

    return &orderID, nil
}
```

### Pattern 2: Batch with Multiple Returns

**Scenario**: Insert multiple claims and get their IDs

```go
func (r *ClaimRepository) CreateMaturityClaimsBatch(ctx context.Context, policies []domain.Policy) ([]domain.ClaimID, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    batch := &pgx.Batch{}
    claimIDs := make([]domain.ClaimID, 0, len(policies))

    for _, policy := range policies {
        query := dblib.Psql.Insert("claims").
            Columns("policy_id", "claim_type", "claim_amount", "status", "claim_date").
            Values(policy.ID, "MATURITY", policy.MaturityAmount, "PENDING", policy.MaturityDate).
            Suffix("RETURNING id")
        
        var claimID domain.ClaimID
        dblib.QueueReturnRow(batch, query, pgx.RowToStructByNameLax[domain.ClaimID], &claimID)
        claimIDs = append(claimIDs, claimID)
    }

    err := r.db.SendBatch(cCtx, batch).Close()
    if err != nil {
        return nil, err
    }

    return claimIDs, nil
}
```

### Pattern 3: Complex Batch (Read + Write)

**Scenario**: Validate policy, calculate amount, create claim, update policy

```go
func (r *ClaimRepository) ProcessDeathClaim(ctx context.Context, req domain.DeathClaimRequest) (*domain.Claim, error) {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    batch := &pgx.Batch{}

    // Query 1: Get policy details
    query1 := dblib.Psql.Select(
        "id", "sum_assured", "accumulated_bonus", "loan_outstanding", "status",
    ).From("policies").
        Where(sq.Eq{"id": req.PolicyID})

    var policy domain.Policy
    dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.Policy], &policy)

    // Query 2: Create claim (will use policy data after batch execution)
    // Note: Calculation happens in application after batch returns
    query2 := dblib.Psql.Insert("claims").
        Columns("policy_id", "claim_type", "claimant_name", "death_date", "status").
        Values(req.PolicyID, "DEATH", req.ClaimantName, req.DeathDate, "REGISTERED").
        Suffix("RETURNING id, policy_id, claim_type, status, created_at")

    var claim domain.Claim
    dblib.QueueReturnRow(batch, query2, pgx.RowToStructByNameLax[domain.Claim], &claim)

    // Query 3: Update policy status
    query3 := dblib.Psql.Update("policies").
        Set("status", "CLAIMED").
        Set("claim_date", time.Now()).
        Where(sq.Eq{"id": req.PolicyID})

    dblib.QueueExecRow(batch, query3)

    // Execute batch
    err := r.db.SendBatch(cCtx, batch).Close()
    if err != nil {
        return nil, err
    }

    // Calculate claim amount from fetched policy data
    claim.ClaimAmount = policy.SumAssured + policy.AccumulatedBonus - policy.LoanOutstanding

    return &claim, nil
}
```

### Pattern 4: Batch with Error Handling

**Best Practice**: Check for specific errors after batch execution

```go
func (r *PaymentRepository) ProcessPayment(ctx context.Context, payment domain.Payment) error {
    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    batch := &pgx.Batch{}

    // Insert payment
    query1 := dblib.Psql.Insert("payments").
        Columns("claim_id", "amount", "payment_method", "status").
        Values(payment.ClaimID, payment.Amount, payment.PaymentMethod, "PROCESSING")
    dblib.QueueExecRow(batch, query1)

    // Update claim status
    query2 := dblib.Psql.Update("claims").
        Set("payment_status", "PROCESSING").
        Set("payment_initiated_at", time.Now()).
        Where(sq.Eq{"id": payment.ClaimID})
    dblib.QueueExecRow(batch, query2)

    // Execute batch
    err := r.db.SendBatch(cCtx, batch).Close()
    if err != nil {
        // Handle specific errors
        if strings.Contains(err.Error(), "duplicate key") {
            return fmt.Errorf("payment already exists for claim %s", payment.ClaimID)
        }
        if strings.Contains(err.Error(), "foreign key") {
            return fmt.Errorf("claim %s not found", payment.ClaimID)
        }
        return fmt.Errorf("failed to process payment: %w", err)
    }

    return nil
}
```

### Pattern 5: Dynamic Batch (Variable Query Count)

**Scenario**: Update multiple policies based on conditions

```go
func (r *PolicyRepository) BulkUpdatePolicyStatus(ctx context.Context, updates []domain.PolicyStatusUpdate) error {
    if len(updates) == 0 {
        return nil
    }

    cCtx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMedium"))
    defer cancel()

    batch := &pgx.Batch{}

    for _, update := range updates {
        query := dblib.Psql.Update("policies").
            Set("status", update.NewStatus).
            Set("status_reason", update.Reason).
            Set("updated_at", time.Now()).
            Set("updated_by", update.UpdatedBy).
            Where(sq.Eq{"id": update.PolicyID})

        dblib.QueueExecRow(batch, query)

        // Also create audit log entry
        auditQuery := dblib.Psql.Insert("policy_audit").
            Columns("policy_id", "action", "old_status", "new_status", "reason", "created_by").
            Values(update.PolicyID, "STATUS_CHANGE", update.OldStatus, update.NewStatus, update.Reason, update.UpdatedBy)
        
        dblib.QueueExecRow(batch, auditQuery)
    }

    err := r.db.SendBatch(cCtx, batch).Close()
    return err
}
```

### Key Principles

1. **Always Use Batch for 2+ Queries**: Single round trip = better performance
2. **No Explicit Transactions**: pgx.Batch provides implicit transaction semantics
3. **Choose Right Queue Function**: 
   - QueueReturnRow for single row result
   - QueueReturn for multiple rows
   - QueueExecRow for no results
4. **Handle Errors After Batch**: Check batch.Close() error for transaction failures
5. **Set Appropriate Timeouts**: Use context with timeout based on query complexity

---

## Temporal Workflow Integration

**Location**: Workflows in `workflows/` directory, called from handlers

**Purpose**: Orchestrate long-running, multi-step business processes with retries, compensation, and state management.

### Decision Tree: When to Use Temporal Workflow?

```
Analyzing endpoint requirements...
│
├─ Simple CRUD operation (< 1 second)?
│  └─ ❌ No workflow needed
│     Flow: Handler → Repository
│
├─ Process duration > 1 minute?
│  └─ ✅ Use Temporal workflow
│     Flow: Handler → Workflow → Activities → Repository
│
├─ Multi-step with waiting/delays?
│  └─ ✅ Use Temporal workflow
│     Example: Wait for investigation report (21 days)
│
├─ Requires human approval?
│  └─ ✅ Use Temporal workflow
│     Example: Claim approval by officer
│
├─ Needs retry/compensation logic?
│  └─ ✅ Use Temporal workflow
│     Example: Payment failure → refund
│
├─ External API calls with timeouts?
│  └─ ✅ Use Temporal workflow
│     Example: Bank verification API
│
└─ Complex business logic with SLA tracking?
   └─ ✅ Use Temporal workflow
      Example: Death claim with investigation
```

### Architecture Flow

#### **Flow 1: Without Temporal (Simple CRUD)**
```
Client Request
    ↓
Handler (validation)
    ↓
Repository (database)
    ↓
Response
```

#### **Flow 2: With Temporal (Complex Process)**
```
Client Request
    ↓
Handler (validation)
    ↓
Start Temporal Workflow
    ↓
Activity 1 (validate policy)
    ↓
Activity 2 (calculate amount)
    ↓
Activity 3 (create claim)
    ↓
Activity 4 (send notification)
    ↓
Return Workflow Result
    ↓
Response
```

### Pattern 1: Basic Workflow Structure

**File**: `workflows/death_claim_workflow.go`

```go
package workflows

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// DeathClaimWorkflowInput defines workflow input
type DeathClaimWorkflowInput struct {
    PolicyID     string
    ClaimantName string
    DeathDate    time.Time
    Documents    []string
}

// DeathClaimWorkflowResult defines workflow output
type DeathClaimWorkflowResult struct {
    ClaimID       string
    Status        string
    ClaimAmount   float64
    ProcessedAt   time.Time
}

// DeathClaimWorkflow orchestrates death claim processing
func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Death claim workflow started", "policyID", input.PolicyID)

    // Configure activity options
    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    1 * time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOptions)

    // Activity 1: Validate policy
    var policyValidation PolicyValidationResult
    err := workflow.ExecuteActivity(ctx, ValidatePolicyActivity, input.PolicyID).Get(ctx, &policyValidation)
    if err != nil {
        return nil, fmt.Errorf("policy validation failed: %w", err)
    }

    // Activity 2: Check investigation requirement
    var investigationRequired bool
    err = workflow.ExecuteActivity(ctx, CheckInvestigationRequirementActivity, 
        input.PolicyID, input.DeathDate).Get(ctx, &investigationRequired)
    if err != nil {
        return nil, err
    }

    // Activity 3: Register claim
    var claimID string
    err = workflow.ExecuteActivity(ctx, RegisterDeathClaimActivity, input).Get(ctx, &claimID)
    if err != nil {
        return nil, err
    }

    // Conditional: If investigation required, wait for completion
    if investigationRequired {
        logger.Info("Investigation required, waiting for completion")
        
        // Wait for investigation signal (max 21 days)
        investigationChannel := workflow.GetSignalChannel(ctx, "investigation-completed")
        
        var investigationResult InvestigationResult
        investigationChannel.Receive(ctx, &investigationResult)
        
        logger.Info("Investigation completed", "result", investigationResult.Status)
    }

    // Activity 4: Approve claim
    var approvalResult ApprovalResult
    err = workflow.ExecuteActivity(ctx, ApproveClaimActivity, claimID).Get(ctx, &approvalResult)
    if err != nil {
        return nil, err
    }

    // Activity 5: Process payment
    var paymentResult PaymentResult
    err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, claimID, approvalResult.ApprovedAmount).Get(ctx, &paymentResult)
    if err != nil {
        // Compensation: Rollback claim status
        workflow.ExecuteActivity(ctx, CompensateClaimActivity, claimID)
        return nil, err
    }

    // Activity 6: Send notification
    workflow.ExecuteActivity(ctx, SendNotificationActivity, claimID, "CLAIM_APPROVED")

    return &DeathClaimWorkflowResult{
        ClaimID:     claimID,
        Status:      "APPROVED",
        ClaimAmount: approvalResult.ApprovedAmount,
        ProcessedAt: workflow.Now(ctx),
    }, nil
}
```

### Pattern 2: Activity Implementation

**File**: `workflows/activities/claim_activities.go`

```go
package activities

import (
    "context"
    "pisapi/repo/postgres"
)

type ClaimActivities struct {
    claimRepo   *postgres.ClaimRepository
    policyRepo  *postgres.PolicyRepository
}

func NewClaimActivities(
    claimRepo *postgres.ClaimRepository,
    policyRepo *postgres.PolicyRepository,
) *ClaimActivities {
    return &ClaimActivities{
        claimRepo:  claimRepo,
        policyRepo: policyRepo,
    }
}

// ValidatePolicyActivity validates policy eligibility
func (a *ClaimActivities) ValidatePolicyActivity(ctx context.Context, policyID string) (*PolicyValidationResult, error) {
    // Call repository
    policy, err := a.policyRepo.GetByID(ctx, policyID)
    if err != nil {
        return nil, err
    }

    // Business validation
    if policy.Status != "ACTIVE" {
        return &PolicyValidationResult{
            Valid: false,
            Error: "Policy is not active",
        }, nil
    }

    return &PolicyValidationResult{
        Valid:       true,
        PolicyData:  policy,
    }, nil
}

// RegisterDeathClaimActivity creates claim record
func (a *ClaimActivities) RegisterDeathClaimActivity(ctx context.Context, input DeathClaimWorkflowInput) (string, error) {
    claim := domain.DeathClaim{
        PolicyID:     input.PolicyID,
        ClaimantName: input.ClaimantName,
        DeathDate:    input.DeathDate,
        Status:       "REGISTERED",
    }

    // Call repository
    claimID, err := a.claimRepo.CreateDeathClaim(ctx, claim)
    if err != nil {
        return "", err
    }

    return claimID.ID, nil
}

// ProcessPaymentActivity handles payment disbursement
func (a *ClaimActivities) ProcessPaymentActivity(ctx context.Context, claimID string, amount float64) (*PaymentResult, error) {
    payment := domain.Payment{
        ClaimID: claimID,
        Amount:  amount,
        Method:  "NEFT",
        Status:  "PROCESSING",
    }

    // Call repository
    paymentID, err := a.claimRepo.ProcessPayment(ctx, payment)
    if err != nil {
        return nil, err
    }

    return &PaymentResult{
        PaymentID: paymentID,
        Status:    "SUCCESS",
    }, nil
}
```

### Pattern 3: Handler Integration

**File**: `handler/claim_handler.go`

```go
package handler

import (
    "net/http"
    "pisapi/workflows"
    "go.temporal.io/sdk/client"
)

type ClaimHandler struct {
    temporalClient client.Client
}

func NewClaimHandler(temporalClient client.Client) *ClaimHandler {
    return &ClaimHandler{
        temporalClient: temporalClient,
    }
}

// RegisterDeathClaim starts death claim workflow
func (h *ClaimHandler) RegisterDeathClaim(c echo.Context) error {
    var req request.RegisterDeathClaimRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    // Validate request
    if err := req.Validate(); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    // Prepare workflow input
    workflowInput := workflows.DeathClaimWorkflowInput{
        PolicyID:     req.PolicyID,
        ClaimantName: req.ClaimantName,
        DeathDate:    req.DeathDate,
        Documents:    req.Documents,
    }

    // Start Temporal workflow
    workflowOptions := client.StartWorkflowOptions{
        ID:        "death-claim-" + req.PolicyID,
        TaskQueue: "claim-processing-queue",
    }

    workflowRun, err := h.temporalClient.ExecuteWorkflow(
        c.Request().Context(),
        workflowOptions,
        workflows.DeathClaimWorkflow,
        workflowInput,
    )
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start workflow"})
    }

    // Return workflow ID for tracking
    return c.JSON(http.StatusAccepted, map[string]string{
        "workflow_id": workflowRun.GetID(),
        "message":     "Death claim processing started",
    })
}

// GetClaimStatus queries workflow status
func (h *ClaimHandler) GetClaimStatus(c echo.Context) error {
    claimID := c.Param("claim_id")
    workflowID := "death-claim-" + claimID

    // Query workflow
    response, err := h.temporalClient.QueryWorkflow(
        c.Request().Context(),
        workflowID,
        "",
        "GetClaimStatus",
    )
    if err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "Workflow not found"})
    }

    var status string
    if err := response.Get(&status); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get status"})
    }

    return c.JSON(http.StatusOK, map[string]string{
        "claim_id": claimID,
        "status":   status,
    })
}
```

### Pattern 4: Signal and Query Handlers

**In Workflow**:

```go
func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    // ... workflow setup ...

    var currentStatus string = "REGISTERED"

    // Register query handler for status
    err := workflow.SetQueryHandler(ctx, "GetClaimStatus", func() (string, error) {
        return currentStatus, nil
    })
    if err != nil {
        return nil, err
    }

    // Register signal handler for document upload
    docChannel := workflow.GetSignalChannel(ctx, "documents-uploaded")
    workflow.Go(ctx, func(ctx workflow.Context) {
        for {
            var uploadSignal DocumentUploadSignal
            docChannel.Receive(ctx, &uploadSignal)
            
            // Process uploaded documents
            workflow.ExecuteActivity(ctx, ProcessDocumentsActivity, uploadSignal.DocumentIDs)
        }
    })

    // ... rest of workflow ...
}
```

### Key Principles

1. **Use Workflows for Long-Running Processes**: > 1 minute duration, human approvals, waits
2. **Activities Call Repositories**: Activities are the bridge to database/external systems
3. **Workflow Orchestrates, Activity Executes**: Workflow = coordinator, Activity = worker
4. **Signal for External Events**: Document upload, investigation completion, approvals
5. **Query for Status**: Real-time status checks without affecting workflow execution
6. **Compensation for Failures**: Rollback/cleanup on errors

---

## Workflow State Optimization

**Location**: Workflow state management in Temporal workflows

**Purpose**: Minimize database round trips by fetching data ONCE and storing in workflow state for reuse by subsequent activities.

**CRITICAL PATTERN**: When multiple activities need the same database data, fetch it ONCE in the first activity and store in workflow state. Subsequent activities read from state instead of querying database again.

### Anti-Pattern: Repeated Database Lookups

**❌ AVOID THIS** (Multiple activities query same data):

```go
func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    ctx = workflow.WithActivityOptions(ctx, activityOptions)

    // Activity 1: Validate policy (fetches policy data)
    var validationResult PolicyValidationResult
    workflow.ExecuteActivity(ctx, ValidatePolicyActivity, input.PolicyID).Get(ctx, &validationResult)

    // Activity 2: Calculate claim amount (fetches policy data AGAIN! ❌)
    var claimAmount float64
    workflow.ExecuteActivity(ctx, CalculateClaimAmountActivity, input.PolicyID).Get(ctx, &claimAmount)

    // Activity 3: Check eligibility (fetches policy data AGAIN! ❌)
    var eligible bool
    workflow.ExecuteActivity(ctx, CheckEligibilityActivity, input.PolicyID).Get(ctx, &eligible)

    // Result: Policy data fetched 3 times from database!
}
```

### ✅ Best Practice: Workflow State Pattern

**Store frequently needed data in workflow state**:

```go
// Define workflow state struct
type DeathClaimWorkflowState struct {
    // Workflow metadata
    ClaimID       string
    WorkflowStage string
    
    // Cached data (fetched once, reused multiple times)
    PolicyData    *domain.Policy        // ✅ Fetched once in first activity
    BenefitCalc   *domain.BenefitCalc   // ✅ Calculated once, reused
    ClaimantInfo  *domain.Claimant      // ✅ Fetched once
    DocumentList  []domain.Document     // ✅ Fetched once
    
    // Workflow progress tracking
    InvestigationRequired bool
    InvestigationResult   *InvestigationResult
    ApprovalResult        *ApprovalResult
}

func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    ctx = workflow.WithActivityOptions(ctx, activityOptions)

    // Initialize workflow state
    state := &DeathClaimWorkflowState{
        WorkflowStage: "INITIALIZATION",
    }

    // Activity 1: Fetch ALL required data ONCE
    var initialData InitialDataResult
    err := workflow.ExecuteActivity(ctx, FetchInitialDataActivity, input.PolicyID).Get(ctx, &initialData)
    if err != nil {
        return nil, err
    }

    // ✅ Store in workflow state
    state.PolicyData = initialData.Policy
    state.ClaimantInfo = initialData.Claimant
    state.DocumentList = initialData.Documents
    state.WorkflowStage = "DATA_LOADED"

    // Activity 2: Calculate benefits (uses state, NO database call)
    var benefitCalc domain.BenefitCalc
    err = workflow.ExecuteActivity(ctx, CalculateBenefitsActivity, state.PolicyData).Get(ctx, &benefitCalc)
    if err != nil {
        return nil, err
    }
    state.BenefitCalc = &benefitCalc // ✅ Store calculated data

    // Activity 3: Validate eligibility (uses state, NO database call)
    var eligible bool
    err = workflow.ExecuteActivity(ctx, ValidateEligibilityActivity, 
        state.PolicyData, state.ClaimantInfo).Get(ctx, &eligible)
    if err != nil {
        return nil, err
    }

    // Activity 4: Register claim (uses state data)
    var claimID string
    err = workflow.ExecuteActivity(ctx, RegisterClaimActivity, 
        state.PolicyData, state.BenefitCalc, input.DeathDate).Get(ctx, &claimID)
    if err != nil {
        return nil, err
    }
    state.ClaimID = claimID
    state.WorkflowStage = "CLAIM_REGISTERED"

    // Register query handler to expose state
    workflow.SetQueryHandler(ctx, "GetWorkflowState", func() (*DeathClaimWorkflowState, error) {
        return state, nil
    })

    // ... rest of workflow uses state ...

    return &DeathClaimWorkflowResult{
        ClaimID:     state.ClaimID,
        ClaimAmount: state.BenefitCalc.TotalAmount,
        Status:      state.WorkflowStage,
    }, nil
}
```

### Pattern 1: Initial Data Fetching Activity

**Fetch all required data in ONE activity**:

```go
type InitialDataResult struct {
    Policy    *domain.Policy
    Claimant  *domain.Claimant
    Documents []domain.Document
    Nominees  []domain.Nominee
}

// FetchInitialDataActivity fetches ALL data needed by workflow in ONE database round trip
func (a *ClaimActivities) FetchInitialDataActivity(ctx context.Context, policyID string) (*InitialDataResult, error) {
    // Use pgx.Batch to fetch all data in single round trip
    batch := &pgx.Batch{}

    // Query 1: Get policy
    query1 := dblib.Psql.Select("*").From("policies").Where(sq.Eq{"id": policyID})
    var policy domain.Policy
    dblib.QueueReturnRow(batch, query1, pgx.RowToStructByNameLax[domain.Policy], &policy)

    // Query 2: Get claimant/policyholder
    query2 := dblib.Psql.Select("*").From("policyholders").Where(sq.Eq{"policy_id": policyID})
    var claimant domain.Claimant
    dblib.QueueReturnRow(batch, query2, pgx.RowToStructByNameLax[domain.Claimant], &claimant)

    // Query 3: Get documents
    query3 := dblib.Psql.Select("*").From("documents").Where(sq.Eq{"policy_id": policyID})
    var documents []domain.Document
    dblib.QueueReturn(batch, query3, pgx.RowToStructByNameLax[domain.Document])

    // Query 4: Get nominees
    query4 := dblib.Psql.Select("*").From("nominees").Where(sq.Eq{"policy_id": policyID})
    var nominees []domain.Nominee
    dblib.QueueReturn(batch, query4, pgx.RowToStructByNameLax[domain.Nominee])

    // Execute all queries in single round trip
    err := a.db.SendBatch(ctx, batch).Close()
    if err != nil {
        return nil, err
    }

    return &InitialDataResult{
        Policy:    &policy,
        Claimant:  &claimant,
        Documents: documents,
        Nominees:  nominees,
    }, nil
}
```

### Pattern 2: Subsequent Activities Use State (No DB Calls)

```go
// CalculateBenefitsActivity - Takes policy from state, NO database call
func (a *ClaimActivities) CalculateBenefitsActivity(ctx context.Context, policy *domain.Policy) (*domain.BenefitCalc, error) {
    // ✅ Pure calculation using data from workflow state
    // ❌ NO database queries here!
    
    totalAmount := policy.SumAssured + policy.AccumulatedBonus - policy.LoanOutstanding
    
    return &domain.BenefitCalc{
        BaseAmount:        policy.SumAssured,
        BonusAmount:       policy.AccumulatedBonus,
        LoanDeduction:     policy.LoanOutstanding,
        TotalAmount:       totalAmount,
        CalculatedAt:      time.Now(),
    }, nil
}

// ValidateEligibilityActivity - Takes data from state, NO database call
func (a *ClaimActivities) ValidateEligibilityActivity(ctx context.Context, 
    policy *domain.Policy, claimant *domain.Claimant) (bool, error) {
    
    // ✅ Pure business logic using state data
    // ❌ NO database queries!
    
    if policy.Status != "ACTIVE" {
        return false, fmt.Errorf("policy not active")
    }
    
    if claimant.KYCStatus != "VERIFIED" {
        return false, fmt.Errorf("KYC not verified")
    }
    
    return true, nil
}
```

### Pattern 3: Update State as Workflow Progresses

```go
func MaturityClaimWorkflow(ctx workflow.Context, input MaturityClaimInput) (*MaturityClaimResult, error) {
    // Initialize state
    state := &MaturityClaimWorkflowState{
        WorkflowStage: "INITIALIZED",
    }

    // Fetch initial data
    var initialData InitialDataResult
    workflow.ExecuteActivity(ctx, FetchInitialDataActivity, input.PolicyID).Get(ctx, &initialData)
    state.PolicyData = initialData.Policy
    state.WorkflowStage = "DATA_LOADED"

    // Calculate maturity amount
    var maturityCalc domain.MaturityCalculation
    workflow.ExecuteActivity(ctx, CalculateMaturityAmountActivity, state.PolicyData).Get(ctx, &maturityCalc)
    state.MaturityCalc = &maturityCalc // ✅ Store calculation in state
    state.WorkflowStage = "AMOUNT_CALCULATED"

    // Verify bank account
    var bankVerification BankVerificationResult
    workflow.ExecuteActivity(ctx, VerifyBankAccountActivity, state.PolicyData.BankAccount).Get(ctx, &bankVerification)
    state.BankVerified = bankVerification.Verified // ✅ Store verification result
    state.WorkflowStage = "BANK_VERIFIED"

    // Register claim (uses all state data)
    var claimID string
    workflow.ExecuteActivity(ctx, RegisterMaturityClaimActivity, state).Get(ctx, &claimID)
    state.ClaimID = claimID
    state.WorkflowStage = "CLAIM_REGISTERED"

    // Wait for approval signal
    approvalChannel := workflow.GetSignalChannel(ctx, "claim-approved")
    var approvalSignal ApprovalSignal
    approvalChannel.Receive(ctx, &approvalSignal)
    state.ApprovalResult = &approvalSignal // ✅ Store approval in state
    state.WorkflowStage = "APPROVED"

    // Process payment (uses state)
    var paymentResult PaymentResult
    workflow.ExecuteActivity(ctx, ProcessPaymentActivity, state.ClaimID, state.MaturityCalc.TotalAmount).Get(ctx, &paymentResult)
    state.PaymentResult = &paymentResult
    state.WorkflowStage = "COMPLETED"

    return &MaturityClaimResult{
        ClaimID: state.ClaimID,
        Amount:  state.MaturityCalc.TotalAmount,
        Status:  state.WorkflowStage,
    }, nil
}
```

### Pattern 4: Query Handler for State Inspection

```go
func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    state := &DeathClaimWorkflowState{}

    // Register query handler to expose current state
    workflow.SetQueryHandler(ctx, "GetWorkflowState", func() (*DeathClaimWorkflowState, error) {
        return state, nil
    })

    // Register specialized queries
    workflow.SetQueryHandler(ctx, "GetClaimAmount", func() (float64, error) {
        if state.BenefitCalc == nil {
            return 0, fmt.Errorf("claim amount not calculated yet")
        }
        return state.BenefitCalc.TotalAmount, nil
    })

    workflow.SetQueryHandler(ctx, "GetCurrentStage", func() (string, error) {
        return state.WorkflowStage, nil
    })

    // ... workflow execution ...
}
```

### Pattern 5: State-Based Error Recovery

```go
func DeathClaimWorkflow(ctx workflow.Context, input DeathClaimWorkflowInput) (*DeathClaimWorkflowResult, error) {
    state := &DeathClaimWorkflowState{
        RetryCount: 0,
    }

    // Fetch initial data with retry
    var initialData InitialDataResult
    for attempt := 0; attempt < 3; attempt++ {
        err := workflow.ExecuteActivity(ctx, FetchInitialDataActivity, input.PolicyID).Get(ctx, &initialData)
        if err == nil {
            state.PolicyData = initialData.Policy
            break
        }
        
        state.RetryCount++
        if attempt == 2 {
            return nil, fmt.Errorf("failed to fetch initial data after 3 attempts")
        }
        
        // Wait before retry
        workflow.Sleep(ctx, time.Duration(attempt+1)*time.Minute)
    }

    // Process payment with state-based compensation
    var paymentResult PaymentResult
    err := workflow.ExecuteActivity(ctx, ProcessPaymentActivity, 
        state.ClaimID, state.BenefitCalc.TotalAmount).Get(ctx, &paymentResult)
    
    if err != nil {
        // Compensation: Use state to rollback
        workflow.ExecuteActivity(ctx, RollbackClaimActivity, state.ClaimID, state.WorkflowStage)
        return nil, err
    }

    state.PaymentResult = &paymentResult

    return &DeathClaimWorkflowResult{
        ClaimID: state.ClaimID,
        Status:  "COMPLETED",
    }, nil
}
```

### Key Principles

1. **Fetch Data ONCE**: Use initial data fetching activity to get all required data
2. **Store in State**: Keep fetched data in workflow state struct
3. **Pass State to Activities**: Activities receive data as parameters, not from DB
4. **Update State Progressively**: Add results to state as workflow proceeds
5. **Expose State via Queries**: Allow external systems to inspect current state
6. **Use pgx.Batch in Initial Fetch**: Minimize round trips even in initial activity
7. **State for Compensation**: Store intermediate states for rollback logic

### Benefits

- **Reduced Database Load**: Data fetched once instead of N times
- **Faster Execution**: No repeated queries in activities
- **Better Auditability**: Complete workflow state available for inspection
- **Easier Testing**: Activities become pure functions when they use state
- **Simplified Compensation**: State provides context for rollback operations

---

## Summary: When to Use Each Pattern

| Scenario | Pattern | Example |
|----------|---------|---------|
| Single query | Direct db.QueryRow/Exec | Get policy by ID |
| 2+ queries in transaction | pgx.Batch | Create order + update inventory |
| Insert derived data | INSERT...SELECT | Generate maturity claims |
| Update based on query | UPDATE...FROM | Update claims from approvals |
| Complex multi-step | WITH (CTE) | Create claim + update policy + audit |
| Long-running process | Temporal Workflow | Death claim with investigation |
| Multi-step with waits | Temporal Workflow | Claim waiting for approval |
| Multiple activities need same data | Workflow State | Store policy data once, reuse |
| External API calls | Temporal Activity | Bank verification API |
| Human approval needed | Temporal Signal | Officer approves claim |

---


---

## Naming Conventions

### Package Names
- `domain` - Business entities
- `handler` - HTTP handlers
- `response` - Response DTOs (subpackage of handler)
- `repo` - Repository interfaces
- `postgres` - PostgreSQL implementations (subpackage of repo)

### Type Names
- Domain: `{Resource}` (e.g., `User`, `Product`)
- Repository: `{Resource}Repository` (e.g., `UserRepository`)
- Handler: `{Resource}Handler` (e.g., `UserHandler`)
- Request: `Create{Resource}Request`, `Update{Resource}Request`, `{Resource}IDUri`, `List{Resources}Params`
- Response: `{Resource}Response`, `{Resource}CreateResponse`, `{Resources}ListResponse`

### Function Names
- Constructor: `New{Resource}Repository`, `New{Resource}Handler`
- Handler methods: `Create{Resource}`, `List{Resources}`, `Get{Resource}ByID`, `Update{Resource}ByID`, `Delete{Resource}ByID`
- Repository methods: `Create`, `FindByID`, `List`, `Update`, `Delete`
- Response converter: `New{Resource}Response`, `New{Resources}Response`

### Field Names
- Go: `PascalCase` (e.g., `FirstName`)
- JSON: `snake_case` (e.g., `first_name`)
- Database: `snake_case` (e.g., `first_name`)
- URL params: `snake_case` (e.g., `:id`, `:user_id`)

### Route Names
- Paths: `/{resources}` (plural, lowercase)
- Route names: `"Create {Resource}"`, `"List {Resources}"` (for Swagger)

---

## Error Handling

**In Handlers**:
```go
result, err := h.svc.SomeMethod(sctx.Ctx, params)
if err != nil {
    if err == pgx.ErrNoRows {
        sctx.Log.Error("resource not found", "id", id)
        return nil, err
    }
    sctx.Log.Error("failed to perform operation", "error", err)
    return nil, err
}
```

**In Repositories**:
```go
err := dblib.SelectOne(ctx, r.db, query, &result)
if err != nil {
    if err == pgx.ErrNoRows {
        return result, err // Let handler decide how to handle
    }
    return result, err
}
return result, nil
```

**Rules**:
- Always log errors before returning
- Use descriptive log messages
- Include relevant context in logs (IDs, parameters)
- Return errors directly (framework handles HTTP status codes)
- Check for `pgx.ErrNoRows` for 404 scenarios
- Don't wrap errors unnecessarily

---

## Complete Example Workflow

When creating a new resource called `Product`, follow these steps:

### Step 1: Create Domain Model
**File**: `core/domain/product.go`
```go
package domain

import "time"

type Product struct {
    ID          int64     `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Price       float64   `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

### Step 2: Create Database Schema
**File**: `db/products.sql`
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
```

### Step 3: Create Repository
**File**: `repo/postgres/product.go`
```go
package repo

import (
    "context"
    "time"

    sq "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
    config "gitlab.cept.gov.in/it-2.0-common/api-config"
    dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
    "pisapi/core/domain"
)

type ProductRepository struct {
    db  *dblib.DB
    cfg *config.Config
}

func NewProductRepository(db *dblib.DB, cfg *config.Config) *ProductRepository {
    return &ProductRepository{
        db:  db,
        cfg: cfg,
    }
}

const productTable = "products"

func (r *ProductRepository) Create(ctx context.Context, data domain.Product) (domain.Product, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Insert(productTable).
        Columns("name", "description", "price", "stock").
        Values(data.Name, data.Description, data.Price, data.Stock).
        Suffix("RETURNING id, name, description, price, stock, created_at, updated_at").
        PlaceholderFormat(sq.Dollar)

    var result domain.Product
    err := dblib.Insert(ctx, r.db, query, &result)
    return result, err
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (domain.Product, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Select("id", "name", "description", "price", "stock", "created_at", "updated_at").
        From(productTable).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    var result domain.Product
    err := dblib.SelectOne(ctx, r.db, query, &result)
    if err != nil {
        return result, err
    }
    return result, nil
}

func (r *ProductRepository) List(ctx context.Context, skip, limit int64, orderBy, sortType string) ([]domain.Product, int64, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
    defer cancel()

    countQuery := sq.Select("COUNT(*)").
        From(productTable).
        PlaceholderFormat(sq.Dollar)

    var totalCount int64
    err := dblib.SelectOne(ctx, r.db, countQuery, &totalCount)
    if err != nil {
        return nil, 0, err
    }

    query := sq.Select("id", "name", "description", "price", "stock", "created_at", "updated_at").
        From(productTable).
        OrderBy(orderBy + " " + sortType).
        Limit(uint64(limit)).
        Offset(uint64(skip)).
        PlaceholderFormat(sq.Dollar)

    var results []domain.Product
    err = dblib.SelectRows(ctx, r.db, query, &results)
    if err != nil {
        return nil, 0, err
    }

    return results, totalCount, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int64, name, description *string, price *float64, stock *int) (domain.Product, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Update(productTable).
        Set("updated_at", time.Now()).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    if name != nil {
        query = query.Set("name", *name)
    }
    if description != nil {
        query = query.Set("description", *description)
    }
    if price != nil {
        query = query.Set("price", *price)
    }
    if stock != nil {
        query = query.Set("stock", *stock)
    }

    query = query.Suffix("RETURNING id, name, description, price, stock, created_at, updated_at")

    var result domain.Product
    err := dblib.Update(ctx, r.db, query, &result)
    return result, err
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()

    query := sq.Delete(productTable).
        Where(sq.Eq{"id": id}).
        PlaceholderFormat(sq.Dollar)

    return dblib.Delete(ctx, r.db, query)
}
```

### Step 4: Create Request DTOs
**File**: `handler/request.go` (add to existing file)
```go
import "pisapi/core/port"

type CreateProductRequest struct {
    Name        string  `json:"name" validate:"required"`
    Description string  `json:"description" validate:"required"`
    Price       float64 `json:"price" validate:"required"`
    Stock       int     `json:"stock" validate:"required"`
}

func (r CreateProductRequest) ToDomain() domain.Product {
    return domain.Product{
        Name:        r.Name,
        Description: r.Description,
        Price:       r.Price,
        Stock:       r.Stock,
    }
}

type UpdateProductRequest struct {
    ID          int64   `uri:"id" validate:"required"`
    Name        string  `json:"name" validate:"omitempty"`
    Description string  `json:"description" validate:"omitempty"`
    Price       float64 `json:"price" validate:"omitempty"`
    Stock       int     `json:"stock" validate:"omitempty"`
}

type ProductIDUri struct {
    ID int64 `uri:"id" validate:"required"`
}

type ListProductsParams struct {
    port.MetadataRequest
}
```

### Step 5: Create Response DTOs
**File**: `handler/response/product.go`
```go
package response

import (
    "pisapi/core/domain"
    "pisapi/core/port"
)

type ProductResponse struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Stock       int     `json:"stock"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

func NewProductResponse(d domain.Product) ProductResponse {
    return ProductResponse{
        ID:          d.ID,
        Name:        d.Name,
        Description: d.Description,
        Price:       d.Price,
        Stock:       d.Stock,
        CreatedAt:   d.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:   d.UpdatedAt.Format("2006-01-02 15:04:05"),
    }
}

func NewProductsResponse(data []domain.Product) []ProductResponse {
    res := make([]ProductResponse, 0, len(data))
    for _, d := range data {
        res = append(res, NewProductResponse(d))
    }
    return res
}

type ProductCreateResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      ProductResponse `json:"data"`
}

type ProductFetchResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      ProductResponse `json:"data"`
}

type ProductsListResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    port.MetaDataResponse     `json:",inline"`
    Data                      []ProductResponse `json:"data"`
}

type ProductUpdateResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
    Data                      ProductResponse `json:"data"`
}

type ProductDeleteResponse struct {
    port.StatusCodeAndMessage `json:",inline"`
}
```

### Step 6: Create Handler
**File**: `handler/product.go`
```go
package handler

import (
    "github.com/jackc/pgx/v5"
    log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
    serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
    serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
    "pisapi/core/port"
    resp "pisapi/handler/response"
    repo "pisapi/repo/postgres"
)

type ProductHandler struct {
    *serverHandler.Base
    svc *repo.ProductRepository
}

func NewProductHandler(svc *repo.ProductRepository) *ProductHandler {
    base := serverHandler.New("Products").
        SetPrefix("/v1").
        AddPrefix("")
    return &ProductHandler{
        Base: base,
        svc:  svc,
    }
}

func (h *ProductHandler) Routes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.POST("/products", h.CreateProduct).Name("Create Product"),
        serverRoute.GET("/products", h.ListProducts).Name("List Products"),
        serverRoute.GET("/products/:id", h.GetProductByID).Name("Get Product By ID"),
        serverRoute.PUT("/products/:id", h.UpdateProductByID).Name("Update Product By ID"),
        serverRoute.DELETE("/products/:id", h.DeleteProductByID).Name("Delete Product By ID"),
    }
}

func (h *ProductHandler) CreateProduct(sctx *serverRoute.Context, req CreateProductRequest) (*resp.ProductCreateResponse, error) {
    data := req.ToDomain()

    result, err := h.svc.Create(sctx.Ctx, data)
    if err != nil {
        log.Error(sctx.Ctx, "Error creating product: %v", err)
        return nil, err
    }

    log.Info(sctx.Ctx, "Product created with ID: %d", result.ID)
    r := &resp.ProductCreateResponse{
        StatusCodeAndMessage: port.CreateSuccess,
        Data:                 resp.NewProductResponse(result),
    }
    return r, nil
}

func (h *ProductHandler) ListProducts(sctx *serverRoute.Context, req ListProductsParams) (*resp.ProductsListResponse, error) {
    results, totalCount, err := h.svc.List(sctx.Ctx, req.Skip, req.Limit, req.OrderBy, req.SortType)
    if err != nil {
        log.Error(sctx.Ctx, "Error fetching products: %v", err)
        return nil, err
    }

    r := &resp.ProductsListResponse{
        StatusCodeAndMessage: port.ListSuccess,
        MetaDataResponse: port.MetaDataResponse{
            TotalCount: totalCount,
            Count:      int64(len(results)),
            Skip:       req.Skip,
            Limit:      req.Limit,
        },
        Data: resp.NewProductsResponse(results),
    }
    return r, nil
}

func (h *ProductHandler) GetProductByID(sctx *serverRoute.Context, req ProductIDUri) (*resp.ProductFetchResponse, error) {
    result, err := h.svc.FindByID(sctx.Ctx, req.ID)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Error(sctx.Ctx, "Product not found with ID: %d", req.ID)
            return nil, err
        }
        log.Error(sctx.Ctx, "Error fetching product by ID: %v", err)
        return nil, err
    }

    r := &resp.ProductFetchResponse{
        StatusCodeAndMessage: port.FetchSuccess,
        Data:                 resp.NewProductResponse(result),
    }
    return r, nil
}

func (h *ProductHandler) UpdateProductByID(sctx *serverRoute.Context, req UpdateProductRequest) (*resp.ProductUpdateResponse, error) {
    var name, description *string
    var price *float64
    var stock *int

    if req.Name != "" {
        name = &req.Name
    }
    if req.Description != "" {
        description = &req.Description
    }
    if req.Price != 0 {
        price = &req.Price
    }
    if req.Stock != 0 {
        stock = &req.Stock
    }

    result, err := h.svc.Update(sctx.Ctx, req.ID, name, description, price, stock)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Error(sctx.Ctx, "Product not found with ID: %d", req.ID)
            return nil, err
        }
        log.Error(sctx.Ctx, "Error updating product by ID: %v", err)
        return nil, err
    }

    r := &resp.ProductUpdateResponse{
        StatusCodeAndMessage: port.UpdateSuccess,
        Data:                 resp.NewProductResponse(result),
    }
    return r, nil
}

func (h *ProductHandler) DeleteProductByID(sctx *serverRoute.Context, req ProductIDUri) (*resp.ProductDeleteResponse, error) {
    err := h.svc.Delete(sctx.Ctx, req.ID)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Error(sctx.Ctx, "Product not found with ID: %d", req.ID)
            return nil, err
        }
        log.Error(sctx.Ctx, "Error deleting product by ID: %v", err)
        return nil, err
    }

    r := &resp.ProductDeleteResponse{
        StatusCodeAndMessage: port.DeleteSuccess,
    }
    return r, nil
}
```

### Step 7: Register Dependencies
**File**: `bootstrap/bootstrapper.go`
```go
var FxRepo = fx.Module(
    "Repomodule",
    fx.Provide(
        repo.NewUserRepository,
        repo.NewProductRepository, // Add this line
    ),
)

var FxHandler = fx.Module(
    "Handlermodule",
    fx.Provide(
        fx.Annotate(
            handler.NewUserHandler,
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
        fx.Annotate(
            handler.NewProductHandler, // Add this block
            fx.As(new(serverHandler.Handler)),
            fx.ResultTags(serverHandler.ServerControllersGroupTag),
        ),
    ),
)
```

### Step 8: Generate Validators
```bash
cd handler
govalid
```

### Step 9: Run Migrations
```bash
# Apply database schema
psql -U username -d database -f db/products.sql
```

### Step 10: Test Endpoints
```bash
# Start the server
go run main.go

# Test endpoints
# Create
curl -X POST http://localhost:8080/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Product 1","description":"Description","price":99.99,"stock":100}'

# List
curl http://localhost:8080/v1/products

# Get by ID
curl http://localhost:8080/v1/products/1

# Update
curl -X PUT http://localhost:8080/v1/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Product"}'

# Delete
curl -X DELETE http://localhost:8080/v1/products/1
```

---

## Quick Reference Checklist

When creating a new API resource, ensure you:

- [ ] Create domain model in `core/domain/{resource}.go`
- [ ] Create database schema in `db/{resources}.sql`
- [ ] Create repository in `repo/postgres/{resource}.go` with:
  - [ ] `Create` method
  - [ ] `FindByID` method
  - [ ] `List` method
  - [ ] `Update` method
  - [ ] `Delete` method
- [ ] Add request DTOs to `handler/request.go`:
  - [ ] `Create{Resource}Request`
  - [ ] `Update{Resource}Request`
  - [ ] `{Resource}IDUri`
  - [ ] `List{Resources}Params`
- [ ] Create response DTOs in `handler/response/{resource}.go`:
  - [ ] `{Resource}Response`
  - [ ] Conversion functions
  - [ ] Operation-specific responses (Create, Fetch, List, Update, Delete)
- [ ] Create handler in `handler/{resource}.go` with:
  - [ ] Constructor
  - [ ] `Routes()` method
  - [ ] CRUD handler methods
- [ ] Register in `bootstrap/bootstrapper.go`:
  - [ ] Add repository to `FxRepo`
  - [ ] Add handler to `FxHandler`
- [ ] Generate validators with `govalid`
- [ ] Run database migrations
- [ ] Test all endpoints

---

## Common Patterns Summary

**REST Operations**:
- POST /{resources} - Create
- GET /{resources} - List (with pagination)
- GET /{resources}/:id - Get single
- PUT /{resources}/:id - Update
- DELETE /{resources}/:id - Delete

**Response Codes**:
- 200 OK - Successful GET, PUT, DELETE
- 201 Created - Successful POST
- 400 Bad Request - Validation errors
- 404 Not Found - Resource not found
- 500 Internal Server Error - Server errors

**Standard Response Format**:
```json
{
  "status_code": 200,
  "success": true,
  "message": "operation successful",
  "data": {...}
}
```

**List Response Format**:
```json
{
  "status_code": 200,
  "success": true,
  "message": "list retrieved successfully",
  "total_count": 100,
  "count": 10,
  "skip": 0,
  "limit": 10,
  "data": [...]
}
```

---

## Notes

- Replace `{Resource}` with your actual resource name (e.g., `Product`, `Order`)
- Replace `{resources}` with plural lowercase (e.g., `products`, `orders`)
- Replace `{project}` with your actual project module name
- All timestamps are stored in UTC
- Pagination defaults: Skip=0, Limit=10, OrderBy="id", SortType="asc"
- Query timeouts prevent long-running queries from blocking
- Use Squirrel for all SQL query building (type-safe, composable)
- All database operations must use context with timeout
- Framework handles request binding, validation, and response serialization automatically

---

## Development Workflow

### Initial Project Setup

```bash
# 1. Create project directory
mkdir {project-name}
cd {project-name}

# 2. Initialize Go module
go mod init {project}

# 3. Create directory structure
mkdir -p bootstrap configs core/domain core/port handler/response repo/postgres db docs

# 4. Install core dependencies
go get gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper@latest
go get gitlab.cept.gov.in/it-2.0-common/n-api-server@latest
go get gitlab.cept.gov.in/it-2.0-common/n-api-db@latest
go get gitlab.cept.gov.in/it-2.0-common/api-config@latest
go get github.com/Masterminds/squirrel@latest
go get github.com/jackc/pgx/v5@latest
go get go.uber.org/fx@latest

# 5. Tidy up dependencies
go mod tidy

# 6. Create config files (copy from template)
cp path/to/template/configs/* configs/

# 7. Create main.go and bootstrap files
# (Follow patterns in this document)
```

### Adding a New Resource

**Step-by-Step Checklist**:

1. **Create Domain Model** (`core/domain/{resource}.go`)
   ```bash
   # Create the file and add domain struct with db tags
   ```

2. **Create Database Schema** (`db/{resources}.sql`)
   ```bash
   # Write CREATE TABLE statement with indexes
   ```

3. **Apply Database Migration**
   ```bash
   psql -U username -d database -f db/{resources}.sql
   ```

4. **Create Repository** (`repo/postgres/{resource}.go`)
   ```bash
   # Implement Create, FindByID, List, Update, Delete methods
   ```

5. **Add Request DTOs** (`handler/request.go`)
   ```bash
   # Add Create, Update, ID, and List request structs
   ```

6. **Create Response DTOs** (`handler/response/{resource}.go`)
   ```bash
   # Add response structs and conversion functions
   ```

7. **Create Handler** (`handler/{resource}.go`)
   ```bash
   # Implement handler with all CRUD methods
   ```

8. **Register Dependencies** (`bootstrap/bootstrapper.go`)
   ```bash
   # Add repository to FxRepo
   # Add handler to FxHandler
   ```

9. **Generate Validators**
   ```bash
   cd handler
   govalid
   cd ..
   ```

10. **Test Endpoints**
    ```bash
    # Start server and test with curl or Postman
    ```

### Running the Application

```bash
# Development (uses config.yaml or config.dev.yaml)
go run main.go

# Specify environment
ENV=dev go run main.go      # Development
ENV=test go run main.go     # Test
ENV=sit go run main.go      # System Integration Test
ENV=staging go run main.go  # Staging
ENV=prod go run main.go     # Production

# Build binary
go build -o bin/app main.go

# Run binary
./bin/app

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o bin/app main.go
```

### Database Operations

```bash
# Connect to database
psql -U username -d database

# Run migration
psql -U username -d database -f db/{resource}.sql

# Check tables
\dt

# Describe table
\d {resources}

# Query data
SELECT * FROM {resources};

# Drop table (careful!)
DROP TABLE IF EXISTS {resources};
```

### Testing Endpoints

```bash
# Create
curl -X POST http://localhost:8080/v1/{resources} \
  -H "Content-Type: application/json" \
  -d '{
    "field1": "value1",
    "field2": "value2",
    "field3": 123
  }'

# List with pagination
curl "http://localhost:8080/v1/{resources}?skip=0&limit=10&orderBy=id&sortType=asc"

# Get by ID
curl http://localhost:8080/v1/{resources}/1

# Update
curl -X PUT http://localhost:8080/v1/{resources}/1 \
  -H "Content-Type: application/json" \
  -d '{
    "field1": "updated value"
  }'

# Delete
curl -X DELETE http://localhost:8080/v1/{resources}/1

# Check response status
curl -i http://localhost:8080/v1/{resources}
```

### Common Development Tasks

**Update Dependencies**:
```bash
# Update specific package
go get gitlab.cept.gov.in/it-2.0-common/n-api-server@latest

# Update all dependencies
go get -u ./...

# Tidy up
go mod tidy
```

**Generate Validators**:
```bash
# Install govalid (once)
go install github.com/twpayne/go-govalid/cmd/govalid@latest

# Generate validators
cd handler
govalid
```

**Format Code**:
```bash
# Format all files
go fmt ./...

# Or use gofmt
gofmt -w .
```

**Lint Code**:
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

**Run Tests**:
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in specific package
go test ./handler/...

# Verbose output
go test -v ./...
```

### Debugging Tips

**Enable Debug Logging**:
```yaml
# In config.yaml
log:
  level: debug  # Change from info to debug
```

**Check Database Connections**:
```bash
# In psql, check active connections
SELECT * FROM pg_stat_activity WHERE datname = 'your_database';
```

**View Server Logs**:
```bash
# Logs are output to stdout by default
# Redirect to file:
go run main.go > app.log 2>&1
```

**Common Issues**:

1. **Port already in use**:
   ```bash
   # Find process using port
   netstat -ano | findstr :8080  # Windows
   lsof -i :8080                 # Linux/Mac
   
   # Kill process
   taskkill /PID <pid> /F        # Windows
   kill -9 <pid>                 # Linux/Mac
   ```

2. **Database connection failed**:
   - Check config.yaml database credentials
   - Ensure PostgreSQL is running
   - Check network connectivity
   - Verify database exists

3. **Validation errors**:
   - Regenerate validators: `cd handler && govalid`
   - Check validation tags in request structs
   - Ensure all required fields are provided

4. **Import errors**:
   - Run `go mod tidy`
   - Check module path in go.mod
   - Verify all imports use correct module paths

### Production Deployment

**Build for Production**:
```bash
# Build with optimizations
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o app main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o app.exe main.go
```

**Environment Variables**:
```bash
# Set environment
export ENV=prod

# Set database password
export DB_PASSWORD=secret

# Set Redis password
export REDIS_PASSWORD=secret
```

**Docker Deployment** (if using Docker):
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/main main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./main"]
```

**Health Check Endpoint**:
```bash
# Check if server is running
curl http://localhost:8080/health

# Or use built-in health endpoint (if available)
curl http://localhost:8080/v1/health
```

### Version Control

**Git Workflow**:
```bash
# Create feature branch
git checkout -b feature/{resource-name}

# Add changes
git add .

# Commit with meaningful message
git commit -m "feat: add {resource} CRUD endpoints"

# Push to remote
git push origin feature/{resource-name}

# After review, merge to main
```

**Commit Message Convention**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Add/update tests
- `chore:` - Maintenance tasks

### Monitoring and Observability

**Tracing** (if enabled):
```yaml
# In config.yaml
trace:
  enabled: true
  processor:
    type: "otlp-grpc"
    options:
      host: "localhost:4317"
```

**Metrics**:
- Database connection pool metrics
- Request latency
- Error rates
- Active requests

**Logs**:
- Structured logging with context
- Error logging with stack traces
- Request/response logging

---
