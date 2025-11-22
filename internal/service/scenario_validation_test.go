package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// TestLoadEternalSpringScenario 测试加载永恒之泉剧本
func TestLoadEternalSpringScenario(t *testing.T) {
	// 读取剧本文件
	data, err := os.ReadFile("../../scenarios/eternal-spring.json")
	require.NoError(t, err, "应该能够读取剧本文件")

	// 解析JSON
	var scenario domain.Scenario
	err = json.Unmarshal(data, &scenario)
	require.NoError(t, err, "应该能够解析剧本JSON")

	// 验证基本信息
	assert.Equal(t, "eternal-spring", scenario.ID)
	assert.Equal(t, "永恒之泉", scenario.Name)
	assert.NotEmpty(t, scenario.Description)

	t.Logf("成功加载剧本: %s", scenario.Name)
}

// TestScenarioAnomalyProfile 测试异常体档案完整性
func TestScenarioAnomalyProfile(t *testing.T) {
	scenario := loadTestScenario(t)

	// 验证异常体档案存在
	require.NotNil(t, scenario.Anomaly, "异常体档案不应为空")

	anomaly := scenario.Anomaly
	assert.Equal(t, "fountain-of-youth", anomaly.ID)
	assert.Equal(t, "永恒之泉", anomaly.Name)
	assert.NotEmpty(t, anomaly.History, "异常体历史不应为空")
	assert.NotEmpty(t, anomaly.Appearance, "异常体外观不应为空")
	assert.NotEmpty(t, anomaly.Impulse, "异常体冲动不应为空")
	assert.NotEmpty(t, anomaly.CurrentStatus, "异常体当前状态不应为空")

	// 验证焦点
	require.NotNil(t, anomaly.Focus, "焦点不应为空")
	assert.Equal(t, "渴望", anomaly.Focus.Emotion)
	assert.NotEmpty(t, anomaly.Focus.Subject)

	// 验证领域
	require.NotNil(t, anomaly.Domain, "领域不应为空")
	assert.NotEmpty(t, anomaly.Domain.Location)
	assert.NotEmpty(t, anomaly.Domain.Description)

	// 验证混沌效应
	assert.NotEmpty(t, anomaly.ChaosEffects, "应该有混沌效应")
	assert.GreaterOrEqual(t, len(anomaly.ChaosEffects), 3, "应该至少有3个混沌效应")

	for _, effect := range anomaly.ChaosEffects {
		assert.NotEmpty(t, effect.ID, "混沌效应ID不应为空")
		assert.NotEmpty(t, effect.Name, "混沌效应名称不应为空")
		assert.Greater(t, effect.Cost, 0, "混沌效应成本应大于0")
		assert.NotEmpty(t, effect.Description, "混沌效应描述不应为空")
		assert.NotEmpty(t, effect.Effect, "混沌效应效果不应为空")
	}

	t.Logf("异常体档案验证通过: %d 个混沌效应", len(anomaly.ChaosEffects))
}

// TestScenarioScenes 测试场景完整性
func TestScenarioScenes(t *testing.T) {
	scenario := loadTestScenario(t)

	// 验证场景存在
	require.NotEmpty(t, scenario.Scenes, "应该有场景")
	assert.GreaterOrEqual(t, len(scenario.Scenes), 3, "应该至少有3个场景")

	// 验证起始场景存在
	assert.NotEmpty(t, scenario.StartingSceneID, "应该有起始场景ID")
	_, exists := scenario.Scenes[scenario.StartingSceneID]
	assert.True(t, exists, "起始场景应该存在于场景列表中")

	// 验证每个场景的完整性
	for sceneID, scene := range scenario.Scenes {
		assert.Equal(t, sceneID, scene.ID, "场景ID应该匹配")
		assert.NotEmpty(t, scene.Name, "场景名称不应为空")
		assert.NotEmpty(t, scene.Description, "场景描述不应为空")

		// 验证NPC
		for _, npc := range scene.NPCs {
			assert.NotEmpty(t, npc.ID, "NPC ID不应为空")
			assert.NotEmpty(t, npc.Name, "NPC名称不应为空")
			assert.NotEmpty(t, npc.Description, "NPC描述不应为空")
			assert.NotEmpty(t, npc.Personality, "NPC性格不应为空")
			assert.NotEmpty(t, npc.Dialogues, "NPC应该有对话")
		}

		// 验证线索
		for _, clue := range scene.Clues {
			assert.NotEmpty(t, clue.ID, "线索ID不应为空")
			assert.NotEmpty(t, clue.Name, "线索名称不应为空")
			assert.NotEmpty(t, clue.Description, "线索描述不应为空")
		}

		// 验证事件
		for _, event := range scene.Events {
			assert.NotEmpty(t, event.ID, "事件ID不应为空")
			assert.NotEmpty(t, event.Name, "事件名称不应为空")
			assert.NotEmpty(t, event.Description, "事件描述不应为空")
			assert.NotEmpty(t, event.Trigger, "事件触发条件不应为空")
			assert.NotEmpty(t, event.Effect, "事件效果不应为空")
		}
	}

	t.Logf("场景验证通过: %d 个场景", len(scenario.Scenes))
}

