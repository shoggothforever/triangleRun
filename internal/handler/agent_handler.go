package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/service"
)

type AgentHandler struct {
	agentService service.AgentService
}

func NewAgentHandler(agentService service.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req service.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	agent, err := h.agentService.CreateAgent(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    agent,
	})
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := h.agentService.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agent,
	})
}

func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents, err := h.agentService.ListAgents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agents,
	})
}
