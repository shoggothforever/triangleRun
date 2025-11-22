package configs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// AnomaliesConfig 异常体配置结构
type AnomaliesConfig struct {
	Anomalies []AnomalyConfig `json:"anomalies"`
}

// AnomalyConfig 单个异常体配置
type AnomalyConfig struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Focus       map[string]string        `json:"focus"`
	Abilities   []*domain.AnomalyAbility `json:"abilities"`
}

// RealitiesConfig 现实配置结构
type RealitiesConfig struct {
	Realities []RealityConfig `json:"realities"`
}

// RealityConfig 单个现实配置
type RealityConfig struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	SpecialFeature   map[string]interface{} `json:"special_feature"`
	Trigger          *domain.RealityTrigger `json:"trigger"`
	OverloadRelief   *domain.OverloadRelief `json:"overload_relief"`
	DegradationTrack DegradationTrackConfig `json:"degradation_track"`
	Relationships    RelationshipsConfig    `json:"relationships"`
}

// DegradationTrackConfig 退化轨道配置
type DegradationTrackConfig struct {
	Name        string `json:"name"`
	Boxes       int    `json:"boxes"`
	Trigger     string `json:"trigger"`
	Consequence string `json:"consequence"`
}

// RelationshipsConfig 人际关系配置
type RelationshipsConfig struct {
	Count                 int      `json:"count"`
	TotalConnection       int      `json:"total_connection"`
	SuggestedDistribution []int    `json:"suggested_distribution"`
	Questions             []string `json:"questions"`
}

// CareersConfig 职能配置结构
type CareersConfig struct {
	Careers []CareerConfig `json:"careers"`
}

