package service

import (
	"encoding/json"
	"sync"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// SceneService 场景服务接口
type SceneService interface {
	// 场景加载
	LoadScene(sessionID string, sceneID string) (*domain.Scene, error)
	GetCurrentScene(sessionID string) (*domain.Scene, error)

	// 场景状态管理
	SaveSceneState(sessionID string, sceneID string, state map[string]any) error
	GetSceneState(sessionID string, sceneID string) (map[string]any, error)
	RestoreSceneState(sessionID string, sceneID string) error

	// 场景切换
	TransitionToScene(sessionID string, targetSceneID string) error
	MarkSceneVisited(sessionID string, sceneID string) error

	// 场景对象交互
	InteractWithObject(sessionID string, objectID string, action string) (*InteractionResult, error)
	GetAvailableInteractions(sessionID string) ([]*Interaction, error)

	// 场景状态持久化
	PersistSceneStates(sessionID string) error
	LoadPersistedSceneStates(sessionID string) (map[string]map[string]any, error)
}

// InteractionResult 交互结果
type InteractionResult struct {
	Success         bool           `json:"success"`
	Description     string         `json:"description"`
	CluesGained     []string       `json:"clues_gained"`
	EventsTriggered []string       `json:"events_triggered"`
	StateChanges    map[string]any `json:"state_changes"`
}

// Interaction 可用交互
type Interaction struct {
	ObjectID    string   `json:"object_id"`
	ObjectName  string   `json:"object_name"`
	Actions     []string `json:"actions"`
	Description string   `json:"description"`
}

// sceneService 场景服务实现
type sceneService struct {
	scenarioService ScenarioService
	gameService     GameService
	sceneStates     map[string]map[string]map[string]any // sessionID -> sceneID -> state
	mu              sync.RWMutex
}

// NewSceneService 创建场景服务
func NewSceneService(scenarioService ScenarioService, gameService GameService) SceneService {
	return &sceneService{
		scenarioService: scenarioService,
		gameService:     gameService,
		sceneStates:     make(map[string]map[string]map[string]any),
	}
}

// LoadScene 加载场景
func (s *sceneService) LoadScene(sessionID string, sceneID string) (*domain.Scene, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 从剧本服务获取场景
	scene, err := s.scenarioService.GetScene(session.ScenarioID, sceneID)
	if err != nil {
		return nil, err
	}

	// 恢复场景状态
	s.mu.RLock()
	if sessionStates, exists := s.sceneStates[sessionID]; exists {
		if sceneState, exists := sessionStates[sceneID]; exists {
			// 创建场景副本并应用保存的状态
			sceneCopy := *scene
			sceneCopy.State = sceneState
			s.mu.RUnlock()
			return &sceneCopy, nil
		}
	}
	s.mu.RUnlock()

	return scene, nil
}

// GetCurrentScene 获取当前场景
func (s *sceneService) GetCurrentScene(sessionID string) (*domain.Scene, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.State == nil || session.State.CurrentSceneID == "" {
		return nil, domain.NewGameError(domain.ErrInvalidState, "当前场景未设置")
	}

	return s.LoadScene(sessionID, session.State.CurrentSceneID)
}

// SaveSceneState 保存场景状态
func (s *sceneService) SaveSceneState(sessionID string, sceneID string, state map[string]any) error {
	if state == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "场景状态不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保会话状态映射存在
	if _, exists := s.sceneStates[sessionID]; !exists {
		s.sceneStates[sessionID] = make(map[string]map[string]any)
	}

	// 保存场景状态
	s.sceneStates[sessionID][sceneID] = state

	return nil
}

// GetSceneState 获取场景状态
func (s *sceneService) GetSceneState(sessionID string, sceneID string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sessionStates, exists := s.sceneStates[sessionID]; exists {
		if sceneState, exists := sessionStates[sceneID]; exists {
			return sceneState, nil
		}
	}

	// 如果没有保存的状态，返回空状态
	return make(map[string]any), nil
}

// RestoreSceneState 恢复场景状态
func (s *sceneService) RestoreSceneState(sessionID string, sceneID string) error {
	// 获取保存的状态
	state, err := s.GetSceneState(sessionID, sceneID)
	if err != nil {
		return err
	}

	// 加载场景并应用状态
	scene, err := s.LoadScene(sessionID, sceneID)
	if err != nil {
		return err
	}

	scene.State = state
	return nil
}

// TransitionToScene 切换到场景
func (s *sceneService) TransitionToScene(sessionID string, targetSceneID string) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 验证目标场景存在
	_, err = s.scenarioService.GetScene(session.ScenarioID, targetSceneID)
	if err != nil {
		return err
	}

	// 保存当前场景状态
	if session.State.CurrentSceneID != "" {
		currentScene, err := s.GetCurrentScene(sessionID)
		if err == nil && currentScene != nil {
			_ = s.SaveSceneState(sessionID, session.State.CurrentSceneID, currentScene.State)
		}
	}

	// 更新当前场景
	session.State.CurrentSceneID = targetSceneID

	// 标记场景为已访问
	if session.State.VisitedScenes == nil {
		session.State.VisitedScenes = make(map[string]bool)
	}
	session.State.VisitedScenes[targetSceneID] = true

	// 保存会话
	return s.gameService.SaveSession(session)
}

