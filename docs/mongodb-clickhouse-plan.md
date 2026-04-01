# stardust.mini — MongoDB & ClickHouse 接入开发计划

## 背景

`stardust.mini` 已具备 MySQL/Postgres（`databases/`）和 Redis（`redis/`）的接入能力。
本计划在此基础上新增两个独立包：

- **`mongodb/`**：提供 MongoDB 连接与集合操作能力
- **`clickhouse/`**：提供 ClickHouse 连接与写入/查询能力

两个包均遵循库内现有模式：`Init(config []byte)` → `Get*Manager()` → 接口操作。

---

## 基础设施信息（docker-dev-env）

| 服务 | 容器名 | 端口 | 用户 | 密码 | 默认库 |
|------|--------|------|------|------|--------|
| MongoDB | `drama-mongodb` | `27017` | `admin` | `drama_mongo_123` | `drama_dev` |
| ClickHouse | `drama-clickhouse` | HTTP `8123` / Native `9000` | `drama_user` | `drama_click_123` | `drama_analytics` |

Docker 网络内 DNS：`mongodb:27017`、`clickhouse:8123`

---

## 一、MongoDB 接入方案

### 1.1 目录结构

```
stardust.mini/mongodb/
├── mongodb.go          # Config 结构、Init()、GetMongoManager()
├── manager.go          # MongoManager（多库管理，线程安全）
├── client.go           # MongoCli 接口定义
├── client_impl.go      # mongoClient 实现
└── errors.go           # 错误常量
```

### 1.2 配置结构

```go
// mongodb.go
type Config struct {
    Name     string `json:"name"`      // 逻辑名称，用于多库区分
    URI      string `json:"uri"`       // mongodb://user:pass@host:port/dbname
    Database string `json:"database"`  // 默认操作的数据库名
    MaxPool  uint64 `json:"max_pool"`  // 最大连接池大小，默认 10
    MinPool  uint64 `json:"min_pool"`  // 最小连接池大小，默认 2
    TimeoutS int    `json:"timeout_s"` // 操作超时秒数，默认 5
}
```

configTest.toml 配置示例：
```toml
[mongodb]
name     = "default"
uri      = "mongodb://admin:drama_mongo_123@mongodb:27017/drama_dev?authSource=admin"
database = "drama_dev"
max_pool = 10
min_pool = 2
timeout_s = 5
```

### 1.3 接口设计（MongoCli）

```go
type MongoCli interface {
    // 文档操作
    InsertOne(ctx context.Context, collection string, doc interface{}) (string, error)
    InsertMany(ctx context.Context, collection string, docs []interface{}) ([]string, error)
    FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error
    FindMany(ctx context.Context, collection string, filter interface{}, results interface{}, opts ...*options.FindOptions) error
    UpdateOne(ctx context.Context, collection string, filter, update interface{}) (int64, error)
    UpdateMany(ctx context.Context, collection string, filter, update interface{}) (int64, error)
    DeleteOne(ctx context.Context, collection string, filter interface{}) (int64, error)
    DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error)
    CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error)

    // 原生集合（用于复杂聚合等）
    Collection(name string) *mongo.Collection

    // 健康检查
    Ping(ctx context.Context) error
}
```

### 1.4 Manager 初始化模式

```go
// 与 databases 包保持一致的初始化模式
func Init(config []byte) { ... }
func GetMongoManager() *MongoManager { ... }

// 使用示例
cli := mongodb.GetMongoManager().GetClient("default")
id, err := cli.InsertOne(ctx, "events", doc)
```

### 1.5 依赖

```
go.mongodb.org/mongo-driver v1.x
```

---

## 二、ClickHouse 接入方案

### 2.1 目录结构

```
stardust.mini/clickhouse/
├── clickhouse.go       # Config 结构、Init()、GetClickHouseManager()
├── manager.go          # ClickHouseManager（多实例管理，线程安全）
├── client.go           # ClickHouseCli 接口定义
├── client_impl.go      # clickhouseClient 实现
└── errors.go           # 错误常量
```

### 2.2 配置结构

