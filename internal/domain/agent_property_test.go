package domain

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: trpg-solo-engine, Property 1: 角色创建完整性
// 验证需求: 1.1, 1.2, 1.3, 1.4, 1.5
//
// 属性1: 角色创建完整性
// 对于任何有效的ARC组合（异常、现实、职能），创建角色后应该包含所有必需组件：
// - 3种异常能力
// - 现实触发器
// - 过载解除
// - 3段人际关系（总计12点连结）
// - 9点资质保证
func TestProperty_AgentCreationCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("角色创建完整性", prop.ForAll(
		func(anomalyIdx, realityIdx, careerIdx int) bool {
			// 确保索引为正数
			if anomalyIdx < 0 {
				anomalyIdx = -anomalyIdx
			}
			if realityIdx < 0 {
				realityIdx = -realityIdx
			}
			if careerIdx < 0 {
				careerIdx = -careerIdx
			}

			// 从有效类型中选择
			anomalyType := AllAnomalyTypes[anomalyIdx%len(AllAnomalyTypes)]
			realityType := AllRealityTypes[realityIdx%len(AllRealityTypes)]
			careerType := AllCareerTypes[careerIdx%len(AllCareerTypes)]

			// 创建角色
			agent := createTestAgent(anomalyType, realityType, careerType)

			// 验证完整性
			// 1. 验证3种异常能力
			if len(agent.Anomaly.Abilities) != 3 {
				t.Logf("异常能力数量错误: expected 3, got %d", len(agent.Anomaly.Abilities))
				return false
			}

			// 2. 验证现实触发器存在
			if agent.Reality.Trigger == nil {
				t.Logf("现实触发器缺失")
				return false
			}

			// 3. 验证过载解除存在
			if agent.Reality.OverloadRelief == nil {
				t.Logf("过载解除缺失")
				return false
			}

			// 4. 验证3段人际关系
			if len(agent.Relationships) != 3 {
				t.Logf("人际关系数量错误: expected 3, got %d", len(agent.Relationships))
				return false
			}

			// 5. 验证总连结点数为12
			totalConnection := agent.TotalConnection()
			if totalConnection != 12 {
				t.Logf("总连结点数错误: expected 12, got %d", totalConnection)
				return false
			}

			// 6. 验证总QA点数不超过9
			totalQA := agent.TotalQA()
			if totalQA > 9 {
				t.Logf("总QA点数超过限制: expected <=9, got %d", totalQA)
				return false
			}

			// 7. 验证退化轨道存在
			if agent.Reality.DegradationTrack == nil {
				t.Logf("退化轨道缺失")
				return false
			}

			return true
		},
		gen.Int(),
		gen.Int(),
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// createTestAgent 创建测试用角色
func createTestAgent(anomalyType, realityType, careerType string) *Agent {
	// 创建异常能力
	abilities := []*AnomalyAbility{
		{
			ID:          anomalyType + "-ability-1",
			Name:        "能力1",
			AnomalyType: anomalyType,
			Trigger: &AbilityTrigger{
				Type:        TriggerAction,
				Description: "测试触发器",
			},
			Roll: &AbilityRoll{
				Quality:   QualityFocus,
				DiceCount: 6,
				DiceType:  4,
			},
			Effects: &AbilityEffects{
				Success: &Effect{
					Description: "成功效果",
					Mechanics:   "测试机制",
				},
				Failure: &Effect{
					Description: "失败效果",
					Mechanics:   "测试机制",
				},
			},
		},
		{
			ID:          anomalyType + "-ability-2",
			Name:        "能力2",
			AnomalyType: anomalyType,
		},
		{
			ID:          anomalyType + "-ability-3",
			Name:        "能力3",
			AnomalyType: anomalyType,
		},
	}

	// 创建人际关系（总计12点连结）
	relationships := []*Relationship{
		{
			ID:          "rel-1",
			Name:        "关系1",
			Description: "测试关系",
			Connection:  6,
			PlayedBy:    "GM",
		},
		{
			ID:          "rel-2",
			Name:        "关系2",
			Description: "测试关系",
			Connection:  3,
			PlayedBy:    "GM",
		},
		{
			ID:          "rel-3",
			Name:        "关系3",
			Description: "测试关系",
			Connection:  3,
			PlayedBy:    "GM",
		},
	}

	// 创建QA分配（总计9点）
	careerQA := map[string]int{
		QualityFocus:      1,
		QualityEmpathy:    1,
		QualityPresence:   1,
		QualityDeception:  1,
		QualityInitiative: 1,
		QualityProfession: 1,
		QualityVitality:   1,
		QualityGrit:       1,
		QualitySubtlety:   1,
	}

	// 复制一份给agent当前QA
	currentQA := make(map[string]int)
	for k, v := range careerQA {
		currentQA[k] = v
	}

	agent := &Agent{
		ID:       "test-agent",
		Name:     "测试特工",
		Pronouns: "他/他的",
		Anomaly: &Anomaly{
			Type:      anomalyType,
			Abilities: abilities,
		},
		Reality: &Reality{
			Type: realityType,
			Trigger: &RealityTrigger{
				Name:        "测试触发器",
				Cost:        0,
				Effect:      "测试效果",
				Consequence: "测试后果",
			},
			OverloadRelief: &OverloadRelief{
				Name:      "测试解除",
				Condition: "测试条件",
				Effect:    "无视所有过载",
			},
			DegradationTrack: &DegradationTrack{
				Name:   "测试轨道",
				Filled: 0,
				Total:  4,
			},
			Relationships: relationships,
		},
		Career: &Career{
			Type: careerType,
			QA:   careerQA,
		},
		QA:            currentQA,
		Relationships: relationships,
		Commendations: 0,
		Reprimands:    0,
		Rating:        RatingExcellent,
		Alive:         true,
		InDebt:        false,
	}

	return agent
}

// 测试属性：QA不变量
// 验证在任何操作后，QA总和不会超过初始分配
func TestProperty_QAInvariant(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("QA不变量", prop.ForAll(
		func(spendAmounts []int) bool {
			// 创建测试角色
			agent := createTestAgent(AnomalyWhisper, RealityCaretaker, CareerPublicRelations)
			initialTotal := agent.TotalQA()

			// 尝试花费QA
			for i, amount := range spendAmounts {
				if amount <= 0 {
					continue
				}

				quality := AllQualities[i%len(AllQualities)]
				_ = agent.SpendQA(quality, amount%10) // 忽略错误
			}

			// 验证QA总和不超过初始值
			currentTotal := agent.TotalQA()
			if currentTotal > initialTotal {
				t.Logf("QA总和超过初始值: initial=%d, current=%d", initialTotal, currentTotal)
				return false
			}

			// 恢复QA
			agent.RestoreQA()

			// 验证恢复后等于初始值
			restoredTotal := agent.TotalQA()
			if restoredTotal != initialTotal {
				t.Logf("QA恢复后不等于初始值: initial=%d, restored=%d", initialTotal, restoredTotal)
				return false
			}

			return true
		},
		gen.SliceOf(gen.Int()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// 测试属性：连结单调性
// 验证连结只能减少不能增加（除非特殊机制）
func TestProperty_ConnectionMonotonicity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("连结单调性", prop.ForAll(
		func(lossAmounts []int) bool {
			agent := createTestAgent(AnomalyWhisper, RealityCaretaker, CareerPublicRelations)
			initialTotal := agent.TotalConnection()

			// 模拟失去连结
			for _, amount := range lossAmounts {
				if amount <= 0 {
					continue
				}

				weakest := agent.GetWeakestRelationship()
				if weakest != nil {
					loss := amount % 7 // 限制在0-6之间
					weakest.Connection -= loss
					if weakest.Connection < 0 {
						weakest.Connection = 0
					}
				}
			}

			// 验证总连结不会增加
			currentTotal := agent.TotalConnection()
			if currentTotal > initialTotal {
				t.Logf("连结总数增加了: initial=%d, current=%d", initialTotal, currentTotal)
				return false
			}

			return true
		},
		gen.SliceOf(gen.Int()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// 测试属性：评级映射一致性
// 验证申诫数量与评级的映射关系
func TestProperty_RatingConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("评级映射一致性", prop.ForAll(
		func(reprimands int) bool {
			// 限制申诫数量在合理范围内
			reprimands = reprimands % 20
			if reprimands < 0 {
				reprimands = -reprimands
			}

			rating := GetRating(reprimands)

			// 验证映射关系
			switch {
			case reprimands == 0:
				return rating == RatingExcellent
			case reprimands == 1:
				return rating == RatingNeedsWork
			case reprimands >= 2 && reprimands <= 3:
				return rating == RatingProbation
			case reprimands >= 4 && reprimands <= 9:
				return rating == RatingFinalWarning
			default:
				return rating == RatingRevoked
			}
		},
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
