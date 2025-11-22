package repository

import (
	"context"
	"encoding/json"
	"sync"
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

// TestGameSessionModel SQLite兼容的测试模型
type TestGameSessionModel struct {
	ID         string `gorm:"primaryKey"`
	AgentID    string
	ScenarioID string
	Phase      string
	State      string
	CreatedAt  int64
	UpdatedAt  int64
}

func (TestGameSessionModel) TableName() string {
	return "game_sessions"
}

// setupSessionTestDB 创建测试数据库
func setupSessionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 使用SQLite兼容的模型进行迁移
	err = db.AutoMigrate(&TestGameSessionModel{})
	require.NoError(t, err)

	return db
}

// setupSessionTestRedis 创建测试Redis客户端
func setupSessionTestRedis(t *testing.T) *redis.Client {
	// 使用nil表示没有Redis（测试时可选）
	return nil
}

// createTestSession 创建测试会话
func createTestSession() *domain.GameSession {
	return &domain.GameSession{
		ID:         uuid.New().String(),
		AgentID:    uuid.New().String(),
		ScenarioID: "eternal-spring",
		Phase:      domain.PhaseMorning,
		State: &domain.GameState{
			CurrentSceneID:    "scene-1",
			VisitedScenes:     map[string]bool{"scene-1": true},
			CollectedClues:    []string{"clue-1", "clue-2"},
			UnlockedLocations: []string{"location-1"},
			DomainUnlocked:    false,
			NPCStates: map[string]*domain.NPCState{
				"npc-1": {
					ID:              "npc-1",
					CurrentState:    "neutral",
					AnomalyAffected: false,
					Relationship:    0,
					CustomData:      map[string]interface{}{"key": "value"},
				},
			},
			ChaosPool:         5,
			LooseEnds:         2,
			LocationOverloads: map[string]int{"location-1": 1},
			AnomalyStatus:     "active",
			MissionOutcome:    "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestSessionRepository_Create 测试创建会话
func TestSessionRepository_Create(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 测试创建
	err := repo.Create(ctx, session)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)

	// 验证数据库中存在
	var model database.GameSessionModel
	err = db.Where("id = ?", session.ID).First(&model).Error
	require.NoError(t, err)
	assert.Equal(t, session.AgentID, model.AgentID)
	assert.Equal(t, session.ScenarioID, model.ScenarioID)
	assert.Equal(t, string(session.Phase), model.Phase)
}

// TestSessionRepository_GetByID 测试获取会话
func TestSessionRepository_GetByID(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 先创建
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// 测试获取
	retrieved, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, session.AgentID, retrieved.AgentID)
	assert.Equal(t, session.ScenarioID, retrieved.ScenarioID)
	assert.Equal(t, session.Phase, retrieved.Phase)
	assert.Equal(t, session.State.CurrentSceneID, retrieved.State.CurrentSceneID)
	assert.Equal(t, session.State.ChaosPool, retrieved.State.ChaosPool)
	assert.Equal(t, session.State.LooseEnds, retrieved.State.LooseEnds)
}

// TestSessionRepository_GetByID_NotFound 测试获取不存在的会话
func TestSessionRepository_GetByID_NotFound(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	nonExistentID := uuid.New().String()

	// 测试获取不存在的会话
	_, err := repo.GetByID(ctx, nonExistentID)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestSessionRepository_Update 测试更新会话
func TestSessionRepository_Update(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 先创建
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// 修改会话
	session.Phase = domain.PhaseInvestigation
	session.State.CurrentSceneID = "scene-2"
	session.State.ChaosPool = 10
	session.State.LooseEnds = 5

	// 测试更新
	err = repo.Update(ctx, session)
	require.NoError(t, err)

	// 验证更新
	retrieved, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseInvestigation, retrieved.Phase)
	assert.Equal(t, "scene-2", retrieved.State.CurrentSceneID)
	assert.Equal(t, 10, retrieved.State.ChaosPool)
	assert.Equal(t, 5, retrieved.State.LooseEnds)
}

// TestSessionRepository_Update_NotFound 测试更新不存在的会话
func TestSessionRepository_Update_NotFound(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 测试更新不存在的会话
	err := repo.Update(ctx, session)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestSessionRepository_Delete 测试删除会话
func TestSessionRepository_Delete(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 先创建
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// 测试删除
	err = repo.Delete(ctx, session.ID)
	require.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, session.ID)
	require.Error(t, err)
}

// TestSessionRepository_Delete_NotFound 测试删除不存在的会话
func TestSessionRepository_Delete_NotFound(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	nonExistentID := uuid.New().String()

	// 测试删除不存在的会话
	err := repo.Delete(ctx, nonExistentID)
	require.Error(t, err)

	// 验证错误类型
	gameErr, ok := err.(*domain.GameError)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, gameErr.Code)
}

