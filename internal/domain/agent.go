package domain

import "time"

// Agent 外勤特工
type Agent struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Pronouns      string          `json:"pronouns"`
	Anomaly       *Anomaly        `json:"anomaly"`
	Reality       *Reality        `json:"reality"`
	Career        *Career         `json:"career"`
	QA            map[string]int  `json:"qa"`
	Relationships []*Relationship `json:"relationships"`
	Commendations int             `json:"commendations"`
	Reprimands    int             `json:"reprimands"`
	Rating        string          `json:"rating"`
	Alive         bool            `json:"alive"`
	InDebt        bool            `json:"in_debt"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Anomaly 异常体
type Anomaly struct {
	Type      string            `json:"type"`
	Abilities []*AnomalyAbility `json:"abilities"`
}

// AnomalyAbility 异常能力
type AnomalyAbility struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	AnomalyType string          `json:"anomaly_type"`
	Trigger     *AbilityTrigger `json:"trigger"`
	Roll        *AbilityRoll    `json:"roll"`
	Effects     *AbilityEffects `json:"effects"`
}

// TriggerType 触发类型
type TriggerType string

const (
	TriggerAction   TriggerType = "action"   // 主动使用
	TriggerResponse TriggerType = "response" // 响应某事
	TriggerPassive  TriggerType = "passive"  // 被动触发
	TriggerReactive TriggerType = "reactive" // 反应性触发
)

// AbilityTrigger 能力触发器
type AbilityTrigger struct {
	Type        TriggerType `json:"type"`
	Description string      `json:"description"`
	Condition   string      `json:"condition,omitempty"`
}

// AbilityRoll 能力掷骰
type AbilityRoll struct {
	Quality   string `json:"quality"`    // 相关资质
	DiceCount int    `json:"dice_count"` // 骰子数量
	DiceType  int    `json:"dice_type"`  // 骰子类型
}

// AbilityEffects 能力效果
type AbilityEffects struct {
	Success    *Effect              `json:"success"`
	Failure    *Effect              `json:"failure"`
	Additional []*ConditionalEffect `json:"additional,omitempty"`
}

// Effect 效果
type Effect struct {
	Description string `json:"description"`
	Mechanics   string `json:"mechanics"`
	Duration    string `json:"duration,omitempty"`
	Target      string `json:"target,omitempty"`
}

// ConditionalEffect 条件效果
type ConditionalEffect struct {
	Condition string  `json:"condition"` // "每额外一个3", "六个或更多3"
	Effect    *Effect `json:"effect"`
}

// Reality 现实
type Reality struct {
	Type             string                 `json:"type"`
	SpecialFeature   map[string]interface{} `json:"special_feature"`
	Trigger          *RealityTrigger        `json:"trigger"`
	OverloadRelief   *OverloadRelief        `json:"overload_relief"`
	DegradationTrack *DegradationTrack      `json:"degradation_track"`
	Relationships    []*Relationship        `json:"relationships"`
}

// RealityTrigger 现实触发器
type RealityTrigger struct {
	Name        string `json:"name"`
	Cost        int    `json:"cost"`        // 混沌消耗
	Effect      string `json:"effect"`      // 触发效果
	Consequence string `json:"consequence"` // 忽视后果
}

// OverloadRelief 过载解除
type OverloadRelief struct {
	Name      string `json:"name"`
	Condition string `json:"condition"` // 激活条件
	Effect    string `json:"effect"`    // 通常是"无视所有过载"
}

// Career 职能
type Career struct {
	Type               string               `json:"type"`
	QA                 map[string]int       `json:"qa"`
	PermittedBehaviors []*PermittedBehavior `json:"permitted_behaviors"`
	PrimeDirective     *PrimeDirective      `json:"prime_directive"`
	Claimables         []string             `json:"claimables"`
}

// PermittedBehavior 许可行为
type PermittedBehavior struct {
	Action    string `json:"action"`
	Reward    int    `json:"reward"`    // 嘉奖数量
	Condition string `json:"condition"` // 可选的额外条件
}

// PrimeDirective 首要指令
type PrimeDirective struct {
	Description string `json:"description"`
	Violation   int    `json:"violation"` // 违反时的申诫数量
}

// Relationship 人际关系
type Relationship struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Connection  int      `json:"connection"`
	PlayedBy    string   `json:"played_by"`
	Notes       []string `json:"notes"`
}

// DegradationTrack 退化轨道
type DegradationTrack struct {
	Name   string `json:"name"`
	Filled int    `json:"filled"`
	Total  int    `json:"total"`
}

// Quality 资质常量
const (
	QualityFocus      = "专注" // 关注细节、发现隐藏
	QualityEmpathy    = "共情" // 建立联系、发现弱点
	QualityPresence   = "气场" // 脱颖而出、鼓舞人心
	QualityDeception  = "欺瞒" // 说谎、说服
	QualityInitiative = "主动" // 前瞻思考、迅速行动
	QualityProfession = "专业" // 保持镇定、抵抗分心
	QualityVitality   = "活力" // 进攻、使用武力
	QualityGrit       = "坚毅" // 拒绝退缩、施加压力
	QualitySubtlety   = "诡秘" // 悄无声息、避免注意
)

// AllQualities 所有资质列表
var AllQualities = []string{
	QualityFocus,
	QualityEmpathy,
	QualityPresence,
	QualityDeception,
	QualityInitiative,
	QualityProfession,
	QualityVitality,
	QualityGrit,
	QualitySubtlety,
}

// AnomalyType 异常体类型常量
const (
	AnomalyWhisper   = "低语"
	AnomalyCatalog   = "目录"
	AnomalySiphon    = "汲取"
	AnomalyTimepiece = "时计"
	AnomalyGrowth    = "生长"
	AnomalyGun       = "枪械"
	AnomalyDream     = "梦境"
	AnomalyManifold  = "流形"
	AnomalyAbsence   = "缺位"
)

// AllAnomalyTypes 所有异常体类型
var AllAnomalyTypes = []string{
	AnomalyWhisper,
	AnomalyCatalog,
	AnomalySiphon,
	AnomalyTimepiece,
	AnomalyGrowth,
	AnomalyGun,
	AnomalyDream,
	AnomalyManifold,
	AnomalyAbsence,
}

// RealityType 现实类型常量
const (
	RealityCaretaker        = "看护者"
	RealityScheduleOverload = "日程过载"
	RealityHunted           = "受追猎者"
	RealityStar             = "明星"
	RealityStruggling       = "挣扎求生"
	RealityNewborn          = "新生儿"
	RealityRomantic         = "浪漫主义"
	RealityPillar           = "支柱"
	RealityOutsider         = "异类"
)

// AllRealityTypes 所有现实类型
var AllRealityTypes = []string{
	RealityCaretaker,
	RealityScheduleOverload,
	RealityHunted,
	RealityStar,
	RealityStruggling,
	RealityNewborn,
	RealityRomantic,
	RealityPillar,
	RealityOutsider,
}

// CareerType 职能类型常量
const (
	CareerPublicRelations = "公关"
	CareerRD              = "研发"
	CareerBarista         = "咖啡师"
	CareerCEO             = "CEO"
	CareerIntern          = "实习生"
	CareerGravedigger     = "掘墓人"
	CareerReception       = "接待处"
	CareerHotline         = "热线"
	CareerClown           = "小丑"
)

// AllCareerTypes 所有职能类型
var AllCareerTypes = []string{
	CareerPublicRelations,
	CareerRD,
	CareerBarista,
	CareerCEO,
	CareerIntern,
	CareerGravedigger,
	CareerReception,
	CareerHotline,
	CareerClown,
}

// AgentRating 机构评级常量
const (
	RatingExcellent    = "评级良好"  // 0申诫
	RatingNeedsWork    = "有待改进"  // 1申诫
	RatingProbation    = "留职察看"  // 2-3申诫
	RatingFinalWarning = "最后警告"  // 4-9申诫
	RatingRevoked      = "权限已撤销" // 10+申诫
)

// GetRating 根据申诫数量获取评级
func GetRating(reprimands int) string {
	switch {
	case reprimands == 0:
		return RatingExcellent
	case reprimands == 1:
		return RatingNeedsWork
	case reprimands >= 2 && reprimands <= 3:
		return RatingProbation
	case reprimands >= 4 && reprimands <= 9:
		return RatingFinalWarning
	default:
		return RatingRevoked
	}
}

// SpendQA 花费资质保证
func (a *Agent) SpendQA(quality string, amount int) error {
	if a.QA[quality] < amount {
		return NewGameError(ErrInsufficientQA, "资质保证不足").
			WithDetails("quality", quality).
			WithDetails("available", a.QA[quality]).
			WithDetails("required", amount)
	}

	a.QA[quality] -= amount
	return nil
}

// RestoreQA 恢复资质保证（任务间隙）
func (a *Agent) RestoreQA() {
	// 根据职能恢复QA到初始值
	// 确保所有资质都被恢复
	for _, quality := range AllQualities {
		if initial, exists := a.Career.QA[quality]; exists {
			a.QA[quality] = initial
		} else {
			a.QA[quality] = 0
		}
	}
}

// AddCommendations 添加嘉奖
func (a *Agent) AddCommendations(amount int) {
	a.Commendations += amount
}

// AddReprimands 添加申诫并更新评级
func (a *Agent) AddReprimands(amount int) {
	a.Reprimands += amount
	a.Rating = GetRating(a.Reprimands)
}

// GetWeakestRelationship 获取连结最低的人际关系
func (a *Agent) GetWeakestRelationship() *Relationship {
	if len(a.Relationships) == 0 {
		return nil
	}

	weakest := a.Relationships[0]
	for _, rel := range a.Relationships[1:] {
		if rel.Connection < weakest.Connection {
			weakest = rel
		}
	}
	return weakest
}

// TotalQA 计算总资质保证点数
func (a *Agent) TotalQA() int {
	total := 0
	for _, qa := range a.QA {
		total += qa
	}
	return total
}

// TotalConnection 计算总连结点数
func (a *Agent) TotalConnection() int {
	total := 0
	for _, rel := range a.Relationships {
		total += rel.Connection
	}
	return total
}

// ValidateARC 验证ARC组合的有效性
func (a *Agent) ValidateARC() error {
	// 验证异常体类型
	validAnomaly := false
	for _, t := range AllAnomalyTypes {
		if a.Anomaly.Type == t {
			validAnomaly = true
			break
		}
	}
	if !validAnomaly {
		return NewGameError(ErrInvalidARC, "无效的异常体类型").
			WithDetails("type", a.Anomaly.Type)
	}

	// 验证现实类型
	validReality := false
	for _, t := range AllRealityTypes {
		if a.Reality.Type == t {
			validReality = true
			break
		}
	}
	if !validReality {
		return NewGameError(ErrInvalidARC, "无效的现实类型").
			WithDetails("type", a.Reality.Type)
	}

	// 验证职能类型
	validCareer := false
	for _, t := range AllCareerTypes {
		if a.Career.Type == t {
			validCareer = true
			break
		}
	}
	if !validCareer {
		return NewGameError(ErrInvalidARC, "无效的职能类型").
			WithDetails("type", a.Career.Type)
	}

	// 验证异常能力数量
	if len(a.Anomaly.Abilities) != 3 {
		return NewGameError(ErrInvalidARC, "异常能力必须为3个").
			WithDetails("count", len(a.Anomaly.Abilities))
	}

	// 验证人际关系数量
	if len(a.Relationships) != 3 {
		return NewGameError(ErrInvalidARC, "人际关系必须为3段").
			WithDetails("count", len(a.Relationships))
	}

	// 验证总连结点数
	totalConnection := a.TotalConnection()
	if totalConnection != 12 {
		return NewGameError(ErrInvalidARC, "总连结点数必须为12").
			WithDetails("total", totalConnection)
	}

	// 验证总QA点数
	totalQA := a.TotalQA()
	if totalQA > 9 {
		return NewGameError(ErrInvalidARC, "总资质保证不能超过9点").
			WithDetails("total", totalQA)
	}

	return nil
}
