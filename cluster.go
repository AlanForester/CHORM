package chorm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ClusterNode представляет узел кластера
type ClusterNode struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Weight   int // Вес для балансировки
	Healthy  bool
	LastPing time.Time
}

// Cluster представляет кластер ClickHouse
type Cluster struct {
	Name  string
	Nodes []*ClusterNode
	mu    sync.RWMutex
}

// NewCluster создает новый кластер
func NewCluster(name string) *Cluster {
	return &Cluster{
		Name:  name,
		Nodes: make([]*ClusterNode, 0),
	}
}

// AddNode добавляет узел в кластер
func (c *Cluster) AddNode(node *ClusterNode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Nodes = append(c.Nodes, node)
}

// RemoveNode удаляет узел из кластера
func (c *Cluster) RemoveNode(host string, port int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, node := range c.Nodes {
		if node.Host == host && node.Port == port {
			c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
			break
		}
	}
}

// GetHealthyNodes возвращает здоровые узлы
func (c *Cluster) GetHealthyNodes() []*ClusterNode {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var healthy []*ClusterNode
	for _, node := range c.Nodes {
		if node.Healthy {
			healthy = append(healthy, node)
		}
	}
	return healthy
}

// GetNodeByWeight возвращает узел по весу (для балансировки)
func (c *Cluster) GetNodeByWeight() *ClusterNode {
	healthy := c.GetHealthyNodes()
	if len(healthy) == 0 {
		return nil
	}

	// Простая реализация round-robin с учетом веса
	// В реальном проекте можно использовать более сложные алгоритмы
	totalWeight := 0
	for _, node := range healthy {
		totalWeight += node.Weight
	}

	if totalWeight == 0 {
		return healthy[0]
	}

	// Выбираем узел на основе веса
	currentWeight := 0
	for _, node := range healthy {
		currentWeight += node.Weight
		if currentWeight >= totalWeight/2 {
			return node
		}
	}

	return healthy[0]
}

// HealthCheck проверяет здоровье узлов кластера
func (c *Cluster) HealthCheck(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.Nodes {
		// Создаем временное подключение для проверки
		config := Config{
			Host:     node.Host,
			Port:     node.Port,
			Database: node.Database,
			Username: node.Username,
			Password: node.Password,
		}

		db, err := Connect(ctx, config)
		if err != nil {
			node.Healthy = false
			continue
		}

		// Проверяем подключение
		if err := db.conn.PingContext(ctx); err != nil {
			node.Healthy = false
		} else {
			node.Healthy = true
			node.LastPing = time.Now()
		}

		db.Close()
	}
}

// ClusterDB представляет подключение к кластеру
type ClusterDB struct {
	cluster *Cluster
	config  Config
}

// NewClusterDB создает новое подключение к кластеру
func NewClusterDB(cluster *Cluster, config Config) *ClusterDB {
	return &ClusterDB{
		cluster: cluster,
		config:  config,
	}
}

// ConnectToCluster подключается к кластеру
func ConnectToCluster(cluster *Cluster, config Config) (*ClusterDB, error) {
	// Проверяем здоровье кластера
	cluster.HealthCheck(context.Background())

	healthy := cluster.GetHealthyNodes()
	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy nodes in cluster")
	}

	return &ClusterDB{
		cluster: cluster,
		config:  config,
	}, nil
}

// GetConnection получает подключение к случайному здоровому узлу
func (cdb *ClusterDB) GetConnection(ctx context.Context) (*DB, error) {
	node := cdb.cluster.GetNodeByWeight()
	if node == nil {
		return nil, fmt.Errorf("no available nodes in cluster")
	}

	config := Config{
		Host:     node.Host,
		Port:     node.Port,
		Database: node.Database,
		Username: node.Username,
		Password: node.Password,
	}

	return Connect(ctx, config)
}

// Query выполняет запрос на случайном узле кластера
func (cdb *ClusterDB) Query(ctx context.Context, result interface{}, query string, args ...interface{}) error {
	db, err := cdb.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Query(ctx, result, query, args...)
}

