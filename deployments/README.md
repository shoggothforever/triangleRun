# 三角机构TRPG单人引擎 - 部署指南

本目录包含三角机构TRPG单人引擎的完整部署配置和脚本。

## 目录结构

```
deployments/
├── README.md                      # 本文件
├── docker-compose.prod.yml        # 生产环境Docker Compose配置
├── kubernetes/                    # Kubernetes部署配置
│   ├── namespace.yaml            # 命名空间
│   ├── configmap.yaml            # 配置映射
│   ├── secrets.yaml              # 密钥（需要更新）
│   ├── postgres-deployment.yaml  # PostgreSQL部署
│   ├── redis-deployment.yaml     # Redis部署
│   ├── backend-deployment.yaml   # 后端服务部署
│   ├── ingress.yaml              # Ingress配置
│   ├── hpa.yaml                  # 水平自动扩缩容
│   └── kustomization.yaml        # Kustomize配置
├── nginx/                         # Nginx配置
│   └── nginx.conf                # Nginx配置文件
├── prometheus/                    # Prometheus监控配置
│   └── prometheus.yml            # Prometheus配置
├── grafana/                       # Grafana配置
│   ├── datasources/              # 数据源配置
│   └── dashboards/               # 仪表板配置
└── scripts/                       # 部署脚本
    ├── deploy-docker.sh          # Docker部署脚本
    └── deploy-k8s.sh             # Kubernetes部署脚本
```

## 部署方式

### 1. Docker Compose部署（推荐用于开发和小规模生产）

#### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少2GB可用内存
- 至少10GB可用磁盘空间

#### 开发环境部署

```bash
# 1. 复制环境变量模板
cp .env.example .env

# 2. 编辑.env文件，填入实际配置
vim .env

# 3. 启动服务
./deployments/scripts/deploy-docker.sh dev

# 或者手动启动
docker-compose up -d
```

#### 生产环境部署

```bash
# 1. 确保.env文件已正确配置
vim .env

# 2. 使用生产配置部署
./deployments/scripts/deploy-docker.sh prod

# 或者手动启动
docker-compose -f docker-compose.yml -f deployments/docker-compose.prod.yml up -d
```

#### 启用监控

```bash
# 启动Prometheus和Grafana
docker-compose --profile monitoring up -d

# 访问监控界面
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3001 (默认用户名/密码: admin/admin)
```

#### 常用命令

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f [service_name]

# 重启服务
docker-compose restart [service_name]

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v

# 进入容器
docker-compose exec backend sh
```

### 2. Kubernetes部署（推荐用于大规模生产）

#### 前置要求

- Kubernetes 1.24+
- kubectl配置正确
- 至少3个工作节点（生产环境）
- Ingress Controller（如nginx-ingress）
- 存储类（StorageClass）
- 可选：cert-manager（用于SSL证书）

#### 部署步骤

```bash
# 1. 更新secrets.yaml中的敏感信息
vim deployments/kubernetes/secrets.yaml

# 2. 更新configmap.yaml中的配置
vim deployments/kubernetes/configmap.yaml

# 3. 更新ingress.yaml中的域名
vim deployments/kubernetes/ingress.yaml

# 4. 执行部署
./deployments/scripts/deploy-k8s.sh apply

# 或者使用kubectl手动部署
cd deployments/kubernetes
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml
kubectl apply -f backend-deployment.yaml
kubectl apply -f ingress.yaml
kubectl apply -f hpa.yaml
```

#### 使用Kustomize部署

```bash
# 1. 更新kustomization.yaml中的镜像信息
vim deployments/kubernetes/kustomization.yaml

# 2. 使用kustomize部署
kubectl apply -k deployments/kubernetes/
```

#### 常用命令

```bash
# 查看部署状态
./deployments/scripts/deploy-k8s.sh status

# 查看Pod日志
kubectl logs -f deployment/trpg-backend -n trpg-solo-engine

# 查看所有资源
kubectl get all -n trpg-solo-engine

# 端口转发（本地测试）
kubectl port-forward svc/trpg-backend-service 8080:8080 -n trpg-solo-engine

# 进入Pod
kubectl exec -it <pod-name> -n trpg-solo-engine -- /bin/sh

# 删除部署
./deployments/scripts/deploy-k8s.sh delete
```

### 3. 手动部署（不推荐）

如果需要手动部署，请参考以下步骤：

1. 安装Go 1.23+
2. 安装PostgreSQL 15+
3. 安装Redis 7+
4. 配置环境变量
5. 构建应用：`go build -o trpg-engine ./cmd/server`
6. 运行应用：`./trpg-engine`

## 配置说明

### 环境变量

所有环境变量都在`.env.example`中有详细说明。关键配置包括：

#### 数据库配置

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=trpg
DATABASE_PASSWORD=your_secure_password
DATABASE_NAME=trpg_solo_engine
```

