package repository

import (
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WindowStateRepository 窗口状态数据访问层
type WindowStateRepository struct {
	db *gorm.DB
}

// NewWindowStateRepository 创建 WindowStateRepository 实例
func NewWindowStateRepository() *WindowStateRepository {
	return &WindowStateRepository{
		db: GetDB(),
	}
}

// GetByKey 根据 key 获取窗口状态
func (r *WindowStateRepository) GetByKey(key string) (*models.WindowState, error) {
	var state models.WindowState
	err := r.db.Where("key = ?", key).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// Save 保存或更新窗口状态(Upsert)
// 如果 key 已存在则更新，否则插入新记录
func (r *WindowStateRepository) Save(state *models.WindowState) error {
	// 使用 clause.OnConflict 实现 Upsert
	// 当 key 冲突时，更新所有字段
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"x", "y", "width", "height", "maximized"}),
	}).Create(state).Error
}

// DeleteByKey 删除指定 key 的窗口状态
func (r *WindowStateRepository) DeleteByKey(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.WindowState{}).Error
}
