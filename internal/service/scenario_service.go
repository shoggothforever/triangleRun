package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// ScenarioService 剧本服务接口
type ScenarioService interface {
	// 剧本管理
	LoadScenario(scenarioID string) (*domain.Scenario, error)
	ListScenarios() ([]*ScenarioSummary, error)
	ValidateScenario(scenario *domain.Scenario) error

	// 场景导航
	GetScene(scenarioID, sceneID string) (*domain.Scene, error)
	GetAvailableScenes(sessionID string, state *domain.GameState) ([]*domain.Scene, error)

	// 线索系统
	GetClue(scenarioID, clueID string) (*domain.Clue, error)
	CheckClueRequirements(clue *domain.Clue, state *domain.GameState) bool

	// 事件系统
	CheckEventTriggers(scenario *domain.Scenario, state *domain.GameState) ([]*domain.Event, error)
}

// ScenarioSummary 剧本摘要
type ScenarioSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// scenarioService 剧本服务实现
type scenarioService struct {
	scenarios    map[string]*domain.Scenario
	scenariosDir string
	mu           sync.RWMutex
}

// NewScenarioService 创建剧本服务
func NewScenarioService(scenariosDir string) ScenarioService {
	return &scenarioService{
		scenarios:    make(map[string]*domain.Scenario),
		scenariosDir: scenariosDir,
	}
}

// LoadScenario 加载剧本
func (s *scenarioService) LoadScenario(scenarioID string) (*domain.Scenario, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查缓存
	if scenario, exists := s.scenarios[scenarioID]; exists {
		return scenario, nil
	}

	// 构建文件路径
	filePath := filepath.Join(s.scenariosDir, scenarioID+".json")

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.NewGameError(domain.ErrNotFound, "剧本不存在").
				WithDetails("scenario_id", scenarioID)
		}
		return nil, domain.NewGameError(domain.ErrInternal, "读取剧本文件失败").
			WithDetails("error", err.Error())
	}

	// 解析JSON
	var scenario domain.Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "剧本数据格式错误").
			WithDetails("scenario_id", scenarioID).
			WithDetails("error", err.Error())
	}

	// 验证剧本
	if err := s.ValidateScenario(&scenario); err != nil {
		return nil, err
	}

	// 缓存剧本
	s.scenarios[scenarioID] = &scenario

	return &scenario, nil
}

// ListScenarios 列出所有剧本
func (s *scenarioService) ListScenarios() ([]*ScenarioSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 读取剧本目录
	entries, err := os.ReadDir(s.scenariosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ScenarioSummary{}, nil
		}
		return nil, domain.NewGameError(domain.ErrInternal, "读取剧本目录失败").
			WithDetails("error", err.Error())
	}

	summaries := make([]*ScenarioSummary, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		scenarioID := entry.Name()[:len(entry.Name())-5] // 移除 .json 扩展名

		// 尝试从缓存获取
		if scenario, exists := s.scenarios[scenarioID]; exists {
			summaries = append(summaries, &ScenarioSummary{
				ID:          scenario.ID,
				Name:        scenario.Name,
				Description: scenario.Description,
			})
			continue
		}

		// 读取文件获取基本信息
		filePath := filepath.Join(s.scenariosDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var scenario domain.Scenario
		if err := json.Unmarshal(data, &scenario); err != nil {
			continue
		}

		summaries = append(summaries, &ScenarioSummary{
			ID:          scenario.ID,
			Name:        scenario.Name,
			Description: scenario.Description,
		})
	}

	return summaries, nil
}

// ValidateScenario 验证剧本
func (s *scenarioService) ValidateScenario(scenario *domain.Scenario) error {
	// 验证基本字段
	if scenario.ID == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本ID不能为空")
	}
	if scenario.Name == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本名称不能为空")
	}

	// 验证异常体档案
	if scenario.Anomaly == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本必须包含异常体档案")
	}
	if scenario.Anomaly.ID == "" || scenario.Anomaly.Name == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "异常体档案信息不完整")
	}

	// 验证场景
	if len(scenario.Scenes) == 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本必须包含至少一个场景")
	}

	// 验证起始场景
	if scenario.StartingSceneID == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本必须指定起始场景")
	}
	if _, exists := scenario.Scenes[scenario.StartingSceneID]; !exists {
		return domain.NewGameError(domain.ErrInvalidInput, "起始场景不存在").
			WithDetails("starting_scene_id", scenario.StartingSceneID)
	}

	// 验证场景连接
	for sceneID, scene := range scenario.Scenes {
		for _, connID := range scene.Connections {
			if _, exists := scenario.Scenes[connID]; !exists {
				return domain.NewGameError(domain.ErrInvalidInput, "场景连接指向不存在的场景").
					WithDetails("scene_id", sceneID).
					WithDetails("connection_id", connID)
			}
		}
	}

	// 验证线索引用
	for sceneID, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			// 验证线索解锁的场景是否存在
			for _, unlockID := range clue.Unlocks {
				if _, exists := scenario.Scenes[unlockID]; !exists {
					return domain.NewGameError(domain.ErrInvalidInput, "线索解锁的场景不存在").
						WithDetails("scene_id", sceneID).
						WithDetails("clue_id", clue.ID).
						WithDetails("unlock_id", unlockID)
				}
			}
		}
	}

	return nil
}

