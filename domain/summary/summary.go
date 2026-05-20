// Package summary 提供群聊总结相关的领域模型
package summary

import "time"

// Summary 表示某群某日的聊天总结
type Summary struct {
	ID        string    // 总结唯一标识
	GroupID   string    // 群组 ID
	Date      string    // 日期，格式 yyyy-mm-dd
	Content   string    // 总结内容
	MsgCount  int       // 原始消息数量
	CreatedAt time.Time // 创建时间
}

// NewSummary 创建一条总结，自动填充 CreatedAt 为当前时间
func NewSummary(id, groupID, date, content string, msgCount int) *Summary {
	return &Summary{
		ID:        id,
		GroupID:   groupID,
		Date:      date,
		Content:   content,
		MsgCount:  msgCount,
		CreatedAt: time.Now(),
	}
}
