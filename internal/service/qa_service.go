package service

import (
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// QAService 资质保证服务接口
type QAService interface {
	// QA管理
	SpendQA(agent *domain.Agent, quality string, amount int) error
	RestoreQA(agent *domain.Agent) error
	GetAvailableQA(agent *domain.Agent, quality string) int
	GetTotalQA(agent *domain.Agent) int

	// 过载机制
	CheckOverload(agent *domain.Agent, quality string) bool
	ApplyOverload(agent *domain.Agent, quality string, roll *domain.RollResult) *domain.RollResult
	CheckOverloadRelief(agent *domain.Agent) bool
	ClearOverload(roll *domain.RollResult) *domain.RollResult

	// 骰子调整
	AdjustDiceWithQA(agent *domain.Agent, quality string, roll *domain.RollResult, adjustments []DiceAdjustment) (*domain.RollResult, error)
}

// DiceAdjustment 骰子调整
type DiceAdjustment struct {
	DiceIndex int // 要调整的骰子索引
	NewValue  int // 新值（1-4）
}

// qaService 资质保证服务实现
type qaService struct {
	diceService domain.DiceService
}

// NewQAService 创建资质保证服务
func NewQAService(diceService domain.DiceService) QAService {
	return &qaService{
		diceService: diceService,
	}
}

// SpendQA 花费资质保证
func (s *qaService) SpendQA(agent *domain.Agent, quality string, amount int) error {
	return agent.SpendQA(quality, amount)
}

// RestoreQA 恢复资质保证（任务间隙）
func (s *qaService) RestoreQA(agent *domain.Agent) error {
	agent.RestoreQA()
	return nil
}

// GetAvailableQA 获取可用的资质保证点数
func (s *qaService) GetAvailableQA(agent *domain.Agent, quality string) int {
	if qa, exists := agent.QA[quality]; exists {
		return qa
	}
	return 0
}

// GetTotalQA 获取总资质保证点数
func (s *qaService) GetTotalQA(agent *domain.Agent) int {
	return agent.TotalQA()
}

// CheckOverload 检查是否需要应用过载
func (s *qaService) CheckOverload(agent *domain.Agent, quality string) bool {
	return agent.QA[quality] == 0
}

// ApplyOverload 应用过载效果
// 过载效果：移除一个"3"并产生1点混沌
func (s *qaService) ApplyOverload(agent *domain.Agent, quality string, roll *domain.RollResult) *domain.RollResult {
	// 检查是否真的需要过载
	if !s.CheckOverload(agent, quality) {
		return roll
	}

	// 检查是否有过载解除
	if s.CheckOverloadRelief(agent) {
		return roll
	}

	// 应用过载
	return s.diceService.ApplyOverload(roll, 1)
}

// CheckOverloadRelief 检查过载解除条件
func (s *qaService) CheckOverloadRelief(agent *domain.Agent) bool {
	// 这里简化实现，实际应该根据现实类型的过载解除条件判断
	// 不同的现实类型有不同的过载解除条件
	if agent.Reality == nil || agent.Reality.OverloadRelief == nil {
		return false
	}

	// 实际游戏中，这里应该检查具体的条件
	// 例如："当你为他人承担风险时"、"当你完成一个重要的个人目标时"等
	// 这需要游戏状态的上下文信息
	// 目前返回false，表示没有激活过载解除
	return false
}

// ClearOverload 清除过载效果（当过载解除条件满足时）
func (s *qaService) ClearOverload(roll *domain.RollResult) *domain.RollResult {
	// 创建新的结果，清除过载
	newRoll := &domain.RollResult{
		Dice:      make([]int, len(roll.Dice)),
		Threes:    roll.Threes,
		Success:   roll.Success,
		Chaos:     roll.Chaos,
		Overload:  0, // 清除过载
		TripleAsc: roll.TripleAsc,
	}
	copy(newRoll.Dice, roll.Dice)

	// 如果之前有过载效果，需要恢复
	// 这里简化处理，实际应该记录过载前的状态
	return newRoll
}

// AdjustDiceWithQA 使用资质保证调整骰子
func (s *qaService) AdjustDiceWithQA(agent *domain.Agent, quality string, roll *domain.RollResult, adjustments []DiceAdjustment) (*domain.RollResult, error) {
	// 检查是否有足够的QA
	if len(adjustments) > agent.QA[quality] {
		return nil, domain.NewGameError(domain.ErrInsufficientQA, "资质保证不足").
			WithDetails("quality", quality).
			WithDetails("available", agent.QA[quality]).
			WithDetails("required", len(adjustments))
	}

	// 花费QA
	if err := s.SpendQA(agent, quality, len(adjustments)); err != nil {
		return nil, err
	}

	// 创建新的结果
	newRoll := &domain.RollResult{
		Dice:      make([]int, len(roll.Dice)),
		Threes:    0,
		Success:   false,
		Chaos:     0,
		Overload:  roll.Overload,
		TripleAsc: roll.TripleAsc,
	}
	copy(newRoll.Dice, roll.Dice)

	// 应用调整
	for _, adj := range adjustments {
		if adj.DiceIndex < 0 || adj.DiceIndex >= len(newRoll.Dice) {
			continue
		}
		if adj.NewValue < 1 || adj.NewValue > 4 {
			continue
		}
		newRoll.Dice[adj.DiceIndex] = adj.NewValue
	}

	// 重新计算"3"的数量
	for _, d := range newRoll.Dice {
		if d == 3 {
			newRoll.Threes++
		}
	}

	// 重新判定成功
	newRoll.Success = newRoll.Threes > 0

	// 重新计算混沌
	if !newRoll.Success {
		newRoll.Chaos = len(newRoll.Dice) - newRoll.Threes
	} else {
		newRoll.Chaos = 0
	}

	// 三重升华时不产生混沌
	if newRoll.TripleAsc {
		newRoll.Chaos = 0
	}

	return newRoll, nil
}
