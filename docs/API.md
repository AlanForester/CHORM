# CHORM API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Core Types](#core-types)
5. [Database Connection](#database-connection)
6. [Struct Mapping](#struct-mapping)
7. [CRUD Operations](#crud-operations)
8. [Query Builder](#query-builder)
9. [Aggregate Functions](#aggregate-functions)
10. [Window Functions](#window-functions)
11. [Migrations](#migrations)
12. [Schema Management](#schema-management)
13. [Cluster Support](#cluster-support)
14. [Transactions](#transactions)
15. [Error Handling](#error-handling)
16. [Performance Tips](#performance-tips)
17. [Examples](#examples)

## Overview

CHORM is a comprehensive Go ORM library for ClickHouse that provides:

- **Struct-to-Table Mapping**: Automatic mapping of Go structs to ClickHouse tables
- **CRUD Operations**: Complete Create, Read, Update, Delete operations
- **Query Builder**: Fluent interface for building complex queries
- **Aggregate Functions**: Support for ClickHouse analytical functions
- **Window Functions**: Advanced analytical capabilities
- **Migrations**: Schema versioning and management
- **Cluster Support**: Distributed table operations
- **Performance Optimizations**: Batch inserts, connection pooling
- **Security**: TLS support, parameterized queries

## Installation

```bash
go get github.com/AlanForester/chorm
```

## Configuration

### Config Struct

```go
type Config struct {
    Host            string        // ClickHouse host (default: localhost)
    Port            int           // ClickHouse port (default: 9000)
    Database        string        // Database name
    Username        string        // Username (default: default)
    Password        string        // Password
    MaxOpenConns    int           // Max open connections (default: 10)
    MaxIdleConns    int           // Max idle connections (default: 5)
    ConnMaxLifetime time.Duration // Connection max lifetime (default: 1h)
    TLS             bool          // Enable TLS
    Compression     bool          // Enable compression
    Debug           bool          // Enable debug logging
}
```

### Example Configuration

```go
config := chorm.Config{
    Host:         "localhost",
    Port:         9000,
    Database:     "analytics",
    Username:     "default",
    Password:     "password",
    MaxOpenConns: 20,
    TLS:          true,
    Compression:  true,
    Debug:        false,
}
```

## Core Types

### ClickHouse Data Types

```go
const (
    // Basic Types
    TypeUInt8       ClickHouseType = "UInt8"
    TypeUInt16      ClickHouseType = "UInt16"
    TypeUInt32      ClickHouseType = "UInt32"
    TypeUInt64      ClickHouseType = "UInt64"
    TypeInt8        ClickHouseType = "Int8"
    TypeInt16       ClickHouseType = "Int16"
    TypeInt32       ClickHouseType = "Int32"
    TypeInt64       ClickHouseType = "Int64"
    TypeFloat32     ClickHouseType = "Float32"
    TypeFloat64     ClickHouseType = "Float64"
    TypeString      ClickHouseType = "String"
    TypeFixedString ClickHouseType = "FixedString"
    TypeDate        ClickHouseType = "Date"
    TypeDateTime    ClickHouseType = "DateTime"
    TypeDateTime64  ClickHouseType = "DateTime64"
    TypeBoolean     ClickHouseType = "Boolean"
    TypeUUID        ClickHouseType = "UUID"

    // Complex Types
    TypeArray          ClickHouseType = "Array"
    TypeNullable       ClickHouseType = "Nullable"
    TypeLowCardinality ClickHouseType = "LowCardinality"
    TypeEnum           ClickHouseType = "Enum"
    TypeNested         ClickHouseType = "Nested"
    TypeTuple          ClickHouseType = "Tuple"
    TypeMap            ClickHouseType = "Map"
)
```

### Table Engines

```go
const (
    EngineMergeTree                    Engine = "MergeTree"
    EngineReplacingMergeTree           Engine = "ReplacingMergeTree"
    EngineSummingMergeTree             Engine = "SummingMergeTree"
    EngineAggregatingMergeTree         Engine = "AggregatingMergeTree"
    EngineCollapsingMergeTree          Engine = "CollapsingMergeTree"
    EngineVersionedCollapsingMergeTree Engine = "VersionedCollapsingMergeTree"
    EngineGraphiteMergeTree            Engine = "GraphiteMergeTree"
    EngineTinyLog                      Engine = "TinyLog"
    EngineLog                          Engine = "Log"
    EngineMemory                       Engine = "Memory"
    EngineSet                          Engine = "Set"
    EngineJoin                         Engine = "Join"
    EngineBuffer                       Engine = "Buffer"
    EngineMaterializedView             Engine = "MaterializedView"
)
```

## Database Connection

### Connect

```go
func Connect(ctx context.Context, config Config) (*DB, error)
```

Creates a new connection to ClickHouse.

```go
db, err := chorm.Connect(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Close

```go
func (db *DB) Close() error
```

Closes the database connection.

## Struct Mapping

### Struct Tags

CHORM uses struct tags to map Go structs to ClickHouse tables:

```go
type User struct {
    ID       uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
    Name     string    `ch:"name" ch_type:"String"`
    Email    string    `ch:"email" ch_type:"String"`
    Age      uint8     `ch:"age" ch_type:"UInt8"`
    Created  time.Time `ch:"created" ch_type:"DateTime"`
    IsActive bool      `ch:"is_active" ch_type:"Boolean"`
    Score    float64   `ch:"score" ch_type:"Float64"`
}
```

### Available Tags

- `ch`: Column name in ClickHouse
- `ch_type`: ClickHouse data type
- `ch_pk`: Primary key flag
- `ch_auto`: Auto-increment flag
- `ch_nullable`: Nullable flag
- `ch_engine`: Table engine (struct-level)

### Model Interface

```go
type Model interface {
    TableName() string
}
```

Implement this interface to specify custom table names:

```go
func (u *User) TableName() string {
    return "users"
}
```

## CRUD Operations

### Create Table

```go
func (db *DB) CreateTable(ctx context.Context, model interface{}) error
```

Automatically creates a table based on struct definition:

```go
err := db.CreateTable(ctx, &User{})
```

### Insert

```go
func (db *DB) Insert(ctx context.Context, model interface{}) error
```

Inserts a single record:

```go
user := &User{
    ID:       1,
    Name:     "John Doe",
    Email:    "john@example.com",
    Age:      30,
    Created:  time.Now(),
    IsActive: true,
    Score:    85.5,
}
err := db.Insert(ctx, user)
```

### Insert Batch

```go
func (db *DB) InsertBatch(ctx context.Context, models []interface{}) error
```

Inserts multiple records efficiently:

```go
var users []interface{}
for i := 1; i <= 1000; i++ {
    user := &User{
        ID:       uint32(i),
        Name:     fmt.Sprintf("User %d", i),
        Email:    fmt.Sprintf("user%d@example.com", i),
        Age:      uint8(20 + i%50),
        Created:  time.Now(),
        IsActive: i%2 == 0,
        Score:    float64(50 + i%50),
    }
    users = append(users, user)
}
err := db.InsertBatch(ctx, users)
```

### Query

```go
func (db *DB) Query(ctx context.Context, result interface{}, query string, args ...interface{}) error
```

Executes a query and scans results into a slice:

```go
var users []User
err := db.Query(ctx, &users, "SELECT * FROM users WHERE age > ?", 25)
```

### QueryRow

```go
func (db *DB) QueryRow(ctx context.Context, result interface{}, query string, args ...interface{}) error
```

Executes a query and scans a single row:

```go
var user User
err := db.QueryRow(ctx, &user, "SELECT * FROM users WHERE id = ?", 1)
```

### Exec

```go
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
```

Executes a query without returning results:

```go
result, err := db.Exec(ctx, "DELETE FROM users WHERE age < ?", 18)
```

## Query Builder

### NewQuery

```go
func (db *DB) NewQuery() *Query
```

Creates a new query builder:

```go
query := db.NewQuery()
```

### Table

```go
func (q *Query) Table(table string) *Query
```

Sets the table for the query:

```go
query := db.NewQuery().Table("users")
```

### Select

```go
func (q *Query) Select(fields ...string) *Query
```

Specifies fields to select:

```go
query := db.NewQuery().
    Table("users").
    Select("id", "name", "email", "age")
```

### Where Conditions

```go
// Basic WHERE
func (q *Query) Where(condition string, args ...interface{}) *Query

// WHERE IN
func (q *Query) WhereIn(field string, values []interface{}) *Query

// WHERE NOT IN
func (q *Query) WhereNotIn(field string, values []interface{}) *Query

// WHERE BETWEEN
func (q *Query) WhereBetween(field string, start, end interface{}) *Query

// WHERE LIKE
func (q *Query) WhereLike(field, pattern string) *Query

// WHERE NULL
func (q *Query) WhereNull(field string) *Query

// WHERE NOT NULL
func (q *Query) WhereNotNull(field string) *Query
```

### Joins

```go
// INNER JOIN
func (q *Query) Join(table, condition string, args ...interface{}) *Query

// LEFT JOIN
func (q *Query) LeftJoin(table, condition string, args ...interface{}) *Query

// RIGHT JOIN
func (q *Query) RightJoin(table, condition string, args ...interface{}) *Query
```

### Grouping and Ordering

```go
// GROUP BY
func (q *Query) GroupBy(fields ...string) *Query

// HAVING
func (q *Query) Having(condition string, args ...interface{}) *Query

// ORDER BY
func (q *Query) OrderBy(field string, direction ...string) *Query

// ORDER BY ASC
func (q *Query) OrderByAsc(field string) *Query

// ORDER BY DESC
func (q *Query) OrderByDesc(field string) *Query
```

### Pagination

```go
// LIMIT
func (q *Query) Limit(limit int) *Query

// OFFSET
func (q *Query) Offset(offset int) *Query

// Paginate
func (q *Query) Paginate(ctx context.Context, page, perPage int, result interface{}) (int64, error)
```

### Query Execution

```go
// Get single record
func (q *Query) Get(ctx context.Context, result interface{}) error

// Get all records
func (q *Query) All(ctx context.Context, result interface{}) error

// Count records
func (q *Query) Count(ctx context.Context) (int64, error)

// Check if exists
func (q *Query) Exists(ctx context.Context) (bool, error)

// Get first record
func (q *Query) First(ctx context.Context, result interface{}) error

// Get last record
func (q *Query) Last(ctx context.Context, result interface{}) error

// Update records
func (q *Query) Update(ctx context.Context, data map[string]interface{}) (Result, error)

// Delete records
func (q *Query) Delete(ctx context.Context) (Result, error)
```

### Example Query

```go
var users []User
err := db.NewQuery().
    Table("users").
    Select("id", "name", "email", "age").
    Where("age > ?", 25).
    Where("is_active = ?", true).
    OrderBy("created", "DESC").
    Limit(10).
    All(ctx, &users)
```

## Aggregate Functions

### NewAggregate

```go
func (q *Query) NewAggregate() *Aggregate
```

Creates a new aggregate query:

```go
agg := db.NewQuery().Table("users").NewAggregate()
```

### Basic Aggregates

```go
// SUM
func (a *Aggregate) Sum(field string) *Aggregate

// AVG
func (a *Aggregate) Avg(field string) *Aggregate

// MIN
func (a *Aggregate) Min(field string) *Aggregate

// MAX
func (a *Aggregate) Max(field string) *Aggregate

// COUNT
func (a *Aggregate) Count(field string) *Aggregate

// COUNT DISTINCT
func (a *Aggregate) CountDistinct(field string) *Aggregate
```

### ClickHouse Specific Aggregates

```go
// Uniq (approximate)
func (a *Aggregate) Uniq(field string) *Aggregate

// UniqExact (exact)
func (a *Aggregate) UniqExact(field string) *Aggregate

// Quantile
func (a *Aggregate) Quantile(level float64, field string) *Aggregate

// Median
func (a *Aggregate) Median(field string) *Aggregate

// Standard Deviation
func (a *Aggregate) StdDev(field string) *Aggregate

// Variance
func (a *Aggregate) Variance(field string) *Aggregate

// Any
func (a *Aggregate) Any(field string) *Aggregate

// ArgMin/ArgMax
func (a *Aggregate) ArgMin(arg, val string) *Aggregate
func (a *Aggregate) ArgMax(arg, val string) *Aggregate

// GroupArray
func (a *Aggregate) GroupArray(field string) *Aggregate

// TopK
func (a *Aggregate) TopK(k int, field string) *Aggregate

// Histogram
func (a *Aggregate) Histogram(bins int, field string) *Aggregate
```

### Statistical Functions

```go
// Correlation
func (a *Aggregate) Corr(x, y string) *Aggregate

// Covariance
func (a *Aggregate) CovarPop(x, y string) *Aggregate
func (a *Aggregate) CovarSamp(x, y string) *Aggregate

// Skewness
func (a *Aggregate) SkewPop(field string) *Aggregate

// Kurtosis
func (a *Aggregate) KurtPop(field string) *Aggregate

// Entropy
func (a *Aggregate) Entropy(field string) *Aggregate

// Geometric Mean
func (a *Aggregate) GeometricMean(field string) *Aggregate

// Harmonic Mean
func (a *Aggregate) HarmonicMean(field string) *Aggregate
```

### Example Aggregate Query

```go
type UserStats struct {
    Count        int64   `ch:"count"`
    AvgAge       float64 `ch:"avg_age"`
    MaxScore     float64 `ch:"max_score"`
    UniqEmails   uint64  `ch:"uniq_emails"`
    MedianAge    float64 `ch:"median_age"`
}

var stats UserStats
err := db.NewQuery().
    Table("users").
    NewAggregate().
    Count("*").
    Avg("age").
    Max("score").
    Uniq("email").
    Median("age").
    Get(ctx, &stats)
```

## Window Functions

### NewWindow

```go
func (q *Query) NewWindow() *Window
```

Creates a new window function:

```go
window := db.NewQuery().Table("users").NewWindow()
```

### Window Functions

```go
// Row Number
func (w *Window) RowNumber() *Window

// Rank
func (w *Window) Rank() *Window

// Dense Rank
func (w *Window) DenseRank() *Window

// Lag/Lead
func (w *Window) Lag(field string, offset int) *Window
func (w *Window) Lead(field string, offset int) *Window

// First/Last Value
func (w *Window) FirstValue(field string) *Window
func (w *Window) LastValue(field string) *Window

// Nth Value
func (w *Window) NthValue(field string, n int) *Window

// NTile
func (w *Window) Ntile(buckets int) *Window

// Percent Rank
func (w *Window) PercentRank() *Window

// Cumulative Distribution
func (w *Window) CumeDist() *Window
```

### Window Specification

```go
// OVER clause
func (w *Window) Over(partitionBy, orderBy string) *Window

// Alias
func (w *Window) As(alias string) *Window
```

### Example Window Query

```go
type UserRank struct {
    ID       uint32  `ch:"id"`
    Name     string  `ch:"name"`
    Score    float64 `ch:"score"`
    Rank     uint64  `ch:"rank"`
    RowNum   uint64  `ch:"row_num"`
}

var rankings []UserRank
err := db.NewQuery().
    Table("users").
    Select("id", "name", "score").
    NewWindow().
    Rank().
    Over("", "score DESC").
    As("rank").
    AddToQuery().
    NewWindow().
    RowNumber().
    Over("", "score DESC").
    As("row_num").
    AddToQuery().
    All(ctx, &rankings)
```

## Migrations

### NewMigrator

```go
func NewMigrator(db *DB) *Migrator
```

Creates a new migrator:

```go
migrator := chorm.NewMigrator(db)
```

### Add Migration

```go
func (m *Migrator) AddMigration(name string, up, down MigrationFunc) *Migrator
```

Adds a migration:

```go
migrator.AddMigration("create_users_table", 
    func(ctx context.Context, db *chorm.DB) error {
        return db.CreateTable(ctx, &User{})
    },
    func(ctx context.Context, db *chorm.DB) error {
        _, err := db.Exec(ctx, "DROP TABLE IF EXISTS users")
        return err
    },
)
```

### Migration Operations

```go
// Apply all migrations
func (m *Migrator) Migrate(ctx context.Context) error

// Rollback last migration
func (m *Migrator) Rollback(ctx context.Context) error

// Rollback specific migration
func (m *Migrator) RollbackMigration(ctx context.Context, name string) error

// Show migration status
func (m *Migrator) Status(ctx context.Context) error
```

### Example Migration

```go
migrator := chorm.NewMigrator(db)

// Create users table
migrator.AddMigration("001_create_users", 
    func(ctx context.Context, db *chorm.DB) error {
        return db.CreateTable(ctx, &User{})
    },
    func(ctx context.Context, db *chorm.DB) error {
        _, err := db.Exec(ctx, "DROP TABLE IF EXISTS users")
        return err
    },
)

// Add email index
migrator.AddMigration("002_add_email_index",
    func(ctx context.Context, db *chorm.DB) error {
        _, err := db.Exec(ctx, "ALTER TABLE users ADD INDEX idx_email (email)")
        return err
    },
    func(ctx context.Context, db *chorm.DB) error {
        _, err := db.Exec(ctx, "ALTER TABLE users DROP INDEX idx_email")
        return err
    },
)

// Apply migrations
err := migrator.Migrate(ctx)
```

## Schema Management

### NewSchema

```go
func NewSchema(db *DB) *Schema
```

Creates a new schema manager:

```go
schema := chorm.NewSchema(db)
```

### Database Operations

```go
// Create database
func (s *Schema) CreateDatabase(ctx context.Context, name string) error

// Drop database
func (s *Schema) DropDatabase(ctx context.Context, name string) error

// Get databases
func (s *Schema) GetDatabases(ctx context.Context) ([]string, error)
```

### Table Operations

```go
// Create table
func (s *Schema) CreateTable(ctx context.Context, tableName string, columns []string, engine string, options map[string]string) error

// Drop table
func (s *Schema) DropTable(ctx context.Context, tableName string) error

// Truncate table
func (s *Schema) TruncateTable(ctx context.Context, tableName string) error

// Rename table
func (s *Schema) RenameTable(ctx context.Context, oldName, newName string) error

// Get tables
func (s *Schema) GetTables(ctx context.Context) ([]string, error)

// Get table info
func (s *Schema) GetTableInfo(ctx context.Context, tableName string) (map[string]interface{}, error)
```

### Column Operations

```go
// Add column
func (s *Schema) AddColumn(ctx context.Context, tableName, columnName, columnType string) error

// Drop column
func (s *Schema) DropColumn(ctx context.Context, tableName, columnName string) error

// Modify column
func (s *Schema) ModifyColumn(ctx context.Context, tableName, columnName, newType string) error

// Rename column
func (s *Schema) RenameColumn(ctx context.Context, tableName, oldName, newName string) error
```

### Index Operations

```go
// Create index
func (s *Schema) CreateIndex(ctx context.Context, indexName, tableName string, columns []string) error

// Drop index
func (s *Schema) DropIndex(ctx context.Context, indexName, tableName string) error
```

### Materialized Views

```go
// Create materialized view
func (s *Schema) CreateMaterializedView(ctx context.Context, viewName, tableName, selectQuery string) error

// Drop materialized view
func (s *Schema) DropMaterializedView(ctx context.Context, viewName string) error
```

## Cluster Support

### NewCluster

```go
func NewCluster(name string) *Cluster
```

Creates a new cluster:

```go
cluster := chorm.NewCluster("my_cluster")
```

### Cluster Node

```go
type ClusterNode struct {
    Host     string
    Port     int
    Database string
    Username string
    Password string
    Weight   int    // Load balancing weight
    Healthy  bool
    LastPing time.Time
}
```

### Cluster Operations

```go
// Add node
func (c *Cluster) AddNode(node *ClusterNode)

// Remove node
func (c *Cluster) RemoveNode(host string, port int)

// Get healthy nodes
func (c *Cluster) GetHealthyNodes() []*ClusterNode

// Health check
func (c *Cluster) HealthCheck(ctx context.Context)
```

### Connect to Cluster

```go
func ConnectToCluster(cluster *Cluster, config Config) (*ClusterDB, error)
```

### ClusterDB Operations

```go
// Get connection
func (cdb *ClusterDB) GetConnection(ctx context.Context) (*DB, error)

// Query on cluster
func (cdb *ClusterDB) Query(ctx context.Context, result interface{}, query string, args ...interface{}) error

// Exec on cluster
func (cdb *ClusterDB) Exec(ctx context.Context, query string, args ...interface{}) (Result, error)

// Create distributed table
func (cdb *ClusterDB) CreateDistributedTable(ctx context.Context, tableName, clusterName, localTableName string, shardingKey string) error

// Insert into distributed table
func (cdb *ClusterDB) InsertIntoDistributed(ctx context.Context, tableName string, data interface{}) error
```

### Example Cluster Usage

```go
// Create cluster
cluster := chorm.NewCluster("analytics_cluster")

// Add nodes
cluster.AddNode(&chorm.ClusterNode{
    Host:     "node1.example.com",
    Port:     9000,
    Database: "analytics",
    Username: "default",
    Password: "password",
    Weight:   1,
})

cluster.AddNode(&chorm.ClusterNode{
    Host:     "node2.example.com",
    Port:     9000,
    Database: "analytics",
    Username: "default",
    Password: "password",
    Weight:   1,
})

// Connect to cluster
clusterDB, err := chorm.ConnectToCluster(cluster, config)

// Create distributed table
err = clusterDB.CreateDistributedTable(ctx, "distributed_users", "analytics_cluster", "users", "id")

// Insert into distributed table
err = clusterDB.InsertIntoDistributed(ctx, "distributed_users", user)
```

## Transactions

### Begin Transaction

```go
func (db *DB) Begin(ctx context.Context) (*Tx, error)
```

### Transaction Operations

```go
// Commit transaction
func (tx *Tx) Commit() error

// Rollback transaction
func (tx *Tx) Rollback() error

// Execute in transaction
func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
```

### Example Transaction

```go
tx, err := db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

// Execute operations
_, err = tx.Exec(ctx, "INSERT INTO users (id, name) VALUES (?, ?)", 1, "John")
if err != nil {
    return err
}

_, err = tx.Exec(ctx, "UPDATE user_stats SET count = count + 1")
if err != nil {
    return err
}

// Commit transaction
return tx.Commit()
```

## Error Handling

### Error Types

```go
type Error struct {
    Code    int
    Message string
    Query   string
}
```

### Error Handling Example

```go
db, err := chorm.Connect(ctx, config)
if err != nil {
    if chormErr, ok := err.(*chorm.Error); ok {
        log.Printf("ClickHouse error %d: %s", chormErr.Code, chormErr.Message)
        log.Printf("Query: %s", chormErr.Query)
    } else {
        log.Printf("Connection error: %v", err)
    }
    return
}
```

## Performance Tips

### 1. Use Batch Inserts

For large datasets, use `InsertBatch` instead of multiple `Insert` calls:

```go
// Good
var users []interface{}
for i := 0; i < 10000; i++ {
    users = append(users, &User{...})
}
err := db.InsertBatch(ctx, users)

// Avoid
for i := 0; i < 10000; i++ {
    err := db.Insert(ctx, &User{...})
}
```

### 2. Optimize Connection Pool

Configure appropriate connection pool settings:

```go
config := chorm.Config{
    MaxOpenConns:    20,
    MaxIdleConns:    10,
    ConnMaxLifetime: 30 * time.Minute,
}
```

### 3. Use Appropriate Data Types

Choose the most efficient ClickHouse data types:

```go
type OptimizedUser struct {
    ID       uint32    `ch:"id" ch_type:"UInt32"`           // Instead of UInt64
    Age      uint8     `ch:"age" ch_type:"UInt8"`           // Instead of UInt16
    Score    float32   `ch:"score" ch_type:"Float32"`       // Instead of Float64
    Name     string    `ch:"name" ch_type:"LowCardinality(String)"` // For repeated values
}
```

### 4. Use Materialized Views

For frequently accessed aggregated data:

```go
err := schema.CreateMaterializedView(ctx, "user_stats", "users", 
    "SELECT age_group, COUNT(*) as count, AVG(score) as avg_score " +
    "FROM users GROUP BY age_group")
```

### 5. Optimize Queries

Use appropriate indexes and query patterns:

```go
// Use WHERE on indexed columns
query := db.NewQuery().
    Table("users").
    Where("created >= ?", startDate).
    Where("age BETWEEN ? AND ?", minAge, maxAge)

// Use LIMIT for large result sets
query := db.NewQuery().
    Table("events").
    OrderBy("timestamp", "DESC").
    Limit(1000)
```

## Examples

### Complete Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/forester/chorm"
)

type User struct {
    ID       uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
    Name     string    `ch:"name" ch_type:"String"`
    Email    string    `ch:"email" ch_type:"String"`
    Age      uint8     `ch:"age" ch_type:"UInt8"`
    Created  time.Time `ch:"created" ch_type:"DateTime"`
    IsActive bool      `ch:"is_active" ch_type:"Boolean"`
    Score    float64   `ch:"score" ch_type:"Float64"`
}

func (u *User) TableName() string {
    return "users"
}

func main() {
    ctx := context.Background()
    
    // Connect to ClickHouse
    db, err := chorm.Connect(ctx, chorm.Config{
        Host:     "localhost",
        Port:     9000,
        Database: "analytics",
        Username: "default",
        Password: "",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create table
    if err := db.CreateTable(ctx, &User{}); err != nil {
        log.Fatal(err)
    }
    
    // Insert data
    user := &User{
        ID:       1,
        Name:     "John Doe",
        Email:    "john@example.com",
        Age:      30,
        Created:  time.Now(),
        IsActive: true,
        Score:    85.5,
    }
    
    if err := db.Insert(ctx, user); err != nil {
        log.Fatal(err)
    }
    
    // Query data
    var users []User
    err = db.NewQuery().
        Table("users").
        Where("age > ?", 25).
        Where("is_active = ?", true).
        OrderBy("score", "DESC").
        All(ctx, &users)
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d users", len(users))
}
```

### Migration Example

```go
func setupMigrations(db *chorm.DB) error {
    migrator := chorm.NewMigrator(db)
    
    // Create users table
    migrator.AddMigration("001_create_users", 
        func(ctx context.Context, db *chorm.DB) error {
            return db.CreateTable(ctx, &User{})
        },
        func(ctx context.Context, db *chorm.DB) error {
            _, err := db.Exec(ctx, "DROP TABLE IF EXISTS users")
            return err
        },
    )
    
    // Add indexes
    migrator.AddMigration("002_add_indexes",
        func(ctx context.Context, db *chorm.DB) error {
            schema := chorm.NewSchema(db)
            return schema.CreateIndex(ctx, "idx_email", "users", []string{"email"})
        },
        func(ctx context.Context, db *chorm.DB) error {
            schema := chorm.NewSchema(db)
            return schema.DropIndex(ctx, "idx_email", "users")
        },
    )
    
    return migrator.Migrate(context.Background())
}
```

### Aggregate Example

```go
type UserAnalytics struct {
    TotalUsers    int64   `ch:"total_users"`
    AvgAge        float64 `ch:"avg_age"`
    MaxScore      float64 `ch:"max_score"`
    ActiveUsers   int64   `ch:"active_users"`
    TopScore      float64 `ch:"top_score"`
}

func getUserAnalytics(db *chorm.DB) (*UserAnalytics, error) {
    var analytics UserAnalytics
    
    err := db.NewQuery().
        Table("users").
        NewAggregate().
        Count("*").
        Avg("age").
        Max("score").
        Count("CASE WHEN is_active = 1 THEN 1 END").
        Quantile(0.95, "score").
        Get(context.Background(), &analytics)
    
    return &analytics, err
}
```

This comprehensive API documentation covers all the features of the CHORM library, providing developers with detailed information on how to use each component effectively.