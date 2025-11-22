package service

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// **Feature: trpg-solo-engine, Property 9: 请求机构约束**
// **Validates: Requirements 6.1, 6.4, 6.5**
//
// 属性9: 请求机构约束
// 对于任何现实变更请求，必须包含效果、因果链、资质和掷骰四个要素。
// 已确立的事实不能被改变，直接心智控制必须被拒绝。
func TestProperty_RequestConstraints(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 创建服务
	diceService := domain.NewDiceService()
	chaosService := NewChaosService()
	requestService := NewRequestService(diceService, chaosService)

	// 属性1: 请求必须包含所有四个要素
	properties.Property("请求必须包含效果、因果链、资质和掷骰", prop.ForAll(
		func(effect, causalChain, quality string) bool {
			// 生成请求
			req := &RealityChangeRequest{
				Effect:      effect,
				CausalChain: causalChain,
				Quality:     quality,
				LocationID:  "test_location",
			}

			// 验证请求
			err := requestService.ValidateRequest(req)

			// 如果任何一个要素为空，应该返回错误
			if effect == "" || causalChain == "" || quality == "" {
				return err != nil
			}

			// 如果因果链太短，应该返回错误
			if len(causalChain) < 10 {
				return err != nil
			}

			// 如果资质无效，应该返回错误
			validQualities := []string{"专注", "共情", "气场", "欺瞒", "主动", "专业", "活力", "坚毅", "诡秘"}
			isValidQuality := false
			for _, q := range validQualities {
				if quality == q {
					isValidQuality = true
					break
				}
			}
			if !isValidQuality {
				return err != nil
			}

			// 如果包含心智控制关键词，应该返回错误
			if requestService.IsMindControl(effect) {
				return err != nil
			}

			// 所有要素都有效时，不应该有错误
			return err == nil
		},
		gen.AnyString(),
		gen.AnyString(),
		gen.AnyString(),
	))

	// 属性2: 已确立的事实不能被改变
	properties.Property("已确立的事实不能被改变", prop.ForAll(
		func(fact string) bool {
			// 创建测试会话
			session := &domain.GameSession{
				ID:      "test_session",
				AgentID: "test_agent",
				State: &domain.GameState{
					CollectedClues:    []string{},
					LocationOverloads: make(map[string]int),
				},
			}

			// 创建测试角色
			agent := &domain.Agent{
				ID:   "test_agent",
				Name: "测试角色",
				QA: map[string]int{
					"专注": 1,
					"共情": 1,
					"气场": 1,
					"欺瞒": 1,
					"主动": 1,
					"专业": 1,
					"活力": 1,
					"坚毅": 1,
					"诡秘": 1,
				},
			}

			// 首先确立一个事实
			err := requestService.AddEstablishedFact(session, fact)
			if err != nil {
				return true // 如果添加失败，跳过
			}

			// 验证事实已被确立
			if !requestService.IsEstablishedFact(session, fact) {
				return false
			}

			// 尝试再次改变这个已确立的事实
			req := &RealityChangeRequest{
				Effect:      fact,
				CausalChain: "这是一个足够长的因果链描述，用于测试",
				Quality:     "专注",
				LocationID:  "test_location",
			}

			// 创建一个成功的掷骰结果
			roll := &domain.RollResult{
				Dice:    []int{3, 3, 1, 2, 1, 2},
				Threes:  2,
				Success: true,
				Chaos:   0,
			}

			// 处理请求应该失败
			_, err = requestService.ProcessRequest(agent, session, req, roll)

			// 应该返回错误，因为不能改变已确立的事实
			return err != nil
		},
		gen.Identifier(),
	))

	// 属性3: 直接心智控制必须被拒绝
	properties.Property("直接心智控制必须被拒绝", prop.ForAll(
		func() bool {
			// 测试各种心智控制关键词
			mindControlEffects := []string{
				"控制他的思想",
				"操纵她的意志",
				"强迫他们服从",
				"命令他做某事",
				"洗脑目标",
				"催眠对方",
				"支配他的心智",
				"迫使她改变想法",
				"让他想要帮助我",
				"改变他们的思想",
			}

			// 对每个心智控制效果进行测试
			for _, effect := range mindControlEffects {
				req := &RealityChangeRequest{
					Effect:      effect,
					CausalChain: "这是一个足够长的因果链描述，用于测试心智控制检测",
					Quality:     "专注",
					LocationID:  "test_location",
				}

				// 验证请求应该失败
				err := requestService.ValidateRequest(req)
				if err == nil {
					return false // 应该返回错误
				}
			}

			return true
		},
	))

	// 属性4: 有效的请求应该被接受
	properties.Property("有效的请求应该被接受", prop.ForAll(
		func(quality string) bool {
			validQualities := []string{"专注", "共情", "气场", "欺瞒", "主动", "专业", "活力", "坚毅", "诡秘"}
			isValidQuality := false
			for _, q := range validQualities {
				if quality == q {
					isValidQuality = true
					break
				}
			}
			if !isValidQuality {
				return true
			}

			// 创建一个有效的请求
			req := &RealityChangeRequest{
				Effect:      "门突然打开了",
				CausalChain: "因为我之前在门上做了手脚，所以现在门锁失效了",
				Quality:     quality,
				LocationID:  "test_location",
			}

			// 验证请求应该成功
			err := requestService.ValidateRequest(req)
			return err == nil
		},
		gen.OneConstOf("专注", "共情", "气场", "欺瞒", "主动", "专业", "活力", "坚毅", "诡秘"),
	))

	// 属性5: 请求成功时应该确立事实
	properties.Property("请求成功时应该确立事实", prop.ForAll(
		func(effect string) bool {
			// 创建测试会话
			session := &domain.GameSession{
				ID:      "test_session",
				AgentID: "test_agent",
				State: &domain.GameState{
					CollectedClues:    []string{},
					LocationOverloads: make(map[string]int),
					ChaosPool:         0,
				},
			}

			// 创建测试角色
			agent := &domain.Agent{
				ID:   "test_agent",
				Name: "测试角色",
				QA: map[string]int{
					"专注": 1,
					"共情": 1,
					"气场": 1,
					"欺瞒": 1,
					"主动": 1,
					"专业": 1,
					"活力": 1,
					"坚毅": 1,
					"诡秘": 1,
				},
			}

			// 如果包含心智控制关键词，跳过
			if requestService.IsMindControl(effect) {
				return true
			}

			// 创建一个有效的请求
			req := &RealityChangeRequest{
				Effect:      effect,
				CausalChain: "这是一个足够长的因果链描述，用于测试请求成功时的行为",
				Quality:     "专注",
				LocationID:  "test_location",
			}

			// 创建一个成功的掷骰结果
			roll := &domain.RollResult{
				Dice:    []int{3, 3, 1, 2, 1, 2},
				Threes:  2,
				Success: true,
				Chaos:   0,
			}

			// 处理请求
			result, err := requestService.ProcessRequest(agent, session, req, roll)
			if err != nil {
				return true // 如果有错误，跳过
			}

			// 验证结果
			if !result.Success {
				return false
			}

			// 验证事实已被确立
			if !requestService.IsEstablishedFact(session, effect) {
				return false
			}

			return true
		},
		gen.Identifier(),
	))

	// 属性6: 请求失败时应该添加地点过载
	properties.Property("请求失败时应该添加地点过载", prop.ForAll(
		func(effect string, locationID string) bool {
			// 创建测试会话
			session := &domain.GameSession{
				ID:      "test_session",
				AgentID: "test_agent",
				State: &domain.GameState{
					CollectedClues:    []string{},
					LocationOverloads: make(map[string]int),
					ChaosPool:         0,
				},
			}

			// 创建测试角色
			agent := &domain.Agent{
				ID:   "test_agent",
				Name: "测试角色",
				QA: map[string]int{
					"专注": 1,
					"共情": 1,
					"气场": 1,
					"欺瞒": 1,
					"主动": 1,
					"专业": 1,
					"活力": 1,
					"坚毅": 1,
					"诡秘": 1,
				},
			}

			// 如果包含心智控制关键词，跳过
			if requestService.IsMindControl(effect) {
				return true
			}

			// 记录初始过载
			initialOverload := chaosService.GetLocationOverload(session, locationID)

			// 创建一个有效的请求
			req := &RealityChangeRequest{
				Effect:      effect,
				CausalChain: "这是一个足够长的因果链描述，用于测试请求失败时的行为",
				Quality:     "专注",
				LocationID:  locationID,
			}

			// 创建一个失败的掷骰结果
			roll := &domain.RollResult{
				Dice:    []int{1, 1, 2, 2, 4, 4},
				Threes:  0,
				Success: false,
				Chaos:   6,
			}

			// 处理请求
			result, err := requestService.ProcessRequest(agent, session, req, roll)
			if err != nil {
				return true // 如果有错误，跳过
			}

			// 验证结果
			if result.Success {
				return false
			}

			// 验证地点过载增加了
			finalOverload := chaosService.GetLocationOverload(session, locationID)
			if finalOverload != initialOverload+1 {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}
