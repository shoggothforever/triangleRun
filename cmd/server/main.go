package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/handler"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
)

func main() {
	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// 加载配置
	if err := loadConfig(); err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// 初始化数据库
	db, err := database.NewPostgresDB(logger)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	// 运行数据库迁移
	if err := database.RunMigrations(db, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// 初始化Redis
	redisClient, err := database.NewRedisClient(logger)
	if err != nil {
		logger.Fatal("failed to initialize redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 初始化服务
	diceService := domain.NewDiceService()

	// 创建Gin路由
	router := setupRouter(logger, db, redisClient, diceService)

	// 创建HTTP服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info("starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认值
			return nil
		}
		return err
	}

	return nil
}

func setupRouter(logger *zap.Logger, db *gorm.DB, redisClient *redis.Client, diceService domain.DiceService) *gin.Engine {
	// 设置Gin模式
	if viper.GetString("server.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(logger))

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		sqlDB, err := db.DB()
		dbStatus := "ok"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
		}

		// 检查Redis连接
		redisStatus := "ok"
		if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus = "error"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "trpg-solo-engine",
			"database": dbStatus,
			"redis":    redisStatus,
		})
	})

	// 初始化处理器
	diceHandler := handler.NewDiceHandler(diceService)

	// API路由组
	api := router.Group("/api")
	{
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": "0.1.0",
				"name":    "TRPG Solo Engine",
			})
		})

		// 骰子API
		dice := api.Group("/dice")
		{
			dice.POST("/roll", diceHandler.RollDice)
		}
	}

	return router
}

func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		)
	}
}
