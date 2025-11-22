package service

import (
	"testing"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

func TestAgentService_CreateAgent(t *testing.T) {
	service := NewAgentService()

	req := &CreateAgentRequest{
		Name:        "测试特工",
		Pronouns:    "他/他的",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Name: "李娜", Connection: 6},
			{Name: "王强", Connection: 3},
			{Name: "陈医生", Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 验证基本信息
	if agent.Name != req.Name {
		t.Errorf("期望名称为 %s, 得到 %s", req.Name, agent.Name)
	}

	if agent.Pronouns != req.Pronouns {
		t.Errorf("期望代词为 %s, 得到 %s", req.Pronouns, agent.Pronouns)
	}

	// 验证ARC组件
	if agent.Anomaly.Type != req.AnomalyType {
		t.Errorf("期望异常体类型为 %s, 得到 %s", req.AnomalyType, agent.Anomaly.Type)
	}

	if len(agent.Anomaly.Abilities) != 3 {
		t.Errorf("期望3个异常能力, 得到 %d", len(agent.Anomaly.Abilities))
	}

	if agent.Reality.Type != req.RealityType {
		t.Errorf("期望现实类型为 %s, 得到 %s", req.RealityType, agent.Reality.Type)
	}

	if agent.Career.Type != req.CareerType {
		t.Errorf("期望职能类型为 %s, 得到 %s", req.CareerType, agent.Career.Type)
	}

	// 验证人际关系
	if len(agent.Relationships) != 3 {
		t.Errorf("期望3段人际关系, 得到 %d", len(agent.Relationships))
	}

	totalConnection := agent.TotalConnection()
	if totalConnection != 12 {
		t.Errorf("期望总连结为12, 得到 %d", totalConnection)
	}

	// 验证资质保证
	totalQA := agent.TotalQA()
	if totalQA != 9 {
		t.Errorf("期望总QA为9, 得到 %d", totalQA)
	}

	// 验证初始状态
	if agent.Commendations != 0 {
		t.Errorf("期望初始嘉奖为0, 得到 %d", agent.Commendations)
	}

	if agent.Reprimands != 0 {
		t.Errorf("期望初始申诫为0, 得到 %d", agent.Reprimands)
	}

	if agent.Rating != domain.RatingExcellent {
		t.Errorf("期望初始评级为 %s, 得到 %s", domain.RatingExcellent, agent.Rating)
	}

	if !agent.Alive {
		t.Error("期望角色初始为存活状态")
	}

	if agent.InDebt {
		t.Error("期望角色初始不在负债状态")
	}
}

func TestAgentService_CreateAgentWithInvalidARC(t *testing.T) {
	service := NewAgentService()

	// 测试无效的异常体类型
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: "无效类型",
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	_, err := service.CreateAgent(req)
	if err == nil {
		t.Error("期望无效异常体类型导致错误")
	}
}

func TestAgentService_GetAgent(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	created, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 获取角色
	retrieved, err := service.GetAgent(created.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("期望ID为 %s, 得到 %s", created.ID, retrieved.ID)
	}

	// 测试获取不存在的角色
	_, err = service.GetAgent("不存在的ID")
	if err == nil {
		t.Error("期望获取不存在的角色导致错误")
	}
}

func TestAgentService_UpdateAgent(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 更新角色
	agent.Name = "更新后的名称"
	err = service.UpdateAgent(agent)
	if err != nil {
		t.Fatalf("更新角色失败: %v", err)
	}

	// 验证更新
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Name != "更新后的名称" {
		t.Errorf("期望名称为 '更新后的名称', 得到 %s", updated.Name)
	}
}

func TestAgentService_DeleteAgent(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 删除角色
	err = service.DeleteAgent(agent.ID)
	if err != nil {
		t.Fatalf("删除角色失败: %v", err)
	}

	// 验证删除
	_, err = service.GetAgent(agent.ID)
	if err == nil {
		t.Error("期望获取已删除的角色导致错误")
	}
}

func TestAgentService_SetAnomaly(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 更改异常体类型
	err = service.SetAnomaly(agent.ID, domain.AnomalyCatalog)
	if err != nil {
		t.Fatalf("设置异常体失败: %v", err)
	}

	// 验证更改
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Anomaly.Type != domain.AnomalyCatalog {
		t.Errorf("期望异常体类型为 %s, 得到 %s", domain.AnomalyCatalog, updated.Anomaly.Type)
	}

	if len(updated.Anomaly.Abilities) != 3 {
		t.Errorf("期望3个异常能力, 得到 %d", len(updated.Anomaly.Abilities))
	}

	// 测试无效类型
	err = service.SetAnomaly(agent.ID, "无效类型")
	if err == nil {
		t.Error("期望无效异常体类型导致错误")
	}
}

func TestAgentService_SetReality(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 更改现实类型
	err = service.SetReality(agent.ID, domain.RealityHunted)
	if err != nil {
		t.Fatalf("设置现实失败: %v", err)
	}

	// 验证更改
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Reality.Type != domain.RealityHunted {
		t.Errorf("期望现实类型为 %s, 得到 %s", domain.RealityHunted, updated.Reality.Type)
	}

	// 测试无效类型
	err = service.SetReality(agent.ID, "无效类型")
	if err == nil {
		t.Error("期望无效现实类型导致错误")
	}
}

func TestAgentService_SetCareer(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 更改职能类型
	err = service.SetCareer(agent.ID, domain.CareerCEO)
	if err != nil {
		t.Fatalf("设置职能失败: %v", err)
	}

	// 验证更改
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Career.Type != domain.CareerCEO {
		t.Errorf("期望职能类型为 %s, 得到 %s", domain.CareerCEO, updated.Career.Type)
	}

	// 验证QA也被更新
	if updated.TotalQA() != 9 {
		t.Errorf("期望总QA为9, 得到 %d", updated.TotalQA())
	}

	// 测试无效类型
	err = service.SetCareer(agent.ID, "无效类型")
	if err == nil {
		t.Error("期望无效职能类型导致错误")
	}
}

func TestAgentService_SpendQA(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	initialQA := agent.QA[domain.QualityFocus]

	// 花费QA
	err = service.SpendQA(agent.ID, domain.QualityFocus, 1)
	if err != nil {
		t.Fatalf("花费QA失败: %v", err)
	}

	// 验证花费
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.QA[domain.QualityFocus] != initialQA-1 {
		t.Errorf("期望QA为 %d, 得到 %d", initialQA-1, updated.QA[domain.QualityFocus])
	}

	// 测试花费超过可用QA
	err = service.SpendQA(agent.ID, domain.QualityFocus, 100)
	if err == nil {
		t.Error("期望花费超过可用QA导致错误")
	}
}

func TestAgentService_RestoreQA(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	initialQA := agent.QA[domain.QualityFocus]

	// 花费QA
	err = service.SpendQA(agent.ID, domain.QualityFocus, 1)
	if err != nil {
		t.Fatalf("花费QA失败: %v", err)
	}

	// 恢复QA
	err = service.RestoreQA(agent.ID)
	if err != nil {
		t.Fatalf("恢复QA失败: %v", err)
	}

	// 验证恢复
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.QA[domain.QualityFocus] != initialQA {
		t.Errorf("期望QA恢复到 %d, 得到 %d", initialQA, updated.QA[domain.QualityFocus])
	}
}

func TestAgentService_AddRelationship(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 添加新关系
	newRel := &domain.Relationship{
		Name:       "新朋友",
		Connection: 2,
	}

	err = service.AddRelationship(agent.ID, newRel)
	if err != nil {
		t.Fatalf("添加人际关系失败: %v", err)
	}

	// 验证添加
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if len(updated.Relationships) != 4 {
		t.Errorf("期望4段人际关系, 得到 %d", len(updated.Relationships))
	}

	// 验证ID被生成
	if newRel.ID == "" {
		t.Error("期望关系ID被自动生成")
	}
}

func TestAgentService_UpdateRelationship(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	relID := agent.Relationships[0].ID

	// 更新连结点数
	err = service.UpdateRelationship(agent.ID, relID, 5)
	if err != nil {
		t.Fatalf("更新人际关系失败: %v", err)
	}

	// 验证更新
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Relationships[0].Connection != 5 {
		t.Errorf("期望连结为5, 得到 %d", updated.Relationships[0].Connection)
	}

	// 测试更新不存在的关系
	err = service.UpdateRelationship(agent.ID, "不存在的ID", 5)
	if err == nil {
		t.Error("期望更新不存在的关系导致错误")
	}
}

func TestAgentService_AddCommendations(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 添加嘉奖
	err = service.AddCommendations(agent.ID, 3)
	if err != nil {
		t.Fatalf("添加嘉奖失败: %v", err)
	}

	// 验证添加
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Commendations != 3 {
		t.Errorf("期望嘉奖为3, 得到 %d", updated.Commendations)
	}
}

func TestAgentService_AddReprimands(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 添加申诫
	err = service.AddReprimands(agent.ID, 1)
	if err != nil {
		t.Fatalf("添加申诫失败: %v", err)
	}

	// 验证添加和评级更新
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Reprimands != 1 {
		t.Errorf("期望申诫为1, 得到 %d", updated.Reprimands)
	}

	if updated.Rating != domain.RatingNeedsWork {
		t.Errorf("期望评级为 %s, 得到 %s", domain.RatingNeedsWork, updated.Rating)
	}
}

func TestAgentService_UpdateRating(t *testing.T) {
	service := NewAgentService()

	// 创建角色
	req := &CreateAgentRequest{
		Name:        "测试特工",
		AnomalyType: domain.AnomalyWhisper,
		RealityType: domain.RealityCaretaker,
		CareerType:  domain.CareerPublicRelations,
		Relationships: []*domain.Relationship{
			{Connection: 6},
			{Connection: 3},
			{Connection: 3},
		},
	}

	agent, err := service.CreateAgent(req)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 手动设置申诫（绕过AddReprimands）
	agent.Reprimands = 5

	// 更新评级
	err = service.UpdateRating(agent.ID)
	if err != nil {
		t.Fatalf("更新评级失败: %v", err)
	}

	// 验证评级
	updated, err := service.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("获取角色失败: %v", err)
	}

	if updated.Rating != domain.RatingFinalWarning {
		t.Errorf("期望评级为 %s, 得到 %s", domain.RatingFinalWarning, updated.Rating)
	}
}

func TestAgentService_ListAgents(t *testing.T) {
	service := NewAgentService()

	// 创建多个角色
	for i := 0; i < 3; i++ {
		req := &CreateAgentRequest{
			Name:        "测试特工",
			AnomalyType: domain.AnomalyWhisper,
			RealityType: domain.RealityCaretaker,
			CareerType:  domain.CareerPublicRelations,
			Relationships: []*domain.Relationship{
				{Connection: 6},
				{Connection: 3},
				{Connection: 3},
			},
		}

		_, err := service.CreateAgent(req)
		if err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
	}

	// 列出所有角色
	agents, err := service.ListAgents()
	if err != nil {
		t.Fatalf("列出角色失败: %v", err)
	}

	if len(agents) != 3 {
		t.Errorf("期望3个角色, 得到 %d", len(agents))
	}
}
