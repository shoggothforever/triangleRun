package domain

import (
	"testing"
	"time"
)

func TestAgentCreation(t *testing.T) {
	agent := &Agent{
		ID:       "test-agent-1",
		Name:     "测试特工",
		Pronouns: "他/他的",
		Anomaly: &Anomaly{
			Type: AnomalyWhisper,
			Abilities: []*AnomalyAbility{
				{
					ID:          "whisper-1",
					Name:        "再说一遍？",
					AnomalyType: AnomalyWhisper,
					Trigger: &AbilityTrigger{
						Type:        TriggerResponse,
						Description: "用'再说一遍？'回应一句说出的话",
					},
					Roll: &AbilityRoll{
						Quality:   QualityPresence,
						DiceCount: 6,
						DiceType:  4,
					},
				},
				{ID: "whisper-2", Name: "能力2", AnomalyType: AnomalyWhisper},
				{ID: "whisper-3", Name: "能力3", AnomalyType: AnomalyWhisper},
			},
		},
		Reality: &Reality{
			Type: RealityCaretaker,
			Trigger: &RealityTrigger{
				Name:        "需要关爱",
				Cost:        0,
				Effect:      "受照料者需要你的关注",
				Consequence: "失去1点连结",
			},
			OverloadRelief: &OverloadRelief{
				Name:      "这是你的最爱！",
				Condition: "做某件能让受照料者开心的事",
				Effect:    "无视所有过载",
			},
			DegradationTrack: &DegradationTrack{
				Name:   "独立",
				Filled: 0,
				Total:  4,
			},
		},
		Career: &Career{
			Type: CareerPublicRelations,
			QA: map[string]int{
				QualityFocus:      1,
				QualityEmpathy:    2,
				QualityPresence:   2,
				QualityDeception:  2,
				QualityInitiative: 1,
				QualityProfession: 1,
			},
		},
		QA: map[string]int{
			QualityFocus:      1,
			QualityEmpathy:    2,
			QualityPresence:   2,
			QualityDeception:  2,
			QualityInitiative: 1,
			QualityProfession: 1,
		},
		Relationships: []*Relationship{
			{ID: "rel-1", Name: "李娜", Connection: 6},
			{ID: "rel-2", Name: "王强", Connection: 3},
			{ID: "rel-3", Name: "陈医生", Connection: 3},
		},
		Commendations: 0,
		Reprimands:    0,
		Rating:        RatingExcellent,
		Alive:         true,
		InDebt:        false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if agent.ID != "test-agent-1" {
		t.Errorf("expected ID to be test-agent-1, got %s", agent.ID)
	}

	if agent.Anomaly.Type != AnomalyWhisper {
		t.Errorf("expected anomaly type to be %s, got %s", AnomalyWhisper, agent.Anomaly.Type)
	}

	if len(agent.Anomaly.Abilities) != 3 {
		t.Errorf("expected 3 abilities, got %d", len(agent.Anomaly.Abilities))
	}

	if agent.Reality.Type != RealityCaretaker {
		t.Errorf("expected reality type to be %s, got %s", RealityCaretaker, agent.Reality.Type)
	}

	if agent.Career.Type != CareerPublicRelations {
		t.Errorf("expected career type to be %s, got %s", CareerPublicRelations, agent.Career.Type)
	}
}

func TestAgentValidation(t *testing.T) {
	// 测试有效的角色
	validAgent := &Agent{
		Anomaly: &Anomaly{
			Type: AnomalyWhisper,
			Abilities: []*AnomalyAbility{
				{ID: "1", Name: "能力1"},
				{ID: "2", Name: "能力2"},
				{ID: "3", Name: "能力3"},
			},
		},
		Reality: &Reality{
			Type: RealityCaretaker,
		},
		Career: &Career{
			Type: CareerPublicRelations,
		},
		QA: map[string]int{
			QualityFocus: 1,
		},
		Relationships: []*Relationship{
			{ID: "1", Connection: 6},
			{ID: "2", Connection: 3},
			{ID: "3", Connection: 3},
		},
	}

	if err := validAgent.ValidateARC(); err != nil {
		t.Errorf("expected valid agent to pass validation, got error: %v", err)
	}

	// 测试无效的异常体类型
	invalidAnomaly := &Agent{
		Anomaly: &Anomaly{
			Type:      "无效类型",
			Abilities: []*AnomalyAbility{{}, {}, {}},
		},
		Reality:       &Reality{Type: RealityCaretaker},
		Career:        &Career{Type: CareerPublicRelations},
		Relationships: []*Relationship{{Connection: 6}, {Connection: 3}, {Connection: 3}},
	}

	if err := invalidAnomaly.ValidateARC(); err == nil {
		t.Error("expected invalid anomaly type to fail validation")
	}

	// 测试错误的能力数量
	wrongAbilityCount := &Agent{
		Anomaly: &Anomaly{
			Type:      AnomalyWhisper,
			Abilities: []*AnomalyAbility{{}, {}}, // 只有2个
		},
		Reality:       &Reality{Type: RealityCaretaker},
		Career:        &Career{Type: CareerPublicRelations},
		Relationships: []*Relationship{{Connection: 6}, {Connection: 3}, {Connection: 3}},
	}

	if err := wrongAbilityCount.ValidateARC(); err == nil {
		t.Error("expected wrong ability count to fail validation")
	}

	// 测试错误的连结总数
	wrongConnection := &Agent{
		Anomaly: &Anomaly{
			Type:      AnomalyWhisper,
			Abilities: []*AnomalyAbility{{}, {}, {}},
		},
		Reality: &Reality{Type: RealityCaretaker},
		Career:  &Career{Type: CareerPublicRelations},
		Relationships: []*Relationship{
			{Connection: 5},
			{Connection: 5},
			{Connection: 5}, // 总计15，应该是12
		},
	}

	if err := wrongConnection.ValidateARC(); err == nil {
		t.Error("expected wrong connection total to fail validation")
	}
}

func TestQAManagement(t *testing.T) {
	agent := &Agent{
		QA: map[string]int{
			QualityFocus:   2,
			QualityEmpathy: 3,
		},
		Career: &Career{
			QA: map[string]int{
				QualityFocus:   2,
				QualityEmpathy: 3,
			},
		},
	}

	// 测试花费QA
	err := agent.SpendQA(QualityFocus, 1)
	if err != nil {
		t.Errorf("expected to spend QA successfully, got error: %v", err)
	}

	if agent.QA[QualityFocus] != 1 {
		t.Errorf("expected QA to be 1 after spending, got %d", agent.QA[QualityFocus])
	}

	// 测试花费超过可用QA
	err = agent.SpendQA(QualityFocus, 5)
	if err == nil {
		t.Error("expected error when spending more QA than available")
	}

	// 测试恢复QA
	agent.RestoreQA()
	if agent.QA[QualityFocus] != 2 {
		t.Errorf("expected QA to be restored to 2, got %d", agent.QA[QualityFocus])
	}
}

func TestRatingSystem(t *testing.T) {
	tests := []struct {
		reprimands int
		expected   string
	}{
		{0, RatingExcellent},
		{1, RatingNeedsWork},
		{2, RatingProbation},
		{3, RatingProbation},
		{4, RatingFinalWarning},
		{9, RatingFinalWarning},
		{10, RatingRevoked},
		{15, RatingRevoked},
	}

	for _, tt := range tests {
		rating := GetRating(tt.reprimands)
		if rating != tt.expected {
			t.Errorf("GetRating(%d) = %s, expected %s", tt.reprimands, rating, tt.expected)
		}
	}

	// 测试添加申诫
	agent := &Agent{
		Reprimands: 0,
		Rating:     RatingExcellent,
	}

	agent.AddReprimands(1)
	if agent.Reprimands != 1 {
		t.Errorf("expected reprimands to be 1, got %d", agent.Reprimands)
	}

	if agent.Rating != RatingNeedsWork {
		t.Errorf("expected rating to be %s, got %s", RatingNeedsWork, agent.Rating)
	}
}

func TestRelationshipManagement(t *testing.T) {
	agent := &Agent{
		Relationships: []*Relationship{
			{ID: "1", Name: "李娜", Connection: 6},
			{ID: "2", Name: "王强", Connection: 3},
			{ID: "3", Name: "陈医生", Connection: 2},
		},
	}

	// 测试获取最弱关系
	weakest := agent.GetWeakestRelationship()
	if weakest.Name != "陈医生" {
		t.Errorf("expected weakest relationship to be 陈医生, got %s", weakest.Name)
	}

	// 测试总连结计算
	total := agent.TotalConnection()
	if total != 11 {
		t.Errorf("expected total connection to be 11, got %d", total)
	}
}

func TestGameError(t *testing.T) {
	err := NewGameError(ErrInvalidInput, "测试错误")
	err.WithDetails("field", "test_field")

	if err.Code != ErrInvalidInput {
		t.Errorf("expected error code INVALID_INPUT, got %s", err.Code)
	}

	if err.Message != "测试错误" {
		t.Errorf("expected message 测试错误, got %s", err.Message)
	}

	if err.Details["field"] != "test_field" {
		t.Errorf("expected field detail to be test_field")
	}

	expectedError := "[INVALID_INPUT] 测试错误"
	if err.Error() != expectedError {
		t.Errorf("expected error string %s, got %s", expectedError, err.Error())
	}
}
