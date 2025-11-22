package service

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// AIService AI服务接口
type AIService interface {
	// 场景描述生成
	GenerateSceneDescription(scene *domain.Scene, state *domain.GameState) (string, error)

	// NPC对话生成
	GenerateNPCDialogue(npc *domain.NPC, context *DialogueContext) (string, error)

	// 混沌效应决策
	SelectChaosEffect(anomaly *domain.AnomalyProfile, chaosPool int, context *domain.GameState) (*domain.ChaosEffect, error)

	// 事件叙述
	DescribeEvent(event *domain.Event, context *domain.GameState) (string, error)

	// 结果描述
	NarrateResult(action *Action, result *ActionResult) (string, error)
}

// DialogueContext NPC对话上下文
type DialogueContext struct {
	PlayerAction string                 `json:"player_action"`
	SceneID      string                 `json:"scene_id"`
	GamePhase    domain.GamePhase       `json:"game_phase"`
	NPCState     *domain.NPCState       `json:"npc_state"`
	CustomData   map[string]interface{} `json:"custom_data"`
}

// Action 玩家行动
type Action struct {
	Type        string                 `json:"type"`
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	Quality     string                 `json:"quality"`
	CustomData  map[string]interface{} `json:"custom_data"`
}

// ActionResult 行动结果
type ActionResult struct {
	Success    bool                   `json:"success"`
	Threes     int                    `json:"threes"`
	Chaos      int                    `json:"chaos"`
	Effects    []string               `json:"effects"`
	CustomData map[string]interface{} `json:"custom_data"`
}

// aiService AI服务实现
type aiService struct {
	// 在实际实现中，这里会包含AI API客户端
	// 例如: openaiClient *openai.Client
	// 目前使用模板化的响应
}

// NewAIService 创建AI服务
func NewAIService() AIService {
	return &aiService{}
}

// GenerateSceneDescription 生成场景描述
func (s *aiService) GenerateSceneDescription(scene *domain.Scene, state *domain.GameState) (string, error) {
	if scene == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "场景不能为空")
	}

	// 构建场景描述
	var description strings.Builder

	// 基础场景描述
	description.WriteString(fmt.Sprintf("【%s】\n\n", scene.Name))
	description.WriteString(scene.Description)
	description.WriteString("\n\n")

	// 添加场景状态信息
	if state != nil {
		// 检查是否首次访问
		if !state.VisitedScenes[scene.ID] {
			description.WriteString("这是你第一次来到这里。")
		} else {
			description.WriteString("你再次来到这个熟悉的地方。")
		}
		description.WriteString("\n\n")

		// 添加NPC信息
		if len(scene.NPCs) > 0 {
			description.WriteString("你注意到以下人物：\n")
			for _, npc := range scene.NPCs {
				npcState, exists := state.NPCStates[npc.ID]
				if exists && npcState.AnomalyAffected {
					description.WriteString(fmt.Sprintf("- %s（似乎有些不对劲）\n", npc.Name))
				} else {
					description.WriteString(fmt.Sprintf("- %s\n", npc.Name))
				}
			}
			description.WriteString("\n")
		}

		// 添加可用线索提示
		if len(scene.Clues) > 0 {
			availableClues := 0
			for _, clue := range scene.Clues {
				// 检查线索是否已收集
				collected := false
				for _, collectedID := range state.CollectedClues {
					if collectedID == clue.ID {
						collected = true
						break
					}
				}
				if !collected {
					availableClues++
				}
			}
			if availableClues > 0 {
				description.WriteString(fmt.Sprintf("这里似乎有 %d 条线索等待发现。\n", availableClues))
			}
		}
	}

	return description.String(), nil
}

