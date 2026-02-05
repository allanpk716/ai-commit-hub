package models

// ProviderInfo 表示 provider 的配置状态
type ProviderInfo struct {
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
	Reason     string `json:"reason,omitempty"` // 未配置的原因（如"缺少 API Key"）
}
