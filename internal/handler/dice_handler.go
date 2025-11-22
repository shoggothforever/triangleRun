package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

type DiceHandler struct {
	diceService domain.DiceService
}

func NewDiceHandler(diceService domain.DiceService) *DiceHandler {
	return &DiceHandler{
		diceService: diceService,
	}
}

// RollDice 基础掷骰
func (h *DiceHandler) RollDice(c *gin.Context) {
	var req struct {
		Count int `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Count = 6 // 默认6颗骰子
	}

	result := h.diceService.Roll(req.Count)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
