package application

import (
	"errors"
	"testing"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
)

func TestAskService_Ask_Success(t *testing.T) {
	// 测试 Ask 成功：Search 返回 5 条相关消息，LLM 返回答案
	mockRepo := &mockMessageRepo{
		searchFunc: func(groupID, keyword string) ([]*message.Message, error) {
			msgs := make([]*message.Message, 5)
			for i := 0; i < 5; i++ {
				msgs[i] = &message.Message{Content: "相关消息"}
			}
			return msgs, nil
		},
	}
	mockLLM := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "这是答案", nil
		},
	}
	svc := NewAskService(mockRepo, mockLLM)
	answer, err := svc.Ask("group1", "今天天气怎么样？")
	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	if answer == "" {
		t.Fatal("期望答案非空")
	}
	if answer != "这是答案" {
		t.Fatalf("期望 '这是答案'，得到: %s", answer)
	}
}

func TestAskService_Ask_NoRelatedMessages(t *testing.T) {
	// 测试无相关消息：Search 返回空，LLM 仍应返回答案
	mockRepo := &mockMessageRepo{
		searchFunc: func(groupID, keyword string) ([]*message.Message, error) {
			return nil, nil
		},
	}
	mockLLM := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "基于无上下文的回答", nil
		},
	}
	svc := NewAskService(mockRepo, mockLLM)
	answer, err := svc.Ask("group1", "今天天气怎么样？")
	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	if answer == "" {
		t.Fatal("期望答案非空")
	}
	if answer != "基于无上下文的回答" {
		t.Fatalf("期望 '基于无上下文的回答'，得到: %s", answer)
	}
}

func TestAskService_Ask_SearchError(t *testing.T) {
	// 测试 Search 失败：错误应被传递
	expectedErr := errors.New("搜索失败")
	mockRepo := &mockMessageRepo{
		searchFunc: func(groupID, keyword string) ([]*message.Message, error) {
			return nil, expectedErr
		},
	}
	mockLLM := &mockLLM{}
	svc := NewAskService(mockRepo, mockLLM)
	_, err := svc.Ask("group1", "今天天气怎么样？")
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期望错误 %v，得到: %v", expectedErr, err)
	}
}

func TestAskService_Ask_LLMError(t *testing.T) {
	// 测试 LLM 失败：错误应被传递
	expectedErr := errors.New("LLM 调用失败")
	mockRepo := &mockMessageRepo{
		searchFunc: func(groupID, keyword string) ([]*message.Message, error) {
			return []*message.Message{{Content: "相关消息"}}, nil
		},
	}
	mockLLM := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "", expectedErr
		},
	}
	svc := NewAskService(mockRepo, mockLLM)
	_, err := svc.Ask("group1", "今天天气怎么样？")
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期望错误 %v，得到: %v", expectedErr, err)
	}
}

func TestAskService_Ask_ContextIncluded(t *testing.T) {
	// 测试答案包含上下文：验证传给 LLM 的 messages 中包含搜索到的消息内容
	searchContent := "昨天的聊天记录内容"
	var capturedMessages []llm.Message
	mockRepo := &mockMessageRepo{
		searchFunc: func(groupID, keyword string) ([]*message.Message, error) {
			return []*message.Message{{Content: searchContent}}, nil
		},
	}
	mockLLM := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			capturedMessages = messages
			return "基于上下文的答案", nil
		},
	}
	svc := NewAskService(mockRepo, mockLLM)
	_, err := svc.Ask("group1", "今天天气怎么样？")
	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}

	// 验证 messages 中包含搜索到的消息内容
	found := false
	for _, m := range capturedMessages {
		if m.Content == searchContent {
			found = true
			break
		}
		// 也可能被拼接到 system/user 消息中
	}
	if !found {
		// 检查是否被包含在某个消息内容中
		for _, m := range capturedMessages {
			if contains(m.Content, searchContent) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("LLM messages 中未包含搜索到的消息内容: %v", capturedMessages)
	}
}

func TestAskService_Ask_EmptyQuestion(t *testing.T) {
	// 测试空问题：应返回错误
	mockRepo := &mockMessageRepo{}
	mockLLM := &mockLLM{}
	svc := NewAskService(mockRepo, mockLLM)
	_, err := svc.Ask("group1", "")
	if err == nil {
		t.Fatal("期望空问题返回错误，但得到 nil")
	}
	if err.Error() != "问题不能为空" {
		t.Fatalf("期望错误消息 '问题不能为空'，得到: %v", err)
	}
}

// contains 辅助函数，判断字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchInString(s, substr)
}

func searchInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
