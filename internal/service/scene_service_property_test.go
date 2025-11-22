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

// TestProperty_SceneStatePersistence 属性16: 场景状态持久化
// Feature: trpg-solo-engine, Property 16: 场景状态持久化
// 验证需求: 13.2, 13.5
//
// 对于任何场景，玩家离开后再返回，场景应该保持离开时的状态，而不是初始状态。
// 场景的变化应该被记录并恢复。
func TestProperty_SceneStatePersistence(t *testing.T) {
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

	// 创建游戏服务
	gameService := NewGameService()

	// 创建场景服务
	sceneService := NewSceneService(scenarioService, gameService)

	properties.Property("场景状态在离开后返回时应该保持", prop.ForAll(
		func(stateValues map[string]any) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 设置当前场景为scene-1
			session.State.CurrentSceneID = "scene-1"
			err = gameService.SaveSession(session)
			if err != nil {
				return false
			}

			// 保存场景状态
			err = sceneService.SaveSceneState(session.ID, "scene-1", stateValues)
			if err != nil {
				return false
			}

			// 切换到scene-2（离开scene-1）
			err = sceneService.TransitionToScene(session.ID, "scene-2")
			if err != nil {
				return false
			}

			// 返回scene-1
			err = sceneService.TransitionToScene(session.ID, "scene-1")
			if err != nil {
				return false
			}

			// 加载scene-1并检查状态
			scene, err := sceneService.LoadScene(session.ID, "scene-1")
			if err != nil {
				return false
			}

			// 验证状态保持不变
			if len(scene.State) != len(stateValues) {
				return false
			}

			for key, expectedValue := range stateValues {
				actualValue, exists := scene.State[key]
				if !exists {
					return false
				}
				if actualValue != expectedValue {
					return false
				}
			}

			return true
		},
		genSceneState(),
	))

	properties.Property("场景状态变化应该被记录", prop.ForAll(
		func(initialState map[string]any, modifiedState map[string]any) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 设置当前场景
			session.State.CurrentSceneID = "scene-1"
			err = gameService.SaveSession(session)
			if err != nil {
				return false
			}

			// 保存初始状态
			err = sceneService.SaveSceneState(session.ID, "scene-1", initialState)
			if err != nil {
				return false
			}

			// 修改状态
			err = sceneService.SaveSceneState(session.ID, "scene-1", modifiedState)
			if err != nil {
				return false
			}

			// 获取状态
			retrievedState, err := sceneService.GetSceneState(session.ID, "scene-1")
			if err != nil {
				return false
			}

			// 验证状态是修改后的状态，而不是初始状态
			if len(retrievedState) != len(modifiedState) {
				return false
			}

			for key, expectedValue := range modifiedState {
				actualValue, exists := retrievedState[key]
				if !exists {
					return false
				}
				if actualValue != expectedValue {
					return false
				}
			}

			return true
		},
		genSceneState(),
		genSceneState(),
	))

	properties.Property("多个场景的状态应该独立保存", prop.ForAll(
		func(state1 map[string]any, state2 map[string]any) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 保存两个场景的状态
			err = sceneService.SaveSceneState(session.ID, "scene-1", state1)
			if err != nil {
				return false
			}

			err = sceneService.SaveSceneState(session.ID, "scene-2", state2)
			if err != nil {
				return false
			}

			// 获取两个场景的状态
			retrievedState1, err := sceneService.GetSceneState(session.ID, "scene-1")
			if err != nil {
				return false
			}

			retrievedState2, err := sceneService.GetSceneState(session.ID, "scene-2")
			if err != nil {
				return false
			}

			// 验证状态独立
			for key, expectedValue := range state1 {
				actualValue, exists := retrievedState1[key]
				if !exists {
					return false
				}
				if actualValue != expectedValue {
					return false
				}
			}

			for key, expectedValue := range state2 {
				actualValue, exists := retrievedState2[key]
				if !exists {
					return false
				}
				if actualValue != expectedValue {
					return false
				}
			}

			return true
		},
		genSceneState(),
		genSceneState(),
	))

	properties.Property("场景状态持久化后应该可以恢复", prop.ForAll(
		func(states map[string]map[string]any) bool {
			// 创建测试会话
			agent := createTestAgent()
			session, err := gameService.CreateSession(agent.ID, testScenario.ID)
			if err != nil {
				return false
			}

			// 保存多个场景的状态
			for sceneID, state := range states {
				// 只使用测试剧本中存在的场景
				if sceneID != "scene-1" && sceneID != "scene-2" {
					continue
				}
				err = sceneService.SaveSceneState(session.ID, sceneID, state)
				if err != nil {
					return false
				}
			}

			// 持久化
			err = sceneService.PersistSceneStates(session.ID)
			if err != nil {
				return false
			}

			// 加载持久化的状态
			loadedStates, err := sceneService.LoadPersistedSceneStates(session.ID)
			if err != nil {
				return false
			}

			// 验证所有状态都被正确恢复
			for sceneID, expectedState := range states {
				if sceneID != "scene-1" && sceneID != "scene-2" {
					continue
				}

				actualState, exists := loadedStates[sceneID]
				if !exists {
					return false
				}

				if len(actualState) != len(expectedState) {
					return false
				}

				for key, expectedValue := range expectedState {
					actualValue, exists := actualState[key]
					if !exists {
						return false
					}
					if actualValue != expectedValue {
						return false
					}
				}
			}

			return true
		},
		genMultipleSceneStates(),
	))

	properties.TestingRun(t)
}

// genSceneState 生成场景状态
func genSceneState() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(map[string]any{"key1": "value1", "key2": 42}),
		gen.Const(map[string]any{"door_opened": true, "light_on": false}),
		gen.Const(map[string]any{"visited": 1, "completed": true, "name": "test"}),
		gen.Const(map[string]any{"counter": 100}),
		gen.Const(map[string]any{"flag": false, "score": 50}),
	)
}

// genMultipleSceneStates 生成多个场景的状态
func genMultipleSceneStates() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(map[string]map[string]any{
			"scene-1": {"key1": "value1", "key2": 42},
		}),
		gen.Const(map[string]map[string]any{
			"scene-2": {"key3": true, "key4": "value4"},
		}),
		gen.Const(map[string]map[string]any{
			"scene-1": {"key5": 100},
			"scene-2": {"key6": false},
		}),
	)
}
