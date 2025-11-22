# 故障排除指南

本文档记录了常见问题及其解决方案。

## 数据库连接问题

### 问题：database "trpg" does not exist

**症状**：
```
postgres-1  | FATAL:  database "trpg" does not exist
```

**根本原因**：
环境变量名称不匹配导致应用无法正确读取数据库配置。

**详细分析**：

1. **Viper配置系统的环境变量映射规则**：
   ```go
   viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
   ```
   这意味着配置键 `database.dbname` 会映射到环境变量 `DATABASE_DBNAME`

2. **Docker Compose中的错误配置**：
   ```yaml
   environment:
     - DATABASE_NAME=trpg_solo_engine  # ❌ 错误
   ```

3. **正确的配置**：
   ```yaml
   environment:
     - DATABASE_DBNAME=trpg_solo_engine  # ✅ 正确
   ```

**解决方案**：

修改 `docker-compose.yml` 中的环境变量名：

```yaml
services:
  backend:
    environment:
      - DATABASE_DBNAME=trpg_solo_engine  # 改为 DBNAME 而不是 NAME
```

然后重启服务：
```bash
docker-compose restart backend
```

**验证修复**：
```bash
# 检查日志
docker-compose logs backend | grep "database connection"

# 应该看到：
# database connection established","host":"postgres","port":5432,"database":"trpg_solo_engine"

# 测试健康检查
curl http://localhost:8080/health
```

### 环境变量命名规范

为了避免类似问题，请遵循以下规范：

| 配置键 | 环境变量名 | 说明 |
|--------|-----------|------|
| `server.port` | `SERVER_PORT` | 服务器端口 |
| `database.host` | `DATABASE_HOST` | 数据库主机 |
| `database.port` | `DATABASE_PORT` | 数据库端口 |
| `database.user` | `DATABASE_USER` | 数据库用户 |
| `database.password` | `DATABASE_PASSWORD` | 数据库密码 |
| `database.dbname` | `DATABASE_DBNAME` | 数据库名称 ⚠️ |
| `database.sslmode` | `DATABASE_SSLMODE` | SSL模式 |
| `redis.host` | `REDIS_HOST` | Redis主机 |
| `redis.port` | `REDIS_PORT` | Redis端口 |

**注意**：配置键中的点号（`.`）会被替换为下划线（`_`）。

## API文档访问问题

### 问题：访问 /api/docs 返回 404

**症状**：
```bash
curl http://localhost:8080/api/docs/
# 404 page not found
```

**原因**：
Docker容器中没有包含 `api/` 目录。

**解决方案1：使用本地运行（推荐）**

```bash
# 停止Docker中的后端
docker-compose stop backend

# 本地运行（使用Docker中的数据库）
go run cmd/server/main.go
```

**解决方案2：更新Docker配置**

1. 确保 `Dockerfile` 包含api目录：
   ```dockerfi