// TestSessionRepository_ListByAgent 测试列出指定角色的所有会话
func TestSessionRepository_ListByAgent(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	agentID := uuid.New().String()

	// 创建多个会话
	session1 := createTestSession()
	session1.AgentID = agentID
	session1.ScenarioID = "scenario-1"
	err := repo.Create(ctx, session1)
	require.NoError(t, err)

	session2 := createTestSession()
	session2.AgentID = agentID
	session2.ScenarioID = "scenario-2"
	err = repo.Create(ctx, session2)
	require.NoError(t, err)

	// 创建另一个角色的会话
	session3 := createTestSession()
	session3.AgentID = uuid.New().String()
	err = repo.Create(ctx, session3)
	require.NoError(t, err)

	// 测试列出
	sessions, err := repo.ListByAgent(ctx, agentID)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// 验证会话属于正确的角色
	for _, s := range sessions {
		assert.Equal(t, agentID, s.AgentID)
	}
}

// TestSessionRepository_WithTx 测试事务支持
func TestSessionRepository_WithTx(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()

	// 开始事务
	tx := db.Begin()
	txRepo := repo.WithTx(tx)

	session := createTestSession()

	// 在事务中创建
	err := txRepo.Create(ctx, session)
	require.NoError(t, err)

	// 回滚事务
	tx.Rollback()

	// 验证会话不存在（因为回滚了）
	_, err = repo.GetByID(ctx, session.ID)
	require.Error(t, err)
}

// TestSessionRepository_WithTx_Commit 测试事务提交
func TestSessionRepository_WithTx_Commit(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()

	// 开始事务
	tx := db.Begin()
	txRepo := repo.WithTx(tx)

	session := createTestSession()

	// 在事务中创建
	err := txRepo.Create(ctx, session)
	require.NoError(t, err)

	// 提交事务
	tx.Commit()

	// 验证会话存在
	retrieved, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
}

// TestSessionRepository_ToModel_ToDomain 测试模型转换
func TestSessionRepository_ToModel_ToDomain(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger).(*sessionRepository)

	session := createTestSession()

	// 测试转换为模型
	model, err := repo.toModel(session)
	require.NoError(t, err)
	assert.Equal(t, session.ID, model.ID)
	assert.Equal(t, session.AgentID, model.AgentID)
	assert.Equal(t, session.ScenarioID, model.ScenarioID)
	assert.Equal(t, string(session.Phase), model.Phase)

	// 验证State序列化
	var state domain.GameState
	err = json.Unmarshal([]byte(model.State), &state)
	require.NoError(t, err)
	assert.Equal(t, session.State.CurrentSceneID, state.CurrentSceneID)
	assert.Equal(t, session.State.ChaosPool, state.ChaosPool)
	assert.Equal(t, session.State.LooseEnds, state.LooseEnds)

	// 测试转换为领域模型
	converted, err := repo.toDomain(model)
	require.NoError(t, err)
	assert.Equal(t, session.ID, converted.ID)
	assert.Equal(t, session.AgentID, converted.AgentID)
	assert.Equal(t, session.ScenarioID, converted.ScenarioID)
	assert.Equal(t, session.Phase, converted.Phase)
	assert.Equal(t, session.State.CurrentSceneID, converted.State.CurrentSceneID)
	assert.Equal(t, session.State.ChaosPool, converted.State.ChaosPool)
	assert.Equal(t, session.State.LooseEnds, converted.State.LooseEnds)
}

// TestSessionRepository_ConcurrentAccess 测试并发访问
func TestSessionRepository_ConcurrentAccess(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 先创建会话
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// 并发更新会话
	var wg sync.WaitGroup
	concurrency := 10
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			// 获取会话
			s, err := repo.GetByID(ctx, session.ID)
			if err != nil {
				t.Logf("Failed to get session: %v", err)
				return
			}

			// 修改会话
			s.State.ChaosPool = index
			s.State.LooseEnds = index * 2

			// 更新会话
			err = repo.Update(ctx, s)
			if err != nil {
				t.Logf("Failed to update session: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证会话仍然存在且可以获取
	retrieved, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, session.ID, retrieved.ID)
}

// TestSessionRepository_ConcurrentReadWrite 测试并发读写
func TestSessionRepository_ConcurrentReadWrite(t *testing.T) {
	db := setupSessionTestDB(t)
	redis := setupSessionTestRedis(t)
	logger, _ := zap.NewDevelopment()
	repo := NewSessionRepository(db, redis, logger)

	ctx := context.Background()
	session := createTestSession()

	// 先创建会话
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// 并发读写
	var wg sync.WaitGroup
	readers := 20
	writers := 5

	// 启动读取goroutines
	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := repo.GetByID(ctx, session.ID)
				if err != nil {
					t.Logf("Read error: %v", err)
				}
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// 启动写入goroutines
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				s, err := repo.GetByID(ctx, session.ID)
				if err != nil {
					t.Logf("Write get error: %v", err)
					continue
				}
				s.State.ChaosPool = index*10 + j
				err = repo.Update(ctx, s)
				if err != nil {
					t.Logf("Write update error: %v", err)
				}
				time.Sleep(time.Millisecond * 2)
			}
		}(i)
	}

	wg.Wait()

	// 验证会话仍然存在
	retrieved, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}