```go
// clickhouse.go
type Config struct {
    Name     string `json:"name"`      // 逻辑名称
    Addr     string `json:"addr"`      // host:port（Native 协议 9000）
    Database string `json:"database"`
    Username string `json:"username"`
    Password string `json:"password"`
    MaxConn  int    `json:"max_conn"`  // 最大连接数，默认 10
    MaxIdle  int    `json:"max_idle"`  // 最大空闲连接数，默认 5
    DialTimeoutS int `json:"dial_timeout_s"` // 默认 5
}
```

configTest.toml 配置示例：
```toml
[clickhouse]
name     = "default"
addr     = "clickhouse:9000"
database = "drama_analytics"
username = "drama_user"
password = "drama_click_123"
max_conn = 10
max_idle = 5
dial_timeout_s = 5
```

### 2.3 接口设计（ClickHouseCli）

```go
type ClickHouseCli interface {
    // 写入
    Exec(ctx context.Context, query string, args ...interface{}) error
    AsyncInsert(ctx context.Context, query string, wait bool) error

    // 查询
    Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error

    // 批量写入（高性能场景）
    PrepareBatch(ctx context.Context, query string) (driver.Batch, error)

    // 原生连接（用于复杂场景）
    Conn() driver.Conn

    // 健康检查
    Ping(ctx context.Context) error
}
```

### 2.4 Manager 初始化模式

```go
func Init(config []byte) { ... }
func GetClickHouseManager() *ClickHouseManager { ... }

// 使用示例
cli := clickhouse.GetClickHouseManager().GetClient("default")
err := cli.Exec(ctx, "INSERT INTO play_events VALUES (?, ?, ?)", uid, vid, ts)
```

### 2.5 依赖

```
github.com/ClickHouse/clickhouse-go/v2 v2.x
```

---

## 三、实现里程碑

### M1 — MongoDB 基础实现

- [ ] 新建 `mongodb/` 包
- [ ] 实现 `Config` 结构及 `SetDefaults()`、`Validate()`
- [ ] 实现 `Init(config []byte)` 和 `GetMongoManager()`（单例，线程安全）
- [ ] 实现 `MongoCli` 接口及 `mongoClient` 全部方法
- [ ] 实现 `errors.go`（`ErrNotFound`、`ErrInsertFailed` 等）
- [ ] 补充单元测试（Mock + 集成测试）

### M2 — ClickHouse 基础实现

- [ ] 新建 `clickhouse/` 包
- [ ] 实现 `Config` 结构及 `SetDefaults()`、`Validate()`
- [ ] 实现 `Init(config []byte)` 和 `GetClickHouseManager()`（单例，线程安全）
- [ ] 实现 `ClickHouseCli` 接口及 `clickhouseClient` 全部方法
- [ ] 实现批量写入（`PrepareBatch`）支持
- [ ] 实现 `errors.go`
- [ ] 补充单元测试

### M3 — 集成验证

- [ ] 在 `analyticsServer` 中接入 ClickHouse，写入播放事件
- [ ] 在 `commentServer` 或 `danmakuServer` 中接入 MongoDB，存储评论/弹幕文档
- [ ] 更新各服务 `configTest.toml`，补充 `[mongodb]` / `[clickhouse]` 节
- [ ] 本地 docker-dev-env 联调验证（`docker-compose-huawei.yaml` 已含两个服务）

---

## 四、设计原则

1. **接口隔离**：业务代码只依赖 `MongoCli` / `ClickHouseCli` 接口，不直接引用驱动类型（便于 Mock 测试）。
2. **单例 + 多实例**：`Manager` 以 `sync.Once` 保证单例初始化，内部以 `name` 为 key 支持多数据库实例。
3. **Context 贯通**：所有操作方法首参数均为 `context.Context`，支持超时和 tracing 透传。
4. **配置兼容**：`Init(config []byte)` 同时支持单个配置对象和数组两种 JSON 格式，与 `databases.Init` 保持一致。
5. **最小依赖**：MongoDB 使用官方 `mongo-driver`；ClickHouse 使用 `clickhouse-go/v2`（Native 协议），不引入冗余驱动。
