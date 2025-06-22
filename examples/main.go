package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AlanForester/chorm"
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

func main() {
	ctx := context.Background()

	fmt.Println("🚀 CHORM - ClickHouse ORM для Go")
	fmt.Println("==================================")

	// Подключаемся к ClickHouse
	fmt.Println("\n📡 Подключение к ClickHouse...")
	db, err := chorm.Connect(ctx, chorm.Config{
		Host:     "localhost",
		Port:     9000,
		Database: "test",
		Username: "default",
		Password: "",
		Debug:    true,
	})
	if err != nil {
		log.Printf("❌ Ошибка подключения: %v", err)
		log.Println("💡 Убедитесь, что ClickHouse запущен на localhost:9000")
		return
	}
	defer db.Close()
	fmt.Println("✅ Подключение установлено")

	// Создаем таблицы
	fmt.Println("\n🏗️ Создание таблиц...")
	if err := db.CreateTable(ctx, &User{}); err != nil {
		log.Printf("❌ Ошибка создания таблицы users: %v", err)
		return
	}
	if err := db.CreateTable(ctx, &Order{}); err != nil {
		log.Printf("❌ Ошибка создания таблицы orders: %v", err)
		return
	}
	fmt.Println("✅ Таблицы созданы")

	// Вставляем тестовые данные
	fmt.Println("\n📝 Вставка тестовых данных...")

	// Пользователи
	users := []*User{
		{
			ID:       1,
			Name:     "Иван Иванов",
			Email:    "ivan@example.com",
			Age:      25,
			Created:  time.Now(),
			Updated:  time.Now(),
			IsActive: true,
			Score:    85.5,
		},
		{
			ID:       2,
			Name:     "Мария Петрова",
			Email:    "maria@example.com",
			Age:      30,
			Created:  time.Now(),
			Updated:  time.Now(),
			IsActive: true,
			Score:    92.3,
		},
		{
			ID:       3,
			Name:     "Алексей Сидоров",
			Email:    "alex@example.com",
			Age:      28,
			Created:  time.Now(),
			Updated:  time.Now(),
			IsActive: false,
			Score:    78.9,
		},
	}

	for _, user := range users {
		if err := db.Insert(ctx, user); err != nil {
			log.Printf("❌ Ошибка вставки пользователя: %v", err)
			return
		}
	}

	// Заказы
	orders := []*Order{
		{
			ID:        1,
			UserID:    1,
			ProductID: 101,
			Quantity:  2,
			Price:     1500.0,
			Total:     3000.0,
			Status:    "completed",
			Created:   time.Now().Add(-24 * time.Hour),
			Completed: time.Now().Add(-23 * time.Hour),
		},
		{
			ID:        2,
			UserID:    2,
			ProductID: 102,
			Quantity:  1,
			Price:     2500.0,
			Total:     2500.0,
			Status:    "pending",
			Created:   time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        3,
			UserID:    1,
			ProductID: 103,
			Quantity:  3,
			Price:     800.0,
			Total:     2400.0,
			Status:    "completed",
			Created:   time.Now().Add(-6 * time.Hour),
			Completed: time.Now().Add(-5 * time.Hour),
		},
	}

	for _, order := range orders {
		if err := db.Insert(ctx, order); err != nil {
			log.Printf("❌ Ошибка вставки заказа: %v", err)
			return
		}
	}

	fmt.Println("✅ Тестовые данные вставлены")

	// Демонстрация построителя запросов
	fmt.Println("\n🔍 Демонстрация построителя запросов...")

	// Поиск активных пользователей старше 25 лет
	query := db.NewQuery().
		Table("users").
		Select("id", "name", "email", "age", "score").
		Where("age > ?", 25).
		Where("is_active = ?", true).
		OrderBy("score", "DESC")

	var activeUsers []User
	if err := query.All(ctx, &activeUsers); err != nil {
		log.Printf("❌ Ошибка запроса: %v", err)
		return
	}

	fmt.Printf("👥 Активные пользователи старше 25 лет (%d):\n", len(activeUsers))
	for _, user := range activeUsers {
		fmt.Printf("  - %s (%s), возраст: %d, рейтинг: %.1f\n",
			user.Name, user.Email, user.Age, user.Score)
	}

	// Подсчет общего количества пользователей
	count, err := query.Count(ctx)
	if err != nil {
		log.Printf("❌ Ошибка подсчета: %v", err)
		return
	}
	fmt.Printf("📊 Общее количество активных пользователей старше 25: %d\n", count)

	// Демонстрация агрегатных функций
	fmt.Println("\n📊 Демонстрация агрегатных функций...")

	aggQuery := db.NewQuery().Table("orders")
	agg := aggQuery.NewAggregate().
		Count("*").
		Sum("total").
		Avg("total").
		Max("total").
		Min("total")

	var aggResult map[string]interface{}
	if err := agg.Get(ctx, &aggResult); err != nil {
		log.Printf("❌ Ошибка агрегации: %v", err)
		return
	}

	fmt.Println("📈 Статистика по заказам:")
	fmt.Printf("  - Общее количество заказов: %v\n", aggResult["count"])
	fmt.Printf("  - Общая сумма: %.2f\n", aggResult["sum_total"])
	fmt.Printf("  - Средняя сумма заказа: %.2f\n", aggResult["avg_total"])
	fmt.Printf("  - Максимальная сумма: %.2f\n", aggResult["max_total"])
	fmt.Printf("  - Минимальная сумма: %.2f\n", aggResult["min_total"])

	// Группировка по пользователям
	fmt.Println("\n👤 Статистика по пользователям:")
	groupQuery := db.NewQuery().
		Table("orders").
		Select("user_id", "COUNT(*) as order_count", "SUM(total) as total_spent").
		GroupBy("user_id").
		OrderBy("total_spent", "DESC")

	var userStats []map[string]interface{}
	if err := groupQuery.All(ctx, &userStats); err != nil {
		log.Printf("❌ Ошибка группировки: %v", err)
		return
	}

	for _, stat := range userStats {
		fmt.Printf("  - Пользователь ID %v: %v заказов, общая сумма %.2f\n",
			stat["user_id"], stat["order_count"], stat["total_spent"])
	}

	// Демонстрация пагинации
	fmt.Println("\n📄 Демонстрация пагинации...")

	paginateQuery := db.NewQuery().
		Table("users").
		Select("id", "name", "email", "score").
		OrderBy("score", "DESC")

	var pageUsers []User
	total, err := paginateQuery.Paginate(ctx, 1, 2, &pageUsers)
	if err != nil {
		log.Printf("❌ Ошибка пагинации: %v", err)
		return
	}

	fmt.Printf("📖 Страница 1 (показано %d из %d пользователей):\n", len(pageUsers), total)
	for _, user := range pageUsers {
		fmt.Printf("  - %s (%s), рейтинг: %.1f\n", user.Name, user.Email, user.Score)
	}

	// Демонстрация JOIN запросов
	fmt.Println("\n🔗 Демонстрация JOIN запросов...")

	joinQuery := db.NewQuery().
		Table("orders").
		Select("orders.id", "users.name", "orders.total", "orders.status").
		Join("users", "orders.user_id = users.id").
		Where("orders.status = ?", "completed").
		OrderBy("orders.total", "DESC")

	var joinResults []map[string]interface{}
	if err := joinQuery.All(ctx, &joinResults); err != nil {
		log.Printf("❌ Ошибка JOIN запроса: %v", err)
		return
	}

	fmt.Println("🛒 Завершенные заказы:")
	for _, result := range joinResults {
		fmt.Printf("  - Заказ #%v: %s, сумма %.2f, статус: %s\n",
			result["id"], result["name"], result["total"], result["status"])
	}

	// Демонстрация обновления данных
	fmt.Println("\n✏️ Демонстрация обновления данных...")

	updateQuery := db.NewQuery().
		Table("users").
		Where("id = ?", 3)

	result, err := updateQuery.Update(ctx, map[string]interface{}{
		"is_active": true,
		"score":     85.0,
		"updated":   time.Now(),
	})
	if err != nil {
		log.Printf("❌ Ошибка обновления: %v", err)
		return
	}

	fmt.Printf("✅ Обновлено записей: %d\n", result.RowsAffected)

	// Проверяем обновление
	var updatedUser User
	err = db.QueryRow(ctx, &updatedUser, "SELECT * FROM users WHERE id = ?", 3)
	if err != nil {
		log.Printf("❌ Ошибка проверки обновления: %v", err)
		return
	}

	fmt.Printf("👤 Обновленный пользователь: %s, активен: %v, рейтинг: %.1f\n",
		updatedUser.Name, updatedUser.IsActive, updatedUser.Score)

	// Демонстрация удаления данных
	fmt.Println("\n🗑️ Демонстрация удаления данных...")

	deleteQuery := db.NewQuery().
		Table("orders").
		Where("status = ?", "pending")

	deleteResult, err := deleteQuery.Delete(ctx)
	if err != nil {
		log.Printf("❌ Ошибка удаления: %v", err)
		return
	}

	fmt.Printf("✅ Удалено заказов со статусом 'pending': %d\n", deleteResult.RowsAffected)

	// Финальная статистика
	fmt.Println("\n📊 Финальная статистика:")

	var finalUserCount int64
	err = db.QueryRow(ctx, &finalUserCount, "SELECT COUNT(*) FROM users")
	if err != nil {
		log.Printf("❌ Ошибка подсчета пользователей: %v", err)
		return
	}

	var finalOrderCount int64
	err = db.QueryRow(ctx, &finalOrderCount, "SELECT COUNT(*) FROM orders")
	if err != nil {
		log.Printf("❌ Ошибка подсчета заказов: %v", err)
		return
	}

	fmt.Printf("👥 Пользователей в базе: %d\n", finalUserCount)
	fmt.Printf("🛒 Заказов в базе: %d\n", finalOrderCount)

	fmt.Println("\n🎉 Демонстрация завершена!")
	fmt.Println("📚 Подробная документация: https://github.com/AlanForester/chorm")
}
