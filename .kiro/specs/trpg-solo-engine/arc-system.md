# ARC系统详细结构

## 概述

ARC是《三角机构》角色构建的三大支柱：
- **A (Anomaly/异常)**: 角色绑定的异常体类型及其能力
- **R (Reality/现实)**: 角色的凡俗生活和人际关系
- **C (Career/职能)**: 角色在机构中的职位和职责

## A - 异常 (Anomaly)

### 异常体类型列表

规则书中提供9种异常体类型，每种有3个获批能力：

1. **低语 (Whisper)** - 操控声音和语言
2. **目录 (Catalog)** - 创造和复制物体
3. **汲取 (Siphon)** - 吸取和转移特质
4. **时计 (Timepiece)** - 操控时间流动
5. **生长 (Growth)** - 改变身体和感知
6. **枪械 (Gun)** - 终结和抹除
7. **梦境 (Dream)** - 创造幻象和进入艺术
8. **流形 (Manifold)** - 操控空间和重力
9. **缺位 (Absence)** - 消失和遗忘

### 异常能力结构

每个异常能力包含以下元素：

```json
{
  "ability": {
    "id": "whisper-say-again",
    "name": "再说一遍？",
    "anomalyType": "低语",
    "trigger": {
      "type": "response",
      "description": "用'再说一遍？'回应一句说出的话"
    },
    "roll": {
      "quality": "气场",
      "dice": "6d4"
    },
    "effects": {
      "success": {
        "description": "目标会相信新的那句话才是他们本来的意思",
        "mechanics": "改写目标刚说的话"
      },
      "failure": {
        "description": "目标不受影响，你只能使用那句话中的词语交谈3小时",
        "mechanics": "限制玩家词汇"
      },
      "additional": [
        {
          "condition": "六个或更多3",
          "effect": "你可以在接下来的一小时内随时替目标说话"
        }
      ]
    },
    "chaosGeneration": {
      "onRoll": "每颗非3骰子产生1点混沌",
      "onFailure": "失败时产生的混沌可被异常体利用"
    }
  }
}
```

### 异常能力的关键属性

#### 1. 触发条件 (Trigger)
```go
type TriggerType string

const (
    TriggerAction    TriggerType = "action"     // 主动使用
    TriggerResponse  TriggerType = "response"   // 响应某事
    TriggerPassive   TriggerType = "passive"    // 被动触发
    TriggerReactive  TriggerType = "reactive"   // 反应性触发
)

type AbilityTrigger struct {
    Type        TriggerType
    Description string
    Condition   string  // 具体触发条件
}
```

#### 2. 掷骰机制
```go
type AbilityRoll struct {
    Quality     string  // 相关资质（专注、共情、气场等）
    DiceCount   int     // 骰子数量（通常为6）
    DiceType    int     // 骰子类型（d4）
    Modifiers   []Modifier
}
```

#### 3. 效果系统
```go
type AbilityEffects struct {
    Success    Effect      // 成功时效果
    Failure    Effect      // 失败时效果
    Additional []ConditionalEffect  // 额外效果
}

type Effect struct {
    Description string
    Mechanics   string
    Duration    string
    Target      string
}

type ConditionalEffect struct {
    Condition string  // "每额外一个3", "每第三个3", "六个或更多3"
    Effect    Effect
}
```

### 示例：完整的异常能力定义

```json
{
  "ability": {
    "id": "growth-eyes",
    "name": "眼睛",
    "anomalyType": "生长",
    "trigger": {
      "type": "action",
      "description": "睁开几只额外的眼睛"
    },
    "roll": {
      "quality": "专业",
      "dice": "6d4"
    },
    "effects": {
      "success": {
        "description": "你的身体上会长出新的眼睛，并带来强大的新能力",
        "mechanics": "在视觉类型上花费3，效果持续1小时",
        "options": [
          {"cost": 1, "type": "热成像、夜视或望远"},
          {"cost": 2, "type": "指纹或X光"},
          {"cost": 3, "type": "现实（洞穿幻象与混淆）"},
          {"cost": 4, "type": "植物手语、异常体追踪"},
          {"cost": 6, "type": "弱点"},
          {"cost": 7, "type": "未来视"}
        ]
      },
      "failure": {
        "description": "你会看见终末的幻象",
        "mechanics": "任务剩余时间所有掷骰受到1点额外过载"
      }
    }
  }
}
```

