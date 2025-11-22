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

func setupTestRouter() (*gin.Engine, *AgentHandler) {
	gin.SetMode(gin.TestMode)

	// 创建服务
	agentService := service.NewAgentService()

	// 创建处理器
	handler := NewAgentHandler(agentService)

	// 创建路由
	router := gin.New()
	api := router.Group("/api")
	{
		agents := api.Group("/agents")
		{
			agents.POST("", handler.CreateAgent)
			agents.GET("", handler.ListAgents)
			agents.GET("/:id", handler.GetAgent)
			agents.PUT("/:id", handler.UpdateAgent)
			agents.DELETE("/:id", handler.DeleteAgent)
		}
	}

	return router, handler
}

// TestAgentHandler_CreateAgent 测试创建角色
func TestAgentHandler_CreateAgent(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		request        service.CreateAgentRequest
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "成功创建角色",
			request: service.CreateAgentRequest{
				Name:        "测试特工",
				Pronouns:    "他/他的",
				AnomalyType: domain.AnomalyWhisper,
				RealityType: domain.RealityCaretaker,
				CareerType:  domain.CareerPublicRelations,
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "缺少必需字段",
			request: service.CreateAgentRequest{
				Name: "测试特工",
				// 缺少AnomalyType, RealityType, CareerType
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "无效的异常体类型",
			request: service.CreateAgentRequest{
				Name:        "测试特工",
				AnomalyType: "invalid-anomaly",
				RealityType: domain.RealityCaretaker,
				CareerType:  domain.CareerPublicRelations,
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
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
				// 验证返回的角色数据
				data, ok := response["data"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["id"])
				assert.Equal(t, tt.request.Name, data["name"])
				assert.Equal(t, tt.request.AnomalyType, data["anomaly"].(map[string]interface{})["type"])
			} else {
				// 验证错误信息
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

// TestAgentHandler_GetAgent 测试获取角色
func TestAgentHandler_GetAgent(t *testing.T) {
	router, _ := setupTestRouter()

	// 先创建一个角色
	createReq := service.CreateAgentRequest{
		Name:        "测试特工",
		Pronouns:    "他/他的",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentData := createResponse["data"].(map[string]interface{})
	agentID := agentData["id"].(string)

	tests := []struct {
		name           string
		agentID        string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功获取角色",
			agentID:        agentID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "角色不存在",
			agentID:        "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/agents/"+tt.agentID, nil)
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
				assert.Equal(t, tt.agentID, data["id"])
				assert.Equal(t, "测试特工", data["name"])
			}
		})
	}
}

// TestAgentHandler_UpdateAgent 测试更新角色
func TestAgentHandler_UpdateAgent(t *testing.T) {
	router, _ := setupTestRouter()

	// 先创建一个角色
	createReq := service.CreateAgentRequest{
		Name:        "测试特工",
		Pronouns:    "他/他的",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentData := createResponse["data"].(map[string]interface{})
	agentID := agentData["id"].(string)

	tests := []struct {
		name           string
		agentID        string
		updateData     map[string]interface{}
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:    "成功更新角色名称",
			agentID: agentID,
			updateData: map[string]interface{}{
				"name": "更新后的特工",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:    "更新多个字段",
			agentID: agentID,
			updateData: map[string]interface{}{
				"name":          "再次更新",
				"pronouns":      "她/她的",
				"commendations": 5,
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:    "更新不存在的角色",
			agentID: "non-existent-id",
			updateData: map[string]interface{}{
				"name": "不存在",
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updateData)
			req, _ := http.NewRequest("PUT", "/api/agents/"+tt.agentID, bytes.NewBuffer(body))
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

				// 验证更新的字段
				if name, ok := tt.updateData["name"].(string); ok {
					assert.Equal(t, name, data["name"])
				}
				if pronouns, ok := tt.updateData["pronouns"].(string); ok {
					assert.Equal(t, pronouns, data["pronouns"])
				}
			}
		})
	}
}

// TestAgentHandler_DeleteAgent 测试删除角色
func TestAgentHandler_DeleteAgent(t *testing.T) {
	router, _ := setupTestRouter()

	// 先创建一个角色
	createReq := service.CreateAgentRequest{
		Name:        "待删除特工",
		Pronouns:    "他/他的",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentData := createResponse["data"].(map[string]interface{})
	agentID := agentData["id"].(string)

	tests := []struct {
		name           string
		agentID        string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "成功删除角色",
			agentID:        agentID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "删除不存在的角色",
			agentID:        "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/agents/"+tt.agentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, response["success"])

			if tt.expectSuccess {
				// 验证角色确实被删除了
				req, _ := http.NewRequest("GET", "/api/agents/"+tt.agentID, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		})
	}
}

// TestAgentHandler_ListAgents 测试列出所有角色
func TestAgentHandler_ListAgents(t *testing.T) {
	router, _ := setupTestRouter()

	// 创建几个角色
	agents := []service.CreateAgentRequest{
		{
			Name:        "特工1",
			AnomalyType: domain.AnomalyWhisper,
			RealityType: domain.RealityCaretaker,
			CareerType:  domain.CareerPublicRelations,
		},
		{
			Name:        "特工2",
			AnomalyType: domain.AnomalyCatalog,
			RealityType: domain.RealityScheduleOverload,
			CareerType:  domain.CareerRD,
		},
	}

	for _, agent := range agents {
		body, _ := json.Marshal(agent)
		req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// 列出所有角色
	req, _ := http.NewRequest("GET", "/api/agents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 2, "应该至少有2个角色")
}

// TestAgentHandler_CreateAgentWithRelationships 测试创建带人际关系的角色
func TestAgentHandler_CreateAgentWithRelationships(t *testing.T) {
	router, _ := setupTestRouter()

	createReq := service.CreateAgentRequest{
		Name:        "有关系的特工",
		Pronouns:    "他/他的",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{
				Name:        "母亲",
				Description: "我的母亲",
				Connection:  6,
			},
			{
				Name:        "朋友",
				Description: "我的好友",
				Connection:  3,
			},
			{
				Name:        "同事",
				Description: "工作伙伴",
				Connection:  3,
			},
		},
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	relationships, ok := data["relationships"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(relationships), "应该有3段人际关系")

	// 验证连结点数总和
	totalConnection := 0
	for _, rel := range relationships {
		relMap := rel.(map[string]interface{})
		totalConnection += int(relMap["connection"].(float64))
	}
	assert.Equal(t, 12, totalConnection, "连结点数总和应该是12")
}

// TestAgentHandler_ErrorResponses 测试错误响应格式
func TestAgentHandler_ErrorResponses(t *testing.T) {
	router, _ := setupTestRouter()

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
			path:           "/api/agents",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "获取不存在的角色",
			method:         "GET",
			path:           "/api/agents/invalid-id",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "更新不存在的角色",
			method:         "PUT",
			path:           "/api/agents/invalid-id",
			body:           map[string]string{"name": "test"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "删除不存在的角色",
			method:         "DELETE",
			path:           "/api/agents/invalid-id",
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
