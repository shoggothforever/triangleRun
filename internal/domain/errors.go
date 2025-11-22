package domain

import "fmt"

// ErrorCode 错误代码
type ErrorCode string

const (
	// 验证错误
	ErrInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrInvalidARC    ErrorCode = "INVALID_ARC"
	ErrInvalidAction ErrorCode = "INVALID_ACTION"

	// 资源错误
	ErrInsufficientQA    ErrorCode = "INSUFFICIENT_QA"
	ErrInsufficientChaos ErrorCode = "INSUFFICIENT_CHAOS"

	// 状态错误
	ErrInvalidPhase ErrorCode = "INVALID_PHASE"
	ErrInvalidState ErrorCode = "INVALID_STATE"

	// 数据错误
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrDataCorrupted ErrorCode = "DATA_CORRUPTED"

	// 系统错误
	ErrInternal  ErrorCode = "INTERNAL_ERROR"
	ErrAIService ErrorCode = "AI_SERVICE_ERROR"
)

// GameError 游戏错误
type GameError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *GameError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewGameError 创建游戏错误
func NewGameError(code ErrorCode, message string) *GameError {
	return &GameError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails 添加错误详情
func (e *GameError) WithDetails(key string, value interface{}) *GameError {
	e.Details[key] = value
	return e
}