// Exec выполняет команду на случайном узле кластера
func (cdb *ClusterDB) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	db, err := cdb.GetConnection(ctx)
	if err != nil {
		return Result{}, err
	}
	defer db.Close()

	return db.Exec(ctx, query, args...)
}

// CreateDistributedTable создает распределенную таблицу
func (cdb *ClusterDB) CreateDistributedTable(ctx context.Context, tableName, clusterName, localTableName string, shardingKey string) error {
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s AS %s
		ENGINE = Distributed(%s, %s, %s, %s)
	`, tableName, localTableName, clusterName, cdb.config.Database, localTableName, shardingKey)

	_, err := cdb.Exec(ctx, sql)
	return err
}

// InsertIntoDistributed вставляет данные в распределенную таблицу
func (cdb *ClusterDB) InsertIntoDistributed(ctx context.Context, tableName string, data interface{}) error {
	db, err := cdb.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Insert(ctx, data)
}

// ReplicatedTable представляет реплицированную таблицу
type ReplicatedTable struct {
	Name          string
	ClusterName   string
	Database      string
	Engine        string
	ReplicaPath   string
	ReplicaName   string
	ZooKeeperPath string
	Columns       []string
	PartitionBy   string
	OrderBy       string
	PrimaryKey    string
	SampleBy      string
	TTL           string
	Settings      map[string]string
}

// NewReplicatedTable создает новую реплицированную таблицу
func NewReplicatedTable(name, clusterName, database string) *ReplicatedTable {
	return &ReplicatedTable{
		Name:        name,
		ClusterName: clusterName,
		Database:    database,
		Engine:      "ReplicatedMergeTree",
		Columns:     make([]string, 0),
		Settings:    make(map[string]string),
	}
}

// AddColumn добавляет колонку в реплицированную таблицу
func (rt *ReplicatedTable) AddColumn(name, dataType string) *ReplicatedTable {
	rt.Columns = append(rt.Columns, fmt.Sprintf("%s %s", name, dataType))
	return rt
}

// SetReplicaPath устанавливает путь реплики
func (rt *ReplicatedTable) SetReplicaPath(path string) *ReplicatedTable {
	rt.ReplicaPath = path
	return rt
}

// SetReplicaName устанавливает имя реплики
func (rt *ReplicatedTable) SetReplicaName(name string) *ReplicatedTable {
	rt.ReplicaName = name
	return rt
}

// SetZooKeeperPath устанавливает путь в ZooKeeper
func (rt *ReplicatedTable) SetZooKeeperPath(path string) *ReplicatedTable {
	rt.ZooKeeperPath = path
	return rt
}

// SetPartitionBy устанавливает PARTITION BY
func (rt *ReplicatedTable) SetPartitionBy(expr string) *ReplicatedTable {
	rt.PartitionBy = expr
	return rt
}

// SetOrderBy устанавливает ORDER BY
func (rt *ReplicatedTable) SetOrderBy(expr string) *ReplicatedTable {
	rt.OrderBy = expr
	return rt
}

// SetPrimaryKey устанавливает PRIMARY KEY
func (rt *ReplicatedTable) SetPrimaryKey(expr string) *ReplicatedTable {
	rt.PrimaryKey = expr
	return rt
}

// SetSampleBy устанавливает SAMPLE BY
func (rt *ReplicatedTable) SetSampleBy(expr string) *ReplicatedTable {
	rt.SampleBy = expr
	return rt
}

// SetTTL устанавливает TTL
func (rt *ReplicatedTable) SetTTL(expr string) *ReplicatedTable {
	rt.TTL = expr
	return rt
}

// AddSetting добавляет настройку
func (rt *ReplicatedTable) AddSetting(key, value string) *ReplicatedTable {
	rt.Settings[key] = value
	return rt
}

// BuildCreateSQL строит SQL для создания реплицированной таблицы
func (rt *ReplicatedTable) BuildCreateSQL() string {
	var parts []string

	// CREATE TABLE
	parts = append(parts, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (", rt.Database, rt.Name))

	// Columns
	parts = append(parts, "  "+strings.Join(rt.Columns, ",\n  "))
	parts = append(parts, ")")

	// Engine
	engine := fmt.Sprintf("ENGINE = %s", rt.Engine)

	// Engine parameters
	var engineParams []string
	if rt.ZooKeeperPath != "" {
		engineParams = append(engineParams, fmt.Sprintf("'%s'", rt.ZooKeeperPath))
	}
	if rt.ReplicaName != "" {
		engineParams = append(engineParams, fmt.Sprintf("'%s'", rt.ReplicaName))
	}

	if len(engineParams) > 0 {
		engine += fmt.Sprintf("(%s)", strings.Join(engineParams, ", "))
	}
	parts = append(parts, engine)

	// PARTITION BY
	if rt.PartitionBy != "" {
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", rt.PartitionBy))
	}

	// ORDER BY
	if rt.OrderBy != "" {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", rt.OrderBy))
	}

	// PRIMARY KEY
	if rt.PrimaryKey != "" {
		parts = append(parts, fmt.Sprintf("PRIMARY KEY %s", rt.PrimaryKey))
	}

	// SAMPLE BY
	if rt.SampleBy != "" {
		parts = append(parts, fmt.Sprintf("SAMPLE BY %s", rt.SampleBy))
	}

	// TTL
	if rt.TTL != "" {
		parts = append(parts, fmt.Sprintf("TTL %s", rt.TTL))
	}

	// SETTINGS
	if len(rt.Settings) > 0 {
		var settings []string
		for k, v := range rt.Settings {
			settings = append(settings, fmt.Sprintf("%s = %s", k, v))
		}
		parts = append(parts, fmt.Sprintf("SETTINGS %s", strings.Join(settings, ", ")))
	}

	return strings.Join(parts, "\n")
}

// Create создает реплицированную таблицу
func (rt *ReplicatedTable) Create(ctx context.Context, db *DB) error {
	sql := rt.BuildCreateSQL()
	_, err := db.Exec(ctx, sql)
	return err
}

// ShardManager представляет менеджер шардов
type ShardManager struct {
	cluster *Cluster
}

// NewShardManager создает новый менеджер шардов
func NewShardManager(cluster *Cluster) *ShardManager {
	return &ShardManager{
		cluster: cluster,
	}
}

// GetShardInfo получает информацию о шардах
func (sm *ShardManager) GetShardInfo(ctx context.Context) (map[string]interface{}, error) {
	// Подключаемся к любому узлу кластера
	node := sm.cluster.GetNodeByWeight()
	if node == nil {
		return nil, fmt.Errorf("no available nodes")
	}

	config := Config{
		Host:     node.Host,
		Port:     node.Port,
		Database: node.Database,
		Username: node.Username,
		Password: node.Password,
	}

	db, err := Connect(ctx, config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var result []map[string]interface{}
	err = db.Query(ctx, &result, "SELECT * FROM system.clusters")
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return result[0], nil
	}
	return nil, fmt.Errorf("no cluster information found")
}

// GetShardNodes получает узлы шарда
func (sm *ShardManager) GetShardNodes(ctx context.Context, clusterName string) ([]map[string]interface{}, error) {
	node := sm.cluster.GetNodeByWeight()
	if node == nil {
		return nil, fmt.Errorf("no available nodes")
	}

	config := Config{
		Host:     node.Host,
		Port:     node.Port,
		Database: node.Database,
		Username: node.Username,
		Password: node.Password,
	}

	db, err := Connect(ctx, config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var result []map[string]interface{}
	err = db.Query(ctx, &result, "SELECT * FROM system.clusters WHERE cluster = ?", clusterName)
	return result, err
}

// BalanceLoad балансирует нагрузку между шардами
func (sm *ShardManager) BalanceLoad(ctx context.Context) error {
	// Здесь можно реализовать логику балансировки нагрузки
	// Например, перераспределение данных между шардами

	// Проверяем здоровье всех узлов
	sm.cluster.HealthCheck(ctx)

	// Получаем статистику по шардам
	healthy := sm.cluster.GetHealthyNodes()
	if len(healthy) == 0 {
		return fmt.Errorf("no healthy nodes for load balancing")
	}

	// Простая реализация - в реальном проекте можно использовать более сложные алгоритмы
	fmt.Printf("Load balancing across %d healthy nodes\n", len(healthy))

	return nil
}
