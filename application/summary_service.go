// Package application 提供业务编排服务
package application

import (
	"fmt"
	"time"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
	"qq-group-bot/domain/summary"

	"github.com/google/uuid"
)

// SummaryService 群聊每日总结服务
type SummaryService struct {
	msgRepo     message.Repository
	summaryRepo summary.Repository
	llm         llm.Provider
}

// NewSummaryService 创建 SummaryService 实例
func NewSummaryService(msgRepo message.Repository, summaryRepo summary.Repository, llm llm.Provider) *SummaryService {
	return &SummaryService{
		msgRepo:     msgRepo,
		summaryRepo: summaryRepo,
		llm:         llm,
	}
}

// Summarize 生成指定群组当天的聊天总结
// 流程：查询当天所有消息 -> 调 LLM 生成摘要 -> 保存 Summary -> 返回摘要文本
func (s *SummaryService) Summarize(groupID string) (string, error) {
	// 获取当天零点
	today := time.Now().Truncate(24 * time.Hour)
	dateStr := today.Format("2006-01-02")

	// 查询当天所有消息
	msgs, err := s.msgRepo.QueryByDate(groupID, today)
	if err != nil {
		return "", fmt.Errorf("查询消息失败: %w", err)
	}

	// 无消息则直接返回空字符串
	if len(msgs) == 0 {
		return "", nil
	}

	// 构建 LLM prompt，包含消息数量和日期
	prompt := fmt.Sprintf("请总结以下 %d 条群聊消息（日期：%s）：\n", len(msgs), dateStr)
	for _, m := range msgs {
		prompt += fmt.Sprintf("%s: %s\n", m.Nickname, m.Content)
	}

	// 调用 LLM 生成摘要
	result, err := s.llm.ChatCompletion([]llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 生成摘要失败: %w", err)
	}

	// 保存总结记录
	summaryRecord := summary.NewSummary(
		uuid.New().String(),
		groupID,
		dateStr,
		result,
		len(msgs),
	)
	if err := s.summaryRepo.Save(summaryRecord); err != nil {
		return "", fmt.Errorf("保存总结失败: %w", err)
	}

	return result, nil
}
