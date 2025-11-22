package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trpg-solo-engine/backend/internal/domain"
)

// TestSaveService_CreateSave 测试存档保存
func TestSaveService_CreateSave(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 创建存档
	snapshot, err := saveService.CreateSave(session.ID, "Test Save")
	require.NoError(t, err)
	assert.NotEmpty(t, snapshot.ID)
	assert.Equal(t, session.ID, snapshot.SessionID)
	assert.Equal(t, "Test Save", snapshot.Name)
	assert.Equal(t, "1.0.0", snapshot.Version)
	assert.NotNil(t, snapshot.Snapshot)
	assert.NotNil(t, snapshot.Metadata)
}

// TestSaveService_GetSave 测试存档加载
func TestSaveService_GetSave(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 创建存档
	snapshot, err := saveService.CreateSave(session.ID, "Test Save")
	require.NoError(t, err)

	// 获取存档
	loaded, err := saveService.GetSave(snapshot.ID)
	require.NoError(t, err)
	assert.Equal(t, snapshot.ID, loaded.ID)
	assert.Equal(t, snapshot.Name, loaded.Name)
	assert.Equal(t, snapshot.SessionID, loaded.SessionID)
}

// TestSaveService_GetSave_NotFound 测试获取不存在的存档
func TestSaveService_GetSave_NotFound(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 尝试获取不存在的存档
	_, err := saveService.GetSave("non-existent-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "存档不存在")
}

// TestSaveService_ListSaves 测试存档列表
func TestSaveService_ListSaves(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 创建多个存档
	_, err = saveService.CreateSave(session.ID, "Save 1")
	require.NoError(t, err)
	_, err = saveService.CreateSave(session.ID, "Save 2")
	require.NoError(t, err)

	// 列出存档
	saves, err := saveService.ListSaves(session.ID)
	require.NoError(t, err)
	assert.Len(t, saves, 2)
}

// TestSaveService_DeleteSave 测试存档删除
func TestSaveService_DeleteSave(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 创建存档
	snapshot, err := saveService.CreateSave(session.ID, "Test Save")
	require.NoError(t, err)

	// 删除存档
	err = saveService.DeleteSave(snapshot.ID)
	require.NoError(t, err)

	// 验证存档已删除
	_, err = saveService.GetSave(snapshot.ID)
	require.Error(t, err)
}

// TestSaveService_LoadSave 测试加载存档
func TestSaveService_LoadSave(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 修改游戏状态
	session.State.ChaosPool = 10
	session.State.LooseEnds = 5
	session.State.DomainUnlocked = true
	err = gameService.SaveSession(session)
	require.NoError(t, err)

	// 创建存档
	snapshot, err := saveService.CreateSave(session.ID, "Test Save")
	require.NoError(t, err)

	// 加载存档
	loadedSession, err := saveService.LoadSave(snapshot.ID)
	require.NoError(t, err)
	assert.Equal(t, session.AgentID, loadedSession.AgentID)
	assert.Equal(t, session.ScenarioID, loadedSession.ScenarioID)
	assert.Equal(t, session.Phase, loadedSession.Phase)
	assert.Equal(t, 10, loadedSession.State.ChaosPool)
	assert.Equal(t, 5, loadedSession.State.LooseEnds)
	assert.True(t, loadedSession.State.DomainUnlocked)
}

// TestSaveService_SerializeDeserialize 测试序列化和反序列化
func TestSaveService_SerializeDeserialize(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 序列化
	data, err := saveService.SerializeSession(session)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// 反序列化
	loaded, err := saveService.DeserializeSession(data)
	require.NoError(t, err)
	assert.Equal(t, session.AgentID, loaded.AgentID)
	assert.Equal(t, session.ScenarioID, loaded.ScenarioID)
	assert.Equal(t, session.Phase, loaded.Phase)
}

// TestSaveService_ValidateVersion 测试版本验证
func TestSaveService_ValidateVersion(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建测试角色
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
	agent, err := agentService.CreateAgent(req)
	require.NoError(t, err)

	// 创建游戏会话
	session, err := gameService.CreateSession(agent.ID, "test-scenario")
	require.NoError(t, err)

	// 序列化
	data, err := saveService.SerializeSession(session)
	require.NoError(t, err)

	// 验证版本
	err = saveService.ValidateVersion(data)
	require.NoError(t, err)
}

// TestSaveService_ValidateVersion_Invalid 测试无效版本
func TestSaveService_ValidateVersion_Invalid(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	// 创建无效版本的数据
	invalidData := []byte(`{"version":"0.0.1","session":{}}`)

	// 验证版本
	err := saveService.ValidateVersion(invalidData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "版本不兼容")
}

// TestSaveService_ErrorHandling 测试错误处理
func TestSaveService_ErrorHandling(t *testing.T) {
	// 创建服务
	gameService := NewGameService()
	agentService := NewAgentService()
	saveService := NewSaveService(gameService, agentService)

	t.Run("创建存档时会话不存在", func(t *testing.T) {
		_, err := saveService.CreateSave("non-existent-session", "Test Save")
		require.Error(t, err)
	})

	t.Run("序列化空会话", func(t *testing.T) {
		_, err := saveService.SerializeSession(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "游戏会话不能为空")
	})

	t.Run("反序列化空数据", func(t *testing.T) {
		_, err := saveService.DeserializeSession([]byte{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "数据不能为空")
	})

	t.Run("反序列化损坏的数据", func(t *testing.T) {
		_, err := saveService.DeserializeSession([]byte("invalid json"))
		require.Error(t, err)
	})

	t.Run("删除不存在的存档", func(t *testing.T) {
		err := saveService.DeleteSave("non-existent-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "存档不存在")
	})
}
