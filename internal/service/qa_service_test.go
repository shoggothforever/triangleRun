package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

func TestQAService_SpendQA(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	t.Run("成功花费QA", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		quality := domain.QualityFocus
		initialQA := agent.QA[quality]

		err := qaService.SpendQA(agent, quality, 1)
		assert.NoError(t, err)
		assert.Equal(t, initialQA-1, agent.QA[quality])
	})

	t.Run("QA不足时返回错误", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		quality := domain.QualityFocus
		agent.QA[quality] = 1

		err := qaService.SpendQA(agent, quality, 2)
		assert.Error(t, err)
		assert.Equal(t, 1, agent.QA[quality]) // QA不应该改变
	})

	t.Run("花费0点QA", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		quality := domain.QualityFocus
		initialQA := agent.QA[quality]

		err := qaService.SpendQA(agent, quality, 0)
		assert.NoError(t, err)
		assert.Equal(t, initialQA, agent.QA[quality])
	})
}

func TestQAService_RestoreQA(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	t.Run("恢复所有QA到初始值", func(t *testing.T) {
		agent := createTestAgentForUnitTest()

		// 花费一些QA
		_ = qaService.SpendQA(agent, domain.QualityFocus, 1)
		_ = qaService.SpendQA(agent, domain.QualityEmpathy, 1)

		// 恢复
		err := qaService.RestoreQA(agent)
		assert.NoError(t, err)
		assert.Equal(t, 9, agent.TotalQA())
	})

	t.Run("完全耗尽后恢复", func(t *testing.T) {
		agent := createTestAgentForUnitTest()

		// 耗尽所有QA
		for quality, qa := range agent.QA {
			_ = qaService.SpendQA(agent, quality, qa)
		}
		assert.Equal(t, 0, agent.TotalQA())

		// 恢复
		err := qaService.RestoreQA(agent)
		assert.NoError(t, err)
		assert.Equal(t, 9, agent.TotalQA())
	})
}

func TestQAService_CheckOverload(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	t.Run("QA为0时需要过载", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 0

		needsOverload := qaService.CheckOverload(agent, domain.QualityFocus)
		assert.True(t, needsOverload)
	})

	t.Run("QA大于0时不需要过载", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 1

		needsOverload := qaService.CheckOverload(agent, domain.QualityFocus)
		assert.False(t, needsOverload)
	})
}

func TestQAService_ApplyOverload(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)

	t.Run("QA为0时应用过载", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 0

		// 创建一个有"3"的掷骰结果
		roll := &domain.RollResult{
			Dice:      []int{3, 3, 2, 1, 4, 2},
			Threes:    2,
			Success:   true,
			Chaos:     0,
			Overload:  0,
			TripleAsc: false,
		}

		result := qaService.ApplyOverload(agent, domain.QualityFocus, roll)
		assert.Equal(t, 1, result.Threes)
		assert.Equal(t, 1, result.Overload)
		assert.Equal(t, 1, result.Chaos)
	})

	t.Run("QA大于0时不应用过载", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 1

		roll := &domain.RollResult{
			Dice:      []int{3, 3, 2, 1, 4, 2},
			Threes:    2,
			Success:   true,
			Chaos:     0,
			Overload:  0,
			TripleAsc: false,
		}

		result := qaService.ApplyOverload(agent, domain.QualityFocus, roll)
		assert.Equal(t, 2, result.Threes)
		assert.Equal(t, 0, result.Overload)
	})

	t.Run("没有3时过载无法移除", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 0

		roll := &domain.RollResult{
			Dice:      []int{1, 2, 4, 1, 4, 2},
			Threes:    0,
			Success:   false,
			Chaos:     6,
			Overload:  0,
			TripleAsc: false,
		}

		result := qaService.ApplyOverload(agent, domain.QualityFocus, roll)
		assert.Equal(t, 0, result.Threes)
		assert.Equal(t, 1, result.Overload)
	})
}

func TestQAService_AdjustDiceWithQA(t *testing.T) {
	diceService := domain.NewDiceService()
	qaService := NewQAService(diceService)

	t.Run("成功调整骰子", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 2

		roll := &domain.RollResult{
			Dice:      []int{1, 2, 4, 1, 4, 2},
			Threes:    0,
			Success:   false,
			Chaos:     6,
			Overload:  0,
			TripleAsc: false,
		}

		adjustments := []DiceAdjustment{
			{DiceIndex: 0, NewValue: 3},
			{DiceIndex: 1, NewValue: 3},
		}

		result, err := qaService.AdjustDiceWithQA(agent, domain.QualityFocus, roll, adjustments)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Threes)
		assert.True(t, result.Success)
		assert.Equal(t, 0, agent.QA[domain.QualityFocus]) // QA被花费
	})

	t.Run("QA不足时返回错误", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 1

		roll := &domain.RollResult{
			Dice:      []int{1, 2, 4, 1, 4, 2},
			Threes:    0,
			Success:   false,
			Chaos:     6,
			Overload:  0,
			TripleAsc: false,
		}

		adjustments := []DiceAdjustment{
			{DiceIndex: 0, NewValue: 3},
			{DiceIndex: 1, NewValue: 3},
		}

		_, err := qaService.AdjustDiceWithQA(agent, domain.QualityFocus, roll, adjustments)
		assert.Error(t, err)
		assert.Equal(t, 1, agent.QA[domain.QualityFocus]) // QA不应该改变
	})

	t.Run("调整失败的掷骰为成功", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 1

		roll := &domain.RollResult{
			Dice:      []int{1, 2, 4, 1, 4, 2},
			Threes:    0,
			Success:   false,
			Chaos:     6,
			Overload:  0,
			TripleAsc: false,
		}

		adjustments := []DiceAdjustment{
			{DiceIndex: 0, NewValue: 3},
		}

		result, err := qaService.AdjustDiceWithQA(agent, domain.QualityFocus, roll, adjustments)
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Threes)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.Chaos) // 成功时混沌为0
	})
}

func TestQAService_GetAvailableQA(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	t.Run("获取可用QA", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		agent.QA[domain.QualityFocus] = 3

		available := qaService.GetAvailableQA(agent, domain.QualityFocus)
		assert.Equal(t, 3, available)
	})

	t.Run("不存在的资质返回0", func(t *testing.T) {
		agent := createTestAgentForUnitTest()

		available := qaService.GetAvailableQA(agent, "不存在的资质")
		assert.Equal(t, 0, available)
	})
}

func TestQAService_GetTotalQA(t *testing.T) {
	qaService := NewQAService(domain.NewDiceService())

	t.Run("获取总QA", func(t *testing.T) {
		agent := createTestAgentForUnitTest()

		total := qaService.GetTotalQA(agent)
		assert.Equal(t, 9, total)
	})

	t.Run("花费后总QA减少", func(t *testing.T) {
		agent := createTestAgentForUnitTest()
		// 花费1点Focus和1点Empathy
		err1 := qaService.SpendQA(agent, domain.QualityFocus, 1)
		err2 := qaService.SpendQA(agent, domain.QualityEmpathy, 1)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		total := qaService.GetTotalQA(agent)
		assert.Equal(t, 7, total)
	})
}

func createTestAgentForUnitTest() *domain.Agent {
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
