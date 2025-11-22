package service

import (
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// **Feature: trpg-solo-engine, Property 10: 异常能力效果一致性**
// **Validates: Requirements 7.2, 7.3, 7.4**
//
// 属性10: 异常能力效果一致性
// 对于任何异常能力的使用，成功时应用"成功时"效果，
// 失败时应用"失败时"效果并产生混沌，
// 满足额外条件时应用对应的额外效果
func TestProperty_AbilityEffectConsistency(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)
	chaosService := NewChaosService()
	abilityService := NewAbilityService(diceService, qaService, chaosService)

	// 测试1: 成功时应用成功效果
	t.Run("SuccessAppliesSuccessEffect", func(t *testing.T) {
		f := func(threes uint8) bool {
			// 确保至少有1个3（成功）
			threesCount := int(threes%5) + 1

			ability := createTestAbilityForTest()

			// 创建成功的掷骰结果
			roll := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    threesCount,
				Success:   true,
				Chaos:     0,
				Overload:  0,
				TripleAsc: false,
			}

			// 手动应用效果（模拟UseAbility的部分逻辑）
			successEffect, err := abilityService.ApplySuccessEffect(ability, roll, &AbilityContext{OnDuty: true})
			if err != nil {
				return false
			}

			// 验证成功效果被应用
			return successEffect != nil && successEffect.Applied
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 失败时应用失败效果并产生混沌
	t.Run("FailureAppliesFailureEffectAndGeneratesChaos", func(t *testing.T) {
		f := func(nonThrees uint8) bool {
			// 确保没有3（失败）
			nonThreesCount := int(nonThrees%6) + 1

			session := createTestSessionForAbility()
			ability := createTestAbilityForTest()

			// 创建失败的掷骰结果
			roll := &domain.RollResult{
				Dice:      make([]int, nonThreesCount),
				Threes:    0,
				Success:   false,
				Chaos:     nonThreesCount,
				Overload:  0,
				TripleAsc: false,
			}

			// 应用失败效果
			failureEffect, err := abilityService.ApplyFailureEffect(ability, roll, &AbilityContext{OnDuty: true})
			if err != nil {
				return false
			}

			// 添加混沌到混沌池
			initialChaos := session.State.ChaosPool
			err = chaosService.AddChaosFromRoll(session, roll)
			if err != nil {
				return false
			}

			// 验证失败效果被应用且混沌增加
			return failureEffect != nil &&
				failureEffect.Applied &&
				session.State.ChaosPool == initialChaos+nonThreesCount
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 成功时不产生混沌
	t.Run("SuccessDoesNotGenerateChaos", func(t *testing.T) {
		f := func(threes uint8) bool {
			threesCount := int(threes%5) + 1

			session := createTestSessionForAbility()
			initialChaos := session.State.ChaosPool

			// 创建成功的掷骰结果
			roll := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    threesCount,
				Success:   true,
				Chaos:     0,
				Overload:  0,
				TripleAsc: false,
			}

			// 尝试添加混沌
			err := chaosService.AddChaosFromRoll(session, roll)
			if err != nil {
				return false
			}

			// 验证混沌池没有变化
			return session.State.ChaosPool == initialChaos
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 满足额外条件时应用额外效果
	t.Run("AdditionalConditionAppliesAdditionalEffect", func(t *testing.T) {
		// 测试"每额外一个3"条件
		t.Run("ExtraThrees", func(t *testing.T) {
			f := func(threes uint8) bool {
				threesCount := int(threes%5) + 2 // 至少2个3

				ability := createTestAbilityWithAdditionalForTest("每额外一个3")

				roll := &domain.RollResult{
					Dice:      make([]int, 6),
					Threes:    threesCount,
					Success:   true,
					Chaos:     0,
					Overload:  0,
					TripleAsc: false,
				}

				additionalEffects, err := abilityService.ApplyAdditionalEffects(ability, roll, &AbilityContext{OnDuty: true})
				if err != nil {
					return false
				}

				// 如果有超过1个3，应该有额外效果
				if threesCount > 1 {
					return len(additionalEffects) > 0 && additionalEffects[0].Applied
				}

				// 否则不应该有额外效果
				return len(additionalEffects) == 0
			}

			if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
				t.Error(err)
			}
		})

		// 测试"六个或更多3"条件
		t.Run("SixOrMoreThrees", func(t *testing.T) {
			ability := createTestAbilityWithAdditionalForTest("六个或更多3")

			// 测试6个3
			roll6 := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    6,
				Success:   true,
				Chaos:     0,
				Overload:  0,
				TripleAsc: false,
			}

			effects6, err := abilityService.ApplyAdditionalEffects(ability, roll6, &AbilityContext{OnDuty: true})
			assert.NoError(t, err)
			assert.NotEmpty(t, effects6)
			assert.True(t, effects6[0].Applied)

			// 测试5个3
			roll5 := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    5,
				Success:   true,
				Chaos:     0,
				Overload:  0,
				TripleAsc: false,
			}

			effects5, err := abilityService.ApplyAdditionalEffects(ability, roll5, &AbilityContext{OnDuty: true})
			assert.NoError(t, err)
			assert.Empty(t, effects5)
		})
	})

	// 测试5: 效果一致性 - 同样的掷骰结果应该产生同样的效果
	t.Run("ConsistentEffectsForSameRoll", func(t *testing.T) {
		f := func(threes uint8, chaos uint8) bool {
			threesCount := int(threes % 7)
			chaosCount := int(chaos % 7)

			ability := createTestAbilityForTest()

			roll := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    threesCount,
				Success:   threesCount > 0,
				Chaos:     chaosCount,
				Overload:  0,
				TripleAsc: false,
			}

			context := &AbilityContext{OnDuty: true}

			// 第一次应用效果
			var effect1, effect2 *EffectResult
			var err error

			if roll.Success {
				effect1, err = abilityService.ApplySuccessEffect(ability, roll, context)
			} else {
				effect1, err = abilityService.ApplyFailureEffect(ability, roll, context)
			}
			if err != nil {
				return false
			}

			// 第二次应用效果
			if roll.Success {
				effect2, err = abilityService.ApplySuccessEffect(ability, roll, context)
			} else {
				effect2, err = abilityService.ApplyFailureEffect(ability, roll, context)
			}
			if err != nil {
				return false
			}

			// 验证两次效果一致
			if effect1 == nil && effect2 == nil {
				return true
			}
			if effect1 == nil || effect2 == nil {
				return false
			}

			return effect1.Description == effect2.Description &&
				effect1.Mechanics == effect2.Mechanics &&
				effect1.Applied == effect2.Applied
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// 辅助函数：创建测试用的特工
func createTestAgentForAbility() *domain.Agent {
	return &domain.Agent{
		ID:       "test-agent",
		Name:     "测试特工",
		Pronouns: "他/她",
		Anomaly: &domain.Anomaly{
			Type:      domain.AnomalyWhisper,
			Abilities: []*domain.AnomalyAbility{},
		},
		Reality: &domain.Reality{
			Type: domain.RealityCaretaker,
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
			{ID: "rel-1", Name: "关系1", Connection: 4},
			{ID: "rel-2", Name: "关系2", Connection: 4},
			{ID: "rel-3", Name: "关系3", Connection: 4},
		},
		Commendations: 0,
		Reprimands:    0,
		Rating:        domain.RatingExcellent,
		Alive:         true,
		InDebt:        false,
	}
}

// 辅助函数：创建测试用的游戏会话
func createTestSessionForAbility() *domain.GameSession {
	return &domain.GameSession{
		ID:         "test-session",
		AgentID:    "test-agent",
		ScenarioID: "test-scenario",
		Phase:      domain.PhaseInvestigation,
		State: &domain.GameState{
			CurrentSceneID:    "scene-1",
			VisitedScenes:     make(map[string]bool),
			CollectedClues:    []string{},
			UnlockedLocations: []string{},
			DomainUnlocked:    false,
			NPCStates:         make(map[string]*domain.NPCState),
			ChaosPool:         0,
			LooseEnds:         0,
			LocationOverloads: make(map[string]int),
			AnomalyStatus:     "active",
			MissionOutcome:    "",
		},
	}
}

// 辅助函数：创建测试用的异常能力
func createTestAbilityForTest() *domain.AnomalyAbility {
	return &domain.AnomalyAbility{
		ID:          "ability-1",
		Name:        "测试能力",
		AnomalyType: domain.AnomalyWhisper,
		Trigger: &domain.AbilityTrigger{
			Type:        domain.TriggerAction,
			Description: "主动使用",
		},
		Roll: &domain.AbilityRoll{
			Quality:   domain.QualityFocus,
			DiceCount: 6,
			DiceType:  4,
		},
		Effects: &domain.AbilityEffects{
			Success: &domain.Effect{
				Description: "成功效果",
				Mechanics:   "获得信息",
			},
			Failure: &domain.Effect{
				Description: "失败效果",
				Mechanics:   "产生混沌",
			},
		},
	}
}

// 辅助函数：创建带有额外效果的测试能力
func createTestAbilityWithAdditionalForTest(condition string) *domain.AnomalyAbility {
	ability := createTestAbilityForTest()
	ability.Effects.Additional = []*domain.ConditionalEffect{
		{
			Condition: condition,
			Effect: &domain.Effect{
				Description: "额外效果",
				Mechanics:   "额外奖励",
			},
		},
	}
	return ability
}

// 单元测试：验证基本功能
func TestAbilityService_BasicFunctionality(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)
	chaosService := NewChaosService()
	abilityService := NewAbilityService(diceService, qaService, chaosService)

	t.Run("ValidateTrigger_Action", func(t *testing.T) {
		ability := createTestAbilityForTest()
		ability.Trigger.Type = domain.TriggerAction

		valid, err := abilityService.ValidateTrigger(ability, &AbilityContext{OnDuty: true})
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("ApplySuccessEffect", func(t *testing.T) {
		ability := createTestAbilityForTest()
		roll := &domain.RollResult{
			Dice:    []int{3, 3, 1, 2, 4, 1},
			Threes:  2,
			Success: true,
			Chaos:   0,
		}

		effect, err := abilityService.ApplySuccessEffect(ability, roll, &AbilityContext{OnDuty: true})
		assert.NoError(t, err)
		assert.NotNil(t, effect)
		assert.True(t, effect.Applied)
		assert.Equal(t, "成功效果", effect.Description)
	})

	t.Run("ApplyFailureEffect", func(t *testing.T) {
		ability := createTestAbilityForTest()
		roll := &domain.RollResult{
			Dice:    []int{1, 1, 2, 2, 4, 4},
			Threes:  0,
			Success: false,
			Chaos:   6,
		}

		effect, err := abilityService.ApplyFailureEffect(ability, roll, &AbilityContext{OnDuty: true})
		assert.NoError(t, err)
		assert.NotNil(t, effect)
		assert.True(t, effect.Applied)
		assert.Equal(t, "失败效果", effect.Description)
	})

	t.Run("ApplyAdditionalEffects_ExtraThrees", func(t *testing.T) {
		ability := createTestAbilityWithAdditionalForTest("每额外一个3")
		roll := &domain.RollResult{
			Dice:    []int{3, 3, 3, 2, 4, 1},
			Threes:  3,
			Success: true,
			Chaos:   0,
		}

		effects, err := abilityService.ApplyAdditionalEffects(ability, roll, &AbilityContext{OnDuty: true})
		assert.NoError(t, err)
		assert.NotEmpty(t, effects)
		assert.True(t, effects[0].Applied)
	})

	t.Run("CheckOffDutyUsage_Morning", func(t *testing.T) {
		agent := createTestAgentForAbility()
		session := createTestSessionForAbility()
		session.Phase = domain.PhaseMorning

		offDuty := abilityService.CheckOffDutyUsage(agent, session)
		assert.True(t, offDuty)
	})

	t.Run("CheckOffDutyUsage_Investigation", func(t *testing.T) {
		agent := createTestAgentForAbility()
		session := createTestSessionForAbility()
		session.Phase = domain.PhaseInvestigation

		offDuty := abilityService.CheckOffDutyUsage(agent, session)
		assert.False(t, offDuty)
	})
}
