package persistence

import (
	"testing"
	"time"

	"qq-group-bot/domain/message"

	_ "modernc.org/sqlite"
)

// setupMessageDB 创建内存 SQLite 数据库，建表，返回 Repository 实例
func setupMessageDB(t *testing.T) *sqliteMessageRepo {
	t.Helper()

	db, err := sqlOpen("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id        TEXT PRIMARY KEY,
			group_id  TEXT NOT NULL,
			user_id   TEXT NOT NULL DEFAULT '',
			nickname  TEXT NOT NULL DEFAULT '',
			content   TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			msg_type  TEXT NOT NULL DEFAULT 'text',
			compacted INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	return NewSQLiteMessageRepo(db)
}

// TestMessageRepo_SaveAndQueryByDate Save 后 QueryByDate 能查到
func TestMessageRepo_SaveAndQueryByDate(t *testing.T) {
	repo := setupMessageDB(t)
	now := time.Now()

	msg := &message.Message{
		ID:        "msg_001",
		GroupID:   "group_123",
		UserID:    "user_456",
		Nickname:  "测试用户",
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
		t.Errorf("msgs[0].ID = %q, want %q", msgs[0].ID, "msg_001")
	}
	if msgs[0].Content != "测试消息" {
		t.Errorf("msgs[0].Content = %q, want %q", msgs[0].Content, "测试消息")
	}
	if msgs[0].UserID != "user_456" {
		t.Errorf("msgs[0].UserID = %q, want %q", msgs[0].UserID, "user_456")
	}
	if msgs[0].Nickname != "测试用户" {
		t.Errorf("msgs[0].Nickname = %q, want %q", msgs[0].Nickname, "测试用户")
	}
	if msgs[0].MsgType != "text" {
		t.Errorf("msgs[0].MsgType = %q, want %q", msgs[0].MsgType, "text")
	}
}

// TestMessageRepo_QueryByDateFiltersCorrectly QueryByDate 多群多日期正确过滤
func TestMessageRepo_QueryByDateFiltersCorrectly(t *testing.T) {
	repo := setupMessageDB(t)
	today := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)

	repo.Save(&message.Message{ID: "msg_001", GroupID: "group_123", Content: "今天群1", Timestamp: today})
	repo.Save(&message.Message{ID: "msg_002", GroupID: "group_456", Content: "今天群2", Timestamp: today})
	repo.Save(&message.Message{ID: "msg_003", GroupID: "group_123", Content: "昨天群1", Timestamp: yesterday})

	// 查询 group_123 今天的消息，应该只返回 msg_001
	msgs, err := repo.QueryByDate("group_123", today)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("QueryByDate(group_123, today) 返回 %d 条, want 1", len(msgs))
	}
	if msgs[0].ID != "msg_001" {
		t.Errorf("返回的消息 ID = %q, want %q", msgs[0].ID, "msg_001")
	}

	// 查询 group_456 今天，应该返回 msg_002
	msgs2, err := repo.QueryByDate("group_456", today)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}
	if len(msgs2) != 1 {
		t.Fatalf("QueryByDate(group_456, today) 返回 %d 条, want 1", len(msgs2))
	}

	// 查询 group_123 昨天，应该返回 msg_003
	msgs3, err := repo.QueryByDate("group_123", yesterday)
	if err != nil {
		t.Fatalf("QueryByDate() 返回错误: %v", err)
	}
	if len(msgs3) != 1 {
		t.Fatalf("QueryByDate(group_123, yesterday) 返回 %d 条, want 1", len(msgs3))
	}
}

// TestMessageRepo_QueryOldMessagesLimit QueryOldMessages 按时间+limit 筛选
func TestMessageRepo_QueryOldMessagesLimit(t *testing.T) {
	repo := setupMessageDB(t)
	now := time.Now()
	earlier := now.Add(-2 * time.Hour)

	// 存入 10 条旧消息，每条不同 ID
	for i := 0; i < 10; i++ {
		repo.Save(&message.Message{
			ID:        time.Now().Add(time.Duration(i) * time.Second).Format("msg_150405.000"),
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
		t.Errorf("QueryOldMessages 返回 %d 条, 限制为 %d", len(msgs), limit)
	}
	if len(msgs) == 0 {
		t.Error("QueryOldMessages 返回 0 条, 期望 > 0")
	}
}

// TestMessageRepo_SearchByKeyword Search 关键词匹配
func TestMessageRepo_SearchByKeyword(t *testing.T) {
	repo := setupMessageDB(t)
	now := time.Now()

	repo.Save(&message.Message{ID: "msg_001", GroupID: "group_123", Content: "今天天气真好", Timestamp: now})
	repo.Save(&message.Message{ID: "msg_002", GroupID: "group_123", Content: "明天天气也不错", Timestamp: now})
	repo.Save(&message.Message{ID: "msg_003", GroupID: "group_123", Content: "晚上要下雨", Timestamp: now})

	msgs, err := repo.Search("group_123", "天气")
	if err != nil {
		t.Fatalf("Search() 返回错误: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("Search('天气') 返回 %d 条, want 2", len(msgs))
	}

	// 搜索不存在的关键词返回空
	msgs2, err := repo.Search("group_123", "不存在")
	if err != nil {
		t.Fatalf("Search() 返回错误: %v", err)
	}
	if len(msgs2) != 0 {
		t.Errorf("Search('不存在') 返回 %d 条, want 0", len(msgs2))
	}
}

// TestMessageRepo_Replace Replace 删除旧消息插入压缩消息
func TestMessageRepo_Replace(t *testing.T) {
	repo := setupMessageDB(t)
	now := time.Now()

	repo.Save(&message.Message{ID: "msg_001", GroupID: "group_123", Content: "第一条", Timestamp: now})
	repo.Save(&message.Message{ID: "msg_002", GroupID: "group_123", Content: "第二条", Timestamp: now})
	repo.Save(&message.Message{ID: "msg_003", GroupID: "group_123", Content: "第三条", Timestamp: now})

	compacted := &message.Message{
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

	// 总数: 3条原始 - 2条删除 + 1条压缩 = 2条
	if len(remaining) != 2 {
		t.Errorf("Replace 后共有 %d 条, want 2", len(remaining))
	}
}

// TestMessageRepo_QueryNonExistentDate 查询不存在的日期返回空列表
func TestMessageRepo_QueryNonExistentDate(t *testing.T) {
	repo := setupMessageDB(t)
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	// 先存入一条其他日期的消息
	repo.Save(&message.Message{
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
		t.Errorf("查询不存在日期应返回 0 条, 得到 %d", len(msgs))
	}
}
