package domain

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// Feature: trpg-solo-engine, Property 2: 骰子判定一致性
// 验证需求: 2.1, 2.2, 2.3
func TestProperty_DiceRollConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)
	service := NewDiceService()

	properties.Property("骰子判定一致性", prop.ForAll(
		func() bool {
			roll := service.Roll(6)

			// 验证骰子数量
			if len(roll.Dice) != 6 {
				return false
			}

			// 验证骰子范围（1-4）
			for _, die := range roll.Dice {
				if die < 1 || die > 4 {
					return false
				}
			}

			// 验证"3"的统计
			actualThrees := CountThrees(roll.Dice)
			if roll.Threes != actualThrees {
				return false
			}

			// 验证成功判定
			if (roll.Threes > 0) != roll.Success {
				return false
			}

			// 验证混沌生成（失败时）
			if !roll.Success {
				expectedChaos := 6 - roll.Threes
				if roll.Chaos != expectedChaos {
					return false
				}
			}

			return true
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: trpg-solo-engine, Property 3: 三重升华零混沌
// 验证需求: 2.4
func TestProperty_TripleAscensionZeroChaos(t *testing.T) {
	properties := gopter.NewProperties(nil)
	service := NewDiceService()

	// 模拟三重升华场景
	properties.Property("三重升华零混沌", prop.ForAll(
		func() bool {
			// 多次掷骰，寻找三重升华
			for i := 0; i < 100; i++ {
				roll := service.Roll(6)

				// 如果是三重升华
				if roll.TripleAsc {
					// 验证混沌为0
					if roll.Chaos != 0 {
						t.Logf("三重升华时混沌不为0: chaos=%d", roll.Chaos)
						return false
					}

					// 验证恰好3个"3"
					if roll.Threes != 3 {
						t.Logf("三重升华时'3'的数量不是3: threes=%d", roll.Threes)
						return false
					}

					return true
				}
			}

			// 如果100次都没有三重升华，跳过此次测试
			return true
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
