package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
)

// TestProperty_NPCStateConsistency 属性18: NPC状态一致性
// Feature: trpg-solo-engine, Property 18: NPC状态一致性
// 验证需求: 15.3, 15.4
//
// 对于任何NPC，其状态变化应该反映在后续的对话和行为中。
// 被异常体影响的NPC应该表现出不同的行为模式。
func TestProperty_NPCStateConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试剧本并保存到文件
	testScenario := CreateTestScenario()
	data, err := json.MarshalIndent(testScenario, "", "  ")
	require.NoError(t, err)

	scenarioPath := filepath.Join(tempDir, testScenario.ID+".json")
	err = os.WriteFile(scenarioPath, data, 0644)
	require.NoError(t, err)

	scenarioService := NewScenarioService(tempDir)
	gameService := NewGameService()
	npcService := NewNPCService(scenarioService, gameService)

	properties.Property("NPC状态变化应该在后续查询中反映", prop.ForAll(
		func(newState string, anomalyAffected bool) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 加载NPC
			npcInfo1, err := npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 记录初始状态
			initialState := npcInfo1.CurrentState
			initialAffected := npcInfo1.AnomalyAffected

			// 更新状态
			err = npcService.UpdateNPCState(session.ID, "npc-1", newState)
			if err != nil {
				return false
			}

			err = npcService.SetAnomalyAffected(session.ID, "npc-1", anomalyAffected)
			if err != nil {
				return false
			}

			// 重新加载NPC
			npcInfo2, err := npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 验证状态已更新
			if npcInfo2.CurrentState != newState {
				return false
			}

			if npcInfo2.AnomalyAffected != anomalyAffected {
				return false
			}

			// 验证状态确实发生了变化（如果参数不同）
			if newState != initialState && npcInfo2.CurrentState == initialState {
				return false
			}

			if anomalyAffected != initialAffected && npcInfo2.AnomalyAffected == initialAffected {
				return false
			}

			return true
		},
		genNPCState(),
		gen.Bool(),
	))

	properties.Property("异常影响应该自动设置AnomalyAffected标志", prop.ForAll(
		func(anomalyType string, effect string) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 加载NPC
			_, err = npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 记录异常影响
			influence := &AnomalyInfluence{
				AnomalyType: anomalyType,
				Effect:      effect,
				Description: "测试影响",
			}

			err = npcService.RecordAnomalyInfluence(session.ID, "npc-1", influence)
			if err != nil {
				return false
			}

			// 验证AnomalyAffected标志被设置
			npcState, err := npcService.GetNPCState(session.ID, "npc-1")
			if err != nil {
				return false
			}

			if !npcState.AnomalyAffected {
				return false
			}

			// 验证通过LoadNPC也能看到
			npcInfo, err := npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			if !npcInfo.AnomalyAffected {
				return false
			}

			// 验证影响记录存在
			if !npcService.HasAnomalyInfluence(session.ID, "npc-1") {
				return false
			}

			return true
		},
		genAnomalyType(),
		genAnomalyEffect(),
	))

	properties.Property("关系值变化应该累积", prop.ForAll(
		func(deltas []int) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 加载NPC
			_, err = npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 计算期望的最终关系值
			expectedRelationship := 0
			for _, delta := range deltas {
				expectedRelationship += delta
			}

			// 应用所有关系变化
			for _, delta := range deltas {
				err = npcService.ModifyRelationship(session.ID, "npc-1", delta)
				if err != nil {
					return false
				}
			}

			// 验证最终关系值
			actualRelationship, err := npcService.GetRelationship(session.ID, "npc-1")
			if err != nil {
				return false
			}

			return actualRelationship == expectedRelationship
		},
		genRelationshipDeltas(),
	))

	properties.Property("多个NPC的状态应该独立", prop.ForAll(
		func(state1 string, state2 string, affected1 bool, affected2 bool) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 加载两个NPC
			_, err = npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			_, err = npcService.LoadNPC(session.ID, "npc-2")
			if err != nil {
				return false
			}

			// 设置不同的状态
			err = npcService.UpdateNPCState(session.ID, "npc-1", state1)
			if err != nil {
				return false
			}

			err = npcService.UpdateNPCState(session.ID, "npc-2", state2)
			if err != nil {
				return false
			}

			err = npcService.SetAnomalyAffected(session.ID, "npc-1", affected1)
			if err != nil {
				return false
			}

			err = npcService.SetAnomalyAffected(session.ID, "npc-2", affected2)
			if err != nil {
				return false
			}

			// 验证状态独立
			npcState1, err := npcService.GetNPCState(session.ID, "npc-1")
			if err != nil {
				return false
			}

			npcState2, err := npcService.GetNPCState(session.ID, "npc-2")
			if err != nil {
				return false
			}

			if npcState1.CurrentState != state1 {
				return false
			}

			if npcState2.CurrentState != state2 {
				return false
			}

			if npcState1.AnomalyAffected != affected1 {
				return false
			}

			if npcState2.AnomalyAffected != affected2 {
				return false
			}

			return true
		},
		genNPCState(),
		genNPCState(),
		gen.Bool(),
		gen.Bool(),
	))

	properties.Property("异常影响记录应该按时间顺序保存", prop.ForAll(
		func(influenceCount int) bool {
			if influenceCount < 1 || influenceCount > 10 {
				return true // 跳过无效范围
			}

			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 加载NPC
			_, err = npcService.LoadNPC(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 记录多个影响
			for i := 0; i < influenceCount; i++ {
				influence := &AnomalyInfluence{
					AnomalyType: "test",
					Effect:      "effect",
					Description: "影响" + string(rune(i)),
				}

				err = npcService.RecordAnomalyInfluence(session.ID, "npc-1", influence)
				if err != nil {
					return false
				}
			}

			// 获取所有影响
			influences, err := npcService.GetAnomalyInfluences(session.ID, "npc-1")
			if err != nil {
				return false
			}

			// 验证数量
			if len(influences) != influenceCount {
				return false
			}

			// 验证时间顺序（每个影响的时间戳应该不早于前一个）
			for i := 1; i < len(influences); i++ {
				if influences[i].Timestamp.Before(influences[i-1].Timestamp) {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// genNPCState 生成NPC状态
func genNPCState() gopter.Gen {
	return gen.OneConstOf(
		"normal",
		"suspicious",
		"hostile",
		"friendly",
		"afraid",
		"confused",
		"controlled",
	)
}

// genAnomalyType 生成异常体类型
func genAnomalyType() gopter.Gen {
	return gen.OneConstOf(
		"whisper",
		"catalog",
		"drain",
		"clock",
		"growth",
		"gun",
		"dream",
		"manifold",
		"absence",
	)
}

// genAnomalyEffect 生成异常效应
func genAnomalyEffect() gopter.Gen {
	return gen.OneConstOf(
		"mind_control",
		"memory_alteration",
		"fear",
		"confusion",
		"hallucination",
		"possession",
	)
}

// genRelationshipDeltas 生成关系值变化序列
func genRelationshipDeltas() gopter.Gen {
	return gen.SliceOf(gen.IntRange(-10, 10))
}
