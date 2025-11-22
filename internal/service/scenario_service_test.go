package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

func TestScenarioService_LoadScenario(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	service := NewScenarioService(tmpDir)

	// 创建测试剧本
	scenario := CreateTestScenario()

	// 保存剧本
	err := service.(*scenarioService).SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 加载剧本
	loaded, err := service.LoadScenario(scenario.ID)
	if err != nil {
		t.Fatalf("加载剧本失败: %v", err)
	}

	// 验证基本信息
	if loaded.ID != scenario.ID {
		t.Errorf("期望ID为 %s, 得到 %s", scenario.ID, loaded.ID)
	}

	if loaded.Name != scenario.Name {
		t.Errorf("期望名称为 %s, 得到 %s", scenario.Name, loaded.Name)
	}

	if loaded.Description != scenario.Description {
		t.Errorf("期望描述为 %s, 得到 %s", scenario.Description, loaded.Description)
	}

	// 验证异常体档案
	if loaded.Anomaly == nil {
		t.Fatal("期望异常体档案不为nil")
	}

	if loaded.Anomaly.ID != scenario.Anomaly.ID {
		t.Errorf("期望异常体ID为 %s, 得到 %s", scenario.Anomaly.ID, loaded.Anomaly.ID)
	}

	// 验证场景
	if len(loaded.Scenes) != len(scenario.Scenes) {
		t.Errorf("期望 %d 个场景, 得到 %d", len(scenario.Scenes), len(loaded.Scenes))
	}

	// 测试加载不存在的剧本
	_, err = service.LoadScenario("不存在的剧本")
	if err == nil {
		t.Error("期望加载不存在的剧本导致错误")
	}
}

func TestScenarioService_LoadScenario_Cache(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建并保存测试剧本
	scenario := CreateTestScenario()
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 第一次加载
	loaded1, err := service.LoadScenario(scenario.ID)
	if err != nil {
		t.Fatalf("第一次加载失败: %v", err)
	}

	// 验证缓存
	cached, exists := service.GetScenarioFromCache(scenario.ID)
	if !exists {
		t.Error("期望剧本在缓存中")
	}

	if cached.ID != loaded1.ID {
		t.Error("缓存的剧本与加载的剧本不一致")
	}

	// 第二次加载（应该从缓存获取）
	loaded2, err := service.LoadScenario(scenario.ID)
	if err != nil {
		t.Fatalf("第二次加载失败: %v", err)
	}

	if loaded2.ID != loaded1.ID {
		t.Error("两次加载的剧本不一致")
	}
}

func TestScenarioService_ListScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建多个测试剧本
	scenarios := []*domain.Scenario{
		CreateTestScenario(),
		{
			ID:          "scenario-2",
			Name:        "剧本2",
			Description: "第二个测试剧本",
			Anomaly: &domain.AnomalyProfile{
				ID:   "anomaly-2",
				Name: "异常体2",
			},
			Scenes: map[string]*domain.Scene{
				"scene-1": {
					ID:          "scene-1",
					Name:        "场景1",
					Description: "测试场景",
					Connections: []string{},
					State:       make(map[string]interface{}),
				},
			},
			StartingSceneID: "scene-1",
		},
	}

	// 保存所有剧本
	for _, scenario := range scenarios {
		err := service.SaveScenario(scenario)
		if err != nil {
			t.Fatalf("保存剧本失败: %v", err)
		}
	}

	// 列出剧本
	summaries, err := service.ListScenarios()
	if err != nil {
		t.Fatalf("列出剧本失败: %v", err)
	}

	if len(summaries) != len(scenarios) {
		t.Errorf("期望 %d 个剧本, 得到 %d", len(scenarios), len(summaries))
	}

	// 验证摘要信息
	for _, summary := range summaries {
		found := false
		for _, scenario := range scenarios {
			if summary.ID == scenario.ID {
				found = true
				if summary.Name != scenario.Name {
					t.Errorf("期望名称为 %s, 得到 %s", scenario.Name, summary.Name)
				}
				if summary.Description != scenario.Description {
					t.Errorf("期望描述为 %s, 得到 %s", scenario.Description, summary.Description)
				}
				break
			}
		}
		if !found {
			t.Errorf("未找到剧本 %s", summary.ID)
		}
	}
}

