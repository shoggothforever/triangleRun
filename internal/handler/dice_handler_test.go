package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

func setupDiceTestRouter() (*gin.Engine, *DiceHandler, service.AgentService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	diceService := domain.NewDiceService()
	agentService := service.NewAgentService()
	diceHandler := NewDiceHandler(diceService, agentService)

	api := router.Group("/api/dice")
	{
		api.POST("/roll", diceHandler.RollDice)
		api.POST("/ability", diceHandler.RollForAbility)
		api.POST("/request", diceHandler.RollForRequest)
	}

	return router, diceHandler, agentService
}

// TestRollDice_BasicRoll 测试基础掷骰
func TestRollDice_BasicRoll(t *testing.T) {
	router, _, _ := setupDiceTestRouter()

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name:       "默认6颗骰子",
			request:    map[string]interface{}{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "指定4颗骰子",
			request:    map[string]interface{}{"count": 4},
			wantStatus: http.StatusOK,
		},
		{
			name:       "指定10颗骰子",
			request:    map[string]interface{}{"count": 10},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/dice/roll", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response["success"].(bool))

			data := response["data"].(map[string]interface{})
			dice := data["dice"].([]interface{})

			// 验证骰子数量
			expectedCount := 6
			if count, ok := tt.request["count"].(int); ok && count > 0 {
				expectedCount = count
			}
			assert.Equal(t, expectedCount, len(dice))

			// 验证骰子范围（1-4）
			for _, d := range dice {
				dieValue := int(d.(float64))
				assert.GreaterOrEqual(t, dieValue, 1)
				assert.LessOrEqual(t, dieValue, 4)
			}

			// 验证结果字段存在
			assert.Contains(t, data, "threes")
			assert.Contains(t, data, "success")
			assert.Contains(t, data, "chaos")
		})
	}
}

// TestRollForAbility_Success 测试能力掷骰成功
func TestRollForAbility_Success(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Greater(t, len(agent.Anomaly.Abilities), 0)

	abilityID := agent.Anomaly.Abilities[0].ID

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name: "不使用QA",
			request: map[string]interface{}{
				"agent_id":   agent.ID,
				"ability_id": abilityID,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "使用1点QA",
			request: map[string]interface{}{
				"agent_id":   agent.ID,
				"ability_id": abilityID,
				"qa_spend":   1,
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/dice/ability", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response["success"].(bool))

			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "roll")
			assert.Contains(t, data, "ability")
			assert.Contains(t, data, "qa_spent")
			assert.Contains(t, data, "qa_remaining")
		})
	}
}

// TestRollForAbility_InvalidAgent 测试能力掷骰 - 无效角色
func TestRollForAbility_InvalidAgent(t *testing.T) {
	router, _, _ := setupDiceTestRouter()

	request := map[string]interface{}{
		"agent_id":   "invalid-id",
		"ability_id": "some-ability",
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/dice/ability", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response, "error")
}

// TestRollForAbility_InvalidAbility 测试能力掷骰 - 无效能力
func TestRollForAbility_InvalidAbility(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)

	request := map[string]interface{}{
		"agent_id":   agent.ID,
		"ability_id": "invalid-ability-id",
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/dice/ability", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "能力不存在")
}

// TestRollForAbility_InsufficientQA 测试能力掷骰 - QA不足
func TestRollForAbility_InsufficientQA(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)

	abilityID := agent.Anomaly.Abilities[0].ID

	request := map[string]interface{}{
		"agent_id":   agent.ID,
		"ability_id": abilityID,
		"qa_spend":   100, // 远超可用QA
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/dice/ability", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "资质保证不足")
}

// TestRollForRequest_Success 测试请求机构掷骰成功
func TestRollForRequest_Success(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name: "基础请求机构",
			request: map[string]interface{}{
				"agent_id":     agent.ID,
				"quality":      domain.QualityFocus,
				"effect":       "我在门口发现了一把钥匙",
				"causal_chain": "因为我之前看到有人掉了东西",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "使用QA的请求机构",
			request: map[string]interface{}{
				"agent_id":     agent.ID,
				"quality":      domain.QualityDeception,
				"effect":       "守卫相信了我的谎言",
				"causal_chain": "因为我伪造了一份文件",
				"qa_spend":     1,
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/dice/request", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response["success"].(bool))

			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "roll")
			assert.Contains(t, data, "quality")
			assert.Contains(t, data, "effect")
			assert.Contains(t, data, "causal_chain")
			assert.Contains(t, data, "qa_spent")
			assert.Contains(t, data, "qa_remaining")
		})
	}
}

// TestRollForRequest_InvalidQuality 测试请求机构掷骰 - 无效资质
func TestRollForRequest_InvalidQuality(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)

	request := map[string]interface{}{
		"agent_id":     agent.ID,
		"quality":      "invalid_quality",
		"effect":       "某个效果",
		"causal_chain": "某个因果链",
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/dice/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "无效的资质类型")
}

// TestRollForRequest_MissingFields 测试请求机构掷骰 - 缺少必需字段
func TestRollForRequest_MissingFields(t *testing.T) {
	router, _, _ := setupDiceTestRouter()

	tests := []struct {
		name    string
		request map[string]interface{}
	}{
		{
			name: "缺少agent_id",
			request: map[string]interface{}{
				"quality":      domain.QualityFocus,
				"effect":       "某个效果",
				"causal_chain": "某个因果链",
			},
		},
		{
			name: "缺少quality",
			request: map[string]interface{}{
				"agent_id":     "some-id",
				"effect":       "某个效果",
				"causal_chain": "某个因果链",
			},
		},
		{
			name: "缺少effect",
			request: map[string]interface{}{
				"agent_id":     "some-id",
				"quality":      domain.QualityFocus,
				"causal_chain": "某个因果链",
			},
		},
		{
			name: "缺少causal_chain",
			request: map[string]interface{}{
				"agent_id": "some-id",
				"quality":  domain.QualityFocus,
				"effect":   "某个效果",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/dice/request", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response["success"].(bool))
		})
	}
}

// TestRollForRequest_InsufficientQA 测试请求机构掷骰 - QA不足
func TestRollForRequest_InsufficientQA(t *testing.T) {
	router, _, agentService := setupDiceTestRouter()

	// 创建测试角色
	agent, err := agentService.CreateAgent(&service.CreateAgentRequest{
		Name:        "测试角色",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
	})
	assert.NoError(t, err)

	request := map[string]interface{}{
		"agent_id":     agent.ID,
		"quality":      domain.QualityFocus,
		"effect":       "某个效果",
		"causal_chain": "某个因果链",
		"qa_spend":     100, // 远超可用QA
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/dice/request", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "资质保证不足")
}
