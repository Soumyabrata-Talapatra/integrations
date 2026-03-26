# DB Library

**API DB** is a Go package that provides an abstraction layer for interacting with postgreSQL databases. It simplifies database connection management, transaction handling, and configuration using `pgxpool`.


## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Examples](#examples)
- [Notes](#notes)

## Features

- **Connection Pooling**: Efficient resource management with `pgxpool`.
- **Transactional Operations**: Easy-to-use functions for read and write transactions.
- **Customizable Isolation Levels**: Support for different isolation levels for transactions.
- **Configurable Connection Settings**: Flexible options for database connection setup.
- **Graceful Cleanup**: Ensures database connections are closed properly.


## Installation

Add the `db` library to your Go project:

```bash
go get gitlab.cept.gov.in/it-2.0-common/n-api-db
```

Import the required package in your Go files:

```go
import (
dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)
```

## Configuration

The library uses a configuration struct (`DBConfig`) to set up database connection parameters. Below are the fields available in the configuration:

### DBConfig Fields

- `DBUsername`: Database username.
- `DBPassword`: Database password.
- `DBHost`: Host address of the database server.
- `DBPort`: Port of the database server.
- `DBDatabase`: Name of the database.
- `Schema`: Database schema.
- `MaxConns`: Maximum number of connections in the pool.
- `MinConns`: Minimum number of connections in the pool.
- `MaxConnLifetime`: Maximum connection lifetime in minutes.
- `MaxConnIdleTime`: Maximum idle time for connections in minutes.
- `HealthCheckPeriod`: Period for health checks in minutes.
- `AppName`: Name of the application for database logging purposes.

### Setting Up Configuration

```go
 dbConfig := dblib.DBConfig{ 
        DBUsername:        c.DBUsername(),
        DBPassword:        c.DBPassword(),
        DBHost:            c.DBHost(),
        DBPort:            c.DBPort(),
        DBDatabase:        c.DBDatabase(),
        Schema:            c.DBSchema(),
        MaxConns:          int32(c.MaxConns()),
        MinConns:          int32(c.MinConns()),
        MaxConnLifetime:   time.Duration(c.MaxConnLifetime()),   // In minutes
        MaxConnIdleTime:   time.Duration(c.MaxConnIdleTime()),   // In minutes
        HealthCheckPeriod: time.Duration(c.HealthCheckPeriod()), // In minutes
        AppName:           c.AppName(),
    }
```
## Usage

### Initialize Database Connection

Use the `DefaultDbFactory` to prepare the configuration and establish a connection:

```go
 preparedConfig := dblib.NewDefaultDbFactory().NewPreparedDBConfig(dbConfig)

    // Step 3: Establish the database connection
    dbConn, err := dblib.NewDefaultDbFactory().CreateConnection(preparedConfig)

    if err != nil {
        log.Warn(nil,"error in db connection %s", err)
        return nil, err
    }
    log.Info(nil,"Successfully connected to the database %s", c.DBConnection())
    defer dbConn.Close()

    if dbConn.Ping() == nil {
		log.info("Connection Established")
	} else {
		log.warn("Failed to establish database connection")
	}
```
## API Reference

### Core Database Operations

#### Query Helpers

##### `SelectOne[T any](ctx context.Context, db *DB, builder sq.SelectBuilder, scanFn pgx.RowToFunc[T]) (T, error)`
Executes a SELECT query and returns a single row. Returns `pgx.ErrNoRows` if no row is found.

```go
user, err := dblib.SelectOne(ctx, db, 
    dblib.Psql.Select("id", "name", "email").From("users").Where(sq.Eq{"id": 1}),
    pgx.RowToStructByName[User])
```

##### `SelectOneOK[T any](ctx context.Context, db *DB, builder sq.SelectBuilder, scanFn pgx.RowToFunc[T]) (T, bool, error)`
Similar to SelectOne but returns a boolean indicating if a row was found instead of returning an error for no rows.

```go
user, found, err := dblib.SelectOneOK(ctx, db, 
    dblib.Psql.Select("*").From("users").Where(sq.Eq{"id": 1}),
    pgx.RowToStructByName[User])
```

##### `SelectRows[T any](ctx context.Context, db *DB, builder sq.SelectBuilder, scanFn pgx.RowToFunc[T]) ([]T, error)`
Executes a SELECT query and returns multiple rows.

```go
users, err := dblib.SelectRows(ctx, db,
    dblib.Psql.Select("*").From("users").Where(sq.Eq{"active": true}),
    pgx.RowToStructByName[User])
```

##### `SelectRowsOK[T any](ctx context.Context, db *DB, builder sq.SelectBuilder, scanFn pgx.RowToFunc[T]) ([]T, bool, error)`
Similar to SelectRows but returns a boolean indicating if any rows were found.

```go
users, found, err := dblib.SelectRowsOK(ctx, db,
    dblib.Psql.Select("*").From("users"),
    pgx.RowToStructByName[User])
```

##### `SelectRowsTag[T any](ctx context.Context, db *DB, builder sq.SelectBuilder, tag string) ([]T, error)`
Executes a SELECT query and maps results to struct fields using custom struct tags.

```go
// For struct with custom tags like `custom:"column_name"`
users, err := dblib.SelectRowsTag[User](ctx, db,
    dblib.Psql.Select("*").From("users"),
    "custom")
```

#### Insert Operations

##### `Insert(ctx context.Context, db *DB, query sq.InsertBuilder) (pgconn.CommandTag, error)`
Executes an INSERT statement and returns command tag with rows affected.

```go
tag, err := dblib.Insert(ctx, db,
    dblib.Psql.Insert("users").Columns("name", "email").Values("John", "john@example.com"))
```

##### `InsertReturning[T any](ctx context.Context, db *DB, builder sq.InsertBuilder, scanFn pgx.RowToFunc[T]) (T, error)`
Executes an INSERT statement with RETURNING clause and returns the inserted row.

```go
user, err := dblib.InsertReturning(ctx, db,
    dblib.Psql.Insert("users").Columns("name", "email").
        Values("John", "john@example.com").
        Suffix("RETURNING *"),
    pgx.RowToStructByName[User])
```

##### `InsertReturningrows[T any](ctx context.Context, db *DB, builder sq.InsertBuilder, scanFn pgx.RowToFunc[T]) ([]T, error)`
Executes an INSERT statement with RETURNING clause and returns multiple inserted rows (useful for bulk inserts).

```go
users, err := dblib.InsertReturningrows(ctx, db,
    dblib.Psql.Insert("users").Columns("name", "email").
        Values("John", "john@example.com").
        Values("Jane", "jane@example.com").
        Suffix("RETURNING *"),
    pgx.RowToStructByName[User])
```

#### Update Operations

##### `Update(ctx context.Context, db *DB, query sq.UpdateBuilder) (pgconn.CommandTag, error)`
Executes an UPDATE statement.

```go
tag, err := dblib.Update(ctx, db,
    dblib.Psql.Update("users").Set("email", "newemail@example.com").Where(sq.Eq{"id": 1}))
```

##### `UpdateReturning[T any](ctx context.Context, db *DB, query sq.UpdateBuilder, scanFn pgx.RowToFunc[T]) (T, error)`
Executes an UPDATE statement with RETURNING clause.

```go
user, err := dblib.UpdateReturning(ctx, db,
    dblib.Psql.Update("users").Set("email", "new@example.com").
        Where(sq.Eq{"id": 1}).
        Suffix("RETURNING *"),
    pgx.RowToStructByName[User])
```

#### Delete Operations

##### `Delete(ctx context.Context, db *DB, query sq.DeleteBuilder) (pgconn.CommandTag, error)`
Executes a DELETE statement.

```go
tag, err := dblib.Delete(ctx, db,
    dblib.Psql.Delete("users").Where(sq.Eq{"id": 1}))
```

#### General Execution

##### `Exec(ctx context.Context, db *DB, sql string, args []any) (pgconn.CommandTag, error)`
Executes a raw SQL query.

```go
tag, err := dblib.Exec(ctx, db, "DELETE FROM users WHERE created_at < $1", oldDate)
```

##### `ExecRow(ctx context.Context, db *DB, sql string, args ...any) (pgconn.CommandTag, error)`
Executes a SQL statement and ensures at least one row was affected. Returns `pgx.ErrNoRows` if no rows were affected.

```go
tag, err := dblib.ExecRow(ctx, db, "UPDATE users SET active = true WHERE id = $1", userId)
```

### Transaction Operations

##### `WithTx(ctx context.Context, fn func(tx pgx.Tx) error, levels ...pgx.TxIsoLevel) error`
Executes a function within a transaction with configurable isolation level (default: ReadCommitted).

```go
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    // Perform operations using tx
    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
    return err
}, pgx.Serializable)
```

##### `ReadTx(ctx context.Context, fn func(tx pgx.Tx) error) error`
Executes a read-only transaction with ReadCommitted isolation level.

```go
err := db.ReadTx(ctx, func(tx pgx.Tx) error {
    // Perform read operations using tx
    rows, err := tx.Query(ctx, "SELECT * FROM users")
    return err
})
```

##### Transaction Helper Functions

**`TxReturnRow[T any](ctx context.Context, tx pgx.Tx, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *T) error`**
Executes a query within a transaction and returns a single row into the result pointer.

**`TxRows[T any](ctx context.Context, tx pgx.Tx, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error`**
Executes a query within a transaction and returns multiple rows into the result pointer.

**`TxExec(ctx context.Context, tx pgx.Tx, builder sq.Sqlizer) error`**
Executes a statement within a transaction.

### Batch Operations

Batch operations allow you to queue multiple queries and execute them in a single round-trip to the database.

#### Standard Batch

##### `QueueExecRow(batch *pgx.Batch, builder sq.Sqlizer) error`
Queues a query that expects to affect at least one row.

```go
batch := &pgx.Batch{}
err := dblib.QueueExecRow(batch, 
    dblib.Psql.Update("users").Set("active", true).Where(sq.Eq{"id": 1}))
```

##### `QueueReturn[T any](batch *pgx.Batch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error`
Queues a query that returns multiple rows.

```go
var users []User
batch := &pgx.Batch{}
err := dblib.QueueReturn(batch,
    dblib.Psql.Select("*").From("users"),
    pgx.RowToStructByName[User],
    &users)
```

##### `QueueReturnRow[T any](batch *pgx.Batch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *T) error`
Queues a query that returns a single row.

```go
var user User
batch := &pgx.Batch{}
err := dblib.QueueReturnRow(batch,
    dblib.Psql.Select("*").From("users").Where(sq.Eq{"id": 1}),
    pgx.RowToStructByName[User],
    &user)
```

##### `QueueReturnBulk[T any](batch *pgx.Batch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error`
Queues a query and appends results to an existing slice (useful for collecting results from multiple queries).

#### Timed Batch

Timed batches automatically set a statement timeout for all queries in the batch.

##### `NewTimedBatch(timeoutMs int) *TimedBatch`
Creates a new batch with timeout configuration.

```go
batch := dblib.NewTimedBatch(5000) // 5 second timeout
```

##### `TimedQueueExecRow(batch *TimedBatch, builder sq.Sqlizer) error`
Queues a query with timeout that expects to affect at least one row.

##### `TimedQueueReturn[T any](batch *TimedBatch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error`
Queues a query with timeout that returns multiple rows.

##### `TimedQueueReturnRow[T any](batch *TimedBatch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *T) error`
Queues a query with timeout that returns a single row.

##### `TimedQueueReturnBulk[T any](batch *TimedBatch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error`
Queues a query with timeout and appends results to existing slice.

### Helper Functions

#### Null Handling

##### `NullString(value string) sql.NullString`
Converts a string to sql.NullString. Empty strings become NULL.

```go
nullStr := dblib.NullString(email) // NULL if email is ""
```

##### `NullInt(value int) sql.NullInt64`
Converts an int to sql.NullInt64. Zero values become NULL.

```go
nullInt := dblib.NullInt(count) // NULL if count is 0
```

##### `NullInt64(value int64) sql.NullInt64`
Converts an int64 to sql.NullInt64. Zero values become NULL.

##### `NullUint64(value uint64) sql.NullInt64`
Converts a uint64 to sql.NullInt64. Zero values become NULL.

##### `NullFloat64(value float64) sql.NullFloat64`
Converts a float64 to sql.NullFloat64. Zero values become NULL.

#### Struct Mapping Utilities

##### `RowToStructByTag[T any](row pgx.CollectableRow, tag string) (T, error)`
Maps a database row to a struct using custom struct tags.

```go
type User struct {
    ID   int    `custom:"user_id"`
    Name string `custom:"user_name"`
}
user, err := dblib.RowToStructByTag[User](row, "custom")
```

##### `StructToSetMap(article interface{}) map[string]interface{}`
Converts a struct to a map for UPDATE operations, excluding zero values. Uses json tags.

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
user := &User{Email: "new@example.com"} // Name is empty
setMap := dblib.StructToSetMap(user)
// setMap = map[string]interface{}{"email": "new@example.com"}
```

##### `GenerateMapFromStruct(instance interface{}, tag string) map[string]interface{}`
Converts a struct to a map using specified struct tags.

```go
colMap := dblib.GenerateMapFromStruct(user, "db")
```

##### `GenerateColumnsFromStruct(instance interface{}, tag string) []string`
Extracts column names from struct tags.

```go
columns := dblib.GenerateColumnsFromStruct(user, "db")
// Returns: ["id", "name", "email"]
```

#### Collection Helpers

##### `CollectRowsOK[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) ([]T, bool, error)`
Collects multiple rows with a boolean indicating if any rows were found.

##### `CollectOneRowOK[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) (T, bool, error)`
Collects one row with a boolean indicating if a row was found.

### Query Builder

##### `Psql` (Variable)
A pre-configured Squirrel statement builder for PostgreSQL with dollar sign placeholders.

```go
query := dblib.Psql.Select("*").From("users").Where(sq.Eq{"active": true})
```

### Database Factory

#### `NewDefaultDbFactory() DBFactory`
Creates a new database factory instance.

```go
factory := dblib.NewDefaultDbFactory()
```

#### `NewPreparedDBConfig(input DBConfig) *DBConfig`
Prepares and validates database configuration.

```go
preparedConfig := factory.NewPreparedDBConfig(dbConfig)
```

#### `CreateConnection(dbConfig *DBConfig, osdktrace *otelsdktrace.TracerProvider, metricsSet *metrics.Set) (*DB, error)`
Creates a database connection with the prepared configuration.

```go
db, err := factory.CreateConnection(preparedConfig, tracerProvider, metricsSet)
```

#### `SetCollectorName(name string)`
Sets a custom name for the metrics collector.

```go
factory.SetCollectorName("my_service_db")
```

### Health Check

#### `NewSQLProbe(db *DB) *SQLProbe`
Creates a health check probe for the database connection.

```go
probe := dblib.NewSQLProbe(db)
probe.SetName("Primary Database")
result := probe.Check(ctx)
```

### Metrics Collection

#### `NewCollector(stater Stater, labels map[string]string, set *metrics.Set) *Collector`
Creates a metrics collector for database connection pool statistics. Automatically tracks:
- Active connections
- Idle connections
- Connection acquisition metrics
- Connection lifecycle metrics

The collector integrates with VictoriaMetrics for monitoring.

## Examples

For complete example usage and implementation, please refer to the [API DB Example](https://gitlab.cept.gov.in/it-2.0-common/n-api-db/-/tree/main/example?ref_type=heads).

## Common Usage Patterns

### Basic CRUD Operations

```go
// Create
user := User{Name: "John", Email: "john@example.com"}
insertedUser, err := dblib.InsertReturning(ctx, db,
    dblib.Psql.Insert("users").Columns("name", "email").
        Values(user.Name, user.Email).
        Suffix("RETURNING *"),
    pgx.RowToStructByName[User])

// Read
user, err := dblib.SelectOne(ctx, db,
    dblib.Psql.Select("*").From("users").Where(sq.Eq{"id": 1}),
    pgx.RowToStructByName[User])

// Update
tag, err := dblib.Update(ctx, db,
    dblib.Psql.Update("users").Set("email", "new@example.com").Where(sq.Eq{"id": 1}))

// Delete
tag, err := dblib.Delete(ctx, db,
    dblib.Psql.Delete("users").Where(sq.Eq{"id": 1}))
```

### Batch Operations Example

```go
// Execute multiple queries in one round-trip
batch := &pgx.Batch{}

var user User
var orders []Order

dblib.QueueReturnRow(batch,
    dblib.Psql.Select("*").From("users").Where(sq.Eq{"id": userId}),
    pgx.RowToStructByName[User],
    &user)

dblib.QueueReturn(batch,
    dblib.Psql.Select("*").From("orders").Where(sq.Eq{"user_id": userId}),
    pgx.RowToStructByName[Order],
    &orders)

br := db.SendBatch(ctx, batch)
defer br.Close()

// Process results
err := br.Close()
```

### Transaction Example

```go
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    // Insert user
    var userId int64
    err := tx.QueryRow(ctx, 
        "INSERT INTO users (name) VALUES ($1) RETURNING id", 
        "John").Scan(&userId)
    if err != nil {
        return err
    }
    
    // Insert related record
    _, err = tx.Exec(ctx,
        "INSERT INTO profiles (user_id, bio) VALUES ($1, $2)",
        userId, "Bio text")
    return err
})
```

## Notes

- Ensure your PostgreSQL server is running and accessible with the provided configuration.
- Use appropriate isolation levels and access modes for transactions based on your application requirements.
- To use the functions from `utility.go` or `helper.go`, prefix them with `dblib.` (assuming the import alias is used as `dblib`).
- Batch operations are more efficient than individual queries when performing multiple database operations.
- Always use prepared statements (via Squirrel builders) to prevent SQL injection.
- The library uses context for cancellation and timeout management - always pass appropriate contexts.
- For production use, consider enabling database tracing by setting `Trace: true` in DBConfig.

## Support

If you encounter any issues or have questions about using the library, feel free to reach out via the following channels:

- **Issue Tracker**: [GitLab Issue Tracker](https://gitlab.cept.gov.in/it-2.0-common/n-api-db/issues)
- **Email**: [maintainer.email@example.com](mailto:soumyabrata.t@indiapost.gov.in)
 




