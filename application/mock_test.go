package application

import (
	"time"

	"qq-group-bot/domain/llm"
	"qq-group-bot/domain/message"
	"qq-group-bot/domain/summary"
)

// mockMessageRepo 模拟 message.Repository，支持所有测试文件的用例
type mockMessageRepo struct {
	// chat_service 用到的字段
	messages []*message.Message
	saveErr  error // saveErr 不为 nil 时 Save 返回此错误

	// ask_service 用到的字段
	searchFunc func(groupID, keyword string) ([]*message.Message, error)

	// compact_service 用到的字段
	queryOldMessagesFunc func(groupID string, before time.Time, limit int) ([]*message.Message, error)
	replaceFunc          func(oldIDs []string, compacted *message.Message) error

	// summary_service 用到的字段
	queryByDateFunc func(groupID string, date time.Time) ([]*message.Message, error)
}

func (m *mockMessageRepo) Save(msg *message.Message) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockMessageRepo) QueryByDate(groupID string, date time.Time) ([]*message.Message, error) {
	if m.queryByDateFunc != nil {
		return m.queryByDateFunc(groupID, date)
	}
	return nil, nil
}

func (m *mockMessageRepo) QueryOldMessages(groupID string, before time.Time, limit int) ([]*message.Message, error) {
	if m.queryOldMessagesFunc != nil {
		return m.queryOldMessagesFunc(groupID, before, limit)
	}
	return nil, nil
}

func (m *mockMessageRepo) Search(groupID string, keyword string) ([]*message.Message, error) {
	if m.searchFunc != nil {
		return m.searchFunc(groupID, keyword)
	}
	return nil, nil
}

func (m *mockMessageRepo) Replace(oldIDs []string, compacted *message.Message) error {
	if m.replaceFunc != nil {
		return m.replaceFunc(oldIDs, compacted)
	}
	return nil
}

// mockLLM 模拟 llm.Provider，支持所有测试文件的用例
type mockLLM struct {
	chatFunc func(messages []llm.Message) (string, error)
}

func (m *mockLLM) ChatCompletion(messages []llm.Message) (string, error) {
	if m.chatFunc != nil {
		return m.chatFunc(messages)
	}
	return "", nil
}

// mockSummaryRepo 模拟 summary.Repository，支持所有测试文件的用例
type mockSummaryRepo struct {
	saveFunc func(s *summary.Summary) error
}

func (m *mockSummaryRepo) Save(s *summary.Summary) error {
	if m.saveFunc != nil {
		return m.saveFunc(s)
	}
	return nil
}

func (m *mockSummaryRepo) QueryByDate(groupID, date string) (*summary.Summary, error) {
	return nil, summary.ErrNotFound
}

func (m *mockSummaryRepo) ListByGroup(groupID string, limit int) ([]*summary.Summary, error) {
	return nil, nil
}
