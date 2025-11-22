package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trpg-solo-engine/backend/internal/domain"
	"github.com/trpg-solo-engine/backend/internal/infrastructure/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SaveRepository 存档仓储接口
type SaveRepository interface {
	// CRUD操作
	Create(ctx context.Context, save *SaveSnapshot) error
	GetByID(ctx context.Context, id string) (*SaveSnapshot, error)
	Delete(ctx context.Context, id string) error
	ListBySession(ctx context.Context, sessionID string) ([]*SaveSnapshot, error)
	List(ctx context.Context) ([]*SaveSnapshot, error)

	// 事务支持
	WithTx(tx *gorm.DB) SaveRepository
}

// SaveSnapshot 存档快照（与service层的SaveSnapshot保持一致）
type SaveSnapshot struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Snapshot  *domain.GameSession    `json:"snapshot"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

type saveRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSaveRepository 创建存档仓储实例
func NewSaveRepository(db *gorm.DB, logger *zap.Logger) SaveRepository {
	return &saveRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建存档
func (r *saveRepository) Create(ctx context.Context, save *SaveSnapshot) error {
	model, err := r.toModel(save)
	if err != nil {
		return fmt.Errorf("failed to convert save to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create save: %w", err)
	}

	// 更新ID（如果数据库生成了新ID）
	save.ID = model.ID

	return nil
}

// GetByID 根据ID获取存档
func (r *saveRepository) GetByID(ctx context.Context, id string) (*SaveSnapshot, error) {
	var model database.SaveModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.NewGameError(domain.ErrNotFound, "存档不存在").
				WithDetails("save_id", id)
		}
		return nil, fmt.Errorf("failed to get save: %w", err)
	}

	save, err := r.toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to save: %w", err)
	}

	return save, nil
}

// Delete 删除存档
func (r *saveRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&database.SaveModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete save: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.NewGameError(domain.ErrNotFound, "存档不存在").
			WithDetails("save_id", id)
	}

	return nil
}

// ListBySession 列出指定会话的所有存档
func (r *saveRepository) ListBySession(ctx context.Context, sessionID string) ([]*SaveSnapshot, error) {
	var models []database.SaveModel
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list saves by session: %w", err)
	}

	saves := make([]*SaveSnapshot, 0, len(models))
	for _, model := range models {
		save, err := r.toDomain(&model)
		if err != nil {
			r.logger.Warn("failed to convert model to save",
				zap.Error(err),
				zap.String("save_id", model.ID))
			continue
		}
		saves = append(saves, save)
	}

	return saves, nil
}

// List 列出所有存档
func (r *saveRepository) List(ctx context.Context) ([]*SaveSnapshot, error) {
	var models []database.SaveModel
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list saves: %w", err)
	}

	saves := make([]*SaveSnapshot, 0, len(models))
	for _, model := range models {
		save, err := r.toDomain(&model)
		if err != nil {
			r.logger.Warn("failed to convert model to save",
				zap.Error(err),
				zap.String("save_id", model.ID))
			continue
		}
		saves = append(saves, save)
	}

	return saves, nil
}

// WithTx 使用事务
func (r *saveRepository) WithTx(tx *gorm.DB) SaveRepository {
	return &saveRepository{
		db:     tx,
		logger: r.logger,
	}
}

// toModel 将SaveSnapshot转换为数据库模型
func (r *saveRepository) toModel(save *SaveSnapshot) (*database.SaveModel, error) {
	// 创建完整的快照数据结构
	snapshotData := map[string]interface{}{
		"version":  save.Version,
		"snapshot": save.Snapshot,
		"metadata": save.Metadata,
	}

	// 序列化快照数据
	snapshotJSON, err := json.Marshal(snapshotData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return &database.SaveModel{
		ID:        save.ID,
		SessionID: save.SessionID,
		Name:      save.Name,
		Snapshot:  string(snapshotJSON),
		CreatedAt: save.CreatedAt.Unix(),
	}, nil
}

// toDomain 将数据库模型转换为SaveSnapshot
func (r *saveRepository) toDomain(model *database.SaveModel) (*SaveSnapshot, error) {
	// 反序列化快照数据
	var snapshotData struct {
		Version  string                 `json:"version"`
		Snapshot *domain.GameSession    `json:"snapshot"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(model.Snapshot), &snapshotData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &SaveSnapshot{
		ID:        model.ID,
		SessionID: model.SessionID,
		Name:      model.Name,
		Version:   snapshotData.Version,
		Snapshot:  snapshotData.Snapshot,
		Metadata:  snapshotData.Metadata,
		CreatedAt: time.Unix(model.CreatedAt, 0),
	}, nil
}
