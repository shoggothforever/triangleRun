package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

func TestAIService_GenerateSceneDescription(t *testing.T) {
	service := NewAIService()

	t.Run("生成基础场景描述", func(t *testing.T) {
		scene := &domain.Scene{
			ID:          "scene1",
			Name:        "咖啡馆",
			Description: "一个温馨的咖啡馆，弥漫着咖啡的香气。",
			NPCs:        []*domain.NPC{},
			Clues:       []*domain.Clue{},
		}

		state := &domain.GameState{
			VisitedScenes:  make(map[string]bool),
			CollectedClues: []string{},
			NPCStates:      make(map[string]*domain.NPCState),
		}

		description, err := service.GenerateSceneDescription(scene, state)

		require.NoError(t, err)
		assert.Contains(t, description, "咖啡馆")
		assert.Contains(t, description, "一个温馨的咖啡馆")
		assert.Contains(t, description, "第一次来到这里")
	})

	t.Run("生成包含NPC的场景描述", func(t *testing.T) {
		scene := &domain.Scene{
			ID:          "scene2",
			Name:        "办公室",
			Description: "一个整洁的办公室。",
			NPCs: []*domain.NPC{
				{
					ID:   "npc1",
					Name: "张经理",
				},
				{
					ID:   "npc2",
					Name: "李秘书",
				},
			},
		}

		state := &domain.GameState{
			VisitedScenes:  map[string]bool{"scene2": true},
			CollectedClues: []string{},
			NPCStates: map[string]*domain.NPCState{
				"npc1": {
					ID:              "npc1",
					AnomalyAffected: false,
				},
				"npc2": {
					ID:              "npc2",
					AnomalyAffected: true,
				},
			},
		}

		description, err := service.GenerateSceneDescription(scene, state)

		require.NoError(t, err)
		assert.Contains(t, description, "办公室")
		assert.Contains(t, description, "张经理")
		assert.Contains(t, description, "李秘书")
		assert.Contains(t, description, "不对劲")
		assert.Contains(t, description, "熟悉的地方")
	})

	t.Run("生成包含线索的场景描述", func(t *testing.T) {
		scene := &domain.Scene{
			ID:          "scene3",
			Name:        "图书馆",
			Description: "一个安静的图书馆。",
			Clues: []*domain.Clue{
				{ID: "clue1", Name: "神秘笔记"},
				{ID: "clue2", Name: "旧照片"},
				{ID: "clue3", Name: "日记"},
			},
		}

		state := &domain.GameState{
			VisitedScenes:  make(map[string]bool),
			CollectedClues: []string{"clue1"}, // 已收集一条线索
			NPCStates:      make(map[string]*domain.NPCState),
		}

		description, err := service.GenerateSceneDescription(scene, state)

		require.NoError(t, err)
		assert.Contains(t, description, "图书馆")
		assert.Contains(t, description, "2 条线索") // 3条线索，已收集1条，剩余2条
	})

	t.Run("场景为空时返回错误", func(t *testing.T) {
		state := &domain.GameState{}

		_, err := service.GenerateSceneDescription(nil, state)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "场景不能为空")
	})
}

