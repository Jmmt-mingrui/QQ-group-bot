package summary

import "errors"

// ErrNotFound 表示未找到总结记录
var ErrNotFound = errors.New("summary not found")

// Repository 定义总结的持久化接口
type Repository interface {
	// Save 保存总结，若已存在相同群组和日期的记录则替换
	Save(s *Summary) error
	// QueryByDate 按群组和日期查询总结
	QueryByDate(groupID string, date string) (*Summary, error)
	// ListByGroup 按群组列出总结，按 limit 限制返回条数
	ListByGroup(groupID string, limit int) ([]*Summary, error)
}
