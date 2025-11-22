# 三角机构TRPG单人引擎 - 设计文档

## 概述

本系统是一个基于《三角机构》规则的单人TRPG游戏引擎，使用Golang构建后端，提供RESTful API供前端调用。系统的核心目标是让个人玩家能够独自体验完整的异常体回收任务，通过AI总经理模拟GM角色，提供沉浸式的单人游戏体验。

### 核心特性

1. **完整的规则实现**：6d4骰子系统、资质保证、混沌池、请求机构
2. **ARC角色系统**：9种异常体、9种现实、9种职能
3. **剧本模组系统**：支持加载和执行多个剧本
4. **三阶段游戏流程**：晨会 → 调查 → 遭遇
5. **AI总经理**：智能叙事、NPC扮演、混沌效应决策
6. **状态持久化**：完整的存档和读档功能

## 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         前端层                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 游戏UI   │  │ 角色创建 │  │ 线索日志 │  │ 场景渲染 │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       │             │              │             │          │
│       └─────────────┴──────────────┴─────────────┘          │
│                         │                                    │
│                    REST API                                  │
└────────────────────────┼────────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────────┐
│                    API Gateway                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  认证中间件 │ 日志中间件 │ 错误处理 │ 限流中间件   │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┼────────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────────┐
│                     应用层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ 游戏控制器   │  │ 角色控制器   │  │ 剧本控制器   │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │              │
│  ┌──────┴──────────────────┴──────────────────┴────────┐   │
│  │                   服务层                              │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │   │
│  │  │ 游戏服务   │  │ 骰子服务   │  │ AI服务     │    │   │
│  │  └────────────┘  └────────────┘  └────────────┘    │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │   │
│  │  │ 剧本服务   │  │ 角色服务   │  │ 状态服务   │    │   │
│  │  └────────────┘  └────────────┘  └────────────┘    │   │
│  └───────────────────────────────────────────────────┘   │
└────────────────────────┼────────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────────┐
│                     领域层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ 游戏引擎     │  │ 规则引擎     │  │ 剧本引擎     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ 骰子系统     │  │ 混沌系统     │  │ 事件系统     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┼────────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────────┐
│                     数据层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ PostgreSQL   │  │ Redis缓存    │  │ 文件存储     │     │
│  │ (游戏状态)   │  │ (会话数据)   │  │ (剧本数据)   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### 技术栈

- **后端语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **AI集成**: OpenAI API / 本地LLM
- **配置管理**: Viper
- **日志**: Zap
- **测试**: Testify + Ginkgo

## 组件和接口

### 1. 核心领域模型

#### 1.1 游戏会话 (Game Session)

```go
package domain

type GameSession struct {
    ID            string
    AgentID       string
    ScenarioID    string
    Phase         GamePhase
    State         *GameState
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type GamePhase string

const (
    PhaseMorning       GamePhase = "morning"
    PhaseInvestigation GamePhase = "investigation"
    PhaseEncounter     GamePhase = "encounter"
    PhaseAftermath     GamePhase = "aftermath"
)

type GameState struct {
    // 当前场景
    CurrentSceneID    string
    VisitedScenes     map[string]bool
    
    // 线索和进度
    CollectedClues    []string
    UnlockedLocations []string
    DomainUnlocked    bool
    
    // NPC状态
    NPCStates         map[string]*NPCState
    
    // 资源
    ChaosPool         int
    LooseEnds         int
    
    // 任务结果
    AnomalyStatus     string
    MissionOutcome    string
}
```

#### 1.2 角色 (Agent)

```go
package domain

type Agent struct {
    ID            string
    Name          string
    Pronouns      string
    
    // ARC组件
    Anomaly       *Anomaly
    Reality       *Reality
    Career        *Career
    
    // 资质保证
    QA            map[string]int
    
    // 人际关系
    Relationships []*Relationship
    
    // 绩效
    Commendations int
    Reprimands    int
    Rating        string
    
    // 状态
    Alive         bool
    InDebt        bool
    
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Anomaly struct {
    Type      string
    Abilities []string
}

type Reality struct {
    Type              string
    SpecialFeature    map[string]interface{}
    DegradationTrack  *DegradationTrack
    Relationships     []*Relationship
}

type Career struct {
    Type       string
    QA         map[string]int
    Claimables []string
}

type Relationship struct {
    ID          string
    Name        string
    Description string
    Connection  int
    PlayedBy    string
    Notes       []string
}

type DegradationTrack struct {
    Name   string
    Filled int
    Total  int
}
```

