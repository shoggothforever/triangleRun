# ============================================
# 构建阶段
# ============================================
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用（带版本信息和优化）
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o trpg-engine \
    ./cmd/server

# 验证构建
RUN ./trpg-engine --version || true

# ============================================
# 运行阶段
# ============================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# 创建非root用户
RUN addgroup -g 1000 trpg && \
    adduser -D -u 1000 -G trpg trpg

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/trpg-engine .

# 复制配置和数据文件
COPY --from=builder --chown=trpg:trpg /app/configs ./configs
COPY --from=builder --chown=trpg:trpg /app/scenarios ./scenarios
COPY --from=builder --chown=trpg:trpg /app/api ./api

# 创建日志目录
RUN mkdir -p /app/logs && chown -R trpg:trpg /app/logs

# 切换到非root用户
USER trpg

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 设置环境变量
ENV SERVER_PORT=8080 \
    SERVER_MODE=release \
    LOG_LEVEL=info

# 运行应用
ENTRYPOINT ["./trpg-engine"]
CMD []