// GenerateNPCDialogue 生成NPC对话
func (s *aiService) GenerateNPCDialogue(npc *domain.NPC, context *DialogueContext) (string, error) {
	if npc == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "NPC不能为空")
	}

	if context == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "对话上下文不能为空")
	}

	var dialogue strings.Builder

	// NPC名称和状态
	dialogue.WriteString(fmt.Sprintf("【%s】", npc.Name))
	if context.NPCState != nil && context.NPCState.AnomalyAffected {
		dialogue.WriteString("（受异常影响）")
	}
	dialogue.WriteString("\n\n")

	// 根据NPC状态和玩家行动生成对话
	if context.NPCState != nil && context.NPCState.AnomalyAffected {
		// 受异常影响的对话
		dialogue.WriteString(generateAffectedDialogue(npc, context))
	} else if len(npc.Dialogues) > 0 {
		// 使用预设对话
		dialogueIndex := rand.Intn(len(npc.Dialogues))
		dialogue.WriteString(npc.Dialogues[dialogueIndex])
	} else {
		// 生成基于性格的对话
		dialogue.WriteString(generatePersonalityDialogue(npc, context))
	}

	return dialogue.String(), nil
}

// SelectChaosEffect 选择混沌效应
func (s *aiService) SelectChaosEffect(anomaly *domain.AnomalyProfile, chaosPool int, context *domain.GameState) (*domain.ChaosEffect, error) {
	if anomaly == nil {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "异常体档案不能为空")
	}

	if len(anomaly.ChaosEffects) == 0 {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "异常体没有可用的混沌效应")
	}

	// 筛选可用的混沌效应（成本不超过混沌池）
	availableEffects := make([]*domain.ChaosEffect, 0)
	for _, effect := range anomaly.ChaosEffects {
		if effect.Cost <= chaosPool {
			availableEffects = append(availableEffects, effect)
		}
	}

	if len(availableEffects) == 0 {
		return nil, domain.NewGameError(domain.ErrInsufficientChaos, "混沌池不足以使用任何效应").
			WithDetails("chaos_pool", chaosPool).
			WithDetails("min_cost", getMinCost(anomaly.ChaosEffects))
	}

	// 根据游戏状态选择最合适的效应
	// 在实际实现中，这里会使用AI来做决策
	// 目前使用简单的启发式规则
	selectedEffect := selectBestEffect(availableEffects, context)

	return selectedEffect, nil
}

// DescribeEvent 描述事件
func (s *aiService) DescribeEvent(event *domain.Event, context *domain.GameState) (string, error) {
	if event == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "事件不能为空")
	}

	var description strings.Builder

	// 事件标题
	description.WriteString(fmt.Sprintf("【事件：%s】\n\n", event.Name))

	// 事件描述
	description.WriteString(event.Description)
	description.WriteString("\n\n")

	// 事件效果
	if event.Effect != "" {
		description.WriteString("效果：")
		description.WriteString(event.Effect)
		description.WriteString("\n")
	}

	// 根据游戏状态添加额外信息
	if context != nil {
		if context.ChaosPool > 5 {
			description.WriteString("\n你感觉到周围的异常能量正在增强...")
		}
		if context.DomainUnlocked {
			description.WriteString("\n异常体的领域已经显现。")
		}
	}

	return description.String(), nil
}

// NarrateResult 叙述行动结果
func (s *aiService) NarrateResult(action *Action, result *ActionResult) (string, error) {
	if action == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "行动不能为空")
	}

	if result == nil {
		return "", domain.NewGameError(domain.ErrInvalidInput, "结果不能为空")
	}

	var narration strings.Builder

	// 行动描述
	narration.WriteString(fmt.Sprintf("你尝试%s", action.Description))
	if action.Target != "" {
		narration.WriteString(fmt.Sprintf("（目标：%s）", action.Target))
	}
	narration.WriteString("...\n\n")

	// 掷骰结果
	narration.WriteString(fmt.Sprintf("掷骰结果：%d个\"3\"", result.Threes))
	if result.Success {
		narration.WriteString(" - 成功！\n\n")
	} else {
		narration.WriteString(" - 失败。\n\n")
	}

	// 混沌生成
	if result.Chaos > 0 {
		narration.WriteString(fmt.Sprintf("产生了 %d 点混沌。\n", result.Chaos))
	}

	// 效果描述
	if len(result.Effects) > 0 {
		narration.WriteString("\n效果：\n")
		for _, effect := range result.Effects {
			narration.WriteString(fmt.Sprintf("- %s\n", effect))
		}
	}

	// 成功/失败的叙事
	if result.Success {
		narration.WriteString("\n")
		narration.WriteString(generateSuccessNarration(action, result))
	} else {
		narration.WriteString("\n")
		narration.WriteString(generateFailureNarration(action, result))
	}

	return narration.String(), nil
}