// MarkSceneVisited 标记场景为已访问
func (s *sceneService) MarkSceneVisited(sessionID string, sceneID string) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 标记场景为已访问
	if session.State.VisitedScenes == nil {
		session.State.VisitedScenes = make(map[string]bool)
	}
	session.State.VisitedScenes[sceneID] = true

	// 保存会话
	return s.gameService.SaveSession(session)
}

// InteractWithObject 与对象交互
func (s *sceneService) InteractWithObject(sessionID string, objectID string, action string) (*InteractionResult, error) {
	// 获取当前场景
	scene, err := s.GetCurrentScene(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	result := &InteractionResult{
		Success:         false,
		CluesGained:     []string{},
		EventsTriggered: []string{},
		StateChanges:    make(map[string]any),
	}

	// 检查是否是线索交互
	for _, clue := range scene.Clues {
		if clue.ID == objectID {
			// 检查线索需求
			if s.scenarioService.CheckClueRequirements(clue, session.State) {
				result.Success = true
				result.Description = clue.Description
				result.CluesGained = append(result.CluesGained, clue.ID)

				// 添加线索到已收集列表
				session.State.CollectedClues = append(session.State.CollectedClues, clue.ID)

				// 解锁新场景
				for _, unlockID := range clue.Unlocks {
					if !contains(session.State.UnlockedLocations, unlockID) {
						session.State.UnlockedLocations = append(session.State.UnlockedLocations, unlockID)
					}
				}

				// 更新场景状态
				if scene.State == nil {
					scene.State = make(map[string]any)
				}
				scene.State["clue_"+clue.ID+"_collected"] = true
				result.StateChanges["clue_"+clue.ID+"_collected"] = true

				// 保存场景状态
				_ = s.SaveSceneState(sessionID, scene.ID, scene.State)

				// 保存会话
				_ = s.gameService.SaveSession(session)

				return result, nil
			} else {
				return nil, domain.NewGameError(domain.ErrInvalidAction, "线索需求未满足").
					WithDetails("clue_id", clue.ID)
			}
		}
	}

	// 检查是否是NPC交互
	for _, npc := range scene.NPCs {
		if npc.ID == objectID {
			result.Success = true
			result.Description = "与" + npc.Name + "交互"

			// 更新场景状态
			if scene.State == nil {
				scene.State = make(map[string]any)
			}
			scene.State["npc_"+npc.ID+"_interacted"] = true
			result.StateChanges["npc_"+npc.ID+"_interacted"] = true

			// 保存场景状态
			_ = s.SaveSceneState(sessionID, scene.ID, scene.State)

			return result, nil
		}
	}

	return nil, domain.NewGameError(domain.ErrNotFound, "对象不存在").
		WithDetails("object_id", objectID)
}

// GetAvailableInteractions 获取可用交互
func (s *sceneService) GetAvailableInteractions(sessionID string) ([]*Interaction, error) {
	// 获取当前场景
	scene, err := s.GetCurrentScene(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	interactions := make([]*Interaction, 0)

	// 添加线索交互
	for _, clue := range scene.Clues {
		// 检查线索是否已收集
		if contains(session.State.CollectedClues, clue.ID) {
			continue
		}

		// 检查线索需求
		if s.scenarioService.CheckClueRequirements(clue, session.State) {
			interactions = append(interactions, &Interaction{
				ObjectID:    clue.ID,
				ObjectName:  clue.Name,
				Actions:     []string{"调查", "检查"},
				Description: clue.Description,
			})
		}
	}

	// 添加NPC交互
	for _, npc := range scene.NPCs {
		interactions = append(interactions, &Interaction{
			ObjectID:    npc.ID,
			ObjectName:  npc.Name,
			Actions:     []string{"对话", "观察"},
			Description: npc.Description,
		})
	}

	return interactions, nil
}

// PersistSceneStates 持久化场景状态
func (s *sceneService) PersistSceneStates(sessionID string) error {
	s.mu.RLock()
	sessionStates, exists := s.sceneStates[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil // 没有状态需要持久化
	}

	// 将场景状态序列化并保存到会话的自定义数据中
	// 这里我们使用一个简单的方法：将状态保存到会话的State中
	// 实际实现中可能需要更复杂的持久化机制

	// 序列化场景状态
	statesJSON, err := json.Marshal(sessionStates)
	if err != nil {
		return domain.NewGameError(domain.ErrInternal, "序列化场景状态失败").
			WithDetails("error", err.Error())
	}

	// 这里我们需要一个地方存储这些数据
	// 由于当前的GameState没有专门的字段，我们可以考虑扩展它
	// 或者使用其他持久化机制
	_ = statesJSON // 暂时忽略，实际实现中需要保存

	return nil
}

// LoadPersistedSceneStates 加载持久化的场景状态
func (s *sceneService) LoadPersistedSceneStates(sessionID string) (map[string]map[string]any, error) {
	// 从持久化存储加载场景状态
	// 这里返回当前内存中的状态
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sessionStates, exists := s.sceneStates[sessionID]; exists {
		// 创建深拷贝
		result := make(map[string]map[string]any)
		for sceneID, state := range sessionStates {
			stateCopy := make(map[string]any)
			for k, v := range state {
				stateCopy[k] = v
			}
			result[sceneID] = stateCopy
		}
		return result, nil
	}

	return make(map[string]map[string]any), nil
}

// contains 辅助函数：检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