func TestScenarioService_ListScenarios_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir)

	// 列出空目录
	summaries, err := service.ListScenarios()
	if err != nil {
		t.Fatalf("列出剧本失败: %v", err)
	}

	if len(summaries) != 0 {
		t.Errorf("期望0个剧本, 得到 %d", len(summaries))
	}
}

func TestScenarioService_ValidateScenario(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir)

	tests := []struct {
		name      string
		scenario  *domain.Scenario
		wantError bool
	}{
		{
			name:      "有效剧本",
			scenario:  CreateTestScenario(),
			wantError: false,
		},
		{
			name: "缺少ID",
			scenario: &domain.Scenario{
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
		{
			name: "缺少名称",
			scenario: &domain.Scenario{
				ID:          "test",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
		{
			name: "缺少异常体档案",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
		{
			name: "缺少场景",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes:          map[string]*domain.Scene{},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
		{
			name: "起始场景不存在",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "不存在的场景",
			},
			wantError: true,
		},
		{
			name: "场景连接无效",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景",
						Connections: []string{"不存在的场景"},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
		{
			name: "线索解锁场景不存在",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:   "scene-1",
						Name: "场景",
						Clues: []*domain.Clue{
							{
								ID:      "clue-1",
								Name:    "线索",
								Unlocks: []string{"不存在的场景"},
							},
						},
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateScenario(tt.scenario)

			if tt.wantError && err == nil {
				t.Error("期望验证失败，但没有错误")
			}

			if !tt.wantError && err != nil {
				t.Errorf("期望验证成功，但得到错误: %v", err)
			}
		})
	}
}

func TestScenarioService_GetScene(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建并保存测试剧本
	scenario := CreateTestScenario()
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 获取场景
	scene, err := service.GetScene(scenario.ID, "scene-1")
	if err != nil {
		t.Fatalf("获取场景失败: %v", err)
	}

	// 验证场景
	if scene.ID != "scene-1" {
		t.Errorf("期望场景ID为 scene-1, 得到 %s", scene.ID)
	}

	if scene.Name != "工厂入口" {
		t.Errorf("期望场景名称为 工厂入口, 得到 %s", scene.Name)
	}

	// 测试获取不存在的场景
	_, err = service.GetScene(scenario.ID, "不存在的场景")
	if err == nil {
		t.Error("期望获取不存在的场景导致错误")
	}

	// 测试获取不存在剧本的场景
	_, err = service.GetScene("不存在的剧本", "scene-1")
	if err == nil {
		t.Error("期望获取不存在剧本的场景导致错误")
	}
}

func TestScenarioService_GetClue(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建并保存测试剧本
	scenario := CreateTestScenario()
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 获取线索
	clue, err := service.GetClue(scenario.ID, "clue-1")
	if err != nil {
		t.Fatalf("获取线索失败: %v", err)
	}

	// 验证线索
	if clue.ID != "clue-1" {
		t.Errorf("期望线索ID为 clue-1, 得到 %s", clue.ID)
	}

	if clue.Name != "脚印" {
		t.Errorf("期望线索名称为 脚印, 得到 %s", clue.Name)
	}

	// 测试获取不存在的线索
	_, err = service.GetClue(scenario.ID, "不存在的线索")
	if err == nil {
		t.Error("期望获取不存在的线索导致错误")
	}

	// 测试获取不存在剧本的线索
	_, err = service.GetClue("不存在的剧本", "clue-1")
	if err == nil {
		t.Error("期望获取不存在剧本的线索导致错误")
	}
}

func TestScenarioService_CheckClueRequirements(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir)

	tests := []struct {
		name     string
		clue     *domain.Clue
		state    *domain.GameState
		expected bool
	}{
		{
			name: "无需求的线索",
			clue: &domain.Clue{
				ID:           "clue-1",
				Requirements: []string{},
			},
			state: &domain.GameState{
				CollectedClues: []string{},
			},
			expected: true,
		},
		{
			name: "需求已满足",
			clue: &domain.Clue{
				ID:           "clue-2",
				Requirements: []string{"clue-1"},
			},
			state: &domain.GameState{
				CollectedClues: []string{"clue-1"},
			},
			expected: true,
		},
		{
			name: "需求未满足",
			clue: &domain.Clue{
				ID:           "clue-2",
				Requirements: []string{"clue-1"},
			},
			state: &domain.GameState{
				CollectedClues: []string{},
			},
			expected: false,
		},
		{
			name: "多个需求全部满足",
			clue: &domain.Clue{
				ID:           "clue-3",
				Requirements: []string{"clue-1", "clue-2"},
			},
			state: &domain.GameState{
				CollectedClues: []string{"clue-1", "clue-2"},
			},
			expected: true,
		},
		{
			name: "多个需求部分满足",
			clue: &domain.Clue{
				ID:           "clue-3",
				Requirements: []string{"clue-1", "clue-2"},
			},
			state: &domain.GameState{
				CollectedClues: []string{"clue-1"},
			},
			expected: false,
		},
		{
			name:     "nil线索",
			clue:     nil,
			state:    &domain.GameState{},
			expected: false,
		},
		{
			name: "nil状态",
			clue: &domain.Clue{
				ID: "clue-1",
			},
			state:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckClueRequirements(tt.clue, tt.state)
			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

func TestScenarioService_CheckEventTriggers(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建测试剧本
	scenario := CreateTestScenario()
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	tests := []struct {
		name          string
		state         *domain.GameState
		expectedCount int
	}{
		{
			name: "第一次访问触发",
			state: &domain.GameState{
				CurrentSceneID: "scene-1",
				VisitedScenes:  map[string]bool{},
			},
			expectedCount: 1,
		},
		{
			name: "非第一次访问不触发",
			state: &domain.GameState{
				CurrentSceneID: "scene-1",
				VisitedScenes: map[string]bool{
					"scene-1": true,
				},
			},
			expectedCount: 0,
		},
		{
			name: "当前场景不存在",
			state: &domain.GameState{
				CurrentSceneID: "不存在的场景",
				VisitedScenes:  map[string]bool{},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := service.CheckEventTriggers(scenario, tt.state)
			if err != nil {
				t.Fatalf("检查事件触发失败: %v", err)
			}

			if len(events) != tt.expectedCount {
				t.Errorf("期望 %d 个事件, 得到 %d", tt.expectedCount, len(events))
			}
		})
	}

	// 测试nil参数
	_, err = service.CheckEventTriggers(nil, &domain.GameState{})
	if err == nil {
		t.Error("期望nil剧本导致错误")
	}

	_, err = service.CheckEventTriggers(scenario, nil)
	if err == nil {
		t.Error("期望nil状态导致错误")
	}
}

func TestScenarioService_EvaluateTrigger(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	tests := []struct {
		name     string
		trigger  string
		state    *domain.GameState
		expected bool
	}{
		{
			name:     "always触发器",
			trigger:  "always",
			state:    &domain.GameState{},
			expected: true,
		},
		{
			name:    "domain_unlocked触发器-已解锁",
			trigger: "domain_unlocked",
			state: &domain.GameState{
				DomainUnlocked: true,
			},
			expected: true,
		},
		{
			name:    "domain_unlocked触发器-未解锁",
			trigger: "domain_unlocked",
			state: &domain.GameState{
				DomainUnlocked: false,
			},
			expected: false,
		},
		{
			name:    "first_visit触发器-第一次",
			trigger: "first_visit",
			state: &domain.GameState{
				CurrentSceneID: "scene-1",
				VisitedScenes:  map[string]bool{},
			},
			expected: true,
		},
		{
			name:    "first_visit触发器-非第一次",
			trigger: "first_visit",
			state: &domain.GameState{
				CurrentSceneID: "scene-1",
				VisitedScenes: map[string]bool{
					"scene-1": true,
				},
			},
			expected: false,
		},
		{
			name:    "线索触发器-已收集",
			trigger: "clue:clue-1",
			state: &domain.GameState{
				CollectedClues: []string{"clue-1"},
			},
			expected: true,
		},
		{
			name:    "线索触发器-未收集",
			trigger: "clue:clue-1",
			state: &domain.GameState{
				CollectedClues: []string{},
			},
			expected: false,
		},
		{
			name:     "未知触发器",
			trigger:  "unknown",
			state:    &domain.GameState{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.evaluateTrigger(tt.trigger, tt.state)
			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

func TestScenarioService_ValidateSceneConnections(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	tests := []struct {
		name      string
		scenario  *domain.Scenario
		wantError bool
	}{
		{
			name:      "所有场景可达",
			scenario:  CreateTestScenario(),
			wantError: false,
		},
		{
			name: "存在孤立场景",
			scenario: &domain.Scenario{
				ID:          "test",
				Name:        "测试",
				Description: "测试",
				Anomaly: &domain.AnomalyProfile{
					ID:   "anomaly-1",
					Name: "异常体",
				},
				Scenes: map[string]*domain.Scene{
					"scene-1": {
						ID:          "scene-1",
						Name:        "场景1",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
					"scene-2": {
						ID:          "scene-2",
						Name:        "场景2",
						Connections: []string{},
						State:       make(map[string]interface{}),
					},
				},
				StartingSceneID: "scene-1",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSceneConnections(tt.scenario)

			if tt.wantError && err == nil {
				t.Error("期望验证失败，但没有错误")
			}

			if !tt.wantError && err != nil {
				t.Errorf("期望验证成功，但得到错误: %v", err)
			}
		})
	}
}

func TestScenarioService_SaveScenario(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建测试剧本
	scenario := CreateTestScenario()

	// 保存剧本
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 验证文件存在
	filePath := filepath.Join(tmpDir, scenario.ID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("期望文件存在")
	}

	// 验证可以加载
	loaded, err := service.LoadScenario(scenario.ID)
	if err != nil {
		t.Fatalf("加载剧本失败: %v", err)
	}

	if loaded.ID != scenario.ID {
		t.Errorf("期望ID为 %s, 得到 %s", scenario.ID, loaded.ID)
	}

	// 测试保存nil剧本
	err = service.SaveScenario(nil)
	if err == nil {
		t.Error("期望保存nil剧本导致错误")
	}

	// 测试保存无效剧本
	invalidScenario := &domain.Scenario{
		ID: "invalid",
		// 缺少必需字段
	}
	err = service.SaveScenario(invalidScenario)
	if err == nil {
		t.Error("期望保存无效剧本导致错误")
	}
}

func TestScenarioService_ClearCache(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir).(*scenarioService)

	// 创建并保存测试剧本
	scenario := CreateTestScenario()
	err := service.SaveScenario(scenario)
	if err != nil {
		t.Fatalf("保存剧本失败: %v", err)
	}

	// 加载剧本（会缓存）
	_, err = service.LoadScenario(scenario.ID)
	if err != nil {
		t.Fatalf("加载剧本失败: %v", err)
	}

	// 验证缓存存在
	_, exists := service.GetScenarioFromCache(scenario.ID)
	if !exists {
		t.Error("期望剧本在缓存中")
	}

	// 清除缓存
	service.ClearCache()

	// 验证缓存已清除
	_, exists = service.GetScenarioFromCache(scenario.ID)
	if exists {
		t.Error("期望缓存已清除")
	}
}

func TestScenarioService_LoadScenario_CorruptedData(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewScenarioService(tmpDir)

	// 创建损坏的JSON文件
	filePath := filepath.Join(tmpDir, "corrupted.json")
	err := os.WriteFile(filePath, []byte("这不是有效的JSON"), 0644)
	if err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 尝试加载损坏的剧本
	_, err = service.LoadScenario("corrupted")
	if err == nil {
		t.Error("期望加载损坏的剧本导致错误")
	}
}
