package message

import (
	"strings"
	"testing"
	"time"
)

// mockRepo 是 Repository 接口的 mock 实现
type mockRepo struct {
	messages []*Message
}

func (m *mockRepo) Save(msg *Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockRepo) QueryByDate(groupID string, date time.Time) ([]*Message, error) {
	var result []*Message
	for _, msg := range m.messages {
		if msg.GroupID == groupID && isSameDate(msg.Timestamp, date) {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *mockRepo) QueryOldMessages(groupID string, before time.Time, limit int) ([]*Message, error) {
	var result []*Message
	for _, msg := range m.messages {
		if msg.GroupID == groupID && msg.Timestamp.Before(before) {
			result = append(result, msg)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepo) Search(groupID string, keyword string) ([]*Message, error) {
	var result []*Message
	for _, msg := range m.messages {
		if msg.GroupID == groupID && strings.Contains(msg.Content, keyword) {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *mockRepo) Replace(oldIDs []string, compacted *Message) error {
	oldSet := make(map[string]bool)
	for _, id := range oldIDs {
		oldSet[id] = true
	}
	var kept []*Message
	for _, msg := range m.messages {
		if !oldSet[msg.ID] {
			kept = append(kept, msg)
		}
	}
	m.messages = kept
	m.messages = append(m.messages, compacted)
	return nil
}

// 辅助函数：判断两个 time.Time 是否在同一天
func isSameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// ---------------------------------------------------------------------------
// 测试用例
// ---------------------------------------------------------------------------

// Save 一条消息后能 QueryByDate 查到
func TestSaveAndQueryByDate(t *testing.T) {
	repo := &mockRepo{}
	now := time.Now()
	msg := &Message{
		ID:        "msg_001",
		GroupID:   "group_123",
		UserID:    "user_456",
		Content:   "测试消息",
		MsgType:   "text",
		Timestamp: now,
	}

	err := repo.Save(msg)
	if err != nil {
		t.Fatalf("Save() 返回错误: %v", err)
	}

	msgs, err := repo.QueryByDate("group_123", now)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("QueryByDate() 返回 %d 条消息, want 1", len(msgs))
	}
	if msgs[0].ID != "msg_001" {
		t.Errorf("QueryByDate()[0].ID = %q, want %q", msgs[0].ID, "msg_001")
	}
}

// QueryByDate 只返回指定群和日期的消息
func TestQueryByDateFiltersCorrectly(t *testing.T) {
	repo := &mockRepo{}
	today := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)

	repo.Save(&Message{ID: "msg_001", GroupID: "group_123", Content: "今天群1", Timestamp: today})
	repo.Save(&Message{ID: "msg_002", GroupID: "group_456", Content: "今天群2", Timestamp: today})
	repo.Save(&Message{ID: "msg_003", GroupID: "group_123", Content: "昨天群1", Timestamp: yesterday})

	// 查询 group_123 今天的消息 — 应该只返回 msg_001
	msgs, err := repo.QueryByDate("group_123", today)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("QueryByDate(group_123, today) 返回 %d 条消息, want 1", len(msgs))
	}
	if msgs[0].ID != "msg_001" {
		t.Errorf("返回的消息 ID = %q, want %q", msgs[0].ID, "msg_001")
	}
}

// QueryOldMessages 只返回 before 之前的消息，限制 limit
func TestQueryOldMessagesLimit(t *testing.T) {
	repo := &mockRepo{}
	now := time.Now()
	earlier := now.Add(-2 * time.Hour)

	// 存入 10 条旧消息
	for i := 0; i < 10; i++ {
		repo.Save(&Message{
			ID:        "msg_001",
			GroupID:   "group_123",
			Content:   "旧消息",
			Timestamp: earlier,
		})
	}

	limit := 3
	msgs, err := repo.QueryOldMessages("group_123", now, limit)
	if err != nil {
		t.Fatalf("QueryOldMessages() 返回错误: %v", err)
	}
	if len(msgs) > limit {
		t.Errorf("QueryOldMessages 返回 %d 条消息, 限制为 %d", len(msgs), limit)
	}
}

// Search 按关键词搜索返回匹配消息
func TestSearchByKeyword(t *testing.T) {
	repo := &mockRepo{}
	now := time.Now()

	repo.Save(&Message{ID: "msg_001", GroupID: "group_123", Content: "今天天气真好", Timestamp: now})
	repo.Save(&Message{ID: "msg_002", GroupID: "group_123", Content: "明天天气也不错", Timestamp: now})
	repo.Save(&Message{ID: "msg_003", GroupID: "group_123", Content: "晚上要下雨", Timestamp: now})

	msgs, err := repo.Search("group_123", "天气")
	if err != nil {
		t.Fatalf("Search() 返回错误: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("Search('天气') 返回 %d 条消息, want 2", len(msgs))
	}
}

// Replace 删除旧消息并插入压缩后的消息
func TestReplace(t *testing.T) {
	repo := &mockRepo{}
	now := time.Now()

	repo.Save(&Message{ID: "msg_001", GroupID: "group_123", Content: "第一条", Timestamp: now})
	repo.Save(&Message{ID: "msg_002", GroupID: "group_123", Content: "第二条", Timestamp: now})
	repo.Save(&Message{ID: "msg_003", GroupID: "group_123", Content: "第三条", Timestamp: now})

	compacted := &Message{
		ID:        "compacted_001",
		GroupID:   "group_123",
		Content:   "压缩后的摘要",
		Compacted: true,
		Timestamp: now,
	}

	err := repo.Replace([]string{"msg_001", "msg_002"}, compacted)
	if err != nil {
		t.Fatalf("Replace() 返回错误: %v", err)
	}

	// 查询剩余消息
	remaining, err := repo.QueryByDate("group_123", now)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}

	// 验证旧消息被删除
	for _, msg := range remaining {
		if msg.ID == "msg_001" || msg.ID == "msg_002" {
			t.Errorf("旧消息 %q 应该被删除，但仍存在", msg.ID)
		}
	}

	// 验证压缩消息被插入
	found := false
	for _, msg := range remaining {
		if msg.ID == "compacted_001" {
			found = true
			if !msg.Compacted {
				t.Error("压缩消息的 Compacted 应该为 true")
			}
			if msg.Content != "压缩后的摘要" {
				t.Errorf("压缩消息 Content = %q, want %q", msg.Content, "压缩后的摘要")
			}
			break
		}
	}
	if !found {
		t.Error("压缩消息未在查询结果中找到")
	}

	// 总数应为: 3条原始 - 2条删除 + 1条压缩 = 2条
	if len(remaining) != 2 {
		t.Errorf("Replace 后共有 %d 条消息, want 2", len(remaining))
	}
}

// 查询不存在的日期返回空列表不报错
func TestQueryNonExistentDate(t *testing.T) {
	repo := &mockRepo{}
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// 先存入一条其他日期的消息
	repo.Save(&Message{
		ID:        "msg_001",
		GroupID:   "group_123",
		Content:   "今天的消息",
		Timestamp: time.Now(),
	})

	msgs, err := repo.QueryByDate("group_123", future)
	if err != nil {
		t.Fatalf("查询不存在日期返回错误: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("查询不存在日期应返回 0 条消息, 得到 %d", len(msgs))
	}
}
