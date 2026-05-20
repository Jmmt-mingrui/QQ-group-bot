package message

import "time"

// Message 表示一条聊天消息
type Message struct {
	ID        string    // 消息 ID
	GroupID   string    // 群组 ID
	UserID    string    // 用户 ID
	Nickname  string    // 用户昵称
	Content   string    // 消息内容
	Timestamp time.Time // 消息时间戳
	MsgType   string    // 消息类型：text / image / at 等
	Compacted bool      // 是否已被压缩
}
