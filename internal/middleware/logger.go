package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter 包装gin.ResponseWriter以捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware 创建请求日志中间件
// 记录请求和响应信息，包括性能追踪
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 获取请求信息
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 读取请求体（如果需要）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 恢复请求体供后续处理使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装ResponseWriter以捕获响应体
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("request_size", len(requestBody)),
			zap.Int("response_size", blw.body.Len()),
		}

		// 如果有用户ID，添加到日志中
		if userID, exists := c.Get("userID"); exists {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}

		// 如果有错误，添加到日志中
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("server error", fields...)
		case statusCode >= 400:
			logger.Warn("client error", fields...)
		case statusCode >= 300:
			logger.Info("redirect", fields...)
		default:
			logger.Info("request completed", fields...)
		}

		// 性能追踪：如果请求耗时超过阈值，记录警告
		if latency > 2*time.Second {
			logger.Warn("slow request detected",
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.String("threshold", "2s"),
			)
		}
	}
}

// DetailedLoggerMiddleware 创建详细的请求日志中间件
// 包含请求和响应体的详细信息（用于调试）
func DetailedLoggerMiddleware(logger *zap.Logger, logBody bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 获取请求信息
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil && logBody {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装ResponseWriter
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 记录请求开始
		logger.Debug("request started",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
		)

		if logBody && len(requestBody) > 0 {
			logger.Debug("request body",
				zap.String("path", path),
				zap.ByteString("body", requestBody),
			)
		}

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("request_size", len(requestBody)),
			zap.Int("response_size", blw.body.Len()),
		}

		// 添加用户ID
		if userID, exists := c.Get("userID"); exists {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 记录响应体
		if logBody && blw.body.Len() > 0 {
			fields = append(fields, zap.ByteString("response_body", blw.body.Bytes()))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("request completed with server error", fields...)
		case statusCode >= 400:
			logger.Warn("request completed with client error", fields...)
		default:
			logger.Info("request completed successfully", fields...)
		}
	}
}

// PerformanceTrackerMiddleware 创建性能追踪中间件
// 专注于性能指标的收集和记录
func PerformanceTrackerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算性能指标
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录性能指标
		logger.Info("performance",
			zap.String("method", method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.Int("status", statusCode),
			zap.Bool("success", statusCode < 400),
		)

		// 性能警告
		if latency > 1*time.Second {
			logger.Warn("performance warning",
				zap.String("method", method),
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.String("severity", getSeverity(latency)),
			)
		}
	}
}

// getSeverity 根据延迟时间返回严重程度
func getSeverity(latency time.Duration) string {
	switch {
	case latency > 5*time.Second:
		return "critical"
	case latency > 3*time.Second:
		return "high"
	case latency > 1*time.Second:
		return "medium"
	default:
		return "low"
	}
}
