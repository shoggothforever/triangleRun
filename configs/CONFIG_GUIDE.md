# 三角机构TRPG单人引擎 - 配置指南

本文档详细说明应用配置系统的使用方法。

## 配置文件概览

### 应用配置
- `config.yaml`: 主应用配置文件（服务器、数据库、Redis、AI、日志等）
- `.env`: 环境变量配置（不提交到版本控制）
- `.env.example`: 环境变量配置示例

### ARC系统配置
- `anomalies.json`: 9种异常体的配置
- `realities.json`: 9种现实的配置
- `careers.json`: 9种职能的配置

详见 [README.md](./README.md) 了解ARC配置。

## 配置优先级

配置加载遵循以下优先级（从高到低）：

1. **环境变量** - 最高优先级，覆盖所有其他配置
2. **config.yaml** - 主配置文件
3. **默认值** - 代码中定义的默认值

### 示例

如果同时存在以下配置：
- 环境变量: `DATABASE_HOST=prod-db.example.com`
- config.yaml: `database.host: localhost`

则实际使用的值为: `prod-db.example.com`

## config.yaml 详解

### 1. 服务器配置 (server)

```yaml
server:
  port: "8080"              # 服务器端口
  mode: "debug"             # 运行模式: debug, release, test
  read_timeout: 30          # 读取超时（秒）
  write_timeout: 30         # 写入超时（秒）
  idle_timeout: 120         # 空闲超时（秒）
  max_header_bytes: 1048576 # 最大请求头大小（字节）
  shutdown_timeout: 5       # 优雅关闭超时（秒）
```

**说明：**
- `mode`: 
  - `debug`: 开发模式，详细日志，Gin详细输出
  - `release`: 生产模式，优化性能，简化日志
  - `test`: 测试模式
- 超时配置根据实际网络情况调整

### 2. 日志配置 (log)

```yaml
log:
  level: "info"             # 日志级别: debug, info, warn, error, fatal
  format: "json"            # 日志格式: json, console
  output: "stdout"          # 输出目标: stdout, stderr, file
  file_path: "logs/app.log" # 日志文件路径（当output为file时）
  max_size: 100             # 日志文件最大大小（MB）
  max_backups: 3            # 保留的旧日志文件数量
  max_age: 28               # 保留日志文件的最大天数
  compress: true            # 是否压缩旧日志文件
  enable_caller: true       # 是否记录调用者信息
  enable_stacktrace: false  # 是否记录堆栈跟踪
```

**日志级别说明：**
- `debug`: 详细调试信息（开发环境）
- `info`: 一般信息（推荐生产环境）
- `warn`: 警告信息
- `error`: 错误信息
- `fatal`: 致命错误

**推荐配置：**
- 开发环境: `level: debug`, `format: console`, `output: stdout`
- 生产环境: `level: info`, `format: json`, `output: file`

### 3. 数据库配置 (database)

```yaml
database:
  host: "localhost"         # 数据库主机
  port: 5432                # 数据库端口
  user: "trpg"              # 数据库用户
  password: "***"           # 数据库密码（建议使用环境变量）
  dbname: "trpg_solo_engine" # 数据库名称
  sslmode: "disable"        # SSL模式
  timezone: "Asia/Shanghai" # 时区
  max_open_conns: 25        # 最大打开连接数
  max_idle_conns: 5         # 最大空闲连接数
  conn_max_lifetime: 300    # 连接最大生命周期（秒）
  conn_max_idle_time: 60    # 连接最大空闲时间（秒）
  log_level: "warn"         # 数据库日志级别
  slow_threshold: 200       # 慢查询阈值（毫秒）
  auto_migrate: true        # 是否自动运行迁移
```

**SSL模式说明：**
- `disable`: 不使用SSL
- `require`: 要求SSL但不验证证书
- `verify-ca`: 验证CA证书
- `verify-full`: 完全验证

