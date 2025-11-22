package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB 创建PostgreSQL数据库连接
func NewPostgresDB(log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.dbname"),
		viper.GetString("database.sslmode"),
	)

	// 配置GORM日志
	gormLogger := logger.Default
	if viper.GetString("log.level") == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 配置连接池
	maxOpenConns := viper.GetInt("database.max_open_conns")
	if maxOpenConns == 0 {
		maxOpenConns = 25
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := viper.GetInt("database.max_idle_conns")
	if maxIdleConns == 0 {
		maxIdleConns = 5
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	connMaxLifetime := viper.GetInt("database.conn_max_lifetime")
	if connMaxLifetime == 0 {
		connMaxLifetime = 300
	}
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	log.Info("database connection established",
		zap.String("host", viper.GetString("database.host")),
		zap.Int("port", viper.GetInt("database.port")),
		zap.String("database", viper.GetString("database.dbname")),
	)

	return db, nil
}
