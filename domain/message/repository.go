package message

import "time"

// Repository 定义消息仓库接口，用于消息的持久化存储与检索
type Repository interface {
	// Save 保存一条消息
	Save(msg *Message) error

	// QueryByDate 按群组和日期查询消息
	QueryByDate(groupID string, date time.Time) ([]*Message, error)

	// QueryOldMessages 查询指定时间之前的消息，最多返回 limit 条
	QueryOldMessages(groupID string, before time.Time, limit int) ([]*Message, error)

	// Search 在指定群组中按关键词搜索消息
	Search(groupID string, keyword string) ([]*Message, error)

	// Replace 删除旧消息（按 ID 列表）并插入一条压缩后的消息
	Replace(oldIDs []string, compacted *Message) error
}
