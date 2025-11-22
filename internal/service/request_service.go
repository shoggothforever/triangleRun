package service

import (
	"strings"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// RequestService 请求机构服务接口
type RequestService interface {
	// 现实变更请求
	ValidateRequest(req *RealityChangeRequest) error
	ProcessRequest(agent *domain.Agent, session *domain.GameSession, req *RealityChangeRequest, roll *domain.RollResult) (*RequestResult, error)

	// 既定事实管理
	AddEstablishedFact(session *domain.GameSession, fact string) error
	IsEstablishedFact(session *domain.GameSession, fact string) bool
	GetEstablishedFacts(session *domain.GameSession) []string

	// 因果链检查
	ValidateCausalChain(chain string) error

	// 心智控制检测
	IsMindControl(effect string) bool
}

// RealityChangeRequest 现实变更请求
type RealityChangeRequest struct {
	Effect      string // 既定效果
	CausalChain string // 因果链
	Quality     string // 相关资质
	LocationID  string // 当前地点
}

// RequestResult 请求结果
type RequestResult struct {
	Success         bool   // 是否成功
	AppliedEffect   string // 应用的效果
	EstablishedFact string // 确立的事实
	Chaos           int    // 产生的混沌
	Overload        int    // 地点过载
}

// requestService 请求机构服务实现
type requestService struct {
	diceService  domain.DiceService
	chaosService ChaosService
}

// NewRequestService 创建请求机构服务
func NewRequestService(diceService domain.DiceService, chaosService ChaosService) RequestService {
	return &requestService{
		diceService:  diceService,
		chaosService: chaosService,
	}
}

// ValidateRequest 验证请求的完整性
// 需求6.1: 要求玩家提供既定效果、因果链、相关资质和掷骰
func (s *requestService) ValidateRequest(req *RealityChangeRequest) error {
	if req == nil {
		return domain.NewGameError(domain.ErrInvalidInput, "请求不能为空")
	}

	// 检查既定效果
	if strings.TrimSpace(req.Effect) == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "必须提供既定效果")
	}

	// 检查因果链
	if strings.TrimSpace(req.CausalChain) == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "必须提供因果链")
	}

	// 检查相关资质
	if strings.TrimSpace(req.Quality) == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "必须提供相关资质")
	}

	// 验证资质名称是否有效
	validQualities := []string{"专注", "共情", "气场", "欺瞒", "主动", "专业", "活力", "坚毅", "诡秘"}
	isValid := false
	for _, q := range validQualities {
		if req.Quality == q {
			isValid = true
			break
		}
	}
	if !isValid {
		return domain.NewGameError(domain.ErrInvalidInput, "无效的资质名称").
			WithDetails("quality", req.Quality)
	}

	// 检查因果链是否有效
	if err := s.ValidateCausalChain(req.CausalChain); err != nil {
		return err
	}

	// 需求6.5: 检查是否为心智控制
	if s.IsMindControl(req.Effect) {
		return domain.NewGameError(domain.ErrInvalidAction, "不能直接进行心智控制，因果链必须是外在的")
	}

	return nil
}

