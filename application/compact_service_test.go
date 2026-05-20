package application

import (
	"errors"
	"testing"
	"time"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
)

// 辅助函数：生成指定数量的测试消息
func generateTestMsgs(n int) []*message.Message {
	msgs := make([]*message.Message, n)
	for i := 0; i < n; i++ {
		msgs[i] = &message.Message{
			ID:      string(rune('a' + i)),
			Content: "消息内容",
		}
	}
	return msgs
}

// TestCompact_Success Compact 成功 — 有 10 条旧消息，LLM 返回压缩文本，Replace 被调用，返回压缩条数
func TestCompact_Success(t *testing.T) {
	llmCalled := false
	replaceCalled := false

	repo := &mockMessageRepo{
		queryOldMessagesFunc: func(groupID string, before time.Time, limit int) ([]*message.Message, error) {
			return generateTestMsgs(10), nil
		},
		replaceFunc: func(oldIDs []string, compacted *message.Message) error {
			replaceCalled = true
			if len(oldIDs) != 10 {
				t.Errorf("期望 oldIDs 长度为 10，实际为 %d", len(oldIDs))
			}
			if compacted.Content != "压缩摘要" {
				t.Errorf("期望压缩内容为 '压缩摘要'，实际为 '%s'", compacted.Content)
			}
			if !compacted.Compacted {
				t.Error("期望压缩消息的 Compacted 标志为 true")
			}
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			llmCalled = true
			return "压缩摘要", nil
		},
	}

	svc := NewCompactService(repo, llmMock)
	count, err := svc.Compact("group1", 7, 100)

	if err != nil {
		t.Fatalf("期望无错误，实际为: %v", err)
	}
	if count != 10 {
		t.Errorf("期望返回 10 条，实际返回 %d", count)
	}
	if !llmCalled {
		t.Error("期望调用了 LLM")
	}
	if !replaceCalled {
		t.Error("期望调用了 Replace")
	}
}

// TestCompact_NoMessages Compact 无旧消息 — QueryOldMessages 返回空，不调 LLM，不调 Replace，返回 0
func TestCompact_NoMessages(t *testing.T) {
	llmCalled := false
	replaceCalled := false

	repo := &mockMessageRepo{
		queryOldMessagesFunc: func(groupID string, before time.Time, limit int) ([]*message.Message, error) {
			return nil, nil
		},
		replaceFunc: func(oldIDs []string, compacted *message.Message) error {
			replaceCalled = true
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			llmCalled = true
			return "", nil
		},
	}

	svc := NewCompactService(repo, llmMock)
	count, err := svc.Compact("group1", 7, 100)

	if err != nil {
		t.Fatalf("期望无错误，实际为: %v", err)
	}
	if count != 0 {
		t.Errorf("期望返回 0 条，实际返回 %d", count)
	}
	if llmCalled {
		t.Error("无消息时不应调用 LLM")
	}
	if replaceCalled {
		t.Error("无消息时不应调用 Replace")
	}
}

// TestCompact_LLMFailure Compact LLM 失败 — LLM 返回 error，Compact 应传递该错误
func TestCompact_LLMFailure(t *testing.T) {
	expectedErr := errors.New("llm service error")
	replaceCalled := false

	repo := &mockMessageRepo{
		queryOldMessagesFunc: func(groupID string, before time.Time, limit int) ([]*message.Message, error) {
			return generateTestMsgs(5), nil
		},
		replaceFunc: func(oldIDs []string, compacted *message.Message) error {
			replaceCalled = true
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "", expectedErr
		},
	}

	svc := NewCompactService(repo, llmMock)
	_, err := svc.Compact("group1", 7, 100)

	if err == nil {
		t.Fatal("期望返回 LLM 错误，实际无错误")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("期望错误包含 %v，实际为 %v", expectedErr, err)
	}
	if replaceCalled {
		t.Error("LLM 失败时不应调用 Replace")
	}
}

// TestCompact_ReplaceFailure Compact Replace 失败 — 确保 Replace 失败时返回错误
func TestCompact_ReplaceFailure(t *testing.T) {
	expectedErr := errors.New("replace error")

	repo := &mockMessageRepo{
		queryOldMessagesFunc: func(groupID string, before time.Time, limit int) ([]*message.Message, error) {
			return generateTestMsgs(3), nil
		},
		replaceFunc: func(oldIDs []string, compacted *message.Message) error {
			return expectedErr
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "压缩摘要", nil
		},
	}

	svc := NewCompactService(repo, llmMock)
	_, err := svc.Compact("group1", 7, 100)

	if err == nil {
		t.Fatal("期望返回 Replace 错误，实际无错误")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("期望错误包含 %v，实际为 %v", expectedErr, err)
	}
}

// TestCompact_LimitPassThrough Compact 限制条数 — 确认 QueryOldMessages 收到的 limit 参数被正确传递
func TestCompact_LimitPassThrough(t *testing.T) {
	var capturedLimit int

	repo := &mockMessageRepo{
		queryOldMessagesFunc: func(groupID string, before time.Time, limit int) ([]*message.Message, error) {
			capturedLimit = limit
			return generateTestMsgs(10), nil
		},
		replaceFunc: func(oldIDs []string, compacted *message.Message) error {
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(messages []llm.Message) (string, error) {
			return "压缩摘要", nil
		},
	}

	svc := NewCompactService(repo, llmMock)
	_, err := svc.Compact("group1", 7, 50)

	if err != nil {
		t.Fatalf("期望无错误，实际为: %v", err)
	}
	if capturedLimit != 50 {
		t.Errorf("期望 limit 为 50，实际为 %d", capturedLimit)
	}
}
