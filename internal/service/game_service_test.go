package service

import (
	"sync"
	"testing"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

func TestGameService_CreateSession(t *testing.T) {
	service := NewGameService()

	agentID := "test-agent-id"
	scenarioID := "test-scenario-id"

	session, err := service.CreateSession(agentID, scenarioID)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 验证基本信息
	if session.AgentID != agentID {
		t.Errorf("期望AgentID为 %s, 得到 %s", agentID, session.AgentID)
	}

	if session.ScenarioID != scenarioID {
		t.Errorf("期望ScenarioID为 %s, 得到 %s", scenarioID, session.ScenarioID)
	}

	// 验证初始阶段
	if session.Phase != domain.PhaseMorning {
		t.Errorf("期望初始阶段为 %s, 得到 %s", domain.PhaseMorning, session.Phase)
	}

	// 验证初始状态
	if session.State == nil {
		t.Fatal("期望状态不为nil")
	}

	if session.State.ChaosPool != 0 {
		t.Errorf("期望初始混沌池为0, 得到 %d", session.State.ChaosPool)
	}

	if session.State.LooseEnds != 0 {
		t.Errorf("期望初始散逸端为0, 得到 %d", session.State.LooseEnds)
	}

	if session.State.DomainUnlocked {
		t.Error("期望初始领域未解锁")
	}

	if len(session.State.CollectedClues) != 0 {
		t.Errorf("期望初始线索数为0, 得到 %d", len(session.State.CollectedClues))
	}

	if session.State.AnomalyStatus != "未知" {
		t.Errorf("期望初始异常体状态为'未知', 得到 %s", session.State.AnomalyStatus)
	}

	if session.State.MissionOutcome != "进行中" {
		t.Errorf("期望初始任务结果为'进行中', 得到 %s", session.State.MissionOutcome)
	}
}

func TestGameService_GetSession(t *testing.T) {
	service := NewGameService()

	// 创建会话
	created, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 获取会话
	retrieved, err := service.GetSession(created.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("期望ID为 %s, 得到 %s", created.ID, retrieved.ID)
	}

	// 测试获取不存在的会话
	_, err = service.GetSession("不存在的ID")
	if err == nil {
		t.Error("期望获取不存在的会话导致错误")
	}
}

func TestGameService_SaveSession(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 修改会话
	session.State.ChaosPool = 5
	session.State.LooseEnds = 3

	// 保存会话
	err = service.SaveSession(session)
	if err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}

	// 验证保存
	retrieved, err := service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if retrieved.State.ChaosPool != 5 {
		t.Errorf("期望混沌池为5, 得到 %d", retrieved.State.ChaosPool)
	}

	if retrieved.State.LooseEnds != 3 {
		t.Errorf("期望散逸端为3, 得到 %d", retrieved.State.LooseEnds)
	}

	// 测试保存不存在的会话
	nonExistent := &domain.GameSession{
		ID: "不存在的ID",
	}
	err = service.SaveSession(nonExistent)
	if err == nil {
		t.Error("期望保存不存在的会话导致错误")
	}
}

func TestGameService_DeleteSession(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 删除会话
	err = service.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("删除会话失败: %v", err)
	}

	// 验证删除
	_, err = service.GetSession(session.ID)
	if err == nil {
		t.Error("期望获取已删除的会话导致错误")
	}

	// 测试删除不存在的会话
	err = service.DeleteSession("不存在的ID")
	if err == nil {
		t.Error("期望删除不存在的会话导致错误")
	}
}

