package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// SaveService 存档服务接口
type SaveService interface {
	// 存档管理
	CreateSave(sessionID, name string) (*SaveSnapshot, error)
	GetSave(saveID string) (*SaveSnapshot, error)
	ListSaves(sessionID string) ([]*SaveMetadata, error)
	DeleteSave(saveID string) error
	LoadSave(saveID string) (*domain.GameSession, error)

	// 序列化和反序列化
	SerializeSession(session *domain.GameSession) ([]byte, error)
	DeserializeSession(data []byte) (*domain.GameSession, error)

	// 版本兼容性
	ValidateVersion(data []byte) error
}

// SaveSnapshot 存档快照
type SaveSnapshot struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Snapshot  *domain.GameSession    `json:"snapshot"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// SaveMetadata 存档元数据
type SaveMetadata struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	AgentName string    `json:"agent_name"`
	Phase     string    `json:"phase"`
	CreatedAt time.Time `json:"created_at"`
}

// saveService 存档服务实现
type saveService struct {
	saves        map[string]*SaveSnapshot
	gameService  GameService
	agentService AgentService
	mu           sync.RWMutex
	version      string
}

// NewSaveService 创建存档服务
func NewSaveService(gameService GameService, agentService AgentService) SaveService {
	return &saveService{
		saves:        make(map[string]*SaveSnapshot),
		gameService:  gameService,
		agentService: agentService,
		version:      "1.0.0",
	}
}

// CreateSave 创建存档
func (s *saveService) CreateSave(sessionID, name string) (*SaveSnapshot, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 验证会话
	if session == nil {
		return nil, domain.NewGameError(domain.ErrNotFound, "游戏会话不存在").
			WithDetails("session_id", sessionID)
	}

	// 获取角色信息用于元数据
	agent, err := s.agentService.GetAgent(session.AgentID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建存档快照
	snapshot := &SaveSnapshot{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Name:      name,
		Version:   s.version,
		Snapshot:  session,
		Metadata: map[string]interface{}{
			"agent_name":  agent.Name,
			"scenario_id": session.ScenarioID,
			"phase":       session.Phase,
		},
		CreatedAt: time.Now(),
	}

	// 保存存档
	s.saves[snapshot.ID] = snapshot

	return snapshot, nil
}

// GetSave 获取存档
func (s *saveService) GetSave(saveID string) (*SaveSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, exists := s.saves[saveID]
	if !exists {
		return nil, domain.NewGameError(domain.ErrNotFound, "存档不存在").
			WithDetails("save_id", saveID)
	}

	return snapshot, nil
}

// ListSaves 列出存档
func (s *saveService) ListSaves(sessionID string) ([]*SaveMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var metadata []*SaveMetadata
	for _, save := range s.saves {
		if sessionID == "" || save.SessionID == sessionID {
			meta := &SaveMetadata{
				ID:        save.ID,
				SessionID: save.SessionID,
				Name:      save.Name,
				Version:   save.Version,
				CreatedAt: save.CreatedAt,
			}

			// 从元数据中提取信息
			if agentName, ok := save.Metadata["agent_name"].(string); ok {
				meta.AgentName = agentName
			}
			if phase, ok := save.Metadata["phase"].(domain.GamePhase); ok {
				meta.Phase = string(phase)
			}

			metadata = append(metadata, meta)
		}
	}

	return metadata, nil
}

// DeleteSave 删除存档
func (s *saveService) DeleteSave(saveID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.saves[saveID]; !exists {
		return domain.NewGameError(domain.ErrNotFound, "存档不存在").
			WithDetails("save_id", saveID)
	}

	delete(s.saves, saveID)
	return nil
}

// LoadSave 加载存档
func (s *saveService) LoadSave(saveID string) (*domain.GameSession, error) {
	snapshot, err := s.GetSave(saveID)
	if err != nil {
		return nil, err
	}

	// 验证版本兼容性
	if snapshot.Version != s.version {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "存档版本不兼容").
			WithDetails("save_version", snapshot.Version).
			WithDetails("current_version", s.version)
	}

	// 创建会话的深拷贝
	sessionCopy := &domain.GameSession{
		ID:         uuid.New().String(), // 生成新的会话ID
		AgentID:    snapshot.Snapshot.AgentID,
		ScenarioID: snapshot.Snapshot.ScenarioID,
		Phase:      snapshot.Snapshot.Phase,
		State:      copyGameState(snapshot.Snapshot.State),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return sessionCopy, nil
}

// SerializeSession 序列化游戏会话
func (s *saveService) SerializeSession(session *domain.GameSession) ([]byte, error) {
	if session == nil {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "游戏会话不能为空")
	}

	// 创建包含版本信息的包装结构
	wrapper := struct {
		Version string              `json:"version"`
		Session *domain.GameSession `json:"session"`
	}{
		Version: s.version,
		Session: session,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, domain.NewGameError(domain.ErrInternal, "序列化失败").
			WithDetails("error", err.Error())
	}

	return data, nil
}

// DeserializeSession 反序列化游戏会话
func (s *saveService) DeserializeSession(data []byte) (*domain.GameSession, error) {
	if len(data) == 0 {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "数据不能为空")
	}

	// 先验证版本
	if err := s.ValidateVersion(data); err != nil {
		return nil, err
	}

	// 解析包装结构
	var wrapper struct {
		Version string              `json:"version"`
		Session *domain.GameSession `json:"session"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "反序列化失败").
			WithDetails("error", err.Error())
	}

	if wrapper.Session == nil {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "存档数据损坏")
	}

	return wrapper.Session, nil
}

