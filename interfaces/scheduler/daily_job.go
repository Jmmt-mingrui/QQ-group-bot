// Package scheduler 定时任务触发器
package scheduler

// 默认压缩参数
const (
	defaultBeforeDays = 7  // 压缩 7 天前的消息
	defaultLimit      = 50 // 最多压缩 50 条
)

// SummaryService 总结服务接口
// 生产环境由 application.SummaryService 实现
type SummaryService interface {
	Summarize(groupID string) (string, error)
}

// CompactService 压缩服务接口
// 生产环境由 application.CompactService 实现
type CompactService interface {
	Compact(groupID string, beforeDays int, limit int) (int, error)
}

// DailyJob 每日定时任务
// 在指定小时触发：对每个群依次执行总结 → 压缩
type DailyJob struct {
	summaryService SummaryService
	compactService CompactService
	hour           int
	groupIDs       []string
}

// NewDailyJob 创建 DailyJob 实例
//   - summary: 总结服务
//   - compact: 压缩服务
//   - hour: 触发时间（小时）
//   - groupIDs: 要处理的群列表
func NewDailyJob(summary SummaryService, compact CompactService, hour int, groupIDs []string) *DailyJob {
	return &DailyJob{
		summaryService: summary,
		compactService: compact,
		hour:           hour,
		groupIDs:       groupIDs,
	}
}

// Run 依次对每个群执行 Summarize + Compact
// 某群总结失败不中断流程，错误被收集并一起返回
func (j *DailyJob) Run() []error {
	var errs []error

	for _, groupID := range j.groupIDs {
		// 第一步：总结
		_, err := j.summaryService.Summarize(groupID)
		if err != nil {
			errs = append(errs, err)
			continue // 总结失败，跳过压缩
		}

		// 第二步：压缩旧消息
		_, err = j.compactService.Compact(groupID, defaultBeforeDays, defaultLimit)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Hour 返回配置的触发小时
func (j *DailyJob) Hour() int {
	return j.hour
}

// GroupIDs 返回配置的群 ID 列表
func (j *DailyJob) GroupIDs() []string {
	return j.groupIDs
}
