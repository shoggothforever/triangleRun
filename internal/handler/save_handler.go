package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/service"
)

type SaveHandler struct {
	saveService service.SaveService
	gameService service.GameService
}

func NewSaveHandler(saveService service.SaveService, gameService service.GameService) *SaveHandler {
	return &SaveHandler{
		saveService: saveService,
		gameService: gameService,
	}
}

// CreateSave 保存游戏 POST /api/saves
func (h *SaveHandler) CreateSave(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
		})
		return
	}

	// 创建存档
	snapshot, err := h.saveService.CreateSave(req.SessionID, req.Name)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			case domain.ErrInvalidInput:
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
		"data":    snapshot,
	})
}

// ListSaves 列出存档 GET /api/saves
func (h *SaveHandler) ListSaves(c *gin.Context) {
	// 可选的session_id查询参数
	sessionID := c.Query("session_id")

	metadata, err := h.saveService.ListSaves(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metadata,
	})
}

// GetSave 获取存档 GET /api/saves/:id
func (h *SaveHandler) GetSave(c *gin.Context) {
	saveID := c.Param("id")

	snapshot, err := h.saveService.GetSave(saveID)
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
		"data":    snapshot,
	})
}

// LoadSave 加载存档 POST /api/saves/:id/load
func (h *SaveHandler) LoadSave(c *gin.Context) {
	saveID := c.Param("id")

	// 加载存档（这会创建一个新的会话ID）
	session, err := h.saveService.LoadSave(saveID)
	if err != nil {
		if gameErr, ok := err.(*domain.GameError); ok {
			switch gameErr.Code {
			case domain.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			case domain.ErrDataCorrupted:
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

	// 将加载的会话注册到游戏服务
	if err := h.gameService.RegisterSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "注册会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session_id": session.ID,
			"session":    session,
			"message":    "存档已加载",
		},
	})
}

// DeleteSave 删除存档 DELETE /api/saves/:id
func (h *SaveHandler) DeleteSave(c *gin.Context) {
	saveID := c.Param("id")

	if err := h.saveService.DeleteSave(saveID); err != nil {
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
		"message": "存档已删除",
	})
}
