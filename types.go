package chorm

import (
	"database/sql"
	"time"
)

// Config представляет конфигурацию подключения к ClickHouse
type Config struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	TLS             bool
	Compression     bool
	Debug           bool
}

// DB представляет основное соединение с ClickHouse
type DB struct {
	conn   *sql.DB
	config Config
}

// QueryBuilder представляет построитель запросов
type QueryBuilder struct {
	table   string
	selects []string
	wheres  []string
	groupBy []string
	orderBy []string
	limit   int
	offset  int
	args    []interface{}
}

// Model представляет интерфейс для моделей
type Model interface {
	TableName() string
}

// FieldInfo содержит информацию о поле структуры
type FieldInfo struct {
	Name     string
	Type     string
	Tag      string
	IsPK     bool
	IsAuto   bool
	Nullable bool
}

// TableInfo содержит информацию о таблице
type TableInfo struct {
	Name    string
	Fields  []FieldInfo
	Engine  string
	Options map[string]string
}

// ClickHouseType представляет типы данных ClickHouse
type ClickHouseType string

const (
	// Базовые типы
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

	// Сложные типы
	TypeArray          ClickHouseType = "Array"
	TypeNullable       ClickHouseType = "Nullable"
	TypeLowCardinality ClickHouseType = "LowCardinality"
	TypeEnum           ClickHouseType = "Enum"
	TypeNested         ClickHouseType = "Nested"
	TypeTuple          ClickHouseType = "Tuple"
	TypeMap            ClickHouseType = "Map"
)

// Engine представляет движки таблиц ClickHouse
type Engine string

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

// Error представляет ошибку ORM
type Error struct {
	Code    int
	Message string
	Query   string
}

func (e *Error) Error() string {
	return e.Message
}

// Result представляет результат выполнения запроса
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Row представляет строку результата
type Row struct {
	values map[string]interface{}
}

// Get возвращает значение по ключу
func (r *Row) Get(key string) interface{} {
	return r.values[key]
}

// GetString возвращает строковое значение
func (r *Row) GetString(key string) string {
	if v, ok := r.values[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt возвращает целочисленное значение
func (r *Row) GetInt(key string) int64 {
	if v, ok := r.values[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int32:
			return int64(val)
		case int16:
			return int64(val)
		case int8:
			return int64(val)
		case uint64:
			return int64(val)
		case uint32:
			return int64(val)
		case uint16:
			return int64(val)
		case uint8:
			return int64(val)
		}
	}
	return 0
}

// GetFloat возвращает значение с плавающей точкой
func (r *Row) GetFloat(key string) float64 {
	if v, ok := r.values[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		}
	}
	return 0.0
}

// GetBool возвращает булево значение
func (r *Row) GetBool(key string) bool {
	if v, ok := r.values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetTime возвращает временное значение
func (r *Row) GetTime(key string) time.Time {
	if v, ok := r.values[key]; ok {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}
