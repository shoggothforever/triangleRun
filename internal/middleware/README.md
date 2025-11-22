# Middleware Package

## Overview

This package provides middleware components for the TRPG Solo Engine API, including:
- JWT-based authentication
- Request/response logging
- Performance tracking

## Usage

### Basic Setup

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/trpg-solo-engine/backend/internal/middleware"
)

func main() {
    router := gin.Default()
    
    // Set your secret key (should be loaded from config)
    secretKey := "your-secret-key"
    
    // Apply middleware to protected routes
    protected := router.Group("/api")
    protected.Use(middleware.AuthMiddleware(secretKey))
    {
        protected.GET("/agents", agentHandler.ListAgents)
        protected.POST("/sessions", sessionHandler.CreateSession)
        // ... other protected routes
    }
    
    router.Run(":8080")
}
```

### Generating Tokens

For testing or initial authentication:

```go
import "github.com/trpg-solo-engine/backend/internal/middleware"

// Generate a token for a user
token, err := middleware.GenerateToken("user-123", "your-secret-key")
if err != nil {
    // handle error
}

// Use the token in requests
// Authorization: Bearer <token>
```

### Accessing User Information

In your handlers, you can access the authenticated user ID:

```go
func MyHandler(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    
    // Use userID for business logic
    userIDStr := userID.(string)
    // ...
}
```

## Configuration

The middleware requires a secret key for signing and verifying JWT tokens. This should be:

1. Stored securely (environment variable or secrets manager)
2. At least 32 characters long
3. Randomly generated
4. Never committed to version control

Example configuration in `config.yaml`:

```yaml
auth:
  jwt_secret: ${JWT_SECRET}  # Load from environment variable
```

## Token Format

The middleware expects tokens in the following format:

```
Authorization: Bearer <jwt-token>
```

The JWT token contains the following claims:

```json
{
  "user_id": "string",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Error Responses

The middleware returns the following error responses:

- `401 Unauthorized` - Missing authorization header
- `401 Unauthorized` - Invalid authorization header format
- `401 Unauthorized` - Invalid token
- `401 Unauthorized` - Invalid token claims

## Logging Middleware

### Overview

The logging middleware provides comprehensive request/response logging with performance tracking capabilities.

### Basic Logger

The `LoggerMiddleware` provides standard request logging:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/trpg-solo-engine/backend/internal/middleware"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    router := gin.New()
    
    // Add logging middleware
    router.Use(middleware.LoggerMiddleware(logger))
    
    // Your routes...
    router.Run(":8080")
}
```

**Features:**
- Request method, path, and query parameters
- Response status code
- Request/response size
- Client IP and user agent
- Request latency
- User ID (if authenticated)
- Error messages (if any)
- Automatic log level selection based on status code:
  - 5xx: Error level
  - 4xx: Warn level
  - 3xx: Info level
  - 2xx: Info level
- Slow request detection (>2s)

### Detailed Logger

The `DetailedLoggerMiddleware` provides verbose logging including request/response bodies:

```go
// Enable detailed logging with body capture
router.Use(middleware.DetailedLoggerMiddleware(logger, true))

// Disable body logging
router.Use(middleware.DetailedLoggerMiddleware(logger, false))
```

**Use Cases:**
- Development and debugging
- Troubleshooting specific issues
- Audit logging

**Warning:** Logging request/response bodies can impact performance and may expose sensitive data. Use with caution in production.

### Performance Tracker

The `PerformanceTrackerMiddleware` focuses on performance metrics:

```go
router.Use(middleware.PerformanceTrackerMiddleware(logger))
```

**Features:**
- Request latency in milliseconds
- Success/failure tracking
- Performance warnings for slow requests (>1s)
- Severity levels:
  - Low: < 1s
  - Medium: 1-3s
  - High: 3-5s
  - Critical: > 5s

### Log Output Example

```json
{
  "level": "info",
  "ts": 1234567890.123,
  "msg": "request completed",
  "method": "POST",
  "path": "/api/agents",
  "query": "",
  "status": 201,
  "latency": "45.2ms",
  "ip": "127.0.0.1",
  "user_agent": "Mozilla/5.0...",
  "user_id": "user-123",
  "request_size": 256,
  "response_size": 128
}
```

### Combining Middlewares

You can combine multiple middlewares:

```go
router := gin.New()
router.Use(gin.Recovery())
router.Use(middleware.LoggerMiddleware(logger))

