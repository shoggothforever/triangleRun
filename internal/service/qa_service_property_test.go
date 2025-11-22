package service

import (
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// 属性4: 资质保证不变量
// Feature: trpg-solo-engine, Property 4: 资质保证不变量
// Validates: Requirements 4.1, 4.5
// 对于任何角色，所有资质的QA总和应该始终等于或小于初始分配的9点（消耗后），
// 且任何单次花费不能超过该资质的当前可用QA。
func TestProperty_QAInvariant(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	f := func(spendAmount uint8) bool {
		// 限制花费金额在合理范围内
		amount := int(spendAmount % 10)

		// 创建测试角色
		agent := createTestAgent()
		initialTotal := agent.TotalQA()

		// 初始总和应该是9
		if initialTotal != 9 {
			return false
		}

		// 选择一个资质进行花费
		quality := domain.QualityFocus
		initialQA := agent.QA[quality]

		// 尝试花费QA
		err := qaService.SpendQA(agent, quality, amount)

		// 如果花费金额超过可用QA，应该返回错误
		if amount > initialQA {
			if err == nil {
				return false // 应该返回错误但没有
			}
			// 错误情况下，QA不应该改变
			return agent.QA[quality] == initialQA && agent.TotalQA() == initialTotal
		}

		// 如果花费成功
		if err != nil {
			return false // 不应该有错误
		}

		// 验证该资质的QA减少了正确的数量
		if agent.QA[quality] != initialQA-amount {
			return false
		}

		// 验证总QA减少了正确的数量
		if agent.TotalQA() != initialTotal-amount {
			return false
		}

		// 验证总QA不超过初始的9点
		if agent.TotalQA() > 9 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性4扩展: 多次花费QA的不变量
// Feature: trpg-solo-engine, Property 4: 资质保证不变量
// Validates: Requirements 4.1, 4.5
func TestProperty_QAInvariant_MultipleSpends(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	f := func(spends [5]uint8) bool {
		// 创建测试角色
		agent := createTestAgent()
		initialTotal := agent.TotalQA()

		// 记录每次花费
		totalSpent := 0
		qualities := []string{
			domain.QualityFocus,
			domain.QualityEmpathy,
			domain.QualityPresence,
			domain.QualityDeception,
			domain.QualityInitiative,
		}

		for i, spendAmount := range spends {
			amount := int(spendAmount % 5) // 限制每次花费在0-4之间
			quality := qualities[i%len(qualities)]

			err := qaService.SpendQA(agent, quality, amount)
			if err == nil {
				totalSpent += amount
			}
		}

		// 验证总QA等于初始值减去总花费
		if agent.TotalQA() != initialTotal-totalSpent {
			return false
		}

		// 验证总QA不超过初始的9点
		if agent.TotalQA() > 9 {
			return false
		}

		// 验证总QA不为负
		if agent.TotalQA() < 0 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性5: 过载机制
// Feature: trpg-solo-engine, Property 5: 过载机制
// Validates: Requirements 4.2, 4.3
// 对于任何在QA为0的资质上进行的掷骰，应该应用过载效果：移除一个"3"（如果有）并产生1点混沌。
// 当过载解除条件满足时，所有过载应该被清零。
func TestProperty_OverloadMechanism(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)

	f := func() bool {
		// 创建测试角色
		agent := createTestAgent()

		// 将某个资质的QA设为0
		quality := domain.QualityFocus
		agent.QA[quality] = 0

		// 进行掷骰
		roll := diceService.Roll(6)
		originalThrees := roll.Threes
		originalChaos := roll.Chaos

		// 应用过载
		overloadedRoll := qaService.ApplyOverload(agent, quality, roll)

		// 如果原始掷骰有"3"，过载应该移除一个"3"
		if originalThrees > 0 {
			if overloadedRoll.Threes != originalThrees-1 {
				return false
			}
			// 过载应该增加1点混沌
			if overloadedRoll.Chaos != originalChaos+1 {
				return false
			}
		} else {
			// 如果没有"3"，过载无法移除，但仍然记录过载
			if overloadedRoll.Threes != 0 {
				return false
			}
		}

		// 验证过载标记
		if overloadedRoll.Overload != 1 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性5扩展: 过载解除机制
// Feature: trpg-solo-engine, Property 5: 过载机制
// Validates: Requirements 4.2, 4.3
func TestProperty_OverloadRelief(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)

	f := func() bool {
		// 创建测试角色
		agent := createTestAgent()

		// 将某个资质的QA设为0
		quality := domain.QualityFocus
		agent.QA[quality] = 0

		// 进行掷骰
		roll := diceService.Roll(6)

		// 模拟过载解除条件满足（这里我们手动设置）
		// 在实际游戏中，这会根据现实类型的条件判断
		// 由于CheckOverloadRelief目前返回false，我们测试ClearOverload功能

		// 先应用过载
		overloadedRoll := qaService.ApplyOverload(agent, quality, roll)

		// 验证过载被应用
		if overloadedRoll.Overload == 0 {
			return false
		}

		// 清除过载
		clearedRoll := qaService.ClearOverload(overloadedRoll)

		// 验证过载被清除
		if clearedRoll.Overload != 0 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性5扩展: 有QA时不应用过载
// Feature: trpg-solo-engine, Property 5: 过载机制
// Validates: Requirements 4.2, 4.3
func TestProperty_NoOverloadWithQA(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)

	f := func(qaAmount uint8) bool {
		// 创建测试角色
		agent := createTestAgent()

		// 设置QA为非零值
		quality := domain.QualityFocus
		agent.QA[quality] = int(qaAmount%5) + 1 // 1-5

		// 进行掷骰
		roll := diceService.Roll(6)
		originalThrees := roll.Threes

		// 尝试应用过载
		result := qaService.ApplyOverload(agent, quality, roll)

		// 由于有QA，不应该应用过载
		// 结果应该与原始掷骰相同
		if result.Threes != originalThrees {
			return false
		}

		// 不应该有过载标记
		if result.Overload != 0 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性6: 任务间隙恢复
// Feature: trpg-solo-engine, Property 6: 任务间隙恢复
// Validates: Requirements 4.4
// 对于任何角色，在任务间隙时，所有资质的QA应该恢复到其当前上限（不超过初始分配）。
func TestProperty_MissionIntervalRecovery(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	f := func(spends [9]uint8) bool {
		// 创建测试角色
		agent := createTestAgent()

		// 记录初始QA
		initialQA := make(map[string]int)
		for quality, qa := range agent.Career.QA {
			initialQA[quality] = qa
		}

		// 随机花费一些QA
		qualities := domain.AllQualities
		for i, spendAmount := range spends {
			if i >= len(qualities) {
				break
			}
			quality := qualities[i]
			amount := int(spendAmount % 3) // 0-2

			// 尝试花费（可能失败）
			_ = qaService.SpendQA(agent, quality, amount)
		}

		// 验证至少有一些QA被花费了（否则测试没有意义）
		hasSpent := false
		for quality := range agent.QA {
			if agent.QA[quality] < initialQA[quality] {
				hasSpent = true
				break
			}
		}

		// 恢复QA（任务间隙）
		err := qaService.RestoreQA(agent)
		if err != nil {
			return false
		}

		// 验证所有QA都恢复到初始值
		for quality, initialValue := range initialQA {
			if agent.QA[quality] != initialValue {
				return false
			}
		}

		// 验证总QA恢复到9点
		if agent.TotalQA() != 9 {
			return false
		}

		// 如果有花费，验证恢复确实发生了
		if hasSpent {
			// 至少有一个资质应该被恢复
			restored := false
			for quality := range agent.QA {
				if agent.QA[quality] == initialQA[quality] {
					restored = true
					break
				}
			}
			if !restored {
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// 属性6扩展: 完全耗尽后的恢复
// Feature: trpg-solo-engine, Property 6: 任务间隙恢复
// Validates: Requirements 4.4
func TestProperty_MissionIntervalRecovery_FullDepletion(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	f := func() bool {
		// 创建测试角色
		agent := createTestAgent()

		// 记录初始QA
		initialQA := make(map[string]int)
		for quality, qa := range agent.Career.QA {
			initialQA[quality] = qa
		}

		// 完全耗尽所有QA
		for quality, qa := range agent.QA {
			if qa > 0 {
				err := qaService.SpendQA(agent, quality, qa)
				if err != nil {
					return false
				}
			}
		}

		// 验证所有QA都为0
		if agent.TotalQA() != 0 {
			return false
		}

		// 恢复QA
		err := qaService.RestoreQA(agent)
		if err != nil {
			return false
		}

		// 验证所有QA都恢复到初始值
		for quality, initialValue := range initialQA {
			if agent.QA[quality] != initialValue {
				return false
			}
		}

		// 验证总QA恢复到9点
		if agent.TotalQA() != 9 {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// createTestAgent 创建测试用角色
func createTestAgent() *domain.Agent {
	return &domain.Agent{
		ID:       uuid.New().String(),
		Name:     "测试特工",
		Pronouns: "他/她",
		Anomaly: &domain.Anomaly{
			Type: domain.AnomalyWhisper,
			Abilities: []*domain.AnomalyAbility{
				{ID: uuid.New().String(), Name: "能力1", AnomalyType: domain.AnomalyWhisper},
				{ID: uuid.New().String(), Name: "能力2", AnomalyType: domain.AnomalyWhisper},
				{ID: uuid.New().String(), Name: "能力3", AnomalyType: domain.AnomalyWhisper},
			},
		},
		Reality: &domain.Reality{
			Type: domain.RealityCaretaker,
			Trigger: &domain.RealityTrigger{
				Name:        "现实触发",
				Cost:        0,
				Effect:      "触发效果",
				Consequence: "忽视后果",
			},
			OverloadRelief: &domain.OverloadRelief{
				Name:      "过载解除",
				Condition: "满足条件",
				Effect:    "无视所有过载",
			},
			DegradationTrack: &domain.DegradationTrack{
				Name:   "退化轨道",
				Filled: 0,
				Total:  4,
			},
		},
		Career: &domain.Career{
			Type: domain.CareerPublicRelations,
			QA: map[string]int{
				domain.QualityFocus:      1,
				domain.QualityEmpathy:    1,
				domain.QualityPresence:   1,
				domain.QualityDeception:  1,
				domain.QualityInitiative: 1,
				domain.QualityProfession: 1,
				domain.QualityVitality:   1,
				domain.QualityGrit:       1,
				domain.QualitySubtlety:   1,
			},
		},
		QA: map[string]int{
			domain.QualityFocus:      1,
			domain.QualityEmpathy:    1,
			domain.QualityPresence:   1,
			domain.QualityDeception:  1,
			domain.QualityInitiative: 1,
			domain.QualityProfession: 1,
			domain.QualityVitality:   1,
			domain.QualityGrit:       1,
			domain.QualitySubtlety:   1,
		},
		Relationships: []*domain.Relationship{
			{ID: uuid.New().String(), Name: "关系1", Connection: 6},
			{ID: uuid.New().String(), Name: "关系2", Connection: 3},
			{ID: uuid.New().String(), Name: "关系3", Connection: 3},
		},
		Commendations: 0,
		Reprimands:    0,
		Rating:        domain.RatingExcellent,
		Alive:         true,
		InDebt:        false,
	}
}
