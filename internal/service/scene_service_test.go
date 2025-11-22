package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

func setupSceneTest(t *testing.T) (ScenarioService, GameService, SceneService, *domain.GameSession) {
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
	sceneService := NewSceneService(scenarioService, gameService)

	agent := createTestAgentForScene()
	session, err := gameService.CreateSession(agent.ID, testScenario.ID)
	require.NoError(t, err)

	return scenarioService, gameService, sceneService, session
}

func TestSceneService_LoadScene(t *testing.T) {
	_, _, sceneService, session := setupSceneTest(t)

	t.Run("加载存在的场景", func(t *testing.T) {
		scene, err := sceneService.LoadScene(session.ID, "scene-1")
		assert.NoError(t, err)
		assert.NotNil(t, scene)
		assert.Equal(t, "scene-1", scene.ID)
		assert.Equal(t, "工厂入口", scene.Name)
	})

	t.Run("加载不存在的场景", func(t *testing.T) {
		_, err := sceneService.LoadScene(session.ID, "non-existent")
		assert.Error(t, err)
	})

	t.Run("加载不存在的会话", func(t *testing.T) {
		_, err := sceneService.LoadScene("non-existent-session", "scene-1")
		assert.Error(t, err)
	})
}

func TestSceneService_GetCurrentScene(t *testing.T) {
	_, gameService, sceneService, session := setupSceneTest(t)

	t.Run("获取当前场景", func(t *testing.T) {
		session.State.CurrentSceneID = "scene-1"
		err := gameService.SaveSession(session)
		require.NoError(t, err)

		scene, err := sceneService.GetCurrentScene(session.ID)
		assert.NoError(t, err)
		assert.NotNil(t, scene)
		assert.Equal(t, "scene-1", scene.ID)
	})

	t.Run("当前场景未设置", func(t *testing.T) {
		agent := createTestAgentForScene()
		session2, err := gameService.CreateSession(agent.ID, "test-scenario")
		require.NoError(t, err)
		session2.State.CurrentSceneID = ""
		err = gameService.SaveSession(session2)
		require.NoError(t, err)

		_, err = sceneService.GetCurrentScene(session2.ID)
		assert.Error(t, err)
	})
}

func TestSceneService_SaveAndGetSceneState(t *testing.T) {
	_, _, sceneService, session := setupSceneTest(t)

	t.Run("保存和获取场景状态", func(t *testing.T) {
		state := map[string]any{
			"door_opened": true,
			"light_on":    false,
			"visited":     1,
		}

		err := sceneService.SaveSceneState(session.ID, "scene-1", state)
		assert.NoError(t, err)

		retrievedState, err := sceneService.GetSceneState(session.ID, "scene-1")
		assert.NoError(t, err)
		assert.Equal(t, true, retrievedState["door_opened"])
		assert.Equal(t, false, retrievedState["light_on"])
		assert.Equal(t, 1, retrievedState["visited"])
	})

	t.Run("获取不存在的场景状态", func(t *testing.T) {
		state, err := sceneService.GetSceneState(session.ID, "non-existent")
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.Empty(t, state)
	})

	t.Run("保存空状态", func(t *testing.T) {
		err := sceneService.SaveSceneState(session.ID, "scene-1", nil)
		assert.Error(t, err)
	})
}

func TestSceneService_TransitionToScene(t *testing.T) {
	_, gameService, sceneService, session := setupSceneTest(t)

	t.Run("切换到新场景", func(t *testing.T) {
		session.State.CurrentSceneID = "scene-1"
		err := gameService.SaveSession(session)
		require.NoError(t, err)

		err = sceneService.TransitionToScene(session.ID, "scene-2")
		assert.NoError(t, err)

		updatedSession, err := gameService.GetSession(session.ID)
		assert.NoError(t, err)
		assert.Equal(t, "scene-2", updatedSession.State.CurrentSceneID)
		assert.True(t, updatedSession.State.VisitedScenes["scene-2"])
	})

	t.Run("切换到不存在的场景", func(t *testing.T) {
		err := sceneService.TransitionToScene(session.ID, "non-existent")
		assert.Error(t, err)
	})

	t.Run("场景状态持久化", func(t *testing.T) {
		state1 := map[string]any{"key": "value1"}
		err := sceneService.SaveSceneState(session.ID, "scene-1", state1)
		require.NoError(t, err)

		session.State.CurrentSceneID = "scene-1"
		err = gameService.SaveSession(session)
		require.NoError(t, err)

		err = sceneService.TransitionToScene(session.ID, "scene-2")
		assert.NoError(t, err)

		savedState, err := sceneService.GetSceneState(session.ID, "scene-1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", savedState["key"])
	})
}

func TestSceneService_InteractWithObject(t *testing.T) {
	_, gameService, sceneService, session := setupSceneTest(t)

	session.State.CurrentSceneID = "scene-1"
	err := gameService.SaveSession(session)
	require.NoError(t, err)

	t.Run("与线索交互", func(t *testing.T) {
		result, err := sceneService.InteractWithObject(session.ID, "clue-1", "调查")
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.CluesGained, "clue-1")
		assert.Contains(t, result.Description, "脚印")

		updatedSession, err := gameService.GetSession(session.ID)
		assert.NoError(t, err)
		assert.Contains(t, updatedSession.State.CollectedClues, "clue-1")
		assert.Contains(t, updatedSession.State.UnlockedLocations, "scene-2")
	})

	t.Run("与NPC交互", func(t *testing.T) {
		result, err := sceneService.InteractWithObject(session.ID, "npc-1", "对话")
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Description, "工厂工人")
	})

	t.Run("与不存在的对象交互", func(t *testing.T) {
		_, err := sceneService.InteractWithObject(session.ID, "non-existent", "action")
		assert.Error(t, err)
	})
}

func createTestAgentForScene() *domain.Agent {
	return &domain.Agent{
		ID:       "test-agent-scene-1",
		Name:     "测试特工",
		Pronouns: "他/him",
		Anomaly: &domain.Anomaly{
			Type: "低语",
			Abilities: []*domain.AnomalyAbility{
				{ID: "ability1", Name: "能力1", AnomalyType: "低语"},
			},
		},
		Reality: &domain.Reality{
			Type:             "看护者",
			Trigger:          &domain.RealityTrigger{Name: "触发", Cost: 0, Effect: "效果", Consequence: "后果"},
			OverloadRelief:   &domain.OverloadRelief{Name: "解除", Condition: "条件", Effect: "效果"},
			DegradationTrack: &domain.DegradationTrack{Name: "轨道", Filled: 0, Total: 5},
		},
		Career: &domain.Career{Type: "公关", QA: map[string]int{"专注": 1}},
		QA:     map[string]int{"专注": 1},
	}
}
