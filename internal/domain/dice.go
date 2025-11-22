package domain

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RollResult 掷骰结果
type RollResult struct {
	Dice      []int `json:"dice"`       // 骰子结果
	Threes    int   `json:"threes"`     // "3"的数量
	Success   bool  `json:"success"`    // 是否成功
	Chaos     int   `json:"chaos"`      // 产生的混沌
	Overload  int   `json:"overload"`   // 过载点数
	TripleAsc bool  `json:"triple_asc"` // 三重升华
}

// DiceService 骰子服务接口
type DiceService interface {
	Roll(count int) *RollResult
	RollForAbility(agent *Agent, ability *AnomalyAbility) *RollResult
	RollForQuality(agent *Agent, quality string) *RollResult
	ApplyQA(roll *RollResult, quality string, amount int) *RollResult
	ApplyOverload(roll *RollResult, amount int) *RollResult
	CheckTripleAscension(roll *RollResult) bool
}

// diceService 骰子服务实现
type diceService struct{}

// NewDiceService 创建骰子服务
func NewDiceService() DiceService {
	return &diceService{}
}

// Roll 基础掷骰（6d4）
func (s *diceService) Roll(count int) *RollResult {
	if count <= 0 {
		count = 6
	}

	dice := make([]int, count)
	threes := 0

	for i := 0; i < count; i++ {
		dice[i] = rand.Intn(4) + 1 // 1-4
		if dice[i] == 3 {
			threes++
		}
	}

	// 检查三重升华（调整前恰好3个"3"）
	tripleAsc := (threes == 3)

	// 判定成功/失败
	success := threes > 0

	// 计算混沌（失败时每颗非"3"骰子产生1点混沌）
	chaos := 0
	if !success {
		chaos = count - threes
	}

	// 三重升华时不产生混沌
	if tripleAsc {
		chaos = 0
	}

	return &RollResult{
		Dice:      dice,
		Threes:    threes,
		Success:   success,
		Chaos:     chaos,
		Overload:  0,
		TripleAsc: tripleAsc,
	}
}

// RollForAbility 为异常能力掷骰
func (s *diceService) RollForAbility(agent *Agent, ability *AnomalyAbility) *RollResult {
	if ability.Roll == nil {
		return s.Roll(6)
	}

	roll := s.Roll(ability.Roll.DiceCount)

	// 检查是否需要应用过载
	quality := ability.Roll.Quality
	if agent.QA[quality] == 0 {
		roll = s.ApplyOverload(roll, 1)
	}

	return roll
}

// RollForQuality 为特定资质掷骰
func (s *diceService) RollForQuality(agent *Agent, quality string) *RollResult {
	roll := s.Roll(6)

	// 检查是否需要应用过载
	if agent.QA[quality] == 0 {
		roll = s.ApplyOverload(roll, 1)
	}

	return roll
}

// ApplyQA 应用资质保证调整
func (s *diceService) ApplyQA(roll *RollResult, quality string, amount int) *RollResult {
	// QA可以将任意骰子调整为"3"或从"3"调整为其他数字
	// 这里简化实现：假设总是将骰子调整为"3"
	for i := 0; i < amount && i < len(roll.Dice); i++ {
		if roll.Dice[i] != 3 {
			roll.Dice[i] = 3
			roll.Threes++
		}
	}

	// 重新计算成功和混沌
	roll.Success = roll.Threes > 0
	if !roll.Success {
		roll.Chaos = len(roll.Dice) - roll.Threes
	} else {
		roll.Chaos = 0
	}

	return roll
}

// ApplyOverload 应用过载效果
func (s *diceService) ApplyOverload(roll *RollResult, amount int) *RollResult {
	roll.Overload += amount

	// 过载效果：移除一个"3"并产生1点混沌
	for i := 0; i < amount && roll.Threes > 0; i++ {
		// 找到第一个"3"并改为其他数字
		for j := 0; j < len(roll.Dice); j++ {
			if roll.Dice[j] == 3 {
				roll.Dice[j] = 1 // 改为1
				roll.Threes--
				roll.Chaos++
				break
			}
		}
	}

	// 重新判定成功
	roll.Success = roll.Threes > 0

	return roll
}

// CheckTripleAscension 检查三重升华
func (s *diceService) CheckTripleAscension(roll *RollResult) bool {
	return roll.TripleAsc
}

// CountThrees 统计"3"的数量
func CountThrees(dice []int) int {
	count := 0
	for _, d := range dice {
		if d == 3 {
			count++
		}
	}
	return count
}
