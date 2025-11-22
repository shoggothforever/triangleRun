package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

type AgentService interface {
	// 角色管理
	CreateAgent(req *CreateAgentRequest) (*domain.Agent, error)
	GetAgent(agentID string) (*domain.Agent, error)
	UpdateAgent(agent *domain.Agent) error
	DeleteAgent(agentID string) error
	ListAgents() ([]*domain.Agent, error)

	// ARC管理
	SetAnomaly(agentID string, anomalyType string) error
	SetReality(agentID string, realityType string) error
	SetCareer(agentID string, careerType string) error

	// 资质保证
	SpendQA(agentID, quality string, amount int) error
	RestoreQA(agentID string) error

	// 人际关系
	AddRelationship(agentID string, rel *domain.Relationship) error
	UpdateRelationship(agentID, relID string, connection int) error

	// 绩效
	AddCommendations(agentID string, amount int) error
	AddReprimands(agentID string, amount int) error
	UpdateRating(agentID string) error
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

func (s *agentService) UpdateAgent(agent *domain.Agent) error {
	if _, exists := s.agents[agent.ID]; !exists {
		return domain.NewGameError(domain.ErrNotFound, "角色不存在")
	}

	// 验证ARC
	if err := agent.ValidateARC(); err != nil {
		return err
	}

	agent.UpdatedAt = time.Now()
	s.agents[agent.ID] = agent
	return nil
}

func (s *agentService) DeleteAgent(agentID string) error {
	if _, exists := s.agents[agentID]; !exists {
		return domain.NewGameError(domain.ErrNotFound, "角色不存在")
	}

	delete(s.agents, agentID)
	return nil
}

func (s *agentService) ListAgents() ([]*domain.Agent, error) {
	agents := make([]*domain.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

// SetAnomaly 设置异常体类型
func (s *agentService) SetAnomaly(agentID string, anomalyType string) error {
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

	return nil
}

// SetReality 设置现实类型
func (s *agentService) SetReality(agentID string, realityType string) error {
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

	return nil
}

// SetCareer 设置职能类型
func (s *agentService) SetCareer(agentID string, careerType string) error {
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

	return nil
}

// SpendQA 花费资质保证
func (s *agentService) SpendQA(agentID, quality string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	if err := agent.SpendQA(quality, amount); err != nil {
		return err
	}

	agent.UpdatedAt = time.Now()
	return nil
}

// RestoreQA 恢复资质保证
func (s *agentService) RestoreQA(agentID string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.RestoreQA()
	agent.UpdatedAt = time.Now()
	return nil
}

// AddRelationship 添加人际关系
func (s *agentService) AddRelationship(agentID string, rel *domain.Relationship) error {
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
	return nil
}

// UpdateRelationship 更新人际关系连结点数
func (s *agentService) UpdateRelationship(agentID, relID string, connection int) error {
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
	return nil
}

// AddCommendations 添加嘉奖
func (s *agentService) AddCommendations(agentID string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.AddCommendations(amount)
	agent.UpdatedAt = time.Now()
	return nil
}

// AddReprimands 添加申诫
func (s *agentService) AddReprimands(agentID string, amount int) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.AddReprimands(amount)
	agent.UpdatedAt = time.Now()
	return nil
}

// UpdateRating 更新机构评级
func (s *agentService) UpdateRating(agentID string) error {
	agent, err := s.GetAgent(agentID)
	if err != nil {
		return err
	}

	agent.Rating = domain.GetRating(agent.Reprimands)
	agent.UpdatedAt = time.Now()
	return nil
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
