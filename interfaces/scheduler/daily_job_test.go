// Package scheduler 定时任务触发器
package scheduler

import (
	"errors"
	"testing"
)

// ---------- mock 定义 ----------

// mockSummaryService 模拟 SummaryService
type mockSummaryService struct {
	// 记录对每个 groupID 的调用次数
	summarizeCalls map[string]int
	// 控制每次调用的返回值（按 groupID）
	summarizeResult map[string]string
	summarizeErr    map[string]error
}

func newMockSummaryService() *mockSummaryService {
	return &mockSummaryService{
		summarizeCalls:  make(map[string]int),
		summarizeResult: make(map[string]string),
		summarizeErr:    make(map[string]error),
	}
}

func (m *mockSummaryService) Summarize(groupID string) (string, error) {
	m.summarizeCalls[groupID]++
	result := m.summarizeResult[groupID]
	err := m.summarizeErr[groupID]
	return result, err
}

// mockCompactService 模拟 CompactService
type mockCompactService struct {
	// 记录对每个 groupID 的调用次数
	compactCalls map[string]int
	// 控制返回值
	compactResult map[string]int
	compactErr    map[string]error
}

func newMockCompactService() *mockCompactService {
	return &mockCompactService{
		compactCalls:   make(map[string]int),
		compactResult:  make(map[string]int),
		compactErr:     make(map[string]error),
	}
}

func (m *mockCompactService) Compact(groupID string, beforeDays int, limit int) (int, error) {
	m.compactCalls[groupID]++
	result := m.compactResult[groupID]
	err := m.compactErr[groupID]
	return result, err
}

// ---------- 测试用例 ----------

// TestNewDailyJob NewDailyJob 创建成功,验证基本字段
func TestNewDailyJob(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()
	hour := 23
	groupIDs := []string{"111", "222"}

	job := NewDailyJob(summary, compact, hour, groupIDs)
	if job == nil {
		t.Fatal("NewDailyJob 返回 nil，期望非 nil")
	}
	if job.Hour() != hour {
		t.Errorf("Hour() = %d, 期望 %d", job.Hour(), hour)
	}
	gotIDs := job.GroupIDs()
	if len(gotIDs) != len(groupIDs) {
		t.Fatalf("GroupIDs() 长度 = %d, 期望 %d", len(gotIDs), len(groupIDs))
	}
	for i, id := range groupIDs {
		if gotIDs[i] != id {
			t.Errorf("GroupIDs()[%d] = %s, 期望 %s", i, gotIDs[i], id)
		}
	}
}

// TestRunNow Run() 触发总结，验证 SummaryService.Summarize 被调用
func TestRunNow(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	// 配置 Mock：Summarize 成功返回
	summary.summarizeResult["111"] = "今日总结摘要"
	// Compact 成功
	compact.compactResult["111"] = 10

	groupIDs := []string{"111"}
	job := NewDailyJob(summary, compact, 23, groupIDs)

	errs := job.Run()
	if len(errs) > 0 {
		t.Fatalf("Run() 返回错误: %v", errs)
	}

	// 验证 Summarize 被调用了
	if summary.summarizeCalls["111"] != 1 {
		t.Errorf("Summarize 调用次数 = %d, 期望 1", summary.summarizeCalls["111"])
	}
	// 验证 Compact 被调用了
	if compact.compactCalls["111"] != 1 {
		t.Errorf("Compact 调用次数 = %d, 期望 1", compact.compactCalls["111"])
	}
}

// TestRunAllGroups Run 对所有群执行，验证每个群都调用了 Summarize
func TestRunAllGroups(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	groupIDs := []string{"111", "222"}
	for _, id := range groupIDs {
		summary.summarizeResult[id] = id + "的总结"
		compact.compactResult[id] = 5
	}

	job := NewDailyJob(summary, compact, 23, groupIDs)

	errs := job.Run()
	if len(errs) > 0 {
		t.Fatalf("Run() 返回错误: %v", errs)
	}

	// 验证两个群都被调用了 Summarize
	for _, id := range groupIDs {
		if summary.summarizeCalls[id] != 1 {
			t.Errorf("群 %s 的 Summarize 调用次数 = %d, 期望 1", id, summary.summarizeCalls[id])
		}
	}
	// 验证两个群都被调用了 Compact
	for _, id := range groupIDs {
		if compact.compactCalls[id] != 1 {
			t.Errorf("群 %s 的 Compact 调用次数 = %d, 期望 1", id, compact.compactCalls[id])
		}
	}
}

