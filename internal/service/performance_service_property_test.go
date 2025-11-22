package service

import (
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// **Feature: trpg-solo-engine, Property 13: 机构评级映射**
// **Validates: Requirements 9.3**
//
// 属性13: 机构评级映射
// 对于任何角色，机构评级应该与申诫总数一一对应：
// 0申诫="评级良好"，1申诫="有待改进"，2-3申诫="留职察看"，
// 4-9申诫="最后警告"，10+申诫="权限已撤销"。
func TestProperty_AgencyRatingMapping(t *testing.T) {
	performanceService := NewPerformanceService()

	// 测试1: 0申诫 = "评级良好"
	t.Run("ZeroReprimandsExcellentRating", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForPerformance()
			agent.Reprimands = 0

			err := performanceService.UpdateRating(agent)
			if err != nil {
				return false
			}

			return agent.Rating == domain.RatingExcellent
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 1申诫 = "有待改进"
	t.Run("OneReprimandNeedsWork", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForPerformance()
			agent.Reprimands = 1

			err := performanceService.UpdateRating(agent)
			if err != nil {
				return false
			}

			return agent.Rating == domain.RatingNeedsWork
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 2-3申诫 = "留职察看"
	t.Run("TwoToThreeReprimandsProbation", func(t *testing.T) {
		f := func(reprimands uint8) bool {
			// 限制在2-3范围内
			rep := int(reprimands%2) + 2 // 2或3

			agent := createTestAgentForPerformance()
			agent.Reprimands = rep

			err := performanceService.UpdateRating(agent)
			if err != nil {
				return false
			}

			return agent.Rating == domain.RatingProbation
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 4-9申诫 = "最后警告"
	t.Run("FourToNineReprimandsFinalWarning", func(t *testing.T) {
		f := func(reprimands uint8) bool {
			// 限制在4-9范围内
			rep := int(reprimands%6) + 4 // 4-9

			agent := createTestAgentForPerformance()
			agent.Reprimands = rep

			err := performanceService.UpdateRating(agent)
			if err != nil {
				return false
			}

			return agent.Rating == domain.RatingFinalWarning
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: 10+申诫 = "权限已撤销"
	t.Run("TenPlusReprimandsRevoked", func(t *testing.T) {
		f := func(reprimands uint8) bool {
			// 限制在10+范围内
			rep := int(reprimands%90) + 10 // 10-99

			agent := createTestAgentForPerformance()
			agent.Reprimands = rep

			err := performanceService.UpdateRating(agent)
			if err != nil {
				return false
			}

			return agent.Rating == domain.RatingRevoked
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试6: AddReprimands自动更新评级
	t.Run("AddReprimandsAutoUpdatesRating", func(t *testing.T) {
		f := func(initialRep uint8, addRep uint8) bool {
			// 限制范围
			initial := int(initialRep % 20) // 0-19
			add := int(addRep%10) + 1       // 1-10
			expected := initial + add

			agent := createTestAgentForPerformance()
			agent.Reprimands = initial

			err := performanceService.AddReprimands(agent, add)
			if err != nil {
				return false
			}

			// 验证申诫数量
			if agent.Reprimands != expected {
				return false
			}

			// 验证评级正确
			expectedRating := domain.GetRating(expected)
			return agent.Rating == expectedRating
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试7: CalculateRating函数的一致性
	t.Run("CalculateRatingConsistency", func(t *testing.T) {
		f := func(reprimands uint8) bool {
			rep := int(reprimands % 100) // 0-99

			// 使用服务计算
			serviceRating := performanceService.CalculateRating(rep)

			// 使用domain函数计算
			domainRating := domain.GetRating(rep)

			// 两者应该一致
			return serviceRating == domainRating
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试8: 评级映射的完整性（所有可能的申诫数都有对应的评级）
	t.Run("RatingMappingCompleteness", func(t *testing.T) {
		f := func(reprimands uint8) bool {
			rep := int(reprimands % 100) // 0-99

			rating := performanceService.CalculateRating(rep)

			// 评级应该是五个有效值之一
			validRatings := []string{
				domain.RatingExcellent,
				domain.RatingNeedsWork,
				domain.RatingProbation,
				domain.RatingFinalWarning,
				domain.RatingRevoked,
			}

			for _, valid := range validRatings {
				if rating == valid {
					return true
				}
			}

			return false
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// **Feature: trpg-solo-engine, Property 14: 任务结果奖励**
// **Validates: Requirements 3.4**
//
// 属性14: 任务结果奖励
// 对于任何完成的任务，捕获异常体应该给予每名特工3次嘉奖，
// 中和异常体无奖惩，逃脱应该给予每名特工3次申诫。
func TestProperty_MissionResultRewards(t *testing.T) {
	performanceService := NewPerformanceService()

	// 测试1: 捕获异常体给予3次嘉奖
	t.Run("CaptureGives3Commendations", func(t *testing.T) {
		f := func(initialComm uint8) bool {
			initial := int(initialComm % 50) // 0-49

			agent := createTestAgentForPerformance()
			agent.Commendations = initial

			err := performanceService.AwardMissionSuccess(agent, OutcomeCaptured)
			if err != nil {
				return false
			}

			// 应该增加3次嘉奖
			return agent.Commendations == initial+3
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 中和异常体无奖惩
	t.Run("NeutralizationNoReward", func(t *testing.T) {
		f := func(initialComm uint8, initialRep uint8) bool {
			comm := int(initialComm % 50) // 0-49
			rep := int(initialRep % 20)   // 0-19

			agent := createTestAgentForPerformance()
			agent.Commendations = comm
			agent.Reprimands = rep

			err := performanceService.AwardMissionSuccess(agent, OutcomeNeutralized)
			if err != nil {
				return false
			}

			// 嘉奖和申诫都不应该改变
			return agent.Commendations == comm && agent.Reprimands == rep
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 异常体逃脱给予3次申诫
	t.Run("EscapeGives3Reprimands", func(t *testing.T) {
		f := func(initialRep uint8) bool {
			initial := int(initialRep % 20) // 0-19

			agent := createTestAgentForPerformance()
			agent.Reprimands = initial

			err := performanceService.AwardMissionSuccess(agent, OutcomeEscaped)
			if err != nil {
				return false
			}

			// 应该增加3次申诫
			if agent.Reprimands != initial+3 {
				return false
			}

			// 评级应该更新
			expectedRating := domain.GetRating(initial + 3)
			return agent.Rating == expectedRating
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: AwardCaptureBonus直接调用
	t.Run("AwardCaptureBonusDirectCall", func(t *testing.T) {
		f := func(initialComm uint8) bool {
			initial := int(initialComm % 50)

			agent := createTestAgentForPerformance()
			agent.Commendations = initial

			err := performanceService.AwardCaptureBonus(agent)
			if err != nil {
				return false
			}

			return agent.Commendations == initial+3
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: AwardEscapePenalty直接调用
	t.Run("AwardEscapePenaltyDirectCall", func(t *testing.T) {
		f := func(initialRep uint8) bool {
			initial := int(initialRep % 20)

			agent := createTestAgentForPerformance()
			agent.Reprimands = initial

			err := performanceService.AwardEscapePenalty(agent)
			if err != nil {
				return false
			}

			return agent.Reprimands == initial+3
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试6: 捕获后负债状态更新
	t.Run("CaptureUpdatesDebtStatus", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForPerformance()
			agent.Commendations = -5 // 负债状态
			agent.InDebt = true

			err := performanceService.AwardCaptureBonus(agent)
			if err != nil {
				return false
			}

			// 嘉奖应该增加到-2
			if agent.Commendations != -2 {
				return false
			}

			// 仍然在负债状态
			return agent.InDebt
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试7: 捕获后脱离负债状态
	t.Run("CaptureExitsDebtStatus", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForPerformance()
			agent.Commendations = -2 // 负债状态
			agent.InDebt = true

			err := performanceService.AwardCaptureBonus(agent)
			if err != nil {
				return false
			}

			// 嘉奖应该增加到1
			if agent.Commendations != 1 {
				return false
			}

			// 应该脱离负债状态
			return !agent.InDebt
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试8: 无效的任务结果返回错误
	t.Run("InvalidOutcomeReturnsError", func(t *testing.T) {
		f := func() bool {
			agent := createTestAgentForPerformance()
			initialComm := agent.Commendations
			initialRep := agent.Reprimands

			err := performanceService.AwardMissionSuccess(agent, "无效结果")

			// 应该返回错误
			if err == nil {
				return false
			}

			// 嘉奖和申诫不应该改变
			return agent.Commendations == initialComm && agent.Reprimands == initialRep
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// createTestAgentForPerformance 创建测试用角色
func createTestAgentForPerformance() *domain.Agent {
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
func TestPerformanceService_BasicFunctionality(t *testing.T) {
	performanceService := NewPerformanceService()

	t.Run("AddCommendations", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Commendations = 5

		err := performanceService.AddCommendations(agent, 3)
		assert.NoError(t, err)
		assert.Equal(t, 8, agent.Commendations)
	})

	t.Run("AddCommendationsNegativeAmount", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		initialComm := agent.Commendations

		err := performanceService.AddCommendations(agent, -3)
		assert.Error(t, err)
		assert.Equal(t, initialComm, agent.Commendations)
	})

	t.Run("AddReprimands", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Reprimands = 0

		err := performanceService.AddReprimands(agent, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, agent.Reprimands)
		assert.Equal(t, domain.RatingProbation, agent.Rating)
	})

	t.Run("AddReprimandsNegativeAmount", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		initialRep := agent.Reprimands

		err := performanceService.AddReprimands(agent, -2)
		assert.Error(t, err)
		assert.Equal(t, initialRep, agent.Reprimands)
	})

	t.Run("GetCommendations", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Commendations = 15

		assert.Equal(t, 15, performanceService.GetCommendations(agent))
	})

	t.Run("GetReprimands", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Reprimands = 5

		assert.Equal(t, 5, performanceService.GetReprimands(agent))
	})

	t.Run("UpdateRating", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Reprimands = 10

		err := performanceService.UpdateRating(agent)
		assert.NoError(t, err)
		assert.Equal(t, domain.RatingRevoked, agent.Rating)
	})

	t.Run("GetRating", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Rating = domain.RatingNeedsWork

		assert.Equal(t, domain.RatingNeedsWork, performanceService.GetRating(agent))
	})

	t.Run("CalculateRating", func(t *testing.T) {
		assert.Equal(t, domain.RatingExcellent, performanceService.CalculateRating(0))
		assert.Equal(t, domain.RatingNeedsWork, performanceService.CalculateRating(1))
		assert.Equal(t, domain.RatingProbation, performanceService.CalculateRating(2))
		assert.Equal(t, domain.RatingProbation, performanceService.CalculateRating(3))
		assert.Equal(t, domain.RatingFinalWarning, performanceService.CalculateRating(4))
		assert.Equal(t, domain.RatingFinalWarning, performanceService.CalculateRating(9))
		assert.Equal(t, domain.RatingRevoked, performanceService.CalculateRating(10))
		assert.Equal(t, domain.RatingRevoked, performanceService.CalculateRating(100))
	})

	t.Run("CheckDebt", func(t *testing.T) {
		agent := createTestAgentForPerformance()

		agent.Commendations = 5
		assert.False(t, performanceService.CheckDebt(agent))

		agent.Commendations = 0
		assert.False(t, performanceService.CheckDebt(agent))

		agent.Commendations = -1
		assert.True(t, performanceService.CheckDebt(agent))

		agent.Commendations = -10
		assert.True(t, performanceService.CheckDebt(agent))
	})

	t.Run("UpdateDebtStatus", func(t *testing.T) {
		agent := createTestAgentForPerformance()

		agent.Commendations = 5
		err := performanceService.UpdateDebtStatus(agent)
		assert.NoError(t, err)
		assert.False(t, agent.InDebt)

		agent.Commendations = -5
		err = performanceService.UpdateDebtStatus(agent)
		assert.NoError(t, err)
		assert.True(t, agent.InDebt)
	})

	t.Run("AwardMissionSuccess_Captured", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Commendations = 5

		err := performanceService.AwardMissionSuccess(agent, OutcomeCaptured)
		assert.NoError(t, err)
		assert.Equal(t, 8, agent.Commendations)
	})

	t.Run("AwardMissionSuccess_Neutralized", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		initialComm := agent.Commendations
		initialRep := agent.Reprimands

		err := performanceService.AwardMissionSuccess(agent, OutcomeNeutralized)
		assert.NoError(t, err)
		assert.Equal(t, initialComm, agent.Commendations)
		assert.Equal(t, initialRep, agent.Reprimands)
	})

	t.Run("AwardMissionSuccess_Escaped", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		agent.Reprimands = 0

		err := performanceService.AwardMissionSuccess(agent, OutcomeEscaped)
		assert.NoError(t, err)
		assert.Equal(t, 3, agent.Reprimands)
		assert.Equal(t, domain.RatingProbation, agent.Rating)
	})

	t.Run("AwardMissionSuccess_InvalidOutcome", func(t *testing.T) {
		agent := createTestAgentForPerformance()
		initialComm := agent.Commendations
		initialRep := agent.Reprimands

		err := performanceService.AwardMissionSuccess(agent, "无效结果")
		assert.Error(t, err)
		assert.Equal(t, initialComm, agent.Commendations)
		assert.Equal(t, initialRep, agent.Reprimands)
	})
}
