package service

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// TestProperty_SaveLoadRoundTrip 属性15: 存档round-trip
// Feature: trpg-solo-engine, Property 15: 存档round-trip
// 验证需求: 11.1, 11.2
//
// 对于任何游戏状态，保存后立即加载应该恢复完全相同的状态（序列化后反序列化应该是恒等操作）。
func TestProperty_SaveLoadRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	properties.Property("序列化后反序列化应该恢复相同的游戏会话", prop.ForAll(
		func(phase domain.GamePhase, chaosPool int, looseEnds int, domainUnlocked bool) bool {
			// 创建测试角色
			req := &CreateAgentRequest{
				Name:        "测试特工",
				Pronouns:    "他/他的",
				AnomalyType: domain.AnomalyWhisper,
				RealityType: domain.RealityCaretaker,
				CareerType:  domain.CareerPublicRelations,
				Relationships: []*domain.Relationship{
					{Name: "李娜", Connection: 6},
					{Name: "王强", Connection: 3},
					{Name: "陈医生", Connection: 3},
				},
			}
			agent, err := agentService.CreateAgent(req)
			if err != nil {
				return false
			}

			// 创建游戏会话
			session, err := gameService.CreateSession(agent.ID, "test-scenario")
			if err != nil {
				return false
			}

			// 设置游戏状态
			session.Phase = phase
			session.State.ChaosPool = chaosPool
			session.State.LooseEnds = looseEnds
			session.State.DomainUnlocked = domainUnlocked

			// 序列化
			data, err := saveService.SerializeSession(session)
			if err != nil {
				return false
			}

			// 反序列化
			loaded, err := saveService.DeserializeSession(data)
			if err != nil {
				return false
			}

			// 验证关键字段相等
			if loaded.AgentID != session.AgentID {
				return false
			}
			if loaded.ScenarioID != session.ScenarioID {
				return false
			}
			if loaded.Phase != session.Phase {
				return false
			}
			if loaded.State.ChaosPool != session.State.ChaosPool {
				return false
			}
			if loaded.State.LooseEnds != session.State.LooseEnds {
				return false
			}
			if loaded.State.DomainUnlocked != session.State.DomainUnlocked {
				return false
			}

			return true
		},
		genGamePhase(),
		gen.IntRange(0, 20),
		gen.IntRange(0, 10),
		gen.Bool(),
	))

	properties.Property("存档创建和加载应该保持游戏状态", prop.ForAll(
		func(saveName string, clues []string, locations []string) bool {
			// 创建测试角色
			req := &CreateAgentRequest{
				Name:        "测试特工",
				Pronouns:    "他/他的",
				AnomalyType: domain.AnomalyWhisper,
				RealityType: domain.RealityCaretaker,
				CareerType:  domain.CareerPublicRelations,
				Relationships: []*domain.Relationship{
					{Name: "李娜", Connection: 6},
					{Name: "王强", Connection: 3},
					{Name: "陈医生", Connection: 3},
				},
			}
			agent, err := agentService.CreateAgent(req)
			if err != nil {
				return false
			}

			// 创建游戏会话
			session, err := gameService.CreateSession(agent.ID, "test-scenario")
			if err != nil {
				return false
			}

			// 设置游戏状态
			session.State.CollectedClues = clues
			session.State.UnlockedLocations = locations
			err = gameService.SaveSession(session)
			if err != nil {
				return false
			}

			// 创建存档
			snapshot, err := saveService.CreateSave(session.ID, saveName)
			if err != nil {
				return false
			}

			// 加载存档
			loadedSession, err := saveService.LoadSave(snapshot.ID)
			if err != nil {
				return false
			}

			// 验证状态
			if len(loadedSession.State.CollectedClues) != len(clues) {
				return false
			}
			for i, clue := range clues {
				if loadedSession.State.CollectedClues[i] != clue {
					return false
				}
			}

			if len(loadedSession.State.UnlockedLocations) != len(locations) {
				return false
			}
			for i, loc := range locations {
				if loadedSession.State.UnlockedLocations[i] != loc {
					return false
				}
			}

			return true
		},
		gen.AlphaString(),
		gen.SliceOf(gen.AlphaString()),
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t)
}

// genGamePhase 生成游戏阶段
func genGamePhase() gopter.Gen {
	return gen.OneConstOf(
		domain.PhaseMorning,
		domain.PhaseInvestigation,
		domain.PhaseEncounter,
		domain.PhaseAftermath,
	)
}
