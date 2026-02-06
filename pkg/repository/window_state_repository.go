package repository

import (
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"gorm.io/gorm"
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
func (r *WindowStateRepository) Save(state *models.WindowState) error {
	return r.db.Save(state).Error
}

// DeleteByKey 删除指定 key 的窗口状态
func (r *WindowStateRepository) DeleteByKey(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.WindowState{}).Error
}
