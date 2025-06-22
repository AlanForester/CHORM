package chorm

import (
	"context"
	"fmt"
	"strings"
)

// Aggregate представляет агрегатную функцию
type Aggregate struct {
	query *Query
	funcs []string
}

// NewAggregate создает новый агрегат
func (q *Query) NewAggregate() *Aggregate {
	return &Aggregate{
		query: q,
		funcs: make([]string, 0),
	}
}

// Sum добавляет функцию SUM
func (a *Aggregate) Sum(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("SUM(%s) as sum_%s", field, field))
	return a
}

// Avg добавляет функцию AVG
func (a *Aggregate) Avg(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("AVG(%s) as avg_%s", field, field))
	return a
}

// Min добавляет функцию MIN
func (a *Aggregate) Min(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("MIN(%s) as min_%s", field, field))
	return a
}

// Max добавляет функцию MAX
func (a *Aggregate) Max(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("MAX(%s) as max_%s", field, field))
	return a
}

// Count добавляет функцию COUNT
func (a *Aggregate) Count(field string) *Aggregate {
	if field == "*" {
		a.funcs = append(a.funcs, "COUNT(*) as count")
	} else {
		a.funcs = append(a.funcs, fmt.Sprintf("COUNT(%s) as count_%s", field, field))
	}
	return a
}

// CountDistinct добавляет функцию COUNT DISTINCT
func (a *Aggregate) CountDistinct(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("COUNT(DISTINCT %s) as count_distinct_%s", field, field))
	return a
}

// Uniq добавляет функцию uniq (ClickHouse специфичная)
func (a *Aggregate) Uniq(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("uniq(%s) as uniq_%s", field, field))
	return a
}

// UniqExact добавляет функцию uniqExact
func (a *Aggregate) UniqExact(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("uniqExact(%s) as uniq_exact_%s", field, field))
	return a
}

// Quantile добавляет функцию quantile
func (a *Aggregate) Quantile(level float64, field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("quantile(%f)(%s) as quantile_%f_%s", level, field, level, field))
	return a
}

// Median добавляет функцию median
func (a *Aggregate) Median(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("median(%s) as median_%s", field, field))
	return a
}

// StdDev добавляет функцию stddev
func (a *Aggregate) StdDev(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("stddev(%s) as stddev_%s", field, field))
	return a
}

// Variance добавляет функцию variance
func (a *Aggregate) Variance(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("varSamp(%s) as variance_%s", field, field))
	return a
}

// Any добавляет функцию any
func (a *Aggregate) Any(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("any(%s) as any_%s", field, field))
	return a
}

// ArgMin добавляет функцию argMin
func (a *Aggregate) ArgMin(arg, val string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("argMin(%s, %s) as argmin_%s_%s", arg, val, arg, val))
	return a
}

// ArgMax добавляет функцию argMax
func (a *Aggregate) ArgMax(arg, val string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("argMax(%s, %s) as argmax_%s_%s", arg, val, arg, val))
	return a
}

// GroupArray добавляет функцию groupArray
func (a *Aggregate) GroupArray(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("groupArray(%s) as group_array_%s", field, field))
	return a
}

// GroupUniqArray добавляет функцию groupUniqArray
func (a *Aggregate) GroupUniqArray(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("groupUniqArray(%s) as group_uniq_array_%s", field, field))
	return a
}

// TopK добавляет функцию topK
func (a *Aggregate) TopK(k int, field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("topK(%d)(%s) as topk_%d_%s", k, field, k, field))
	return a
}

// TopKWeighted добавляет функцию topKWeighted
func (a *Aggregate) TopKWeighted(k int, field, weight string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("topKWeighted(%d)(%s, %s) as topk_weighted_%d_%s_%s", k, field, weight, k, field, weight))
	return a
}

// Histogram добавляет функцию histogram
func (a *Aggregate) Histogram(bins int, field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("histogram(%d)(%s) as histogram_%d_%s", bins, field, bins, field))
	return a
}

// Corr добавляет функцию корреляции
func (a *Aggregate) Corr(x, y string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("corr(%s, %s) as corr_%s_%s", x, y, x, y))
	return a
}

// CovarPop добавляет функцию ковариации
func (a *Aggregate) CovarPop(x, y string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("covarPop(%s, %s) as covar_pop_%s_%s", x, y, x, y))
	return a
}

// CovarSamp добавляет функцию выборочной ковариации
func (a *Aggregate) CovarSamp(x, y string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("covarSamp(%s, %s) as covar_samp_%s_%s", x, y, x, y))
	return a
}

// SkewPop добавляет функцию асимметрии
func (a *Aggregate) SkewPop(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("skewPop(%s) as skew_pop_%s", field, field))
	return a
}