// Protected routes with authentication
protected := router.Group("/api")
protected.Use(middleware.AuthMiddleware(secretKey))
{
    protected.POST("/agents", handler.CreateAgent)
}
```

## Testing

Run the tests with:

```bash
go test ./internal/middleware/...
```

For coverage:

```bash
go test ./internal/middleware/... -cover
```

## Best Practices

1. **Authentication**: Always use HTTPS in production and store JWT secrets securely
2. **Logging**: Use `LoggerMiddleware` for production, `DetailedLoggerMiddleware` for debugging only
3. **Performance**: Monitor slow request warnings and optimize accordingly
4. **Order**: Apply middlewares in the correct order:
   - Recovery (first)
   - Logging
   - Authentication
   - Custom middlewares
   - Route handlers (last)


## Error Handler Middleware

### Overview

The error handler middleware provides global error handling with panic recovery, standardized error responses, and comprehensive error logging.

### Basic Usage

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/trpg-solo-engine/backend/internal/middleware"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    router := gin.New()
    
    // Add error handling middleware (includes panic recovery)
    router.Use(middleware.ErrorHandlerMiddleware(logger))
    
    // Add 404 and 405 handlers
    router.NoRoute(middleware.NotFoundHandler(logger))
    router.HandleMethodNotAllowed = true
    router.NoMethod(middleware.MethodNotAllowedHandler(logger))
    
    // Your routes...
    router.Run(":8080")
}
```

### Features

- **Panic Recovery**: Catches panics and returns 500 errors with stack traces logged
- **Standardized Responses**: All errors return consistent JSON format
- **Automatic Status Codes**: Maps error codes to appropriate HTTP status codes
- **Detailed Logging**: Logs all errors with context (path, method, user, etc.)
- **GameError Support**: Special handling for domain-specific errors with codes and details

### Error Response Format

All errors return a standardized JSON response:

```json
{
  "success": false,
  "error": "错误消息",
  "code": "ERROR_CODE",
  "details": {
    "field": "value",
    "additional": "info"
  }
}
```

### HTTP Status Code Mapping

The middleware automatically maps error codes to HTTP status codes:

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `INVALID_INPUT` | 400 | Bad Request |
| `INVALID_ARC` | 400 | Bad Request |
| `INVALID_ACTION` | 400 | Bad Request |
| `INSUFFICIENT_QA` | 400 | Bad Request |
| `INSUFFICIENT_CHAOS` | 400 | Bad Request |
| `INVALID_PHASE` | 400 | Bad Request |
| `INVALID_STATE` | 400 | Bad Request |
| `NOT_FOUND` | 404 | Not Found |
| `ALREADY_EXISTS` | 409 | Conflict |
| `DATA_CORRUPTED` | 500 | Internal Server Error |
| `INTERNAL_ERROR` | 500 | Internal Server Error |
| `AI_SERVICE_ERROR` | 500 | Internal Server Error |

### Using in Handlers

#### Method 1: Using Helper Functions (Recommended)

```go
func CreateAgent(c *gin.Context) {
    var req CreateAgentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Abort with validation error
        middleware.AbortWithGameError(c, domain.ErrInvalidInput, 
            "请求参数无效: "+err.Error(), logger)
        return
    }
    
    agent, err := agentService.CreateAgent(&req)
    if err != nil {
        // Abort with service error
        middleware.AbortWithError(c, err, logger)
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data": agent,
    })
}
```

#### Method 2: Using c.Error()

```go
func GetAgent(c *gin.Context) {
    agentID := c.Param("id")
    
    agent, err := agentService.GetAgent(agentID)
    if err != nil {
        // Add error to context - middleware will handle it
        c.Error(err)
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": agent,
    })
}
```

#### Method 3: With Error Details

```go
func UpdateAgent(c *gin.Context) {
    var req UpdateAgentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        details := map[string]interface{}{
            "field": "name",
            "reason": "名称不能为空",
        }
        middleware.AbortWithGameErrorDetails(c, domain.ErrInvalidInput, 
            "验证失败", details, logger)
        return
    }
    
    // ... rest of handler
}
```

### Separate Middleware Components

You can also use individual middleware components:

```go
// Just panic recovery
router.Use(middleware.RecoveryMiddleware(logger))

// Just validation error handling
router.Use(middleware.ValidationErrorMiddleware(logger))
```

### Example: Complete Setup

