package chorm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Connect создает подключение к ClickHouse
func Connect(ctx context.Context, config Config) (*DB, error) {
	if config.Port == 0 {
		config.Port = 9000
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 10
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}

	// Создаем DSN для подключения
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=10s&max_execution_time=60",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.TLS {
		dsn += "&secure=true"
	}

	if config.Compression {
		dsn += "&compress=true"
	}

	// Подключаемся к базе данных
	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Настраиваем пул соединений
	conn.SetMaxOpenConns(config.MaxOpenConns)
	conn.SetMaxIdleConns(config.MaxIdleConns)
	conn.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Проверяем подключение
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &DB{
		conn:   conn,
		config: config,
	}, nil
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	return db.conn.Close()
}

// CreateTable создает таблицу на основе структуры
func (db *DB) CreateTable(ctx context.Context, model interface{}) error {
	mapper := NewMapper()
	info, err := mapper.ParseStruct(model)
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	sql := mapper.BuildCreateTableSQL(info)

	if db.config.Debug {
		fmt.Printf("Creating table with SQL: %s\n", sql)
	}

	_, err = db.conn.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Insert вставляет одну запись
func (db *DB) Insert(ctx context.Context, model interface{}) error {
	mapper := NewMapper()
	info, err := mapper.ParseStruct(model)
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	// Получаем значения полей
	var columns []string
	var values []interface{}
	var placeholders []string

	for _, field := range info.Fields {
		value, err := mapper.GetFieldValue(model, field.Name)
		if err != nil {
			continue // Пропускаем поля, которые не удалось получить
		}

		columns = append(columns, fmt.Sprintf("`%s`", field.Name))
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		info.Name, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	if db.config.Debug {
		fmt.Printf("Insert SQL: %s\n", sql)
		fmt.Printf("Values: %v\n", values)
	}

	_, err = db.conn.ExecContext(ctx, sql, values...)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	return nil
}

// InsertBatch вставляет множество записей
func (db *DB) InsertBatch(ctx context.Context, models []interface{}) error {
	if len(models) == 0 {
		return nil
	}

	mapper := NewMapper()
	info, err := mapper.ParseStruct(models[0])
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	// Получаем колонки из первой модели
	var columns []string
	for _, field := range info.Fields {
		columns = append(columns, fmt.Sprintf("`%s`", field.Name))
	}

	// Строим SQL для batch insert
	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES ",
		info.Name, strings.Join(columns, ", "))

	var allValues []interface{}
	var valueGroups []string

	for _, model := range models {
		var values []interface{}
		var placeholders []string

		for _, field := range info.Fields {
			value, err := mapper.GetFieldValue(model, field.Name)
			if err != nil {
				value = nil // Используем NULL для недоступных полей
			}
			values = append(values, value)
			placeholders = append(placeholders, "?")
		}

		valueGroups = append(valueGroups, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		allValues = append(allValues, values...)
	}

	sql += strings.Join(valueGroups, ", ")

	if db.config.Debug {
		fmt.Printf("Batch Insert SQL: %s\n", sql)
	}

	_, err = db.conn.ExecContext(ctx, sql, allValues...)
	if err != nil {
		return fmt.Errorf("failed to batch insert records: %w", err)
	}

	return nil
}

// Query выполняет запрос и заполняет результат в slice
func (db *DB) Query(ctx context.Context, result interface{}, query string, args ...interface{}) error {
	if db.config.Debug {
		fmt.Printf("Query SQL: %s\n", query)
		fmt.Printf("Args: %v\n", args)
	}

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return db.scanRows(rows, result)
}

// QueryRow выполняет запрос и возвращает одну строку
func (db *DB) QueryRow(ctx context.Context, result interface{}, query string, args ...interface{}) error {
	if db.config.Debug {
		fmt.Printf("QueryRow SQL: %s\n", query)
		fmt.Printf("Args: %v\n", args)
	}

	row := db.conn.QueryRowContext(ctx, query, args...)
	return db.scanRow(row, result)
}

// Exec выполняет запрос без возврата результата
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	if db.config.Debug {
		fmt.Printf("Exec SQL: %s\n", query)
		fmt.Printf("Args: %v\n", args)
	}

	result, err := db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, fmt.Errorf("failed to execute query: %w", err)
	}

	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return Result{
		LastInsertID: lastInsertID,
		RowsAffected: rowsAffected,
	}, nil
}

