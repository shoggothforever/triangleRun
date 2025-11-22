package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// GameService 游戏会话服务接口
type GameService interface {
	// 会话管理
	CreateSession(agentID, scenarioID string) (*domain.GameSession, error)
	GetSession(sessionID string) (*domain.GameSession, error)
	SaveSession(session *domain.GameSession) error
	DeleteSession(sessionID string) error
	ListSessions() ([]*domain.GameSession, error)

	// 游戏流程
	StartMorningPhase(sessionID string) (*MorningPhaseResult, error)
	StartInvestigationPhase(sessionID string) (*InvestigationPhaseResult, error)
	StartEncounterPhase(sessionID string) (*EncounterPhaseResult, error)

	// 阶段转换
	TransitionPhase(sessionID string, toPhase domain.GamePhase) error

	// 状态管理
	UpdateState(sessionID string, updateFn func(*domain.GameState) error) error
	GetState(sessionID string) (*domain.GameState, error)
}

// MorningPhaseResult 晨会阶段结果
type MorningPhaseResult struct {
	SessionID   string                 `json:"session_id"`
	Briefing    *domain.Briefing       `json:"briefing"`
	Goals       []*domain.OptionalGoal `json:"goals"`
	Description string                 `json:"description"`
}

// InvestigationPhaseResult 调查阶段结果
type InvestigationPhaseResult struct {
	SessionID       string   `json:"session_id"`
	CurrentSceneID  string   `json:"current_scene_id"`
	AvailableScenes []string `json:"available_scenes"`
	Description     string   `json:"description"`
}

// EncounterPhaseResult 遭遇阶段结果
type EncounterPhaseResult struct {
	SessionID   string `json:"session_id"`
	AnomalyName string `json:"anomaly_name"`
	Description string `json:"description"`
}

// gameService 游戏会话服务实现
type gameService struct {
	sessions map[string]*domain.GameSession
	agents   map[string]*domain.Agent // 用于测试的角色存储
	mu       sync.RWMutex             // 并发控制
}

// NewGameService 创建游戏会话服务
func NewGameService() GameService {
	return &gameService{
		sessions: make(map[string]*domain.GameSession),
		agents:   make(map[string]*domain.Agent),
	}
}

// SaveAgent 保存角色（用于测试）
func (s *gameService) SaveAgent(agent *domain.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "角色不能为空")
	}

	s.agents[agent.ID] = agent
	return nil
}

// CreateSession 创建游戏会话
func (s *gameService) CreateSession(agentID, scenarioID string) (*domain.GameSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建初始游戏状态
	state := &domain.GameState{
		CurrentSceneID:    "",
		VisitedScenes:     make(map[string]bool),
		CollectedClues:    []string{},
		UnlockedLocations: []string{},
		DomainUnlocked:    false,
		NPCStates:         make(map[string]*domain.NPCState),
		ChaosPool:         0,
		LooseEnds:         0,
		LocationOverloads: make(map[string]int),
		AnomalyStatus:     "未知",
		MissionOutcome:    "进行中",
	}

	// 创建会话
	session := &domain.GameSession{
		ID:         uuid.New().String(),
		AgentID:    agentID,
		ScenarioID: scenarioID,
		Phase:      domain.PhaseMorning,
		State:      state,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存会话
	s.sessions[session.ID] = session

	return session, nil
}

// GetSession 获取游戏会话
func (s *gameService) GetSession(sessionID string) (*domain.GameSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, domain.NewGameError(domain.ErrNotFound, "游戏会话不存在").
			WithDetails("session_id", sessionID)
	}

	return session, nil
}

// SaveSession 保存游戏会话（如果不存在则创建）
func (s *gameService) SaveSession(session *domain.GameSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "游戏会话不能为空")
	}

	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session

	return nil
}

// DeleteSession 删除游戏会话
func (s *gameService) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return domain.NewGameError(domain.ErrNotFound, "游戏会话不存在").
			WithDetails("session_id", sessionID)
	}

	delete(s.sessions, sessionID)
	return nil
}

