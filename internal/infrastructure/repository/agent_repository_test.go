package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAgentModel SQLite兼容的测试模型
type TestAgentModel struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	Pronouns      string
	AnomalyType   string
	RealityType   string
	CareerType    string
	QA            string
	Relationships string
	Commendations int
	Reprimands    int
	Rating        string
	Alive         bool
	InDebt        bool
	CreatedAt     int64
	UpdatedAt     int64
}

func (TestAgentModel) TableName() string {
	return "agents"
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 使用SQLite兼容的模型进行迁移
	err = db.AutoMigrate(&TestAgentModel{})
	require.NoError(t, err)

	return db
}

// setupTestRedis 创建测试Redis客户端（使用miniredis或mock）
func setupTestRedis(t *testing.T) *redis.Client {
	// 使用nil表示没有Redis（测试时可选）
	return nil
}

// createTestAgent 创建测试角色
func createTestAgent() *domain.Agent {
	return &domain.Agent{
		ID:       uuid.New().String(),
		Name:     "测试特工",
		Pronouns: "他/她",
		Anomaly: &domain.Anomaly{
			Type:      domain.AnomalyWhisper,
			Abilities: createDefaultAbilities(domain.AnomalyWhisper),
		},
		Reality: &domain.Reality{
			Type: domain.RealityCaretaker,
			Trigger: &domain.RealityTrigger{
				Name:        "现实触发",
				Cost:        0,
				Effect:      "触发效果",
				Consequence: "忽视后果",
			},
			OverloadRelief: &domain.OverloadRelief{
				Name:      "过载解除",
				Condition: "满足条件",
				Effect:    "无视所有过载",
			},
			DegradationTrack: &domain.DegradationTrack{
				Name:   "退化轨道",
				Filled: 0,
				Total:  4,
			},
		},
		Career: &domain.Career{
			Type: domain.CareerPublicRelations,
			QA: map[string]int{
				domain.QualityFocus:      1,
				domain.QualityEmpathy:    1,
				domain.QualityPresence:   1,
				domain.QualityDeception:  1,
				domain.QualityInitiative: 1,
				domain.QualityProfession: 1,
				domain.QualityVitality:   1,
				domain.QualityGrit:       1,
				domain.QualitySubtlety:   1,
			},
		},
		QA: map[string]int{
			domain.QualityFocus:      1,
			domain.QualityEmpathy:    1,
			domain.QualityPresence:   1,
			domain.QualityDeception:  1,
			domain.QualityInitiative: 1,
			domain.QualityProfession: 1,
			domain.QualityVitality:   1,
			domain.QualityGrit:       1,
			domain.QualitySubtlety:   1,
		},
		Relationships: []*domain.Relationship{
			{ID: uuid.New().String(), Name: "关系1", Connection: 6},
			{ID: uuid.New().String(), Name: "关系2", Connection: 3},
			{ID: uuid.New().String(), Name: "关系3", Connection: 3},
		},
		Commendations: 0,
		Reprimands:    0,
		Rating:        domain.RatingExcellent,
		Alive:         true,
		InDebt:        false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// TestAgentRepository_Create 测试创建角色
func TestAgentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	agent := createTestAgent()

	// 测试创建
	err := repo.Create(ctx, agent)
	require.NoError(t, err)
	assert.NotEmpty(t, agent.ID)

	// 验证数据库中存在
	var model database.AgentModel
	err = db.Where("id = ?", agent.ID).First(&model).Error
	require.NoError(t, err)
	assert.Equal(t, agent.Name, model.Name)
	assert.Equal(t, agent.Anomaly.Type, model.AnomalyType)
}

// TestAgentRepository_GetByID 测试获取角色
func TestAgentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	agent := createTestAgent()

	// 先创建
	err := repo.Create(ctx, agent)
	require.NoError(t, err)

	// 测试获取
	retrieved, err := repo.GetByID(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
	assert.Equal(t, agent.Name, retrieved.Name)
	assert.Equal(t, agent.Anomaly.Type, retrieved.Anomaly.Type)
	assert.Equal(t, agent.Reality.Type, retrieved.Reality.Type)
	assert.Equal(t, agent.Career.Type, retrieved.Career.Type)
}

// TestAgentRepository_GetByID_NotFound 测试获取不存在的角色
func TestAgentRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	nonExistentID := uuid.New().String()

	// 测试获取不存在的角色
	_, err := repo.GetByID(ctx, nonExistentID)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestAgentRepository_Update 测试更新角色
func TestAgentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	agent := createTestAgent()

	// 先创建
	err := repo.Create(ctx, agent)
	require.NoError(t, err)

	// 修改角色
	agent.Name = "更新后的特工"
	agent.Commendations = 5
	agent.Reprimands = 2

	// 测试更新
	err = repo.Update(ctx, agent)
	require.NoError(t, err)

	// 验证更新
	retrieved, err := repo.GetByID(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新后的特工", retrieved.Name)
	assert.Equal(t, 5, retrieved.Commendations)
	assert.Equal(t, 2, retrieved.Reprimands)
}

// TestAgentRepository_Update_NotFound 测试更新不存在的角色
func TestAgentRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	agent := createTestAgent()

	// 测试更新不存在的角色
	err := repo.Update(ctx, agent)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestAgentRepository_Delete 测试删除角色
func TestAgentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	agent := createTestAgent()

	// 先创建
	err := repo.Create(ctx, agent)
	require.NoError(t, err)

	// 测试删除
	err = repo.Delete(ctx, agent.ID)
	require.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, agent.ID)
	require.Error(t, err)
}

