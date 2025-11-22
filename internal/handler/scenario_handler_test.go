package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/service"
)

func setupScenarioTestRouter() (*gin.Engine, *ScenarioHandler, string) {
	gin.SetMode(gin.TestMode)

	// 创建临时目录用于测试剧本
	tempDir, _ := os.MkdirTemp("", "scenarios-test-*")

	// 创建测试剧本文件
	testScenario := service.CreateTestScenario()
	data, _ := json.Marshal(testScenario)
	scenarioPath := filepath.Join(tempDir, "test-scenario.json")
	os.WriteFile(scenarioPath, data, 0644)

	// 创建服务
	scenarioService := service.NewScenarioService(tempDir)

	// 创建处理器
	handler := NewScenarioHandler(scenarioService)

	// 创建路由
	router := gin.New()
	api := router.Group("/api")
	{
		scenarios := api.Group("/scenarios")
		{
			scenarios.GET("", handler.ListScenarios)
			scenarios.GET("/:id", handler.GetScenario)
			scenarios.GET("/:id/scenes/:sceneId", handler.GetScene)
		}
	}

	return router, handler, tempDir
}

// TestScenarioHandler_ListScenarios 测试列出剧本
func TestScenarioHandler_ListScenarios(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	req, _ := http.NewRequest("GET", "/api/scenarios", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 1, "应该至少有1个剧本")

	// 验证剧本摘要包含必需字段
	if len(data) > 0 {
		scenario := data[0].(map[string]interface{})
		assert.NotEmpty(t, scenario["id"])
		assert.NotEmpty(t, scenario["name"])
		assert.NotEmpty(t, scenario["description"])
	}
}

// TestScenarioHandler_GetScenario 测试获取剧本详情
func TestScenarioHandler_GetScenario(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		scenarioID     string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功获取剧本",
			scenarioID:     "test-scenario",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "剧本不存在",
			scenarioID:     "non-existent-scenario",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:           "空剧本ID",
			scenarioID:     "",
			expectedStatus: http.StatusMovedPermanently, // Gin会返回301重定向
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/scenarios/" + tt.scenarioID
			if tt.scenarioID == "" {
				path = "/api/scenarios/"
			}

			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// 对于重定向响应，跳过JSON解析
			if w.Code == http.StatusMovedPermanently {
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				// 验证剧本数据完整性
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, tt.scenarioID, data["id"])
				assert.NotEmpty(t, data["name"])
				assert.NotEmpty(t, data["description"])

				// 验证异常体档案
				anomaly, ok := data["anomaly"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, anomaly["id"])
				assert.NotEmpty(t, anomaly["name"])

				// 验证场景
				scenes, ok := data["scenes"].(map[string]interface{})
				require.True(t, ok)
				assert.Greater(t, len(scenes), 0, "应该至少有一个场景")

				// 验证起始场景
				assert.NotEmpty(t, data["starting_scene_id"])
			} else {
				// 验证错误信息
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestScenarioHandler_GetScene 测试获取场景
func TestScenarioHandler_GetScene(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		scenarioID     string
		sceneID        string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功获取场景",
			scenarioID:     "test-scenario",
			sceneID:        "scene-1",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "场景不存在",
			scenarioID:     "test-scenario",
			sceneID:        "non-existent-scene",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:           "剧本不存在",
			scenarioID:     "non-existent-scenario",
			sceneID:        "scene-1",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:           "空场景ID",
			scenarioID:     "test-scenario",
			sceneID:        "",
			expectedStatus: http.StatusNotFound, // Gin路由不匹配返回404
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/scenarios/" + tt.scenarioID + "/scenes/" + tt.sceneID
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// 对于重定向或404响应，跳过JSON解析（可能是HTML响应）
			if w.Code == http.StatusMovedPermanently || (w.Code == http.StatusNotFound && tt.sceneID == "") {
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				// 验证场景数据
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, tt.sceneID, data["id"])
				assert.NotEmpty(t, data["name"])
				assert.NotEmpty(t, data["description"])

				// 验证场景包含NPCs
				npcs, ok := data["npcs"].([]interface{})
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(npcs), 0)

				// 验证场景包含线索
				clues, ok := data["clues"].([]interface{})
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(clues), 0)

				// 验证场景包含事件
				events, ok := data["events"].([]interface{})
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(events), 0)

				// 验证场景连接
				connections, ok := data["connections"].([]interface{})
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(connections), 0)
			} else {
				// 验证错误信息
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestScenarioHandler_GetScenarioWithMultipleScenes 测试获取包含多个场景的剧本
func TestScenarioHandler_GetScenarioWithMultipleScenes(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	req, _ := http.NewRequest("GET", "/api/scenarios/test-scenario", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	scenes, ok := data["scenes"].(map[string]interface{})
	require.True(t, ok)

	// 验证测试剧本有2个场景
	assert.Equal(t, 2, len(scenes), "测试剧本应该有2个场景")

	// 验证场景1
	scene1, ok := scenes["scene-1"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "scene-1", scene1["id"])
	assert.Equal(t, "工厂入口", scene1["name"])

	// 验证场景2
	scene2, ok := scenes["scene-2"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "scene-2", scene2["id"])
	assert.Equal(t, "工厂车间", scene2["name"])
}

// TestScenarioHandler_ErrorResponses 测试错误响应
func TestScenarioHandler_ErrorResponses(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "获取不存在的剧本",
			path:           "/api/scenarios/invalid-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "获取不存在剧本的场景",
			path:           "/api/scenarios/invalid-id/scenes/scene-1",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "获取不存在的场景",
			path:           "/api/scenarios/test-scenario/scenes/invalid-scene",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.NotEmpty(t, response["error"])
		})
	}
}

