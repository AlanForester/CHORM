package chorm

import (
	"context"
	"fmt"
	"strings"
)

// Query представляет построитель запросов
type Query struct {
	db       *DB
	table    string
	selects  []string
	wheres   []string
	groupBy  []string
	orderBy  []string
	limit    int
	offset   int
	args     []interface{}
	distinct bool
	having   []string
	joins    []string
}

// NewQuery создает новый построитель запросов
func (db *DB) NewQuery() *Query {
	return &Query{
		db:      db,
		selects: []string{"*"},
		args:    make([]interface{}, 0),
	}
}

// Table устанавливает таблицу для запроса
func (q *Query) Table(table string) *Query {
	q.table = table
	return q
}

// Select устанавливает поля для выборки
func (q *Query) Select(fields ...string) *Query {
	if len(fields) > 0 {
		q.selects = fields
	}
	return q
}

// Distinct добавляет DISTINCT к запросу
func (q *Query) Distinct() *Query {
	q.distinct = true
	return q
}

// Where добавляет условие WHERE
func (q *Query) Where(condition string, args ...interface{}) *Query {
	q.wheres = append(q.wheres, condition)
	q.args = append(q.args, args...)
	return q
}

// WhereIn добавляет условие WHERE IN
func (q *Query) WhereIn(field string, values []interface{}) *Query {
	if len(values) == 0 {
		return q
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	q.wheres = append(q.wheres, condition)
	q.args = append(q.args, values...)
	return q
}

// WhereNotIn добавляет условие WHERE NOT IN
func (q *Query) WhereNotIn(field string, values []interface{}) *Query {
	if len(values) == 0 {
		return q
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
	q.wheres = append(q.wheres, condition)
	q.args = append(q.args, values...)
	return q
}

// WhereBetween добавляет условие WHERE BETWEEN
func (q *Query) WhereBetween(field string, start, end interface{}) *Query {
	condition := fmt.Sprintf("%s BETWEEN ? AND ?", field)
	q.wheres = append(q.wheres, condition)
	q.args = append(q.args, start, end)
	return q
}

// WhereLike добавляет условие WHERE LIKE
func (q *Query) WhereLike(field, pattern string) *Query {
	condition := fmt.Sprintf("%s LIKE ?", field)
	q.wheres = append(q.wheres, condition)
	q.args = append(q.args, pattern)
	return q
}

// WhereNull добавляет условие WHERE IS NULL
func (q *Query) WhereNull(field string) *Query {
	condition := fmt.Sprintf("%s IS NULL", field)
	q.wheres = append(q.wheres, condition)
	return q
}

// WhereNotNull добавляет условие WHERE IS NOT NULL
func (q *Query) WhereNotNull(field string) *Query {
	condition := fmt.Sprintf("%s IS NOT NULL", field)
	q.wheres = append(q.wheres, condition)
	return q
}

// Join добавляет JOIN
func (q *Query) Join(table, condition string, args ...interface{}) *Query {
	join := fmt.Sprintf("JOIN %s ON %s", table, condition)
	q.joins = append(q.joins, join)
	q.args = append(q.args, args...)
	return q
}

// LeftJoin добавляет LEFT JOIN
func (q *Query) LeftJoin(table, condition string, args ...interface{}) *Query {
	join := fmt.Sprintf("LEFT JOIN %s ON %s", table, condition)
	q.joins = append(q.joins, join)
	q.args = append(q.args, args...)
	return q
}

// RightJoin добавляет RIGHT JOIN
func (q *Query) RightJoin(table, condition string, args ...interface{}) *Query {
	join := fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition)
	q.joins = append(q.joins, join)
	q.args = append(q.args, args...)
	return q
}

// GroupBy добавляет GROUP BY
func (q *Query) GroupBy(fields ...string) *Query {
	q.groupBy = append(q.groupBy, fields...)
	return q
}

// Having добавляет HAVING
func (q *Query) Having(condition string, args ...interface{}) *Query {
	q.having = append(q.having, condition)
	q.args = append(q.args, args...)
	return q
}

// OrderBy добавляет ORDER BY
func (q *Query) OrderBy(field string, direction ...string) *Query {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	q.orderBy = append(q.orderBy, fmt.Sprintf("%s %s", field, dir))
	return q
}

// OrderByAsc добавляет ORDER BY ASC
func (q *Query) OrderByAsc(field string) *Query {
	q.orderBy = append(q.orderBy, fmt.Sprintf("%s ASC", field))
	return q
}

// OrderByDesc добавляет ORDER BY DESC
func (q *Query) OrderByDesc(field string) *Query {
	q.orderBy = append(q.orderBy, fmt.Sprintf("%s DESC", field))
	return q
}

// Limit устанавливает LIMIT
func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

// Offset устанавливает OFFSET
func (q *Query) Offset(offset int) *Query {
	q.offset = offset
	return q
}

// buildSQL строит SQL запрос
func (q *Query) buildSQL() string {
	var parts []string

	// SELECT
	selectClause := "SELECT "
	if q.distinct {
		selectClause += "DISTINCT "
	}
	selectClause += strings.Join(q.selects, ", ")
	parts = append(parts, selectClause)

	// FROM
	if q.table != "" {
		parts = append(parts, fmt.Sprintf("FROM %s", q.table))
	}

	// JOIN
	if len(q.joins) > 0 {
		parts = append(parts, strings.Join(q.joins, " "))
	}

	// WHERE
	if len(q.wheres) > 0 {
		parts = append(parts, fmt.Sprintf("WHERE %s", strings.Join(q.wheres, " AND ")))
	}

	// GROUP BY
	if len(q.groupBy) > 0 {
		parts = append(parts, fmt.Sprintf("GROUP BY %s", strings.Join(q.groupBy, ", ")))
	}

	// HAVING
	if len(q.having) > 0 {
		parts = append(parts, fmt.Sprintf("HAVING %s", strings.Join(q.having, " AND ")))
	}

	// ORDER BY
	if len(q.orderBy) > 0 {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(q.orderBy, ", ")))
	}

	// LIMIT
	if q.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", q.limit))
	}

	// OFFSET
	if q.offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", q.offset))
	}

	return strings.Join(parts, " ")
}

