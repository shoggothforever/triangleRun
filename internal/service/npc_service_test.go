package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

func setupNPCTest(t *testing.T) (ScenarioService, GameService, NPCService, *domain.GameSession) {
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

	agent := createTestAgent()
	session, err := gameService.CreateSession(agent.ID, testScenario.ID)
	require.NoError(t, err)

	return scenarioService, gameService, npcService, session
}

// TestNPCService_LoadNPC 测试NPC加载
func TestNPCService_LoadNPC(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("加载存在的NPC", func(t *testing.T) {
		npcInfo, err := npcService.LoadNPC(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.NotNil(t, npcInfo)
		assert.Equal(t, "npc-1", npcInfo.ID)
		assert.Equal(t, "工厂工人", npcInfo.Name)
		assert.Equal(t, "normal", npcInfo.CurrentState)
		assert.False(t, npcInfo.AnomalyAffected)
		assert.Equal(t, 0, npcInfo.Relationship)
	})

	t.Run("加载不存在的NPC", func(t *testing.T) {
		_, err := npcService.LoadNPC(session.ID, "non-existent")
		assert.Error(t, err)
		gameErr, ok := err.(*domain.GameError)
		assert.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, gameErr.Code)
	})

	t.Run("加载不存在的会话", func(t *testing.T) {
		_, err := npcService.LoadNPC("non-existent-session", "npc-1")
		assert.Error(t, err)
	})
}

// TestNPCService_UpdateNPCState 测试状态更新
func TestNPCService_UpdateNPCState(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("更新NPC状态", func(t *testing.T) {
		// 初始加载
		npcInfo, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)
		assert.Equal(t, "normal", npcInfo.CurrentState)

		// 更新状态
		err = npcService.UpdateNPCState(session.ID, "npc-1", "suspicious")
		assert.NoError(t, err)

		// 验证状态已更新
		npcState, err := npcService.GetNPCState(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, "suspicious", npcState.CurrentState)
	})

	t.Run("更新多个状态", func(t *testing.T) {
		// 更新为suspicious
		err := npcService.UpdateNPCState(session.ID, "npc-1", "suspicious")
		require.NoError(t, err)

		// 再更新为hostile
		err = npcService.UpdateNPCState(session.ID, "npc-1", "hostile")
		require.NoError(t, err)

		// 验证最终状态
		npcState, err := npcService.GetNPCState(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, "hostile", npcState.CurrentState)
	})

	t.Run("更新不存在的会话", func(t *testing.T) {
		err := npcService.UpdateNPCState("non-existent-session", "npc-1", "suspicious")
		assert.Error(t, err)
	})
}

// TestNPCService_SetAnomalyAffected 测试异常影响
func TestNPCService_SetAnomalyAffected(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("设置异常影响", func(t *testing.T) {
		// 初始状态
		npcInfo, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)
		assert.False(t, npcInfo.AnomalyAffected)

		// 设置为受影响
		err = npcService.SetAnomalyAffected(session.ID, "npc-1", true)
		assert.NoError(t, err)

		// 验证状态
		npcState, err := npcService.GetNPCState(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.True(t, npcState.AnomalyAffected)
	})

	t.Run("取消异常影响", func(t *testing.T) {
		// 先设置为受影响
		err := npcService.SetAnomalyAffected(session.ID, "npc-1", true)
		require.NoError(t, err)

		// 取消影响
		err = npcService.SetAnomalyAffected(session.ID, "npc-1", false)
		assert.NoError(t, err)

		// 验证状态
		npcState, err := npcService.GetNPCState(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.False(t, npcState.AnomalyAffected)
	})
}

// TestNPCService_Relationship 测试关系变化
func TestNPCService_Relationship(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("获取初始关系值", func(t *testing.T) {
		// 初始化NPC
		_, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)

		relationship, err := npcService.GetRelationship(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, 0, relationship)
	})

	t.Run("修改关系值", func(t *testing.T) {
		// 初始化NPC
		_, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)

		// 增加关系值
		err = npcService.ModifyRelationship(session.ID, "npc-1", 5)
		assert.NoError(t, err)

		relationship, err := npcService.GetRelationship(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, 5, relationship)

		// 再次增加
		err = npcService.ModifyRelationship(session.ID, "npc-1", 3)
		assert.NoError(t, err)

		relationship, err = npcService.GetRelationship(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, 8, relationship)

		// 减少关系值
		err = npcService.ModifyRelationship(session.ID, "npc-1", -10)
		assert.NoError(t, err)

		relationship, err = npcService.GetRelationship(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.Equal(t, -2, relationship)
	})

	t.Run("设置关系值", func(t *testing.T) {
		// 初始化NPC
		_, err := npcService.LoadNPC(session.ID, "npc-2")
		require.NoError(t, err)

		// 直接设置关系值
		err = npcService.SetRelationship(session.ID, "npc-2", 10)
		assert.NoError(t, err)

		relationship, err := npcService.GetRelationship(session.ID, "npc-2")
		assert.NoError(t, err)
		assert.Equal(t, 10, relationship)

		// 重新设置
		err = npcService.SetRelationship(session.ID, "npc-2", -5)
		assert.NoError(t, err)

		relationship, err = npcService.GetRelationship(session.ID, "npc-2")
		assert.NoError(t, err)
		assert.Equal(t, -5, relationship)
	})
}