// GetScene 获取场景
func (s *scenarioService) GetScene(scenarioID, sceneID string) (*domain.Scene, error) {
	// 加载剧本
	scenario, err := s.LoadScenario(scenarioID)
	if err != nil {
		return nil, err
	}

	// 获取场景
	scene, exists := scenario.Scenes[sceneID]
	if !exists {
		return nil, domain.NewGameError(domain.ErrNotFound, "场景不存在").
			WithDetails("scenario_id", scenarioID).
			WithDetails("scene_id", sceneID)
	}

	return scene, nil
}

// GetAvailableScenes 获取可用场景
func (s *scenarioService) GetAvailableScenes(sessionID string, state *domain.GameState) ([]*domain.Scene, error) {
	// 这里需要从state中获取scenarioID，但当前GameState没有这个字段
	// 为了实现，我们假设可以从其他地方获取，或者修改接口
	// 暂时返回空列表
	return []*domain.Scene{}, nil
}

// GetClue 获取线索
func (s *scenarioService) GetClue(scenarioID, clueID string) (*domain.Clue, error) {
	// 加载剧本
	scenario, err := s.LoadScenario(scenarioID)
	if err != nil {
		return nil, err
	}

	// 在所有场景中查找线索
	for _, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			if clue.ID == clueID {
				return clue, nil
			}
		}
	}

	return nil, domain.NewGameError(domain.ErrNotFound, "线索不存在").
		WithDetails("scenario_id", scenarioID).
		WithDetails("clue_id", clueID)
}

