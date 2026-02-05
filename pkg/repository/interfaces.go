package repository

import (
	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// IGitProjectRepository 定义 Git 项目仓库接口
// 注意：此接口只包含实际实现中存在的方法
type IGitProjectRepository interface {
	// GetAll 获取所有项目
	GetAll() ([]models.GitProject, error)

	// GetByID 根据 ID 获取项目
	GetByID(id uint) (*models.GitProject, error)

	// GetByPath 根据路径获取项目
	GetByPath(path string) (*models.GitProject, error)

	// Create 创建新项目
	Create(project *models.GitProject) error

	// Update 更新项目
	Update(project *models.GitProject) error

	// Delete 删除项目
	Delete(id uint) error

	// GetMaxSortOrder 获取最大排序值
	GetMaxSortOrder() (int, error)
}

// ICommitHistoryRepository 定义提交历史仓库接口
// 注意：此接口只包含实际实现中存在的方法
type ICommitHistoryRepository interface {
	// Create 创建提交历史记录
	Create(history *models.CommitHistory) error

	// GetByProjectID 获取项目的提交历史
	GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error)

	// GetRecent 获取最近的提交历史
	GetRecent(limit int) ([]models.CommitHistory, error)
}