// TestNPCService_AnomalyInfluence 测试异常影响记录
func TestNPCService_AnomalyInfluence(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("记录异常影响", func(t *testing.T) {
		influence := &AnomalyInfluence{
			AnomalyType: "whisper",
			Effect:      "mind_control",
			Description: "NPC被低语异常体影响",
			Data: map[string]interface{}{
				"intensity": 5,
			},
		}

		err := npcService.RecordAnomalyInfluence(session.ID, "npc-1", influence)
		assert.NoError(t, err)

		// 验证NPC被标记为受影响
		npcState, err := npcService.GetNPCState(session.ID, "npc-1")
		assert.NoError(t, err)
		assert.True(t, npcState.AnomalyAffected)

		// 验证影响记录存在
		assert.True(t, npcService.HasAnomalyInfluence(session.ID, "npc-1"))
	})

	t.Run("获取异常影响记录", func(t *testing.T) {
		// 使用新的会话避免测试间干扰
		_, _, npcService2, session2 := setupNPCTest(t)

		influence1 := &AnomalyInfluence{
			AnomalyType: "whisper",
			Effect:      "mind_control",
			Description: "第一次影响",
		}

		influence2 := &AnomalyInfluence{
			AnomalyType: "catalog",
			Effect:      "memory_alteration",
			Description: "第二次影响",
		}

		err := npcService2.RecordAnomalyInfluence(session2.ID, "npc-1", influence1)
		require.NoError(t, err)

		err = npcService2.RecordAnomalyInfluence(session2.ID, "npc-1", influence2)
		require.NoError(t, err)

		influences, err := npcService2.GetAnomalyInfluences(session2.ID, "npc-1")
		assert.NoError(t, err)
		assert.Len(t, influences, 2)
		assert.Equal(t, "whisper", influences[0].AnomalyType)
		assert.Equal(t, "catalog", influences[1].AnomalyType)
	})

	t.Run("检查无影响的NPC", func(t *testing.T) {
		// 初始化一个新NPC
		_, err := npcService.LoadNPC(session.ID, "npc-2")
		require.NoError(t, err)

		// 检查是否有影响
		assert.False(t, npcService.HasAnomalyInfluence(session.ID, "npc-2"))

		influences, err := npcService.GetAnomalyInfluences(session.ID, "npc-2")
		assert.NoError(t, err)
		assert.Empty(t, influences)
	})

	t.Run("影响记录包含时间戳", func(t *testing.T) {
		influence := &AnomalyInfluence{
			AnomalyType: "dream",
			Effect:      "nightmare",
			Description: "梦境影响",
		}

		before := time.Now()
		err := npcService.RecordAnomalyInfluence(session.ID, "npc-1", influence)
		require.NoError(t, err)
		after := time.Now()

		influences, err := npcService.GetAnomalyInfluences(session.ID, "npc-1")
		require.NoError(t, err)
		require.NotEmpty(t, influences)

		lastInfluence := influences[len(influences)-1]
		assert.False(t, lastInfluence.Timestamp.IsZero())
		assert.True(t, lastInfluence.Timestamp.After(before) || lastInfluence.Timestamp.Equal(before))
		assert.True(t, lastInfluence.Timestamp.Before(after) || lastInfluence.Timestamp.Equal(after))
	})
}