// KurtPop добавляет функцию эксцесса
func (a *Aggregate) KurtPop(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("kurtPop(%s) as kurt_pop_%s", field, field))
	return a
}

// Entropy добавляет функцию энтропии
func (a *Aggregate) Entropy(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("entropy(%s) as entropy_%s", field, field))
	return a
}

// GeometricMean добавляет функцию геометрического среднего
func (a *Aggregate) GeometricMean(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("geometricMean(%s) as geometric_mean_%s", field, field))
	return a
}

// HarmonicMean добавляет функцию гармонического среднего
func (a *Aggregate) HarmonicMean(field string) *Aggregate {
	a.funcs = append(a.funcs, fmt.Sprintf("harmonicMean(%s) as harmonic_mean_%s", field, field))
	return a
}

// Get выполняет агрегатный запрос и возвращает результат
func (a *Aggregate) Get(ctx context.Context, result interface{}) error {
	if len(a.funcs) == 0 {
		return fmt.Errorf("no aggregate functions specified")
	}

	// Устанавливаем SELECT с агрегатными функциями
	a.query.selects = a.funcs

	// Выполняем запрос
	return a.query.Get(ctx, result)
}

// All выполняет агрегатный запрос и возвращает все результаты
func (a *Aggregate) All(ctx context.Context, result interface{}) error {
	if len(a.funcs) == 0 {
		return fmt.Errorf("no aggregate functions specified")
	}

	// Устанавливаем SELECT с агрегатными функциями
	a.query.selects = a.funcs

	// Выполняем запрос
	return a.query.All(ctx, result)
}

// Window представляет оконную функцию
type Window struct {
	query    *Query
	function string
	over     string
	alias    string
}

// NewWindow создает новую оконную функцию
func (q *Query) NewWindow() *Window {
	return &Window{
		query: q,
	}
}

// RowNumber добавляет ROW_NUMBER()
func (w *Window) RowNumber() *Window {
	w.function = "ROW_NUMBER()"
	return w
}

// Rank добавляет RANK()
func (w *Window) Rank() *Window {
	w.function = "RANK()"
	return w
}

// DenseRank добавляет DENSE_RANK()
func (w *Window) DenseRank() *Window {
	w.function = "DENSE_RANK()"
	return w
}

// Lag добавляет LAG()
func (w *Window) Lag(field string, offset int) *Window {
	w.function = fmt.Sprintf("LAG(%s, %d)", field, offset)
	return w
}

// Lead добавляет LEAD()
func (w *Window) Lead(field string, offset int) *Window {
	w.function = fmt.Sprintf("LEAD(%s, %d)", field, offset)
	return w
}

// FirstValue добавляет FIRST_VALUE()
func (w *Window) FirstValue(field string) *Window {
	w.function = fmt.Sprintf("FIRST_VALUE(%s)", field)
	return w
}

// LastValue добавляет LAST_VALUE()
func (w *Window) LastValue(field string) *Window {
	w.function = fmt.Sprintf("LAST_VALUE(%s)", field)
	return w
}

// NthValue добавляет NTH_VALUE()
func (w *Window) NthValue(field string, n int) *Window {
	w.function = fmt.Sprintf("NTH_VALUE(%s, %d)", field, n)
	return w
}

// Ntile добавляет NTILE()
func (w *Window) Ntile(buckets int) *Window {
	w.function = fmt.Sprintf("NTILE(%d)", buckets)
	return w
}

// PercentRank добавляет PERCENT_RANK()
func (w *Window) PercentRank() *Window {
	w.function = "PERCENT_RANK()"
	return w
}

// CumeDist добавляет CUME_DIST()
func (w *Window) CumeDist() *Window {
	w.function = "CUME_DIST()"
	return w
}

// Over устанавливает OVER clause
func (w *Window) Over(partitionBy, orderBy string) *Window {
	var parts []string

	if partitionBy != "" {
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", partitionBy))
	}

	if orderBy != "" {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", orderBy))
	}

	w.over = fmt.Sprintf("OVER (%s)", strings.Join(parts, " "))
	return w
}

// As устанавливает алиас
func (w *Window) As(alias string) *Window {
	w.alias = alias
	return w
}

// Build строит оконную функцию
func (w *Window) Build() string {
	if w.function == "" {
		return ""
	}

	result := w.function
	if w.over != "" {
		result += " " + w.over
	}

	if w.alias != "" {
		result += " AS " + w.alias
	}

	return result
}

// AddToQuery добавляет оконную функцию к запросу
func (w *Window) AddToQuery() *Query {
	if w.function == "" {
		return w.query
	}

	windowFunc := w.Build()
	if windowFunc != "" {
		w.query.selects = append(w.query.selects, windowFunc)
	}

	return w.query
}
