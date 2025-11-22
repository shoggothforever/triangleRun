package service

import (
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// **Feature: trpg-solo-engine, Property 7: 混沌守恒**
// **Validates: Requirements 5.1, 5.2, 5.3, 5.4**
//
// 属性7: 混沌守恒
// 对于任何游戏会话，混沌池的变化应该遵循：
// - 失败掷骰增加混沌（每颗非"3"骰子+1）
// - 异常体使用效应减少混沌
// - 任务开始时等于累积散逸端
// - 任务结束时归零
func TestProperty_ChaosConservation(t *testing.T) {
	chaosService := NewChaosService()

	// 测试1: 任务开始时混沌池等于散逸端数量
	t.Run("InitializationEqualsLooseEnds", func(t *testing.T) {
		f := func(looseEnds uint8) bool {
			// 限制范围避免过大的值
			looseEndsInt := int(looseEnds % 50)

			session := createTestSession()
			err := chaosService.InitializeChaosPool(session, looseEndsInt)
			if err != nil {
				return false
			}

			// 验证混沌池等于散逸端数量
			return chaosService.GetChaosPool(session) == looseEndsInt &&
				session.State.LooseEnds == looseEndsInt
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 失败掷骰增加混沌（每颗非"3"骰子+1）
	t.Run("FailedRollAddsChaos", func(t *testing.T) {
		f := func(initialChaos uint8, chaosFromRoll uint8) bool {
			// 限制范围
			initial := int(initialChaos % 50)
			chaos := int(chaosFromRoll % 20)

			session := createTestSession()
			session.State.ChaosPool = initial

			// 创建失败的掷骰结果
			roll := &domain.RollResult{
				Dice:      make([]int, 6),
				Threes:    0,
				Success:   false,
				Chaos:     chaos,
				Overload:  0,
				TripleAsc: false,
			}

			err := chaosService.AddChaosFromRoll(session, roll)
			if err != nil {
				return false
			}

			// 验证混沌池增加了正确的数量
			return chaosService.GetChaosPool(session) == initial+chaos
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 成功掷骰不增加混沌
	t.Run("SuccessfulRollDoesNotAddChaos", func(t *testing.T) {
		f := func(initialChaos uint8) bool {
			initial := int(initialChaos % 50)

			session := createTestSession()
			session.State.ChaosPool = initial

			// 创建成功的掷骰结果
			roll := &domain.RollResult{
				Dice:      []int{3, 3, 1, 2, 4, 1},
				Threes:    2,
				Success:   true,
				Chaos:     0,
				Overload:  0,
				TripleAsc: false,
			}

			err := chaosService.AddChaosFromRoll(session, roll)
			if err != nil {
				return false
			}

			// 验证混沌池没有变化
			return chaosService.GetChaosPool(session) == initial
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 异常体使用效应减少混沌
	t.Run("SpendingChaosReducesPool", func(t *testing.T) {
		f := func(initialChaos uint8, spendAmount uint8) bool {
			initial := int(initialChaos%50) + 10 // 确保至少有10点混沌
			spend := int(spendAmount % 10)       // 花费不超过10点

			session := createTestSession()
			session.State.ChaosPool = initial

			err := chaosService.SpendChaos(session, spend)
			if err != nil {
				return false
			}

			// 验证混沌池减少了正确的数量
			return chaosService.GetChaosPool(session) == initial-spend
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: 不能花费超过混沌池的混沌
	t.Run("CannotSpendMoreThanAvailable", func(t *testing.T) {
		f := func(initialChaos uint8, extraSpend uint8) bool {
			initial := int(initialChaos % 20)
			spend := initial + int(extraSpend%10) + 1 // 总是超过可用量

			session := createTestSession()
			session.State.ChaosPool = initial

			err := chaosService.SpendChaos(session, spend)

			// 应该返回错误
			if err == nil {
				return false
			}

			// 混沌池应该没有变化
			return chaosService.GetChaosPool(session) == initial
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试6: 任务结束时混沌池归零
	t.Run("ClearPoolAtMissionEnd", func(t *testing.T) {
		f := func(initialChaos uint8) bool {
			initial := int(initialChaos % 100)

			session := createTestSession()
			session.State.ChaosPool = initial

			err := chaosService.ClearChaosPool(session)
			if err != nil {
				return false
			}

			// 验证混沌池归零
			return chaosService.GetChaosPool(session) == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试7: 混沌守恒 - 添加和消耗的总和应该等于变化量
	t.Run("ChaosConservationAddAndSpend", func(t *testing.T) {
		f := func(looseEnds uint8, adds []uint8, spends []uint8) bool {
			// 限制数组大小
			if len(adds) > 10 {
				adds = adds[:10]
			}
			if len(spends) > 10 {
				spends = spends[:10]
			}

			initial := int(looseEnds % 50)
			session := createTestSession()

			// 初始化混沌池
			err := chaosService.InitializeChaosPool(session, initial)
			if err != nil {
				return false
			}

			expectedChaos := initial

			// 添加混沌
			for _, add := range adds {
				amount := int(add % 10)
				err := chaosService.AddChaos(session, amount)
				if err != nil {
					return false
				}
				expectedChaos += amount
			}

			// 消耗混沌（只消耗可用的）
			for _, spend := range spends {
				amount := int(spend % 5)
				if amount <= chaosService.GetChaosPool(session) {
					err := chaosService.SpendChaos(session, amount)
					if err != nil {
						return false
					}
					expectedChaos -= amount
				}
			}

			// 验证最终混沌池等于预期值
			return chaosService.GetChaosPool(session) == expectedChaos
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// **Feature: trpg-solo-engine, Property 8: 请求机构地点过载**
// **Validates: Requirements 5.5**
//
// 属性8: 请求机构地点过载
// 对于任何地点，如果玩家在该地点请求机构失败，
// 该地点的后续请求应该累积1点过载，直到离开该地点
func TestProperty_LocationOverload(t *testing.T) {
	chaosService := NewChaosService()

	// 测试1: 请求失败增加地点过载
	t.Run("FailedRequestAddsOverload", func(t *testing.T) {
		f := func(locationID string, failures uint8) bool {
			// 确保locationID不为空
			if locationID == "" {
				locationID = "test_location"
			}

			failureCount := int(failures%20) + 1 // 1-20次失败

			session := createTestSession()

			// 模拟多次失败
			for i := 0; i < failureCount; i++ {
				err := chaosService.AddLocationOverload(session, locationID)
				if err != nil {
					return false
				}
			}

			// 验证过载数量等于失败次数
			return chaosService.GetLocationOverload(session, locationID) == failureCount
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试2: 不同地点的过载独立追踪
	t.Run("DifferentLocationsIndependentOverload", func(t *testing.T) {
		f := func(failures1 uint8, failures2 uint8) bool {
			loc1 := "location_1"
			loc2 := "location_2"
			count1 := int(failures1 % 10)
			count2 := int(failures2 % 10)

			session := createTestSession()

			// 为第一个地点添加过载
			for i := 0; i < count1; i++ {
				err := chaosService.AddLocationOverload(session, loc1)
				if err != nil {
					return false
				}
			}

			// 为第二个地点添加过载
			for i := 0; i < count2; i++ {
				err := chaosService.AddLocationOverload(session, loc2)
				if err != nil {
					return false
				}
			}

			// 验证两个地点的过载独立
			return chaosService.GetLocationOverload(session, loc1) == count1 &&
				chaosService.GetLocationOverload(session, loc2) == count2
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试3: 离开地点清除过载
	t.Run("LeavingLocationClearsOverload", func(t *testing.T) {
		f := func(locationID string, failures uint8) bool {
			if locationID == "" {
				locationID = "test_location"
			}

			failureCount := int(failures%20) + 1

			session := createTestSession()

			// 添加过载
			for i := 0; i < failureCount; i++ {
				err := chaosService.AddLocationOverload(session, locationID)
				if err != nil {
					return false
				}
			}

			// 验证过载存在
			if chaosService.GetLocationOverload(session, locationID) != failureCount {
				return false
			}

			// 清除过载
			err := chaosService.ClearLocationOverload(session, locationID)
			if err != nil {
				return false
			}

			// 验证过载已清除
			return chaosService.GetLocationOverload(session, locationID) == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试4: 过载单调递增（在同一地点）
	t.Run("OverloadMonotonicallyIncreases", func(t *testing.T) {
		f := func(locationID string, failures []uint8) bool {
			if locationID == "" {
				locationID = "test_location"
			}

			// 限制数组大小
			if len(failures) > 20 {
				failures = failures[:20]
			}

			session := createTestSession()
			previousOverload := 0

			// 逐次添加过载并验证单调递增
			for range failures {
				err := chaosService.AddLocationOverload(session, locationID)
				if err != nil {
					return false
				}

				currentOverload := chaosService.GetLocationOverload(session, locationID)
				if currentOverload != previousOverload+1 {
					return false
				}
				previousOverload = currentOverload
			}

			return true
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// 测试5: 清除不存在的地点过载不会出错
	t.Run("ClearingNonexistentLocationSafe", func(t *testing.T) {
		f := func(locationID string) bool {
			if locationID == "" {
				locationID = "nonexistent_location"
			}

			session := createTestSession()

			// 清除不存在的地点过载
			err := chaosService.ClearLocationOverload(session, locationID)
			if err != nil {
				return false
			}

			// 验证过载为0
			return chaosService.GetLocationOverload(session, locationID) == 0
		}

		if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// 辅助函数：创建测试用的游戏会话
func createTestSession() *domain.GameSession {
	return &domain.GameSession{
		ID:         "test-session",
		AgentID:    "test-agent",
		ScenarioID: "test-scenario",
		Phase:      domain.PhaseInvestigation,
		State: &domain.GameState{
			CurrentSceneID:    "scene-1",
			VisitedScenes:     make(map[string]bool),
			CollectedClues:    []string{},
			UnlockedLocations: []string{},
			DomainUnlocked:    false,
			NPCStates:         make(map[string]*domain.NPCState),
			ChaosPool:         0,
			LooseEnds:         0,
			LocationOverloads: make(map[string]int),
			AnomalyStatus:     "active",
			MissionOutcome:    "",
		},
	}
}

// 单元测试：验证基本功能
func TestChaosService_BasicFunctionality(t *testing.T) {
	chaosService := NewChaosService()

	t.Run("InitializeChaosPool", func(t *testing.T) {
		session := createTestSession()
		err := chaosService.InitializeChaosPool(session, 5)
		assert.NoError(t, err)
		assert.Equal(t, 5, chaosService.GetChaosPool(session))
		assert.Equal(t, 5, session.State.LooseEnds)
	})

	t.Run("AddChaos", func(t *testing.T) {
		session := createTestSession()
		session.State.ChaosPool = 10

		err := chaosService.AddChaos(session, 5)
		assert.NoError(t, err)
		assert.Equal(t, 15, chaosService.GetChaosPool(session))
	})

	t.Run("SpendChaos", func(t *testing.T) {
		session := createTestSession()
		session.State.ChaosPool = 10

		err := chaosService.SpendChaos(session, 3)
		assert.NoError(t, err)
		assert.Equal(t, 7, chaosService.GetChaosPool(session))
	})

	t.Run("SpendChaos_Insufficient", func(t *testing.T) {
		session := createTestSession()
		session.State.ChaosPool = 5

		err := chaosService.SpendChaos(session, 10)
		assert.Error(t, err)
		assert.Equal(t, 5, chaosService.GetChaosPool(session))
	})

	t.Run("ClearChaosPool", func(t *testing.T) {
		session := createTestSession()
		session.State.ChaosPool = 20

		err := chaosService.ClearChaosPool(session)
		assert.NoError(t, err)
		assert.Equal(t, 0, chaosService.GetChaosPool(session))
	})

	t.Run("LocationOverload", func(t *testing.T) {
		session := createTestSession()

		// 添加过载
		err := chaosService.AddLocationOverload(session, "location-1")
		assert.NoError(t, err)
		assert.Equal(t, 1, chaosService.GetLocationOverload(session, "location-1"))

		// 再次添加
		err = chaosService.AddLocationOverload(session, "location-1")
		assert.NoError(t, err)
		assert.Equal(t, 2, chaosService.GetLocationOverload(session, "location-1"))

		// 清除过载
		err = chaosService.ClearLocationOverload(session, "location-1")
		assert.NoError(t, err)
		assert.Equal(t, 0, chaosService.GetLocationOverload(session, "location-1"))
	})
}
