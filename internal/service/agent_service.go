package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

type AgentService interface {
	CreateAgent(req *CreateAgentRequest) (*domain.Agent, error)
	GetAgent(agentID string) (*domain.Agent, error)
	ListAgents() ([]*domain.Agent, error)
}

type CreateAgentRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Pronouns      string                 `json:"pronouns"`
	AnomalyType   string                 `json:"anomaly_type" binding:"required"`
	RealityType   string                 `json:"reality_type" binding:"required"`
	CareerType    string                 `json:"career_type" binding:"required"`
	Relationships []*domain.Relationship `json:"relationships"`
}

type agentService struct {
	agents map[string]*domain.Agent // 简化实现，使用内存存储
}

func NewAgentService() AgentService {
	return &agentService{
		agents: make(map[string]*domain.Agent),
	}
}

func (s *agentService) CreateAgent(req *CreateAgentRequest) (*domain.Agent, error) {
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

	// 保存
	s.agents[agent.ID] = agent

	return agent, nil
}

func (s *agentService) GetAgent(agentID string) (*domain.Agent, error) {
	agent, exists := s.agents[agentID]
	if !exists {
		return nil, domain.NewGameError(domain.ErrNotFound, "角色不存在")
	}
	return agent, nil
}

func (s *agentService) ListAgents() ([]*domain.Agent, error) {
	agents := make([]*domain.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func createDefaultAbilities(anomalyType string) []*domain.AnomalyAbility {
	return []*domain.AnomalyAbility{
		{
			ID:          uuid.New().String(),
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
			ID:          uuid.New().String(),
			Name:        "能力2",
			AnomalyType: anomalyType,
		},
		{
			ID:          uuid.New().String(),
			Name:        "能力3",
			AnomalyType: anomalyType,
		},
	}
}

func getDefaultQA(careerType string) map[string]int {
	// 简化实现：所有职能都是平均分配
	return map[string]int{
		domain.QualityFocus:      1,
		domain.QualityEmpathy:    1,
		domain.QualityPresence:   1,
		domain.QualityDeception:  1,
		domain.QualityInitiative: 1,
		domain.QualityProfession: 1,
		domain.QualityVitality:   1,
		domain.QualityGrit:       1,
		domain.QualitySubtlety:   1,
	}
}
