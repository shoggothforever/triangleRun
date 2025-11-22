# ARC系统配置文件

本目录包含《三角机构》TRPG系统的ARC（异常-现实-职能）配置数据。

## 文件说明

### anomalies.json
包含9种异常体类型的完整定义：
- 低语 (Whisper) - 操控声音和语言
- 目录 (Catalog) - 创造和复制物体
- 汲取 (Siphon) - 吸取和转移特质
- 时计 (Timepiece) - 操控时间流动
- 生长 (Growth) - 改变身体和感知
- 枪械 (Gun) - 终结和抹除
- 梦境 (Dream) - 创造幻象和进入艺术
- 流形 (Manifold) - 操控空间和重力
- 缺位 (Absence) - 消失和遗忘

每个异常体包含：
- 3个获批能力
- 每个能力的触发条件、掷骰机制、成功/失败效果

### realities.json
包含9种现实类型的完整定义：
- 看护者 (Caretaker) - 照顾受照料者
- 日程过载 (Schedule Overload) - 兼顾多份工作
- 受追猎者 (Hunted) - 躲避过去
- 明星 (Star) - 公众人物
- 挣扎求生 (Struggling) - 经济困难
- 新生儿 (Newborn) - 初来乍到
- 浪漫主义 (Romantic) - 情感复杂
- 支柱 (Pillar) - 社区依靠
- 异类 (Outsider) - 格格不入

每个现实包含：
- 特殊特性
- 现实触发器
- 过载解除条件
- 退化轨道（4格）
- 人际关系配置（3段，总计12点连结）

### careers.json
包含9种职能类型的完整定义：
- 公关 (Public Relations) - 管理形象
- 研发 (R&D) - 研究异常
- 咖啡师 (Barista) - 提供服务
- CEO - 享受特权
- 实习生 (Intern) - 学习成长
- 掘墓人 (Gravedigger) - 处理后果
- 接待处 (Reception) - 协调沟通
- 热线 (Hotline) - 提供支持
- 小丑 (Clown) - 娱乐士气

每个职能包含：
- 初始资质保证（QA）分配（总计9点）
- 许可行为及奖励
- 首要指令及违反惩罚
- 初始申领物
- 评估问题

## 使用方法

### 加载配置

```go
import (
    "encoding/json"
    "os"
)

// 加载异常体配置
func LoadAnomalies() (*AnomaliesConfig, error) {
    data, err := os.ReadFile("configs/anomalies.json")
    if err != nil {
        return nil, err
    }
    
    var config AnomaliesConfig
    err = json.Unmarshal(data, &config)
    return &config, err
}

// 加载现实配置
func LoadRealities() (*RealitiesConfig, error) {
    data, err := os.ReadFile("configs/realities.json")
    if err != nil {
        return nil, err
    }
    
    var config RealitiesConfig
    err = json.Unmarshal(data, &config)
    return &config, err
}

// 加载职能配置
func LoadCareers() (*CareersConfig, error) {
    data, err := os.ReadFile("configs/careers.json")
    if err != nil {
        return nil, err
    }
    
    var config CareersConfig
    err = json.Unmarshal(data, &config)
    return &config, err
}
```

### 创建角色

```go
// 根据配置创建角色
func CreateAgentFromConfig(anomalyType, realityType, careerType string) (*domain.Agent, error) {
    // 1. 加载配置
    anomalies, _ := LoadAnomalies()
    realities, _ := LoadRealities()
    careers, _ := LoadCareers()
    
    // 2. 查找对应的配置
    var anomalyConfig *AnomalyConfig
    for _, a := range anomalies.Anomalies {
        if a.Name == anomalyType {
            anomalyConfig = &a
            break
        }
    }
    
    var realityConfig *RealityConfig
    for _, r := range realities.Realities {
        if r.Name == realityType {
            realityConfig = &r
            break
        }
    }
    
    var careerConfig *CareerConfig
    for _, c := range careers.Careers {
        if c.Name == careerType {
            careerConfig = &c
            break
        }
    }
    
    // 3. 创建角色
    agent := &domain.Agent{
        Anomaly: &domain.Anomaly{
            Type:      anomalyConfig.Name,
            Abilities: anomalyConfig.Abilities,
        },
        Reality: &domain.Reality{
            Type:             realityConfig.Name,
            SpecialFeature:   realityConfig.SpecialFeature,
            Trigger:          realityConfig.Trigger,
            OverloadRelief:   realityConfig.OverloadRelief,
            DegradationTrack: &domain.DegradationTrack{
                Name:   realityConfig.DegradationTrack.Name,
                Filled: 0,
                Total:  realityConfig.DegradationTrack.Boxes,
            },
        },
        Career: &domain.Career{
            Type:               careerConfig.Name,
            QA:                 careerConfig.InitialQA.Distribution,
            PermittedBehaviors: careerConfig.PermittedBehaviors,
            PrimeDirective:     careerConfig.PrimeDirective,
        },
        QA: make(map[string]int),
    }
    
    // 4. 复制QA分配
    for quality, amount := range careerConfig.InitialQA.Distribution {
        agent.QA[quality] = amount
    }
    
    return agent, nil
}
```

## 测试

运行测试以验证配置完整性：

```bash
go test ./configs/...
```

测试包括：
- 配置文件加载测试
- 数据完整性验证
- 类型匹配验证
- 资质有效性验证

## 配置验证规则

### 异常体配置
- ✓ 必须有9种异常体
- ✓ 每个异常体必须有3个能力
- ✓ 每个能力必须有触发器、掷骰配置和效果
- ✓ 掷骰必须是6d4
- ✓ 使用的资质必须是有效的9种资质之一

### 现实配置
- ✓ 必须有9种现实
- ✓ 每个现实必须有触发器、过载解除和退化轨道
- ✓ 人际关系必须是3段，总连结12点

### 职能配置
- ✓ 必须有9种职能
- ✓ 每个职能的QA总和必须是9点
- ✓ 必须定义所有9种资质的QA分配
- ✓ 必须有许可行为和首要指令

## 扩展配置

如需添加新的异常体、现实或职能：

1. 在对应的JSON文件中添加新条目
2. 确保遵循现有的数据结构
3. 运行测试验证配置有效性
4. 更新domain包中的类型常量列表

## 相关文档

- [ARC系统详细结构](.kiro/specs/trpg-solo-engine/arc-system.md)
- [需求文档](.kiro/specs/trpg-solo-engine/requirements.md)
- [设计文档](.kiro/specs/trpg-solo-engine/design.md)
