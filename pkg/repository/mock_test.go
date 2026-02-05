package repository

import (
	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

// MockGitProjectRepository 是用于测试的 Mock 实现
type MockGitProjectRepository struct {
	projects []models.GitProject
	err      error
}

// NewMockGitProjectRepository 创建一个 Mock Git 项目仓库
func NewMockGitProjectRepository(projects []models.GitProject, err error) IGitProjectRepository {
	return &MockGitProjectRepository{
		projects: projects,
		err:      err,
	}
}

func (m *MockGitProjectRepository) GetAll() ([]models.GitProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.projects, nil
}

func (m *MockGitProjectRepository) GetByID(id uint) (*models.GitProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.projects {
		if m.projects[i].ID == id {
			// 返回值的副本，避免外部修改直接影响内部状态
			project := m.projects[i]
			return &project, nil
		}
	}
	return nil, nil
}

func (m *MockGitProjectRepository) GetByPath(path string) (*models.GitProject, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.projects {
		if m.projects[i].Path == path {
			return &m.projects[i], nil
		}
	}
	return nil, nil
}

func (m *MockGitProjectRepository) Create(project *models.GitProject) error {
	if m.err != nil {
		return m.err
	}
	m.projects = append(m.projects, *project)
	return nil
}

func (m *MockGitProjectRepository) Update(project *models.GitProject) error {
	if m.err != nil {
		return m.err
	}
	// 直接修改切片
	for i := range m.projects {
		if m.projects[i].ID == project.ID {
			m.projects[i].Name = project.Name
			m.projects[i].Path = project.Path
			m.projects[i].SortOrder = project.SortOrder
			// 复制其他字段...
			return nil
		}
	}
	return nil
}

func (m *MockGitProjectRepository) Delete(id uint) error {
	if m.err != nil {
		return m.err
	}
	for i := range m.projects {
		if m.projects[i].ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockGitProjectRepository) UpdateLastCommitTime(id uint, commitTime int64) error {
	// GitProject 模型没有 LastCommitTime 字段
	// 这个方法不在当前接口中
	return nil
}

func (m *MockGitProjectRepository) UpdateHookStatus(id uint, needsUpdate bool) error {
	// 这个方法不在当前接口中
	return nil
}

func (m *MockGitProjectRepository) GetMaxSortOrder() (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	maxOrder := -1
	for _, p := range m.projects {
		if p.SortOrder > maxOrder {
			maxOrder = p.SortOrder
		}
	}
	return maxOrder, nil
}

// MockCommitHistoryRepository 是用于测试的 Mock 实现
type MockCommitHistoryRepository struct {
	histories []models.CommitHistory
	err       error
}

// NewMockCommitHistoryRepository 创建一个 Mock 提交历史仓库
func NewMockCommitHistoryRepository(histories []models.CommitHistory, err error) ICommitHistoryRepository {
	return &MockCommitHistoryRepository{
		histories: histories,
		err:       err,
	}
}

func (m *MockCommitHistoryRepository) Create(history *models.CommitHistory) error {
	if m.err != nil {
		return m.err
	}
	m.histories = append(m.histories, *history)
	return nil
}

func (m *MockCommitHistoryRepository) GetByProjectID(projectID uint, limit int) ([]models.CommitHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.CommitHistory
	count := 0
	for _, h := range m.histories {
		if h.ProjectID == projectID {
			result = append(result, h)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockCommitHistoryRepository) GetRecent(limit int) ([]models.CommitHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	if limit > 0 && len(m.histories) > limit {
		return m.histories[:limit], nil
	}
	return m.histories, nil
}

func (m *MockCommitHistoryRepository) Delete(id uint) error {
	// Delete 方法不在当前接口中，但保留以避免破坏现有代码
	return m.err
}

func (m *MockCommitHistoryRepository) DeleteByProjectID(projectID uint) error {
	// DeleteByProjectID 方法不在当前接口中，但保留以避免破坏现有代码
	return m.err
}