// CareerConfig 单个职能配置
type CareerConfig struct {
	ID                  string                      `json:"id"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	InitialQA           InitialQAConfig             `json:"initial_qa"`
	PermittedBehaviors  []*domain.PermittedBehavior `json:"permitted_behaviors"`
	PrimeDirective      *domain.PrimeDirective      `json:"prime_directive"`
	InitialClaimable    map[string]string           `json:"initial_claimable"`
	AssessmentQuestions []string                    `json:"assessment_questions"`
}

// InitialQAConfig 初始资质保证配置
type InitialQAConfig struct {
	Total        int            `json:"total"`
	Distribution map[string]int `json:"distribution"`
	Note         string         `json:"note,omitempty"`
}

// TestAnomaliesConfigLoad 测试异常体配置加载
func TestAnomaliesConfigLoad(t *testing.T) {
	data, err := os.ReadFile("anomalies.json")
	require.NoError(t, err, "应该能够读取异常体配置文件")

	var config AnomaliesConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err, "应该能够解析异常体配置")

	// 验证有9种异常体
	assert.Len(t, config.Anomalies, 9, "应该有9种异常体")

	// 验证每个异常体的完整性
	for _, anomaly := range config.Anomalies {
		t.Run(anomaly.Name, func(t *testing.T) {
			assert.NotEmpty(t, anomaly.ID, "异常体ID不能为空")
			assert.NotEmpty(t, anomaly.Name, "异常体名称不能为空")
			assert.NotEmpty(t, anomaly.Description, "异常体描述不能为空")

			// 验证焦点
			assert.NotEmpty(t, anomaly.Focus, "异常体应该有焦点")

			// 验证有3个能力
			assert.Len(t, anomaly.Abilities, 3, "每个异常体应该有3个能力")

			// 验证每个能力的完整性
			for _, ability := range anomaly.Abilities {
				assert.NotEmpty(t, ability.ID, "能力ID不能为空")
				assert.NotEmpty(t, ability.Name, "能力名称不能为空")
				assert.Equal(t, anomaly.Name, ability.AnomalyType, "能力类型应该匹配异常体")

				// 验证触发器
				assert.NotNil(t, ability.Trigger, "能力应该有触发器")
				assert.NotEmpty(t, ability.Trigger.Description, "触发器应该有描述")

				// 验证掷骰
				assert.NotNil(t, ability.Roll, "能力应该有掷骰配置")
				assert.NotEmpty(t, ability.Roll.Quality, "掷骰应该指定资质")
				assert.Equal(t, 6, ability.Roll.DiceCount, "应该掷6颗骰子")
				assert.Equal(t, 4, ability.Roll.DiceType, "应该是d4骰子")

				// 验证效果
				assert.NotNil(t, ability.Effects, "能力应该有效果")
				assert.NotNil(t, ability.Effects.Success, "应该有成功效果")
				assert.NotNil(t, ability.Effects.Failure, "应该有失败效果")
			}
		})
	}
}

// TestRealitiesConfigLoad 测试现实配置加载
func TestRealitiesConfigLoad(t *testing.T) {
	data, err := os.ReadFile("realities.json")
	require.NoError(t, err, "应该能够读取现实配置文件")

	var config RealitiesConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err, "应该能够解析现实配置")

	// 验证有9种现实
	assert.Len(t, config.Realities, 9, "应该有9种现实")

	// 验证每个现实的完整性
	for _, reality := range config.Realities {
		t.Run(reality.Name, func(t *testing.T) {
			assert.NotEmpty(t, reality.ID, "现实ID不能为空")
			assert.NotEmpty(t, reality.Name, "现实名称不能为空")
			assert.NotEmpty(t, reality.Description, "现实描述不能为空")

			// 验证特殊特性
			assert.NotEmpty(t, reality.SpecialFeature, "现实应该有特殊特性")

			// 验证触发器
			assert.NotNil(t, reality.Trigger, "现实应该有触发器")
			assert.NotEmpty(t, reality.Trigger.Name, "触发器应该有名称")
			assert.NotEmpty(t, reality.Trigger.Effect, "触发器应该有效果")

			// 验证过载解除
			assert.NotNil(t, reality.OverloadRelief, "现实应该有过载解除")
			assert.NotEmpty(t, reality.OverloadRelief.Name, "过载解除应该有名称")
			assert.NotEmpty(t, reality.OverloadRelief.Condition, "过载解除应该有条件")

			// 验证退化轨道
			assert.NotEmpty(t, reality.DegradationTrack.Name, "退化轨道应该有名称")
			assert.Greater(t, reality.DegradationTrack.Boxes, 0, "退化轨道应该有格子")

			// 验证人际关系配置
			assert.Equal(t, 3, reality.Relationships.Count, "应该有3段人际关系")
			assert.Equal(t, 12, reality.Relationships.TotalConnection, "总连结应该是12点")
			assert.Len(t, reality.Relationships.Questions, 3, "应该有3个人际关系问题")
		})
	}
}

// TestCareersConfigLoad 测试职能配置加载
func TestCareersConfigLoad(t *testing.T) {
	data, err := os.ReadFile("careers.json")
	require.NoError(t, err, "应该能够读取职能配置文件")

	var config CareersConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err, "应该能够解析职能配置")

	// 验证有9种职能
	assert.Len(t, config.Careers, 9, "应该有9种职能")

	// 验证每个职能的完整性
	for _, career := range config.Careers {
		t.Run(career.Name, func(t *testing.T) {
			assert.NotEmpty(t, career.ID, "职能ID不能为空")
			assert.NotEmpty(t, career.Name, "职能名称不能为空")
			assert.NotEmpty(t, career.Description, "职能描述不能为空")

			// 验证初始QA
			assert.Equal(t, 9, career.InitialQA.Total, "总QA应该是9点")
			assert.NotEmpty(t, career.InitialQA.Distribution, "应该有QA分配")

			// 验证QA分配总和
			totalQA := 0
			for quality, amount := range career.InitialQA.Distribution {
				assert.Contains(t, domain.AllQualities, quality, "资质应该是有效的")
				assert.GreaterOrEqual(t, amount, 0, "QA点数不能为负")
				totalQA += amount
			}
			assert.Equal(t, 9, totalQA, "QA分配总和应该是9点")

			// 验证所有9种资质都有定义
			assert.Len(t, career.InitialQA.Distribution, 9, "应该定义所有9种资质")

			// 验证许可行为
			assert.NotEmpty(t, career.PermittedBehaviors, "应该有许可行为")
			for _, behavior := range career.PermittedBehaviors {
				assert.NotEmpty(t, behavior.Action, "许可行为应该有描述")
				assert.Greater(t, behavior.Reward, 0, "许可行为应该有奖励")
			}

			// 验证首要指令
			assert.NotNil(t, career.PrimeDirective, "应该有首要指令")
			assert.NotEmpty(t, career.PrimeDirective.Description, "首要指令应该有描述")
			assert.Greater(t, career.PrimeDirective.Violation, 0, "违反首要指令应该有惩罚")

			// 验证初始申领物
			assert.NotEmpty(t, career.InitialClaimable, "应该有初始申领物")

			// 验证评估问题
			assert.NotEmpty(t, career.AssessmentQuestions, "应该有评估问题")
		})
	}
}

// TestAnomalyTypesMatch 测试异常体类型与domain定义匹配
func TestAnomalyTypesMatch(t *testing.T) {
	data, err := os.ReadFile("anomalies.json")
	require.NoError(t, err)

	var config AnomaliesConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// 收集配置中的异常体类型
	configTypes := make(map[string]bool)
	for _, anomaly := range config.Anomalies {
		configTypes[anomaly.Name] = true
	}

	// 验证domain中定义的所有类型都在配置中
	for _, anomalyType := range domain.AllAnomalyTypes {
		assert.True(t, configTypes[anomalyType], "配置应该包含异常体类型: %s", anomalyType)
	}

	// 验证配置中的类型数量与domain定义匹配
	assert.Len(t, configTypes, len(domain.AllAnomalyTypes), "配置中的异常体数量应该匹配")
}

// TestRealityTypesMatch 测试现实类型与domain定义匹配
func TestRealityTypesMatch(t *testing.T) {
	data, err := os.ReadFile("realities.json")
	require.NoError(t, err)

	var config RealitiesConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// 收集配置中的现实类型
	configTypes := make(map[string]bool)
	for _, reality := range config.Realities {
		configTypes[reality.Name] = true
	}

	// 验证domain中定义的所有类型都在配置中
	for _, realityType := range domain.AllRealityTypes {
		assert.True(t, configTypes[realityType], "配置应该包含现实类型: %s", realityType)
	}

	// 验证配置中的类型数量与domain定义匹配
	assert.Len(t, configTypes, len(domain.AllRealityTypes), "配置中的现实数量应该匹配")
}

// TestCareerTypesMatch 测试职能类型与domain定义匹配
func TestCareerTypesMatch(t *testing.T) {
	data, err := os.ReadFile("careers.json")
	require.NoError(t, err)

	var config CareersConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// 收集配置中的职能类型
	configTypes := make(map[string]bool)
	for _, career := range config.Careers {
		configTypes[career.Name] = true
	}

	// 验证domain中定义的所有类型都在配置中
	for _, careerType := range domain.AllCareerTypes {
		assert.True(t, configTypes[careerType], "配置应该包含职能类型: %s", careerType)
	}

	// 验证配置中的类型数量与domain定义匹配
	assert.Len(t, configTypes, len(domain.AllCareerTypes), "配置中的职能数量应该匹配")
}

// TestAbilityQualitiesValid 测试能力使用的资质都是有效的
func TestAbilityQualitiesValid(t *testing.T) {
	data, err := os.ReadFile("anomalies.json")
	require.NoError(t, err)

	var config AnomaliesConfig
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	validQualities := make(map[string]bool)
	for _, quality := range domain.AllQualities {
		validQualities[quality] = true
	}

	for _, anomaly := range config.Anomalies {
		for _, ability := range anomaly.Abilities {
			assert.True(t, validQualities[ability.Roll.Quality],
				"能力 %s 使用的资质 %s 应该是有效的", ability.Name, ability.Roll.Quality)
		}
	}
}
