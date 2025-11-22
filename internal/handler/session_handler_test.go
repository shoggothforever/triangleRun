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

func setupSessionTestRouter() (*gin.Engine, *SessionHandler, service.GameService) {
	gin.SetMode(gin.TestMode)

	// 创建服务
	gameService := service.NewGameService()

	// 创建处理器
	handler := NewSessionHandler(gameService)

	// 创建路由
	router := gin.New()
	api := router.Group("/api")
	{
		sessions := api.Group("/sessions")
		{
			sessions.POST("", handler.CreateSession)
			sessions.GET("/:id", handler.GetSession)
			sessions.POST("/:id/actions", handler.ExecuteAction)
			sessions.POST("/:id/phase", handler.TransitionPhase)
		}
	}

	return router, handler, gameService
}

// TestSessionHandler_CreateSession 测试创建会话
func TestSessionHandler_CreateSession(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	tests := []struct {
		name           string
		request        map[string]string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "成功创建会话",
			request: map[string]string{
				"agent_id":    "test-agent-id",
				"scenario_id": "eternal-spring",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "缺少agent_id",
			request: map[string]string{
				"scenario_id": "eternal-spring",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "缺少scenario_id",
			request: map[string]string{
				"agent_id": "test-agent-id",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:           "空请求体",
			request:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				// 验证返回的会话数据
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["id"])
				assert.Equal(t, tt.request["agent_id"], data["agent_id"])
				assert.Equal(t, tt.request["scenario_id"], data["scenario_id"])
				assert.Equal(t, string(domain.PhaseMorning), data["phase"])
				assert.NotNil(t, data["state"])
			} else {
				// 验证错误信息
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSessionHandler_GetSession 测试获取会话
func TestSessionHandler_GetSession(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 先创建一个会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功获取会话",
			sessionID:      sessionID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "会话不存在",
			sessionID:      "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/sessions/"+tt.sessionID, nil)
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
				assert.Equal(t, tt.sessionID, data["id"])
				assert.Equal(t, "test-agent-id", data["agent_id"])
				assert.Equal(t, "eternal-spring", data["scenario_id"])
			}
		})
	}
}

