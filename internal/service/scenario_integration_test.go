package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// TestScenarioService_LoadEternalSpring 测试通过服务加载永恒之泉剧本
func TestScenarioService_LoadEternalSpring(t *testing.T) {
	// 创建剧本服务
	service := NewScenarioService("../../scenarios")

	// 加载永恒之泉剧本
	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err, "应该能够加载永恒之泉剧本")
	require.NotNil(t, scenario, "剧本不应为空")

	// 验证基本信息
	assert.Equal(t, "eternal-spring", scenario.ID)
	assert.Equal(t, "永恒之泉", scenario.Name)
	assert.NotEmpty(t, scenario.Description)

	// 验证异常体
	require.NotNil(t, scenario.Anomaly)
	assert.Equal(t, "永恒之泉", scenario.Anomaly.Name)
	assert.Equal(t, "渴望", scenario.Anomaly.Focus.Emotion)

	// 验证场景
	assert.NotEmpty(t, scenario.Scenes)
	assert.Equal(t, "commercial-avenue", scenario.StartingSceneID)

	// 验证起始场景存在
	startScene, exists := scenario.Scenes[scenario.StartingSceneID]
	assert.True(t, exists, "起始场景应该存在")
	assert.Equal(t, "商业大道", startScene.Name)

	t.Logf("成功加载剧本: %s，包含 %d 个场景", scenario.Name, len(scenario.Scenes))
}

// TestScenarioService_ValidateEternalSpring 测试验证永恒之泉剧本
func TestScenarioService_ValidateEternalSpring(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	// 加载剧本
	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err)

	// 验证剧本
	err = service.ValidateScenario(scenario)
	assert.NoError(t, err, "永恒之泉剧本应该通过验证")
}

// TestScenarioService_GetEternalSpringScene 测试获取永恒之泉的场景
func TestScenarioService_GetEternalSpringScene(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	// 获取商业大道场景
	scene, err := service.GetScene("eternal-spring", "commercial-avenue")
	require.NoError(t, err, "应该能够获取商业大道场景")
	require.NotNil(t, scene)

	assert.Equal(t, "commercial-avenue", scene.ID)
	assert.Equal(t, "商业大道", scene.Name)
	assert.NotEmpty(t, scene.Description)
	assert.NotEmpty(t, scene.NPCs, "商业大道应该有NPC")
	assert.NotEmpty(t, scene.Clues, "商业大道应该有线索")

	t.Logf("场景: %s，包含 %d 个NPC，%d 条线索", scene.Name, len(scene.NPCs), len(scene.Clues))
}

// TestScenarioService_ListScenariosIncludesEternalSpring 测试列表包含永恒之泉
func TestScenarioService_ListScenariosIncludesEternalSpring(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	// 列出所有剧本
	scenarios, err := service.ListScenarios()
	require.NoError(t, err)
	require.NotEmpty(t, scenarios, "应该至少有一个剧本")

	// 查找永恒之泉
	found := false
	for _, summary := range scenarios {
		if summary.ID == "eternal-spring" {
			found = true
			assert.Equal(t, "永恒之泉", summary.Name)
			assert.NotEmpty(t, summary.Description)
			break
		}
	}

	assert.True(t, found, "剧本列表应该包含永恒之泉")
	t.Logf("找到 %d 个剧本", len(scenarios))
}

// TestScenarioService_EternalSpringSceneConnections 测试永恒之泉的场景连接
func TestScenarioService_EternalSpringSceneConnections(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err)

	// 验证所有场景连接
	for sceneID, scene := range scenario.Scenes {
		for _, connectionID := range scene.Connections {
			connectedScene, exists := scenario.Scenes[connectionID]
			assert.True(t, exists, "场景 %s 的连接 %s 应该存在", sceneID, connectionID)
			if exists {
				t.Logf("%s -> %s", scene.Name, connectedScene.Name)
			}
		}
	}
}

// TestScenarioService_EternalSpringNPCs 测试永恒之泉的NPC
func TestScenarioService_EternalSpringNPCs(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err)

	// 收集所有NPC
	npcs := make(map[string]*domain.NPC)
	for _, scene := range scenario.Scenes {
		for _, npc := range scene.NPCs {
			npcs[npc.ID] = npc
		}
	}

	// 验证关键NPC存在
	keyNPCs := []string{"maya-ng", "serena-evermore"}
	for _, npcID := range keyNPCs {
		npc, exists := npcs[npcID]
		assert.True(t, exists, "关键NPC %s 应该存在", npcID)
		if exists {
			assert.NotEmpty(t, npc.Name)
			assert.NotEmpty(t, npc.Description)
			assert.NotEmpty(t, npc.Dialogues, "NPC %s 应该有对话", npc.Name)
			t.Logf("NPC: %s，%d 条对话", npc.Name, len(npc.Dialogues))
		}
	}
}

// TestScenarioService_EternalSpringClues 测试永恒之泉的线索系统
func TestScenarioService_EternalSpringClues(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err)

	// 收集所有线索
	clues := make(map[string]*domain.Clue)
	for _, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			clues[clue.ID] = clue
		}
	}

	t.Logf("剧本包含 %d 条线索", len(clues))

	// 验证线索有解锁链
	hasUnlocks := 0
	for _, clue := range clues {
		if len(clue.Unlocks) > 0 {
			hasUnlocks++
			t.Logf("线索 %s 解锁: %v", clue.Name, clue.Unlocks)
		}
	}

	assert.Greater(t, hasUnlocks, 0, "应该有线索能够解锁新内容")
}

// TestScenarioService_EternalSpringChaosEffects 测试永恒之泉的混沌效应
func TestScenarioService_EternalSpringChaosEffects(t *testing.T) {
	service := NewScenarioService("../../scenarios")

	scenario, err := service.LoadScenario("eternal-spring")
	require.NoError(t, err)

	effects := scenario.Anomaly.ChaosEffects
	require.NotEmpty(t, effects, "应该有混沌效应")

	t.Logf("异常体有 %d 个混沌效应:", len(effects))
	for _, effect := range effects {
		assert.NotEmpty(t, effect.ID)
		assert.NotEmpty(t, effect.Name)
		assert.Greater(t, effect.Cost, 0)
		assert.NotEmpty(t, effect.Description)
		assert.NotEmpty(t, effect.Effect)
		t.Logf("  - %s (成本: %d): %s", effect.Name, effect.Cost, effect.Description)
	}
}
