package summary

import (
	"fmt"
	"testing"
)

// mockRepository 是 Repository 接口的 mock 实现，用于接口测试
type mockRepository struct {
	summaries []*Summary
}

// Save 保存总结，若已存在相同群组和日期的记录则替换
func (m *mockRepository) Save(s *Summary) error {
	for i, existing := range m.summaries {
		if existing.GroupID == s.GroupID && existing.Date == s.Date {
			m.summaries[i] = s
			return nil
		}
	}
	m.summaries = append(m.summaries, s)
	return nil
}

// QueryByDate 按群组和日期查询总结
func (m *mockRepository) QueryByDate(groupID string, date string) (*Summary, error) {
	for _, s := range m.summaries {
		if s.GroupID == groupID && s.Date == date {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

// ListByGroup 按群组列出总结，按 limit 限制返回条数
func (m *mockRepository) ListByGroup(groupID string, limit int) ([]*Summary, error) {
	var result []*Summary
	for _, s := range m.summaries {
		if s.GroupID == groupID {
			result = append(result, s)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	if result == nil {
		return []*Summary{}, nil
	}
	return result, nil
}

// TestRepository_SaveAndQueryByDate 测试 Save 后 QueryByDate 能查到
func TestRepository_SaveAndQueryByDate(t *testing.T) {
	repo := &mockRepository{}
	s := &Summary{
		ID:      "sum_001",
		GroupID: "group_abc",
		Date:    "2026-05-20",
		Content: "测试总结",
	}

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
}

// TestRepository_QueryByDate_NotFound 测试查询不存在的日期返回 nil 和 ErrNotFound
func TestRepository_QueryByDate_NotFound(t *testing.T) {
	repo := &mockRepository{}

	got, err := repo.QueryByDate("group_empty", "2026-05-20")
	if err != ErrNotFound {
		t.Errorf("err = %v, want %v", err, ErrNotFound)
	}
	if got != nil {
		t.Errorf("got = %v, want nil", got)
	}
}

// TestRepository_ListByGroup_Limit 测试 ListByGroup 按 limit 限制返回条数
func TestRepository_ListByGroup_Limit(t *testing.T) {
	repo := &mockRepository{}

	// 保存5条总结，然后限制返回3条
	for i := 1; i <= 5; i++ {
		date := fmt.Sprintf("2026-05-%02d", i)
		err := repo.Save(&Summary{
			ID:      fmt.Sprintf("sum_%02d", i),
			GroupID: "group_limit",
			Date:    date,
			Content: "test",
		})
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

// TestRepository_ListByGroup_Empty 测试空群返回空列表（非 nil）
func TestRepository_ListByGroup_Empty(t *testing.T) {
	repo := &mockRepository{}

	summaries, err := repo.ListByGroup("group_empty", 10)
	if err != nil {
		t.Fatalf("ListByGroup 失败: %v", err)
	}
	if summaries == nil {
		t.Error("返回了 nil, 应返回空切片 []*Summary{}")
	}
	if len(summaries) != 0 {
		t.Errorf("len(summaries) = %d, want 0", len(summaries))
	}
}
