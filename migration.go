package chorm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Migration представляет миграцию
type Migration struct {
	ID        int64     `ch:"id" ch_type:"UInt64"`
	Name      string    `ch:"name" ch_type:"String"`
	AppliedAt time.Time `ch:"applied_at" ch_type:"DateTime"`
	Checksum  string    `ch:"checksum" ch_type:"String"`
}

// TableName возвращает имя таблицы для миграций
func (m *Migration) TableName() string {
	return "migrations"
}

// MigrationFunc представляет функцию миграции
type MigrationFunc func(ctx context.Context, db *DB) error

// MigrationRecord представляет запись о миграции
type MigrationRecord struct {
	Name     string
	Up       MigrationFunc
	Down     MigrationFunc
	Checksum string
}

// Migrator представляет мигратор
type Migrator struct {
	db         *DB
	migrations []MigrationRecord
}

// NewMigrator создает новый мигратор
func NewMigrator(db *DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]MigrationRecord, 0),
	}
}

// AddMigration добавляет миграцию
func (m *Migrator) AddMigration(name string, up, down MigrationFunc) *Migrator {
	checksum := generateChecksum(name)
	m.migrations = append(m.migrations, MigrationRecord{
		Name:     name,
		Up:       up,
		Down:     down,
		Checksum: checksum,
	})
	return m
}

// CreateMigrationsTable создает таблицу для отслеживания миграций
func (m *Migrator) CreateMigrationsTable(ctx context.Context) error {
	return m.db.CreateTable(ctx, &Migration{})
}

// GetAppliedMigrations получает список примененных миграций
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	var migrations []Migration
	err := m.db.Query(ctx, &migrations, "SELECT * FROM migrations ORDER BY id")
	return migrations, err
}

// IsMigrationApplied проверяет, применена ли миграция
func (m *Migrator) IsMigrationApplied(ctx context.Context, name string) (bool, error) {
	var count int64
	err := m.db.QueryRow(ctx, &count, "SELECT COUNT(*) FROM migrations WHERE name = ?", name)
	return count > 0, err
}

// ApplyMigration применяет миграцию
func (m *Migrator) ApplyMigration(ctx context.Context, migration MigrationRecord) error {
	// Проверяем, не применена ли уже миграция
	applied, err := m.IsMigrationApplied(ctx, migration.Name)
	if err != nil {
		return fmt.Errorf("failed to check if migration is applied: %w", err)
	}

	if applied {
		return fmt.Errorf("migration %s is already applied", migration.Name)
	}

	// Начинаем транзакцию
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Выполняем миграцию
	if err := migration.Up(ctx, m.db); err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
	}

	// Записываем информацию о миграции
	_, err = tx.Exec(ctx,
		"INSERT INTO migrations (name, applied_at, checksum) VALUES (?, ?, ?)",
		migration.Name, time.Now(), migration.Checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Подтверждаем транзакцию
	return tx.Commit()
}

// RollbackMigration откатывает миграцию
func (m *Migrator) RollbackMigration(ctx context.Context, name string) error {
	// Проверяем, применена ли миграция
	applied, err := m.IsMigrationApplied(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if migration is applied: %w", err)
	}

	if !applied {
		return fmt.Errorf("migration %s is not applied", name)
	}

	// Находим миграцию
	var migration MigrationRecord
	for _, m := range m.migrations {
		if m.Name == name {
			migration = m
			break
		}
	}

	if migration.Name == "" {
		return fmt.Errorf("migration %s not found", name)
	}

	// Начинаем транзакцию
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Выполняем откат
	if migration.Down != nil {
		if err := migration.Down(ctx, m.db); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Name, err)
		}
	}

	// Удаляем запись о миграции
	_, err = tx.Exec(ctx, "DELETE FROM migrations WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Подтверждаем транзакцию
	return tx.Commit()
}

// Migrate применяет все непримененные миграции
func (m *Migrator) Migrate(ctx context.Context) error {
	// Создаем таблицу миграций, если она не существует
	if err := m.CreateMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем примененные миграции
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Создаем карту примененных миграций
	appliedMap := make(map[string]bool)
	for _, migration := range applied {
		appliedMap[migration.Name] = true
	}

	// Применяем непримененные миграции
	for _, migration := range m.migrations {
		if !appliedMap[migration.Name] {
			if err := m.ApplyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
			}
			fmt.Printf("Applied migration: %s\n", migration.Name)
		}
	}

	return nil
}

// Rollback откатывает последнюю миграцию
func (m *Migrator) Rollback(ctx context.Context) error {
	// Получаем примененные миграции
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Откатываем последнюю миграцию
	lastMigration := applied[len(applied)-1]
	return m.RollbackMigration(ctx, lastMigration.Name)
}