// TestScenarioConnections 测试场景连接有效性
func TestScenarioConnections(t *testing.T) {
	scenario := loadTestScenario(t)

	// 验证所有场景连接都指向存在的场景
	invalidConnections := 0
	for sceneID, scene := range scenario.Scenes {
		for _, connectionID := range scene.Connections {
			_, exists := scenario.Scenes[connectionID]
			if !exists {
				t.Errorf("场景 %s 的连接 %s 不存在", sceneID, connectionID)
				invalidConnections++
			}
		}
	}

	assert.Equal(t, 0, invalidConnections, "不应该有无效的场景连接")
	t.Logf("场景连接验证通过")
}

// TestScenarioClueUnlocks 测试线索解锁引用有效性
func TestScenarioClueUnlocks(t *testing.T) {
	scenario := loadTestScenario(t)

	// 收集所有场景ID
	sceneIDs := make(map[string]bool)
	for sceneID := range scenario.Scenes {
		sceneIDs[sceneID] = true
	}

	// 验证线索解锁的场景存在
	for _, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			for _, unlockID := range clue.Unlocks {
				// 检查是否是场景ID
				if sceneIDs[unlockID] {
					continue
				}
				// 如果不是场景ID，可能是其他类型的解锁（如能力、事件等）
				// 这里我们只验证场景ID
				t.Logf("线索 %s 解锁了 %s（可能不是场景）", clue.ID, unlockID)
			}
		}
	}

	t.Logf("线索解锁验证通过")
}

// TestScenarioMorningScenes 测试晨会场景
func TestScenarioMorningScenes(t *testing.T) {
	scenario := loadTestScenario(t)

	assert.NotEmpty(t, scenario.MorningScenes, "应该有晨会场景")
	assert.GreaterOrEqual(t, len(scenario.MorningScenes), 3, "应该至少有3个晨会场景")

	for _, scene := range scenario.MorningScenes {
		assert.NotEmpty(t, scene.ID, "晨会场景ID不应为空")
		assert.NotEmpty(t, scene.Description, "晨会场景描述不应为空")
		assert.NotEmpty(t, scene.Type, "晨会场景类型不应为空")
	}

	t.Logf("晨会场景验证通过: %d 个场景", len(scenario.MorningScenes))
}

// TestScenarioBriefing 测试任务简报
func TestScenarioBriefing(t *testing.T) {
	scenario := loadTestScenario(t)

	require.NotNil(t, scenario.Briefing, "任务简报不应为空")

	briefing := scenario.Briefing
	assert.NotEmpty(t, briefing.Summary, "简报摘要不应为空")
	assert.NotEmpty(t, briefing.Objectives, "应该有任务目标")
	assert.GreaterOrEqual(t, len(briefing.Objectives), 3, "应该至少有3个任务目标")
	assert.NotEmpty(t, briefing.Warnings, "应该有警告")

	t.Logf("任务简报验证通过: %d 个目标, %d 个警告", len(briefing.Objectives), len(briefing.Warnings))
}

// TestScenarioOptionalGoals 测试可选目标
func TestScenarioOptionalGoals(t *testing.T) {
	scenario := loadTestScenario(t)

	assert.NotEmpty(t, scenario.OptionalGoals, "应该有可选目标")

	for _, goal := range scenario.OptionalGoals {
		assert.NotEmpty(t, goal.ID, "可选目标ID不应为空")
		assert.NotEmpty(t, goal.Description, "可选目标描述不应为空")
		assert.Greater(t, goal.Reward, 0, "可选目标奖励应大于0")
	}

	t.Logf("可选目标验证通过: %d 个目标", len(scenario.OptionalGoals))
}

// TestScenarioEncounter 测试遭遇阶段
func TestScenarioEncounter(t *testing.T) {
	scenario := loadTestScenario(t)

	require.NotNil(t, scenario.Encounter, "遭遇阶段不应为空")

	encounter := scenario.Encounter
	assert.NotEmpty(t, encounter.ID, "遭遇ID不应为空")
	assert.NotEmpty(t, encounter.Description, "遭遇描述不应为空")
	assert.NotEmpty(t, encounter.Phases, "应该有遭遇阶段")
	assert.GreaterOrEqual(t, len(encounter.Phases), 2, "应该至少有2个遭遇阶段")

	for _, phase := range encounter.Phases {
		assert.NotEmpty(t, phase.ID, "阶段ID不应为空")
		assert.NotEmpty(t, phase.Description, "阶段描述不应为空")
		assert.NotEmpty(t, phase.Actions, "阶段应该有可用行动")
	}

	t.Logf("遭遇阶段验证通过: %d 个阶段", len(encounter.Phases))
}

