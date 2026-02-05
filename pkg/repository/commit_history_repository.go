package repository

import (
	"fmt"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"gorm.io/gorm"
)

// 确保类型实现了接口
var _ ICommitHistoryRepository = (*CommitHistoryRepository)(nil)

// CommitHistoryRepository handles commit history data operations
type CommitHistoryRepository struct {
	db *gorm.DB
}

// NewCommitHistoryRepository creates a new CommitHistoryRepository
func NewCommitHistoryRepository() *CommitHistoryRepository {
	return &CommitHistoryRepository{
		db: GetDB(),
	}
}

// Create creates a new commit history record
func (r *CommitHistoryRepository) Create(history *models.CommitHistory) error {
	if err := r.db.Create(history).Error; err != nil {
		return fmt.Errorf("failed to create commit history: %w", err)
	}
	return nil
}

// GetByProjectID retrieves commit histories for a project, ordered by most recent
func (r *CommitHistoryRepository) GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error) {
	var histories []models.CommitHistory
	err := r.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// GetRecent retrieves recent commit histories across all projects
func (r *CommitHistoryRepository) GetRecent(limit int) ([]models.CommitHistory, error) {
	var histories []models.CommitHistory
	err := r.db.Preload("Project").
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}
