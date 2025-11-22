package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// ClueService 线索服务接口
type ClueService interface {
	// 线索追踪
	AddClue(sessionID string, clueID string, source string) error
	GetCollectedClues(sessionID string) ([]*CollectedClue, error)
	HasClue(sessionID string, clueID string) bool

	// 线索解锁
	CheckUnlockConditions(sessionID string, clueID string) (bool, []string, error)
	GetAvailableClues(sessionID string) ([]*domain.Clue, error)
	UnlockLocations(sessionID string, clueID string) ([]string, error)

	// 调查报告
	GenerateInvestigationReport(sessionID string) (*InvestigationReport, error)
	GetMissingClues(sessionID string) ([]*domain.Clue, error)
	GetClueProgress(sessionID string) (*ClueProgress, error)
}

// CollectedClue 已收集的线索
type CollectedClue struct {
	ClueID      string    `json:"clue_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	CollectedAt time.Time `json:"collected_at"`
	Unlocks     []string  `json:"unlocks"`
}

// InvestigationReport 调查报告
type InvestigationReport struct {
	SessionID      string           `json:"session_id"`
	ScenarioName   string           `json:"scenario_name"`
	GeneratedAt    time.Time        `json:"generated_at"`
	CollectedClues []*CollectedClue `json:"collected_clues"`
	MissingClues   []*domain.Clue   `json:"missing_clues"`
	Progress       *ClueProgress    `json:"progress"`
	UnlockedScenes []string         `json:"unlocked_scenes"`
	VisitedScenes  []string         `json:"visited_scenes"`
	Summary        string           `json:"summary"`
}

// ClueProgress 线索进度
type ClueProgress struct {
	TotalClues     int     `json:"total_clues"`
	CollectedCount int     `json:"collected_count"`
	MissingCount   int     `json:"missing_count"`
	Percentage     float64 `json:"percentage"`
}

// clueService 线索服务实现
type clueService struct {
	scenarioService ScenarioService
	gameService     GameService
	clueMetadata    map[string]map[string]*clueMetadata // sessionID -> clueID -> metadata
	mu              sync.RWMutex
}

// clueMetadata 线索元数据
type clueMetadata struct {
	Source      string
	CollectedAt time.Time
}

// NewClueService 创建线索服务
func NewClueService(scenarioService ScenarioService, gameService GameService) ClueService {
	return &clueService{
		scenarioService: scenarioService,
		gameService:     gameService,
		clueMetadata:    make(map[string]map[string]*clueMetadata),
	}
}

// AddClue 添加线索
func (s *clueService) AddClue(sessionID string, clueID string, source string) error {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 检查线索是否已收集
	if s.HasClue(sessionID, clueID) {
		return domain.NewGameError(domain.ErrInvalidAction, "线索已收集").
			WithDetails("clue_id", clueID)
	}

	// 验证线索存在
	clue, err := s.scenarioService.GetClue(session.ScenarioID, clueID)
	if err != nil {
		return err
	}

	// 检查线索需求
	if !s.scenarioService.CheckClueRequirements(clue, session.State) {
		missingReqs := s.getMissingRequirements(clue, session.State)
		return domain.NewGameError(domain.ErrInvalidAction, "线索需求未满足").
			WithDetails("clue_id", clueID).
			WithDetails("missing_requirements", strings.Join(missingReqs, ", "))
	}

	// 添加线索到已收集列表
	if session.State.CollectedClues == nil {
		session.State.CollectedClues = []string{}
	}
	session.State.CollectedClues = append(session.State.CollectedClues, clueID)

	// 解锁新场景
	if session.State.UnlockedLocations == nil {
		session.State.UnlockedLocations = []string{}
	}
	for _, unlockID := range clue.Unlocks {
		if !contains(session.State.UnlockedLocations, unlockID) {
			session.State.UnlockedLocations = append(session.State.UnlockedLocations, unlockID)
		}
	}

	// 保存元数据
	s.mu.Lock()
	if _, exists := s.clueMetadata[sessionID]; !exists {
		s.clueMetadata[sessionID] = make(map[string]*clueMetadata)
	}
	s.clueMetadata[sessionID][clueID] = &clueMetadata{
		Source:      source,
		CollectedAt: time.Now(),
	}
	s.mu.Unlock()

	// 保存会话
	return s.gameService.SaveSession(session)
}

// GetCollectedClues 获取已收集的线索
func (s *clueService) GetCollectedClues(sessionID string) ([]*CollectedClue, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.State.CollectedClues == nil {
		return []*CollectedClue{}, nil
	}

	collectedClues := make([]*CollectedClue, 0, len(session.State.CollectedClues))

	for _, clueID := range session.State.CollectedClues {
		// 获取线索详情
		clue, err := s.scenarioService.GetClue(session.ScenarioID, clueID)
		if err != nil {
			continue // 跳过无效线索
		}

		// 获取元数据
		s.mu.RLock()
		metadata := s.clueMetadata[sessionID][clueID]
		s.mu.RUnlock()

		collected := &CollectedClue{
			ClueID:      clue.ID,
			Name:        clue.Name,
			Description: clue.Description,
			Unlocks:     clue.Unlocks,
		}

		if metadata != nil {
			collected.Source = metadata.Source
			collected.CollectedAt = metadata.CollectedAt
		}

		collectedClues = append(collectedClues, collected)
	}

	return collectedClues, nil
}

// HasClue 检查是否拥有线索
func (s *clueService) HasClue(sessionID string, clueID string) bool {
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return false
	}

	if session.State.CollectedClues == nil {
		return false
	}

	return contains(session.State.CollectedClues, clueID)
}

// CheckUnlockConditions 检查解锁条件
func (s *clueService) CheckUnlockConditions(sessionID string, clueID string) (bool, []string, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return false, nil, err
	}

	// 获取线索
	clue, err := s.scenarioService.GetClue(session.ScenarioID, clueID)
	if err != nil {
		return false, nil, err
	}

	// 检查需求
	if len(clue.Requirements) == 0 {
		return true, []string{}, nil
	}

	missingReqs := s.getMissingRequirements(clue, session.State)
	return len(missingReqs) == 0, missingReqs, nil
}

// GetAvailableClues 获取可用线索
func (s *clueService) GetAvailableClues(sessionID string) ([]*domain.Clue, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取当前场景
	if session.State.CurrentSceneID == "" {
		return []*domain.Clue{}, nil
	}

	scene, err := s.scenarioService.GetScene(session.ScenarioID, session.State.CurrentSceneID)
	if err != nil {
		return nil, err
	}

	availableClues := make([]*domain.Clue, 0)

	for _, clue := range scene.Clues {
		// 跳过已收集的线索
		if s.HasClue(sessionID, clue.ID) {
			continue
		}

		// 检查需求
		if s.scenarioService.CheckClueRequirements(clue, session.State) {
			availableClues = append(availableClues, clue)
		}
	}

	return availableClues, nil
}

// UnlockLocations 解锁地点
func (s *clueService) UnlockLocations(sessionID string, clueID string) ([]string, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取线索
	clue, err := s.scenarioService.GetClue(session.ScenarioID, clueID)
	if err != nil {
		return nil, err
	}

	// 检查线索是否已收集
	if !s.HasClue(sessionID, clueID) {
		return nil, domain.NewGameError(domain.ErrInvalidAction, "线索未收集").
			WithDetails("clue_id", clueID)
	}

	// 返回解锁的地点
	return clue.Unlocks, nil
}

// GenerateInvestigationReport 生成调查报告
func (s *clueService) GenerateInvestigationReport(sessionID string) (*InvestigationReport, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 加载剧本
	scenario, err := s.scenarioService.LoadScenario(session.ScenarioID)
	if err != nil {
		return nil, err
	}

	// 获取已收集的线索
	collectedClues, err := s.GetCollectedClues(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取遗漏的线索
	missingClues, err := s.GetMissingClues(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取进度
	progress, err := s.GetClueProgress(sessionID)
	if err != nil {
		return nil, err
	}

	// 获取已访问的场景
	visitedScenes := make([]string, 0)
	if session.State.VisitedScenes != nil {
		for sceneID, visited := range session.State.VisitedScenes {
			if visited {
				visitedScenes = append(visitedScenes, sceneID)
			}
		}
	}
	sort.Strings(visitedScenes)

	// 生成摘要
	summary := s.generateSummary(scenario, collectedClues, missingClues, progress)

	report := &InvestigationReport{
		SessionID:      sessionID,
		ScenarioName:   scenario.Name,
		GeneratedAt:    time.Now(),
		CollectedClues: collectedClues,
		MissingClues:   missingClues,
		Progress:       progress,
		UnlockedScenes: session.State.UnlockedLocations,
		VisitedScenes:  visitedScenes,
		Summary:        summary,
	}

	return report, nil
}

// GetMissingClues 获取遗漏的线索
func (s *clueService) GetMissingClues(sessionID string) ([]*domain.Clue, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 加载剧本
	scenario, err := s.scenarioService.LoadScenario(session.ScenarioID)
	if err != nil {
		return nil, err
	}

	// 收集所有线索
	allClues := make(map[string]*domain.Clue)
	for _, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			allClues[clue.ID] = clue
		}
	}

	// 找出遗漏的线索
	missingClues := make([]*domain.Clue, 0)
	for clueID, clue := range allClues {
		if !s.HasClue(sessionID, clueID) {
			missingClues = append(missingClues, clue)
		}
	}

	// 按ID排序
	sort.Slice(missingClues, func(i, j int) bool {
		return missingClues[i].ID < missingClues[j].ID
	})

	return missingClues, nil
}

// GetClueProgress 获取线索进度
func (s *clueService) GetClueProgress(sessionID string) (*ClueProgress, error) {
	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 加载剧本
	scenario, err := s.scenarioService.LoadScenario(session.ScenarioID)
	if err != nil {
		return nil, err
	}

	// 统计总线索数
	allClues := make(map[string]bool)
	for _, scene := range scenario.Scenes {
		for _, clue := range scene.Clues {
			allClues[clue.ID] = true
		}
	}
	totalClues := len(allClues)

	// 统计已收集线索数
	collectedCount := 0
	if session.State.CollectedClues != nil {
		collectedCount = len(session.State.CollectedClues)
	}

	// 计算百分比
	percentage := 0.0
	if totalClues > 0 {
		percentage = float64(collectedCount) / float64(totalClues) * 100
	}

	progress := &ClueProgress{
		TotalClues:     totalClues,
		CollectedCount: collectedCount,
		MissingCount:   totalClues - collectedCount,
		Percentage:     percentage,
	}

	return progress, nil
}

// getMissingRequirements 获取缺失的需求
func (s *clueService) getMissingRequirements(clue *domain.Clue, state *domain.GameState) []string {
	missing := make([]string, 0)

	for _, req := range clue.Requirements {
		found := false
		if state.CollectedClues != nil {
			for _, collected := range state.CollectedClues {
				if collected == req {
					found = true
					break
				}
			}
		}
		if !found {
			missing = append(missing, req)
		}
	}

	return missing
}

// generateSummary 生成摘要
func (s *clueService) generateSummary(scenario *domain.Scenario, collected []*CollectedClue, missing []*domain.Clue, progress *ClueProgress) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("调查剧本：%s\n\n", scenario.Name))
	summary.WriteString(fmt.Sprintf("线索收集进度：%d/%d (%.1f%%)\n\n",
		progress.CollectedCount, progress.TotalClues, progress.Percentage))

	if len(collected) > 0 {
		summary.WriteString("已收集线索：\n")
		for i, clue := range collected {
			summary.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, clue.Name, clue.Description))
			if clue.Source != "" {
				summary.WriteString(fmt.Sprintf("   来源：%s\n", clue.Source))
			}
			if len(clue.Unlocks) > 0 {
				summary.WriteString(fmt.Sprintf("   解锁：%s\n", strings.Join(clue.Unlocks, ", ")))
			}
		}
		summary.WriteString("\n")
	}

	if len(missing) > 0 {
		summary.WriteString(fmt.Sprintf("遗漏线索：%d条\n", len(missing)))
		if len(missing) <= 5 {
			for i, clue := range missing {
				summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, clue.Name))
			}
		} else {
			summary.WriteString("（线索过多，请查看详细报告）\n")
		}
	}

	return summary.String()
}
