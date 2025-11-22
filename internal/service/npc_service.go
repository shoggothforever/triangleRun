package service

import (
	"sync"
	"time"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// NPCService NPC服务接口
type NPCService interface {
	// NPC状态管理
	LoadNPC(sessionID string, npcID string) (*NPCInfo, error)
	GetNPCState(sessionID string, npcID string) (*domain.NPCState, error)
	UpdateNPCState(sessionID string, npcID string, newState string) error
	SetAnomalyAffected(sessionID string, npcID string, affected bool) error

	// NPC关系追踪
	GetRelationship(sessionID string, npcID string) (int, error)
	ModifyRelationship(sessionID string, npcID string, delta int) error
	SetRelationship(sessionID string, npcID string, value int) error

	// 异常影响记录
	RecordAnomalyInfluence(sessionID string, npcID string, influence *AnomalyInfluence) error
	GetAnomalyInfluences(sessionID string, npcID string) ([]*AnomalyInfluence, error)
	HasAnomalyInfluence(sessionID string, npcID string) bool

	// NPC查询
	GetNPCsInScene(sessionID string, sceneID string) ([]*NPCInfo, error)
	GetAllNPCStates(sessionID string) (map[string]*domain.NPCState, error)
}

// NPCInfo NPC信息
type NPCInfo struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Personality     string                 `json:"personality"`
	CurrentState    string                 `json:"current_state"`
	AnomalyAffected bool                   `json:"anomaly_affected"`
	Relationship    int                    `json:"relationship"`
	Dialogues       []string               `json:"dialogues"`
	CustomData      map[string]interface{} `json:"custom_data"`
}

