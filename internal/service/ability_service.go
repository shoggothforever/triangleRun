package service

import (
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// AbilityService 异常能力服务接口
type AbilityService interface {
	// 能力使用
	UseAbility(agent *domain.Agent, session *domain.GameSession, abilityID string, context *AbilityContext) (*AbilityResult, error)

	// 能力验证
	ValidateTrigger(ability *domain.AnomalyAbility, context *AbilityContext) (bool, error)
	CheckCondition(ability *domain.AnomalyAbility, context *AbilityContext) (bool, error)

	// 效果应用
	ApplySuccessEffect(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) (*EffectResult, error)
	ApplyFailureEffect(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) (*EffectResult, error)
	ApplyAdditionalEffects(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) ([]*EffectResult, error)

	// 工作外使用检测
	CheckOffDutyUsage(agent *domain.Agent, session *domain.GameSession) bool
}

// AbilityContext 能力使用上下文
type AbilityContext struct {
	TargetID    string                 // 目标ID（NPC、对象等）
	LocationID  string                 // 当前地点
	OnDuty      bool                   // 是否在工作中
	CustomData  map[string]interface{} // 自定义数据
	Description string                 // 玩家描述的使用方式
}

// AbilityResult 能力使用结果
type AbilityResult struct {
	Ability           *domain.AnomalyAbility `json:"ability"`
	Roll              *domain.RollResult     `json:"roll"`
	Success           bool                   `json:"success"`
	SuccessEffect     *EffectResult          `json:"success_effect,omitempty"`
	FailureEffect     *EffectResult          `json:"failure_effect,omitempty"`
	AdditionalEffects []*EffectResult        `json:"additional_effects,omitempty"`
	ChaosGenerated    int                    `json:"chaos_generated"`
	ReprimandAdded    bool                   `json:"reprimand_added"`
}

// EffectResult 效果结果
type EffectResult struct {
	Description string                 `json:"description"`
	Mechanics   string                 `json:"mechanics"`
	Applied     bool                   `json:"applied"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// abilityService 异常能力服务实现
type abilityService struct {
	diceService  domain.DiceService
	qaService    QAService
	chaosService ChaosService
}

// NewAbilityService 创建异常能力服务
func NewAbilityService(diceService domain.DiceService, qaService QAService, chaosService ChaosService) AbilityService {
	return &abilityService{
		diceService:  diceService,
		qaService:    qaService,
		chaosService: chaosService,
	}
}

// UseAbility 使用异常能力
func (s *abilityService) UseAbility(agent *domain.Agent, session *domain.GameSession, abilityID string, context *AbilityContext) (*AbilityResult, error) {
	// 查找能力
	var ability *domain.AnomalyAbility
	for _, a := range agent.Anomaly.Abilities {
		if a.ID == abilityID {
			ability = a
			break
		}
	}

	if ability == nil {
		return nil, domain.NewGameError(domain.ErrNotFound, "未找到指定的异常能力").
			WithDetails("ability_id", abilityID)
	}

	// 验证触发条件
	valid, err := s.ValidateTrigger(ability, context)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, domain.NewGameError(domain.ErrInvalidAction, "能力触发条件不满足").
			WithDetails("ability", ability.Name)
	}

	// 执行掷骰
	roll := s.diceService.RollForAbility(agent, ability)

	// 添加混沌到混沌池（如果失败）
	if !roll.Success {
		if err := s.chaosService.AddChaosFromRoll(session, roll); err != nil {
			return nil, err
		}
	}

	result := &AbilityResult{
		Ability:        ability,
		Roll:           roll,
		Success:        roll.Success,
		ChaosGenerated: roll.Chaos,
		ReprimandAdded: false,
	}

	// 应用效果
	if roll.Success {
		successEffect, err := s.ApplySuccessEffect(ability, roll, context)
		if err != nil {
			return nil, err
		}
		result.SuccessEffect = successEffect
	} else {
		failureEffect, err := s.ApplyFailureEffect(ability, roll, context)
		if err != nil {
			return nil, err
		}
		result.FailureEffect = failureEffect
	}

	// 应用额外效果
	additionalEffects, err := s.ApplyAdditionalEffects(ability, roll, context)
	if err != nil {
		return nil, err
	}
	result.AdditionalEffects = additionalEffects

	// 检查是否在工作外使用
	if s.CheckOffDutyUsage(agent, session) {
		agent.AddReprimands(1)
		result.ReprimandAdded = true
	}

	return result, nil
}

// ValidateTrigger 验证能力触发条件
func (s *abilityService) ValidateTrigger(ability *domain.AnomalyAbility, context *AbilityContext) (bool, error) {
	if ability.Trigger == nil {
		return true, nil
	}

	// 根据触发类型验证
	switch ability.Trigger.Type {
	case domain.TriggerAction:
		// 主动使用，总是可以触发
		return true, nil

	case domain.TriggerResponse:
		// 响应某事，需要检查上下文
		// 这里简化实现，实际应该检查具体的响应条件
		return context != nil, nil

	case domain.TriggerPassive:
		// 被动触发，需要特定条件
		return s.CheckCondition(ability, context)

	case domain.TriggerReactive:
		// 反应性触发，需要特定事件
		return context != nil && context.CustomData != nil, nil

	default:
		return false, domain.NewGameError(domain.ErrInvalidInput, "未知的触发类型").
			WithDetails("type", ability.Trigger.Type)
	}
}

// CheckCondition 检查能力条件
func (s *abilityService) CheckCondition(ability *domain.AnomalyAbility, context *AbilityContext) (bool, error) {
	if ability.Trigger == nil || ability.Trigger.Condition == "" {
		return true, nil
	}

	// 这里简化实现，实际应该解析和评估条件表达式
	// 例如："当你看到某人受伤时"、"当你进入新地点时"等
	// 目前返回true，表示条件满足
	return true, nil
}

// ApplySuccessEffect 应用成功效果
func (s *abilityService) ApplySuccessEffect(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) (*EffectResult, error) {
	if ability.Effects == nil || ability.Effects.Success == nil {
		return nil, nil
	}

	effect := ability.Effects.Success

	return &EffectResult{
		Description: effect.Description,
		Mechanics:   effect.Mechanics,
		Applied:     true,
		Details: map[string]interface{}{
			"duration": effect.Duration,
			"target":   effect.Target,
			"threes":   roll.Threes,
		},
	}, nil
}

// ApplyFailureEffect 应用失败效果
func (s *abilityService) ApplyFailureEffect(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) (*EffectResult, error) {
	if ability.Effects == nil || ability.Effects.Failure == nil {
		return nil, nil
	}

	effect := ability.Effects.Failure

	return &EffectResult{
		Description: effect.Description,
		Mechanics:   effect.Mechanics,
		Applied:     true,
		Details: map[string]interface{}{
			"duration":       effect.Duration,
			"target":         effect.Target,
			"chaos_produced": roll.Chaos,
		},
	}, nil
}

// ApplyAdditionalEffects 应用额外效果
func (s *abilityService) ApplyAdditionalEffects(ability *domain.AnomalyAbility, roll *domain.RollResult, context *AbilityContext) ([]*EffectResult, error) {
	if ability.Effects == nil || len(ability.Effects.Additional) == 0 {
		return nil, nil
	}

	var results []*EffectResult

	for _, conditional := range ability.Effects.Additional {
		// 检查条件是否满足
		if s.checkAdditionalCondition(conditional.Condition, roll) {
			results = append(results, &EffectResult{
				Description: conditional.Effect.Description,
				Mechanics:   conditional.Effect.Mechanics,
				Applied:     true,
				Details: map[string]interface{}{
					"condition": conditional.Condition,
					"threes":    roll.Threes,
				},
			})
		}
	}

	return results, nil
}

// checkAdditionalCondition 检查额外效果条件
func (s *abilityService) checkAdditionalCondition(condition string, roll *domain.RollResult) bool {
	// 解析条件字符串
	// 例如："每额外一个3"、"每第三个3"、"六个或更多3"

	switch condition {
	case "每额外一个3":
		// 超过1个3时，每多一个3触发一次
		return roll.Threes > 1

	case "每第三个3":
		// 每3个3触发一次
		return roll.Threes >= 3 && roll.Threes%3 == 0

	case "六个或更多3":
		// 6个或更多3
		return roll.Threes >= 6

	case "四个或更多3":
		// 4个或更多3
		return roll.Threes >= 4

	case "两个或更多3":
		// 2个或更多3
		return roll.Threes >= 2

	default:
		// 未知条件，默认不满足
		return false
	}
}

// CheckOffDutyUsage 检查是否在工作外使用能力
func (s *abilityService) CheckOffDutyUsage(agent *domain.Agent, session *domain.GameSession) bool {
	// 如果会话为空或上下文标记为非工作时间，则认为是工作外使用
	if session == nil {
		return true
	}

	// 在晨会和余波阶段使用能力算工作外
	if session.Phase == domain.PhaseMorning || session.Phase == domain.PhaseAftermath {
		return true
	}

	// 在调查和遭遇阶段使用能力算工作中
	return false
}
