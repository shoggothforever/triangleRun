package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SessionRepository 会话仓储接口
type SessionRepository interface {
	// CRUD操作
	Create(ctx context.Context, session *domain.GameSession) error
	GetByID(ctx context.Context, id string) (*domain.GameSession, error)
	Update(ctx context.Context, session *domain.GameSession) error
	Delete(ctx context.Context, id string) error
	ListByAgent(ctx context.Context, agentID string) ([]*domain.GameSession, error)

	// 事务支持
	WithTx(tx *gorm.DB) SessionRepository
}

type sessionRepository struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
	ttl    time.Duration
	// 并发控制：为每个会话维护一个锁
	locks *sync.Map // map[string]*sync.RWMutex
}

// NewSessionRepository 创建会话仓储实例
func NewSessionRepository(db *gorm.DB, redis *redis.Client, logger *zap.Logger) SessionRepository {
	return &sessionRepository{
		db:     db,
		redis:  redis,
		logger: logger,
		ttl:    24 * time.Hour, // 缓存24小时
		locks:  &sync.Map{},
	}
}

// getLock 获取会话的读写锁
func (r *sessionRepository) getLock(sessionID string) *sync.RWMutex {
	lock, _ := r.locks.LoadOrStore(sessionID, &sync.RWMutex{})
	return lock.(*sync.RWMutex)
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, session *domain.GameSession) error {
	model, err := r.toModel(session)
	if err != nil {
		return fmt.Errorf("failed to convert session to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	session.ID = model.ID

	if err := r.cacheSession(ctx, session); err != nil {
		r.logger.Warn("failed to cache session", zap.Error(err), zap.String("session_id", session.ID))
	}

	return nil
}

// GetByID 根据ID获取会话
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.GameSession, error) {
	lock := r.getLock(id)
	lock.RLock()
	defer lock.RUnlock()

	session, err := r.getFromCache(ctx, id)
	if err == nil && session != nil {
		return session, nil
	}

	var model database.GameSessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.NewGameError(domain.ErrNotFound, "会话不存在").
				WithDetails("session_id", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session, err = r.toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to session: %w", err)
	}

	if err := r.cacheSession(ctx, session); err != nil {
		r.logger.Warn("failed to cache session", zap.Error(err), zap.String("session_id", id))
	}

	return session, nil
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, session *domain.GameSession) error {
	lock := r.getLock(session.ID)
	lock.Lock()
	defer lock.Unlock()

	model, err := r.toModel(session)
	if err != nil {
		return fmt.Errorf("failed to convert session to model: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&database.GameSessionModel{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"agent_id":    model.AgentID,
			"scenario_id": model.ScenarioID,
			"phase":       model.Phase,
			"state":       model.State,
			"updated_at":  time.Now().Unix(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewGameError(domain.ErrNotFound, "会话不存在").
			WithDetails("session_id", session.ID)
	}

	if err := r.cacheSession(ctx, session); err != nil {
		r.logger.Warn("failed to update cache", zap.Error(err), zap.String("session_id", session.ID))
	}

	return nil
}

// Delete 删除会话
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	lock := r.getLock(id)
	lock.Lock()
	defer lock.Unlock()

	result := r.db.WithContext(ctx).Delete(&database.GameSessionModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewGameError(domain.ErrNotFound, "会话不存在").
			WithDetails("session_id", id)
	}

	if err := r.invalidateCache(ctx, id); err != nil {
		r.logger.Warn("failed to invalidate cache", zap.Error(err), zap.String("session_id", id))
	}

	r.locks.Delete(id)

	return nil
}

// ListByAgent 列出指定角色的所有会话
func (r *sessionRepository) ListByAgent(ctx context.Context, agentID string) ([]*domain.GameSession, error) {
	var models []database.GameSessionModel
	if err := r.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := make([]*domain.GameSession, 0, len(models))
	for _, model := range models {
		session, err := r.toDomain(&model)
		if err != nil {
			r.logger.Warn("failed to convert model to session", zap.Error(err), zap.String("session_id", model.ID))
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// WithTx 使用事务
func (r *sessionRepository) WithTx(tx *gorm.DB) SessionRepository {
	return &sessionRepository{
		db:     tx,
		redis:  r.redis,
		logger: r.logger,
		ttl:    r.ttl,
		locks:  r.locks,
	}
}

// toModel 将领域模型转换为数据库模型
func (r *sessionRepository) toModel(session *domain.GameSession) (*database.GameSessionModel, error) {
	stateJSON, err := json.Marshal(session.State)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	return &database.GameSessionModel{
		ID:         session.ID,
		AgentID:    session.AgentID,
		ScenarioID: session.ScenarioID,
		Phase:      string(session.Phase),
		State:      string(stateJSON),
		CreatedAt:  session.CreatedAt.Unix(),
		UpdatedAt:  session.UpdatedAt.Unix(),
	}, nil
}

// toDomain 将数据库模型转换为领域模型
func (r *sessionRepository) toDomain(model *database.GameSessionModel) (*domain.GameSession, error) {
	var state domain.GameState
	if err := json.Unmarshal([]byte(model.State), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &domain.GameSession{
		ID:         model.ID,
		AgentID:    model.AgentID,
		ScenarioID: model.ScenarioID,
		Phase:      domain.GamePhase(model.Phase),
		State:      &state,
		CreatedAt:  time.Unix(model.CreatedAt, 0),
		UpdatedAt:  time.Unix(model.UpdatedAt, 0),
	}, nil
}

// cacheSession 缓存会话到Redis
func (r *sessionRepository) cacheSession(ctx context.Context, session *domain.GameSession) error {
	if r.redis == nil {
		return nil
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.ID)
	return r.redis.Set(ctx, key, data, r.ttl).Err()
}

// getFromCache 从Redis缓存获取会话
func (r *sessionRepository) getFromCache(ctx context.Context, id string) (*domain.GameSession, error) {
	if r.redis == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("session:%s", id)
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var session domain.GameSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// invalidateCache 使缓存失效
func (r *sessionRepository) invalidateCache(ctx context.Context, id string) error {
	if r.redis == nil {
		return nil
	}

	key := fmt.Sprintf("session:%s", id)
	return r.redis.Del(ctx, key).Err()
}
