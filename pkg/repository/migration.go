package repository

import (
	"fmt"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"gorm.io/gorm"
)

// MigrateAddProjectAIConfig 添加项目 AI 配置字段的迁移
func MigrateAddProjectAIConfig(db *gorm.DB) error {
	logger.Info("开始迁移：添加项目 AI 配置字段")

	// AutoMigrate 会自动添加新字段
	if err := db.AutoMigrate(&models.GitProject{}); err != nil {
		return fmt.Errorf("AutoMigrate 失败: %w", err)
	}

	// 将现有项目标记为使用默认配置
	result := db.Model(&models.GitProject{}).
		Where("use_default IS NULL OR use_default = false").
		Update("use_default", true)

	if result.Error != nil {
		return fmt.Errorf("更新现有项目失败: %w", result.Error)
	}

	logger.Infof("迁移完成：已更新 %d 个项目", result.RowsAffected)
	return nil
}
