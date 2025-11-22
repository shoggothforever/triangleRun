package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"go.uber.org/zap"
)

func setupTestRouter() (*gin.Engine, *zap.Logger) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	router := gin.New()
	return router, logger
}

func TestErrorHandlerMiddleware_PanicRecovery(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ErrorHandlerMiddleware(logger))

	// 创建一个会panic的路由
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "服务器内部错误", response.Error)
	assert.Equal(t, string(domain.ErrInternal), response.Code)
}

func TestErrorHandlerMiddleware_GameError(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      domain.ErrorCode
		errorMessage   string
		expectedStatus int
	}{
		{
			name:           "Invalid Input Error",
			errorCode:      domain.ErrInvalidInput,
			errorMessage:   "输入无效",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found Error",
			errorCode:      domain.ErrNotFound,
			errorMessage:   "资源未找到",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Already Exists Error",
			errorCode:      domain.ErrAlreadyExists,
			errorMessage:   "资源已存在",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Internal Error",
			errorCode:      domain.ErrInternal,
			errorMessage:   "内部错误",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid ARC Error",
			errorCode:      domain.ErrInvalidARC,
			errorMessage:   "ARC组合无效",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Insufficient QA Error",
			errorCode:      domain.ErrInsufficientQA,
			errorMessage:   "资质保证不足",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, logger := setupTestRouter()
			router.Use(ErrorHandlerMiddleware(logger))

			// 创建返回GameError的路由
			router.GET("/error", func(c *gin.Context) {
				err := domain.NewGameError(tt.errorCode, tt.errorMessage)
				c.Error(err)
			})

			// 发送请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/error", nil)
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.Equal(t, tt.errorMessage, response.Error)
			assert.Equal(t, string(tt.errorCode), response.Code)
		})
	}
}

func TestErrorHandlerMiddleware_GameErrorWithDetails(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ErrorHandlerMiddleware(logger))

	// 创建返回带详情的GameError的路由
	router.GET("/error-details", func(c *gin.Context) {
		err := domain.NewGameError(domain.ErrInvalidInput, "验证失败")
		err.Details = map[string]interface{}{
			"field":  "name",
			"reason": "名称不能为空",
		}
		c.Error(err)
	})

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error-details", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "验证失败", response.Error)
	assert.Equal(t, string(domain.ErrInvalidInput), response.Code)
	assert.NotNil(t, response.Details)
	assert.Equal(t, "name", response.Details["field"])
	assert.Equal(t, "名称不能为空", response.Details["reason"])
}

func TestErrorHandlerMiddleware_GenericError(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ErrorHandlerMiddleware(logger))

	// 创建返回普通错误的路由
	router.GET("/generic-error", func(c *gin.Context) {
		c.Error(errors.New("generic error"))
	})

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/generic-error", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "generic error", response.Error)
	assert.Equal(t, string(domain.ErrInternal), response.Code)
}

func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ErrorHandlerMiddleware(logger))

	// 创建正常的路由
	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "GameError - Invalid Input",
			error:          domain.NewGameError(domain.ErrInvalidInput, "输入无效"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   string(domain.ErrInvalidInput),
		},
		{
			name:           "GameError - Not Found",
			error:          domain.NewGameError(domain.ErrNotFound, "未找到"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   string(domain.ErrNotFound),
		},
		{
			name:           "Generic Error",
			error:          errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   string(domain.ErrInternal),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, logger := setupTestRouter()

			router.GET("/test", func(c *gin.Context) {
				HandleError(c, tt.error, logger)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.Equal(t, tt.expectedCode, response.Code)
		})
	}
}

