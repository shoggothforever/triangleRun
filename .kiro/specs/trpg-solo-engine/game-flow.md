# 三角机构游戏流程详解

## 概述

每个《三角机构》任务由三个核心阶段组成：**晨会 (Morning Scenes)** → **调查 (Investigation)** → **遭遇 (Encounter)**

## 阶段一：晨会 (Morning Scenes)

### 目的
- 让玩家从日常生活过渡到任务状态
- 为任务埋下伏笔和线索
- 建立角色的凡俗生活背景
- 低压力的角色扮演热身

### 运作方式

```
1. 系统随机选择 2-3 个晨会场景
2. 每个场景关联一名特工或其重要之人
3. AI总经理描述场景
4. 玩家可以进行简单的角色扮演互动
5. 场景中隐含任务相关的暗示
```

### 示例（永恒之泉）

```json
{
  "morningScenes": [
    {
      "id": "morning-001",
      "trigger": "特工的伴侣",
      "description": "一名特工的伴侣兴奋地分享了他们期待已久的水疗日计划",
      "hint": "暗示源泉水疗中心",
      "interaction": {
        "type": "dialogue",
        "options": [
          "询问水疗中心的名字",
          "表示支持",
          "建议改天再去"
        ]
      }
    },
    {
      "id": "morning-002",
      "trigger": "街上的年轻人",
      "description": "一群年轻人走过，他们似乎都长着同一张脸",
      "hint": "暗示异常体的焕新效应",
      "interaction": {
        "type": "observation",
        "options": [
          "仔细观察他们的脸",
          "跟踪他们",
          "忽略继续前行"
        ]
      }
    }
  ]
}
```

### 技术实现

```go
type MorningScene struct {
    ID          string
    Description string
    Hint        string
    Trigger     string  // 触发对象（特工/重要之人）
    Interaction InteractionType
}

func (g *Game) PlayMorningScenes() {
    // 1. 从剧本中随机选择2-3个晨会场景
    scenes := g.Scenario.SelectMorningScenes(2, 3)
    
    // 2. 对每个场景
    for _, scene := range scenes {
        // 2.1 AI生成场景描述
        description := g.AI.GenerateSceneDescription(scene)
        g.Display(description)
        
        // 2.2 提供互动选项
        options := scene.GetInteractionOptions()
        choice := g.GetPlayerChoice(options)
        
        // 2.3 处理玩家选择
        g.ProcessChoice(scene, choice)
    }
    
    // 3. 过渡到任务简报
    g.TransitionToBriefing()
}
```

### 晨会结束条件
- 所有选定的晨会场景都已呈现
- 玩家选择"前往机构总部"
- 自动过渡到任务简报

---

## 阶段二：调查 (Investigation)

### 目的
- 收集关于异常体的线索
- 找到异常体的领域位置
- 理解异常体的焦点
- 管理散逸端风险

### 运作方式

```
1. 任务简报：AI总经理介绍已知情报
2. 调查循环：
   a. 玩家选择要前往的地点
   b. 系统呈现场景描述
   c. 玩家选择行动（对话/调查/使用能力/移动）
   d. 系统处理行动结果
   e. 更新线索日志
   f. 检查事件触发
   g. 处理混沌效应
   h. 重复直到找到领域
3. 过渡到遭遇
```

### 调查循环详解

#### 2.1 任务简报

```json
{
  "briefing": {
    "summary": "三联城市中心商业大道街区出现异常行为模式",
    "knownFacts": [
      "无故哭泣频率增加",
      "外貌相似的不同人",
      "监控录像异常"
    ],
    "startingLocation": "commercial-avenue",
    "objectives": [
      "找到异常体的领域",
      "理解异常体的焦点",
      "最小化散逸端"
    ]
  }
}
```

#### 2.2 场景探索

```go
type InvestigationLoop struct {
    CurrentScene    *Scene
    VisitedScenes   map[string]bool
    CollectedClues  []*Clue
    ActiveNPCs      map[string]*NPC
    ChaosPool       int
    LooseEnds       int  // 散逸端计数
}

func (inv *InvestigationLoop) Run() {
    for !inv.DomainFound() {
        // 1. 显示当前场景
        inv.DisplayScene()
        
        // 2. 列出可用行动
        actions := inv.GetAvailableActions()
        
        // 3. 玩家选择行动
        action := inv.GetPlayerAction(actions)
        
        // 4. 处理行动
        result := inv.ProcessAction(action)
        
        // 5. 更新状态
        inv.UpdateState(result)
        
        // 6. 检查事件
        inv.CheckEvents()
        
        // 7. 处理混沌
        if inv.ChaosPool > 0 {
            inv.ProcessChaosEffects()
        }
    }
}
```

