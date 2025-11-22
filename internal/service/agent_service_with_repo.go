package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/repository"
)

// agentServiceWithRepo 使用仓储的角色服务实现
type agentServiceWithRepo struct {
	repo repository.AgentRepository
}

// NewAgentServiceWithRepo 创建使用仓储的角色服务
func NewAgentServiceWithRepo(repo repository.AgentRepository) AgentService {
	return &agentServiceWithRepo{
		repo: repo,
	}
}

func (s *agentServiceWithRepo) CreateAgent(req *CreateAgentRequest) (*domain.Agent, error) {
	ctx := context.Background()

	// 创建角色
	agent := &domain.Agent{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Pronouns: req.Pronouns,
		Anomaly: &domain.Anomaly{
			Type:      req.AnomalyType,
			Abilities: createDefaultAbilities(req.AnomalyType),
		},
		Reality: &domain.Reality{
			Type: req.RealityType,
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
			Type: req.CareerType,
			QA:   getDefaultQA(req.CareerType),
		},
		QA:            getDefaultQA(req.CareerType),
		Relationships: req.Relationships,
		Commendations: 0,
		Reprimands:    0,
		Rating:        domain.RatingExcellent,
		Alive:         true,
		InDebt:        false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 如果没有提供人际关系，创建默认的
	if len(agent.Relationships) == 0 {
		agent.Relationships = []*domain.Relationship{
			{ID: uuid.New().String(), Name: "关系1", Connection: 6},
			{ID: uuid.New().String(), Name: "关系2", Connection: 3},
			{ID: uuid.New().String(), Name: "关系3", Connection: 3},
		}
	}

	// 验证ARC
	if err := agent.ValidateARC(); err != nil {
		return nil, err
	}

	// 保存到仓储
	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

func (s *agentServiceWithRepo) GetAgent(agentID string) (*domain.Agent, error) {
	ctx := context.Background()
	return s.repo.GetByID(ctx, agentID)
}

func (s *agentServiceWithRepo) UpdateAgent(agent *domain.Agent) error {
	ctx := context.Background()

	// 验证ARC
	if err := agent.ValidateARC(); err != nil {
		return err
	}

	agent.UpdatedAt = time.Now()
	return s.repo.Update(ctx, agent)
}

func (s *agentServiceWithRepo) DeleteAgent(agentID string) error {
	ctx := context.Background()
	return s.repo.Delete(ctx, agentID)
}

func (s *agentServiceWithRepo) ListAgents() ([]*domain.Agent, error) {
	ctx := context.Background()
	return s.repo.List(ctx)
}

func (s *agentServiceWithRepo) SetAnomaly(agentID string, anomalyType string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	// 验证异常体类型
	validType := false
	for _, t := range domain.AllAnomalyTypes {
		if t == anomalyType {
			validType = true
			break
		}
	}
	if !validType {
		return domain.NewGameError(domain.ErrInvalidARC, "无效的异常体类型").
			WithDetails("type", anomalyType)
	}

	agent.Anomaly = &domain.Anomaly{
		Type:      anomalyType,
		Abilities: createDefaultAbilities(anomalyType),
	}
	agent.UpdatedAt = time.Now()

	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) SetReality(agentID string, realityType string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	// 验证现实类型
	validType := false
	for _, t := range domain.AllRealityTypes {
		if t == realityType {
			validType = true
			break
		}
	}
	if !validType {
		return domain.NewGameError(domain.ErrInvalidARC, "无效的现实类型").
			WithDetails("type", realityType)
	}

	agent.Reality = &domain.Reality{
		Type: realityType,
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
	agent.UpdatedAt = time.Now()

	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) SetCareer(agentID string, careerType string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	// 验证职能类型
	validType := false
	for _, t := range domain.AllCareerTypes {
		if t == careerType {
			validType = true
			break
		}
	}
	if !validType {
		return domain.NewGameError(domain.ErrInvalidARC, "无效的职能类型").
			WithDetails("type", careerType)
	}

	qa := getDefaultQA(careerType)
	agent.Career = &domain.Career{
		Type: careerType,
		QA:   qa,
	}
	agent.QA = qa
	agent.UpdatedAt = time.Now()

	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) SpendQA(agentID, quality string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	if err := agent.SpendQA(quality, amount); err != nil {
		return err
	}

	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) RestoreQA(agentID string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.RestoreQA()
	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) AddRelationship(agentID string, rel *domain.Relationship) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	// 生成ID如果没有
	if rel.ID == "" {
		rel.ID = uuid.New().String()
	}

	agent.Relationships = append(agent.Relationships, rel)
	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) UpdateRelationship(agentID, relID string, connection int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	found := false
	for _, rel := range agent.Relationships {
		if rel.ID == relID {
			rel.Connection = connection
			found = true
			break
		}
	}

	if !found {
		return domain.NewGameError(domain.ErrNotFound, "人际关系不存在").
			WithDetails("relationship_id", relID)
	}

	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) AddCommendations(agentID string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.AddCommendations(amount)
	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) AddReprimands(agentID string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.AddReprimands(amount)
	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}

func (s *agentServiceWithRepo) UpdateRating(agentID string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.Rating = domain.GetRating(agent.Reprimands)
	agent.UpdatedAt = time.Now()
	return s.UpdateAgent(agent)
}
