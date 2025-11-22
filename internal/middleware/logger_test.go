package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		statusCode     int
		requestBody    string
		responseBody   string
		expectedLevel  zapcore.Level
		expectedFields map[string]interface{}
	}{
		{
			name:          "successful GET request",
			method:        "GET",
			path:          "/api/agents",
			query:         "limit=10",
			statusCode:    200,
			expectedLevel: zapcore.InfoLevel,
			expectedFields: map[string]interface{}{
				"method": "GET",
				"path":   "/api/agents",
				"query":  "limit=10",
				"status": 200,
			},
		},
		{
			name:          "successful POST request",
			method:        "POST",
			path:          "/api/agents",
			statusCode:    201,
			requestBody:   `{"name":"test"}`,
			responseBody:  `{"id":"123"}`,
			expectedLevel: zapcore.InfoLevel,
			expectedFields: map[string]interface{}{
				"method": "POST",
				"path":   "/api/agents",
				"status": 201,
			},
		},
		{
			name:          "client error",
			method:        "GET",
			path:          "/api/agents/invalid",
			statusCode:    404,
			expectedLevel: zapcore.WarnLevel,
			expectedFields: map[string]interface{}{
				"method": "GET",
				"path":   "/api/agents/invalid",
				"status": 404,
			},
		},
		{
			name:          "server error",
			method:        "POST",
			path:          "/api/agents",
			statusCode:    500,
			expectedLevel: zapcore.ErrorLevel,
			expectedFields: map[string]interface{}{
				"method": "POST",
				"path":   "/api/agents",
				"status": 500,
			},
		},
		{
			name:          "redirect",
			method:        "GET",
			path:          "/old-path",
			statusCode:    301,
			expectedLevel: zapcore.InfoLevel,
			expectedFields: map[string]interface{}{
				"method": "GET",
				"path":   "/old-path",
				"status": 301,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建观察者来捕获日志
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			// 创建测试路由
			router := gin.New()
			router.Use(LoggerMiddleware(logger))
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				if tt.responseBody != "" {
					c.String(tt.statusCode, tt.responseBody)
				} else {
					c.Status(tt.statusCode)
				}
			})

			// 创建请求
			var req *http.Request
			if tt.requestBody != "" {
				req = httptest.NewRequest(tt.method, tt.path+"?"+tt.query, strings.NewReader(tt.requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path+"?"+tt.query, nil)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应状态码
			assert.Equal(t, tt.statusCode, w.Code)

			// 验证日志记录
			logs := recorded.All()
			assert.Greater(t, len(logs), 0, "should have logged at least one entry")

			// 查找主要的请求日志
			var mainLog *observer.LoggedEntry
			for i := range logs {
				if logs[i].Level == tt.expectedLevel {
					mainLog = &logs[i]
					break
				}
			}

			assert.NotNil(t, mainLog, "should have logged entry with expected level")

			// 验证日志级别
			assert.Equal(t, tt.expectedLevel, mainLog.Level)

			// 验证日志字段
			for key, expectedValue := range tt.expectedFields {
				found := false
				for _, field := range mainLog.Context {
					if field.Key == key {
						found = true
						switch v := expectedValue.(type) {
						case string:
							assert.Equal(t, v, field.String)
						case int:
							assert.Equal(t, int64(v), field.Integer)
						}
						break
					}
				}
				assert.True(t, found, "should have field: %s", key)
			}

			// 验证必需字段存在
			requiredFields := []string{"method", "path", "status", "latency", "ip"}
			for _, field := range requiredFields {
				found := false
				for _, logField := range mainLog.Context {
					if logField.Key == field {
						found = true
						break
					}
				}
				assert.True(t, found, "should have required field: %s", field)
			}
		})
	}
}