func TestAbortWithError(t *testing.T) {
	router, logger := setupTestRouter()

	router.GET("/abort", func(c *gin.Context) {
		err := domain.NewGameError(domain.ErrInvalidInput, "测试错误")
		AbortWithError(c, err, logger)
		// 使用return确保不会继续执行
		return
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abort", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "测试错误", response.Error)
}

func TestAbortWithGameError(t *testing.T) {
	router, logger := setupTestRouter()

	router.GET("/abort-game-error", func(c *gin.Context) {
		AbortWithGameError(c, domain.ErrNotFound, "资源未找到", logger)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abort-game-error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "资源未找到", response.Error)
	assert.Equal(t, string(domain.ErrNotFound), response.Code)
}

func TestAbortWithGameErrorDetails(t *testing.T) {
	router, logger := setupTestRouter()

	router.GET("/abort-with-details", func(c *gin.Context) {
		details := map[string]interface{}{
			"field": "email",
			"value": "invalid",
		}
		AbortWithGameErrorDetails(c, domain.ErrInvalidInput, "邮箱格式错误", details, logger)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abort-with-details", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "邮箱格式错误", response.Error)
	assert.NotNil(t, response.Details)
	assert.Equal(t, "email", response.Details["field"])
}

func TestRecoveryMiddleware(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(RecoveryMiddleware(logger))

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic in recovery middleware")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "服务器内部错误", response.Error)
}

func TestValidationErrorMiddleware(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ValidationErrorMiddleware(logger))

	router.GET("/validation-error", func(c *gin.Context) {
		c.Error(errors.New("validation failed"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validation-error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "请求验证失败")
}

func TestNotFoundHandler(t *testing.T) {
	router, logger := setupTestRouter()
	router.NoRoute(NotFoundHandler(logger))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "路由不存在")
	assert.Equal(t, string(domain.ErrNotFound), response.Code)
}

func TestMethodNotAllowedHandler(t *testing.T) {
	router, logger := setupTestRouter()
	router.HandleMethodNotAllowed = true
	router.NoMethod(MethodNotAllowedHandler(logger))

	// 只定义GET方法
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// 尝试使用POST方法
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "方法不允许")
}

func TestGetStatusCodeFromErrorCode(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      domain.ErrorCode
		expectedStatus int
	}{
		{"Invalid Input", domain.ErrInvalidInput, http.StatusBadRequest},
		{"Invalid ARC", domain.ErrInvalidARC, http.StatusBadRequest},
		{"Invalid Action", domain.ErrInvalidAction, http.StatusBadRequest},
		{"Insufficient QA", domain.ErrInsufficientQA, http.StatusBadRequest},
		{"Insufficient Chaos", domain.ErrInsufficientChaos, http.StatusBadRequest},
		{"Invalid Phase", domain.ErrInvalidPhase, http.StatusBadRequest},
		{"Invalid State", domain.ErrInvalidState, http.StatusBadRequest},
		{"Not Found", domain.ErrNotFound, http.StatusNotFound},
		{"Already Exists", domain.ErrAlreadyExists, http.StatusConflict},
		{"Data Corrupted", domain.ErrDataCorrupted, http.StatusInternalServerError},
		{"Internal Error", domain.ErrInternal, http.StatusInternalServerError},
		{"AI Service Error", domain.ErrAIService, http.StatusInternalServerError},
		{"Unknown Error", domain.ErrorCode("UNKNOWN"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode := getStatusCodeFromErrorCode(tt.errorCode)
			assert.Equal(t, tt.expectedStatus, statusCode)
		})
	}
}

func TestErrorResponse_Format(t *testing.T) {
	// 测试ErrorResponse的JSON序列化
	response := ErrorResponse{
		Success: false,
		Error:   "测试错误",
		Code:    "TEST_ERROR",
		Details: map[string]interface{}{
			"field": "test",
			"value": 123,
		},
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded ErrorResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, response.Success, decoded.Success)
	assert.Equal(t, response.Error, decoded.Error)
	assert.Equal(t, response.Code, decoded.Code)
	assert.Equal(t, "test", decoded.Details["field"])
}

func TestMultipleErrors(t *testing.T) {
	router, logger := setupTestRouter()
	router.Use(ErrorHandlerMiddleware(logger))

	router.GET("/multiple-errors", func(c *gin.Context) {
		c.Error(errors.New("first error"))
		c.Error(errors.New("second error"))
		c.Error(domain.NewGameError(domain.ErrInvalidInput, "third error"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multiple-errors", nil)
	router.ServeHTTP(w, req)

	// 应该处理最后一个错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "third error", response.Error)
}
