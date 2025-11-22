package service

import (
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// PerformanceService 绩效服务接口
type PerformanceService interface {
	// 嘉奖管理
	AddCommendations(agent *domain.Agent, amount int) error
	GetCommendations(agent *domain.Agent) int

	// 申诫管理
	AddReprimands(agent *domain.Agent, amount int) error
	GetReprimands(agent *domain.Agent) int

	// 机构评级
	UpdateRating(agent *domain.Agent) error
	GetRating(agent *domain.Agent) string
	CalculateRating(reprimands int) string

	// 任务奖励
	AwardMissionSuccess(agent *domain.Agent, outcome string) error
	AwardCaptureBonus(agent *domain.Agent) error
	AwardNeutralizationPenalty(agent *domain.Agent) error
	AwardEscapePenalty(agent *domain.Agent) error

	// 负债检查
	CheckDebt(agent *domain.Agent) bool
	UpdateDebtStatus(agent *domain.Agent) error
}

// MissionOutcome 任务结果常量
const (
	OutcomeCaptured    = "已捕获" // 捕获异常体
	OutcomeNeutralized = "已中和" // 中和异常体
	OutcomeEscaped     = "已逃脱" // 异常体逃脱
)

// performanceService 绩效服务实现
type performanceService struct{}

// NewPerformanceService 创建绩效服务
func NewPerformanceService() PerformanceService {
	return &performanceService{}
}

// AddCommendations 添加嘉奖
func (s *performanceService) AddCommendations(agent *domain.Agent, amount int) error {
	if amount < 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "嘉奖数量不能为负数").
			WithDetails("amount", amount)
	}

	agent.AddCommendations(amount)
	return s.UpdateDebtStatus(agent)
}

// GetCommendations 获取嘉奖数量
func (s *performanceService) GetCommendations(agent *domain.Agent) int {
	return agent.Commendations
}

// AddReprimands 添加申诫并更新评级
func (s *performanceService) AddReprimands(agent *domain.Agent, amount int) error {
	if amount < 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "申诫数量不能为负数").
			WithDetails("amount", amount)
	}

	agent.AddReprimands(amount)
	return s.UpdateRating(agent)
}

// GetReprimands 获取申诫数量
func (s *performanceService) GetReprimands(agent *domain.Agent) int {
	return agent.Reprimands
}

// UpdateRating 更新机构评级
func (s *performanceService) UpdateRating(agent *domain.Agent) error {
	agent.Rating = s.CalculateRating(agent.Reprimands)
	return nil
}

// GetRating 获取机构评级
func (s *performanceService) GetRating(agent *domain.Agent) string {
	return agent.Rating
}

// CalculateRating 根据申诫数量计算评级
func (s *performanceService) CalculateRating(reprimands int) string {
	return domain.GetRating(reprimands)
}

// AwardMissionSuccess 根据任务结果授予奖励
func (s *performanceService) AwardMissionSuccess(agent *domain.Agent, outcome string) error {
	switch outcome {
	case OutcomeCaptured:
		return s.AwardCaptureBonus(agent)
	case OutcomeNeutralized:
		// 中和异常体无奖惩
		return nil
	case OutcomeEscaped:
		return s.AwardEscapePenalty(agent)
	default:
		return domain.NewGameError(domain.ErrInvalidInput, "无效的任务结果").
			WithDetails("outcome", outcome)
	}
}

// AwardCaptureBonus 捕获异常体奖励（3次嘉奖）
func (s *performanceService) AwardCaptureBonus(agent *domain.Agent) error {
	return s.AddCommendations(agent, 3)
}

// AwardNeutralizationPenalty 中和异常体（无奖惩）
func (s *performanceService) AwardNeutralizationPenalty(agent *domain.Agent) error {
	// 中和异常体无奖惩
	return nil
}

// AwardEscapePenalty 异常体逃脱惩罚（3次申诫）
func (s *performanceService) AwardEscapePenalty(agent *domain.Agent) error {
	return s.AddReprimands(agent, 3)
}

// CheckDebt 检查是否处于负债状态
func (s *performanceService) CheckDebt(agent *domain.Agent) bool {
	return agent.Commendations < 0
}

// UpdateDebtStatus 更新负债状态
func (s *performanceService) UpdateDebtStatus(agent *domain.Agent) error {
	agent.InDebt = s.CheckDebt(agent)
	return nil
}