```go
func setupRouter(logger *zap.Logger) *gin.Engine {
    router := gin.New()
    
    // 1. Recovery middleware (catches panics)
    router.Use(middleware.RecoveryMiddleware(logger))
    
    // 2. Logging middleware
    router.Use(middleware.LoggerMiddleware(logger))
    
    // 3. Error handler middleware
    router.Use(middleware.ErrorHandlerMiddleware(logger))
    
    // 4. 404 and 405 handlers
    router.NoRoute(middleware.NotFoundHandler(logger))
    router.HandleMethodNotAllowed = true
    router.NoMethod(middleware.MethodNotAllowedHandler(logger))
    
    // 5. Your routes
    api := router.Group("/api")
    {
        api.POST("/agents", CreateAgent)
        api.GET("/agents/:id", GetAgent)
    }
    
    return router
}
```

### Testing Error Handling

```go
func TestErrorHandling(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    router := gin.New()
    router.Use(middleware.ErrorHandlerMiddleware(logger))
    
    router.GET("/error", func(c *gin.Context) {
        err := domain.NewGameError(domain.ErrNotFound, "资源未找到")
        c.Error(err)
    })
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/error", nil)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusNotFound, w.Code)
    
    var response middleware.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.False(t, response.Success)
    assert.Equal(t, "资源未找到", response.Error)
}
```

### Best Practices

1. **Always use return after AbortWith* functions** to prevent further handler execution
2. **Use GameError for domain-specific errors** to get proper error codes and status codes
3. **Add details to errors** when they provide useful context for debugging
4. **Log errors at appropriate levels** - the middleware handles this automatically
5. **Don't expose sensitive information** in error messages or details
6. **Test error paths** to ensure proper error handling and responses

### Middleware Order

The recommended middleware order is:

```go
router.Use(gin.Recovery())                          // 1. Catch panics (optional if using ErrorHandlerMiddleware)
router.Use(middleware.LoggerMiddleware(logger))     // 2. Log requests
router.Use(middleware.ErrorHandlerMiddleware(logger)) // 3. Handle errors
router.Use(middleware.AuthMiddleware(secretKey))    // 4. Authenticate (for protected routes)
// ... your routes
```

## Rate Limiter Middleware

### Overview

The rate limiter middleware provides flexible rate limiting capabilities using Redis for distributed rate limiting. It supports IP-based, user-based, and combined rate limiting strategies with a sliding window algorithm.

### Basic Usage

```go
import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/trpg-solo-engine/backend/internal/middleware"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    redisClient := setupRedis() // Your Redis client
    
    router := gin.New()
    
    // IP-based rate limiting: 100 requests per minute
    config := middleware.RateLimitConfig{
        MaxRequests: 100,
        Window:      time.Minute,
        ByUser:      false, // IP-based
    }
    
    router.Use(middleware.RateLimitMiddleware(redisClient, logger, config))
    
    // Your routes...
    router.Run(":8080")
}
```

### Rate Limiting Strategies

#### 1. IP-Based Rate Limiting

Limits requests based on client IP address:

```go
config := middleware.RateLimitConfig{
    MaxRequests: 100,
    Window:      time.Minute,
    ByUser:      false,
}
```

#### 2. User-Based Rate Limiting

Limits requests based on authenticated user ID:

```go
config := middleware.RateLimitConfig{
    MaxRequests: 50,
    Window:      time.Minute,
    ByUser:      true,
}
```

**Note:** Requires authentication middleware to set `userID` in context. Falls back to IP-based if no user ID is present.

#### 3. Combined IP and User Rate Limiting

Enforces limits on both IP and user simultaneously:

```go
config := middleware.RateLimitConfig{
    MaxRequests: 100,
    Window:      time.Minute,
    ByBoth:      true,
}
```

### Endpoint-Specific Rate Limiting

Apply different rate limits to different endpoints:

```go
configs := map[string]middleware.RateLimitConfig{
    "/api/dice/*": {
        MaxRequests: 100,
        Window:      time.Minute,
        ByUser:      true,
    },
    "/api/ai/*": {
        MaxRequests: 10,
        Window:      time.Minute,
        ByUser:      true,
    },
    "/api/saves/*": {
        MaxRequests: 20,
        Window:      time.Minute,
        ByUser:      true,
    },
}

router.Use(middleware.RateLimitByEndpoint(redisClient, logger, configs))
```

### Default Rate Limit Configurations

The middleware provides sensible defaults for common endpoints:

```go
configs := middleware.DefaultRateLimitConfigs()
router.Use(middleware.RateLimitByEndpoint(redisClient, logger, configs))
```