// ValidateVersion 验证版本兼容性
func (s *saveService) ValidateVersion(data []byte) error {
	if len(data) == 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "数据不能为空")
	}

	// 只解析版本字段
	var versionCheck struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &versionCheck); err != nil {
		return domain.NewGameError(domain.ErrDataCorrupted, "无法读取版本信息").
			WithDetails("error", err.Error())
	}

	if versionCheck.Version != s.version {
		return domain.NewGameError(domain.ErrDataCorrupted, "版本不兼容").
			WithDetails("save_version", versionCheck.Version).
			WithDetails("current_version", s.version)
	}

	return nil
}

// copyGameState 深拷贝游戏状态
func copyGameState(state *domain.GameState) *domain.GameState {
	if state == nil {
		return nil
	}

	// 拷贝visited scenes
	visitedScenes := make(map[string]bool)
	for k, v := range state.VisitedScenes {
		visitedScenes[k] = v
	}

	// 拷贝collected clues
	collectedClues := make([]string, len(state.CollectedClues))
	copy(collectedClues, state.CollectedClues)

	// 拷贝unlocked locations
	unlockedLocations := make([]string, len(state.UnlockedLocations))
	copy(unlockedLocations, state.UnlockedLocations)

	// 拷贝NPC states
	npcStates := make(map[string]*domain.NPCState)
	for k, v := range state.NPCStates {
		customData := make(map[string]interface{})
		for ck, cv := range v.CustomData {
			customData[ck] = cv
		}

		npcStates[k] = &domain.NPCState{
			ID:              v.ID,
			CurrentState:    v.CurrentState,
			AnomalyAffected: v.AnomalyAffected,
			Relationship:    v.Relationship,
			CustomData:      customData,
		}
	}

	// 拷贝location overloads
	locationOverloads := make(map[string]int)
	for k, v := range state.LocationOverloads {
		locationOverloads[k] = v
	}

	return &domain.GameState{
		CurrentSceneID:    state.CurrentSceneID,
		VisitedScenes:     visitedScenes,
		CollectedClues:    collectedClues,
		UnlockedLocations: unlockedLocations,
		DomainUnlocked:    state.DomainUnlocked,
		NPCStates:         npcStates,
		ChaosPool:         state.ChaosPool,
		LooseEnds:         state.LooseEnds,
		LocationOverloads: locationOverloads,
		AnomalyStatus:     state.AnomalyStatus,
		MissionOutcome:    state.MissionOutcome,
	}
}