#### 2.3 场景结构

每个场景包含：

```json
{
  "scene": {
    "id": "commercial-avenue",
    "name": "商业大道",
    "description": "时髦的精品店、沙龙、餐厅和酒吧...",
    "locations": [
      {
        "id": "tryptik-cafe",
        "name": "Tryptik咖啡馆",
        "npcs": ["barista", "crying-customers"],
        "clues": ["unusual-behavior", "forget-me-not-drink"],
        "interactables": ["menu", "customers"]
      },
      {
        "id": "the-source-exterior",
        "name": "源泉门外",
        "npcs": ["indigo-jones", "quinn-wilmar", "gregory-doppler"],
        "clues": ["jay-hsieh-promotion", "crowd-gathering"],
        "interactables": ["door", "crowd"]
      }
    ],
    "connections": ["the-source-interior", "laundrocade"],
    "state": {
      "visited": false,
      "eventsTriggered": [],
      "npcStates": {}
    }
  }
}
```

#### 2.4 玩家行动类型

```go
type ActionType string

const (
    ActionMove        ActionType = "move"         // 移动到新地点
    ActionTalk        ActionType = "talk"         // 与NPC对话
    ActionInvestigate ActionType = "investigate"  // 调查对象/环境
    ActionUseAbility  ActionType = "use_ability"  // 使用异常能力
    ActionRequest     ActionType = "request"      // 请求机构
    ActionWait        ActionType = "wait"         // 等待/观察
)

type PlayerAction struct {
    Type   ActionType
    Target string  // NPC ID, 对象 ID, 或地点 ID
    Params map[string]interface{}
}
```

#### 2.5 线索收集

```go
type Clue struct {
    ID          string
    Name        string
    Type        ClueType  // mundane 或 anomalous
    Content     string
    Source      string    // NPC ID 或 location ID
    Category    string    // "anomaly-effect", "location", "npc-info"
    Unlocks     []string  // 解锁的地点或选项
    Required    bool      // 是否为必要线索
    Collected   bool
}

func (inv *InvestigationLoop) CollectClue(clue *Clue) {
    // 1. 添加到线索日志
    inv.CollectedClues = append(inv.CollectedClues, clue)
    
    // 2. 检查是否解锁新内容
    for _, unlock := range clue.Unlocks {
        inv.UnlockContent(unlock)
    }
    
    // 3. 检查是否满足进入领域的条件
    if inv.HasRequiredClues() {
        inv.UnlockDomain()
    }
}
```

#### 2.6 NPC互动

```go
type NPCInteraction struct {
    NPC      *NPC
    Dialogue []DialogueOption
    State    NPCState
}

type DialogueOption struct {
    ID          string
    Text        string
    Requirements *Requirement  // 可能需要特定线索或能力
    Reveals     []*Clue
    Effects     []Effect
}

func (inv *InvestigationLoop) TalkToNPC(npcID string) {
    npc := inv.ActiveNPCs[npcID]
    
    // 1. 获取可用对话选项
    options := npc.GetDialogueOptions(inv.CollectedClues)
    
    // 2. 玩家选择对话
    choice := inv.GetPlayerChoice(options)
    
    // 3. 处理对话结果
    result := npc.ProcessDialogue(choice)
    
    // 4. 揭示新线索
    for _, clue := range result.Reveals {
        inv.CollectClue(clue)
    }
    
    // 5. 应用效果
    for _, effect := range result.Effects {
        inv.ApplyEffect(effect)
    }
}
```

#### 2.7 事件触发

```go
type Event struct {
    ID          string
    Name        string
    Trigger     TriggerCondition
    Description string
    Effects     []Effect
    Resolution  []ResolutionOption
}

type TriggerCondition struct {
    Type      string  // "proximity", "time", "chaos", "player-action"
    Condition string
    Params    map[string]interface{}
}

func (inv *InvestigationLoop) CheckEvents() {
    for _, event := range inv.CurrentScene.Events {
        if event.ShouldTrigger(inv) {
            inv.TriggerEvent(event)
        }
    }
}

func (inv *InvestigationLoop) TriggerEvent(event *Event) {
    // 1. 显示事件描述
    inv.Display(event.Description)
    
    // 2. 应用效果
    for _, effect := range event.Effects {
        inv.ApplyEffect(effect)
    }
    
    // 3. 提供解决选项
    if len(event.Resolution) > 0 {
        choice := inv.GetPlayerChoice(event.Resolution)
        inv.ResolveEvent(event, choice)
    }
}
```