#### 1.3 剧本 (Scenario)

```go
package domain

type Scenario struct {
    ID          string
    Name        string
    Description string
    
    // 异常体档案
    Anomaly     *AnomalyProfile
    
    // 任务前夕
    MorningScenes    []*MorningScene
    Briefing         *Briefing
    OptionalGoals    []*OptionalGoal
    
    // 调查阶段
    Scenes           map[string]*Scene
    StartingSceneID  string
    
    // 遭遇阶段
    Encounter        *Encounter
    
    // 余波
    Aftermath        *Aftermath
    
    // 奖励
    Rewards          *Rewards
}

type AnomalyProfile struct {
    ID            string
    Name          string
    History       string
    Focus         *Focus
    Domain        *Domain
    Appearance    string
    Impulse       string
    CurrentStatus string
    ChaosEffects  []*ChaosEffect
}

type Focus struct {
    Emotion string
    Subject string
}

type Domain struct {
    Location    string
    Description string
}
```

### 2. 核心服务接口

#### 2.1 游戏服务

```go
package service

type GameService interface {
    // 会话管理
    CreateSession(agentID, scenarioID string) (*GameSession, error)
    GetSession(sessionID string) (*GameSession, error)
    SaveSession(session *GameSession) error
    DeleteSession(sessionID string) error
    
    // 游戏流程
    StartMorningPhase(sessionID string) (*MorningPhaseResult, error)
    StartInvestigationPhase(sessionID string) (*InvestigationPhaseResult, error)
    StartEncounterPhase(sessionID string) (*EncounterPhaseResult, error)
    
    // 阶段转换
    TransitionPhase(sessionID string, toPhase GamePhase) error
}
```

#### 2.2 骰子服务

```go
package service

type DiceService interface {
    // 基础掷骰
    Roll(count int) *RollResult
    
    // 能力掷骰
    RollForAbility(agent *Agent, ability *Ability) *RollResult
    
    // 请求机构掷骰
    RollForRequest(agent *Agent, quality string) *RollResult
    
    // 应用QA调整
    ApplyQA(roll *RollResult, quality string, amount int) *RollResult
    
    // 应用过载
    ApplyOverload(roll *RollResult, amount int) *RollResult
    
    // 检查三重升华
    CheckTripleAscension(roll *RollResult) bool
}

type RollResult struct {
    Dice      []int
    Threes    int
    Success   bool
    Chaos     int
    Overload  int
    TripleAsc bool
}
```

#### 2.3 AI服务

```go
package service

type AIService interface {
    // 场景描述
    GenerateSceneDescription(scene *Scene, state *GameState) (string, error)
    
    // NPC对话
    GenerateNPCDialogue(npc *NPC, context *DialogueContext) (string, error)
    
    // 混沌效应决策
    SelectChaosEffect(anomaly *AnomalyProfile, chaosPool int, context *GameState) (*ChaosEffect, error)
    
    // 事件描述
    DescribeEvent(event *Event, context *GameState) (string, error)
    
    // 结果叙述
    NarrateResult(action *Action, result *ActionResult) (string, error)
}
```

#### 2.4 剧本服务

```go
package service

type ScenarioService interface {
    // 剧本管理
    LoadScenario(scenarioID string) (*Scenario, error)
    ListScenarios() ([]*ScenarioSummary, error)
    ValidateScenario(scenario *Scenario) error
    
    // 场景导航
    GetScene(scenarioID, sceneID string) (*Scene, error)
    GetAvailableScenes(sessionID string) ([]*Scene, error)
    
    // 线索系统
    GetClue(scenarioID, clueID string) (*Clue, error)
    CheckClueRequirements(clue *Clue, state *GameState) bool
    
    // 事件系统
    CheckEventTriggers(scenario *Scenario, state *GameState) ([]*Event, error)
}
```