func TestGameService_ListSessions(t *testing.T) {
	service := NewGameService()

	// 创建多个会话
	for i := 0; i < 3; i++ {
		_, err := service.CreateSession("agent-1", "scenario-1")
		if err != nil {
			t.Fatalf("创建会话失败: %v", err)
		}
	}

	// 列出所有会话
	sessions, err := service.ListSessions()
	if err != nil {
		t.Fatalf("列出会话失败: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("期望3个会话, 得到 %d", len(sessions))
	}
}

func TestGameService_TransitionPhase(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 测试有效的阶段转换: 晨会 -> 调查
	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	// 验证转换
	updated, err := service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if updated.Phase != domain.PhaseInvestigation {
		t.Errorf("期望阶段为 %s, 得到 %s", domain.PhaseInvestigation, updated.Phase)
	}

	// 测试有效的阶段转换: 调查 -> 遭遇
	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	updated, err = service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if updated.Phase != domain.PhaseEncounter {
		t.Errorf("期望阶段为 %s, 得到 %s", domain.PhaseEncounter, updated.Phase)
	}

	// 测试有效的阶段转换: 遭遇 -> 余波
	err = service.TransitionPhase(session.ID, domain.PhaseAftermath)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	updated, err = service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if updated.Phase != domain.PhaseAftermath {
		t.Errorf("期望阶段为 %s, 得到 %s", domain.PhaseAftermath, updated.Phase)
	}

	// 测试无效的阶段转换: 余波 -> 调查（跳过晨会）
	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err == nil {
		t.Error("期望无效的阶段转换导致错误")
	}
}

func TestGameService_StartMorningPhase(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 开始晨会阶段
	result, err := service.StartMorningPhase(session.ID)
	if err != nil {
		t.Fatalf("开始晨会阶段失败: %v", err)
	}

	// 验证结果
	if result.SessionID != session.ID {
		t.Errorf("期望SessionID为 %s, 得到 %s", session.ID, result.SessionID)
	}

	if result.Briefing == nil {
		t.Fatal("期望简报不为nil")
	}

	if len(result.Goals) == 0 {
		t.Error("期望至少有一个可选目标")
	}

	if result.Description == "" {
		t.Error("期望描述不为空")
	}

	// 测试在错误阶段调用
	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	_, err = service.StartMorningPhase(session.ID)
	if err == nil {
		t.Error("期望在非晨会阶段调用导致错误")
	}
}

func TestGameService_StartInvestigationPhase(t *testing.T) {
	service := NewGameService()

	// 创建会话并转换到调查阶段
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	// 开始调查阶段
	result, err := service.StartInvestigationPhase(session.ID)
	if err != nil {
		t.Fatalf("开始调查阶段失败: %v", err)
	}

	// 验证结果
	if result.SessionID != session.ID {
		t.Errorf("期望SessionID为 %s, 得到 %s", session.ID, result.SessionID)
	}

	if result.Description == "" {
		t.Error("期望描述不为空")
	}

	// 测试在错误阶段调用
	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	_, err = service.StartInvestigationPhase(session.ID)
	if err == nil {
		t.Error("期望在非调查阶段调用导致错误")
	}
}

func TestGameService_StartEncounterPhase(t *testing.T) {
	service := NewGameService()

	// 创建会话并转换到遭遇阶段
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	// 开始遭遇阶段
	result, err := service.StartEncounterPhase(session.ID)
	if err != nil {
		t.Fatalf("开始遭遇阶段失败: %v", err)
	}

	// 验证结果
	if result.SessionID != session.ID {
		t.Errorf("期望SessionID为 %s, 得到 %s", session.ID, result.SessionID)
	}

	if result.AnomalyName == "" {
		t.Error("期望异常体名称不为空")
	}

	if result.Description == "" {
		t.Error("期望描述不为空")
	}

	// 测试在错误阶段调用
	err = service.TransitionPhase(session.ID, domain.PhaseAftermath)
	if err != nil {
		t.Fatalf("阶段转换失败: %v", err)
	}

	_, err = service.StartEncounterPhase(session.ID)
	if err == nil {
		t.Error("期望在非遭遇阶段调用导致错误")
	}
}

func TestGameService_UpdateState(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 更新状态
	err = service.UpdateState(session.ID, func(state *domain.GameState) error {
		state.ChaosPool = 10
		state.LooseEnds = 5
		state.DomainUnlocked = true
		state.CollectedClues = append(state.CollectedClues, "线索1", "线索2")
		return nil
	})
	if err != nil {
		t.Fatalf("更新状态失败: %v", err)
	}

	// 验证更新
	state, err := service.GetState(session.ID)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if state.ChaosPool != 10 {
		t.Errorf("期望混沌池为10, 得到 %d", state.ChaosPool)
	}

	if state.LooseEnds != 5 {
		t.Errorf("期望散逸端为5, 得到 %d", state.LooseEnds)
	}

	if !state.DomainUnlocked {
		t.Error("期望领域已解锁")
	}

	if len(state.CollectedClues) != 2 {
		t.Errorf("期望2个线索, 得到 %d", len(state.CollectedClues))
	}

	// 测试更新函数返回错误
	testErr := domain.NewGameError(domain.ErrInvalidState, "测试错误")
	err = service.UpdateState(session.ID, func(state *domain.GameState) error {
		return testErr
	})
	if err == nil {
		t.Error("期望更新函数错误被传播")
	}
}

func TestGameService_GetState(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 获取状态
	state, err := service.GetState(session.ID)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if state == nil {
		t.Fatal("期望状态不为nil")
	}

	// 测试获取不存在会话的状态
	_, err = service.GetState("不存在的ID")
	if err == nil {
		t.Error("期望获取不存在会话的状态导致错误")
	}
}

func TestGameService_ConcurrentAccess(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 并发更新状态
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			err := service.UpdateState(session.ID, func(state *domain.GameState) error {
				state.ChaosPool++
				return nil
			})
			if err != nil {
				t.Errorf("并发更新失败: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	state, err := service.GetState(session.ID)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if state.ChaosPool != concurrency {
		t.Errorf("期望混沌池为 %d, 得到 %d", concurrency, state.ChaosPool)
	}
}

func TestGameService_ConcurrentSessionCreation(t *testing.T) {
	service := NewGameService()

	// 并发创建会话
	var wg sync.WaitGroup
	concurrency := 10
	sessionIDs := make(chan string, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			session, err := service.CreateSession("agent-1", "scenario-1")
			if err != nil {
				t.Errorf("并发创建会话失败: %v", err)
				return
			}
			sessionIDs <- session.ID
		}(i)
	}

	wg.Wait()
	close(sessionIDs)

	// 验证所有会话都被创建
	sessions, err := service.ListSessions()
	if err != nil {
		t.Fatalf("列出会话失败: %v", err)
	}

	if len(sessions) != concurrency {
		t.Errorf("期望 %d 个会话, 得到 %d", concurrency, len(sessions))
	}

	// 验证所有ID都是唯一的
	idMap := make(map[string]bool)
	for id := range sessionIDs {
		if idMap[id] {
			t.Errorf("发现重复的会话ID: %s", id)
		}
		idMap[id] = true
	}
}

func TestGameService_PhaseTransitionValidation(t *testing.T) {
	tests := []struct {
		name      string
		from      domain.GamePhase
		to        domain.GamePhase
		wantError bool
	}{
		{"晨会到调查", domain.PhaseMorning, domain.PhaseInvestigation, false},
		{"调查到遭遇", domain.PhaseInvestigation, domain.PhaseEncounter, false},
		{"遭遇到余波", domain.PhaseEncounter, domain.PhaseAftermath, false},
		{"余波到晨会", domain.PhaseAftermath, domain.PhaseMorning, false},
		{"晨会到遭遇（跳过调查）", domain.PhaseMorning, domain.PhaseEncounter, true},
		{"调查到余波（跳过遭遇）", domain.PhaseInvestigation, domain.PhaseAftermath, true},
		{"遭遇到晨会（跳过余波）", domain.PhaseEncounter, domain.PhaseMorning, true},
		{"调查到晨会（倒退）", domain.PhaseInvestigation, domain.PhaseMorning, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameService()

			// 创建会话
			session, err := service.CreateSession("agent-1", "scenario-1")
			if err != nil {
				t.Fatalf("创建会话失败: %v", err)
			}

			// 设置初始阶段
			session.Phase = tt.from
			err = service.SaveSession(session)
			if err != nil {
				t.Fatalf("保存会话失败: %v", err)
			}

			// 尝试转换
			err = service.TransitionPhase(session.ID, tt.to)

			if tt.wantError && err == nil {
				t.Error("期望阶段转换导致错误，但没有错误")
			}

			if !tt.wantError && err != nil {
				t.Errorf("期望阶段转换成功，但得到错误: %v", err)
			}
		})
	}
}