#### 2.8 混沌效应

```go
func (inv *InvestigationLoop) ProcessChaosEffects() {
    anomaly := inv.Scenario.Anomaly
    
    // 1. AI总经理决定使用哪个混沌效应
    effect := inv.AI.SelectChaosEffect(anomaly, inv.ChaosPool)
    
    // 2. 检查是否有足够混沌
    if inv.ChaosPool >= effect.Cost {
        // 3. 扣除混沌
        inv.ChaosPool -= effect.Cost
        
        // 4. 应用效应
        inv.ApplyChaosEffect(effect)
        
        // 5. 描述效应
        description := inv.AI.DescribeChaosEffect(effect, inv.CurrentScene)
        inv.Display(description)
    }
}
```

### 调查阶段结束条件

```go
func (inv *InvestigationLoop) DomainFound() bool {
    // 1. 收集了所有必要线索
    hasRequiredClues := inv.HasAllRequiredClues()
    
    // 2. 找到了领域位置
    domainUnlocked := inv.IsDomainUnlocked()
    
    // 3. 玩家选择进入领域
    playerReady := inv.PlayerChoosesToEnter()
    
    return hasRequiredClues && domainUnlocked && playerReady
}
```

---

## 阶段三：遭遇 (Encounter)

### 目的
- 直接面对异常体
- 理解并应对异常体的焦点
- 决定异常体的命运（捕获/中和/逃脱）
- 解决剧情冲突

### 运作方式

```
1. 进入领域：描述异常体的领域环境
2. 遭遇阶段：
   a. 阶段1：初始接触
   b. 阶段2：理解焦点
   c. 阶段3：关键冲突
   d. 阶段4：最终决战或和解
3. 结算：判定任务结果
```

### 遭遇结构

```json
{
  "encounter": {
    "domain": {
      "id": "underground-rainforest",
      "name": "地下雨林",
      "description": "三联城下方有一片雨林...",
      "hazards": ["藤蔓阻挡", "次级异常体", "尸体"],
      "atmosphere": "潮湿、温暖、充满生机"
    },
    "phases": [
      {
        "id": "phase-1",
        "name": "进入领域",
        "description": "特工们穿过下水道进入雨林",
        "challenges": [
          {
            "type": "obstacle",
            "description": "粗壮的藤蔓挡住去路",
            "solutions": ["砍伐", "请求机构", "寻找其他路径"]
          },
          {
            "type": "discovery",
            "description": "发现五具尸体",
            "clue": "异常体的杀伤力"
          }
        ],
        "chaosThreshold": 10
      },
      {
        "id": "phase-2",
        "name": "与异常体对话",
        "description": "异常体通过花朵与特工交流",
        "dialogue": [
          "我不想再被囚禁",
          "我想回到我的家园",
          "你们能帮我吗？"
        ],
        "options": [
          {
            "action": "承诺帮助",
            "requirement": "理解焦点",
            "outcome": "异常体信任增加"
          },
          {
            "action": "欺骗",
            "requirement": "欺瞒检定",
            "outcome": "可能被识破"
          },
          {
            "action": "攻击",
            "outcome": "进入战斗"
          }
        ]
      },
      {
        "id": "phase-3",
        "name": "Serena的忏悔",
        "trigger": "特工威胁异常体",
        "description": "Serena冲出保护异常体",
        "conflict": {
          "parties": ["特工", "Serena", "异常体"],
          "stakes": "Serena的生命，异常体的命运",
          "resolution": "需要调解或战斗"
        }
      }
    ],
    "outcomes": {
      "captured": {
        "condition": "异常体被安抚或筋疲力尽",
        "method": "普通手提箱",
        "rewards": "+3嘉奖每名特工"
      },
      "neutralized": {
        "condition": "使用波纹枪",
        "consequences": "无嘉奖或申诫"
      },
      "escaped": {
        "condition": "异常体逃离",
        "consequences": "-3嘉奖每名特工"
      }
    }
  }
}
```