#### 2.5 角色服务

```go
package service

type AgentService interface {
    // 角色管理
    CreateAgent(req *CreateAgentRequest) (*Agent, error)
    GetAgent(agentID string) (*Agent, error)
    UpdateAgent(agent *Agent) error
    DeleteAgent(agentID string) error
    
    // ARC管理
    SetAnomaly(agentID string, anomalyType string) error
    SetReality(agentID string, realityType string) error
    SetCareer(agentID string, careerType string) error
    
    // 资质保证
    SpendQA(agentID, quality string, amount int) error
    RestoreQA(agentID string) error
    
    // 人际关系
    AddRelationship(agentID string, rel *Relationship) error
    UpdateRelationship(agentID, relID string, connection int) error
    
    // 绩效
    AddCommendations(agentID string, amount int) error
    AddReprimands(agentID string, amount int) error
    UpdateRating(agentID string) error
}
```

## 数据模型

### 数据库Schema

```sql
-- 角色表
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    pronouns VARCHAR(50),
    anomaly_type VARCHAR(50) NOT NULL,
    reality_type VARCHAR(50) NOT NULL,
    career_type VARCHAR(50) NOT NULL,
    qa JSONB NOT NULL,
    relationships JSONB NOT NULL,
    commendations INTEGER DEFAULT 0,
    reprimands INTEGER DEFAULT 0,
    rating VARCHAR(50) DEFAULT '评级良好',
    alive BOOLEAN DEFAULT true,
    in_debt BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 游戏会话表
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    scenario_id VARCHAR(100) NOT NULL,
    phase VARCHAR(50) NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 存档表
CREATE TABLE saves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id),
    name VARCHAR(255) NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_agents_name ON agents(name);
CREATE INDEX idx_sessions_agent ON game_sessions(agent_id);
CREATE INDEX idx_sessions_phase ON game_sessions(phase);
CREATE INDEX idx_saves_session ON saves(session_id);
```

### Redis缓存结构

```
# 会话缓存
session:{sessionID} -> GameSession JSON (TTL: 24h)

# 场景缓存
scene:{scenarioID}:{sceneID} -> Scene JSON (TTL: 1h)

# 剧本缓存
scenario:{scenarioID} -> Scenario JSON (TTL: 1h)

# 角色缓存
agent:{agentID} -> Agent JSON (TTL: 1h)
```

## 正确性属性

*一个属性是一个特征或行为，应该在系统的所有有效执行中保持为真——本质上，是关于系统应该做什么的正式陈述。属性作为人类可读规范和机器可验证正确性保证之间的桥梁。*


### 属性反思

在编写正确性属性之前，我们需要识别并消除冗余：

1. **骰子系统属性**: 需求2.1-2.3可以合并为一个综合属性，涵盖掷骰、成功判定和混沌生成
2. **资质保证属性**: 需求4.1和4.5可以合并，都是关于QA的不变量
3. **状态持久化属性**: 需求11.1和11.2是round-trip property，应该合并
4. **场景状态属性**: 需求13.2和13.5都是关于场景状态保存，可以合并

### 核心正确性属性

基于prework分析，以下是系统的核心正确性属性：

#### 属性1: 角色创建完整性
*对于任何*有效的ARC组合（异常、现实、职能），创建角色后应该包含所有必需组件：3种异常能力、现实触发器、过载解除、3段人际关系（总计12点连结）、9点资质保证。

**验证需求**: 1.1, 1.2, 1.3, 1.4, 1.5

#### 属性2: 骰子判定一致性
*对于任何*掷骰，应该投掷恰好6颗四面骰（每颗结果1-4），统计"3"的数量，至少一个"3"为成功，零个"3"为失败，失败时每颗非"3"骰子产生1点混沌。

**验证需求**: 2.1, 2.2, 2.3