// TestSessionHandler_ExecuteAction 测试执行行动
func TestSessionHandler_ExecuteAction(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 先创建一个会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	tests := []struct {
		name           string
		sessionID      string
		action         map[string]interface{}
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:      "移动到场景",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "move_to_scene",
				"target":      "scene-001",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:      "收集线索",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "collect_clue",
				"target":      "clue-001",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:      "解锁地点",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "unlock_location",
				"target":      "location-001",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:      "添加混沌",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "add_chaos",
				"parameters": map[string]interface{}{
					"amount": 3.0,
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:      "更新NPC状态",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "update_npc_state",
				"target":      "npc-001",
				"parameters": map[string]interface{}{
					"status":            "influenced",
					"anomaly_influence": true,
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:      "缺少action_type",
			sessionID: sessionID,
			action: map[string]interface{}{
				"target": "scene-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:      "未知的行动类型",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "unknown_action",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:      "会话不存在",
			sessionID: "non-existent-id",
			action: map[string]interface{}{
				"action_type": "move_to_scene",
				"target":      "scene-001",
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:      "移动到场景但缺少目标",
			sessionID: sessionID,
			action: map[string]interface{}{
				"action_type": "move_to_scene",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.action)
			req, _ := http.NewRequest("POST", "/api/sessions/"+tt.sessionID+"/actions", bytes.NewBuffer(body))
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
				assert.NotEmpty(t, data["action"])
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSessionHandler_TransitionPhase 测试阶段转换
func TestSessionHandler_TransitionPhase(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 先创建一个会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	tests := []struct {
		name           string
		sessionID      string
		phaseReq       map[string]string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:      "成功转换到调查阶段",
			sessionID: sessionID,
			phaseReq: map[string]string{
				"phase": string(domain.PhaseInvestigation),
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "缺少phase字段",
			sessionID:      sessionID,
			phaseReq:       map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:      "无效的阶段值",
			sessionID: sessionID,
			phaseReq: map[string]string{
				"phase": "invalid-phase",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:      "会话不存在",
			sessionID: "non-existent-id",
			phaseReq: map[string]string{
				"phase": string(domain.PhaseInvestigation),
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.phaseReq)
			req, _ := http.NewRequest("POST", "/api/sessions/"+tt.sessionID+"/phase", bytes.NewBuffer(body))
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
				assert.Equal(t, tt.sessionID, data["session_id"])
				assert.NotEmpty(t, data["phase"])
				assert.NotEmpty(t, data["message"])
			} else {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestSessionHandler_PhaseTransitionFlow 测试完整的阶段转换流程
func TestSessionHandler_PhaseTransitionFlow(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 创建会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	// 验证初始阶段是晨会
	assert.Equal(t, string(domain.PhaseMorning), sessionData["phase"])

	// 转换到调查阶段
	phaseReq := map[string]string{"phase": string(domain.PhaseInvestigation)}
	body, _ = json.Marshal(phaseReq)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/phase", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var phaseResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &phaseResponse)
	assert.True(t, phaseResponse["success"].(bool))

	// 验证阶段已更新
	req, _ = http.NewRequest("GET", "/api/sessions/"+sessionID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResponse)
	sessionData = getResponse["data"].(map[string]interface{})
	assert.Equal(t, string(domain.PhaseInvestigation), sessionData["phase"])

	// 转换到遭遇阶段
	phaseReq = map[string]string{"phase": string(domain.PhaseEncounter)}
	body, _ = json.Marshal(phaseReq)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/phase", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证阶段已更新
	req, _ = http.NewRequest("GET", "/api/sessions/"+sessionID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &getResponse)
	sessionData = getResponse["data"].(map[string]interface{})
	assert.Equal(t, string(domain.PhaseEncounter), sessionData["phase"])
}

// TestSessionHandler_ActionSequence 测试行动序列
func TestSessionHandler_ActionSequence(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 创建会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	// 执行一系列行动
	actions := []map[string]interface{}{
		{
			"action_type": "move_to_scene",
			"target":      "scene-001",
		},
		{
			"action_type": "collect_clue",
			"target":      "clue-001",
		},
		{
			"action_type": "collect_clue",
			"target":      "clue-002",
		},
		{
			"action_type": "unlock_location",
			"target":      "location-001",
		},
		{
			"action_type": "add_chaos",
			"parameters": map[string]interface{}{
				"amount": 2.0,
			},
		},
	}

	for i, action := range actions {
		body, _ := json.Marshal(action)
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/actions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "行动 %d 应该成功", i+1)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))
	}

	// 验证会话状态
	req, _ = http.NewRequest("GET", "/api/sessions/"+sessionID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &getResponse)
	sessionData = getResponse["data"].(map[string]interface{})
	state := sessionData["state"].(map[string]interface{})

	// 验证状态更新
	assert.Equal(t, "scene-001", state["current_scene_id"])

	collectedClues := state["collected_clues"].([]interface{})
	assert.Equal(t, 2, len(collectedClues), "应该收集了2个线索")

	unlockedLocations := state["unlocked_locations"].([]interface{})
	assert.Equal(t, 1, len(unlockedLocations), "应该解锁了1个地点")

	chaosPool := int(state["chaos_pool"].(float64))
	assert.Equal(t, 2, chaosPool, "混沌池应该有2点混沌")
}

// TestSessionHandler_DuplicateClueCollection 测试重复收集线索
func TestSessionHandler_DuplicateClueCollection(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

	// 创建会话
	createReq := map[string]string{
		"agent_id":    "test-agent-id",
		"scenario_id": "eternal-spring",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	sessionData := createResponse["data"].(map[string]interface{})
	sessionID := sessionData["id"].(string)

	// 第一次收集线索
	action := map[string]interface{}{
		"action_type": "collect_clue",
		"target":      "clue-001",
	}
	body, _ = json.Marshal(action)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/actions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 第二次收集同一个线索
	body, _ = json.Marshal(action)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/actions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"].(string), "线索已收集")
}

// TestSessionHandler_ErrorResponses 测试错误响应格式
func TestSessionHandler_ErrorResponses(t *testing.T) {
	router, _, _ := setupSessionTestRouter()

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
			path:           "/api/sessions",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "获取不存在的会话",
			method:         "GET",
			path:           "/api/sessions/invalid-id",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "执行行动时会话不存在",
			method: "POST",
			path:   "/api/sessions/invalid-id/actions",
			body: map[string]string{
				"action_type": "move_to_scene",
				"target":      "scene-001",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "转换阶段时会话不存在",
			method: "POST",
			path:   "/api/sessions/invalid-id/phase",
			body: map[string]string{
				"phase": string(domain.PhaseInvestigation),
			},
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