// TestGameFlow_CompleteSequence 测试完整的游戏流程序列
func TestGameFlow_CompleteSequence(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 验证初始阶段为晨会
	if session.Phase != domain.PhaseMorning {
		t.Errorf("期望初始阶段为晨会，得到 %s", session.Phase)
	}

	// 1. 晨会阶段
	morningResult, err := service.StartMorningPhase(session.ID)
	if err != nil {
		t.Fatalf("开始晨会阶段失败: %v", err)
	}

	if morningResult.Briefing == nil {
		t.Error("期望晨会阶段返回简报")
	}

	if len(morningResult.Goals) == 0 {
		t.Error("期望晨会阶段返回可选目标")
	}

	// 转换到调查阶段
	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("转换到调查阶段失败: %v", err)
	}

	// 2. 调查阶段
	investigationResult, err := service.StartInvestigationPhase(session.ID)
	if err != nil {
		t.Fatalf("开始调查阶段失败: %v", err)
	}

	if investigationResult.SessionID != session.ID {
		t.Errorf("期望SessionID为 %s，得到 %s", session.ID, investigationResult.SessionID)
	}

	// 转换到遭遇阶段
	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("转换到遭遇阶段失败: %v", err)
	}

	// 3. 遭遇阶段
	encounterResult, err := service.StartEncounterPhase(session.ID)
	if err != nil {
		t.Fatalf("开始遭遇阶段失败: %v", err)
	}

	if encounterResult.AnomalyName == "" {
		t.Error("期望遭遇阶段返回异常体名称")
	}

	// 转换到余波阶段
	err = service.TransitionPhase(session.ID, domain.PhaseAftermath)
	if err != nil {
		t.Fatalf("转换到余波阶段失败: %v", err)
	}

	// 验证最终阶段
	finalSession, err := service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取最终会话失败: %v", err)
	}

	if finalSession.Phase != domain.PhaseAftermath {
		t.Errorf("期望最终阶段为余波，得到 %s", finalSession.Phase)
	}
}