#### 属性3: 三重升华零混沌
*对于任何*掷骰结果，如果在任何调整前恰好有三个"3"，则该掷骰产生零点混沌，无论是否应用QA或过载。

**验证需求**: 2.4

#### 属性4: 资质保证不变量
*对于任何*角色，所有资质的QA总和应该始终等于或小于初始分配的9点（消耗后），且任何单次花费不能超过该资质的当前可用QA。

**验证需求**: 4.1, 4.5

#### 属性5: 过载机制
*对于任何*在QA为0的资质上进行的掷骰，应该应用过载效果：移除一个"3"（如果有）并产生1点混沌。当过载解除条件满足时，所有过载应该被清零。

**验证需求**: 4.2, 4.3

#### 属性6: 任务间隙恢复
*对于任何*角色，在任务间隙时，所有资质的QA应该恢复到其当前上限（不超过初始分配）。

**验证需求**: 4.4

#### 属性7: 混沌守恒
*对于任何*游戏会话，混沌池的变化应该遵循：失败掷骰增加混沌（每颗非"3"骰子+1），异常体使用效应减少混沌，任务开始时等于累积散逸端，任务结束时归零。

**验证需求**: 5.1, 5.2, 5.3, 5.4

#### 属性8: 请求机构地点过载
*对于任何*地点，如果玩家在该地点请求机构失败，该地点的后续请求应该累积1点过载，直到离开该地点。

**验证需求**: 5.5

#### 属性9: 请求机构约束
*对于任何*现实变更请求，必须包含效果、因果链、资质和掷骰四个要素。已确立的事实不能被改变，直接心智控制必须被拒绝。

**验证需求**: 6.1, 6.4, 6.5

#### 属性10: 异常能力效果一致性
*对于任何*异常能力的使用，成功时应用"成功时"效果，失败时应用"失败时"效果并产生混沌，满足额外条件时应用对应的额外效果。

**验证需求**: 7.2, 7.3, 7.4

#### 属性11: 人寿保险机制
*对于任何*伤害，玩家可以花费等量QA（任意资质）来无视伤害。如果无法或不愿花费，玩家死亡并扣除5次嘉奖，然后在休息室复活。

**验证需求**: 8.1, 8.2, 8.3

#### 属性12: 伤害与散逸端
*对于任何*超过1点的伤害，如果有目击者，应该产生等于伤害点数的散逸端。

**验证需求**: 8.4

#### 属性13: 机构评级映射
*对于任何*角色，机构评级应该与申诫总数一一对应：0申诫="评级良好"，1申诫="有待改进"，...，10+申诫="权限已撤销"。

**验证需求**: 9.3

#### 属性14: 任务结果奖励
*对于任何*完成的任务，捕获异常体应该给予每名特工3次嘉奖，中和异常体无奖惩，逃脱应该给予每名特工3次申诫。

**验证需求**: 3.4

#### 属性15: 存档round-trip
*对于任何*游戏状态，保存后立即加载应该恢复完全相同的状态（序列化后反序列化应该是恒等操作）。

**验证需求**: 11.1, 11.2

#### 属性16: 场景状态持久化
*对于任何*场景，玩家离开后再返回，场景应该保持离开时的状态，而不是初始状态。场景的变化应该被记录并恢复。

**验证需求**: 13.2, 13.5

#### 属性17: 线索解锁单调性
*对于任何*游戏会话，已收集的线索集合应该单调递增（只增不减），已解锁的地点和选项也应该单调递增。

**验证需求**: 14.1, 14.3

#### 属性18: NPC状态一致性
*对于任何*NPC，其状态变化应该反映在后续的对话和行为中。被异常体影响的NPC应该表现出不同的行为模式。

**验证需求**: 15.3, 15.4

## 错误处理

### 错误类型

