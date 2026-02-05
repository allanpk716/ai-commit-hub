package repository

import (
	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// GitProjectRepository 定义 Git 项目仓库接口
type GitProjectRepository interface {
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

	// UpdateLastCommitTime 更新最后提交时间
	UpdateLastCommitTime(id uint, commitTime int64) error

	// UpdateHookStatus 更新 Hook 状态
	UpdateHookStatus(id uint, needsUpdate bool) error
}

// CommitHistoryRepository 定义提交历史仓库接口
type CommitHistoryRepository interface {
	// Create 创建提交历史记录
	Create(history *models.CommitHistory) error

	// GetByProjectID 获取项目的提交历史
	GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error)

	// Delete 删除历史记录
	Delete(id uint) error

	// DeleteByProjectID 删除项目的所有历史
	DeleteByProjectID(projectID uint) error
}
