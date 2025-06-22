package chorm

import (
	"context"
	"testing"
	"time"
)

// TestUser представляет тестового пользователя
type TestUser struct {
	ID       uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
	Name     string    `ch:"name" ch_type:"String"`
	Email    string    `ch:"email" ch_type:"String"`
	Age      uint8     `ch:"age" ch_type:"UInt8"`
	Created  time.Time `ch:"created" ch_type:"DateTime"`
	IsActive bool      `ch:"is_active" ch_type:"Boolean"`
	Score    float64   `ch:"score" ch_type:"Float64"`
}

// TableName возвращает имя таблицы
func (u *TestUser) TableName() string {
	return "test_users"
}

// TestConnect тестирует подключение к базе данных
func TestConnect(t *testing.T) {
	ctx := context.Background()

	// Пропускаем тест, если нет подключения к ClickHouse
	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Проверяем, что подключение работает
	if err := db.conn.PingContext(ctx); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

// TestCreateTable тестирует создание таблицы
func TestCreateTable(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем тестовую таблицу
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Проверяем, что таблица существует
	var count int64
	err = db.QueryRow(ctx, &count, "SELECT COUNT(*) FROM test_users")
	if err != nil {
		t.Errorf("Failed to query table: %v", err)
	}
}

// TestInsert тестирует вставку данных
func TestInsert(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Вставляем тестового пользователя
	testUser := &TestUser{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Created:  time.Now(),
		IsActive: true,
		Score:    85.5,
	}

	if err := db.Insert(ctx, testUser); err != nil {
		t.Errorf("Failed to insert user: %v", err)
	}

	// Проверяем, что пользователь был вставлен
	var insertedUser TestUser
	err = db.QueryRow(ctx, &insertedUser, "SELECT * FROM test_users WHERE id = ?", 1)
	if err != nil {
		t.Errorf("Failed to query inserted user: %v", err)
	}

	if insertedUser.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", insertedUser.Name)
	}
}

// TestInsertBatch тестирует массовую вставку
func TestInsertBatch(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Создаем тестовых пользователей
	var users []interface{}
	for i := 1; i <= 5; i++ {
		user := &TestUser{
			ID:       uint32(i),
			Name:     "Test User " + string(rune(i+'0')),
			Email:    "test" + string(rune(i+'0')) + "@example.com",
			Age:      uint8(20 + i),
			Created:  time.Now(),
			IsActive: i%2 == 0,
			Score:    float64(70 + i*5),
		}
		users = append(users, user)
	}

	// Вставляем пользователей
	if err := db.InsertBatch(ctx, users); err != nil {
		t.Errorf("Failed to batch insert users: %v", err)
	}

	// Проверяем количество вставленных записей
	var count int64
	err = db.QueryRow(ctx, &count, "SELECT COUNT(*) FROM test_users")
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 users, got %d", count)
	}
}

// TestQuery тестирует выполнение запросов
func TestQuery(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу и вставляем тестовые данные
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	testUser := &TestUser{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Created:  time.Now(),
		IsActive: true,
		Score:    85.5,
	}

	if err := db.Insert(ctx, testUser); err != nil {
		t.Errorf("Failed to insert user: %v", err)
	}

	// Выполняем запрос
	var users []TestUser
	err = db.Query(ctx, &users, "SELECT * FROM test_users WHERE age > ?", 20)
	if err != nil {
		t.Errorf("Failed to query users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0].Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", users[0].Name)
	}
}