**连接池调优：**
- 低负载: `max_open_conns: 10`, `max_idle_conns: 2`
- 中等负载: `max_open_conns: 25`, `max_idle_conns: 5`
- 高负载: `max_open_conns: 100`, `max_idle_conns: 10`

### 4. Redis配置 (redis)

```yaml
redis:
  host: "localhost"         # Redis主机
  port: 6379                # Redis端口
  password: ""              # Redis密码
  db: 0                     # Redis数据库编号
  pool_size: 10             # 连接池大小
  min_idle_conns: 2         # 最小空闲连接数
  max_retries: 3            # 最大重试次数
  dial_timeout: 5           # 连接超时（秒）
  read_timeout: 3           # 读取超时（秒）
  write_timeout: 3          # 写入超时（秒）
  pool_timeout: 4           # 连接池超时（秒）
  idle_timeout: 300         # 空闲连接超时（秒）
  cache_ttl:                # 缓存TTL配置（秒）
    session: 86400          # 会话缓存24小时
    agent: 3600             # 角色缓存1小时
    scenario: 3600          # 剧本缓存1小时
    scene: 3600             # 场景缓存1小时
```

**缓存策略：**
- `session`: 游戏会话数据，较长TTL
- `agent`: 角色数据，中等TTL
- `scenario/scene`: 静态数据，可以较长TTL

### 5. AI服务配置 (ai)

```yaml
ai:
  provider: "openai"        # AI提供商: openai, local, azure
  api_key: ""               # API密钥（建议使用环境变量）
  model: "gpt-4"            # 使用的模型
  base_url: ""              # 自定义API端点（可选）
  timeout: 30               # 请求超时（秒）
  max_tokens: 2000          # 最大生成token数
  temperature: 0.7          # 温度参数（0-2）
  top_p: 1.0                # Top-p采样参数
  frequency_penalty: 0.0    # 频率惩罚（-2.0到2.0）
  presence_penalty: 0.0     # 存在惩罚（-2.0到2.0）
  retry_attempts: 3         # 重试次数
  retry_delay: 1            # 重试延迟（秒）
  enable_cache: true        # 是否启用响应缓存
  cache_ttl: 3600           # 响应缓存时间（秒）
```

**模型选择：**
- `gpt-4`: 最高质量，较慢，较贵
- `gpt-3.5-turbo`: 平衡性能和成本
- `gpt-4-turbo`: 更快的GPT-4

**参数调优：**
- `temperature`: 控制随机性，0=确定性，2=高随机性
- `max_tokens`: 根据需要的响应长度调整
- `enable_cache`: 开发环境可关闭，生产环境建议开启

### 6. 认证配置 (auth)

```yaml
auth:
  jwt_secret: "***"         # JWT密钥（必须使用环境变量）
  jwt_expiration: 86400     # JWT过期时间（秒，24小时）
  jwt_refresh_expiration: 604800  # 刷新token过期时间（秒，7天）
  enable_auth: false        # 是否启用认证
  token_header: "Authorization"   # Token请求头名称
  token_prefix: "Bearer"    # Token前缀
```

**安全建议：**
- 生产环境必须启用认证: `enable_auth: true`
- JWT密钥必须使用强随机字符串（至少32字符）
- 定期轮换JWT密钥
- 使用HTTPS传输token

### 7. 速率限制配置 (rate_limit)

```yaml
rate_limit:
  enabled: true             # 是否启用速率限制
  global:                   # 全局默认限制
    max_requests: 1000      # 每个时间窗口的最大请求数
    window: 60              # 时间窗口（秒）
    by_ip: true             # 是否基于IP限流
    by_user: false          # 是否基于用户限流
  endpoints:                # 端点特定限制
    "/api/dice/*":
      max_requests: 100
      window: 60
      by_user: true
    "/api/ai/*":
      max_requests: 10
      window: 60
      by_user: true
    "/api/saves/*":
      max_requests: 20
      window: 60
      by_user: true
```

