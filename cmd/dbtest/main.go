package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
)

func main() {
	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "trpg")
	viper.SetDefault("database.password", "trpg_password")
	viper.SetDefault("database.dbname", "trpg_solo_engine")
	viper.SetDefault("database.sslmode", "disable")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Fatal("failed to read config", zap.Error(err))
		}
		logger.Info("using default configuration")
	}

	// 测试PostgreSQL连接
	logger.Info("testing PostgreSQL connection...")
	db, err := database.NewPostgresDB(logger)
	if err != nil {
		logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("✓ PostgreSQL connection successful")

	// 运行迁移
	logger.Info("running database migrations...")
	if err := database.RunMigrations(db, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("✓ Database migrations successful")

	// 测试Redis连接
	logger.Info("testing Redis connection...")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	redisClient, err := database.NewRedisClient(logger)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("✓ Redis connection successful")

	logger.Info("✓ All database connections successful!")
}
