package service

import (
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// DamageService 伤害服务接口
type DamageService interface {
	// ApplyDamage 应用伤害
	// 返回是否死亡、是否使用了人寿保险、产生的散逸端数量
	ApplyDamage(agent *domain.Agent, damage int, hasWitnesses bool) (died bool, usedInsurance bool, looseEnds int, err error)

	// UseLifeInsurance 使用人寿保险无视伤害
	UseLifeInsurance(agent *domain.Agent, damage int) error

	// HandleDeath 处理死亡
	HandleDeath(agent *domain.Agent) error

	// Revive 复活角色
	Revive(agent *domain.Agent) error

	// GenerateLooseEnds 生成散逸端
	GenerateLooseEnds(damage int, hasWitnesses bool) int

	// CheckDebt 检查嘉奖负债状态
	CheckDebt(agent *domain.Agent) bool

	// CanAffordInsurance 检查是否能支付人寿保险
	CanAffordInsurance(agent *domain.Agent, damage int) bool
}

type damageService struct{}

// NewDamageService 创建伤害服务
func NewDamageService() DamageService {
	return &damageService{}
}

// ApplyDamage 应用伤害
func (s *damageService) ApplyDamage(agent *domain.Agent, damage int, hasWitnesses bool) (died bool, usedInsurance bool, looseEnds int, err error) {
	if damage <= 0 {
		return false, false, 0, nil
	}

	// 检查是否能使用人寿保险
	canAfford := s.CanAffordInsurance(agent, damage)

	// 如果能支付且选择使用人寿保险（这里默认使用）
	if canAfford {
		err = s.UseLifeInsurance(agent, damage)
		if err != nil {
			return false, false, 0, err
		}
		return false, true, 0, nil
	}

	// 无法或不愿使用人寿保险，角色死亡
	err = s.HandleDeath(agent)
	if err != nil {
		return true, false, 0, err
	}

	// 生成散逸端
	looseEnds = s.GenerateLooseEnds(damage, hasWitnesses)

	return true, false, looseEnds, nil
}

// UseLifeInsurance 使用人寿保险无视伤害
func (s *damageService) UseLifeInsurance(agent *domain.Agent, damage int) error {
	if damage <= 0 {
		return nil
	}

	// 需要花费等量的QA（可以从任意资质中扣除）
	totalQA := agent.TotalQA()
	if totalQA < damage {
		return domain.NewGameError(domain.ErrInsufficientQA, "资质保证不足以支付人寿保险").
			WithDetails("required", damage).
			WithDetails("available", totalQA)
	}

	// 从QA中扣除（优先从高的资质扣除）
	remaining := damage
	for remaining > 0 {
		// 找到QA最高的资质
		maxQuality := ""
		maxQA := 0
		for quality, qa := range agent.QA {
			if qa > maxQA {
				maxQA = qa
				maxQuality = quality
			}
		}

		if maxQA == 0 {
			// 理论上不应该到这里，因为前面已经检查过总QA
			return domain.NewGameError(domain.ErrInsufficientQA, "资质保证不足")
		}

		// 从这个资质中扣除
		deduct := remaining
		if deduct > maxQA {
			deduct = maxQA
		}

		agent.QA[maxQuality] -= deduct
		remaining -= deduct
	}

	return nil
}

// HandleDeath 处理死亡
func (s *damageService) HandleDeath(agent *domain.Agent) error {
	// 标记为死亡
	agent.Alive = false

	// 扣除5次嘉奖
	agent.Commendations -= 5

	// 检查是否进入嘉奖负债
	if agent.Commendations < 0 {
		agent.InDebt = true
	}

	// 立即复活
	return s.Revive(agent)
}

// Revive 复活角色
func (s *damageService) Revive(agent *domain.Agent) error {
	// 在分部休息室复活
	agent.Alive = true
	// 保留死亡前的记忆和状态
	return nil
}

// GenerateLooseEnds 生成散逸端
func (s *damageService) GenerateLooseEnds(damage int, hasWitnesses bool) int {
	// 如果伤害超过1点且有目击者，产生等于伤害点数的散逸端
	if damage > 1 && hasWitnesses {
		return damage
	}
	return 0
}

// CheckDebt 检查嘉奖负债状态
func (s *damageService) CheckDebt(agent *domain.Agent) bool {
	return agent.Commendations < 0
}

// CanAffordInsurance 检查是否能支付人寿保险
func (s *damageService) CanAffordInsurance(agent *domain.Agent, damage int) bool {
	return agent.TotalQA() >= damage
}
