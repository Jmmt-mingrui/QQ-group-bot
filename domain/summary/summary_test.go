package summary

import (
	"testing"
	"time"
)

// TestSummary_Create 测试创建一条总结，所有字段正确
func TestSummary_Create(t *testing.T) {
	now := time.Now()
	s := &Summary{
		ID:        "sum_001",
		GroupID:   "group_abc",
		Date:      "2026-05-20",
		Content:   "今日群聊主要讨论了项目进展和技术方案。",
		MsgCount:  42,
		CreatedAt: now,
	}

	if s.ID != "sum_001" {
		t.Errorf("ID = %q, want %q", s.ID, "sum_001")
	}
	if s.GroupID != "group_abc" {
		t.Errorf("GroupID = %q, want %q", s.GroupID, "group_abc")
	}
	if s.Date != "2026-05-20" {
		t.Errorf("Date = %q, want %q", s.Date, "2026-05-20")
	}
	if s.Content != "今日群聊主要讨论了项目进展和技术方案。" {
		t.Errorf("Content = %q, want %q", s.Content, "今日群聊主要讨论了项目进展和技术方案。")
	}
	if s.MsgCount != 42 {
		t.Errorf("MsgCount = %d, want %d", s.MsgCount, 42)
	}
	if !s.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", s.CreatedAt, now)
	}
}

// TestSummary_MsgCount 测试 MsgCount 反映原始消息数量
func TestSummary_MsgCount(t *testing.T) {
	s := &Summary{MsgCount: 128}

	if s.MsgCount != 128 {
		t.Errorf("MsgCount = %d, want %d", s.MsgCount, 128)
	}
}

// TestSummary_DateFormat 测试 Date 格式为 yyyy-mm-dd
func TestSummary_DateFormat(t *testing.T) {
	s := &Summary{Date: "2026-05-20"}

	if len(s.Date) != 10 {
		t.Errorf("Date length = %d, want 10 (yyyy-mm-dd)", len(s.Date))
	}
	if s.Date[4] != '-' || s.Date[7] != '-' {
		t.Errorf("Date format = %q, want yyyy-mm-dd", s.Date)
	}
}

// TestSummary_ContentEmpty 测试 Content 可以为空（没有可总结的内容时）
func TestSummary_ContentEmpty(t *testing.T) {
	s := &Summary{
		ID:       "sum_002",
		GroupID:  "group_def",
		Date:     "2026-05-20",
		Content:  "",
		MsgCount: 0,
	}

	if s.Content != "" {
		t.Errorf("Content = %q, want empty string", s.Content)
	}
	if s.MsgCount != 0 {
		t.Errorf("MsgCount = %d, want 0", s.MsgCount)
	}
}

// TestSummary_CreatedAtDefault 测试 CreatedAt 默认为当前时间
func TestSummary_CreatedAtDefault(t *testing.T) {
	s := NewSummary("sum_003", "group_ghi", "2026-05-20", "", 0)

	// CreatedAt 不应为零值（期望构造函数或初始化逻辑设置了当前时间）
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt 不应为零值，应默认为当前时间")
	}

	// 检查 CreatedAt 是否在合理的时间范围内（前后5分钟）
	now := time.Now()
	diff := s.CreatedAt.Sub(now)
	if diff < -5*time.Minute || diff > 5*time.Minute {
		t.Errorf("CreatedAt = %v, 与当前时间 %v 的差异超出5分钟", s.CreatedAt, now)
	}
}