// TestNPCService_GetNPCsInScene 测试场景中的NPC
func TestNPCService_GetNPCsInScene(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("获取场景中的所有NPC", func(t *testing.T) {
		npcs, err := npcService.GetNPCsInScene(session.ID, "scene-1")
		assert.NoError(t, err)
		assert.Len(t, npcs, 2)

		// 验证NPC信息
		npcMap := make(map[string]*NPCInfo)
		for _, npc := range npcs {
			npcMap[npc.ID] = npc
		}

		assert.Contains(t, npcMap, "npc-1")
		assert.Contains(t, npcMap, "npc-2")
		assert.Equal(t, "工厂工人", npcMap["npc-1"].Name)
		assert.Equal(t, "保安", npcMap["npc-2"].Name)
	})

	t.Run("获取空场景的NPC", func(t *testing.T) {
		npcs, err := npcService.GetNPCsInScene(session.ID, "scene-2")
		assert.NoError(t, err)
		assert.Empty(t, npcs)
	})

	t.Run("获取不存在的场景", func(t *testing.T) {
		_, err := npcService.GetNPCsInScene(session.ID, "non-existent")
		assert.Error(t, err)
	})
}

// TestNPCService_GetAllNPCStates 测试获取所有NPC状态
func TestNPCService_GetAllNPCStates(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("获取所有NPC状态", func(t *testing.T) {
		// 初始化几个NPC
		_, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)
		_, err = npcService.LoadNPC(session.ID, "npc-2")
		require.NoError(t, err)

		// 修改一些状态
		err = npcService.UpdateNPCState(session.ID, "npc-1", "suspicious")
		require.NoError(t, err)
		err = npcService.SetAnomalyAffected(session.ID, "npc-2", true)
		require.NoError(t, err)

		// 获取所有状态
		allStates, err := npcService.GetAllNPCStates(session.ID)
		assert.NoError(t, err)
		assert.Len(t, allStates, 2)

		// 验证状态
		assert.Equal(t, "suspicious", allStates["npc-1"].CurrentState)
		assert.True(t, allStates["npc-2"].AnomalyAffected)
	})

	t.Run("空会话返回空map", func(t *testing.T) {
		_, gameService2, npcService2, _ := setupNPCTest(t)
		agent := createTestAgent()
		newSession, err := gameService2.CreateSession(agent.ID, "test-scenario")
		require.NoError(t, err)

		allStates, err := npcService2.GetAllNPCStates(newSession.ID)
		assert.NoError(t, err)
		assert.Empty(t, allStates)
	})
}

// TestNPCService_StateConsistency 测试状态一致性
func TestNPCService_StateConsistency(t *testing.T) {
	_, _, npcService, session := setupNPCTest(t)

	t.Run("状态变化应该在后续查询中反映", func(t *testing.T) {
		// 加载NPC
		npcInfo1, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)
		assert.Equal(t, "normal", npcInfo1.CurrentState)
		assert.False(t, npcInfo1.AnomalyAffected)

		// 修改状态
		err = npcService.UpdateNPCState(session.ID, "npc-1", "hostile")
		require.NoError(t, err)
		err = npcService.SetAnomalyAffected(session.ID, "npc-1", true)
		require.NoError(t, err)

		// 重新加载NPC
		npcInfo2, err := npcService.LoadNPC(session.ID, "npc-1")
		require.NoError(t, err)
		assert.Equal(t, "hostile", npcInfo2.CurrentState)
		assert.True(t, npcInfo2.AnomalyAffected)
	})

	t.Run("异常影响应该自动设置AnomalyAffected标志", func(t *testing.T) {
		// 加载NPC
		_, err := npcService.LoadNPC(session.ID, "npc-2")
		require.NoError(t, err)

		// 记录异常影响
		influence := &AnomalyInfluence{
			AnomalyType: "whisper",
			Effect:      "mind_control",
			Description: "测试影响",
		}
		err = npcService.RecordAnomalyInfluence(session.ID, "npc-2", influence)
		require.NoError(t, err)

		// 验证标志被设置
		npcState, err := npcService.GetNPCState(session.ID, "npc-2")
		assert.NoError(t, err)
		assert.True(t, npcState.AnomalyAffected)

		// 验证通过LoadNPC也能看到
		npcInfo, err := npcService.LoadNPC(session.ID, "npc-2")
		assert.NoError(t, err)
		assert.True(t, npcInfo.AnomalyAffected)
	})
}