// AnomalyInfluence 异常影响记录
type AnomalyInfluence struct {
	Timestamp   time.Time              `json:"timestamp"`
	AnomalyType string                 `json:"anomaly_type"`
	Effect      string                 `json:"effect"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

// npcService NPC服务实现
type npcService struct {
	scenarioService ScenarioService
	gameService     GameService
	influences      map[string]map[string][]*AnomalyInfluence // sessionID -> npcID -> influences
	mu              sync.RWMutex
}

// NewNPCService 创建NPC服务
func NewNPCService(scenarioService ScenarioService, gameService GameService) NPCService {
	return &npcService{
		scenarioService: scenarioService,
		gameService:     gameService,
		influences:      make(map[string]map[string][]*AnomalyInfluence),
	}
}

// LoadNPC 加载NPC
func (s *npcService) LoadNPC(sessionID string, npcID string) (*NPCInfo, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 加载剧本
	scenario, err := s.scenarioService.LoadScenario(session.ScenarioID)
	if err != nil {
		return nil, err
	}

	// 在所有场景中查找NPC
	var npc *domain.NPC
	for _, scene := range scenario.Scenes {
		for _, n := range scene.NPCs {
			if n.ID == npcID {
				npc = n
				break
			}
		}
		if npc != nil {
			break
		}
	}

	if npc == nil {
		return nil, domain.NewGameError(domain.ErrNotFound, "NPC不存在").
			WithDetails("npc_id", npcID)
	}

	// 获取或初始化NPC状态
	npcState := s.getOrInitNPCState(session, npcID, npc.State)

	// 构建NPC信息
	npcInfo := &NPCInfo{
		ID:              npc.ID,
		Name:            npc.Name,
		Description:     npc.Description,
		Personality:     npc.Personality,
		CurrentState:    npcState.CurrentState,
		AnomalyAffected: npcState.AnomalyAffected,
		Relationship:    npcState.Relationship,
		Dialogues:       npc.Dialogues,
		CustomData:      npcState.CustomData,
	}

	return npcInfo, nil
}

// GetNPCState 获取NPC状态
func (s *npcService) GetNPCState(sessionID string, npcID string) (*domain.NPCState, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 检查NPC状态是否存在
	if session.State.NPCStates == nil {
		return nil, domain.NewGameError(domain.ErrNotFound, "NPC状态不存在").
			WithDetails("npc_id", npcID)
	}

	npcState, exists := session.State.NPCStates[npcID]
	if !exists {
		return nil, domain.NewGameError(domain.ErrNotFound, "NPC状态不存在").
			WithDetails("npc_id", npcID)
	}

	return npcState, nil
}

// UpdateNPCState 更新NPC状态
func (s *npcService) UpdateNPCState(sessionID string, npcID string, newState string) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 获取或初始化NPC状态
	npcState := s.getOrInitNPCState(session, npcID, newState)

	// 更新状态
	npcState.CurrentState = newState

	// 保存会话
	return s.gameService.SaveSession(session)
}

// SetAnomalyAffected 设置异常影响
func (s *npcService) SetAnomalyAffected(sessionID string, npcID string, affected bool) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 获取或初始化NPC状态
	npcState := s.getOrInitNPCState(session, npcID, "")

	// 更新异常影响状态
	npcState.AnomalyAffected = affected

	// 保存会话
	return s.gameService.SaveSession(session)
}

// GetRelationship 获取关系值
func (s *npcService) GetRelationship(sessionID string, npcID string) (int, error) {
	// 获取NPC状态
	npcState, err := s.GetNPCState(sessionID, npcID)
	if err != nil {
		return 0, err
	}

	return npcState.Relationship, nil
}

// ModifyRelationship 修改关系值
func (s *npcService) ModifyRelationship(sessionID string, npcID string, delta int) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 获取或初始化NPC状态
	npcState := s.getOrInitNPCState(session, npcID, "")

	// 修改关系值
	npcState.Relationship += delta

	// 保存会话
	return s.gameService.SaveSession(session)
}

// SetRelationship 设置关系值
func (s *npcService) SetRelationship(sessionID string, npcID string, value int) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 获取或初始化NPC状态
	npcState := s.getOrInitNPCState(session, npcID, "")

	// 设置关系值
	npcState.Relationship = value

	// 保存会话
	return s.gameService.SaveSession(session)
}

// RecordAnomalyInfluence 记录异常影响
func (s *npcService) RecordAnomalyInfluence(sessionID string, npcID string, influence *AnomalyInfluence) error {
	// 验证会话存在
	_, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 设置时间戳
	if influence.Timestamp.IsZero() {
		influence.Timestamp = time.Now()
	}

	// 记录影响
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.influences[sessionID]; !exists {
		s.influences[sessionID] = make(map[string][]*AnomalyInfluence)
	}

	if _, exists := s.influences[sessionID][npcID]; !exists {
		s.influences[sessionID][npcID] = make([]*AnomalyInfluence, 0)
	}

	s.influences[sessionID][npcID] = append(s.influences[sessionID][npcID], influence)

	// 标记NPC为受异常影响
	return s.SetAnomalyAffected(sessionID, npcID, true)
}

// GetAnomalyInfluences 获取异常影响记录
func (s *npcService) GetAnomalyInfluences(sessionID string, npcID string) ([]*AnomalyInfluence, error) {
	// 验证会话存在
	_, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.influences[sessionID]; !exists {
		return []*AnomalyInfluence{}, nil
	}

	if influences, exists := s.influences[sessionID][npcID]; exists {
		return influences, nil
	}

	return []*AnomalyInfluence{}, nil
}

// HasAnomalyInfluence 检查是否有异常影响
func (s *npcService) HasAnomalyInfluence(sessionID string, npcID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.influences[sessionID]; !exists {
		return false
	}

	if influences, exists := s.influences[sessionID][npcID]; exists {
		return len(influences) > 0
	}

	return false
}

// GetNPCsInScene 获取场景中的NPC
func (s *npcService) GetNPCsInScene(sessionID string, sceneID string) ([]*NPCInfo, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取场景
	scene, err := s.scenarioService.GetScene(session.ScenarioID, sceneID)
	if err != nil {
		return nil, err
	}

	// 加载所有NPC信息
	npcInfos := make([]*NPCInfo, 0, len(scene.NPCs))
	for _, npc := range scene.NPCs {
		npcInfo, err := s.LoadNPC(sessionID, npc.ID)
		if err != nil {
			continue // 跳过加载失败的NPC
		}
		npcInfos = append(npcInfos, npcInfo)
	}

	return npcInfos, nil
}

// GetAllNPCStates 获取所有NPC状态
func (s *npcService) GetAllNPCStates(sessionID string) (map[string]*domain.NPCState, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.State.NPCStates == nil {
		return make(map[string]*domain.NPCState), nil
	}

	return session.State.NPCStates, nil
}

// getOrInitNPCState 获取或初始化NPC状态
func (s *npcService) getOrInitNPCState(session *domain.GameSession, npcID string, defaultState string) *domain.NPCState {
	// 初始化NPCStates map
	if session.State.NPCStates == nil {
		session.State.NPCStates = make(map[string]*domain.NPCState)
	}

	// 检查NPC状态是否存在
	npcState, exists := session.State.NPCStates[npcID]
	if !exists {
		// 创建新的NPC状态
		npcState = &domain.NPCState{
			ID:              npcID,
			CurrentState:    defaultState,
			AnomalyAffected: false,
			Relationship:    0,
			CustomData:      make(map[string]interface{}),
		}
		session.State.NPCStates[npcID] = npcState
	}

	return npcState
}
