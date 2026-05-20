package persistence

import (
	"testing"

	"qq-group-bot/domain/summary"

	_ "modernc.org/sqlite"
)

// setupSummaryDB 创建内存 SQLite 数据库，建表，返回 Repository 实例
func setupSummaryDB(t *testing.T) *sqliteSummaryRepo {
	t.Helper()

	db, err := sqlOpen("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS summaries (
			id        TEXT PRIMARY KEY,
			group_id  TEXT NOT NULL,
			date      TEXT NOT NULL,
			content   TEXT NOT NULL,
			msg_count INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			UNIQUE(group_id, date)
		)
	`)
	if err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	return NewSQLiteSummaryRepo(db)
}

// TestSummaryRepo_SaveAndQueryByDate Save 后 QueryByDate 能查到
func TestSummaryRepo_SaveAndQueryByDate(t *testing.T) {
	repo := setupSummaryDB(t)

	s := summary.NewSummary("sum_001", "group_abc", "2026-05-20", "测试总结", 10)

	err := repo.Save(s)
	if err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	got, err := repo.QueryByDate("group_abc", "2026-05-20")
	if err != nil {
		t.Fatalf("QueryByDate 失败: %v", err)
	}
	if got == nil {
		t.Fatal("QueryByDate 返回 nil，期望返回总结")
	}
	if got.ID != "sum_001" {
		t.Errorf("got.ID = %q, want %q", got.ID, "sum_001")
	}
	if got.Content != "测试总结" {
		t.Errorf("got.Content = %q, want %q", got.Content, "测试总结")
	}
	if got.MsgCount != 10 {
		t.Errorf("got.MsgCount = %d, want 10", got.MsgCount)
	}
}

// TestSummaryRepo_QueryByDate_NotFound 查询不存在的日期返回 ErrNotFound
func TestSummaryRepo_QueryByDate_NotFound(t *testing.T) {
	repo := setupSummaryDB(t)

	got, err := repo.QueryByDate("group_empty", "2026-05-20")
	if err != summary.ErrNotFound {
		t.Errorf("err = %v, want %v", err, summary.ErrNotFound)
	}
	if got != nil {
		t.Errorf("got = %v, want nil", got)
	}
}

// TestSummaryRepo_Save_Update 保存相同群组日期的总结会自动替换
func TestSummaryRepo_Save_Update(t *testing.T) {
	repo := setupSummaryDB(t)

	s1 := summary.NewSummary("sum_001", "group_abc", "2026-05-20", "原始总结", 5)
	err := repo.Save(s1)
	if err != nil {
		t.Fatalf("第一次 Save 失败: %v", err)
	}

	s2 := summary.NewSummary("sum_002", "group_abc", "2026-05-20", "更新后的总结", 10)
	err = repo.Save(s2)
	if err != nil {
		t.Fatalf("第二次 Save 失败: %v", err)
	}

	got, err := repo.QueryByDate("group_abc", "2026-05-20")
	if err != nil {
		t.Fatalf("QueryByDate 失败: %v", err)
	}
	if got.Content != "更新后的总结" {
		t.Errorf("got.Content = %q, want %q", got.Content, "更新后的总结")
	}
	if got.MsgCount != 10 {
		t.Errorf("got.MsgCount = %d, want 10", got.MsgCount)
	}
}

// TestSummaryRepo_ListByGroup_Limit ListByGroup 按 limit 限制返回条数
func TestSummaryRepo_ListByGroup_Limit(t *testing.T) {
	repo := setupSummaryDB(t)

	// 保存 5 条总结
	for i := 1; i <= 5; i++ {
		date := "2026-05-" + twoDigit(i)
		s := summary.NewSummary(
			"sum_"+twoDigit(i),
			"group_limit",
			date,
			"test",
			0,
		)
		err := repo.Save(s)
		if err != nil {
			t.Fatalf("Save 第%d条失败: %v", i, err)
		}
	}

	summaries, err := repo.ListByGroup("group_limit", 3)
	if err != nil {
		t.Fatalf("ListByGroup 失败: %v", err)
	}
	if len(summaries) != 3 {
		t.Errorf("len(summaries) = %d, want 3", len(summaries))
	}
}

// TestSummaryRepo_ListByGroup_Empty 空群返回空列表
func TestSummaryRepo_ListByGroup_Empty(t *testing.T) {
	repo := setupSummaryDB(t)

	summaries, err := repo.ListByGroup("group_empty", 10)
	if err != nil {
		t.Fatalf("ListByGroup 失败: %v", err)
	}
	if summaries == nil {
		t.Error("返回了 nil, 应返回空切片")
	}
	if len(summaries) != 0 {
		t.Errorf("len(summaries) = %d, want 0", len(summaries))
	}
}

// twoDigit 将整数格式化为至少两位的字符串
func twoDigit(i int) string {
	if i < 10 {
		return "0" + string(rune('0'+i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}
