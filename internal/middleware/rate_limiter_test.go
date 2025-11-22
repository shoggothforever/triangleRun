package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupTestRedis 创建测试用的Redis客户端
// 使用真实的Redis连接进行测试
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // 使用测试数据库
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping rate limiter tests")
	}

	// 清空测试数据库
	client.FlushDB(ctx)

	return client
}

func TestRateLimitMiddleware_IPBased(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 5,
		Window:      time.Second,
		ByUser:      false, // 基于IP
	}

	router := gin.New()
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试：前5个请求应该成功
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i+1))

		// 检查响应头
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	}

	// 测试：第6个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "RATE_LIMIT_EXCEEDED")
}

func TestRateLimitMiddleware_UserBased(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 3,
		Window:      time.Second,
		ByUser:      true, // 基于用户
	}

	router := gin.New()
	// 模拟认证中间件设置userID
	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	})
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试：用户1的前3个请求应该成功
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "user1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("User1 request %d should succeed", i+1))
	}

	// 测试：用户1的第4个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 测试：用户2的请求应该成功（不同用户独立计数）
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user2")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_Reset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 2,
		Window:      500 * time.Millisecond, // 短窗口便于测试
		ByUser:      false,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 发送2个请求，用完配额
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 第3个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 等待窗口重置
	time.Sleep(600 * time.Millisecond)

	// 窗口重置后，请求应该再次成功
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 2,
		Window:      time.Second,
		ByUser:      false,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// IP1: 发送2个请求
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.10:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP1: 第3个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.10:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// IP2: 请求应该成功（不同IP独立计数）
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.20:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitByEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	configs := map[string]RateLimitConfig{
		"/api/dice/*": {
			MaxRequests: 3,
			Window:      time.Second,
			ByUser:      false,
		},
		"/api/ai/*": {
			MaxRequests: 1,
			Window:      time.Second,
			ByUser:      false,
		},
	}

	router := gin.New()
	router.Use(RateLimitByEndpoint(redisClient, logger, configs))
	router.GET("/api/dice/roll", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "dice rolled"})
	})
	router.GET("/api/ai/generate", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ai generated"})
	})
	router.GET("/api/other", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "no limit"})
	})

	// 测试 /api/dice/* 端点（限制3次）
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/dice/roll", nil)
		req.RemoteAddr = "192.168.1.30:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Dice request %d should succeed", i+1))
	}

	// 第4个请求应该被限流
	req := httptest.NewRequest(http.MethodGet, "/api/dice/roll", nil)
	req.RemoteAddr = "192.168.1.30:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 测试 /api/ai/* 端点（限制1次）- 使用不同的IP
	req = httptest.NewRequest(http.MethodGet, "/api/ai/generate", nil)
	req.RemoteAddr = "192.168.1.31:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 第2个AI请求应该被限流
	req = httptest.NewRequest(http.MethodGet, "/api/ai/generate", nil)
	req.RemoteAddr = "192.168.1.31:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 测试没有限制的端点 - 使用不同的IP
	for i := 0; i < 10; i++ {
		req = httptest.NewRequest(http.MethodGet, "/api/other", nil)
		req.RemoteAddr = "192.168.1.32:12345"
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Unlimited endpoint should always succeed")
	}
}

func TestRateLimitMiddleware_ResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 5,
		Window:      time.Second,
		ByUser:      false,
	}

	router := gin.New()
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 第一个请求
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.40:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))

	// 第二个请求
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.40:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "3", w.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimitMiddleware_BothIPAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 2,
		Window:      time.Second,
		ByBoth:      true, // 同时基于IP和用户
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	})
	router.Use(RateLimitMiddleware(redisClient, logger, config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 用户1从IP1发送2个请求
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "user1")
		req.RemoteAddr = "192.168.1.50:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 用户1从IP1的第3个请求应该被限流（用户维度）
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user1")
	req.RemoteAddr = "192.168.1.50:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 用户2从IP1的请求应该被限流（因为IP1已经达到限制）
	// 当ByBoth为true时，IP和用户都会被检查
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user2")
	req.RemoteAddr = "192.168.1.50:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 用户2从不同IP的请求应该成功
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user2")
	req.RemoteAddr = "192.168.1.51:12345"
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMatchEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"/api/dice/roll", "/api/dice/*", true},
		{"/api/dice/ability", "/api/dice/*", true},
		{"/api/ai/generate", "/api/dice/*", false},
		{"/api/dice", "/api/dice/*", false},
		{"/api/dice/roll", "/api/dice/roll", true},
		{"/api/dice/ability", "/api/dice/roll", false},
		{"/api/saves/123", "/api/saves/*", true},
		{"/api/saves", "/api/saves/*", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s matches %s", tt.path, tt.pattern), func(t *testing.T) {
			result := matchEndpoint(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultRateLimitConfigs(t *testing.T) {
	configs := DefaultRateLimitConfigs()

	// 验证配置存在
	assert.NotNil(t, configs)
	assert.NotEmpty(t, configs)

	// 验证关键端点的配置
	diceConfig, exists := configs["/api/dice/*"]
	assert.True(t, exists)
	assert.Equal(t, 100, diceConfig.MaxRequests)
	assert.Equal(t, time.Minute, diceConfig.Window)
	assert.True(t, diceConfig.ByUser)

	aiConfig, exists := configs["/api/ai/*"]
	assert.True(t, exists)
	assert.Equal(t, 10, aiConfig.MaxRequests)
	assert.Equal(t, time.Minute, aiConfig.Window)
	assert.True(t, aiConfig.ByUser)

	savesConfig, exists := configs["/api/saves/*"]
	assert.True(t, exists)
	assert.Equal(t, 20, savesConfig.MaxRequests)
	assert.Equal(t, time.Minute, savesConfig.Window)
	assert.True(t, savesConfig.ByUser)
}

func TestRateLimiter_Allow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	config := RateLimitConfig{
		MaxRequests: 3,
		Window:      time.Second,
		ByUser:      false,
	}

	limiter := NewRateLimiter(redisClient, logger, config)
	ctx := context.Background()
	key := "test:rate_limit:ip:192.168.1.100"

	// 前3个请求应该被允许
	for i := 0; i < 3; i++ {
		allowed, remaining, resetTime, err := limiter.allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed, fmt.Sprintf("Request %d should be allowed", i+1))
		assert.Equal(t, 2-i, remaining)
		assert.True(t, resetTime.After(time.Now()))
	}

	// 第4个请求应该被拒绝
	allowed, remaining, _, err := limiter.allow(ctx, key)
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}
