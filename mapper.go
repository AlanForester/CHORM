package chorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Mapper представляет маппер для работы со структурами
type Mapper struct {
	registry map[string]*TableInfo
}

// NewMapper создает новый маппер
func NewMapper() *Mapper {
	return &Mapper{
		registry: make(map[string]*TableInfo),
	}
}

// ParseStruct парсит структуру и возвращает информацию о таблице
func (m *Mapper) ParseStruct(model interface{}) (*TableInfo, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct, got %s", val.Kind())
	}

	typ := val.Type()
	tableName := m.getTableName(model, typ)

	// Проверяем кэш
	if info, exists := m.registry[tableName]; exists {
		return info, nil
	}

	info := &TableInfo{
		Name:    tableName,
		Fields:  make([]FieldInfo, 0),
		Engine:  string(EngineMergeTree),
		Options: make(map[string]string),
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldInfo, err := m.parseField(field)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %w", field.Name, err)
		}

		if fieldInfo.Name != "" {
			info.Fields = append(info.Fields, fieldInfo)
		}
	}

	// Кэшируем результат
	m.registry[tableName] = info

	return info, nil
}

// parseField парсит отдельное поле структуры
func (m *Mapper) parseField(field reflect.StructField) (FieldInfo, error) {
	info := FieldInfo{
		Name: field.Name,
		Type: string(TypeString), // По умолчанию
	}

	// Парсим тег ch
	if tag := field.Tag.Get("ch"); tag != "" {
		info.Name = tag
	}

	// Парсим тип ClickHouse
	if chType := field.Tag.Get("ch_type"); chType != "" {
		info.Type = chType
	} else {
		// Автоматическое определение типа
		info.Type = m.goTypeToClickHouseType(field.Type)
	}

	// Проверяем дополнительные опции
	if field.Tag.Get("ch_pk") == "true" {
		info.IsPK = true
	}

	if field.Tag.Get("ch_auto") == "true" {
		info.IsAuto = true
	}

	if field.Tag.Get("ch_nullable") == "true" {
		info.Nullable = true
	}

	// Парсим движок таблицы
	if engine := field.Tag.Get("ch_engine"); engine != "" {
		// Это должно быть на уровне структуры, но для простоты обрабатываем здесь
	}

	return info, nil
}

// goTypeToClickHouseType конвертирует Go тип в тип ClickHouse
func (m *Mapper) goTypeToClickHouseType(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return string(TypeBoolean)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return string(TypeInt32)
	case reflect.Int64:
		return string(TypeInt64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return string(TypeUInt32)
	case reflect.Uint64:
		return string(TypeUInt64)
	case reflect.Float32:
		return string(TypeFloat32)
	case reflect.Float64:
		return string(TypeFloat64)
	case reflect.String:
		return string(TypeString)
	case reflect.Slice, reflect.Array:
		elemType := m.goTypeToClickHouseType(typ.Elem())
		return fmt.Sprintf("Array(%s)", elemType)
	case reflect.Struct:
		// Проверяем специальные типы
		if typ.String() == "time.Time" {
			return string(TypeDateTime)
		}
		return string(TypeString) // По умолчанию
	default:
		return string(TypeString)
	}
}

// getTableName получает имя таблицы из модели
func (m *Mapper) getTableName(model interface{}, typ reflect.Type) string {
	// Проверяем, реализует ли модель интерфейс Model
	if modelWithTable, ok := model.(Model); ok {
		return modelWithTable.TableName()
	}

	// Проверяем тег на уровне структуры
	if tag := typ.Field(0).Tag.Get("ch_table"); tag != "" {
		return tag
	}

	// Используем имя типа в нижнем регистре
	return strings.ToLower(typ.Name())
}

// GetFieldValue получает значение поля из структуры
func (m *Mapper) GetFieldValue(model interface{}, fieldName string) (interface{}, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct")
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	return field.Interface(), nil
}

// SetFieldValue устанавливает значение поля в структуре
func (m *Mapper) SetFieldValue(model interface{}, fieldName string, value interface{}) error {
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer to struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a pointer to struct")
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s is not settable", fieldName)
	}

	// Конвертируем значение в нужный тип
	fieldType := field.Type()
	valueType := reflect.TypeOf(value)

	if fieldType == valueType {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	// Простые конвертации
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
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
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(i)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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
		case string:
			if i, err := strconv.ParseUint(v, 10, 64); err == nil {
				field.SetUint(i)
			}
		}
	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case float32:
			field.SetFloat(float64(v))
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				field.SetFloat(f)
			}
		}
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				field.SetBool(b)
			}
		}
	}

	return nil
}

// GetPrimaryKey получает первичный ключ из структуры
func (m *Mapper) GetPrimaryKey(model interface{}) (string, interface{}, error) {
	info, err := m.ParseStruct(model)
	if err != nil {
		return "", nil, err
	}

	for _, field := range info.Fields {
		if field.IsPK {
			value, err := m.GetFieldValue(model, field.Name)
			return field.Name, value, err
		}
	}

	return "", nil, fmt.Errorf("no primary key found")
}

// BuildCreateTableSQL строит SQL для создания таблицы
func (m *Mapper) BuildCreateTableSQL(info *TableInfo) string {
	var columns []string

	for _, field := range info.Fields {
		columnDef := fmt.Sprintf("`%s` %s", field.Name, field.Type)

		if field.IsPK {
			columnDef += " PRIMARY KEY"
		}

		if field.IsAuto {
			columnDef += " AUTO_INCREMENT"
		}

		columns = append(columns, columnDef)
	}

	engine := info.Engine
	if engine == "" {
		engine = string(EngineMergeTree)
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  %s\n) ENGINE = %s",
		info.Name, strings.Join(columns, ",\n  "), engine)

	// Добавляем опции движка
	if len(info.Options) > 0 {
		var options []string
		for k, v := range info.Options {
			options = append(options, fmt.Sprintf("%s = %s", k, v))
		}
		sql += fmt.Sprintf("(%s)", strings.Join(options, ", "))
	}

	return sql
}
