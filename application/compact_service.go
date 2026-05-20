package application

import (
	"fmt"
	"time"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
)

// CompactService 压缩旧聊天记录：查询 N 天前的消息，调用 LLM 压缩，替换为一条压缩消息
type CompactService struct {
	msgRepo message.Repository
	llm     llm.Provider
}

// NewCompactService 创建 CompactService
func NewCompactService(msgRepo message.Repository, llm llm.Provider) *CompactService {
	return &CompactService{
		msgRepo: msgRepo,
		llm:     llm,
	}
}

// Compact 执行消息压缩流程：
//  1. 计算 beforeTime = now - beforeDays 天
//  2. 查询旧消息 (QueryOldMessages)
//  3. 若无消息，返回 0
//  4. 拼装消息文本，调用 LLM 压缩
//  5. 收集旧消息 ID，调用 Replace 替换
//  6. 返回压缩的消息条数
func (s *CompactService) Compact(groupID string, beforeDays int, limit int) (int, error) {
	// 计算截止时间
	beforeTime := time.Now().Add(-time.Duration(beforeDays) * 24 * time.Hour)

	// 查询旧消息
	msgs, err := s.msgRepo.QueryOldMessages(groupID, beforeTime, limit)
	if err != nil {
		return 0, fmt.Errorf("查询旧消息失败: %w", err)
	}

	// 无消息则直接返回
	if len(msgs) == 0 {
		return 0, nil
	}

	// 将消息拼装为文本
	text := message.CompactMessages(msgs)

	// 调用 LLM 生成压缩摘要
	llmMessages := []llm.Message{
		{Role: "system", Content: "你是一个聊天记录压缩助手，请将以下聊天记录压缩为一段简洁的摘要，保留关键信息。"},
		{Role: "user", Content: text},
	}
	compressed, err := s.llm.ChatCompletion(llmMessages)
	if err != nil {
		return 0, fmt.Errorf("LLM 压缩失败: %w", err)
	}

	// 收集旧消息 ID
	oldIDs := make([]string, len(msgs))
	for i, msg := range msgs {
		oldIDs[i] = msg.ID
	}

	// 构建压缩后的消息
	compactedMsg := &message.Message{
		GroupID:   groupID,
		Content:   compressed,
		Timestamp: time.Now(),
		Compacted: true,
	}

	// 替换旧消息
	if err := s.msgRepo.Replace(oldIDs, compactedMsg); err != nil {
		return 0, fmt.Errorf("替换旧消息失败: %w", err)
	}

	return len(msgs), nil
}
