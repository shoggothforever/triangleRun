package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AgentRepository 角色仓储接口
type AgentRepository interface {
	// CRUD操作
	Create(ctx context.Context, agent *domain.Agent) error
	GetByID(ctx context.Context, id string) (*domain.Agent, error)
	Update(ctx context.Context, agent *domain.Agent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.Agent, error)

	// 事务支持
	WithTx(tx *gorm.DB) AgentRepository
}

type agentRepository struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// NewAgentRepository 创建角色仓储实例
func NewAgentRepository(db *gorm.DB, redis *redis.Client, logger *zap.Logger) AgentRepository {
	return &agentRepository{
		db:     db,
		redis:  redis,
		logger: logger,
		ttl:    1 * time.Hour, // 缓存1小时
	}
}

// Create 创建角色
func (r *agentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	// 转换为数据库模型
	model, err := r.toModel(agent)
	if err != nil {
		return fmt.Errorf("failed to convert agent to model: %w", err)
	}

	// 保存到数据库
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// 更新ID（如果数据库生成了新ID）
	agent.ID = model.ID

	// 缓存到Redis
	if err := r.cacheAgent(ctx, agent); err != nil {
		r.logger.Warn("failed to cache agent", zap.Error(err), zap.String("agent_id", agent.ID))
	}

	return nil
}

// GetByID 根据ID获取角色
func (r *agentRepository) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	// 先尝试从缓存获取
	agent, err := r.getFromCache(ctx, id)
	if err == nil && agent != nil {
		return agent, nil
	}

	// 从数据库获取
	var model database.AgentModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.NewGameError(domain.ErrNotFound, "角色不存在").
				WithDetails("agent_id", id)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// 转换为领域模型
	agent, err = r.toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to agent: %w", err)
	}

	// 缓存到Redis
	if err := r.cacheAgent(ctx, agent); err != nil {
		r.logger.Warn("failed to cache agent", zap.Error(err), zap.String("agent_id", id))
	}

	return agent, nil
}

