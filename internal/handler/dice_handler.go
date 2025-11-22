package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

type DiceHandler struct {
	diceService  domain.DiceService
	agentService service.AgentService
}

func NewDiceHandler(diceService domain.DiceService, agentService service.AgentService) *DiceHandler {
	return &DiceHandler{
		diceService:  diceService,
		agentService: agentService,
	}
}

// RollDice 基础掷骰 POST /api/dice/roll
func (h *DiceHandler) RollDice(c *gin.Context) {
	var req struct {
		Count int `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Count = 6 // 默认6颗骰子
	}

	if req.Count <= 0 {
		req.Count = 6
	}

	result := h.diceService.Roll(req.Count)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RollForAbility 异常能力掷骰 POST /api/dice/ability
func (h *DiceHandler) RollForAbility(c *gin.Context) {
	var req struct {
		AgentID   string `json:"agent_id" binding:"required"`
		AbilityID string `json:"ability_id" binding:"required"`
		QASpend   int    `json:"qa_spend"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 获取角色
	agent, err := h.agentService.GetAgent(req.AgentID)
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

	// 查找能力
	var ability *domain.AnomalyAbility
	for _, ab := range agent.Anomaly.Abilities {
		if ab.ID == req.AbilityID {
			ability = ab
			break
		}
	}

	if ability == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "能力不存在",
		})
		return
	}

	// 执行掷骰
	result := h.diceService.RollForAbility(agent, ability)

	// 应用QA调整
	if req.QASpend > 0 && ability.Roll != nil {
		quality := ability.Roll.Quality
		// 检查QA是否足够
		if agent.QA[quality] < req.QASpend {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "资质保证不足",
				"details": gin.H{
					"quality":   quality,
					"available": agent.QA[quality],
					"required":  req.QASpend,
				},
			})
			return
		}

		// 应用QA
		result = h.diceService.ApplyQA(result, quality, req.QASpend)

		// 扣除QA
		if err := h.agentService.SpendQA(req.AgentID, quality, req.QASpend); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "扣除资质保证失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roll":         result,
			"ability":      ability,
			"qa_spent":     req.QASpend,
			"qa_remaining": agent.QA,
		},
	})
}

// RollForRequest 请求机构掷骰 POST /api/dice/request
func (h *DiceHandler) RollForRequest(c *gin.Context) {
	var req struct {
		AgentID     string `json:"agent_id" binding:"required"`
		Quality     string `json:"quality" binding:"required"`
		Effect      string `json:"effect" binding:"required"`
		CausalChain string `json:"causal_chain" binding:"required"`
		QASpend     int    `json:"qa_spend"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证资质类型
	validQualities := []string{
		domain.QualityFocus,
		domain.QualityEmpathy,
		domain.QualityPresence,
		domain.QualityDeception,
		domain.QualityInitiative,
		domain.QualityProfession,
		domain.QualityVitality,
		domain.QualityGrit,
		domain.QualitySubtlety,
	}

	validQuality := false
	for _, q := range validQualities {
		if q == req.Quality {
			validQuality = true
			break
		}
	}

	if !validQuality {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的资质类型: " + req.Quality,
		})
		return
	}

	// 获取角色
	agent, err := h.agentService.GetAgent(req.AgentID)
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

	// 执行掷骰
	result := h.diceService.RollForQuality(agent, req.Quality)

	// 应用QA调整
	if req.QASpend > 0 {
		// 检查QA是否足够
		if agent.QA[req.Quality] < req.QASpend {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "资质保证不足",
				"details": gin.H{
					"quality":   req.Quality,
					"available": agent.QA[req.Quality],
					"required":  req.QASpend,
				},
			})
			return
		}

		// 应用QA
		result = h.diceService.ApplyQA(result, req.Quality, req.QASpend)

		// 扣除QA
		if err := h.agentService.SpendQA(req.AgentID, req.Quality, req.QASpend); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "扣除资质保证失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roll":         result,
			"quality":      req.Quality,
			"effect":       req.Effect,
			"causal_chain": req.CausalChain,
			"qa_spent":     req.QASpend,
			"qa_remaining": agent.QA,
		},
	})
}
