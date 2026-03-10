# stardust.mini

Go 微服务框架库，提供构建微服务应用所需的核心功能。

## 功能模块

| 模块 | 说明 |
|------|------|
| `conf` | TOML 配置文件管理 |
| `databases` | 基于 GORM 的数据库 ORM 层，支持 MySQL/PostgreSQL |
| `errors` | 错误处理框架，支持堆栈跟踪和 Try-Catch 机制 |
| `http_server` | 基于 Echo v4 的 HTTP 服务器 |
| `jwt` | JWT 认证（HMAC-SHA256） |
| `logs` | 基于 Zap 的日志系统，支持日志轮转 |
| `nats` | NATS 消息队列客户端，支持 JetStream |
| `redis` | Redis 客户端，支持多种数据结构 |
| `uuid` | 基于 Snowflake 算法的分布式 ID 生成器 |
| `utils` | HTTP 请求、文件下载等工具函数 |

## 安装

```bash
go get github.com/jxncyjq/stardust.mini
```

## 主要依赖

- `gorm.io/gorm` - ORM 框架
- `github.com/labstack/echo/v4` - HTTP 框架
- `github.com/redis/go-redis/v9` - Redis 客户端
- `github.com/nats-io/nats.go` - NATS 客户端
- `go.uber.org/zap` - 日志库

## 要求

- Go 1.25+