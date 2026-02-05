package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// 确保类型实现了接口
var _ GitProjectRepository = (*GitProjectRepository)(nil)

// GitProjectRepository handles git project data operations
type GitProjectRepository struct {
	db *gorm.DB
}

// NewGitProjectRepository creates a new GitProjectRepository
func NewGitProjectRepository() *GitProjectRepository {
	return &GitProjectRepository{
		db: GetDB(),
	}
}

// Create creates a new git project
func (r *GitProjectRepository) Create(project *models.GitProject) error {
	if err := r.db.Create(project).Error; err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetAll retrieves all projects ordered by sort_order
func (r *GitProjectRepository) GetAll() ([]models.GitProject, error) {
	var projects []models.GitProject
	if err := r.db.Order("sort_order asc").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}

// GetByID retrieves a project by ID
func (r *GitProjectRepository) GetByID(id uint) (*models.GitProject, error) {
	var project models.GitProject
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// Update updates a project
func (r *GitProjectRepository) Update(project *models.GitProject) error {
	if err := r.db.Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// Delete deletes a project by ID
func (r *GitProjectRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.GitProject{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// GetMaxSortOrder returns the maximum sort_order value
func (r *GitProjectRepository) GetMaxSortOrder() (int, error) {
	var maxOrder int
	if err := r.db.Model(&models.GitProject{}).
		Select("COALESCE(MAX(sort_order), -1)").
		Scan(&maxOrder).Error; err != nil {
		return 0, fmt.Errorf("failed to get max sort order: %w", err)
	}
	return maxOrder, nil
}

// GetByPath 根据路径获取项目
func (r *GitProjectRepository) GetByPath(path string) (*models.GitProject, error) {
	var project models.GitProject
	if err := r.db.Where("path = ?", path).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to get project by path: %w", err)
	}
	return &project, nil
}