// Get выполняет запрос и возвращает одну запись
func (q *Query) Get(ctx context.Context, result interface{}) error {
	q.limit = 1
	sql := q.buildSQL()

	if q.db.config.Debug {
		fmt.Printf("Get SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", q.args)
	}

	return q.db.QueryRow(ctx, result, sql, q.args...)
}

// All выполняет запрос и возвращает все записи
func (q *Query) All(ctx context.Context, result interface{}) error {
	sql := q.buildSQL()

	if q.db.config.Debug {
		fmt.Printf("All SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", q.args)
	}

	return q.db.Query(ctx, result, sql, q.args...)
}

// Count выполняет запрос COUNT
func (q *Query) Count(ctx context.Context) (int64, error) {
	// Сохраняем оригинальные selects
	originalSelects := q.selects
	q.selects = []string{"COUNT(*)"}

	sql := q.buildSQL()

	if q.db.config.Debug {
		fmt.Printf("Count SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", q.args)
	}

	var count int64
	err := q.db.QueryRow(ctx, &count, sql, q.args...)

	// Восстанавливаем оригинальные selects
	q.selects = originalSelects

	return count, err
}

// Exists проверяет существование записей
func (q *Query) Exists(ctx context.Context) (bool, error) {
	q.selects = []string{"1"}
	q.limit = 1

	sql := q.buildSQL()

	if q.db.config.Debug {
		fmt.Printf("Exists SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", q.args)
	}

	var exists int
	err := q.db.QueryRow(ctx, &exists, sql, q.args...)

	return err == nil, err
}

// First выполняет запрос и возвращает первую запись
func (q *Query) First(ctx context.Context, result interface{}) error {
	q.limit = 1
	return q.Get(ctx, result)
}

// Last выполняет запрос и возвращает последнюю запись
func (q *Query) Last(ctx context.Context, result interface{}) error {
	// Сохраняем оригинальный orderBy
	originalOrderBy := q.orderBy

	// Если нет ORDER BY, добавляем по первичному ключу
	if len(q.orderBy) == 0 {
		// Здесь можно добавить логику для определения первичного ключа
		q.orderBy = []string{"id DESC"}
	} else {
		// Инвертируем существующий ORDER BY
		var invertedOrderBy []string
		for _, order := range q.orderBy {
			if strings.Contains(order, "ASC") {
				invertedOrderBy = append(invertedOrderBy, strings.Replace(order, "ASC", "DESC", 1))
			} else if strings.Contains(order, "DESC") {
				invertedOrderBy = append(invertedOrderBy, strings.Replace(order, "DESC", "ASC", 1))
			} else {
				invertedOrderBy = append(invertedOrderBy, order+" DESC")
			}
		}
		q.orderBy = invertedOrderBy
	}

	q.limit = 1
	err := q.Get(ctx, result)

	// Восстанавливаем оригинальный orderBy
	q.orderBy = originalOrderBy

	return err
}

// Paginate выполняет пагинацию
func (q *Query) Paginate(ctx context.Context, page, perPage int, result interface{}) (int64, error) {
	// Получаем общее количество записей
	total, err := q.Count(ctx)
	if err != nil {
		return 0, err
	}

	// Вычисляем offset
	offset := (page - 1) * perPage

	// Устанавливаем limit и offset
	q.limit = perPage
	q.offset = offset

	// Выполняем запрос
	err = q.All(ctx, result)

	return total, err
}

// Update выполняет UPDATE запрос
func (q *Query) Update(ctx context.Context, data map[string]interface{}) (Result, error) {
	if len(data) == 0 {
		return Result{}, fmt.Errorf("no data to update")
	}

	var sets []string
	var args []interface{}

	for field, value := range data {
		sets = append(sets, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	// Добавляем аргументы WHERE
	args = append(args, q.args...)

	sql := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(sets, ", "))

	if len(q.wheres) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(q.wheres, " AND "))
	}

	if q.db.config.Debug {
		fmt.Printf("Update SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", args)
	}

	return q.db.Exec(ctx, sql, args...)
}

// Delete выполняет DELETE запрос
func (q *Query) Delete(ctx context.Context) (Result, error) {
	sql := fmt.Sprintf("DELETE FROM %s", q.table)

	if len(q.wheres) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(q.wheres, " AND "))
	}

	if q.db.config.Debug {
		fmt.Printf("Delete SQL: %s\n", sql)
		fmt.Printf("Args: %v\n", q.args)
	}

	return q.db.Exec(ctx, sql, q.args...)
}