// Status показывает статус миграций
func (m *Migrator) Status(ctx context.Context) error {
	// Создаем таблицу миграций, если она не существует
	if err := m.CreateMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем примененные миграции
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Создаем карту примененных миграций
	appliedMap := make(map[string]Migration)
	for _, migration := range applied {
		appliedMap[migration.Name] = migration
	}

	fmt.Println("Migration Status:")
	fmt.Println("==================")

	for _, migration := range m.migrations {
		if applied, exists := appliedMap[migration.Name]; exists {
			fmt.Printf("✓ %s (applied at %s)\n", migration.Name, applied.AppliedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("✗ %s (pending)\n", migration.Name)
		}
	}

	return nil
}

// generateChecksum генерирует контрольную сумму для миграции
func generateChecksum(name string) string {
	// Простая реализация - в реальном проекте можно использовать более сложные алгоритмы
	return fmt.Sprintf("%d", len(name))
}

// Schema представляет схему базы данных
type Schema struct {
	db *DB
}

// NewSchema создает новый объект схемы
func NewSchema(db *DB) *Schema {
	return &Schema{db: db}
}

// CreateDatabase создает базу данных
func (s *Schema) CreateDatabase(ctx context.Context, name string) error {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", name)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// DropDatabase удаляет базу данных
func (s *Schema) DropDatabase(ctx context.Context, name string) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// CreateTable создает таблицу
func (s *Schema) CreateTable(ctx context.Context, tableName string, columns []string, engine string, options map[string]string) error {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n) ENGINE = %s",
		tableName, strings.Join(columns, ",\n  "), engine)

	if len(options) > 0 {
		var opts []string
		for k, v := range options {
			opts = append(opts, fmt.Sprintf("%s = %s", k, v))
		}
		sql += fmt.Sprintf("(%s)", strings.Join(opts, ", "))
	}

	_, err := s.db.Exec(ctx, sql)
	return err
}

// DropTable удаляет таблицу
func (s *Schema) DropTable(ctx context.Context, tableName string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// TruncateTable очищает таблицу
func (s *Schema) TruncateTable(ctx context.Context, tableName string) error {
	sql := fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// RenameTable переименовывает таблицу
func (s *Schema) RenameTable(ctx context.Context, oldName, newName string) error {
	sql := fmt.Sprintf("RENAME TABLE %s TO %s", oldName, newName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// AddColumn добавляет колонку
func (s *Schema) AddColumn(ctx context.Context, tableName, columnName, columnType string) error {
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// DropColumn удаляет колонку
func (s *Schema) DropColumn(ctx context.Context, tableName, columnName string) error {
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, columnName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// ModifyColumn изменяет тип колонки
func (s *Schema) ModifyColumn(ctx context.Context, tableName, columnName, newType string) error {
	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", tableName, columnName, newType)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// RenameColumn переименовывает колонку
func (s *Schema) RenameColumn(ctx context.Context, tableName, oldName, newName string) error {
	sql := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, oldName, newName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// CreateIndex создает индекс
func (s *Schema) CreateIndex(ctx context.Context, indexName, tableName string, columns []string) error {
	sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, strings.Join(columns, ", "))
	_, err := s.db.Exec(ctx, sql)
	return err
}

// DropIndex удаляет индекс
func (s *Schema) DropIndex(ctx context.Context, indexName, tableName string) error {
	sql := fmt.Sprintf("DROP INDEX %s ON %s", indexName, tableName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// CreateMaterializedView создает материализованное представление
func (s *Schema) CreateMaterializedView(ctx context.Context, viewName, tableName, selectQuery string) error {
	sql := fmt.Sprintf("CREATE MATERIALIZED VIEW %s TO %s AS %s", viewName, tableName, selectQuery)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// DropMaterializedView удаляет материализованное представление
func (s *Schema) DropMaterializedView(ctx context.Context, viewName string) error {
	sql := fmt.Sprintf("DROP VIEW IF EXISTS %s", viewName)
	_, err := s.db.Exec(ctx, sql)
	return err
}

// GetTableInfo получает информацию о таблице
func (s *Schema) GetTableInfo(ctx context.Context, tableName string) (map[string]interface{}, error) {
	var result []map[string]interface{}
	err := s.db.Query(ctx, &result, "DESCRIBE TABLE "+tableName)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0], nil
	}
	return nil, fmt.Errorf("table %s not found", tableName)
}

// GetTables получает список таблиц
func (s *Schema) GetTables(ctx context.Context) ([]string, error) {
	var tables []string
	err := s.db.Query(ctx, &tables, "SHOW TABLES")
	return tables, err
}

// GetDatabases получает список баз данных
func (s *Schema) GetDatabases(ctx context.Context) ([]string, error) {
	var databases []string
	err := s.db.Query(ctx, &databases, "SHOW DATABASES")
	return databases, err
}
