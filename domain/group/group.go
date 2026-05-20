// Package group 定义群组领域模型
package group

// Group 表示一个 QQ 群组
// ID 为群号（唯一标识），Name 为群名称
type Group struct {
	ID   string // 群号，唯一标识
	Name string // 群名称
}