**Default Limits:**
- `/api/dice/*`: 100 requests/minute (user-based)
- `/api/ai/*`: 10 requests/minute (user-based)
- `/api/saves/*`: 20 requests/minute (user-based)
- `/api/sessions/*`: 30 requests/minute (user-based)
- `/api/agents/*`: 20 requests/minute (user-based)
- `/api/scenarios/*`: 50 requests/minute (IP-based)

### Response Headers

The middleware adds standard rate limit headers to all responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1234567890
```

### Error Response

When rate limit is exceeded, returns HTTP 429:

```json
{
  "success": false,
  "error": "请求过于频繁，请稍后再试",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "limit": 100,
    "window": "1m0s",
    "reset_at": 1234567890,
    "remaining": 0
  }
}
```

### Sliding Window Algorithm

The rate limiter uses a sliding window algorithm with Redis sorted sets:

1. Removes expired entries outside the time window
2. Counts current requests in the window
3. Adds the new request with timestamp
4. Sets expiration on the key

This provides accurate rate limiting without the "burst" problem of fixed windows.

### Complete Example

```go
func setupRouter(logger *zap.Logger, redisClient *redis.Client) *gin.Engine {
    router := gin.New()
    
    // 1. Recovery and logging
    router.Use(middleware.RecoveryMiddleware(logger))
    router.Use(middleware.LoggerMiddleware(logger))
    
    // 2. Error handling
    router.Use(middleware.ErrorHandlerMiddleware(logger))
    
    // 3. Rate limiting (before authentication)
    configs := middleware.DefaultRateLimitConfigs()
    router.Use(middleware.RateLimitByEndpoint(redisClient, logger, configs))
    
    // 4. Authentication for protected routes
    api := router.Group("/api")
    api.Use(middleware.AuthMiddleware(secretKey))
    {
        api.POST("/agents", CreateAgent)
        api.GET("/agents/:id", GetAgent)
        api.POST("/dice/roll", RollDice)
        api.POST("/ai/generate", GenerateAI)
    }
    
    return router
}
```

### Testing

The rate limiter requires a Redis instance for testing:

```bash
# Start Redis for testing
docker run -d -p 6379:6379 redis:7-alpine

# Run tests
go test ./internal/middleware -run TestRateLimit
```

Tests will be skipped if Redis is not available.

### Configuration Best Practices

1. **Choose appropriate limits**: Balance user experience with server capacity
2. **Use user-based limits for authenticated endpoints**: Prevents abuse while allowing legitimate users
3. **Use IP-based limits for public endpoints**: Protects against DDoS
4. **Set shorter windows for expensive operations**: AI generation, database writes
5. **Set longer windows for cheap operations**: Static content, reads
6. **Monitor rate limit hits**: Track how often users hit limits to adjust accordingly

### Redis Requirements

- Redis 2.8.0+ (for sorted set operations)
- Sufficient memory for rate limit keys
- Consider using a separate Redis database for rate limiting (e.g., DB 1)
- Set appropriate maxmemory-policy (e.g., `allkeys-lru`)

### Performance Considerations

- Each request requires 4 Redis operations (pipeline)
- Redis operations are fast (~1ms)
- Use connection pooling for better performance
- Consider using Redis Cluster for high-traffic applications

### Fail-Open Behavior

If Redis is unavailable, the middleware allows requests through (fail-open) to prevent service disruption. Errors are logged for monitoring.

### Monitoring

Monitor these metrics:
- Rate limit hits (429 responses)
- Redis connection errors
- Average request latency
- Top rate-limited IPs/users

### Advanced Usage

#### Custom Key Generation

For custom rate limiting logic, use the `RateLimiter` directly:

```go
limiter := middleware.NewRateLimiter(redisClient, logger, config)
key := fmt.Sprintf("custom:rate_limit:%s", customID)
allowed, remaining, resetTime, err := limiter.Allow(ctx, key)
```

#### Dynamic Rate Limits

Adjust limits based on user tier or time of day:

```go
func DynamicRateLimitMiddleware(redisClient *redis.Client, logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        userTier := getUserTier(c)
        
        config := middleware.RateLimitConfig{
            MaxRequests: getTierLimit(userTier),
            Window:      time.Minute,
            ByUser:      true,
        }
        
        limiter := middleware.NewRateLimiter(redisClient, logger, config)
        // ... apply rate limiting
    }
}
```

