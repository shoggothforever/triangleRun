package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/repository"
)

// saveServiceWithRepo 使用仓储的存档服务实现
type saveServiceWithRepo struct {
	saveRepo     repository.SaveRepository
	gameService  GameService
	agentService AgentService
	version      string
}

// NewSaveServiceWithRepo 创建使用仓储的存档服务
func NewSaveServiceWithRepo(
	saveRepo repository.SaveRepository,
	gameService GameService,
	agentService AgentService,
) SaveService {
	return &saveServiceWithRepo{
		saveRepo:     saveRepo,
		gameService:  gameService,
		agentService: agentService,
		version:      "1.0.0",
	}
}

// CreateSave 创建存档
func (s *saveServiceWithRepo) CreateSave(sessionID, name string) (*SaveSnapshot, error) {
	ctx := context.Background()

	// 获取游戏会话
	session, err := s.gameService.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 验证会话
	if session == nil {
		return nil, domain.NewGameError(domain.ErrNotFound, "游戏会话不存在").
			WithDetails("session_id", sessionID)
	}

	// 获取角色信息用于元数据
	agent, err := s.agentService.GetAgent(session.AgentID)
	if err != nil {
		return nil, err
	}

	// 创建存档快照
	snapshot := &repository.SaveSnapshot{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Name:      name,
		Version:   s.version,
		Snapshot:  session,
		Metadata: map[string]interface{}{
			"agent_name":  agent.Name,
			"scenario_id": session.ScenarioID,
			"phase":       session.Phase,
		},
		CreatedAt: time.Now(),
	}

	// 保存到仓储
	if err := s.saveRepo.Create(ctx, snapshot); err != nil {
		return nil, err
	}

	// 转换为服务层的SaveSnapshot
	return &SaveSnapshot{
		ID:        snapshot.ID,
		SessionID: snapshot.SessionID,
		Name:      snapshot.Name,
		Version:   snapshot.Version,
		Snapshot:  snapshot.Snapshot,
		Metadata:  snapshot.Metadata,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// GetSave 获取存档
func (s *saveServiceWithRepo) GetSave(saveID string) (*SaveSnapshot, error) {
	ctx := context.Background()

	snapshot, err := s.saveRepo.GetByID(ctx, saveID)
	if err != nil {
		return nil, err
	}

	// 转换为服务层的SaveSnapshot
	return &SaveSnapshot{
		ID:        snapshot.ID,
		SessionID: snapshot.SessionID,
		Name:      snapshot.Name,
		Version:   snapshot.Version,
		Snapshot:  snapshot.Snapshot,
		Metadata:  snapshot.Metadata,
		CreatedAt: snapshot.CreatedAt,
	}, nil
}

// ListSaves 列出存档
func (s *saveServiceWithRepo) ListSaves(sessionID string) ([]*SaveMetadata, error) {
	ctx := context.Background()

	var snapshots []*repository.SaveSnapshot
	var err error

	if sessionID == "" {
		snapshots, err = s.saveRepo.List(ctx)
	} else {
		snapshots, err = s.saveRepo.ListBySession(ctx, sessionID)
	}

	if err != nil {
		return nil, err
	}

	metadata := make([]*SaveMetadata, 0, len(snapshots))
	for _, save := range snapshots {
		meta := &SaveMetadata{
			ID:        save.ID,
			SessionID: save.SessionID,
			Name:      save.Name,
			Version:   save.Version,
			CreatedAt: save.CreatedAt,
		}

		// 从元数据中提取信息
		if agentName, ok := save.Metadata["agent_name"].(string); ok {
			meta.AgentName = agentName
		}
		if phase, ok := save.Metadata["phase"].(domain.GamePhase); ok {
			meta.Phase = string(phase)
		} else if phaseStr, ok := save.Metadata["phase"].(string); ok {
			meta.Phase = phaseStr
		}

		metadata = append(metadata, meta)
	}

	return metadata, nil
}

// DeleteSave 删除存档
func (s *saveServiceWithRepo) DeleteSave(saveID string) error {
	ctx := context.Background()
	return s.saveRepo.Delete(ctx, saveID)
}

// LoadSave 加载存档
func (s *saveServiceWithRepo) LoadSave(saveID string) (*domain.GameSession, error) {
	ctx := context.Background()

	snapshot, err := s.saveRepo.GetByID(ctx, saveID)
	if err != nil {
		return nil, err
	}

	// 验证版本兼容性
	if snapshot.Version != s.version {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "存档版本不兼容").
			WithDetails("save_version", snapshot.Version).
			WithDetails("current_version", s.version)
	}

	// 创建会话的深拷贝
	sessionCopy := &domain.GameSession{
		ID:         uuid.New().String(), // 生成新的会话ID
		AgentID:    snapshot.Snapshot.AgentID,
		ScenarioID: snapshot.Snapshot.ScenarioID,
		Phase:      snapshot.Snapshot.Phase,
		State:      copyGameState(snapshot.Snapshot.State),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return sessionCopy, nil
}

// SerializeSession 序列化游戏会话
func (s *saveServiceWithRepo) SerializeSession(session *domain.GameSession) ([]byte, error) {
	if session == nil {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "游戏会话不能为空")
	}

	// 创建包含版本信息的包装结构
	wrapper := struct {
		Version string              `json:"version"`
		Session *domain.GameSession `json:"session"`
	}{
		Version: s.version,
		Session: session,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, domain.NewGameError(domain.ErrInternal, "序列化失败").
			WithDetails("error", err.Error())
	}

	return data, nil
}

// DeserializeSession 反序列化游戏会话
func (s *saveServiceWithRepo) DeserializeSession(data []byte) (*domain.GameSession, error) {
	if len(data) == 0 {
		return nil, domain.NewGameError(domain.ErrInvalidInput, "数据不能为空")
	}

	// 先验证版本
	if err := s.ValidateVersion(data); err != nil {
		return nil, err
	}

	// 解析包装结构
	var wrapper struct {
		Version string              `json:"version"`
		Session *domain.GameSession `json:"session"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "反序列化失败").
			WithDetails("error", err.Error())
	}

	if wrapper.Session == nil {
		return nil, domain.NewGameError(domain.ErrDataCorrupted, "存档数据损坏")
	}

	return wrapper.Session, nil
}

// ValidateVersion 验证版本兼容性
func (s *saveServiceWithRepo) ValidateVersion(data []byte) error {
	if len(data) == 0 {
		return domain.NewGameError(domain.ErrInvalidInput, "数据不能为空")
	}

	// 只解析版本字段
	var versionCheck struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &versionCheck); err != nil {
		return domain.NewGameError(domain.ErrDataCorrupted, "无法读取版本信息").
			WithDetails("error", err.Error())
	}

	if versionCheck.Version != s.version {
		return domain.NewGameError(domain.ErrDataCorrupted, "版本不兼容").
			WithDetails("save_version", versionCheck.Version).
			WithDetails("current_version", s.version)
	}

	return nil
}