// CheckClueRequirements 检查线索需求
func (s *scenarioService) CheckClueRequirements(clue *domain.Clue, state *domain.GameState) bool {
	if clue == nil || state == nil {
		return false
	}

	// 如果没有需求，直接可用
	if len(clue.Requirements) == 0 {
		return true
	}

	// 检查所有需求是否满足
	for _, req := range clue.Requirements {
		found := false
		for _, collected := range state.CollectedClues {
			if collected == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// CheckEventTriggers 检查事件触发
func (s *scenarioService) CheckEventTriggers(scenario *domain.Scenario, state *domain.GameState) ([]*domain.Event, error) {
	if scenario == nil || state == nil {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "参数不能为空")
	}

	triggeredEvents := make([]*domain.Event, 0)

	// 获取当前场景
	currentScene, exists := scenario.Scenes[state.CurrentSceneID]
	if !exists {
		return triggeredEvents, nil
	}

	// 检查场景中的所有事件
	for _, event := range currentScene.Events {
		if s.evaluateTrigger(event.Trigger, state) {
			triggeredEvents = append(triggeredEvents, event)
		}
	}

	return triggeredEvents, nil
}

// evaluateTrigger 评估触发条件
func (s *scenarioService) evaluateTrigger(trigger string, state *domain.GameState) bool {
	// 简单的触发条件评估
	// 实际实现中可能需要更复杂的表达式解析
	switch trigger {
	case "always":
		return true
	case "domain_unlocked":
		return state.DomainUnlocked
	case "first_visit":
		return !state.VisitedScenes[state.CurrentSceneID]
	default:
		// 检查是否是线索触发
		if len(trigger) > 5 && trigger[:5] == "clue:" {
			clueID := trigger[5:]
			for _, collected := range state.CollectedClues {
				if collected == clueID {
					return true
				}
			}
		}
		return false
	}
}

// GetScenarioByID 通过ID获取剧本（内部辅助方法）
func (s *scenarioService) GetScenarioByID(scenarioID string) (*domain.Scenario, error) {
	return s.LoadScenario(scenarioID)
}

// ClearCache 清除缓存
func (s *scenarioService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scenarios = make(map[string]*domain.Scenario)
}

// GetScenarioFromCache 从缓存获取剧本
func (s *scenarioService) GetScenarioFromCache(scenarioID string) (*domain.Scenario, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scenario, exists := s.scenarios[scenarioID]
	return scenario, exists
}

// CreateTestScenario 创建测试剧本（用于测试）
func CreateTestScenario() *domain.Scenario {
	return &domain.Scenario{
		ID:          "test-scenario",
		Name:        "测试剧本",
		Description: "这是一个测试剧本",
		Anomaly: &domain.AnomalyProfile{
			ID:            "test-anomaly",
			Name:          "测试异常体",
			History:       "测试历史",
			Focus:         &domain.Focus{Emotion: "恐惧", Subject: "黑暗"},
			Domain:        &domain.Domain{Location: "废弃工厂", Description: "阴暗潮湿"},
			Appearance:    "黑色影子",
			Impulse:       "吞噬光明",
			CurrentStatus: "活跃",
			ChaosEffects: []*domain.ChaosEffect{
				{
					ID:          "effect-1",
					Name:        "黑暗降临",
					Cost:        3,
					Description: "周围陷入黑暗",
					Effect:      "所有掷骰-1",
				},
			},
		},
		MorningScenes: []*domain.MorningScene{
			{
				ID:          "morning-1",
				Description: "晨会场景",
				Type:        "briefing",
			},
		},
		Briefing: &domain.Briefing{
			Summary:    "调查废弃工厂的异常现象",
			Objectives: []string{"找到异常体", "捕获或中和"},
			Warnings:   []string{"注意安全", "避免直视"},
		},
		OptionalGoals: []*domain.OptionalGoal{
			{
				ID:          "goal-1",
				Description: "救出被困工人",
				Reward:      3,
			},
		},
		Scenes: map[string]*domain.Scene{
			"scene-1": {
				ID:          "scene-1",
				Name:        "工厂入口",
				Description: "锈迹斑斑的大门",
				NPCs: []*domain.NPC{
					{
						ID:          "npc-1",
						Name:        "工厂工人",
						Description: "一个疲惫的工厂工人",
						Personality: "友好但紧张",
						Dialogues:   []string{"这里很危险", "小心点"},
						State:       "normal",
					},
					{
						ID:          "npc-2",
						Name:        "保安",
						Description: "警惕的保安",
						Personality: "严肃",
						Dialogues:   []string{"请出示证件", "这里禁止进入"},
						State:       "normal",
					},
				},
				Clues: []*domain.Clue{
					{
						ID:           "clue-1",
						Name:         "脚印",
						Description:  "地上有奇怪的脚印",
						Requirements: []string{},
						Unlocks:      []string{"scene-2"},
					},
				},
				Events: []*domain.Event{
					{
						ID:          "event-1",
						Name:        "第一次访问",
						Description: "你第一次来到这里",
						Trigger:     "first_visit",
						Effect:      "获得线索",
					},
				},
				Connections: []string{"scene-2"},
				State:       make(map[string]interface{}),
			},
			"scene-2": {
				ID:          "scene-2",
				Name:        "工厂车间",
				Description: "空旷的车间",
				NPCs:        []*domain.NPC{},
				Clues: []*domain.Clue{
					{
						ID:           "clue-2",
						Name:         "血迹",
						Description:  "墙上有血迹",
						Requirements: []string{"clue-1"},
						Unlocks:      []string{},
					},
				},
				Events:      []*domain.Event{},
				Connections: []string{"scene-1"},
				State:       make(map[string]interface{}),
			},
		},
		StartingSceneID: "scene-1",
		Encounter: &domain.Encounter{
			ID:          "encounter-1",
			Description: "与异常体的最终对决",
			Phases: []*domain.Phase{
				{
					ID:          "phase-1",
					Description: "接近异常体",
					Actions:     []string{"观察", "攻击", "逃跑"},
				},
			},
		},
		Aftermath: &domain.Aftermath{
			Captured:    "成功捕获异常体",
			Neutralized: "成功中和异常体",
			Escaped:     "异常体逃脱了",
		},
		Rewards: &domain.Rewards{
			Commendations: 3,
			Claimables:    []string{"波纹枪", "防护服"},
		},
	}
}

// SaveScenario 保存剧本到文件（用于测试）
func (s *scenarioService) SaveScenario(scenario *domain.Scenario) error {
	if scenario == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "剧本不能为空")
	}

	// 验证剧本
	if err := s.ValidateScenario(scenario); err != nil {
		return err
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return domain.NewGameError(domain.ErrInternal, "序列化剧本失败").
			WithDetails("error", err.Error())
	}

	// 构建文件路径
	filePath := filepath.Join(s.scenariosDir, scenario.ID+".json")

	// 确保目录存在
	if err := os.MkdirAll(s.scenariosDir, 0755); err != nil {
		return domain.NewGameError(domain.ErrInternal, "创建目录失败").
			WithDetails("error", err.Error())
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return domain.NewGameError(domain.ErrInternal, "写入文件失败").
			WithDetails("error", err.Error())
	}

	// 更新缓存
	s.mu.Lock()
	s.scenarios[scenario.ID] = scenario
	s.mu.Unlock()

	return nil
}

// ValidateSceneConnections 验证场景连接的完整性
func (s *scenarioService) ValidateSceneConnections(scenario *domain.Scenario) error {
	if scenario == nil {
		return fmt.Errorf("scenario cannot be nil")
	}

	visited := make(map[string]bool)
	queue := []string{scenario.StartingSceneID}

	// BFS遍历所有可达场景
	for len(queue) > 0 {
		sceneID := queue[0]
		queue = queue[1:]

		if visited[sceneID] {
			continue
		}
		visited[sceneID] = true

		scene, exists := scenario.Scenes[sceneID]
		if !exists {
			continue
		}

		for _, connID := range scene.Connections {
			if !visited[connID] {
				queue = append(queue, connID)
			}
		}
	}

	// 检查是否有孤立场景
	for sceneID := range scenario.Scenes {
		if !visited[sceneID] {
			return domain.NewGameError(domain.ErrInvalidInput, "存在无法到达的场景").
				WithDetails("scene_id", sceneID)
		}
	}

	return nil
}
