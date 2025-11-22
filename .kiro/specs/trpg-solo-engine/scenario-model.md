# 三角机构剧本模型

## 概述

基于《永恒之泉》等官方模组的分析，本文档定义了一个通用的剧本结构模型，适用于所有《三角机构》异常体回收任务。

## 剧本核心结构

### 1. 异常体档案 (Anomaly Profile)

每个剧本的核心是一个异常体，包含以下属性：

```json
{
  "anomaly": {
    "id": "fountain-of-youth",
    "name": "永恒之泉",
    "history": "异常体的起源和发展历史",
    "focus": {
      "emotion": "渴望",
      "subject": "回到某个失去的时光或自我"
    },
    "domain": {
      "location": "三联城下水道中的雨林",
      "description": "领域的详细描述"
    },
    "appearance": "异常体的外观描述",
    "impulse": "异常体的驱动目标",
    "currentStatus": "当前状况和威胁等级",
    "chaosEffects": [
      {
        "cost": 2,
        "name": "焕新",
        "description": "改变目标外貌"
      }
    ]
  }
}
```

### 2. 任务前夕 (Pre-Mission)

#### 2.1 晨会场景 (Morning Scenes)
- 4-5个可选的日常场景，为任务埋下伏笔
- 每个场景关联一名特工或其重要之人

#### 2.2 任务简报 (Mission Briefing)
- 异常体的初步情报
- 调查起点和已知线索
- 机构的特殊指示

#### 2.3 可选目标 (Optional Objectives)
- 额外的嘉奖/申诫条件
- 与剧本主题相关的挑战

### 3. 调查阶段 (Investigation Phase)

调查阶段是剧本的核心，由多个场景组成：

```json
{
  "investigation": {
    "scenes": [
      {
        "id": "commercial-avenue",
        "name": "商业大道",
        "description": "场景的整体描述",
        "locations": [
          {
            "id": "tryptik-cafe",
            "name": "Tryptik咖啡馆",
            "description": "地点描述",
            "clues": [
              {
                "id": "clue-001",
                "type": "mundane",  // mundane 或 anomalous
                "content": "线索内容",
                "requirements": "获取条件",
                "leadsTo": ["location-id", "npc-id"]
              }
            ],
            "npcs": ["maya-ng", "barista"],
            "interactables": [
              {
                "id": "water-puddle",
                "name": "水洼",
                "description": "描述",
                "interaction": "触发的效果或事件"
              }
            ]
          }
        ],
        "connections": ["the-source", "laundrocade"]
      }
    ]
  }
}
```

### 4. NPC系统 (NPC System)

```json
{
  "npcs": [
    {
      "id": "serena-evermore",
      "name": "Serena Evermore",
      "pronouns": "她/她的",
      "role": "水疗协调员",
      "description": "外貌和性格描述",
      "location": "the-source-reception",
      "knowledge": [
        {
          "info": "她知道的信息",
          "revealCondition": "揭示条件（对话、能力、事件）"
        }
      ],
      "dialogue": [
        {
          "trigger": "first-meeting",
          "content": "对话内容"
        }
      ],
      "state": {
        "affected": false,  // 是否被异常体影响
        "relationship": 0,  // 与玩家的关系值
        "alive": true
      }
    }
  ]
}
```

### 5. 线索系统 (Clue System)

线索分为两类：
- **凡俗线索** (Mundane Clues): 通过普通调查获得
- **异常线索** (Anomalous Clues): 需要异常能力或请求机构

```json
{
  "clues": [
    {
      "id": "clue-001",
      "name": "奥可菲产品副作用",
      "type": "mundane",
      "content": "使用者外貌变得相似",
      "source": "maya-ng",
      "category": "anomaly-effect",
      "unlocks": ["the-source-interior"],
      "required": false  // 是否为必要线索
    }
  ]
}
```

### 6. 事件系统 (Event System)

事件可以由玩家行动、时间推移或混沌效应触发：

```json
{
  "events": [
    {
      "id": "event-001",
      "name": "Maya被水洼捕获",
      "trigger": {
        "type": "proximity",  // proximity, time, chaos, player-action
        "condition": "玩家接近水洼"
      },
      "effects": [
        {
          "type": "create-threat",
          "target": "maya-ng",
          "description": "Maya消失在水洼中"
        },
        {
          "type": "散逸端-risk",
          "value": 15,
          "condition": "如果不阻止水洼"
        }
      ],
      "resolution": [
        {
          "action": "救出Maya",
          "result": "获得关键线索"
        }
      ]
    }
  ]
}
```

### 7. 遭遇阶段 (Encounter Phase)

最终对抗异常体的阶段：

