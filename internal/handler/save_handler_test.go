package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

func setupSaveTestRouter() (*gin.Engine, *SaveHandler, service.GameService, service.AgentService) {
	gin.SetMode(gin.TestMode)

	// 创建服务
	agentService := service.NewAgentService()
	gameService := service.NewGameService()
	saveService := service.NewSaveService(gameService, agentService)

	// 创建处理器
	saveHandler := NewSaveHandler(saveService, gameService)

	// 创建路由
	router := gin.New()
	api := router.Group("/api")
	{
		saves := api.Group("/saves")
		{
			saves.POST("", saveHandler.CreateSave)
			saves.GET("", saveHandler.ListSaves)
			saves.GET("/:id", saveHandler.GetSave)
			saves.POST("/:id/load", saveHandler.LoadSave)
			saves.DELETE("/:id", saveHandler.DeleteSave)
		}
	}

	return router, saveHandler, gameService, agentService
}

// TestSaveHandler_CreateSave 测试存档保存
func TestSaveHandler_CreateSave(t *testing.T) {
	router, _, gameService, agentService := setupSaveTestRouter()

	// 先创建一个角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	require.NoError(t, err)

	// 创建一个游戏会话
	session, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "成功创建存档",
			request: map[string]interface{}{
				"session_id": session.ID,
				"name":       "测试存档1",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "缺少session_id",
			request: map[string]interface{}{
				"name": "测试存档2",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "缺少name",
			request: map[string]interface{}{
				"session_id": session.ID,
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "会话不存在",
			request: map[string]interface{}{
				"session_id": "non-existent-session",
				"name":       "测试存档3",
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/saves", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["id"])
				assert.Equal(t, tt.request["name"], data["name"])
				assert.Equal(t, tt.request["session_id"], data["session_id"])
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSaveHandler_ListSaves 测试存档列表
func TestSaveHandler_ListSaves(t *testing.T) {
	router, _, gameService, agentService := setupSaveTestRouter()

	// 创建角色和会话
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	require.NoError(t, err)

	session1, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	session2, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	// 创建几个存档
	saves := []map[string]interface{}{
		{
			"session_id": session1.ID,
			"name":       "存档1",
		},
		{
			"session_id": session1.ID,
			"name":       "存档2",
		},
		{
			"session_id": session2.ID,
			"name":       "存档3",
		},
	}

	for _, save := range saves {
		body, _ := json.Marshal(save)
		req, _ := http.NewRequest("POST", "/api/saves", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	tests := []struct {
		name          string
		sessionID     string
		expectedCount int
	}{
		{
			name:          "列出所有存档",
			sessionID:     "",
			expectedCount: 3,
		},
		{
			name:          "列出session1的存档",
			sessionID:     session1.ID,
			expectedCount: 2,
		},
		{
			name:          "列出session2的存档",
			sessionID:     session2.ID,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/saves"
			if tt.sessionID != "" {
				url += "?session_id=" + tt.sessionID
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response["success"].(bool))

			data, ok := response["data"].([]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.expectedCount, len(data))
		})
	}
}

// TestSaveHandler_GetSave 测试获取存档
func TestSaveHandler_GetSave(t *testing.T) {
	router, _, gameService, agentService := setupSaveTestRouter()

	// 创建角色和会话
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	require.NoError(t, err)

	session, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	// 创建一个存档
	createReq := map[string]interface{}{
		"session_id": session.ID,
		"name":       "测试存档",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/saves", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	saveData := createResponse["data"].(map[string]interface{})
	saveID := saveData["id"].(string)

	tests := []struct {
		name           string
		saveID         string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功获取存档",
			saveID:         saveID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "存档不存在",
			saveID:         "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/saves/"+tt.saveID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, tt.saveID, data["id"])
				assert.Equal(t, "测试存档", data["name"])
				assert.NotNil(t, data["snapshot"])
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSaveHandler_LoadSave 测试存档加载
func TestSaveHandler_LoadSave(t *testing.T) {
	router, _, gameService, agentService := setupSaveTestRouter()

	// 创建角色和会话
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	require.NoError(t, err)

	session, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	// 修改会话状态
	err = gameService.UpdateState(session.ID, func(state *domain.GameState) error {
		state.ChaosPool = 5
		state.LooseEnds = 3
		state.CollectedClues = []string{"clue1", "clue2"}
		return nil
	})
	require.NoError(t, err)

	// 创建存档
	createReq := map[string]interface{}{
		"session_id": session.ID,
		"name":       "测试存档",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/saves", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	saveData := createResponse["data"].(map[string]interface{})
	saveID := saveData["id"].(string)

	tests := []struct {
		name           string
		saveID         string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功加载存档",
			saveID:         saveID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "存档不存在",
			saveID:         "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/saves/"+tt.saveID+"/load", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if w.Code != tt.expectedStatus {
				t.Logf("Response body: %s", w.Body.String())
				t.Logf("Error: %v", response["error"])
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["session_id"])
				assert.Equal(t, "存档已加载", data["message"])

				// 验证加载的会话状态
				newSessionID := data["session_id"].(string)
				loadedSession, err := gameService.GetSession(newSessionID)
				require.NoError(t, err)
				assert.Equal(t, 5, loadedSession.State.ChaosPool)
				assert.Equal(t, 3, loadedSession.State.LooseEnds)
				assert.Equal(t, 2, len(loadedSession.State.CollectedClues))
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSaveHandler_DeleteSave 测试存档删除
func TestSaveHandler_DeleteSave(t *testing.T) {
	router, _, gameService, agentService := setupSaveTestRouter()

	// 创建角色和会话
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	require.NoError(t, err)

	session, err := gameService.CreateSession(agent.ID, "eternal-spring")
	require.NoError(t, err)

	// 创建存档
	createReq := map[string]interface{}{
		"session_id": session.ID,
		"name":       "待删除存档",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/saves", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	saveData := createResponse["data"].(map[string]interface{})
	saveID := saveData["id"].(string)

	tests := []struct {
		name           string
		saveID         string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功删除存档",
			saveID:         saveID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "删除不存在的存档",
			saveID:         "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/saves/"+tt.saveID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				assert.Equal(t, "存档已删除", response["message"])

				// 验证存档确实被删除了
				req, _ := http.NewRequest("GET", "/api/saves/"+tt.saveID, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusNotFound, w.Code)
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSaveHandler_ErrorResponses 测试错误响应
func TestSaveHandler_ErrorResponses(t *testing.T) {
	router, _, _, _ := setupSaveTestRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "无效的JSON",
			method:         "POST",
			path:           "/api/saves",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "获取不存在的存档",
			method:         "GET",
			path:           "/api/saves/invalid-id",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "加载不存在的存档",
			method:         "POST",
			path:           "/api/saves/invalid-id/load",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "删除不存在的存档",
			method:         "DELETE",
			path:           "/api/saves/invalid-id",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, _ = json.Marshal(tt.body)
				}
			}

			req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(reqBody))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
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