---

## R - 现实 (Reality)

### 现实类型列表

规则书中提供9种现实类型：

1. **看护者 (Caretaker)** - 照顾受照料者
2. **日程过载 (Schedule Overload)** - 兼顾多份工作
3. **受追猎者 (Hunted)** - 躲避过去
4. **明星 (Star)** - 公众人物
5. **挣扎求生 (Struggling)** - 经济困难
6. **新生儿 (Newborn)** - 初来乍到
7. **浪漫主义 (Romantic)** - 情感复杂
8. **支柱 (Pillar)** - 社区依靠
9. **异类 (Outsider)** - 格格不入

### 现实系统结构

每个现实类型包含：

```json
{
  "reality": {
    "id": "caretaker",
    "name": "看护者",
    "description": "与受照料者有深厚羁绊",
    "specialFeature": {
      "type": "dependent",
      "name": "受照料者",
      "options": ["婴儿", "动物", "新生AI", "外星人"],
      "mechanics": "共享角色，享受人寿保险福利"
    },
    "realityTrigger": {
      "name": "需要关爱",
      "cost": "混沌消耗",
      "effect": "受照料者需要你的关注",
      "consequence": "如果忽视，与受照料者情谊最淡的人际关系失去1点连结"
    },
    "overloadRelief": {
      "name": "这是你的最爱！",
      "condition": "做某件能让受照料者开心的事",
      "effect": "无视所有过载"
    },
    "degradationTrack": {
      "name": "独立",
      "boxes": 4,
      "trigger": "让受照料者自己解决问题、伤害他们或交给他人监管",
      "consequence": "填满后受照料者不再依赖你，必须选择新现实"
    },
    "relationships": {
      "count": 3,
      "totalConnection": 12,
      "distribution": [6, 3, 3],
      "questions": [
        "如果你不在了，谁会获得受照料者的监护权？",
        "谁怀念你曾经拥有的自由？",
        "受照料者总是很兴奋能和谁待在一起？"
      ]
    }
  }
}
```

### 现实的关键组件

#### 1. 现实触发器 (Reality Trigger)
```go
type RealityTrigger struct {
    Name        string
    Cost        int     // 混沌消耗
    Effect      string  // 触发效果
    Consequence string  // 忽视后果
}

// GM可以消耗混沌来激活现实触发器
func (gm *GM) ActivateRealityTrigger(agent *Agent) {
    trigger := agent.Reality.Trigger
    
    if gm.ChaosPool >= trigger.Cost {
        gm.ChaosPool -= trigger.Cost
        gm.ApplyTriggerEffect(agent, trigger)
        
        // 如果玩家忽视
        if agent.IgnoresTrigger() {
            gm.ApplyConsequence(agent, trigger.Consequence)
        }
    }
}
```

#### 2. 过载解除 (Overload Relief)
```go
type OverloadRelief struct {
    Name      string
    Condition string  // 激活条件
    Effect    string  // "无视所有过载"
}

// 检查是否满足过载解除条件
func (agent *Agent) CheckOverloadRelief(action Action) bool {
    relief := agent.Reality.OverloadRelief
    return relief.ConditionMet(action)
}

// 应用过载解除
func (roll *DiceRoll) ApplyOverloadRelief(agent *Agent) {
    if agent.HasActiveOverloadRelief() {
        roll.Overload = 0  // 清除所有过载
    }
}
```

#### 3. 退化轨道 (Degradation Track)
```go
type DegradationTrack struct {
    Name        string
    Boxes       int
    Filled      int
    Trigger     string  // 什么行为会标记格子
    Consequence string  // 填满后的后果
}

func (track *DegradationTrack) Mark() {
    track.Filled++
    
    if track.IsFull() {
        // 触发后果：必须选择新现实
        return true
    }
    return false
}
```