```json
{
  "encounter": {
    "location": "underground-rainforest",
    "description": "遭遇场景描述",
    "phases": [
      {
        "name": "进入领域",
        "description": "特工进入异常体领域",
        "challenges": ["藤蔓阻挡", "次级异常体"],
        "chaosThreshold": 10
      },
      {
        "name": "与异常体交涉",
        "description": "异常体提出交易或威胁",
        "options": [
          {
            "action": "说服异常体",
            "requirements": "理解焦点",
            "outcome": "peaceful-capture"
          },
          {
            "action": "战斗",
            "outcome": "combat"
          }
        ]
      },
      {
        "name": "Serena的忏悔",
        "description": "关键NPC介入",
        "trigger": "特工威胁异常体",
        "resolution": "需要调解冲突"
      }
    ],
    "captureConditions": [
      "异常体被安抚",
      "异常体筋疲力尽",
      "异常体自愿"
    ]
  }
}
```

### 8. 余波 (Aftermath)

任务结束后的后续影响：

```json
{
  "aftermath": {
    "captured": {
      "description": "异常体被捕获后的影响",
      "npcFates": {
        "serena-evermore": "可能成为共鸣者",
        "源泉员工": "需要处理散逸端"
      },
      "worldChanges": "奥可菲产品失效"
    },
    "neutralized": {
      "description": "异常体被中和后的影响"
    },
    "escaped": {
      "description": "异常体逃脱后的影响",
      "futureThreats": "可能的后续威胁"
    }
  }
}
```

### 9. 申领物和异常能力 (Rewards)

任务完成后可获得的奖励：

```json
{
  "rewards": {
    "claimables": [
      {
        "id": "reunion-rsvp",
        "name": "同学会回执",
        "cost": 3,
        "description": "效果描述",
        "usage": "使用方法"
      }
    ],
    "anomalousAbilities": [
      {
        "id": "harvest-time",
        "name": "收获之时",
        "anomalyType": "生长",
        "description": "能力描述",
        "mechanics": "游戏机制"
      }
    ]
  }
}
```

## 剧本执行流程

### 阶段1: 初始化
1. 加载异常体档案
2. 设置初始混沌池（基于累积散逸端）
3. 呈现晨会场景
4. 播放任务简报

### 阶段2: 调查循环
```
while (未找到异常体领域) {
  1. 呈现当前场景描述
  2. 列出可用行动（移动、对话、调查、使用能力）
  3. 处理玩家行动
  4. 更新场景状态
  5. 检查事件触发条件
  6. 处理混沌效应
  7. 追踪散逸端
  8. 更新线索日志
}
```

### 阶段3: 遭遇
1. 进入异常体领域
2. 执行遭遇阶段
3. 处理玩家与异常体的互动
4. 判定任务结果（已捕获/已中和/已逃脱）

### 阶段4: 结算
1. 计算嘉奖和申诫
2. 生成任务报告
3. 应用余波效果
4. 解锁奖励

## 数据结构设计原则

1. **模块化**: 每个组件（场景、NPC、线索）独立定义，通过ID关联
2. **状态追踪**: 所有可变元素都有状态字段
3. **条件系统**: 使用统一的条件表达式系统
4. **扩展性**: 支持自定义混沌效应和事件类型
5. **可复用**: 通用组件（如次级异常体类型）可在多个剧本间共享

## 示例：完整场景定义

```json
{
  "scene": {
    "id": "the-source-interior",
    "name": "源泉内部",
    "description": "踏入源泉，就像走进了一张商业图库照片...",
    "requiredClues": ["clue-001"],
    "locations": [
      {
        "id": "reception",
        "name": "接待处",
        "description": "Serena坐在半圆形办公桌后...",
        "npcs": ["serena-evermore"],
        "interactables": [
          {
            "id": "serena-tablet",
            "name": "Serena的平板电脑",
            "type": "device",
            "clues": ["serena-schedule", "product-dev-messages"],
            "requirements": {
              "or": [
                {"type": "skill", "skill": "专注"},
                {"type": "ability", "category": "anomalous"}
              ]
            }
          }
        ],
        "exits": ["medical-aesthetics", "spa-area"]
      }
    ],
    "events": [
      {
        "id": "crowd-enters",
        "trigger": {
          "type": "chaos",
          "effect": "吸引",
          "location": "outside"
        },
        "description": "人群涌入室内",
        "effects": ["serena-distracted"]
      }
    ]
  }
}
```

## 实现建议

### 后端架构
- **剧本加载器**: 解析JSON格式的剧本文件
- **状态管理器**: 追踪场景、NPC、线索状态
- **事件引擎**: 处理触发条件和效果
- **对话系统**: 管理NPC对话树
- **混沌管理器**: 处理混沌池和效应

### 前端展示
- **场景渲染**: 动态生成场景描述和可用行动
- **线索日志**: 可视化已收集的线索和关联
- **地图系统**: 显示已探索的地点和连接
- **NPC追踪**: 记录遇到的NPC和对话历史

### AI集成
- **场景描述生成**: 基于模板和状态生成动态描述
- **NPC对话**: 根据性格和状态生成自然对话
- **GM决策**: 处理玩家的创造性行动
- **混沌效应创意**: 生成符合异常体特性的效应