// ProcessRequest 处理现实变更请求
func (s *requestService) ProcessRequest(agent *domain.Agent, session *domain.GameSession, req *RealityChangeRequest, roll *domain.RollResult) (*RequestResult, error) {
	if agent == nil || session == nil || req == nil || roll == nil {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "参数不能为空")
	}

	// 验证请求
	if err := s.ValidateRequest(req); err != nil {
		return nil, err
	}

	// 需求6.4: 检查是否尝试改变已确立的事实
	if s.IsEstablishedFact(session, req.Effect) {
		return nil, domain.NewGameError(domain.ErrInvalidAction, "不能改变已确立的事实").
			WithDetails("fact", req.Effect)
	}

	result := &RequestResult{
		Success:  roll.Success,
		Chaos:    0,
		Overload: 0,
	}

	if roll.Success {
		// 需求6.2: 成功时应用玩家描述的效果并确立为既定事实
		result.AppliedEffect = req.Effect
		result.EstablishedFact = req.Effect

		// 将效果确立为既定事实
		if err := s.AddEstablishedFact(session, req.Effect); err != nil {
			return nil, err
		}
	} else {
		// 需求6.3: 失败时应用反向效果并为该地点添加过载
		result.AppliedEffect = "GM描述的反向效果"
		result.Chaos = roll.Chaos

		// 为当前地点添加过载
		if req.LocationID != "" {
			if err := s.chaosService.AddLocationOverload(session, req.LocationID); err != nil {
				return nil, err
			}
			result.Overload = s.chaosService.GetLocationOverload(session, req.LocationID)
		}

		// 添加混沌到混沌池
		if err := s.chaosService.AddChaosFromRoll(session, roll); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// AddEstablishedFact 添加既定事实
func (s *requestService) AddEstablishedFact(session *domain.GameSession, fact string) error {
	if session == nil || session.State == nil {
		return domain.NewGameError(domain.ErrInvalidState, "游戏会话或状态为空")
	}

	fact = strings.TrimSpace(fact)
	if fact == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "事实不能为空")
	}

	// 初始化既定事实列表（使用CollectedClues字段暂存，实际应该有专门的字段）
	// 这里为了简化，我们使用一个特殊的前缀来标记既定事实
	factKey := "ESTABLISHED_FACT:" + fact

	// 检查是否已存在
	for _, clue := range session.State.CollectedClues {
		if clue == factKey {
			return nil // 已存在，不重复添加
		}
	}

	session.State.CollectedClues = append(session.State.CollectedClues, factKey)
	return nil
}

// IsEstablishedFact 检查是否为既定事实
func (s *requestService) IsEstablishedFact(session *domain.GameSession, fact string) bool {
	if session == nil || session.State == nil {
		return false
	}

	fact = strings.TrimSpace(fact)
	factKey := "ESTABLISHED_FACT:" + fact

	for _, clue := range session.State.CollectedClues {
		if clue == factKey {
			return true
		}
	}

	return false
}

// GetEstablishedFacts 获取所有既定事实
func (s *requestService) GetEstablishedFacts(session *domain.GameSession) []string {
	if session == nil || session.State == nil {
		return []string{}
	}

	facts := []string{}
	prefix := "ESTABLISHED_FACT:"

	for _, clue := range session.State.CollectedClues {
		if strings.HasPrefix(clue, prefix) {
			fact := strings.TrimPrefix(clue, prefix)
			facts = append(facts, fact)
		}
	}

	return facts
}

// ValidateCausalChain 验证因果链
func (s *requestService) ValidateCausalChain(chain string) error {
	chain = strings.TrimSpace(chain)

	if chain == "" {
		return domain.NewGameError(domain.ErrInvalidInput, "因果链不能为空")
	}

	// 因果链应该描述一个合理的因果关系
	// 这里做基本的长度检查，实际游戏中可能需要AI来判断
	if len(chain) < 10 {
		return domain.NewGameError(domain.ErrInvalidInput, "因果链描述过于简短，请提供详细的因果关系")
	}

	return nil
}

// IsMindControl 检测是否为直接心智控制
// 需求6.5: 直接心智控制必须被拒绝
func (s *requestService) IsMindControl(effect string) bool {
	effectLower := strings.ToLower(strings.TrimSpace(effect))

	// 检测直接心智控制的关键词
	mindControlKeywords := []string{
		"控制",
		"操纵",
		"强迫",
		"命令",
		"洗脑",
		"催眠",
		"支配",
		"迫使",
		"让他想",
		"让她想",
		"让他们想",
	}

	// 检测"改变"+"想法/思想/意志"的组合
	if strings.Contains(effectLower, "改变") {
		if strings.Contains(effectLower, "想法") ||
			strings.Contains(effectLower, "思想") ||
			strings.Contains(effectLower, "意志") {
			return true
		}
	}

	for _, keyword := range mindControlKeywords {
		if strings.Contains(effectLower, keyword) {
			return true
		}
	}

	return false
}