#### 4. 人际关系系统 (Relationships)
```go
type Relationship struct {
    ID          string
    Name        string
    Description string
    Connection  int     // 连结点数（1-6）
    PlayedBy    string  // 由哪位玩家/GM扮演
    Notes       []string
}

type RelationshipSystem struct {
    Relationships []*Relationship
    TotalConnection int  // 总连结点数（固定12）
}

// 失去连结
func (rel *Relationship) LoseConnection(amount int) {
    rel.Connection -= amount
    if rel.Connection < 0 {
        rel.Connection = 0
    }
}

// 获得连结
func (rel *Relationship) GainConnection(amount int) {
    rel.Connection += amount
    if rel.Connection > 6 {
        rel.Connection = 6
    }
}
```

### 人际关系的作用

```go
// 人际关系在游戏中的影响
type RelationshipEffects struct {
    // 1. 现实触发器的目标
    TriggerTarget bool
    
    // 2. 过载解除的动机
    OverloadMotivation bool
    
    // 3. 剧情钩子
    StoryHooks []string
    
    // 4. 情感支持
    EmotionalSupport bool
}

// 示例：现实触发器影响人际关系
func (gm *GM) ApplyTriggerConsequence(agent *Agent, trigger RealityTrigger) {
    // 找到连结最低的人际关系
    weakestRel := agent.GetWeakestRelationship()
    
    // 失去1点连结
    weakestRel.LoseConnection(1)
    
    gm.Narrate(fmt.Sprintf("%s与%s的关系受损", agent.Name, weakestRel.Name))
}
```

### 示例：完整的现实定义

```json
{
  "reality": {
    "id": "hunted",
    "name": "受追猎者",
    "description": "正在躲避自己的过去",
    "specialFeature": {
      "type": "dark-past",
      "categories": ["罪犯", "被猎者", "不告而别"],
      "secret": "玩家私下告知GM"
    },
    "realityTrigger": {
      "name": "踪迹暴露",
      "cost": 0,
      "effect": "有人认出了你",
      "consequence": "如果不消除踪迹，最不熟的人际关系失去1点连结并问棘手问题"
    },
    "overloadRelief": {
      "name": "不是我",
      "condition": "做某件能掩盖踪迹的事",
      "effect": "无视所有过载"
    },
    "degradationTrack": {
      "name": "败露",
      "boxes": 4,
      "trigger": "向能认出你的人暴露新身份或位置",
      "consequence": "过去追上你，必须选择新现实"
    },
    "relationships": {
      "count": 3,
      "totalConnection": 12,
      "distribution": [6, 3, 3],
      "questions": [
        "谁知道你的追猎者？",
        "谁会因发现真相而受最深的伤？",
        "谁对新的你痴迷不已？"
      ]
    }
  }
}
```

---

## C - 职能 (Career)

### 职能类型列表

规则书中提供9种职能：

1. **公关 (Public Relations)** - 管理形象
2. **研发 (R&D)** - 研究异常
3. **咖啡师 (Barista)** - 提供服务
4. **首席执行官 (CEO)** - 享受特权
5. **实习生 (Intern)** - 学习成长
6. **掘墓人 (Gravedigger)** - 处理后果
7. **接待处 (Reception)** - 协调沟通
8. **热线 (Hotline)** - 提供支持
9. **小丑 (Clown)** - 娱乐士气

### 职能系统结构

每个职能包含：

```json
{
  "career": {
    "id": "public-relations",
    "name": "公关",
    "description": "负责管理机构形象和控制信息流",
    "initialQA": {
      "total": 9,
      "distribution": {
        "专注": 1,
        "共情": 2,
        "气场": 2,
        "欺瞒": 2,
        "主动": 1,
        "专业": 1,
        "活力": 0,
        "坚毅": 0,
        "诡秘": 0
      }
    },
    "permittedBehaviors": [
      {
        "action": "说服他人相信一切正常",
        "reward": "+1嘉奖"
      },
      {
        "action": "成功消除散逸端",
        "reward": "+2嘉奖"
      },
      {
        "action": "在公众场合代表机构发言",
        "reward": "+1嘉奖"
      }
    ],
    "primeDirective": {
      "description": "绝不让机构的真实运作被公众知晓",
      "violation": "-3申诫"
    },
    "initialClaimable": {
      "id": "pr-spin-kit",
      "name": "公关应急包",
      "description": "包含各种用于控制叙事的工具"
    },
    "assessmentQuestions": [
      "你最擅长哪种社交场合？",
      "你如何处理危机？",
      "你最不想让人知道的秘密是什么？"
    ]
  }
}
```

