package application

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
	"qq-group-bot/domain/summary"
)

// generateTestMessages 生成指定数量的测试消息
func generateTestMessages(count int, groupID string) []*message.Message {
	msgs := make([]*message.Message, count)
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < count; i++ {
		msgs[i] = &message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			GroupID:   groupID,
			UserID:    fmt.Sprintf("user-%d", i),
			Nickname:  fmt.Sprintf("用户%d", i),
			Content:   fmt.Sprintf("这是一条测试消息 #%d", i),
			Timestamp: today.Add(time.Duration(i) * time.Minute),
			MsgType:   "text",
		}
	}
	return msgs
}

// TestSummarize_Success 当天有 50 条消息，LLM 返回摘要，Save 被调用，返回摘要文本
func TestSummarize_Success(t *testing.T) {
	groupID := "test-group-1"
	expectedSummary := "这是今天的总结：群聊活跃，讨论了多个话题。"
	var savedSummary *summary.Summary

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return generateTestMessages(50, gid), nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			savedSummary = s
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			return expectedSummary, nil
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err != nil {
		t.Fatalf("预期无错误，得到: %v", err)
	}
	if result != expectedSummary {
		t.Errorf("预期摘要 %q，得到 %q", expectedSummary, result)
	}
	if savedSummary == nil {
		t.Fatal("Save 未被调用")
	}
	if savedSummary.GroupID != groupID {
		t.Errorf("Summary.GroupID 预期 %q，得到 %q", groupID, savedSummary.GroupID)
	}
	if savedSummary.MsgCount != 50 {
		t.Errorf("Summary.MsgCount 预期 50，得到 %d", savedSummary.MsgCount)
	}
	if savedSummary.Content != expectedSummary {
		t.Errorf("Summary.Content 预期 %q，得到 %q", expectedSummary, savedSummary.Content)
	}
}

// TestSummarize_NoMessages 当天无消息，返回空字符串，不调 LLM
func TestSummarize_NoMessages(t *testing.T) {
	groupID := "test-group-empty"

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return []*message.Message{}, nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			t.Fatal("无消息时应不调用 Save")
			return nil
		},
	}
	llmCalled := false
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			llmCalled = true
			return "", nil
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err != nil {
		t.Fatalf("预期无错误，得到: %v", err)
	}
	if result != "" {
		t.Errorf("预期空字符串，得到 %q", result)
	}
	if llmCalled {
		t.Error("LLM 不应被调用")
	}
}

// TestSummarize_FewMessages 只有 3 条消息也正常总结
func TestSummarize_FewMessages(t *testing.T) {
	groupID := "test-group-few"
	expectedSummary := "简短总结。"
	var savedSummary *summary.Summary

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return generateTestMessages(3, gid), nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			savedSummary = s
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			return expectedSummary, nil
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err != nil {
		t.Fatalf("预期无错误，得到: %v", err)
	}
	if result != expectedSummary {
		t.Errorf("预期 %q，得到 %q", expectedSummary, result)
	}
	if savedSummary == nil {
		t.Fatal("Save 应被调用")
	}
	if savedSummary.MsgCount != 3 {
		t.Errorf("Summary.MsgCount 预期 3，得到 %d", savedSummary.MsgCount)
	}
}

// TestSummarize_LLMFailure LLM 返回 error，SummaryService 应传递该错误
func TestSummarize_LLMFailure(t *testing.T) {
	groupID := "test-group-llm-err"
	expectedErr := errors.New("LLM 服务异常")

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return generateTestMessages(1, gid), nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			t.Fatal("LLM 失败时不应用调用 Save")
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			return "", expectedErr
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err == nil {
		t.Fatal("预期有错误，得到 nil")
	}
	if result != "" {
		t.Errorf("预期空字符串，得到 %q", result)
	}
	// 验证传递了原始错误（或包含该错误的包装错误）
	if !errors.Is(err, expectedErr) && err.Error() != expectedErr.Error() {
		t.Errorf("预期错误包含 %v，得到 %v", expectedErr, err)
	}
}

// TestSummarize_SaveFailure summaryRepo.Save 失败时返回错误
func TestSummarize_SaveFailure(t *testing.T) {
	groupID := "test-group-save-err"
	expectedErr := errors.New("保存失败")

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return generateTestMessages(1, gid), nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			return expectedErr
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			return "some summary", nil
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err == nil {
		t.Fatal("预期有错误，得到 nil")
	}
	if result != "" {
		t.Errorf("预期空字符串，得到 %q", result)
	}
	// 验证传递了原始错误
	if !errors.Is(err, expectedErr) && err.Error() != expectedErr.Error() {
		t.Errorf("预期错误包含 %v，得到 %v", expectedErr, err)
	}
}

// TestSummarize_ReturnsContent 验证返回的摘要文本不空
func TestSummarize_ReturnsContent(t *testing.T) {
	groupID := "test-group-content"
	expectedSummary := "群聊包含50条消息的详细总结内容。"

	msgRepo := &mockMessageRepo{
		queryByDateFunc: func(gid string, date time.Time) ([]*message.Message, error) {
			return generateTestMessages(50, gid), nil
		},
	}
	summaryRepo := &mockSummaryRepo{
		saveFunc: func(s *summary.Summary) error {
			return nil
		},
	}
	llmMock := &mockLLM{
		chatFunc: func(msgs []llm.Message) (string, error) {
			return expectedSummary, nil
		},
	}

	svc := NewSummaryService(msgRepo, summaryRepo, llmMock)
	result, err := svc.Summarize(groupID)
	if err != nil {
		t.Fatalf("预期无错误，得到: %v", err)
	}
	if result == "" {
		t.Error("返回的摘要不应为空")
	}
	// 验证返回结果就是 LLM 返回的内容
	if result != expectedSummary {
		t.Errorf("预期 %q，得到 %q", expectedSummary, result)
	}
}
