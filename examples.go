package chorm

import (
	"context"
	"fmt"
	"log"
	"time"
)

// User представляет пользователя
type User struct {
	ID       uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
	Name     string    `ch:"name" ch_type:"String"`
	Email    string    `ch:"email" ch_type:"String"`
	Age      uint8     `ch:"age" ch_type:"UInt8"`
	Created  time.Time `ch:"created" ch_type:"DateTime"`
	Updated  time.Time `ch:"updated" ch_type:"DateTime"`
	IsActive bool      `ch:"is_active" ch_type:"Boolean"`
	Score    float64   `ch:"score" ch_type:"Float64"`
}

// TableName возвращает имя таблицы
func (u *User) TableName() string {
	return "users"
}

// Order представляет заказ
type Order struct {
	ID        uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
	UserID    uint32    `ch:"user_id" ch_type:"UInt32"`
	ProductID uint32    `ch:"product_id" ch_type:"UInt32"`
	Quantity  uint16    `ch:"quantity" ch_type:"UInt16"`
	Price     float64   `ch:"price" ch_type:"Float64"`
	Total     float64   `ch:"total" ch_type:"Float64"`
	Status    string    `ch:"status" ch_type:"String"`
	Created   time.Time `ch:"created" ch_type:"DateTime"`
	Completed time.Time `ch:"completed" ch_type:"DateTime"`
}

// TableName возвращает имя таблицы
func (o *Order) TableName() string {
	return "orders"
}

// Product представляет продукт
type Product struct {
	ID          uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
	Name        string    `ch:"name" ch_type:"String"`
	Description string    `ch:"description" ch_type:"String"`
	Price       float64   `ch:"price" ch_type:"Float64"`
	Category    string    `ch:"category" ch_type:"String"`
	InStock     bool      `ch:"in_stock" ch_type:"Boolean"`
	Created     time.Time `ch:"created" ch_type:"DateTime"`
}

// TableName возвращает имя таблицы
func (p *Product) TableName() string {
	return "products"
}

// UserStats представляет статистику пользователей
type UserStats struct {
	UserID       uint32  `ch:"user_id" ch_type:"UInt32"`
	TotalOrders  uint32  `ch:"total_orders" ch_type:"UInt32"`
	TotalSpent   float64 `ch:"total_spent" ch_type:"Float64"`
	AvgOrderSize float64 `ch:"avg_order_size" ch_type:"Float64"`
	LastOrder    string  `ch:"last_order" ch_type:"String"`
}

// TableName возвращает имя таблицы
func (us *UserStats) TableName() string {
	return "user_stats"
}

// ExampleBasicUsage демонстрирует базовое использование ORM
func ExampleBasicUsage() {
	ctx := context.Background()

	// Подключаемся к ClickHouse
	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
		Debug:    true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицы
	if err := db.CreateTable(ctx, &User{}); err != nil {
		log.Fatal(err)
	}
	if err := db.CreateTable(ctx, &Product{}); err != nil {
		log.Fatal(err)
	}
	if err := db.CreateTable(ctx, &Order{}); err != nil {
		log.Fatal(err)
	}

	// Вставляем пользователя
	user := &User{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Created:  time.Now(),
		Updated:  time.Now(),
		IsActive: true,
		Score:    95.5,
	}

	if err := db.Insert(ctx, user); err != nil {
		log.Fatal(err)
	}

	// Вставляем продукт
	product := &Product{
		ID:          1,
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999.99,
		Category:    "Electronics",
		InStock:     true,
		Created:     time.Now(),
	}

	if err := db.Insert(ctx, product); err != nil {
		log.Fatal(err)
	}

	// Вставляем заказ
	order := &Order{
		ID:        1,
		UserID:    1,
		ProductID: 1,
		Quantity:  2,
		Price:     999.99,
		Total:     1999.98,
		Status:    "pending",
		Created:   time.Now(),
	}

	if err := db.Insert(ctx, order); err != nil {
		log.Fatal(err)
	}

	// Выполняем запрос
	var users []User
	if err := db.Query(ctx, &users, "SELECT * FROM users WHERE age > ?", 25); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d users\n", len(users))
}

// ExampleQueryBuilder демонстрирует использование построителя запросов
func ExampleQueryBuilder() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Используем построитель запросов
	query := db.NewQuery().
		Table("users").
		Select("id", "name", "email", "age").
		Where("age > ?", 25).
		Where("is_active = ?", true).
		OrderBy("created", "DESC").
		Limit(10)

	var users []User
	if err := query.All(ctx, &users); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d active users over 25\n", len(users))

	// Подсчет записей
	count, err := query.Count(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total count: %d\n", count)

	// Пагинация
	var pageUsers []User
	total, err := query.Paginate(ctx, 1, 5, &pageUsers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Page 1: %d users, Total: %d\n", len(pageUsers), total)
}

// ExampleAggregations демонстрирует использование агрегатных функций
func ExampleAggregations() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Используем агрегатные функции
	query := db.NewQuery().Table("orders")
	agg := query.NewAggregate().
		Count("*").
		Sum("total").
		Avg("total").
		Min("total").
		Max("total").
		Uniq("user_id")

	var result map[string]interface{}
	if err := agg.Get(ctx, &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Aggregation results: %+v\n", result)

	// Группировка
	groupQuery := db.NewQuery().
		Table("orders").
		Select("user_id", "COUNT(*) as order_count", "SUM(total) as total_spent").
		GroupBy("user_id").
		Having("total_spent > ?", 1000).
		OrderBy("total_spent", "DESC")

	var userStats []UserStats
	if err := groupQuery.All(ctx, &userStats); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User statistics: %+v\n", userStats)
}

// ExampleJoins демонстрирует использование JOIN
func ExampleJoins() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// JOIN запрос
	query := db.NewQuery().
		Table("orders").
		Select("orders.id", "users.name", "products.name as product_name", "orders.quantity", "orders.total").
		Join("users", "orders.user_id = users.id").
		Join("products", "orders.product_id = products.id").
		Where("orders.status = ?", "completed").
		OrderBy("orders.created", "DESC")

	var results []map[string]interface{}
	if err := query.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Joined results: %+v\n", results)
}

