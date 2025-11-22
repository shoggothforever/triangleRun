package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 时间窗口内允许的最大请求数
	MaxRequests int
	// 时间窗口大小
	Window time.Duration
	// 是否基于用户ID限流（如果为false，则基于IP）
	ByUser bool
	// 是否同时基于IP和用户限流
	ByBoth bool
}

// RateLimiter 速率限制器
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	config RateLimitConfig
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(redis *redis.Client, logger *zap.Logger, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: logger,
		config: config,
	}
}

// RateLimitMiddleware 创建速率限制中间件
func RateLimitMiddleware(redis *redis.Client, logger *zap.Logger, config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(redis, logger, config)

	return func(c *gin.Context) {
		ctx := context.Background()

		// 根据配置决定限流策略
		var keys []string

		if config.ByBoth {
			// 同时基于IP和用户限流
			ipKey := limiter.getIPKey(c)
			keys = append(keys, ipKey)

			if userID, exists := c.Get("userID"); exists {
				userKey := limiter.getUserKey(userID.(string))
				keys = append(keys, userKey)
			}
		} else if config.ByUser {
			// 仅基于用户限流
			userID, exists := c.Get("userID")
			if !exists {
				// 如果没有用户ID，跳过限流或使用IP
				ipKey := limiter.getIPKey(c)
				keys = append(keys, ipKey)
			} else {
				userKey := limiter.getUserKey(userID.(string))
				keys = append(keys, userKey)
			}
		} else {
			// 仅基于IP限流
			ipKey := limiter.getIPKey(c)
			keys = append(keys, ipKey)
		}

		// 检查所有key的限流状态
		for _, key := range keys {
			allowed, remaining, resetTime, err := limiter.allow(ctx, key)
			if err != nil {
				logger.Error("rate limit check failed",
					zap.Error(err),
					zap.String("key", key),
				)
				// 如果Redis出错，允许请求通过（fail open）
				continue
			}

			// 设置响应头
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				logger.Warn("rate limit exceeded",
					zap.String("key", key),
					zap.String("ip", c.ClientIP()),
					zap.String("path", c.Request.URL.Path),
					zap.Int("limit", config.MaxRequests),
					zap.Duration("window", config.Window),
				)

				c.JSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"error":   "请求过于频繁，请稍后再试",
					"code":    "RATE_LIMIT_EXCEEDED",
					"details": map[string]interface{}{
						"limit":     config.MaxRequests,
						"window":    config.Window.String(),
						"reset_at":  resetTime.Unix(),
						"remaining": 0,
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(ctx context.Context, key string) (allowed bool, remaining int, resetTime time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)

	// 使用Redis的ZSET实现滑动窗口算法
	pipe := rl.redis.Pipeline()

	// 1. 移除窗口外的旧记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 2. 统计当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	// 3. 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// 4. 设置key的过期时间
	pipe.Expire(ctx, key, rl.config.Window+time.Minute)

	// 执行pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	// 获取当前请求数
	count, err := countCmd.Result()
	if err != nil {
		return false, 0, time.Time{}, err
	}

	// 判断是否超过限制
	allowed = int(count) < rl.config.MaxRequests
	remaining = rl.config.MaxRequests - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	// 计算重置时间（窗口结束时间）
	resetTime = now.Add(rl.config.Window)

	return allowed, remaining, resetTime, nil
}

// getIPKey 获取基于IP的key
func (rl *RateLimiter) getIPKey(c *gin.Context) string {
	ip := c.ClientIP()
	return fmt.Sprintf("rate_limit:ip:%s", ip)
}

// getUserKey 获取基于用户的key
func (rl *RateLimiter) getUserKey(userID string) string {
	return fmt.Sprintf("rate_limit:user:%s", userID)
}

// RateLimitByEndpoint 创建基于端点的速率限制中间件
// 不同的端点可以有不同的限流配置
func RateLimitByEndpoint(redis *redis.Client, logger *zap.Logger, configs map[string]RateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	for endpoint, config := range configs {
		limiters[endpoint] = NewRateLimiter(redis, logger, config)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 查找匹配的端点配置
		var limiter *RateLimiter
		var config RateLimitConfig
		for endpoint, cfg := range configs {
			if matchEndpoint(path, endpoint) {
				limiter = limiters[endpoint]
				config = cfg
				break
			}
		}

		// 如果没有匹配的配置，跳过限流
		if limiter == nil {
			c.Next()
			return
		}

		ctx := context.Background()

		// 根据配置决定限流策略
		var keys []string

		if config.ByBoth {
			ipKey := limiter.getIPKey(c)
			keys = append(keys, ipKey)

			if userID, exists := c.Get("userID"); exists {
				userKey := limiter.getUserKey(userID.(string))
				keys = append(keys, userKey)
			}
		} else if config.ByUser {
			userID, exists := c.Get("userID")
			if !exists {
				ipKey := limiter.getIPKey(c)
				keys = append(keys, ipKey)
			} else {
				userKey := limiter.getUserKey(userID.(string))
				keys = append(keys, userKey)
			}
		} else {
			ipKey := limiter.getIPKey(c)
			keys = append(keys, ipKey)
		}

		// 检查所有key的限流状态
		for _, key := range keys {
			allowed, remaining, resetTime, err := limiter.allow(ctx, key)
			if err != nil {
				logger.Error("rate limit check failed",
					zap.Error(err),
					zap.String("key", key),
				)
				continue
			}

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				logger.Warn("rate limit exceeded",
					zap.String("key", key),
					zap.String("ip", c.ClientIP()),
					zap.String("path", path),
					zap.String("endpoint", path),
					zap.Int("limit", config.MaxRequests),
					zap.Duration("window", config.Window),
				)

				c.JSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"error":   "请求过于频繁，请稍后再试",
					"code":    "RATE_LIMIT_EXCEEDED",
					"details": map[string]interface{}{
						"limit":     config.MaxRequests,
						"window":    config.Window.String(),
						"reset_at":  resetTime.Unix(),
						"remaining": 0,
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// matchEndpoint 检查路径是否匹配端点模式
func matchEndpoint(path, pattern string) bool {
	// 简单的前缀匹配
	// 可以扩展为更复杂的模式匹配
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return path == pattern
}

// DefaultRateLimitConfigs 默认的速率限制配置
func DefaultRateLimitConfigs() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		// 骰子掷骰 - 每分钟100次
		"/api/dice/*": {
			MaxRequests: 100,
			Window:      time.Minute,
			ByUser:      true,
		},
		// AI生成 - 每分钟10次
		"/api/ai/*": {
			MaxRequests: 10,
			Window:      time.Minute,
			ByUser:      true,
		},
		// 存档操作 - 每分钟20次
		"/api/saves/*": {
			MaxRequests: 20,
			Window:      time.Minute,
			ByUser:      true,
		},
		// 会话操作 - 每分钟30次
		"/api/sessions/*": {
			MaxRequests: 30,
			Window:      time.Minute,
			ByUser:      true,
		},
		// 角色操作 - 每分钟20次
		"/api/agents/*": {
			MaxRequests: 20,
			Window:      time.Minute,
			ByUser:      true,
		},
		// 剧本查询 - 每分钟50次
		"/api/scenarios/*": {
			MaxRequests: 50,
			Window:      time.Minute,
			ByUser:      false, // 基于IP
		},
	}
}