// TestScenarioHandler_ScenarioValidation 测试剧本验证
func TestScenarioHandler_ScenarioValidation(t *testing.T) {
	router, _, tempDir := setupScenarioTestRouter()
	defer os.RemoveAll(tempDir)

	// 获取有效的剧本应该成功
	req, _ := http.NewRequest("GET", "/api/scenarios/test-scenario", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	// 验证剧本包含所有必需组件
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	// 验证异常体档案
	assert.NotNil(t, data["anomaly"])

	// 验证晨会场景
	morningScenes, ok := data["morning_scenes"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(morningScenes), 0)

	// 验证简报
	assert.NotNil(t, data["briefing"])

	// 验证可选目标
	optionalGoals, ok := data["optional_goals"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(optionalGoals), 0)

	// 验证遭遇
	assert.NotNil(t, data["encounter"])

	// 验证余波
	assert.NotNil(t, data["aftermath"])

	// 验证奖励
	assert.NotNil(t, data["rewards"])
}

// TestScenarioHandler_CorruptedScenarioFile 测试损坏的剧本文件
func TestScenarioHandler_CorruptedScenarioFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时目录
	tempDir, _ := os.MkdirTemp("", "scenarios-corrupted-*")
	defer os.RemoveAll(tempDir)

	// 创建损坏的JSON文件
	corruptedFile := filepath.Join(tempDir, "corrupted.json")
	os.WriteFile(corruptedFile, []byte("{ invalid json }"), 0644)

	// 创建服务和路由
	scenarioService := service.NewScenarioService(tempDir)
	handler := NewScenarioHandler(scenarioService)

	router := gin.New()
	api := router.Group("/api")
	{
		scenarios := api.Group("/scenarios")
		{
			scenarios.GET("/:id", handler.GetScenario)
		}
	}

	// 尝试加载损坏的剧本
	req, _ := http.NewRequest("GET", "/api/scenarios/corrupted", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotEmpty(t, response["error"])
}

// TestScenarioHandler_EmptyScenarioDirectory 测试空剧本目录
func TestScenarioHandler_EmptyScenarioDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建空的临时目录
	tempDir, _ := os.MkdirTemp("", "scenarios-empty-*")
	defer os.RemoveAll(tempDir)

	// 创建服务和路由
	scenarioService := service.NewScenarioService(tempDir)
	handler := NewScenarioHandler(scenarioService)

	router := gin.New()
	api := router.Group("/api")
	{
		scenarios := api.Group("/scenarios")
		{
			scenarios.GET("", handler.ListScenarios)
		}
	}

	// 列出空目录中的剧本
	req, _ := http.NewRequest("GET", "/api/scenarios", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(data), "空目录应该返回空列表")
}
