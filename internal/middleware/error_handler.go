package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"go.uber.org/zap"
)

// ErrorResponse 标准化错误响应结构
type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandlerMiddleware 创建全局错误处理中间件
// 捕获panic和处理错误，返回标准化的错误响应
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 捕获panic
				stack := debug.Stack()

				// 记录panic详情
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", stack),
				)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Error:   "服务器内部错误",
					Code:    string(domain.ErrInternal),
				})

				// 中止请求处理
				c.Abort()
			}
		}()

		// 处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err

			// 记录错误
			logger.Error("request error",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Int("status", c.Writer.Status()),
			)

			// 如果还没有写入响应，处理错误
			if !c.Writer.Written() {
				HandleError(c, err, logger)
			}
		}
	}
}

// HandleError 处理错误并返回标准化响应
func HandleError(c *gin.Context, err error, logger *zap.Logger) {
	// 检查是否是GameError
	if gameErr, ok := err.(*domain.GameError); ok {
		statusCode := getStatusCodeFromErrorCode(gameErr.Code)

		// 记录错误详情
		logger.Error("game error",
			zap.String("code", string(gameErr.Code)),
			zap.String("message", gameErr.Message),
			zap.Any("details", gameErr.Details),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   gameErr.Message,
			Code:    string(gameErr.Code),
			Details: gameErr.Details,
		})
		return
	}

	// 其他错误类型
	logger.Error("unhandled error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    string(domain.ErrInternal),
	})
}

// getStatusCodeFromErrorCode 根据错误代码返回HTTP状态码
func getStatusCodeFromErrorCode(code domain.ErrorCode) int {
	switch code {
	// 验证错误 -> 400 Bad Request
	case domain.ErrInvalidInput, domain.ErrInvalidARC, domain.ErrInvalidAction:
		return http.StatusBadRequest

	// 资源错误 -> 400 Bad Request
	case domain.ErrInsufficientQA, domain.ErrInsufficientChaos:
		return http.StatusBadRequest

	// 状态错误 -> 400 Bad Request
	case domain.ErrInvalidPhase, domain.ErrInvalidState:
		return http.StatusBadRequest

	// 数据错误 -> 404 Not Found 或 409 Conflict
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrAlreadyExists:
		return http.StatusConflict
	case domain.ErrDataCorrupted:
		return http.StatusInternalServerError

	// 系统错误 -> 500 Internal Server Error
	case domain.ErrInternal, domain.ErrAIService:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

// AbortWithError 中止请求并返回错误
// 这是一个辅助函数，可以在handler中使用
func AbortWithError(c *gin.Context, err error, logger *zap.Logger) {
	c.Error(err)
	HandleError(c, err, logger)
	c.Abort()
}

// AbortWithGameError 中止请求并返回GameError
func AbortWithGameError(c *gin.Context, code domain.ErrorCode, message string, logger *zap.Logger) {
	err := domain.NewGameError(code, message)
	AbortWithError(c, err, logger)
}

// AbortWithGameErrorDetails 中止请求并返回带详情的GameError
func AbortWithGameErrorDetails(c *gin.Context, code domain.ErrorCode, message string, details map[string]interface{}, logger *zap.Logger) {
	err := domain.NewGameError(code, message)
	err.Details = details
	AbortWithError(c, err, logger)
}

// RecoveryMiddleware 创建panic恢复中间件
// 这是一个独立的恢复中间件，可以单独使用
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 捕获panic
				stack := debug.Stack()

				// 记录panic详情
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.ByteString("stack", stack),
				)

				// 返回500错误
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Success: false,
						Error:   "服务器内部错误",
						Code:    string(domain.ErrInternal),
					})
				}

				// 中止请求处理
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ValidationErrorMiddleware 创建验证错误处理中间件
// 专门处理请求验证错误
func ValidationErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有验证错误
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// 记录验证错误
				logger.Warn("validation error",
					zap.Error(e.Err),
					zap.Uint("type", uint(e.Type)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
			}

			// 如果还没有写入响应
			if !c.Writer.Written() {
				// 获取第一个错误
				firstError := c.Errors[0].Err

				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("请求验证失败: %s", firstError.Error()),
					Code:    string(domain.ErrInvalidInput),
				})
			}
		}
	}
}

// NotFoundHandler 处理404错误
func NotFoundHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Warn("route not found",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("ip", c.ClientIP()),
		)

		c.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("路由不存在: %s %s", c.Request.Method, c.Request.URL.Path),
			Code:    string(domain.ErrNotFound),
		})
	}
}

// MethodNotAllowedHandler 处理405错误
func MethodNotAllowedHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Warn("method not allowed",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("ip", c.ClientIP()),
		)

		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("方法不允许: %s", c.Request.Method),
			Code:    "METHOD_NOT_ALLOWED",
		})
	}
}