// scanRows сканирует результаты запроса в slice структур
func (db *DB) scanRows(rows *sql.Rows, result interface{}) error {
	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr || resultVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to slice")
	}

	sliceVal := resultVal.Elem()
	elementType := sliceVal.Type().Elem()

	// Получаем колонки
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Создаем слайс для значений
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Сканируем каждую строку
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Создаем новый элемент
		element := reflect.New(elementType).Elem()

		// Заполняем элемент значениями
		for i, column := range columns {
			if i < len(values) {
				db.setFieldValue(element, column, values[i])
			}
		}

		// Добавляем элемент в slice
		sliceVal.Set(reflect.Append(sliceVal, element))
	}

	return rows.Err()
}

// scanRow сканирует одну строку результата
func (db *DB) scanRow(row *sql.Row, result interface{}) error {
	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	// Получаем тип результата
	resultType := resultVal.Type().Elem()
	if resultType.Kind() != reflect.Struct {
		return fmt.Errorf("result must be a pointer to struct")
	}

	// Создаем временную структуру для получения колонок
	temp := reflect.New(resultType).Interface()
	mapper := NewMapper()
	info, err := mapper.ParseStruct(temp)
	if err != nil {
		return fmt.Errorf("failed to parse struct: %w", err)
	}

	// Создаем слайс для значений
	values := make([]interface{}, len(info.Fields))
	valuePtrs := make([]interface{}, len(values))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Сканируем строку
	err = row.Scan(valuePtrs...)
	if err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Заполняем результат
	element := resultVal.Elem()
	for i, field := range info.Fields {
		if i < len(values) {
			db.setFieldValue(element, field.Name, values[i])
		}
	}

	return nil
}

// setFieldValue устанавливает значение поля в структуре
func (db *DB) setFieldValue(element reflect.Value, fieldName string, value interface{}) {
	field := element.FieldByName(fieldName)
	if !field.IsValid() || !field.CanSet() {
		return
	}

	// Конвертируем значение в нужный тип
	fieldType := field.Type()

	switch fieldType.Kind() {
	case reflect.String:
		if value != nil {
			field.SetString(fmt.Sprintf("%v", value))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value != nil {
			switch v := value.(type) {
			case int64:
				field.SetInt(v)
			case int32:
				field.SetInt(int64(v))
			case int16:
				field.SetInt(int64(v))
			case int8:
				field.SetInt(int64(v))
			case uint64:
				field.SetInt(int64(v))
			case uint32:
				field.SetInt(int64(v))
			case uint16:
				field.SetInt(int64(v))
			case uint8:
				field.SetInt(int64(v))
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value != nil {
			switch v := value.(type) {
			case uint64:
				field.SetUint(v)
			case uint32:
				field.SetUint(uint64(v))
			case uint16:
				field.SetUint(uint64(v))
			case uint8:
				field.SetUint(uint64(v))
			case int64:
				field.SetUint(uint64(v))
			case int32:
				field.SetUint(uint64(v))
			case int16:
				field.SetUint(uint64(v))
			case int8:
				field.SetUint(uint64(v))
			}
		}
	case reflect.Float32, reflect.Float64:
		if value != nil {
			switch v := value.(type) {
			case float64:
				field.SetFloat(v)
			case float32:
				field.SetFloat(float64(v))
			}
		}
	case reflect.Bool:
		if value != nil {
			if b, ok := value.(bool); ok {
				field.SetBool(b)
			}
		}
	}
}

// Begin начинает транзакцию
func (db *DB) Begin(ctx context.Context) (*Tx, error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Tx{tx: tx, db: db}, nil
}

// Tx представляет транзакцию
type Tx struct {
	tx *sql.Tx
	db *DB
}

// Commit подтверждает транзакцию
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback откатывает транзакцию
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Exec выполняет запрос в транзакции
func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, fmt.Errorf("failed to execute query in transaction: %w", err)
	}

	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return Result{
		LastInsertID: lastInsertID,
		RowsAffected: rowsAffected,
	}, nil
}
