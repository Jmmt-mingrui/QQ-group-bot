package application

import (
	"fmt"
	"strings"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
)

// AskService 处理群内 @Bot 问答请求
type AskService struct {
	msgRepo message.Repository // 消息仓库，用于检索相关聊天记录
	llm     llm.Provider       // LLM 提供者，用于生成答案
}

// NewAskService 创建 AskService 实例
func NewAskService(msgRepo message.Repository, llm llm.Provider) *AskService {
	return &AskService{
		msgRepo: msgRepo,
		llm:     llm,
	}
}

// Ask 处理问答请求：搜索相关聊天记录，构建上下文，调用 LLM 生成答案
func (s *AskService) Ask(groupID string, question string) (string, error) {
	// 校验问题不能为空
	if strings.TrimSpace(question) == "" {
		return "", fmt.Errorf("问题不能为空")
	}

	// 检索相关聊天记录（最多 20 条）
	relatedMsgs, err := s.msgRepo.Search(groupID, question)
	if err != nil {
		return "", err
	}

	// 构建 system prompt
	systemPrompt := "你是一个智能问答助手，基于提供的聊天上下文回答用户问题。"

	// 构建 messages
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}

	// 如果有相关消息，添加上下文
	if len(relatedMsgs) > 0 {
		var contextBuilder strings.Builder
		contextBuilder.WriteString("以下是相关的聊天记录：\n")
		for _, msg := range relatedMsgs {
			contextBuilder.WriteString(fmt.Sprintf("- [%s] %s\n", msg.Nickname, msg.Content))
		}
		messages = append(messages, llm.Message{Role: "user", Content: contextBuilder.String()})
	}

	// 添加用户问题
	messages = append(messages, llm.Message{Role: "user", Content: question})

	// 调用 LLM 生成答案
	answer, err := s.llm.ChatCompletion(messages)
	if err != nil {
		return "", err
	}

	return answer, nil
}