// ListSessions 列出所有游戏会话
func (s *gameService) ListSessions() ([]*domain.GameSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*domain.GameSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// StartMorningPhase 开始晨会阶段
func (s *gameService) StartMorningPhase(sessionID string) (*MorningPhaseResult, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 验证阶段转换
	if session.Phase != domain.PhaseMorning {
		return nil, domain.NewGameError(domain.ErrInvalidPhase, "当前不在晨会阶段").
			WithDetails("current_phase", session.Phase).
			WithDetails("expected_phase", domain.PhaseMorning)
	}

	// 创建晨会结果
	result := &MorningPhaseResult{
		SessionID: sessionID,
		Briefing: &domain.Briefing{
			Summary:    "任务简报",
			Objectives: []string{"捕获异常体", "最小化散逸端"},
			Warnings:   []string{"注意安全", "遵守规则"},
		},
		Goals: []*domain.OptionalGoal{
			{
				ID:          uuid.New().String(),
				Description: "完成可选目标",
				Reward:      3,
			},
		},
		Description: "晨会开始，总经理正在介绍任务详情...",
	}

	return result, nil
}

// StartInvestigationPhase 开始调查阶段
func (s *gameService) StartInvestigationPhase(sessionID string) (*InvestigationPhaseResult, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 验证阶段转换
	if session.Phase != domain.PhaseInvestigation {
		return nil, domain.NewGameError(domain.ErrInvalidPhase, "当前不在调查阶段").
			WithDetails("current_phase", session.Phase).
			WithDetails("expected_phase", domain.PhaseInvestigation)
	}

	// 创建调查结果
	result := &InvestigationPhaseResult{
		SessionID:       sessionID,
		CurrentSceneID:  session.State.CurrentSceneID,
		AvailableScenes: session.State.UnlockedLocations,
		Description:     "调查阶段开始，你可以开始探索各个地点...",
	}

	return result, nil
}

// StartEncounterPhase 开始遭遇阶段
func (s *gameService) StartEncounterPhase(sessionID string) (*EncounterPhaseResult, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 验证阶段转换
	if session.Phase != domain.PhaseEncounter {
		return nil, domain.NewGameError(domain.ErrInvalidPhase, "当前不在遭遇阶段").
			WithDetails("current_phase", session.Phase).
			WithDetails("expected_phase", domain.PhaseEncounter)
	}

	// 创建遭遇结果
	result := &EncounterPhaseResult{
		SessionID:   sessionID,
		AnomalyName: "未知异常体",
		Description: "遭遇阶段开始，你进入了异常体的领域...",
	}

	return result, nil
}

// TransitionPhase 转换游戏阶段
func (s *gameService) TransitionPhase(sessionID string, toPhase domain.GamePhase) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 验证阶段转换的有效性
	if !isValidPhaseTransition(session.Phase, toPhase) {
		return domain.NewGameError(domain.ErrInvalidPhase, "无效的阶段转换").
			WithDetails("from_phase", session.Phase).
			WithDetails("to_phase", toPhase)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新阶段
	session.Phase = toPhase
	session.UpdatedAt = time.Now()

	return nil
}

// UpdateState 更新游戏状态
func (s *gameService) UpdateState(sessionID string, updateFn func(*domain.GameState) error) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 执行更新函数
	if err := updateFn(session.State); err != nil {
		return err
	}

	session.UpdatedAt = time.Now()
	return nil
}

// GetState 获取游戏状态
func (s *gameService) GetState(sessionID string) (*domain.GameState, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.State, nil
}

// isValidPhaseTransition 验证阶段转换是否有效
func isValidPhaseTransition(from, to domain.GamePhase) bool {
	// 定义有效的阶段转换
	validTransitions := map[domain.GamePhase][]domain.GamePhase{
		domain.PhaseMorning:       {domain.PhaseInvestigation},
		domain.PhaseInvestigation: {domain.PhaseEncounter},
		domain.PhaseEncounter:     {domain.PhaseAftermath},
		domain.PhaseAftermath:     {domain.PhaseMorning}, // 可以开始新任务
	}

	allowedPhases, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, phase := range allowedPhases {
		if phase == to {
			return true
		}
	}

	return false
}