// TestGameFlow_MorningPhaseDetails 测试晨会阶段的详细功能
func TestGameFlow_MorningPhaseDetails(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 开始晨会阶段
	result, err := service.StartMorningPhase(session.ID)
	if err != nil {
		t.Fatalf("开始晨会阶段失败: %v", err)
	}

	// 验证简报内容
	if result.Briefing == nil {
		t.Fatal("期望简报不为nil")
	}

	if result.Briefing.Summary == "" {
		t.Error("期望简报包含摘要")
	}

	if len(result.Briefing.Objectives) == 0 {
		t.Error("期望简报包含目标")
	}

	if len(result.Briefing.Warnings) == 0 {
		t.Error("期望简报包含警告")
	}

	// 验证可选目标
	if len(result.Goals) == 0 {
		t.Error("期望至少有一个可选目标")
	}

	for _, goal := range result.Goals {
		if goal.ID == "" {
			t.Error("期望可选目标有ID")
		}
		if goal.Description == "" {
			t.Error("期望可选目标有描述")
		}
		if goal.Reward <= 0 {
			t.Error("期望可选目标有正数奖励")
		}
	}

	// 验证描述
	if result.Description == "" {
		t.Error("期望晨会阶段有描述")
	}
}

// TestGameFlow_InvestigationPhaseTracking 测试调查阶段的状态追踪
func TestGameFlow_InvestigationPhaseTracking(t *testing.T) {
	service := NewGameService()

	// 创建会话并转换到调查阶段
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("转换到调查阶段失败: %v", err)
	}

	// 模拟调查过程中的状态变化
	err = service.UpdateState(session.ID, func(state *domain.GameState) error {
		// 添加线索
		state.CollectedClues = append(state.CollectedClues, "线索1", "线索2", "线索3")

		// 解锁地点
		state.UnlockedLocations = append(state.UnlockedLocations, "地点A", "地点B")

		// 增加散逸端
		state.LooseEnds = 5

		// 设置当前场景
		state.CurrentSceneID = "scene-1"

		// 标记访问过的场景
		state.VisitedScenes["scene-1"] = true
		state.VisitedScenes["scene-2"] = true

		return nil
	})
	if err != nil {
		t.Fatalf("更新状态失败: %v", err)
	}

	// 验证状态追踪
	state, err := service.GetState(session.ID)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if len(state.CollectedClues) != 3 {
		t.Errorf("期望收集3个线索，得到 %d", len(state.CollectedClues))
	}

	if len(state.UnlockedLocations) != 2 {
		t.Errorf("期望解锁2个地点，得到 %d", len(state.UnlockedLocations))
	}

	if state.LooseEnds != 5 {
		t.Errorf("期望5个散逸端，得到 %d", state.LooseEnds)
	}

	if state.CurrentSceneID != "scene-1" {
		t.Errorf("期望当前场景为scene-1，得到 %s", state.CurrentSceneID)
	}

	if len(state.VisitedScenes) != 2 {
		t.Errorf("期望访问2个场景，得到 %d", len(state.VisitedScenes))
	}

	// 开始调查阶段并验证结果
	result, err := service.StartInvestigationPhase(session.ID)
	if err != nil {
		t.Fatalf("开始调查阶段失败: %v", err)
	}

	if result.CurrentSceneID != "scene-1" {
		t.Errorf("期望当前场景为scene-1，得到 %s", result.CurrentSceneID)
	}

	if len(result.AvailableScenes) != 2 {
		t.Errorf("期望2个可用场景，得到 %d", len(result.AvailableScenes))
	}
}