// ExampleBatchOperations демонстрирует массовые операции
func ExampleBatchOperations() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Массовая вставка пользователей
	var users []interface{}
	for i := 1; i <= 100; i++ {
		user := &User{
			ID:       uint32(i),
			Name:     fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Age:      uint8(20 + (i % 50)),
			Created:  time.Now(),
			Updated:  time.Now(),
			IsActive: i%2 == 0,
			Score:    float64(50 + (i % 50)),
		}
		users = append(users, user)
	}

	if err := db.InsertBatch(ctx, users); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted %d users in batch\n", len(users))
}

// ExampleTransactions демонстрирует использование транзакций
func ExampleTransactions() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Начинаем транзакцию
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Выполняем операции в транзакции
	_, err = tx.Exec(ctx, "INSERT INTO users (id, name, email, age, created, updated, is_active, score) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		999, "Transaction User", "tx@example.com", 25, time.Now(), time.Now(), true, 100.0)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	_, err = tx.Exec(ctx, "UPDATE users SET score = score + 10 WHERE id = ?", 999)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction completed successfully")
}

// ExampleMigrations демонстрирует использование миграций
func ExampleMigrations() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем мигратор
	migrator := NewMigrator(db)

	// Добавляем миграции
	migrator.AddMigration("create_users_table", func(ctx context.Context, db *DB) error {
		return db.CreateTable(ctx, &User{})
	}, func(ctx context.Context, db *DB) error {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS users")
		return err
	})

	migrator.AddMigration("create_products_table", func(ctx context.Context, db *DB) error {
		return db.CreateTable(ctx, &Product{})
	}, func(ctx context.Context, db *DB) error {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS products")
		return err
	})

	migrator.AddMigration("create_orders_table", func(ctx context.Context, db *DB) error {
		return db.CreateTable(ctx, &Order{})
	}, func(ctx context.Context, db *DB) error {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS orders")
		return err
	})

	// Применяем миграции
	if err := migrator.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	// Показываем статус
	if err := migrator.Status(ctx); err != nil {
		log.Fatal(err)
	}
}

// ExampleCluster демонстрирует работу с кластером
func ExampleCluster() {
	ctx := context.Background()

	// Создаем кластер
	cluster := NewCluster("my_cluster")

	// Добавляем узлы
	cluster.AddNode(&ClusterNode{
		Host:     "node1.example.com",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
		Weight:   1,
	})

	cluster.AddNode(&ClusterNode{
		Host:     "node2.example.com",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
		Weight:   1,
	})

	// Подключаемся к кластеру
	clusterDB, err := ConnectToCluster(cluster, Config{
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Выполняем запрос на кластере
	var users []User
	if err := clusterDB.Query(ctx, &users, "SELECT * FROM users LIMIT 10"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved %d users from cluster\n", len(users))

	// Создаем распределенную таблицу
	if err := clusterDB.CreateDistributedTable(ctx, "users_distributed", "my_cluster", "users", "cityHash64(id)"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created distributed table")
}

// ExampleReplicatedTable демонстрирует создание реплицированной таблицы
func ExampleReplicatedTable() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем реплицированную таблицу
	replicatedTable := NewReplicatedTable("users_replicated", "my_cluster", "test").
		AddColumn("id", "UInt32").
		AddColumn("name", "String").
		AddColumn("email", "String").
		AddColumn("age", "UInt8").
		AddColumn("created", "DateTime").
		SetReplicaName("replica_1").
		SetZooKeeperPath("/clickhouse/tables/users_replicated").
		SetPartitionBy("toYYYYMM(created)").
		SetOrderBy("id").
		SetPrimaryKey("id").
		AddSetting("index_granularity", "8192")

	// Создаем таблицу
	if err := replicatedTable.Create(ctx, db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created replicated table")
}

// ExampleWindowFunctions демонстрирует оконные функции
func ExampleWindowFunctions() {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Используем оконные функции
	query := db.NewQuery().Table("orders")
	window := query.NewWindow().
		RowNumber().
		Over("PARTITION BY user_id", "ORDER BY created DESC").
		As("row_num")

	query = window.AddToQuery().
		Select("user_id", "total", "created").
		Where("row_num <= 3")

	var results []map[string]interface{}
	if err := query.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Top 3 orders per user: %+v\n", results)
}