// TestRunPartialFailure 某群总结失败不中断，其余群继续执行
func TestRunPartialFailure(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	groupIDs := []string{"111", "222", "333"}

	// 群1：成功
	summary.summarizeResult["111"] = "群1总结"
	compact.compactResult["111"] = 5
	// 群2：总结失败
	summary.summarizeErr["222"] = errors.New("群2 LLM 调用超时")
	// 群3：成功（且不依赖 compact）
	summary.summarizeResult["333"] = "群3总结"
	compact.compactResult["333"] = 3

	job := NewDailyJob(summary, compact, 23, groupIDs)

	errs := job.Run()

	// 验证返回了 1 个错误
	if len(errs) != 1 {
		t.Fatalf("Run() 返回 %d 个错误, 期望 1 个", len(errs))
	}

	// 验证所有群都调了 Summarize（即使失败也调了）
	if summary.summarizeCalls["111"] != 1 {
		t.Errorf("群1 Summarize 调用次数 = %d, 期望 1", summary.summarizeCalls["111"])
	}
	if summary.summarizeCalls["222"] != 1 {
		t.Errorf("群2 Summarize 调用次数 = %d, 期望 1", summary.summarizeCalls["222"])
	}
	if summary.summarizeCalls["333"] != 1 {
		t.Errorf("群3 Summarize 调用次数 = %d, 期望 1", summary.summarizeCalls["333"])
	}

	// 群2 总结失败，不执行 Compact
	if compact.compactCalls["222"] != 0 {
		t.Errorf("群2 Compact 不应被调用, 被调了 %d 次", compact.compactCalls["222"])
	}

	// 群3 总结成功，Compact 应被调用
	if compact.compactCalls["333"] != 1 {
		t.Errorf("群3 Compact 调用次数 = %d, 期望 1", compact.compactCalls["333"])
	}
}

// TestRunCompact 验证总结完成后对每个群执行 Compact
func TestRunCompact(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	groupIDs := []string{"g1", "g2"}
	for _, id := range groupIDs {
		summary.summarizeResult[id] = id + "总结"
		compact.compactResult[id] = 3
	}

	job := NewDailyJob(summary, compact, 23, groupIDs)
	_ = job.Run()

	// 验证 Compact 被调用，且传入了默认参数 (beforeDays=7, limit=50)
	if compact.compactCalls["g1"] != 1 {
		t.Errorf("Compact 未对 g1 调用")
	}
	if compact.compactCalls["g2"] != 1 {
		t.Errorf("Compact 未对 g2 调用")
	}
}

// TestHour 验证 Hour() 返回配置的小时值
func TestHour(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	job := NewDailyJob(summary, compact, 8, []string{"111"})
	if job.Hour() != 8 {
		t.Errorf("Hour() = %d, 期望 8", job.Hour())
	}

	job2 := NewDailyJob(summary, compact, 23, []string{"111"})
	if job2.Hour() != 23 {
		t.Errorf("Hour() = %d, 期望 23", job2.Hour())
	}
}

// TestGroupIDs 验证 GroupIDs() 返回配置的群 ID 列表
func TestGroupIDs(t *testing.T) {
	summary := newMockSummaryService()
	compact := newMockCompactService()

	ids := []string{"a", "b", "c"}
	job := NewDailyJob(summary, compact, 23, ids)

	got := job.GroupIDs()
	if len(got) != 3 {
		t.Fatalf("GroupIDs() 长度 = %d, 期望 3", len(got))
	}
	for i, id := range ids {
		if got[i] != id {
			t.Errorf("GroupIDs()[%d] = %s, 期望 %s", i, got[i], id)
		}
	}
}