// TestQueryBuilder тестирует построитель запросов
func TestQueryBuilder(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу и вставляем тестовые данные
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	testUser := &TestUser{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Created:  time.Now(),
		IsActive: true,
		Score:    85.5,
	}

	if err := db.Insert(ctx, testUser); err != nil {
		t.Errorf("Failed to insert user: %v", err)
	}

	// Используем построитель запросов
	query := db.NewQuery().
		Table("test_users").
		Select("id", "name", "email").
		Where("age > ?", 20).
		Where("is_active = ?", true)

	var users []TestUser
	err = query.All(ctx, &users)
	if err != nil {
		t.Errorf("Failed to execute query: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Тестируем подсчет
	count, err := query.Count(ctx)
	if err != nil {
		t.Errorf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

// TestAggregate тестирует агрегатные функции
func TestAggregate(t *testing.T) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		t.Skipf("Skipping test - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу и вставляем тестовые данные
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Вставляем несколько пользователей
	var users []interface{}
	for i := 1; i <= 3; i++ {
		user := &TestUser{
			ID:       uint32(i),
			Name:     "Test User " + string(rune(i+'0')),
			Email:    "test" + string(rune(i+'0')) + "@example.com",
			Age:      uint8(20 + i*5),
			Created:  time.Now(),
			IsActive: true,
			Score:    float64(70 + i*10),
		}
		users = append(users, user)
	}

	if err := db.InsertBatch(ctx, users); err != nil {
		t.Errorf("Failed to batch insert users: %v", err)
	}

	// Тестируем агрегатные функции
	query := db.NewQuery().Table("test_users")
	agg := query.NewAggregate().
		Count("*").
		Avg("score").
		Max("age").
		Min("age")

	var result map[string]interface{}
	err = agg.Get(ctx, &result)
	if err != nil {
		t.Errorf("Failed to execute aggregate query: %v", err)
	}

	// Проверяем результаты (базовые проверки)
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestMapper тестирует маппер
func TestMapper(t *testing.T) {
	mapper := NewMapper()

	// Тестируем парсинг структуры
	user := &TestUser{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Created:  time.Now(),
		IsActive: true,
		Score:    85.5,
	}

	info, err := mapper.ParseStruct(user)
	if err != nil {
		t.Errorf("Failed to parse struct: %v", err)
	}

	if info.Name != "test_users" {
		t.Errorf("Expected table name 'test_users', got '%s'", info.Name)
	}

	if len(info.Fields) == 0 {
		t.Error("Expected non-empty fields")
	}

	// Тестируем получение значения поля
	value, err := mapper.GetFieldValue(user, "Name")
	if err != nil {
		t.Errorf("Failed to get field value: %v", err)
	}

	if value != "Test User" {
		t.Errorf("Expected field value 'Test User', got '%v'", value)
	}

	// Тестируем установку значения поля
	newUser := &TestUser{}
	err = mapper.SetFieldValue(newUser, "Name", "New User")
	if err != nil {
		t.Errorf("Failed to set field value: %v", err)
	}

	if newUser.Name != "New User" {
		t.Errorf("Expected field value 'New User', got '%s'", newUser.Name)
	}
}

// TestConfig тестирует конфигурацию
func TestConfig(t *testing.T) {
	config := Config{
		Host:            "localhost",
		Port:            9000,
		Database:        "test",
		Username:        "default",
		Password:        "",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		TLS:             false,
		Compression:     true,
		Debug:           true,
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.Host)
	}

	if config.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", config.Port)
	}

	if config.Database != "test" {
		t.Errorf("Expected database 'test', got '%s'", config.Database)
	}
}

// TestTypes тестирует типы данных
func TestTypes(t *testing.T) {
	// Тестируем типы ClickHouse
	if TypeUInt32 != "UInt32" {
		t.Errorf("Expected TypeUInt32 'UInt32', got '%s'", TypeUInt32)
	}

	if TypeString != "String" {
		t.Errorf("Expected TypeString 'String', got '%s'", TypeString)
	}

	if TypeDateTime != "DateTime" {
		t.Errorf("Expected TypeDateTime 'DateTime', got '%s'", TypeDateTime)
	}

	if TypeBoolean != "Boolean" {
		t.Errorf("Expected TypeBoolean 'Boolean', got '%s'", TypeBoolean)
	}

	// Тестируем движки
	if EngineMergeTree != "MergeTree" {
		t.Errorf("Expected EngineMergeTree 'MergeTree', got '%s'", EngineMergeTree)
	}

	if EngineReplacingMergeTree != "ReplacingMergeTree" {
		t.Errorf("Expected EngineReplacingMergeTree 'ReplacingMergeTree', got '%s'", EngineReplacingMergeTree)
	}
}

// BenchmarkInsert тестирует производительность вставки
func BenchmarkInsert(b *testing.B) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		b.Skipf("Skipping benchmark - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		b.Errorf("Failed to create table: %v", err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testUser := &TestUser{
			ID:       uint32(i + 1),
			Name:     "Benchmark User",
			Email:    "benchmark@example.com",
			Age:      25,
			Created:  time.Now(),
			IsActive: true,
			Score:    85.5,
		}

		if err := db.Insert(ctx, testUser); err != nil {
			b.Errorf("Failed to insert user: %v", err)
		}
	}
}

// BenchmarkInsertBatch тестирует производительность массовой вставки
func BenchmarkInsertBatch(b *testing.B) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		b.Skipf("Skipping benchmark - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		b.Errorf("Failed to create table: %v", err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var users []interface{}
		for j := 0; j < 100; j++ {
			testUser := &TestUser{
				ID:       uint32(i*100 + j + 1),
				Name:     "Benchmark User",
				Email:    "benchmark@example.com",
				Age:      25,
				Created:  time.Now(),
				IsActive: true,
				Score:    85.5,
			}
			users = append(users, testUser)
		}

		if err := db.InsertBatch(ctx, users); err != nil {
			b.Errorf("Failed to batch insert users: %v", err)
		}
	}
}

// BenchmarkQuery тестирует производительность запросов
func BenchmarkQuery(b *testing.B) {
	ctx := context.Background()

	db, err := Connect(ctx, Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
	})

	if err != nil {
		b.Skipf("Skipping benchmark - no ClickHouse connection: %v", err)
		return
	}
	defer db.Close()

	// Создаем таблицу и вставляем тестовые данные
	user := &TestUser{}
	if err := db.CreateTable(ctx, user); err != nil {
		b.Errorf("Failed to create table: %v", err)
		return
	}

	// Вставляем тестовые данные
	var users []interface{}
	for i := 0; i < 1000; i++ {
		testUser := &TestUser{
			ID:       uint32(i + 1),
			Name:     "Benchmark User",
			Email:    "benchmark@example.com",
			Age:      25,
			Created:  time.Now(),
			IsActive: true,
			Score:    85.5,
		}
		users = append(users, testUser)
	}

	if err := db.InsertBatch(ctx, users); err != nil {
		b.Errorf("Failed to insert test data: %v", err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result []TestUser
		err := db.Query(ctx, &result, "SELECT * FROM test_users WHERE age > ? LIMIT 100", 20)
		if err != nil {
			b.Errorf("Failed to query users: %v", err)
		}
	}
}
