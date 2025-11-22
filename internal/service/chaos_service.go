package service

import (
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// ChaosService 混沌服务接口
type ChaosService interface {
	// 混沌池管理
	InitializeChaosPool(session *domain.GameSession, looseEnds int) error
	GetChaosPool(session *domain.GameSession) int
	ClearChaosPool(session *domain.GameSession) error

	// 混沌生成和消耗
	AddChaos(session *domain.GameSession, amount int) error
	AddChaosFromRoll(session *domain.GameSession, roll *domain.RollResult) error
	SpendChaos(session *domain.GameSession, amount int) error

	// 地点过载追踪
	AddLocationOverload(session *domain.GameSession, locationID string) error
	GetLocationOverload(session *domain.GameSession, locationID string) int
	ClearLocationOverload(session *domain.GameSession, locationID string) error
}

// chaosService 混沌服务实现
type chaosService struct{}

// NewChaosService 创建混沌服务
func NewChaosService() ChaosService {
	return &chaosService{}
}

// InitializeChaosPool 初始化混沌池（任务开始时）
// 根据累积的散逸端数量初始化混沌池
func (s *chaosService) InitializeChaosPool(session *domain.GameSession, looseEnds int) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	session.State.ChaosPool = looseEnds
	session.State.LooseEnds = looseEnds

	return nil
}

// GetChaosPool 获取当前混沌池数量
func (s *chaosService) GetChaosPool(session *domain.GameSession) int {
	if session == nil || session.State == nil {
		return 0
	}
	return session.State.ChaosPool
}

// ClearChaosPool 清空混沌池（任务结束时）
func (s *chaosService) ClearChaosPool(session *domain.GameSession) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	session.State.ChaosPool = 0
	return nil
}

// AddChaos 向混沌池添加混沌
func (s *chaosService) AddChaos(session *domain.GameSession, amount int) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	if amount < 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "混沌数量不能为负数").
			WithDetails("amount", amount)
	}

	session.State.ChaosPool += amount
	return nil
}

// AddChaosFromRoll 从掷骰结果添加混沌
// 失败时每颗非"3"骰子产生1点混沌
func (s *chaosService) AddChaosFromRoll(session *domain.GameSession, roll *domain.RollResult) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	if roll == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "掷骰结果为空")
	}

	// 只有失败时才产生混沌
	if !roll.Success && roll.Chaos > 0 {
		return s.AddChaos(session, roll.Chaos)
	}

	return nil
}

// SpendChaos 从混沌池中消耗混沌（异常体使用效应时）
func (s *chaosService) SpendChaos(session *domain.GameSession, amount int) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	if amount < 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "混沌数量不能为负数").
			WithDetails("amount", amount)
	}

	if session.State.ChaosPool < amount {
		return domain.NewGameError(domain.ErrInsufficientChaos, "混沌池混沌不足").
			WithDetails("available", session.State.ChaosPool).
			WithDetails("required", amount)
	}

	session.State.ChaosPool -= amount
	return nil
}

// AddLocationOverload 为地点添加过载（请求机构失败时）
func (s *chaosService) AddLocationOverload(session *domain.GameSession, locationID string) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	if locationID == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "地点ID不能为空")
	}

	// 初始化地点过载映射
	if session.State.LocationOverloads == nil {
		session.State.LocationOverloads = make(map[string]int)
	}

	session.State.LocationOverloads[locationID]++
	return nil
}

// GetLocationOverload 获取地点的过载数量
func (s *chaosService) GetLocationOverload(session *domain.GameSession, locationID string) int {
	if session == nil || session.State == nil || session.State.LocationOverloads == nil {
		return 0
	}

	return session.State.LocationOverloads[locationID]
}

// ClearLocationOverload 清除地点过载（离开地点时）
func (s *chaosService) ClearLocationOverload(session *domain.GameSession, locationID string) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	if session.State.LocationOverloads != nil {
		delete(session.State.LocationOverloads, locationID)
	}

	return nil
}