func TestAIService_GenerateNPCDialogue(t *testing.T) {
	service := NewAIService()

	t.Run("生成正常NPC对话", func(t *testing.T) {
		npc := &domain.NPC{
			ID:          "npc1",
			Name:        "王老板",
			Personality: "友好",
			Dialogues: []string{
				"欢迎光临！",
				"今天天气不错。",
			},
		}

		context := &DialogueContext{
			PlayerAction: "打招呼",
			SceneID:      "scene1",
			GamePhase:    domain.PhaseInvestigation,
			NPCState: &domain.NPCState{
				ID:              "npc1",
				AnomalyAffected: false,
			},
		}

		dialogue, err := service.GenerateNPCDialogue(npc, context)

		require.NoError(t, err)
		assert.Contains(t, dialogue, "王老板")
		assert.NotEmpty(t, dialogue)
	})

	t.Run("生成受异常影响的NPC对话", func(t *testing.T) {
		npc := &domain.NPC{
			ID:          "npc2",
			Name:        "李员工",
			Personality: "正常",
		}

		context := &DialogueContext{
			PlayerAction: "询问情况",
			SceneID:      "scene2",
			GamePhase:    domain.PhaseInvestigation,
			NPCState: &domain.NPCState{
				ID:              "npc2",
				AnomalyAffected: true,
			},
		}

		dialogue, err := service.GenerateNPCDialogue(npc, context)

		require.NoError(t, err)
		assert.Contains(t, dialogue, "李员工")
		assert.Contains(t, dialogue, "受异常影响")
		// 受影响的对话应该包含一些异常的内容（检查多次以确保随机性）
		foundAnomalyContent := false
		for i := 0; i < 10; i++ {
			d, _ := service.GenerateNPCDialogue(npc, context)
			if strings.Contains(d, "不记得") ||
				strings.Contains(d, "声音") ||
				strings.Contains(d, "不对劲") ||
				strings.Contains(d, "发生") {
				foundAnomalyContent = true
				break
			}
		}
		assert.True(t, foundAnomalyContent, "应该生成包含异常内容的对话")
	})

	t.Run("根据性格生成对话", func(t *testing.T) {
		testCases := []struct {
			name        string
			personality string
			expectKey   string
		}{
			{"友好性格", "友好", "帮助"},
			{"冷漠性格", "冷漠", "有事吗"},
			{"紧张性格", "紧张", "你是谁"},
			{"可疑性格", "可疑", "不认识"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				npc := &domain.NPC{
					ID:          "npc_test",
					Name:        "测试NPC",
					Personality: tc.personality,
					Dialogues:   []string{}, // 空对话列表，触发性格生成
				}

				context := &DialogueContext{
					PlayerAction: "交谈",
					NPCState: &domain.NPCState{
						ID:              "npc_test",
						AnomalyAffected: false,
					},
				}

				dialogue, err := service.GenerateNPCDialogue(npc, context)

				require.NoError(t, err)
				assert.Contains(t, dialogue, tc.expectKey)
			})
		}
	})

	t.Run("NPC为空时返回错误", func(t *testing.T) {
		context := &DialogueContext{}

		_, err := service.GenerateNPCDialogue(nil, context)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "NPC不能为空")
	})

	t.Run("上下文为空时返回错误", func(t *testing.T) {
		npc := &domain.NPC{
			ID:   "npc1",
			Name: "测试NPC",
		}

		_, err := service.GenerateNPCDialogue(npc, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "对话上下文不能为空")
	})
}

func TestAIService_SelectChaosEffect(t *testing.T) {
	service := NewAIService()

	t.Run("选择可用的混沌效应", func(t *testing.T) {
		anomaly := &domain.AnomalyProfile{
			ID:   "anomaly1",
			Name: "永恒之泉",
			ChaosEffects: []*domain.ChaosEffect{
				{
					ID:          "effect1",
					Name:        "低级效应",
					Cost:        2,
					Description: "一个低成本效应",
				},
				{
					ID:          "effect2",
					Name:        "中级效应",
					Cost:        5,
					Description: "一个中等成本效应",
				},
				{
					ID:          "effect3",
					Name:        "高级效应",
					Cost:        8,
					Description: "一个高成本效应",
				},
			},
		}

		context := &domain.GameState{
			ChaosPool: 6,
		}

		effect, err := service.SelectChaosEffect(anomaly, 6, context)

		require.NoError(t, err)
		assert.NotNil(t, effect)
		assert.LessOrEqual(t, effect.Cost, 6)
		// 应该选择成本最高的可用效应
		assert.Equal(t, "effect2", effect.ID)
	})

	t.Run("混沌池不足时返回错误", func(t *testing.T) {
		anomaly := &domain.AnomalyProfile{
			ID:   "anomaly1",
			Name: "永恒之泉",
			ChaosEffects: []*domain.ChaosEffect{
				{
					ID:   "effect1",
					Name: "高级效应",
					Cost: 10,
				},
			},
		}

		context := &domain.GameState{
			ChaosPool: 3,
		}

		_, err := service.SelectChaosEffect(anomaly, 3, context)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "混沌池不足")
	})

	t.Run("异常体为空时返回错误", func(t *testing.T) {
		context := &domain.GameState{}

		_, err := service.SelectChaosEffect(nil, 10, context)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "异常体档案不能为空")
	})

	t.Run("没有混沌效应时返回错误", func(t *testing.T) {
		anomaly := &domain.AnomalyProfile{
			ID:           "anomaly1",
			Name:         "测试异常体",
			ChaosEffects: []*domain.ChaosEffect{},
		}

		context := &domain.GameState{}

		_, err := service.SelectChaosEffect(anomaly, 10, context)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "没有可用的混沌效应")
	})
}