### 职能的关键组件

#### 1. 资质保证分配 (QA Distribution)
```go
type QualityAssurance struct {
    Qualities map[string]int  // 9种资质的QA点数
    Total     int             // 总计9点
}

const (
    QualityFocus       = "专注"  // 关注细节、发现隐藏
    QualityEmpathy     = "共情"  // 建立联系、发现弱点
    QualityPresence    = "气场"  // 脱颖而出、鼓舞人心
    QualityDeception   = "欺瞒"  // 说谎、说服
    QualityInitiative  = "主动"  // 前瞻思考、迅速行动
    QualityProfession  = "专业"  // 保持镇定、抵抗分心
    QualityVitality    = "活力"  // 进攻、使用武力
    QualityGrit        = "坚毅"  // 拒绝退缩、施加压力
    QualitySubtlety    = "诡秘"  // 悄无声息、避免注意
)

// 职能决定初始QA分配
func NewAgent(career Career) *Agent {
    agent := &Agent{
        Career: career,
        QA:     career.InitialQA.Copy(),
    }
    return agent
}
```

#### 2. 许可行为 (Permitted Behaviors)
```go
type PermittedBehavior struct {
    Action      string
    Reward      int     // 嘉奖数量
    Condition   string  // 可选的额外条件
}

// 检查行为是否获得奖励
func (agent *Agent) CheckPermittedBehavior(action Action) int {
    for _, behavior := range agent.Career.PermittedBehaviors {
        if behavior.Matches(action) {
            return behavior.Reward
        }
    }
    return 0
}
```

#### 3. 首要指令 (Prime Directive)
```go
type PrimeDirective struct {
    Description string
    Violation   int  // 违反时的申诫数量
}

// 检查是否违反首要指令
func (gm *GM) CheckPrimeDirective(agent *Agent, action Action) {
    directive := agent.Career.PrimeDirective
    
    if directive.IsViolated(action) {
        agent.AddReprimands(directive.Violation)
        gm.Narrate(fmt.Sprintf("%s违反了首要指令！", agent.Name))
    }
}
```

#### 4. 自我评估问题 (Assessment Questions)
```go
type AssessmentQuestions struct {
    Questions []string
    Answers   map[string]string
}

// 评估问题用于进一步细化QA分配
func (agent *Agent) CompleteAssessment(answers map[string]string) {
    // 根据答案调整QA分配
    for question, answer := range answers {
        adjustment := agent.Career.EvaluateAnswer(question, answer)
        agent.ApplyQAAdjustment(adjustment)
    }
}
```

### 示例：完整的职能定义

```json
{
  "career": {
    "id": "intern",
    "name": "实习生",
    "description": "机构的新人，渴望学习和证明自己",
    "initialQA": {
      "total": 9,
      "distribution": {
        "专注": 1,
        "共情": 1,
        "气场": 1,
        "欺瞒": 1,
        "主动": 1,
        "专业": 1,
        "活力": 1,
        "坚毅": 1,
        "诡秘": 1
      },
      "note": "实习生在所有资质上都有1点，体现其全面但不精通"
    },
    "permittedBehaviors": [
      {
        "action": "向资深特工请教并应用建议",
        "reward": "+1嘉奖"
      },
      {
        "action": "主动承担额外工作",
        "reward": "+1嘉奖"
      },
      {
        "action": "在任务中学到新技能",
        "reward": "+1嘉奖"
      }
    ],
    "primeDirective": {
      "description": "绝不质疑上级的命令或机构的决策",
      "violation": "-2申诫"
    },
    "specialMechanic": {
      "name": "快速学习",
      "description": "每次任务后可以重新分配1点QA到任何资质"
    },
    "initialClaimable": {
      "id": "intern-handbook",
      "name": "实习生手册",
      "description": "包含基础指南和紧急联系方式"
    },
    "assessmentQuestions": [
      "你为什么想加入机构？",
      "你最崇拜哪位资深特工？",
      "你最害怕在工作中犯什么错误？"
    ]
  }
}
```

---

## ARC系统的交互

### 1. 异常能力 ↔ 资质保证