// Update 更新角色
func (r *agentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	// 转换为数据库模型
	model, err := r.toModel(agent)
	if err != nil {
		return fmt.Errorf("failed to convert agent to model: %w", err)
	}

	// 更新数据库
	result := r.db.WithContext(ctx).Model(&database.AgentModel{}).
		Where("id = ?", agent.ID).
		Updates(map[string]interface{}{
			"name":          model.Name,
			"pronouns":      model.Pronouns,
			"anomaly_type":  model.AnomalyType,
			"reality_type":  model.RealityType,
			"career_type":   model.CareerType,
			"qa":            model.QA,
			"relationships": model.Relationships,
			"commendations": model.Commendations,
			"reprimands":    model.Reprimands,
			"rating":        model.Rating,
			"alive":         model.Alive,
			"in_debt":       model.InDebt,
			"updated_at":    time.Now().Unix(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update agent: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewGameError(domain.ErrNotFound, "角色不存在").
			WithDetails("agent_id", agent.ID)
	}

	// 使缓存失效
	if err := r.invalidateCache(ctx, agent.ID); err != nil {
		r.logger.Warn("failed to invalidate cache", zap.Error(err), zap.String("agent_id", agent.ID))
	}

	return nil
}

// Delete 删除角色
func (r *agentRepository) Delete(ctx context.Context, id string) error {
	// 从数据库删除
	result := r.db.WithContext(ctx).Delete(&database.AgentModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete agent: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewGameError(domain.ErrNotFound, "角色不存在").
			WithDetails("agent_id", id)
	}

	// 使缓存失效
	if err := r.invalidateCache(ctx, id); err != nil {
		r.logger.Warn("failed to invalidate cache", zap.Error(err), zap.String("agent_id", id))
	}

	return nil
}

// List 列出所有角色
func (r *agentRepository) List(ctx context.Context) ([]*domain.Agent, error) {
	var models []database.AgentModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	agents := make([]*domain.Agent, 0, len(models))
	for _, model := range models {
		agent, err := r.toDomain(&model)
		if err != nil {
			r.logger.Warn("failed to convert model to agent", zap.Error(err), zap.String("agent_id", model.ID))
			continue
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// WithTx 使用事务
func (r *agentRepository) WithTx(tx *gorm.DB) AgentRepository {
	return &agentRepository{
		db:     tx,
		redis:  r.redis,
		logger: r.logger,
		ttl:    r.ttl,
	}
}

// toModel 将领域模型转换为数据库模型
func (r *agentRepository) toModel(agent *domain.Agent) (*database.AgentModel, error) {
	// 序列化QA
	qaJSON, err := json.Marshal(agent.QA)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QA: %w", err)
	}

	// 序列化Relationships
	relsJSON, err := json.Marshal(agent.Relationships)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relationships: %w", err)
	}

	return &database.AgentModel{
		ID:            agent.ID,
		Name:          agent.Name,
		Pronouns:      agent.Pronouns,
		AnomalyType:   agent.Anomaly.Type,
		RealityType:   agent.Reality.Type,
		CareerType:    agent.Career.Type,
		QA:            string(qaJSON),
		Relationships: string(relsJSON),
		Commendations: agent.Commendations,
		Reprimands:    agent.Reprimands,
		Rating:        agent.Rating,
		Alive:         agent.Alive,
		InDebt:        agent.InDebt,
		CreatedAt:     agent.CreatedAt.Unix(),
		UpdatedAt:     agent.UpdatedAt.Unix(),
	}, nil
}

// toDomain 将数据库模型转换为领域模型
func (r *agentRepository) toDomain(model *database.AgentModel) (*domain.Agent, error) {
	// 反序列化QA
	var qa map[string]int
	if err := json.Unmarshal([]byte(model.QA), &qa); err != nil {
		return nil, fmt.Errorf("failed to unmarshal QA: %w", err)
	}

	// 反序列化Relationships
	var relationships []*domain.Relationship
	if err := json.Unmarshal([]byte(model.Relationships), &relationships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relationships: %w", err)
	}

	// 创建基本的ARC组件（简化版本，实际应该从配置加载完整数据）
	anomaly := &domain.Anomaly{
		Type:      model.AnomalyType,
		Abilities: createDefaultAbilities(model.AnomalyType),
	}

	reality := &domain.Reality{
		Type: model.RealityType,
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
	}

	career := &domain.Career{
		Type: model.CareerType,
		QA:   qa,
	}

	return &domain.Agent{
		ID:            model.ID,
		Name:          model.Name,
		Pronouns:      model.Pronouns,
		Anomaly:       anomaly,
		Reality:       reality,
		Career:        career,
		QA:            qa,
		Relationships: relationships,
		Commendations: model.Commendations,
		Reprimands:    model.Reprimands,
		Rating:        model.Rating,
		Alive:         model.Alive,
		InDebt:        model.InDebt,
		CreatedAt:     time.Unix(model.CreatedAt, 0),
		UpdatedAt:     time.Unix(model.UpdatedAt, 0),
	}, nil
}

// cacheAgent 缓存角色到Redis
func (r *agentRepository) cacheAgent(ctx context.Context, agent *domain.Agent) error {
	if r.redis == nil {
		return nil
	}

	data, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	key := fmt.Sprintf("agent:%s", agent.ID)
	return r.redis.Set(ctx, key, data, r.ttl).Err()
}

// getFromCache 从Redis缓存获取角色
func (r *agentRepository) getFromCache(ctx context.Context, id string) (*domain.Agent, error) {
	if r.redis == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("agent:%s", id)
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var agent domain.Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
	}

	return &agent, nil
}

// invalidateCache 使缓存失效
func (r *agentRepository) invalidateCache(ctx context.Context, id string) error {
	if r.redis == nil {
		return nil
	}

	key := fmt.Sprintf("agent:%s", id)
	return r.redis.Del(ctx, key).Err()
}

// createDefaultAbilities 创建默认异常能力（辅助函数）
func createDefaultAbilities(anomalyType string) []*domain.AnomalyAbility {
	return []*domain.AnomalyAbility{
		{
			ID:          "ability-1",
			Name:        "能力1",
			AnomalyType: anomalyType,
			Trigger: &domain.AbilityTrigger{
				Type:        domain.TriggerAction,
				Description: "主动使用",
			},
			Roll: &domain.AbilityRoll{
				Quality:   domain.QualityFocus,
				DiceCount: 6,
				DiceType:  4,
			},
		},
		{
			ID:          "ability-2",
			Name:        "能力2",
			AnomalyType: anomalyType,
		},
		{
			ID:          "ability-3",
			Name:        "能力3",
			AnomalyType: anomalyType,
		},
	}
}