func TestAIService_DescribeEvent(t *testing.T) {
	service := NewAIService()

	t.Run("描述基础事件", func(t *testing.T) {
		event := &domain.Event{
			ID:          "event1",
			Name:        "神秘声音",
			Description: "你听到了一个奇怪的声音。",
			Effect:      "混沌池+2",
		}

		context := &domain.GameState{
			ChaosPool: 3,
		}

		description, err := service.DescribeEvent(event, context)

		require.NoError(t, err)
		assert.Contains(t, description, "神秘声音")
		assert.Contains(t, description, "奇怪的声音")
		assert.Contains(t, description, "混沌池+2")
	})

	t.Run("高混沌池时添加额外描述", func(t *testing.T) {
		event := &domain.Event{
			ID:          "event2",
			Name:        "异常波动",
			Description: "周围的空气开始扭曲。",
		}

		context := &domain.GameState{
			ChaosPool: 8, // 高混沌池
		}

		description, err := service.DescribeEvent(event, context)

		require.NoError(t, err)
		assert.Contains(t, description, "异常能量正在增强")
	})

	t.Run("领域解锁时添加额外描述", func(t *testing.T) {
		event := &domain.Event{
			ID:          "event3",
			Name:        "领域显现",
			Description: "现实开始扭曲。",
		}

		context := &domain.GameState{
			DomainUnlocked: true,
		}

		description, err := service.DescribeEvent(event, context)

		require.NoError(t, err)
		assert.Contains(t, description, "异常体的领域已经显现")
	})

	t.Run("事件为空时返回错误", func(t *testing.T) {
		context := &domain.GameState{}

		_, err := service.DescribeEvent(nil, context)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "事件不能为空")
	})
}

func TestAIService_NarrateResult(t *testing.T) {
	service := NewAIService()

	t.Run("叙述成功的行动", func(t *testing.T) {
		action := &Action{
			Type:        "ability",
			Target:      "门",
			Description: "使用异常能力打开门",
			Quality:     "专注",
		}

		result := &ActionResult{
			Success: true,
			Threes:  3,
			Chaos:   0,
			Effects: []string{"门被打开了", "发现了新线索"},
		}

		narration, err := service.NarrateResult(action, result)

		require.NoError(t, err)
		assert.Contains(t, narration, "使用异常能力打开门")
		assert.Contains(t, narration, "3个\"3\"")
		assert.Contains(t, narration, "成功")
		assert.Contains(t, narration, "门被打开了")
		assert.Contains(t, narration, "发现了新线索")
	})

	t.Run("叙述失败的行动", func(t *testing.T) {
		action := &Action{
			Type:        "request",
			Description: "请求机构改变过去",
			Quality:     "欺瞒",
		}

		result := &ActionResult{
			Success: false,
			Threes:  0,
			Chaos:   6,
			Effects: []string{"请求失败", "地点过载+1"},
		}

		narration, err := service.NarrateResult(action, result)

		require.NoError(t, err)
		assert.Contains(t, narration, "请求机构改变过去")
		assert.Contains(t, narration, "0个\"3\"")
		assert.Contains(t, narration, "失败")
		assert.Contains(t, narration, "6 点混沌")
		assert.Contains(t, narration, "请求失败")
	})

	t.Run("叙述出色的成功", func(t *testing.T) {
		action := &Action{
			Type:        "ability",
			Description: "使用能力",
		}

		result := &ActionResult{
			Success: true,
			Threes:  5, // 很多"3"
			Chaos:   0,
			Effects: []string{},
		}

		narration, err := service.NarrateResult(action, result)

		require.NoError(t, err)
		assert.Contains(t, narration, "出色")
	})

	t.Run("叙述产生大量混沌的失败", func(t *testing.T) {
		action := &Action{
			Type:        "ability",
			Description: "尝试控制异常",
		}

		result := &ActionResult{
			Success: false,
			Threes:  0,
			Chaos:   6, // 大量混沌
			Effects: []string{},
		}

		narration, err := service.NarrateResult(action, result)

		require.NoError(t, err)
		assert.True(t, strings.Contains(narration, "更糟") ||
			strings.Contains(narration, "增强"))
	})

	t.Run("行动为空时返回错误", func(t *testing.T) {
		result := &ActionResult{}

		_, err := service.NarrateResult(nil, result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "行动不能为空")
	})

	t.Run("结果为空时返回错误", func(t *testing.T) {
		action := &Action{
			Description: "测试行动",
		}

		_, err := service.NarrateResult(action, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "结果不能为空")
	})
}