**限流策略：**
- `by_ip`: 基于客户端IP地址限流
- `by_user`: 基于认证用户ID限流
- 可以同时启用两种策略

**端点配置建议：**
- 高频操作（骰子）: 100次/分钟
- AI生成: 10次/分钟（避免API配额耗尽）
- 数据修改: 20-30次/分钟
- 数据查询: 50-100次/分钟

### 8. 游戏配置 (game)

```yaml
game:
  scenarios_path: "scenarios"  # 剧本文件目录
  arc_configs_path: "configs"  # ARC配置文件目录
  rules:                       # 游戏规则配置
    dice_count: 6              # 骰子数量
    dice_sides: 4              # 骰子面数
    success_value: 3           # 成功值
    triple_ascension_count: 3  # 三重升华所需的3的数量
    initial_qa_points: 9       # 初始资质保证点数
    relationship_count: 3      # 人际关系数量
    relationship_total_connection: 12  # 人际关系总连结点数
    commendations_for_capture: 3  # 捕获异常体的嘉奖数
    reprimands_for_escape: 3   # 异常体逃脱的申诫数
    death_commendation_cost: 5 # 死亡扣除的嘉奖数
  session:
    max_active_sessions: 10    # 每个用户最大活跃会话数
    session_timeout: 86400     # 会话超时时间（秒）
    auto_save_interval: 300    # 自动保存间隔（秒）
```

**注意：** 游戏规则配置应与《三角机构》规则书保持一致，不建议修改。

### 9. 性能配置 (performance)

```yaml
performance:
  targets:                  # 响应时间目标（毫秒）
    api_response: 200       # API响应时间（p95）
    dice_roll: 10           # 骰子掷骰
    scene_load: 100         # 场景加载
    ai_generation: 2000     # AI生成
    save_operation: 500     # 存档保存
  concurrency:
    max_goroutines: 1000    # 最大goroutine数量
    worker_pool_size: 10    # 工作池大小
  cache:
    enable_memory_cache: true  # 是否启用内存缓存
    memory_cache_size: 100  # 内存缓存大小（MB）
```

### 10. CORS配置 (cors)

```yaml
cors:
  enabled: true             # 是否启用CORS
  allow_origins:            # 允许的源
    - "http://localhost:3000"
    - "http://localhost:5173"
  allow_methods:            # 允许的HTTP方法
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:            # 允许的请求头
    - "Origin"
    - "Content-Type"
    - "Authorization"
  expose_headers:           # 暴露的响应头
    - "Content-Length"
    - "X-RateLimit-Limit"
  allow_credentials: true   # 是否允许凭证
  max_age: 86400            # 预检请求缓存时间（秒）
```

**生产环境配置：**
```yaml
cors:
  allow_origins:
    - "https://yourdomain.com"
  allow_credentials: true
```

### 11. 监控配置 (monitoring)

```yaml
monitoring:
  enabled: true             # 是否启用监控
  metrics_path: "/metrics"  # Prometheus指标端点
  health_path: "/health"    # 健康检查端点
  prometheus:
    enabled: true
    namespace: "trpg"
    subsystem: "solo_engine"
  tracing:
    enabled: false
    jaeger_endpoint: ""
    sample_rate: 0.1
```

### 12. 开发配置 (development)

```yaml
development:
  enable_pprof: false       # 是否启用pprof性能分析
  pprof_port: "6060"        # pprof端口
  enable_swagger: true      # 是否启用Swagger文档
  swagger_path: "/swagger"  # Swagger文档路径
  mock_ai: false            # 是否使用模拟AI
  seed_data: false          # 是否加载种子数据
```

**注意：** 生产环境应禁用所有开发功能。

## 环境变量

环境变量可以覆盖config.yaml中的配置。

### 命名规则

- 使用大写字母和下划线
- 嵌套配置使用下划线分隔
- 例如: `database.host` → `DATABASE_HOST`

### 关键环境变量

**必须设置（生产环境）：**