// 辅助函数

// generateAffectedDialogue 生成受异常影响的对话
func generateAffectedDialogue(npc *domain.NPC, context *DialogueContext) string {
	affectedDialogues := []string{
		"...我...我不记得了...",
		"那个声音...一直在我脑海里...",
		"你也听到了吗？那个声音...",
		"我感觉...不太对劲...",
		"这里...这里发生了什么？",
	}

	return affectedDialogues[rand.Intn(len(affectedDialogues))]
}

// generatePersonalityDialogue 根据性格生成对话
func generatePersonalityDialogue(npc *domain.NPC, context *DialogueContext) string {
	// 根据NPC性格生成对话
	personality := strings.ToLower(npc.Personality)

	if strings.Contains(personality, "友好") || strings.Contains(personality, "热情") {
		return fmt.Sprintf("你好！我是%s。有什么我可以帮助你的吗？", npc.Name)
	} else if strings.Contains(personality, "冷漠") || strings.Contains(personality, "疏远") {
		return "...有事吗？"
	} else if strings.Contains(personality, "紧张") || strings.Contains(personality, "焦虑") {
		return "你...你是谁？发生什么事了吗？"
	} else if strings.Contains(personality, "可疑") || strings.Contains(personality, "警惕") {
		return "我不认识你。你来这里做什么？"
	}

	// 默认对话
	return fmt.Sprintf("我是%s。", npc.Name)
}

// selectBestEffect 选择最佳混沌效应
func selectBestEffect(effects []*domain.ChaosEffect, context *domain.GameState) *domain.ChaosEffect {
	if len(effects) == 0 {
		return nil
	}

	// 简单策略：优先选择成本较高的效应（通常更强力）
	bestEffect := effects[0]
	for _, effect := range effects {
		if effect.Cost > bestEffect.Cost {
			bestEffect = effect
		}
	}

	return bestEffect
}

// getMinCost 获取最小成本
func getMinCost(effects []*domain.ChaosEffect) int {
	if len(effects) == 0 {
		return 0
	}

	minCost := effects[0].Cost
	for _, effect := range effects {
		if effect.Cost < minCost {
			minCost = effect.Cost
		}
	}

	return minCost
}

// generateSuccessNarration 生成成功叙事
func generateSuccessNarration(action *Action, result *ActionResult) string {
	successNarrations := []string{
		"你的行动取得了成功。",
		"一切都按照计划进行。",
		"你成功地完成了这个行动。",
		"你的技能和判断力得到了回报。",
	}

	narration := successNarrations[rand.Intn(len(successNarrations))]

	// 根据"3"的数量添加额外描述
	if result.Threes >= 4 {
		narration += " 这是一次出色的表现！"
	} else if result.Threes >= 3 {
		narration += " 你的表现相当不错。"
	}

	return narration
}

// generateFailureNarration 生成失败叙事
func generateFailureNarration(action *Action, result *ActionResult) string {
	failureNarrations := []string{
		"事情没有按照你预期的方向发展。",
		"你的尝试失败了。",
		"情况变得更加复杂了。",
		"这次行动没有成功。",
	}

	narration := failureNarrations[rand.Intn(len(failureNarrations))]

	// 根据混沌添加额外描述
	if result.Chaos >= 5 {
		narration += " 而且情况变得更糟了..."
	} else if result.Chaos >= 3 {
		narration += " 异常能量正在增强。"
	}

	return narration
}
