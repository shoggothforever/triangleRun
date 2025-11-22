package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key"
	userID := "test-user-123"

	// 生成有效token
	token, err := GenerateToken(userID, secretKey)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware(secretKey))
	router.GET("/protected", func(c *gin.Context) {
		// 验证用户ID是否正确设置
		extractedUserID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, userID, extractedUserID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建请求
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key"

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware(secretKey))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name          string
		token         string
		expectedCode  int
		expectedError string
	}{
		{
			name:          "malformed token",
			token:         "invalid.token.here",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid token",
		},
		{
			name:          "token with wrong signature",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.wrong_signature",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid token",
		},
		{
			name: "token signed with different key",
			token: func() string {
				token, _ := GenerateToken("test-user", "different-secret")
				return token
			}(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key"

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware(secretKey))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name             string
		authHeader       string
		expectedCode     int
		expectedErrorMsg string
	}{
		{
			name:             "no authorization header",
			authHeader:       "",
			expectedCode:     http.StatusUnauthorized,
			expectedErrorMsg: "missing authorization header",
		},
		{
			name:             "missing Bearer prefix",
			authHeader:       "some-token",
			expectedCode:     http.StatusUnauthorized,
			expectedErrorMsg: "invalid authorization header format",
		},
		{
			name:             "wrong prefix",
			authHeader:       "Basic some-token",
			expectedCode:     http.StatusUnauthorized,
			expectedErrorMsg: "invalid authorization header format",
		},
		{
			name:             "Bearer without token",
			authHeader:       "Bearer",
			expectedCode:     http.StatusUnauthorized,
			expectedErrorMsg: "invalid authorization header format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	userID := "test-user-123"

	// 生成token
	tokenString, err := GenerateToken(userID, secretKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 验证token可以被解析
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	// 验证claims
	claims, ok := token.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
}

func TestAuthMiddleware_UserIDExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key"
	userID := "user-456"

	// 生成token
	token, err := GenerateToken(userID, secretKey)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware(secretKey))
	router.GET("/user", func(c *gin.Context) {
		extractedUserID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": extractedUserID})
	})

	// 创建请求
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
}