func TestLoggerMiddleware_WithUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建观察者
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	// 创建测试路由
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Status(200)
	})

	// 执行请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证日志包含用户ID
	logs := recorded.All()
	assert.Greater(t, len(logs), 0)

	found := false
	for _, log := range logs {
		for _, field := range log.Context {
			if field.Key == "user_id" && field.String == "user123" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "should log user_id")
}

func TestLoggerMiddleware_WithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建观察者
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	// 创建测试路由
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.Status(500)
	})

	// 执行请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证日志包含错误信息
	logs := recorded.All()
	assert.Greater(t, len(logs), 0)

	found := false
	for _, log := range logs {
		for _, field := range log.Context {
			if field.Key == "errors" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "should log errors")
}

func TestLoggerMiddleware_SlowRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建观察者
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	// 创建测试路由
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(2100 * time.Millisecond)
		c.Status(200)
	})

	// 执行请求
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证有慢请求警告
	logs := recorded.All()
	slowWarningFound := false
	for _, log := range logs {
		if log.Level == zapcore.WarnLevel && strings.Contains(log.Message, "slow request") {
			slowWarningFound = true
			break
		}
	}
	assert.True(t, slowWarningFound, "should log slow request warning")
}

func TestDetailedLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		logBody      bool
		requestBody  string
		responseBody string
		shouldLogReq bool
		shouldLogRes bool
	}{
		{
			name:         "log body enabled",
			logBody:      true,
			requestBody:  `{"name":"test"}`,
			responseBody: `{"id":"123"}`,
			shouldLogReq: true,
			shouldLogRes: true,
		},
		{
			name:         "log body disabled",
			logBody:      false,
			requestBody:  `{"name":"test"}`,
			responseBody: `{"id":"123"}`,
			shouldLogReq: false,
			shouldLogRes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建观察者
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			// 创建测试路由
			router := gin.New()
			router.Use(DetailedLoggerMiddleware(logger, tt.logBody))
			router.POST("/test", func(c *gin.Context) {
				c.String(200, tt.responseBody)
			})

			// 执行请求
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证日志
			logs := recorded.All()
			assert.Greater(t, len(logs), 0)

			// 检查是否记录了请求体
			reqBodyLogged := false
			resBodyLogged := false
			for _, log := range logs {
				// 检查请求体日志（在单独的debug日志中）
				if log.Message == "request body" {
					for _, field := range log.Context {
						if field.Key == "body" {
							reqBodyLogged = true
						}
					}
				}
				// 检查响应体（在完成日志中）
				if strings.Contains(log.Message, "request completed") {
					for _, field := range log.Context {
						if field.Key == "response_body" {
							resBodyLogged = true
						}
					}
				}
			}

			if tt.shouldLogReq {
				assert.True(t, reqBodyLogged, "should log request body")
			} else {
				assert.False(t, reqBodyLogged, "should not log request body")
			}

			if tt.shouldLogRes {
				assert.True(t, resBodyLogged, "should log response body")
			} else {
				assert.False(t, resBodyLogged, "should not log response body")
			}
		})
	}
}

func TestPerformanceTrackerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		delay         time.Duration
		shouldWarn    bool
		expectedLevel zapcore.Level
	}{
		{
			name:          "fast request",
			delay:         50 * time.Millisecond,
			shouldWarn:    false,
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:          "slow request",
			delay:         1100 * time.Millisecond,
			shouldWarn:    true,
			expectedLevel: zapcore.WarnLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建观察者
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			// 创建测试路由
			router := gin.New()
			router.Use(PerformanceTrackerMiddleware(logger))
			router.GET("/test", func(c *gin.Context) {
				time.Sleep(tt.delay)
				c.Status(200)
			})

			// 执行请求
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证日志
			logs := recorded.All()
			assert.Greater(t, len(logs), 0)

			// 检查性能日志
			perfLogFound := false
			warnLogFound := false
			for _, log := range logs {
				if log.Message == "performance" {
					perfLogFound = true
					// 验证必需字段
					hasLatency := false
					hasLatencyMs := false
					for _, field := range log.Context {
						if field.Key == "latency" {
							hasLatency = true
						}
						if field.Key == "latency_ms" {
							hasLatencyMs = true
						}
					}
					assert.True(t, hasLatency, "should have latency field")
					assert.True(t, hasLatencyMs, "should have latency_ms field")
				}
				if log.Level == zapcore.WarnLevel && strings.Contains(log.Message, "performance warning") {
					warnLogFound = true
				}
			}

			assert.True(t, perfLogFound, "should log performance metrics")
			if tt.shouldWarn {
				assert.True(t, warnLogFound, "should log performance warning")
			}
		})
	}
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		latency  time.Duration
		expected string
	}{
		{
			name:     "low severity",
			latency:  500 * time.Millisecond,
			expected: "low",
		},
		{
			name:     "medium severity",
			latency:  1500 * time.Millisecond,
			expected: "medium",
		},
		{
			name:     "high severity",
			latency:  3500 * time.Millisecond,
			expected: "high",
		},
		{
			name:     "critical severity",
			latency:  6 * time.Second,
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSeverity(tt.latency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建Gin上下文来获取正确的ResponseWriter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 创建测试ResponseWriter
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(""),
	}

	// 写入数据
	testData := []byte("test response")
	n, err := rw.Write(testData)

	// 验证
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, rw.body.Bytes())
}

func TestLoggerMiddleware_RequestResponseSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建观察者
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	requestBody := `{"name":"test","description":"this is a test"}`
	responseBody := `{"id":"123","status":"created"}`

	// 创建测试路由
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.POST("/test", func(c *gin.Context) {
		c.String(200, responseBody)
	})

	// 执行请求
	req := httptest.NewRequest("POST", "/test", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证日志包含大小信息
	logs := recorded.All()
	assert.Greater(t, len(logs), 0)

	found := false
	for _, log := range logs {
		hasReqSize := false
		hasResSize := false
		for _, field := range log.Context {
			if field.Key == "request_size" {
				hasReqSize = true
				assert.Equal(t, int64(len(requestBody)), field.Integer)
			}
			if field.Key == "response_size" {
				hasResSize = true
				assert.Equal(t, int64(len(responseBody)), field.Integer)
			}
		}
		if hasReqSize && hasResSize {
			found = true
			break
		}
	}
	assert.True(t, found, "should log request and response sizes")
}

func TestLoggerMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建观察者
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	// 创建完整的测试路由
	router := gin.New()
	router.Use(LoggerMiddleware(logger))

	// 模拟真实的API端点
	router.POST("/api/agents", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		c.Set("userID", "test-user")
		c.JSON(201, gin.H{"id": "agent-123", "name": body["name"]})
	})

	// 执行请求
	reqBody := `{"name":"Test Agent","anomaly":"低语"}`
	req := httptest.NewRequest("POST", "/api/agents", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "agent-123", response["id"])

	// 验证日志
	logs := recorded.All()
	assert.Greater(t, len(logs), 0)

	// 查找主日志
	var mainLog *observer.LoggedEntry
	for i := range logs {
		if logs[i].Message == "request completed" {
			mainLog = &logs[i]
			break
		}
	}

	assert.NotNil(t, mainLog)
	assert.Equal(t, zapcore.InfoLevel, mainLog.Level)

	// 验证所有关键字段
	fieldMap := make(map[string]interface{})
	for _, field := range mainLog.Context {
		fieldMap[field.Key] = field
	}

	assert.Contains(t, fieldMap, "method")
	assert.Contains(t, fieldMap, "path")
	assert.Contains(t, fieldMap, "status")
	assert.Contains(t, fieldMap, "latency")
	assert.Contains(t, fieldMap, "ip")
	assert.Contains(t, fieldMap, "user_id")
	assert.Contains(t, fieldMap, "request_size")
	assert.Contains(t, fieldMap, "response_size")
}
