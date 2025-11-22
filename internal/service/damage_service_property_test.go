package service

import (
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// **Feature: trpg-solo-engine, Property 11: 人寿保险机制**
// **Validates: Requirements 8.1, 8.2, 8.3**
//
// 属性11: 人寿保险机制
// 对于任何伤害，玩家可以花费等量QA（任意资质）来无视伤害。
// 如果无法或不愿花费，玩家死亡并扣除5次嘉奖，然后在休息室复活。
func TestProperty_LifeInsuranceMechanism(t *testing.T) {
	damageService := NewDamageService()

	// 测试1: 有足够QA时可以使用人寿保险无视伤害
	t.Run("CanUseInsuranceWithSufficientQA", func(t *testing.T) {
		f := func(damage uint8) bool {
			// 限制伤害在合理范围内
			dmg := int(damage%10) + 1 // 1-10点伤害

			agent := createTestAgentForDamage()
			initialQA := agent.TotalQA()

			// 确保有足够的QA
			if initialQA < dmg {
				return true // 跳过这个测试用例
			}

			// 使用人寿保险
			err := damageService.UseLifeInsurance(agent, dmg)
			if err != nil {
				return false
			}

			// 验证QA减少了等量
			if agent.TotalQA() != initialQA-dmg {
				return false
			}

			// 验证角色仍然存活
			if !agent.Alive {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: QA不足时无法使用人寿保险
	t.Run("CannotUseInsuranceWithInsufficientQA", func(t *testing.T) {
		f := func(extraDamage uint8) bool {
			agent := createTestAgentForDamage()
			totalQA := agent.TotalQA()

			// 伤害超过总QA
			damage := totalQA + int(extraDamage%10) + 1

			// 尝试使用人寿保险
			err := damageService.UseLifeInsurance(agent, damage)

			// 应该返回错误
			if err == nil {
				return false
			}

			// QA不应该改变
			if agent.TotalQA() != totalQA {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 死亡时扣除5次嘉奖
	t.Run("DeathDeducts5Commendations", func(t *testing.T) {
		f := func(initialCommendations uint8) bool {
			agent := createTestAgentForDamage()
			agent.Commendations = int(initialCommendations % 20) // 0-19

			initialComm := agent.Commendations

			// 处理死亡
			err := damageService.HandleDeath(agent)
			if err != nil {
				return false
			}

			// 验证扣除了5次嘉奖
			if agent.Commendations != initialComm-5 {
				return false
			}

			// 验证角色复活
			if !agent.Alive {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 嘉奖为负时进入负债状态
	t.Run("NegativeCommendationsEntersDebt", func(t *testing.T) {
		f := func(initialCommendations uint8) bool {
			agent := createTestAgentForDamage()
			// 设置较少的嘉奖，确保死亡后会变负
			agent.Commendations = int(initialCommendations % 5) // 0-4

			// 处理死亡
			err := damageService.HandleDeath(agent)
			if err != nil {
				return false
			}

			// 如果嘉奖为负，应该进入负债状态
			if agent.Commendations < 0 {
				if !agent.InDebt {
					return false
				}
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: 复活后保留记忆（状态不变）
	t.Run("RevivePreservesState", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForDamage()
			agent.Commendations = 10
			agent.Reprimands = 2

			initialComm := agent.Commendations
			initialRep := agent.Reprimands

			// 标记为死亡
			agent.Alive = false

			// 复活
			err := damageService.Revive(agent)
			if err != nil {
				return false
			}

			// 验证复活
			if !agent.Alive {
				return false
			}

			// 验证其他状态保留（除了Alive标志）
			if agent.Commendations != initialComm {
				return false
			}
			if agent.Reprimands != initialRep {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试6: ApplyDamage集成测试 - 有足够QA
	t.Run("ApplyDamageWithSufficientQA", func(t *testing.T) {
		f := func(damage uint8) bool {
			dmg := int(damage%5) + 1 // 1-5点伤害

			agent := createTestAgentForDamage()
			initialQA := agent.TotalQA()

			// 确保有足够的QA
			if initialQA < dmg {
				return true // 跳过
			}

			died, usedInsurance, looseEnds, err := damageService.ApplyDamage(agent, dmg, true)
			if err != nil {
				return false
			}

			// 应该使用保险，不死亡
			if died {
				return false
			}
			if !usedInsurance {
				return false
			}
			if looseEnds != 0 {
				return false
			}

			// QA应该减少
			if agent.TotalQA() != initialQA-dmg {
				return false
			}

			// 应该存活
			if !agent.Alive {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试7: ApplyDamage集成测试 - QA不足
	t.Run("ApplyDamageWithInsufficientQA", func(t *testing.T) {
		f := func(extraDamage uint8) bool {
			agent := createTestAgentForDamage()
			totalQA := agent.TotalQA()
			initialComm := agent.Commendations

			// 伤害超过总QA
			damage := totalQA + int(extraDamage%10) + 1

			died, usedInsurance, _, err := damageService.ApplyDamage(agent, damage, false)
			if err != nil {
				return false
			}

			// 应该死亡，不使用保险
			if !died {
				return false
			}
			if usedInsurance {
				return false
			}

			// 应该扣除5次嘉奖
			if agent.Commendations != initialComm-5 {
				return false
			}

			// 应该复活
			if !agent.Alive {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// **Feature: trpg-solo-engine, Property 12: 伤害与散逸端**
// **Validates: Requirements 8.4**
//
// 属性12: 伤害与散逸端
// 对于任何超过1点的伤害，如果有目击者，应该产生等于伤害点数的散逸端。
func TestProperty_DamageAndLooseEnds(t *testing.T) {
	damageService := NewDamageService()

	// 测试1: 伤害>1且有目击者时产生散逸端
	t.Run("DamageAbove1WithWitnessesGeneratesLooseEnds", func(t *testing.T) {
		f := func(damage uint8) bool {
			dmg := int(damage%20) + 2 // 2-21点伤害（确保>1）

			looseEnds := damageService.GenerateLooseEnds(dmg, true)

			// 应该产生等于伤害点数的散逸端
			return looseEnds == dmg
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 伤害>1但无目击者时不产生散逸端
	t.Run("DamageAbove1WithoutWitnessesNoLooseEnds", func(t *testing.T) {
		f := func(damage uint8) bool {
			dmg := int(damage%20) + 2 // 2-21点伤害

			looseEnds := damageService.GenerateLooseEnds(dmg, false)

			// 没有目击者，不产生散逸端
			return looseEnds == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 伤害=1时不产生散逸端（即使有目击者）
	t.Run("Damage1NoLooseEnds", func(t *testing.T) {
		f := func(hasWitnesses bool) bool {
			looseEnds := damageService.GenerateLooseEnds(1, hasWitnesses)

			// 伤害为1，不产生散逸端
			return looseEnds == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 伤害=0时不产生散逸端
	t.Run("Damage0NoLooseEnds", func(t *testing.T) {
		f := func(hasWitnesses bool) bool {
			looseEnds := damageService.GenerateLooseEnds(0, hasWitnesses)

			// 伤害为0，不产生散逸端
			return looseEnds == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: 负伤害不产生散逸端
	t.Run("NegativeDamageNoLooseEnds", func(t *testing.T) {
		f := func(damage uint8, hasWitnesses bool) bool {
			dmg := -int(damage%20) - 1 // 负伤害

			looseEnds := damageService.GenerateLooseEnds(dmg, hasWitnesses)

			// 负伤害不产生散逸端
			return looseEnds == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试6: ApplyDamage集成测试 - 验证散逸端生成
	t.Run("ApplyDamageGeneratesLooseEnds", func(t *testing.T) {
		f := func(damage uint8) bool {
			dmg := int(damage%10) + 2 // 2-11点伤害

			agent := createTestAgentForDamage()
			// 设置QA为0，确保会死亡
			for quality := range agent.QA {
				agent.QA[quality] = 0
			}

			died, _, looseEnds, err := damageService.ApplyDamage(agent, dmg, true)
			if err != nil {
				return false
			}

			// 应该死亡
			if !died {
				return false
			}

			// 应该产生等于伤害点数的散逸端
			if looseEnds != dmg {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试7: ApplyDamage集成测试 - 无目击者不产生散逸端
	t.Run("ApplyDamageNoWitnessesNoLooseEnds", func(t *testing.T) {
		f := func(damage uint8) bool {
			dmg := int(damage%10) + 2 // 2-11点伤害

			agent := createTestAgentForDamage()
			// 设置QA为0，确保会死亡
			for quality := range agent.QA {
				agent.QA[quality] = 0
			}

			died, _, looseEnds, err := damageService.ApplyDamage(agent, dmg, false)
			if err != nil {
				return false
			}

			// 应该死亡
			if !died {
				return false
			}

			// 没有目击者，不产生散逸端
			if looseEnds != 0 {
				return false
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// createTestAgentForDamage 创建测试用角色
func createTestAgentForDamage() *domain.Agent {
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
		Commendations: 10,
		Reprimands:    0,
		Rating:        domain.RatingExcellent,
		Alive:         true,
		InDebt:        false,
	}
}

// 单元测试：验证基本功能
func TestDamageService_BasicFunctionality(t *testing.T) {
	damageService := NewDamageService()

	t.Run("CanAffordInsurance", func(t *testing.T) {
		agent := createTestAgentForDamage()

		// 可以支付5点伤害
		assert.True(t, damageService.CanAffordInsurance(agent, 5))

		// 可以支付9点伤害（总QA）
		assert.True(t, damageService.CanAffordInsurance(agent, 9))

		// 不能支付10点伤害
		assert.False(t, damageService.CanAffordInsurance(agent, 10))
	})

	t.Run("CheckDebt", func(t *testing.T) {
		agent := createTestAgentForDamage()

		// 初始不在负债状态
		assert.False(t, damageService.CheckDebt(agent))

		// 设置为负数
		agent.Commendations = -1
		assert.True(t, damageService.CheckDebt(agent))
	})

	t.Run("UseLifeInsurance", func(t *testing.T) {
		agent := createTestAgentForDamage()
		initialQA := agent.TotalQA()

		err := damageService.UseLifeInsurance(agent, 3)
		assert.NoError(t, err)
		assert.Equal(t, initialQA-3, agent.TotalQA())
	})

	t.Run("HandleDeath", func(t *testing.T) {
		agent := createTestAgentForDamage()
		agent.Commendations = 10

		err := damageService.HandleDeath(agent)
		assert.NoError(t, err)
		assert.Equal(t, 5, agent.Commendations)
		assert.True(t, agent.Alive) // 应该复活
	})

	t.Run("GenerateLooseEnds", func(t *testing.T) {
		// 伤害>1，有目击者
		assert.Equal(t, 5, damageService.GenerateLooseEnds(5, true))

		// 伤害>1，无目击者
		assert.Equal(t, 0, damageService.GenerateLooseEnds(5, false))

		// 伤害=1
		assert.Equal(t, 0, damageService.GenerateLooseEnds(1, true))

		// 伤害=0
		assert.Equal(t, 0, damageService.GenerateLooseEnds(0, true))
	})
}