```bash
# 数据库密码
DATABASE_PASSWORD=your_secure_password

# Redis密码（如果启用）
REDIS_PASSWORD=your_redis_password

# AI API密钥
AI_API_KEY=your_openai_api_key

# JWT密钥
JWT_SECRET=your_jwt_secret_key_at_least_32_chars
```

**可选环境变量：**

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=release

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# 功能开关
ENABLE_AUTH=true
RATE_LIMIT_ENABLED=true
MONITORING_ENABLED=true
```

## 配置最佳实践

### 1. 安全性

✅ **DO:**
- 使用环境变量存储敏感信息
- 在生产环境使用强密码和密钥
- 定期轮换密钥和密码
- 启用SSL/TLS连接

❌ **DON'T:**
- 将密码、密钥提交到版本控制
- 在日志中输出敏感信息
- 使用默认密码
- 在生产环境禁用认证

### 2. 环境分离

为不同环境创建不同的配置：

```
configs/
├── config.yaml          # 开发环境
├── config.test.yaml     # 测试环境
└── config.prod.yaml     # 生产环境
```

使用环境变量指定配置文件：
```bash
CONFIG_FILE=config.prod.yaml ./trpg-engine
```

### 3. 性能调优

根据负载调整以下参数：

**低负载（< 100 req/s）：**
```yaml
database:
  max_open_conns: 10
  max_idle_conns: 2
redis:
  pool_size: 5
```

**中等负载（100-1000 req/s）：**
```yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
redis:
  pool_size: 10
```

**高负载（> 1000 req/s）：**
```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
redis:
  pool_size: 50
```

### 4. 监控和日志

生产环境建议配置：

```yaml
log:
  level: "info"              # 不要使用debug
  format: "json"             # 便于日志聚合
  output: "file"             # 输出到文件
  enable_caller: true        # 记录调用者
  enable_stacktrace: false   # 仅error级别记录堆栈

monitoring:
  enabled: true              # 启用监控
  prometheus:
    enabled: true            # 启用Prometheus指标
```

## 故障排查

### 配置文件未找到

```
Error: Config File "config" Not Found in "[.]"
```

**解决方案：**
1. 确保config.yaml在正确的位置
2. 使用`--config`参数指定路径
3. 检查文件权限

### 数据库连接失败

```
Error: failed to connect to database
```

**检查清单：**
- [ ] 数据库服务是否运行
- [ ] 主机和端口是否正确
- [ ] 用户名和密码是否正确
- [ ] 数据库是否存在
- [ ] 防火墙是否允许连接
- [ ] SSL模式是否正确

**测试连接：**
```bash
psql -h localhost -p 5432 -U trpg -d trpg_solo_engine
```

### Redis连接失败

```
Error: failed to connect to redis
```

**检查清单：**
- [ ] Redis服务是否运行
- [ ] 主机和端口是否正确
- [ ] 密码是否正确
- [ ] 网络是否可达

**测试连接：**
```bash
redis-cli -h localhost -p 6379 ping
```

### AI服务调用失败

```
Error: AI service error
```

**检查清单：**
- [ ] API密钥是否正确
- [ ] 网络是否可以访问AI服务
- [ ] API配额是否充足
- [ ] 模型名称是否正确
- [ ] 请求超时是否合理

**测试API：**
```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $AI_API_KEY"
```

## 配置验证

启动前验证配置：

```bash
# 检查YAML语法
yamllint configs/config.yaml

# 验证配置加载
./trpg-engine --validate-config

# 测试数据库连接
./trpg-engine --test-db

# 测试Redis连接
./trpg-engine --test-redis
```

## 参考资料

- [Viper配置库文档](https://github.com/spf13/viper)
- [YAML语法参考](https://yaml.org/)
- [12-Factor App配置原则](https://12factor.net/config)
- [PostgreSQL连接参数](https://www.postgresql.org/docs/current/libpq-connect.html)
- [Redis配置](https://redis.io/docs/management/config/)