#### Redis配置

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
```

#### AI服务配置

```env
AI_PROVIDER=openai
AI_API_KEY=your_openai_api_key
AI_MODEL=gpt-4
```

#### 认证配置

```env
JWT_SECRET=your_very_secure_random_key
ENABLE_AUTH=true
```

### 资源要求

#### 最小配置（开发环境）

- CPU: 2核
- 内存: 2GB
- 磁盘: 10GB

#### 推荐配置（生产环境）

- CPU: 4核
- 内存: 8GB
- 磁盘: 50GB SSD

#### 大规模生产环境

- 后端服务: 3-10个副本，每个512MB-1GB内存
- PostgreSQL: 2核，2GB内存，50GB存储
- Redis: 1核，512MB内存，5GB存储

## 监控和日志

### Prometheus指标

后端服务暴露以下指标端点：

- `/metrics` - Prometheus格式的指标

关键指标：

- `http_requests_total` - HTTP请求总数
- `http_request_duration_seconds` - 请求延迟
- `dice_rolls_total` - 骰子掷骰次数
- `game_sessions_active` - 活跃游戏会话数
- `database_connections` - 数据库连接数

### 日志

日志输出到stdout，可以通过以下方式查看：

```bash
# Docker
docker-compose logs -f backend

# Kubernetes
kubectl logs -f deployment/trpg-backend -n trpg-solo-engine
```

日志级别：

- `debug` - 详细调试信息
- `info` - 一般信息（默认）
- `warn` - 警告信息
- `error` - 错误信息
- `fatal` - 致命错误

## 备份和恢复

### PostgreSQL备份

```bash
# Docker环境
docker-compose exec postgres pg_dump -U trpg trpg_solo_engine > backup.sql

# Kubernetes环境
kubectl exec -n trpg-solo-engine postgres-0 -- pg_dump -U trpg trpg_solo_engine > backup.sql
```

### PostgreSQL恢复

```bash
# Docker环境
docker-compose exec -T postgres psql -U trpg trpg_solo_engine < backup.sql

# Kubernetes环境
kubectl exec -i -n trpg-solo-engine postgres-0 -- psql -U trpg trpg_solo_engine < backup.sql
```

### Redis备份

Redis使用AOF持久化，数据自动保存在数据卷中。

## 安全建议

### 生产环境必做

1. **更新所有默认密码**
   - 数据库密码
   - Redis密码
   - JWT密钥

2. **启用HTTPS**
   - 配置SSL证书
   - 强制HTTPS重定向

3. **启用认证**
   - 设置`ENABLE_AUTH=true`
   - 配置强JWT密钥

4. **配置防火墙**
   - 只暴露必要端口
   - 限制数据库访问

5. **定期备份**
   - 设置自动备份
   - 测试恢复流程

6. **监控和告警**
   - 配置Prometheus告警
   - 设置日志监控

7. **更新依赖**
   - 定期更新Docker镜像
   - 更新Go依赖

### 网络安全

- 使用私有网络
- 配置网络策略（Kubernetes）
- 启用速率限制
- 配置CORS正确的源

## 故障排查

### 服务无法启动

1. 检查日志：`docker-compose logs backend`
2. 检查环境变量配置
3. 确认数据库和Redis可访问
4. 检查端口是否被占用

### 数据库连接失败

1. 检查数据库是否运行：`docker-compose ps postgres`
2. 检查数据库凭证
3. 检查网络连接
4. 查看数据库日志：`docker-compose logs postgres`

### 性能问题

1. 检查资源使用：`docker stats`
2. 查看慢查询日志
3. 检查Redis缓存命中率
4. 分析Prometheus指标

### 内存不足

1. 增加容器内存限制
2. 优化数据库查询
3. 调整Redis最大内存
4. 启用HPA自动扩缩容

## 升级指南

### Docker Compose升级

```bash
# 1. 拉取最新代码
git pull

# 2. 备份数据库
docker-compose exec postgres pg_dump -U trpg trpg_solo_engine > backup.sql

# 3. 停止服务
docker-compose down

# 4. 重新构建和启动
docker-compose up -d --build

# 5. 检查服务状态
docker-compose ps
```

### Kubernetes滚动升级

```bash
# 1. 更新镜像
kubectl set image deployment/trpg-backend trpg-backend=new-image:tag -n trpg-solo-engine

# 2. 监控升级过程
kubectl rollout status deployment/trpg-backend -n trpg-solo-engine

# 3. 如果出现问题，回滚
kubectl rollout undo deployment/trpg-backend -n trpg-solo-engine
```

## 支持和联系

如有问题，请：

1. 查看项目README
2. 查看故障排查文档
3. 提交GitHub Issue
4. 联系维护团队

## 许可证

本项目遵循项目根目录的LICENSE文件。