```go
package errors

type ErrorCode string

const (
    // 验证错误
    ErrInvalidInput      ErrorCode = "INVALID_INPUT"
    ErrInvalidARC        ErrorCode = "INVALID_ARC"
    ErrInvalidAction     ErrorCode = "INVALID_ACTION"
    
    // 资源错误
    ErrInsufficientQA    ErrorCode = "INSUFFICIENT_QA"
    ErrInsufficientChaos ErrorCode = "INSUFFICIENT_CHAOS"
    
    // 状态错误
    ErrInvalidPhase      ErrorCode = "INVALID_PHASE"
    ErrInvalidState      ErrorCode = "INVALID_STATE"
    
    // 数据错误
    ErrNotFound          ErrorCode = "NOT_FOUND"
    ErrAlreadyExists     ErrorCode = "ALREADY_EXISTS"
    ErrDataCorrupted     ErrorCode = "DATA_CORRUPTED"
    
    // 系统错误
    ErrInternal          ErrorCode = "INTERNAL_ERROR"
    ErrAIService         ErrorCode = "AI_SERVICE_ERROR"
)

type GameError struct {
    Code    ErrorCode
    Message string
    Details map[string]interface{}
}

func (e *GameError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

### 错误处理策略

1. **验证错误**: 返回详细的错误信息，指导用户修正输入
2. **资源不足**: 明确告知缺少的资源和当前可用量
3. **状态错误**: 说明当前状态和期望状态，提供可行的操作
4. **数据错误**: 保护原始数据，提供恢复选项
5. **系统错误**: 记录详细日志，返回用户友好的错误信息

## 测试策略

### 单元测试

单元测试覆盖具体的功能点和边界情况：

1. **骰子系统**
   - 测试6d4掷骰的基本功能
   - 测试"3"的统计
   - 测试成功/失败判定
   - 测试三重升华的特殊情况

2. **资质保证系统**
   - 测试QA的花费和恢复
   - 测试过载的应用
   - 测试过载解除

3. **混沌系统**
   - 测试混沌的生成和消耗
   - 测试混沌池的初始化和清空

4. **角色系统**
   - 测试ARC组件的创建
   - 测试人际关系的管理
   - 测试绩效的追踪

### 属性测试

属性测试验证系统的通用正确性，使用Go的testing/quick包或第三方库如gopter：

```go
package domain_test

import (
    "testing"
    "testing/quick"
)

// 属性1: 角色创建完整性
// Feature: trpg-solo-engine, Property 1: 角色创建完整性
func TestProperty_AgentCreationCompleteness(t *testing.T) {
    f := func(anomalyType, realityType, careerType string) bool {
        // 生成随机的ARC组合
        agent, err := CreateAgent(anomalyType, realityType, careerType)
        if err != nil {
            return true // 无效组合跳过
        }
        
        // 验证完整性
        return len(agent.Anomaly.Abilities) == 3 &&
               agent.Reality.Trigger != nil &&
               agent.Reality.OverloadRelief != nil &&
               len(agent.Relationships) == 3 &&
               sumConnections(agent.Relationships) == 12 &&
               sumQA(agent.QA) == 9
    }
    
    if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}

// 属性2: 骰子判定一致性
// Feature: trpg-solo-engine, Property 2: 骰子判定一致性
func TestProperty_DiceRollConsistency(t *testing.T) {
    f := func() bool {
        roll := RollDice(6)
        
        // 验证骰子数量
        if len(roll.Dice) != 6 {
            return false
        }
        
        // 验证骰子范围
        for _, die := range roll.Dice {
            if die < 1 || die > 4 {
                return false
            }
        }
        
        // 验证成功判定
        threes := countThrees(roll.Dice)
        if (threes > 0) != roll.Success {
            return false
        }
        
        // 验证混沌生成
        expectedChaos := 6 - threes
        return roll.Chaos == expectedChaos
    }
    
    if err := quick.Check(f, &quick.Config{MaxCount: 1000}); err != nil {
        t.Error(err)
    }
}

