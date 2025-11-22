package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

type ScenarioHandler struct {
	scenarioService service.ScenarioService
}

func NewScenarioHandler(scenarioService service.ScenarioService) *ScenarioHandler {
	return &ScenarioHandler{
		scenarioService: scenarioService,
	}
}

// ListScenarios 列出所有剧本 GET /api/scenarios
func (h *ScenarioHandler) ListScenarios(c *gin.Context) {
	summaries, err := h.scenarioService.ListScenarios()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summaries,
	})
}

// GetScenario 获取剧本详情 GET /api/scenarios/:id
func (h *ScenarioHandler) GetScenario(c *gin.Context) {
	scenarioID := c.Param("id")

	if scenarioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "剧本ID不能为空",
		})
		return
	}

	scenario, err := h.scenarioService.LoadScenario(scenarioID)
	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			if gameErr.Code == domain.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			if gameErr.Code == domain.ErrDataCorrupted {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
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
		"data":    scenario,
	})
}

// GetScene 获取场景 GET /api/scenarios/:id/scenes/:sceneId
func (h *ScenarioHandler) GetScene(c *gin.Context) {
	scenarioID := c.Param("id")
	sceneID := c.Param("sceneId")

	if scenarioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "剧本ID不能为空",
		})
		return
	}

	if sceneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "场景ID不能为空",
		})
		return
	}

	scene, err := h.scenarioService.GetScene(scenarioID, sceneID)
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
		"data":    scene,
	})
}
