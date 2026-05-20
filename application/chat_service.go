package application

import (
	"fmt"
	"time"

	"qq-group-bot/domain/message"
)

// ChatService 消息处理服务，负责接收 OneBot 事件、创建 Message 实体并持久化存储
type ChatService struct {
	repo message.Repository
}

// NewChatService 创建 ChatService 实例
func NewChatService(repo message.Repository) *ChatService {
	return &ChatService{repo: repo}
}

// HandleMessage 处理一条消息：创建 Message 实体并调用 repo.Save 持久化
func (s *ChatService) HandleMessage(groupID, userID, nickname, content, msgType string) error {
	msg := &message.Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		GroupID:   groupID,
		UserID:    userID,
		Nickname:  nickname,
		Content:   content,
		Timestamp: time.Now(),
		MsgType:   msgType,
	}
	return s.repo.Save(msg)
}