// TestGameFlow_EncounterPhaseActivation 测试遭遇阶段的激活
func TestGameFlow_EncounterPhaseActivation(t *testing.T) {
	service := NewGameService()

	// 创建会话并转换到遭遇阶段
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("转换到调查阶段失败: %v", err)
	}

	// 模拟进入异常体领域
	err = service.UpdateState(session.ID, func(state *domain.GameState) error {
		state.DomainUnlocked = true
		state.CurrentSceneID = "domain-scene"
		state.ChaosPool = 10 // 初始化混沌池
		return nil
	})
	if err != nil {
		t.Fatalf("更新状态失败: %v", err)
	}

	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("转换到遭遇阶段失败: %v", err)
	}

	// 开始遭遇阶段
	result, err := service.StartEncounterPhase(session.ID)
	if err != nil {
		t.Fatalf("开始遭遇阶段失败: %v", err)
	}

	// 验证遭遇阶段结果
	if result.AnomalyName == "" {
		t.Error("期望遭遇阶段返回异常体名称")
	}

	if result.Description == "" {
		t.Error("期望遭遇阶段返回描述")
	}

	// 验证混沌池已初始化
	state, err := service.GetState(session.ID)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if state.ChaosPool != 10 {
		t.Errorf("期望混沌池为10，得到 %d", state.ChaosPool)
	}

	if !state.DomainUnlocked {
		t.Error("期望领域已解锁")
	}
}

// TestGameFlow_PhaseTransitionSequence 测试阶段转换序列的正确性
func TestGameFlow_PhaseTransitionSequence(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 定义正确的阶段序列
	phaseSequence := []domain.GamePhase{
		domain.PhaseMorning,
		domain.PhaseInvestigation,
		domain.PhaseEncounter,
		domain.PhaseAftermath,
	}

	// 验证初始阶段
	currentSession, err := service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if currentSession.Phase != phaseSequence[0] {
		t.Errorf("期望初始阶段为 %s，得到 %s", phaseSequence[0], currentSession.Phase)
	}

	// 按序列转换阶段
	for i := 1; i < len(phaseSequence); i++ {
		err = service.TransitionPhase(session.ID, phaseSequence[i])
		if err != nil {
			t.Fatalf("转换到阶段 %s 失败: %v", phaseSequence[i], err)
		}

		// 验证转换成功
		currentSession, err = service.GetSession(session.ID)
		if err != nil {
			t.Fatalf("获取会话失败: %v", err)
		}

		if currentSession.Phase != phaseSequence[i] {
			t.Errorf("期望阶段为 %s，得到 %s", phaseSequence[i], currentSession.Phase)
		}
	}

	// 验证可以从余波回到晨会（开始新任务）
	err = service.TransitionPhase(session.ID, domain.PhaseMorning)
	if err != nil {
		t.Fatalf("从余波转换到晨会失败: %v", err)
	}

	currentSession, err = service.GetSession(session.ID)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	if currentSession.Phase != domain.PhaseMorning {
		t.Errorf("期望阶段为晨会，得到 %s", currentSession.Phase)
	}
}

// TestGameFlow_InvalidPhaseOperations 测试在错误阶段执行操作
func TestGameFlow_InvalidPhaseOperations(t *testing.T) {
	service := NewGameService()

	// 创建会话
	session, err := service.CreateSession("agent-1", "scenario-1")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	// 测试在晨会阶段调用调查阶段方法
	_, err = service.StartInvestigationPhase(session.ID)
	if err == nil {
		t.Error("期望在晨会阶段调用调查阶段方法导致错误")
	}

	// 测试在晨会阶段调用遭遇阶段方法
	_, err = service.StartEncounterPhase(session.ID)
	if err == nil {
		t.Error("期望在晨会阶段调用遭遇阶段方法导致错误")
	}

	// 转换到调查阶段
	err = service.TransitionPhase(session.ID, domain.PhaseInvestigation)
	if err != nil {
		t.Fatalf("转换到调查阶段失败: %v", err)
	}

	// 测试在调查阶段调用晨会阶段方法
	_, err = service.StartMorningPhase(session.ID)
	if err == nil {
		t.Error("期望在调查阶段调用晨会阶段方法导致错误")
	}

	// 测试在调查阶段调用遭遇阶段方法
	_, err = service.StartEncounterPhase(session.ID)
	if err == nil {
		t.Error("期望在调查阶段调用遭遇阶段方法导致错误")
	}

	// 转换到遭遇阶段
	err = service.TransitionPhase(session.ID, domain.PhaseEncounter)
	if err != nil {
		t.Fatalf("转换到遭遇阶段失败: %v", err)
	}

	// 测试在遭遇阶段调用晨会阶段方法
	_, err = service.StartMorningPhase(session.ID)
	if err == nil {
		t.Error("期望在遭遇阶段调用晨会阶段方法导致错误")
	}

	// 测试在遭遇阶段调用调查阶段方法
	_, err = service.StartInvestigationPhase(session.ID)
	if err == nil {
		t.Error("期望在遭遇阶段调用调查阶段方法导致错误")
	}
}
