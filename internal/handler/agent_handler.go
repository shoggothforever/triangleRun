package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
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

// CreateAgent 创建角色 POST /api/agents
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req service.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	agent, err := h.agentService.CreateAgent(&req)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrInvalidARC, domain.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{
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
		"data":    agent,
	})
}

// GetAgent 获取角色 GET /api/agents/:id
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := h.agentService.GetAgent(agentID)
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
		"data":    agent,
	})
}

// UpdateAgent 更新角色 PUT /api/agents/:id
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	agentID := c.Param("id")

	// 先获取现有角色
	agent, err := h.agentService.GetAgent(agentID)
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

	// 绑定更新请求
	var updateReq struct {
		Name          *string                `json:"name"`
		Pronouns      *string                `json:"pronouns"`
		Relationships []*domain.Relationship `json:"relationships"`
		QA            map[string]int         `json:"qa"`
		Commendations *int                   `json:"commendations"`
		Reprimands    *int                   `json:"reprimands"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 更新字段
	if updateReq.Name != nil {
		agent.Name = *updateReq.Name
	}
	if updateReq.Pronouns != nil {
		agent.Pronouns = *updateReq.Pronouns
	}
	if updateReq.Relationships != nil {
		agent.Relationships = updateReq.Relationships
	}
	if updateReq.QA != nil {
		agent.QA = updateReq.QA
	}
	if updateReq.Commendations != nil {
		agent.Commendations = *updateReq.Commendations
	}
	if updateReq.Reprimands != nil {
		agent.Reprimands = *updateReq.Reprimands
	}

	// 保存更新
	if err := h.agentService.UpdateAgent(agent); err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrInvalidARC, domain.ErrInvalidInput:
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agent,
	})
}

// DeleteAgent 删除角色 DELETE /api/agents/:id
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.agentService.DeleteAgent(agentID); err != nil {
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
		"message": "角色已删除",
	})
}

// ListAgents 列出所有角色 GET /api/agents
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