### 遭遇执行流程

```go
type Encounter struct {
    Domain    *Domain
    Phases    []*EncounterPhase
    Anomaly   *Anomaly
    NPCs      map[string]*NPC
    ChaosPool int
}

func (enc *Encounter) Run() EncounterOutcome {
    // 1. 进入领域
    enc.EnterDomain()
    
    // 2. 执行各个阶段
    for _, phase := range enc.Phases {
        result := enc.ExecutePhase(phase)
        
        // 检查是否提前结束
        if result.IsTerminal {
            return enc.ResolveOutcome(result)
        }
    }
    
    // 3. 最终决战
    finalResult := enc.FinalConfrontation()
    
    // 4. 判定结果
    return enc.ResolveOutcome(finalResult)
}

func (enc *Encounter) ExecutePhase(phase *EncounterPhase) PhaseResult {
    // 1. 描述阶段
    enc.Display(phase.Description)
    
    // 2. 处理挑战
    for _, challenge := range phase.Challenges {
        enc.HandleChallenge(challenge)
    }
    
    // 3. 处理对话
    if phase.HasDialogue() {
        enc.HandleDialogue(phase)
    }
    
    // 4. 检查触发条件
    if phase.HasTrigger() && phase.Trigger.IsMet(enc) {
        enc.TriggerPhaseEvent(phase)
    }
    
    // 5. 处理混沌
    if enc.ChaosPool >= phase.ChaosThreshold {
        enc.ProcessChaosEffects()
    }
    
    return PhaseResult{
        Success: true,
        IsTerminal: false,
    }
}
```

### 捕获条件判定

```go
func (enc *Encounter) CanCapture() bool {
    anomaly := enc.Anomaly
    
    // 检查捕获条件
    conditions := []bool{
        anomaly.IsSatisfied(),    // 焦点被满足
        anomaly.IsConfused(),     // 陷入困惑
        anomaly.IsDespairing(),   // 陷入绝望
        anomaly.IsExhausted(),    // 筋疲力尽
        anomaly.IsWilling(),      // 自愿
    }
    
    // 任一条件满足即可捕获
    for _, condition := range conditions {
        if condition {
            return true
        }
    }
    
    return false
}

func (enc *Encounter) CaptureAnomaly() {
    // 1. 使用普通手提箱
    enc.Display("你打开普通手提箱...")
    
    // 2. 捕获动画/描述
    enc.AI.DescribeCapture(enc.Anomaly)
    
    // 3. 更新状态
    enc.Anomaly.Status = "captured"
    
    // 4. 奖励嘉奖
    enc.AwardCommendations(3)
}
```

### 战斗系统（如需要）

```go
type Combat struct {
    Participants []*Combatant
    TurnOrder    []string
    Round        int
}

type Combatant struct {
    ID        string
    Name      string
    Threat    int  // 攻击力
    Stability int  // 生命值
    Abilities []*Ability
}

func (combat *Combat) Run() CombatResult {
    for !combat.IsOver() {
        // 1. 确定行动顺序
        combat.DetermineTurnOrder()
        
        // 2. 每个参与者行动
        for _, id := range combat.TurnOrder {
            combatant := combat.GetCombatant(id)
            
            if combatant.IsPlayer() {
                // 玩家回合
                action := combat.GetPlayerAction()
                combat.ExecuteAction(combatant, action)
            } else {
                // AI回合
                action := combat.AI.DecideAction(combatant)
                combat.ExecuteAction(combatant, action)
            }
            
            // 检查战斗是否结束
            if combat.IsOver() {
                break
            }
        }
        
        combat.Round++
    }
    
    return combat.DetermineWinner()
}
```

---

## 三阶段流转示意图

