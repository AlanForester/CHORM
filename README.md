# CHORM - ClickHouse ORM для Go

CHORM - это мощная ORM библиотека для работы с ClickHouse в Go, предоставляющая удобный интерфейс для маппинга структур, выполнения CRUD операций и построения сложных аналитических запросов.

## Возможности

- 🗂️ **Маппинг структур** - Автоматическое преобразование Go структур в таблицы ClickHouse
- 🔄 **CRUD операции** - Полная поддержка Create, Read, Update, Delete операций
- 📊 **Аналитические функции** - Поддержка агрегатных функций и сложных запросов
- 🚀 **Производительность** - Оптимизированные массовые вставки и пул соединений
- 🏗️ **Управление схемой** - Автоматическое создание таблиц и миграции
- 🔒 **Безопасность** - Параметризованные запросы и TLS поддержка
- 🎯 **Кластерная поддержка** - Работа с распределенными таблицами

## Установка

```bash
go get github.com/AlanForester/chorm
```

## Быстрый старт

```go
package main

import (
    "context"
    "log"
    
    "github.com/AlanForester/chorm"
)

// Определяем структуру для маппинга
type User struct {
    ID       uint32 `ch:"id" ch_type:"UInt32"`
    Name     string `ch:"name" ch_type:"String"`
    Email    string `ch:"email" ch_type:"String"`
    Age      uint8  `ch:"age" ch_type:"UInt8"`
    Created  string `ch:"created" ch_type:"DateTime"`
}

func main() {
    // Подключаемся к ClickHouse
    db, err := chorm.Connect(context.Background(), chorm.Config{
        Host:     "localhost:9000",
        Database: "test",
        Username: "default",
        Password: "",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Создаем таблицу автоматически
    err = db.CreateTable(context.Background(), &User{})
    if err != nil {
        log.Fatal(err)
    }

    // Вставляем данные
    user := &User{
        ID:      1,
        Name:    "John Doe",
        Email:   "john@example.com",
        Age:     30,
        Created: "2024-01-01 10:00:00",
    }
    
    err = db.Insert(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }

    // Выполняем запрос
    var users []User
    err = db.Query(context.Background(), &users, "SELECT * FROM users WHERE age > ?", 25)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d users", len(users))
}
```

## Документация

Полная документация доступна в [docs/](docs/) директории.

## Лицензия

MIT License - см. файл [LICENSE](LICENSE) для деталей. 