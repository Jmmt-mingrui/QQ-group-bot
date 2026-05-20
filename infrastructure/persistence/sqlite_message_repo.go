package persistence

import (
	"database/sql"
	"fmt"
	"time"

	"qq-group-bot/domain/message"
)

// sqliteMessageRepo 是 message.Repository 的 SQLite 实现
type sqliteMessageRepo struct {
	db *sql.DB
}

// NewSQLiteMessageRepo 创建 SQLite 消息仓储
func NewSQLiteMessageRepo(db *sql.DB) *sqliteMessageRepo {
	return &sqliteMessageRepo{db: db}
}

// sqlOpen 包装 sql.Open，方便测试中统一引包
func sqlOpen(driverName, dataSourceName string) (*sql.DB, error) {
	return sql.Open(driverName, dataSourceName)
}

// Save 保存一条消息
func (r *sqliteMessageRepo) Save(msg *message.Message) error {
	_, err := r.db.Exec(
		`INSERT INTO messages (id, group_id, user_id, nickname, content, timestamp, msg_type, compacted)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.GroupID, msg.UserID, msg.Nickname, msg.Content,
		msg.Timestamp.Unix(), msg.MsgType, boolToInt(msg.Compacted),
	)
	return err
}

// QueryByDate 按群组和日期查询消息
func (r *sqliteMessageRepo) QueryByDate(groupID string, date time.Time) ([]*message.Message, error) {
	dateStr := date.Format("2006-01-02")

	rows, err := r.db.Query(
		`SELECT id, group_id, user_id, nickname, content, timestamp, msg_type, compacted
		 FROM messages
		 WHERE group_id = ? AND date(timestamp, 'unixepoch') = ?
		 ORDER BY timestamp ASC`,
		groupID, dateStr,
	)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	defer rows.Close()

	return scanMessages(rows)
}

// QueryOldMessages 查询指定时间之前的消息，最多返回 limit 条
func (r *sqliteMessageRepo) QueryOldMessages(groupID string, before time.Time, limit int) ([]*message.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, group_id, user_id, nickname, content, timestamp, msg_type, compacted
		 FROM messages
		 WHERE group_id = ? AND timestamp < ?
		 ORDER BY timestamp DESC
		 LIMIT ?`,
		groupID, before.Unix(), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询旧消息失败: %w", err)
	}
	defer rows.Close()

	return scanMessages(rows)
}

// Search 在指定群组中按关键词搜索消息（大小写敏感）
func (r *sqliteMessageRepo) Search(groupID string, keyword string) ([]*message.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, group_id, user_id, nickname, content, timestamp, msg_type, compacted
		 FROM messages
		 WHERE group_id = ? AND content LIKE '%' || ? || '%'
		 ORDER BY timestamp ASC`,
		groupID, keyword,
	)
	if err != nil {
		return nil, fmt.Errorf("搜索消息失败: %w", err)
	}
	defer rows.Close()

	return scanMessages(rows)
}

// Replace 删除旧消息（按 ID 列表）并插入一条压缩后的消息（事务）
func (r *sqliteMessageRepo) Replace(oldIDs []string, compacted *message.Message) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	for _, id := range oldIDs {
		_, err := tx.Exec("DELETE FROM messages WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("删除旧消息 %q 失败: %w", id, err)
		}
	}

	_, err = tx.Exec(
		`INSERT INTO messages (id, group_id, user_id, nickname, content, timestamp, msg_type, compacted)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		compacted.ID, compacted.GroupID, compacted.UserID, compacted.Nickname, compacted.Content,
		compacted.Timestamp.Unix(), compacted.MsgType, boolToInt(compacted.Compacted),
	)
	if err != nil {
		return fmt.Errorf("插入压缩消息失败: %w", err)
	}

	return tx.Commit()
}

// scanMessages 将 SQL 行扫描为 Message 切片
func scanMessages(rows *sql.Rows) ([]*message.Message, error) {
	var result []*message.Message
	for rows.Next() {
		var (
			msg       message.Message
			unixTS    int64
			compacted int
		)
		err := rows.Scan(&msg.ID, &msg.GroupID, &msg.UserID, &msg.Nickname,
			&msg.Content, &unixTS, &msg.MsgType, &compacted)
		if err != nil {
			return nil, fmt.Errorf("扫描消息行失败: %w", err)
		}
		msg.Timestamp = time.Unix(unixTS, 0)
		msg.Compacted = compacted != 0
		result = append(result, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息行失败: %w", err)
	}
	if result == nil {
		return []*message.Message{}, nil
	}
	return result, nil
}

// boolToInt 将 bool 转为 0/1，用于 SQLite 存储
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
