package models

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
)

// GitProject represents a git repository project
type GitProject struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Path      string `gorm:"not null;uniqueIndex" json:"path"`
	Name      string `json:"name"`
	SortOrder int    `gorm:"index" json:"sort_order"`

	// 项目级别 AI 配置（可选）
	Provider   *string `json:"provider,omitempty"`              // nil 表示使用默认
	Language   *string `json:"language,omitempty"`              // nil 表示使用默认
	Model      *string `json:"model,omitempty"`                 // nil 表示使用默认
	UseDefault bool    `gorm:"default:true" json:"use_default"` // true=使用默认配置

	// Pushover Hook 配置
	HookInstalled    bool       `gorm:"default:false" json:"hook_installed"`
	NotificationMode string     `gorm:"default:'enabled'" json:"notification_mode"` // enabled/pushover_only/windows_only/disabled
	HookVersion      string     `gorm:"size:50" json:"hook_version"`
	HookInstalledAt  *time.Time `json:"hook_installed_at,omitempty"`

	// 运行时状态字段（不持久化到数据库）
	HasUncommittedChanges bool `json:"has_uncommitted_changes" gorm:"-"`
	UntrackedCount        int  `json:"untracked_count" gorm:"-"`
	PushoverNeedsUpdate   bool `json:"pushover_needs_update" gorm:"-"`
}

// TableName specifies the table name for GitProject
func (GitProject) TableName() string {
	return "git_projects"
}

// Validate checks if the project is valid
func (gp *GitProject) Validate() error {
	if gp.Path == "" {
		return fmt.Errorf("项目路径不能为空")
	}

	// Check if path exists
	if _, err := os.Stat(gp.Path); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", gp.Path)
	}

	// Check if it's a git repository
	if _, err := git.PlainOpen(gp.Path); err != nil {
		return fmt.Errorf("不是有效的 git 仓库: %s", gp.Path)
	}

	return nil
}

// DetectName attempts to detect the project name from path or git config
func (gp *GitProject) DetectName() (string, error) {
	// Try folder name first
	folderName := filepath.Base(gp.Path)
	if folderName != "" && folderName != "." && folderName != "/" {
		return folderName, nil
	}

	// Try git config
	repo, err := git.PlainOpen(gp.Path)
	if err != nil {
		return "", fmt.Errorf("无法打开 git 仓库: %w", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return folderName, nil // fallback to folder name
	}

	// Try to get name from remote URL or use folder name
	if len(cfg.Remotes) > 0 {
		for _, remote := range cfg.Remotes {
			if len(remote.URLs) > 0 && remote.URLs[0] != "" {
				return folderName, nil // Use folder name for clarity
			}
		}
	}

	return folderName, nil
}

// SingleProjectStatus 表示单个项目的运行时状态
// 用于增量更新项目状态，避免检查所有项目
type SingleProjectStatus struct {
	Path                  string `json:"path"`
	HasUncommittedChanges bool   `json:"has_uncommitted_changes"`
	UntrackedCount        int    `json:"untracked_count"`
	PushoverNeedsUpdate   bool   `json:"pushover_needs_update"`
}
