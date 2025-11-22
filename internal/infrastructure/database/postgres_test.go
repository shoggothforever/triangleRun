package database

import (
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestDatabaseModels(t *testing.T) {
	// 测试模型结构
	agent := &AgentModel{
		Name:          "测试特工",
		AnomalyType:   "低语",
		RealityType:   "看护者",
		CareerType:    "公关",
		QA:            `{"专注": 1}`,
		Relationships: `[]`,
	}

	if agent.Name != "测试特工" {
		t.Errorf("expected name to be 测试特工, got %s", agent.Name)
	}

	session := &GameSessionModel{
		AgentID:    "test-agent-id",
		ScenarioID: "test-scenario",
		Phase:      "morning",
		State:      `{}`,
	}

	if session.Phase != "morning" {
		t.Errorf("expected phase to be morning, got %s", session.Phase)
	}

	save := &SaveModel{
		SessionID: "test-session-id",
		Name:      "测试存档",
		Snapshot:  `{}`,
	}

	if save.Name != "测试存档" {
		t.Errorf("expected save name to be 测试存档, got %s", save.Name)
	}
}

func TestConfigDefaults(t *testing.T) {
	// 测试配置默认值
	viper.Reset()
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)

	if viper.GetInt("database.max_open_conns") != 25 {
		t.Errorf("expected max_open_conns to be 25")
	}

	if viper.GetInt("database.max_idle_conns") != 5 {
		t.Errorf("expected max_idle_conns to be 5")
	}
}

// 注意：实际的数据库连接测试需要运行的PostgreSQL实例
// 这些测试在CI/CD环境中应该使用测试数据库
func TestNewPostgresDB_WithoutDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// 设置无效的数据库配置
	viper.Reset()
	viper.Set("database.host", "invalid-host")
	viper.Set("database.port", 5432)
	viper.Set("database.user", "test")
	viper.Set("database.password", "test")
	viper.Set("database.dbname", "test")
	viper.Set("database.sslmode", "disable")

	_, err := NewPostgresDB(logger)
	if err == nil {
		t.Error("expected error when connecting to invalid database")
	}
}