```
┌─────────────────────────────────────────────────────────────┐
│                        游戏开始                              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   阶段1: 晨会                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 1. 选择2-3个晨会场景                                  │  │
│  │ 2. AI描述场景                                        │  │
│  │ 3. 玩家简单互动                                      │  │
│  │ 4. 埋下任务伏笔                                      │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   任务简报                                   │
│  - AI总经理介绍已知情报                                     │
│  - 说明调查起点                                             │
│  - 列出可选目标                                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   阶段2: 调查                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 调查循环 (重复直到找到领域):                         │  │
│  │                                                       │  │
│  │ 1. 显示当前场景                                      │  │
│  │    ├─ 场景描述                                       │  │
│  │    ├─ 可见NPC                                        │  │
│  │    └─ 可交互对象                                     │  │
│  │                                                       │  │
│  │ 2. 列出可用行动                                      │  │
│  │    ├─ 移动到新地点                                   │  │
│  │    ├─ 与NPC对话                                      │  │
│  │    ├─ 调查对象                                       │  │
│  │    ├─ 使用异常能力                                   │  │
│  │    └─ 请求机构                                       │  │
│  │                                                       │  │
│  │ 3. 玩家选择行动                                      │  │
│  │                                                       │  │
│  │ 4. 处理行动结果                                      │  │
│  │    ├─ 掷骰（如需要）                                 │  │
│  │    ├─ 收集线索                                       │  │
│  │    ├─ 更新NPC状态                                    │  │
│  │    └─ 产生混沌                                       │  │
│  │                                                       │  │
│  │ 5. 检查事件触发                                      │  │
│  │    └─ 执行触发的事件                                 │  │
│  │                                                       │  │
│  │ 6. 处理混沌效应                                      │  │
│  │    └─ AI选择并应用混沌效应                          │  │
│  │                                                       │  │
│  │ 7. 追踪散逸端                                        │  │
│  │                                                       │  │
│  │ 8. 检查是否找到领域                                  │  │
│  │    ├─ 是 → 退出循环                                  │  │
│  │    └─ 否 → 继续循环                                  │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   阶段3: 遭遇                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 1. 进入领域                                          │  │
│  │    ├─ 描述领域环境                                   │  │
│  │    └─ 处理进入障碍                                   │  │
│  │                                                       │  │
│  │ 2. 执行遭遇阶段                                      │  │
│  │    ├─ 阶段1: 初始接触                                │  │
│  │    ├─ 阶段2: 理解焦点                                │  │
│  │    ├─ 阶段3: 关键冲突                                │  │
│  │    └─ 阶段4: 最终决战/和解                           │  │
│  │                                                       │  │
│  │ 3. 与异常体互动                                      │  │
│  │    ├─ 对话                                           │  │
│  │    ├─ 战斗                                           │  │
│  │    └─ 谈判                                           │  │
│  │                                                       │  │
│  │ 4. 判定结果                                          │  │
│  │    ├─ 已捕获 (+3嘉奖)                                │  │
│  │    ├─ 已中和 (无奖惩)                                │  │
│  │    └─ 已逃脱 (-3嘉奖)                                │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   任务结算                                   │
│  1. 生成任务报告                                            │
│  2. 计算嘉奖和申诫                                          │
│  3. 处理散逸端                                              │
│  4. 应用余波效果                                            │
│  5. 解锁奖励                                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                        游戏结束                              │
└─────────────────────────────────────────────────────────────┘
```

## 关键数据流

### 状态追踪

```go
type GameState struct {
    // 当前阶段
    Phase GamePhase  // "morning", "investigation", "encounter"
    
    // 调查状态
    CurrentScene    *Scene
    VisitedScenes   map[string]bool
    CollectedClues  []*Clue
    UnlockedLocations []string
    
    // NPC状态
    NPCStates       map[string]*NPCState
    
    // 资源
    ChaosPool       int
    LooseEnds       int
    
    // 玩家状态
    QualityAssurance map[string]int
    Commendations   int
    Reprimands      int
    
    // 任务进度
    DomainUnlocked  bool
    AnomalyStatus   string  // "active", "captured", "neutralized", "escaped"
}
```

### 阶段转换

```go
func (g *Game) TransitionPhase(from, to GamePhase) {
    switch to {
    case PhaseMorning:
        g.InitializeMorning()
        
    case PhaseInvestigation:
        g.PlayBriefing()
        g.InitializeInvestigation()
        
    case PhaseEncounter:
        g.SaveInvestigationState()
        g.InitializeEncounter()
        
    case PhaseAftermath:
        g.CalculateResults()
        g.GenerateReport()
    }
    
    g.State.Phase = to
}
```

## 总结

三个阶段的核心区别：

1. **晨会**: 线性、低互动、叙事导向
2. **调查**: 非线性、高互动、探索导向
3. **遭遇**: 半线性、高风险、冲突导向

每个阶段都有明确的开始和结束条件，通过状态追踪和条件检查实现自然流转。
