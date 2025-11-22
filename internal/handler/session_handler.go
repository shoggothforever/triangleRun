package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

type SessionHandler struct {
	gameService service.GameService
}

func NewSessionHandler(gameService service.GameService) *SessionHandler {
	return &SessionHandler{
		gameService: gameService,
	}
}

// CreateSession 创建游戏会话 POST /api/sessions
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req struct {
		AgentID    string `json:"agent_id" binding:"required"`
		ScenarioID string `json:"scenario_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	session, err := h.gameService.CreateSession(req.AgentID, req.ScenarioID)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrInvalidInput, domain.ErrInvalidARC:
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			case domain.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    session,
	})
}

// GetSession 获取游戏会话 GET /api/sessions/:id
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := h.gameService.GetSession(sessionID)
	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			if gameErr.Code == domain.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

// ExecuteAction 执行行动 POST /api/sessions/:id/actions
func (h *SessionHandler) ExecuteAction(c *gin.Context) {
	sessionID := c.Param("id")

	var req struct {
		ActionType string                 `json:"action_type" binding:"required"`
		Target     string                 `json:"target"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 获取会话
	session, err := h.gameService.GetSession(sessionID)
	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			if gameErr.Code == domain.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 根据行动类型执行不同的逻辑
	var result interface{}
	switch req.ActionType {
	case "move_to_scene":
		// 移动到场景
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "目标场景ID不能为空",
			})
			return
		}
		err = h.gameService.UpdateState(sessionID, func(state *domain.GameState) error {
			state.CurrentSceneID = req.Target
			state.VisitedScenes[req.Target] = true
			return nil
		})
		result = gin.H{
			"action":   "move_to_scene",
			"scene_id": req.Target,
			"message":  "已移动到场景: " + req.Target,
		}

	case "collect_clue":
		// 收集线索
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "线索ID不能为空",
			})
			return
		}
		err = h.gameService.UpdateState(sessionID, func(state *domain.GameState) error {
			// 检查是否已收集
			for _, clue := range state.CollectedClues {
				if clue == req.Target {
					return domain.NewGameError(domain.ErrInvalidAction, "线索已收集")
				}
			}
			state.CollectedClues = append(state.CollectedClues, req.Target)
			return nil
		})
		result = gin.H{
			"action":  "collect_clue",
			"clue_id": req.Target,
			"message": "已收集线索: " + req.Target,
		}

	case "unlock_location":
		// 解锁地点
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "地点ID不能为空",
			})
			return
		}
		err = h.gameService.UpdateState(sessionID, func(state *domain.GameState) error {
			// 检查是否已解锁
			for _, loc := range state.UnlockedLocations {
				if loc == req.Target {
					return domain.NewGameError(domain.ErrInvalidAction, "地点已解锁")
				}
			}
			state.UnlockedLocations = append(state.UnlockedLocations, req.Target)
			return nil
		})
		result = gin.H{
			"action":      "unlock_location",
			"location_id": req.Target,
			"message":     "已解锁地点: " + req.Target,
		}

	case "add_chaos":
		// 添加混沌
		amount, ok := req.Parameters["amount"].(float64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "混沌数量参数无效",
			})
			return
		}
		err = h.gameService.UpdateState(sessionID, func(state *domain.GameState) error {
			state.ChaosPool += int(amount)
			return nil
		})
		result = gin.H{
			"action":     "add_chaos",
			"amount":     int(amount),
			"chaos_pool": session.State.ChaosPool + int(amount),
			"message":    "已添加混沌",
		}

	case "update_npc_state":
		// 更新NPC状态
		if req.Target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "NPC ID不能为空",
			})
			return
		}
		err = h.gameService.UpdateState(sessionID, func(state *domain.GameState) error {
			if state.NPCStates == nil {
				state.NPCStates = make(map[string]*domain.NPCState)
			}
			// 创建或更新NPC状态
			npcState := &domain.NPCState{
				ID:              req.Target,
				CurrentState:    "",
				AnomalyAffected: false,
				Relationship:    0,
				CustomData:      make(map[string]interface{}),
			}
			if status, ok := req.Parameters["status"].(string); ok {
				npcState.CurrentState = status
			}
			if influenced, ok := req.Parameters["anomaly_affected"].(bool); ok {
				npcState.AnomalyAffected = influenced
			}
			if relationship, ok := req.Parameters["relationship"].(float64); ok {
				npcState.Relationship = int(relationship)
			}
			state.NPCStates[req.Target] = npcState
			return nil
		})
		result = gin.H{
			"action":  "update_npc_state",
			"npc_id":  req.Target,
			"message": "已更新NPC状态",
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "未知的行动类型: " + req.ActionType,
		})
		return
	}

	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrInvalidAction, domain.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			case domain.ErrInvalidPhase:
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// TransitionPhase 转换阶段 POST /api/sessions/:id/phase
func (h *SessionHandler) TransitionPhase(c *gin.Context) {
	sessionID := c.Param("id")

	var req struct {
		Phase string `json:"phase" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证阶段值
	phase := domain.GamePhase(req.Phase)
	if !isValidPhase(phase) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的阶段: " + req.Phase,
		})
		return
	}

	// 执行阶段转换
	err := h.gameService.TransitionPhase(sessionID, phase)
	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			case domain.ErrInvalidPhase:
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 获取更新后的会话
	session, err := h.gameService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session_id": sessionID,
			"phase":      session.Phase,
			"message":    "阶段已转换为: " + string(phase),
		},
	})
}

// isValidPhase 验证阶段是否有效
func isValidPhase(phase domain.GamePhase) bool {
	validPhases := []domain.GamePhase{
		domain.PhaseMorning,
		domain.PhaseInvestigation,
		domain.PhaseEncounter,
		domain.PhaseAftermath,
	}

	for _, p := range validPhases {
		if p == phase {
			return true
		}
	}

	return false
}