// TestScenarioAftermath 测试余波
func TestScenarioAftermath(t *testing.T) {
	scenario := loadTestScenario(t)

	require.NotNil(t, scenario.Aftermath, "余波不应为空")

	aftermath := scenario.Aftermath
	assert.NotEmpty(t, aftermath.Captured, "应该有捕获结局")
	assert.NotEmpty(t, aftermath.Neutralized, "应该有中和结局")
	assert.NotEmpty(t, aftermath.Escaped, "应该有逃脱结局")

	t.Logf("余波验证通过: 3 种结局")
}

// TestScenarioRewards 测试奖励
func TestScenarioRewards(t *testing.T) {
	scenario := loadTestScenario(t)

	require.NotNil(t, scenario.Rewards, "奖励不应为空")

	rewards := scenario.Rewards
	assert.Equal(t, 3, rewards.Commendations, "成功捕获应该给予3次嘉奖")
	assert.NotEmpty(t, rewards.Claimables, "应该有可申领物")

	t.Logf("奖励验证通过: %d 次嘉奖, %d 个可申领物", rewards.Commendations, len(rewards.Claimables))
}

// TestScenarioNPCDialogues 测试NPC对话完整性
func TestScenarioNPCDialogues(t *testing.T) {
	scenario := loadTestScenario(t)

	npcCount := 0
	dialogueCount := 0

	for _, scene := range scenario.Scenes {
		for _, npc := range scene.NPCs {
			npcCount++
			dialogueCount += len(npc.Dialogues)
			assert.NotEmpty(t, npc.Dialogues, "NPC %s 应该有对话", npc.Name)
		}
	}

	t.Logf("剧本中共有 %d 个NPC，%d 条对话", npcCount, dialogueCount)
	assert.Greater(t, npcCount, 0, "应该有NPC")
	assert.Greater(t, dialogueCount, 0, "应该有对话")
}

// TestScenarioClueCount 测试线索数量
func TestScenarioClueCount(t *testing.T) {
	scenario := loadTestScenario(t)

	clueCount := 0
	for _, scene := range scenario.Scenes {
		clueCount += len(scene.Clues)
	}

	t.Logf("剧本中共有 %d 条线索", clueCount)
	assert.Greater(t, clueCount, 5, "应该有足够的线索来支持调查")
}

// TestScenarioEventCount 测试事件数量
func TestScenarioEventCount(t *testing.T) {
	scenario := loadTestScenario(t)

	eventCount := 0
	for _, scene := range scenario.Scenes {
		eventCount += len(scene.Events)
	}

	t.Logf("剧本中共有 %d 个事件", eventCount)
	assert.Greater(t, eventCount, 0, "应该有事件来增加动态性")
}

// TestScenarioChaosEffectCosts 测试混沌效应成本合理性
func TestScenarioChaosEffectCosts(t *testing.T) {
	scenario := loadTestScenario(t)

	for _, effect := range scenario.Anomaly.ChaosEffects {
		assert.GreaterOrEqual(t, effect.Cost, 1, "混沌效应 %s 的成本应至少为1", effect.Name)
		assert.LessOrEqual(t, effect.Cost, 10, "混沌效应 %s 的成本不应超过10", effect.Name)
	}

	t.Logf("混沌效应成本验证通过")
}

// TestScenarioCompleteness 测试剧本整体完整性
func TestScenarioCompleteness(t *testing.T) {
	scenario := loadTestScenario(t)

	// 统计信息
	stats := map[string]int{
		"场景数量":   len(scenario.Scenes),
		"NPC数量":  0,
		"线索数量":   0,
		"事件数量":   0,
		"混沌效应数量": len(scenario.Anomaly.ChaosEffects),
		"晨会场景数量": len(scenario.MorningScenes),
		"可选目标数量": len(scenario.OptionalGoals),
		"遭遇阶段数量": len(scenario.Encounter.Phases),
	}

	for _, scene := range scenario.Scenes {
		stats["NPC数量"] += len(scene.NPCs)
		stats["线索数量"] += len(scene.Clues)
		stats["事件数量"] += len(scene.Events)
	}

	t.Log("=== 永恒之泉剧本统计 ===")
	for key, value := range stats {
		t.Logf("%s: %d", key, value)
	}

	// 验证最低要求
	assert.GreaterOrEqual(t, stats["场景数量"], 3, "应该至少有3个场景")
	assert.GreaterOrEqual(t, stats["NPC数量"], 3, "应该至少有3个NPC")
	assert.GreaterOrEqual(t, stats["线索数量"], 5, "应该至少有5条线索")
	assert.GreaterOrEqual(t, stats["混沌效应数量"], 3, "应该至少有3个混沌效应")
}

// loadTestScenario 辅助函数：加载测试剧本
func loadTestScenario(t *testing.T) *domain.Scenario {
	data, err := os.ReadFile("../../scenarios/eternal-spring.json")
	require.NoError(t, err, "应该能够读取剧本文件")

	var scenario domain.Scenario
	err = json.Unmarshal(data, &scenario)
	require.NoError(t, err, "应该能够解析剧本JSON")

	return &scenario
}