```go
// 使用异常能力时消耗QA
func (agent *Agent) UseAbility(ability Ability) RollResult {
    // 1. 掷骰
    roll := agent.RollDice(ability.Roll.Quality)
    
    // 2. 玩家可以花费QA调整结果
    if agent.WantsToSpendQA() {
        quality := ability.Roll.Quality
        spent := agent.SpendQA(quality, amount)
        roll.AdjustWithQA(spent)
    }
    
    // 3. 应用过载（如果该资质QA为0）
    if agent.QA[ability.Roll.Quality] == 0 {
        roll.ApplyOverload(1)
    }
    
    return roll
}
```

### 2. 现实触发器 ↔ 人际关系

```go
// 现实触发器影响人际关系
func (gm *GM) TriggerReality(agent *Agent) {
    trigger := agent.Reality.Trigger
    
    // 消耗混沌
    if gm.ChaosPool >= trigger.Cost {
        gm.ChaosPool -= trigger.Cost
        
        // 应用效果
        gm.ApplyTriggerEffect(agent, trigger)
        
        // 如果玩家忽视
        if !agent.AddressedTrigger {
            // 影响最弱的人际关系
            weakest := agent.GetWeakestRelationship()
            weakest.LoseConnection(1)
        }
    }
}
```

### 3. 职能 ↔ 嘉奖/申诫

```go
// 职能决定哪些行为获得奖励
func (gm *GM) EvaluateAction(agent *Agent, action Action) {
    // 检查许可行为
    reward := agent.CheckPermittedBehavior(action)
    if reward > 0 {
        agent.AddCommendations(reward)
    }
    
    // 检查首要指令
    if agent.Career.PrimeDirective.IsViolated(action) {
        agent.AddReprimands(agent.Career.PrimeDirective.Violation)
    }
}
```

### 4. 完整的角色状态

```go
type Agent struct {
    // 基本信息
    ID   string
    Name string
    
    // ARC组件
    Anomaly  Anomaly
    Reality  Reality
    Career   Career
    
    // 资质保证
    QA map[string]int
    
    // 人际关系
    Relationships []*Relationship
    
    // 绩效
    Commendations int
    Reprimands    int
    Rating        string
    
    // 状态
    Alive         bool
    InDebt        bool
}
```

---

## 数据存储格式

### 角色存档结构

```json
{
  "agent": {
    "id": "agent-001",
    "name": "张伟",
    "anomaly": {
      "type": "低语",
      "abilities": [
        "再说一遍？",
        "话到嘴边",
        "静默"
      ]
    },
    "reality": {
      "type": "看护者",
      "specialFeature": {
        "dependent": "婴儿",
        "name": "小明"
      },
      "degradationTrack": {
        "name": "独立",
        "filled": 1,
        "total": 4
      },
      "relationships": [
        {
          "id": "rel-001",
          "name": "李娜",
          "description": "妻子，小明的母亲",
          "connection": 6,
          "playedBy": "GM"
        },
        {
          "id": "rel-002",
          "name": "王强",
          "description": "大学朋友",
          "connection": 3,
          "playedBy": "player-2"
        },
        {
          "id": "rel-003",
          "name": "陈医生",
          "description": "小明的儿科医生",
          "connection": 3,
          "playedBy": "GM"
        }
      ]
    },
    "career": {
      "type": "公关",
      "qa": {
        "专注": 1,
        "共情": 2,
        "气场": 2,
        "欺瞒": 2,
        "主动": 1,
        "专业": 1,
        "活力": 0,
        "坚毅": 0,
        "诡秘": 0
      },
      "claimables": [
        "公关应急包"
      ]
    },
    "performance": {
      "commendations": 15,
      "reprimands": 2,
      "rating": "评级良好"
    },
    "status": {
      "alive": true,
      "inDebt": false
    }
  }
}
```

## 总结

ARC系统的三个维度各有侧重：

- **异常 (A)**: 提供**战术能力**，通过掷骰和QA消耗实现
- **现实 (R)**: 提供**叙事深度**，通过人际关系和触发器实现
- **职能 (C)**: 提供**角色定位**，通过奖惩机制和QA分配实现

三者相互交织，共同构成一个完整的角色系统。
