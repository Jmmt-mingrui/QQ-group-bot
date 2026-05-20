package persistence

import (
	"database/sql"
	"fmt"
	"time"

	"qq-group-bot/domain/summary"
)

// sqliteSummaryRepo 是 summary.Repository 的 SQLite 实现
type sqliteSummaryRepo struct {
	db *sql.DB
}

// NewSQLiteSummaryRepo 创建 SQLite 总结仓储
func NewSQLiteSummaryRepo(db *sql.DB) *sqliteSummaryRepo {
	return &sqliteSummaryRepo{db: db}
}

// Save 保存总结，若已存在相同群组和日期的记录则替换
func (r *sqliteSummaryRepo) Save(s *summary.Summary) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO summaries (id, group_id, date, content, msg_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.GroupID, s.Date, s.Content, s.MsgCount, s.CreatedAt.Unix(),
	)
	return err
}

// QueryByDate 按群组和日期查询总结
func (r *sqliteSummaryRepo) QueryByDate(groupID string, date string) (*summary.Summary, error) {
	row := r.db.QueryRow(
		`SELECT id, group_id, date, content, msg_count, created_at
		 FROM summaries
		 WHERE group_id = ? AND date = ?`,
		groupID, date,
	)

	var (
		s        summary.Summary
		unixTS   int64
	)
	err := row.Scan(&s.ID, &s.GroupID, &s.Date, &s.Content, &s.MsgCount, &unixTS)
	if err == sql.ErrNoRows {
		return nil, summary.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询总结失败: %w", err)
	}

	s.CreatedAt = time.Unix(unixTS, 0)
	return &s, nil
}

// ListByGroup 按群组列出总结，按 limit 限制返回条数
func (r *sqliteSummaryRepo) ListByGroup(groupID string, limit int) ([]*summary.Summary, error) {
	rows, err := r.db.Query(
		`SELECT id, group_id, date, content, msg_count, created_at
		 FROM summaries
		 WHERE group_id = ?
		 ORDER BY date DESC
		 LIMIT ?`,
		groupID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("列出总结失败: %w", err)
	}
	defer rows.Close()

	var result []*summary.Summary
	for rows.Next() {
		var (
			s      summary.Summary
			unixTS int64
		)
		err := rows.Scan(&s.ID, &s.GroupID, &s.Date, &s.Content, &s.MsgCount, &unixTS)
		if err != nil {
			return nil, fmt.Errorf("扫描总结行失败: %w", err)
		}
		s.CreatedAt = time.Unix(unixTS, 0)
		result = append(result, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历总结行失败: %w", err)
	}

	if result == nil {
		return []*summary.Summary{}, nil
	}
	return result, nil
}