// TestAgentRepository_Delete_NotFound 测试删除不存在的角色
func TestAgentRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()
	nonExistentID := uuid.New().String()

	// 测试删除不存在的角色
	err := repo.Delete(ctx, nonExistentID)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestAgentRepository_List 测试列出所有角色
func TestAgentRepository_List(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()

	// 创建多个角色
	agent1 := createTestAgent()
	agent1.Name = "特工1"
	err := repo.Create(ctx, agent1)
	require.NoError(t, err)

	agent2 := createTestAgent()
	agent2.Name = "特工2"
	err = repo.Create(ctx, agent2)
	require.NoError(t, err)

	// 测试列出
	agents, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, agents, 2)

	// 验证角色名称
	names := []string{agents[0].Name, agents[1].Name}
	assert.Contains(t, names, "特工1")
	assert.Contains(t, names, "特工2")
}

// TestAgentRepository_WithTx 测试事务支持
func TestAgentRepository_WithTx(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()

	// 开始事务
	tx := db.Begin()
	txRepo := repo.WithTx(tx)

	agent := createTestAgent()

	// 在事务中创建
	err := txRepo.Create(ctx, agent)
	require.NoError(t, err)

	// 回滚事务
	tx.Rollback()

	// 验证角色不存在（因为回滚了）
	_, err = repo.GetByID(ctx, agent.ID)
	require.Error(t, err)
}

// TestAgentRepository_WithTx_Commit 测试事务提交
func TestAgentRepository_WithTx_Commit(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger)

	ctx := context.Background()

	// 开始事务
	tx := db.Begin()
	txRepo := repo.WithTx(tx)

	agent := createTestAgent()

	// 在事务中创建
	err := txRepo.Create(ctx, agent)
	require.NoError(t, err)

	// 提交事务
	tx.Commit()

	// 验证角色存在
	retrieved, err := repo.GetByID(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, retrieved.ID)
}

// TestAgentRepository_ToModel_ToDomain 测试模型转换
func TestAgentRepository_ToModel_ToDomain(t *testing.T) {
	db := setupTestDB(t)
	redis := setupTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewAgentRepository(db, redis, logger).(*agentRepository)

	agent := createTestAgent()

	// 测试转换为模型
	model, err := repo.toModel(agent)
	require.NoError(t, err)
	assert.Equal(t, agent.Name, model.Name)
	assert.Equal(t, agent.Anomaly.Type, model.AnomalyType)

	// 验证QA序列化
	var qa map[string]int
	err = json.Unmarshal([]byte(model.QA), &qa)
	require.NoError(t, err)
	assert.Equal(t, agent.QA, qa)

	// 验证Relationships序列化
	var rels []*domain.Relationship
	err = json.Unmarshal([]byte(model.Relationships), &rels)
	require.NoError(t, err)
	assert.Len(t, rels, 3)

	// 测试转换为领域模型
	converted, err := repo.toDomain(model)
	require.NoError(t, err)
	assert.Equal(t, agent.Name, converted.Name)
	assert.Equal(t, agent.Anomaly.Type, converted.Anomaly.Type)
	assert.Equal(t, agent.Reality.Type, converted.Reality.Type)
	assert.Equal(t, agent.Career.Type, converted.Career.Type)
	assert.Equal(t, agent.QA, converted.QA)
	assert.Len(t, converted.Relationships, 3)
}
