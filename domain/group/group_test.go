package group_test

import (
	"testing"

	"qq-group-bot/domain/group"
)

// 测试创建群组，ID 和 Name 正确赋值
func TestGroupCreation(t *testing.T) {
	g := group.Group{ID: "123456", Name: "测试群组"}
	if g.ID != "123456" {
		t.Errorf("期望 ID 为 123456，实际得到 %s", g.ID)
	}
	if g.Name != "测试群组" {
		t.Errorf("期望 Name 为 测试群组，实际得到 %s", g.Name)
	}
}

// 测试 ID 允许为空字符串（用于验证阶段）
func TestGroupEmptyID(t *testing.T) {
	g := group.Group{Name: "未验证群组"}
	if g.ID != "" {
		t.Errorf("期望 ID 为空字符串，实际得到 %s", g.ID)
	}
}

// 测试 Name 允许为空（未获取到群名时）
func TestGroupEmptyName(t *testing.T) {
	g := group.Group{ID: "789012"}
	if g.Name != "" {
		t.Errorf("期望 Name 为空字符串，实际得到 %s", g.Name)
	}
}

// 测试相同 ID 的 Group 应视为同一群组（以 ID 为唯一标识）
func TestGroupEqualityByID(t *testing.T) {
	g1 := group.Group{ID: "123456", Name: "群组A"}
	g2 := group.Group{ID: "123456", Name: "群组B"}
	// 群组以 ID 作为唯一标识，ID 相同即为同一群组
	if g1.ID != g2.ID {
		t.Errorf("相同 ID 的 Group 应视为同一群组，但 g1.ID=%s != g2.ID=%s", g1.ID, g2.ID)
	}
}

// 测试群号支持长数字字符串
func TestGroupLongID(t *testing.T) {
	longID := "1234567890123456789"
	g := group.Group{ID: longID, Name: "长ID群组"}
	if g.ID != longID {
		t.Errorf("期望 ID 为 %s，实际得到 %s", longID, g.ID)
	}
}
