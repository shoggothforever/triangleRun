package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/handler"
	"github.com/trpg-solo-engine/backend/internal/service"
)

func main() {
	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// 初始化服务（不需要数据库）
	diceService := domain.NewDiceService()
	agentService := service.NewAgentService()

	// 创建Gin路由
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "trpg-solo-engine-test",
		})
	})

	// 初始化处理器
	diceHandler := handler.NewDiceHandler(diceService)
	agentHandler := handler.NewAgentHandler(agentService)

	// API路由
	api := router.Group("/api")
	{
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": "0.1.0",
				"name":    "TRPG Solo Engine",
			})
		})

		// 骰子API
		dice := api.Group("/dice")
		{
			dice.POST("/roll", diceHandler.RollDice)
		}

		// 角色API
		agents := api.Group("/agents")
		{
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("", agentHandler.ListAgents)
			agents.GET("/:id", agentHandler.GetAgent)
		}
	}

	// 启动服务器
	port := "8080"
	logger.Info("starting test server", zap.String("port", port))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