// 属性15: 存档round-trip
// Feature: trpg-solo-engine, Property 15: 存档round-trip
func TestProperty_SaveLoadRoundTrip(t *testing.T) {
    f := func(state GameState) bool {
        // 序列化
        data, err := json.Marshal(state)
        if err != nil {
            return false
        }
        
        // 反序列化
        var loaded GameState
        err = json.Unmarshal(data, &loaded)
        if err != nil {
            return false
        }
        
        // 验证相等
        return reflect.DeepEqual(state, loaded)
    }
    
    if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

### 集成测试

集成测试验证组件间的协作：

1. **完整游戏流程**: 从角色创建到任务完成的端到端测试
2. **API接口**: 测试所有REST API端点
3. **数据持久化**: 测试数据库操作和缓存
4. **AI集成**: 测试AI服务的调用和响应

### 测试配置

```yaml
# test_config.yaml
testing:
  unit:
    parallel: true
    timeout: 5s
  
  property:
    iterations: 100
    seed: random
  
  integration:
    database: test_db
    redis: test_redis
    ai_mock: true
```

## 性能考虑

### 响应时间目标

- API响应: < 200ms (p95)
- 骰子掷骰: < 10ms
- 场景加载: < 100ms
- AI生成: < 2s
- 存档保存: < 500ms

### 优化策略

1. **缓存策略**
   - 剧本数据缓存1小时
   - 场景数据缓存1小时
   - 角色数据缓存1小时
   - 会话数据缓存24小时

2. **数据库优化**
   - 为常用查询添加索引
   - 使用连接池
   - 批量操作

3. **AI调用优化**
   - 异步生成非关键内容
   - 缓存常见响应
   - 使用流式响应

## 安全考虑

### 认证和授权

```go
// JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        claims, err := ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### 数据验证

所有用户输入必须经过验证：

1. **ARC选择**: 验证类型是否在有效列表中
2. **掷骰请求**: 验证资质名称和QA花费
3. **行动请求**: 验证行动类型和目标
4. **存档操作**: 验证会话所有权

### 速率限制

```go
// 速率限制配置
rateLimits := map[string]int{
    "/api/dice/roll":     100,  // 每分钟100次
    "/api/ai/generate":   10,   // 每分钟10次
    "/api/session/save":  20,   // 每分钟20次
}
```

## 部署架构

### 容器化

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o trpg-engine ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trpg-engine .
COPY --from=builder /app/scenarios ./scenarios
EXPOSE 8080
CMD ["./trpg-engine"]
```

### Kubernetes部署

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trpg-engine
spec:
  replicas: 3
  selector:
    matchLabels:
      app: trpg-engine
  template:
    metadata:
      labels:
        app: trpg-engine
    spec:
      containers:
      - name: trpg-engine
        image: trpg-engine:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: trpg-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: trpg-secrets
              key: redis-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 监控和日志

### 日志级别

```go
// 日志配置
logger := zap.NewProduction()
defer logger.Sync()

// 不同级别的日志
logger.Debug("Dice rolled", zap.Int("threes", threes))
logger.Info("Session created", zap.String("sessionID", id))
logger.Warn("Low QA", zap.String("quality", quality))
logger.Error("AI service failed", zap.Error(err))
```

### 指标收集

```go
// Prometheus指标
var (
    diceRolls = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "trpg_dice_rolls_total",
            Help: "Total number of dice rolls",
        },
    )
    
    sessionDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "trpg_session_duration_seconds",
            Help: "Duration of game sessions",
        },
    )
)
```

## 未来扩展

### 多人模式

虽然当前版本专注于单人体验，但架构设计考虑了未来的多人扩展：

1. **会话共享**: 多个玩家共享同一个游戏会话
2. **角色分配**: 每个玩家控制一个特工
3. **实时同步**: 使用WebSocket进行状态同步
4. **冲突解决**: 实现规则书中的直接冲突机制

### 自定义剧本编辑器

提供可视化工具让用户创建自己的剧本：

1. **场景编辑器**: 拖拽式场景设计
2. **NPC编辑器**: 可视化NPC配置
3. **事件编辑器**: 条件和效果的图形化编辑
4. **验证工具**: 自动检查剧本完整性

### 社区功能

1. **剧本分享**: 用户可以分享自己创建的剧本
2. **评分系统**: 对剧本进行评分和评论
3. **排行榜**: 展示最受欢迎的剧本和最佳玩家
