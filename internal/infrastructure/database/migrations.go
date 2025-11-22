package database

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AgentModel 角色数据库模型
type AgentModel struct {
	ID            string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string `gorm:"type:varchar(255);not null"`
	Pronouns      string `gorm:"type:varchar(50)"`
	AnomalyType   string `gorm:"type:varchar(50);not null"`
	RealityType   string `gorm:"type:varchar(50);not null"`
	CareerType    string `gorm:"type:varchar(50);not null"`
	QA            string `gorm:"type:jsonb;not null"`
	Relationships string `gorm:"type:jsonb;not null"`
	Commendations int    `gorm:"default:0"`
	Reprimands    int    `gorm:"default:0"`
	Rating        string `gorm:"type:varchar(50);default:'评级良好'"`
	Alive         bool   `gorm:"default:true"`
	InDebt        bool   `gorm:"default:false"`
	CreatedAt     int64  `gorm:"autoCreateTime"`
	UpdatedAt     int64  `gorm:"autoUpdateTime"`
}

func (AgentModel) TableName() string {
	return "agents"
}

// GameSessionModel 游戏会话数据库模型
type GameSessionModel struct {
	ID         string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID    string `gorm:"type:uuid;not null;index"`
	ScenarioID string `gorm:"type:varchar(100);not null"`
	Phase      string `gorm:"type:varchar(50);not null;index"`
	State      string `gorm:"type:jsonb;not null"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"`
}

func (GameSessionModel) TableName() string {
	return "game_sessions"
}

// SaveModel 存档数据库模型
type SaveModel struct {
	ID        string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID string `gorm:"type:uuid;not null;index"`
	Name      string `gorm:"type:varchar(255);not null"`
	Snapshot  string `gorm:"type:jsonb;not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}

func (SaveModel) TableName() string {
	return "saves"
}

// RunMigrations 执行数据库迁移
func RunMigrations(db *gorm.DB, log *zap.Logger) error {
	log.Info("running database migrations...")

	// 自动迁移
	if err := db.AutoMigrate(
		&AgentModel{},
		&GameSessionModel{},
		&SaveModel{},
	); err != nil {
		return err
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		return err
	}

	log.Info("database migrations completed")
	return nil
}

func createIndexes(db *gorm.DB) error {
	// 为agents表创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name)").Error; err != nil {
		return err
	}

	// 为game_sessions表创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_agent ON game_sessions(agent_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_phase ON game_sessions(phase)").Error; err != nil {
		return err
	}

	// 为saves表创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_saves_session ON saves(session_id)").Error; err != nil {
		return err
	}

	return nil
}
