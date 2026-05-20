package message

import (
	"testing"
	"time"
)

// 创建一条文本消息，字段正确
func TestNewTextMessage(t *testing.T) {
	now := time.Now()
	msg := &Message{
		ID:        "msg_001",
		GroupID:   "group_123",
		UserID:    "user_456",
		Nickname:  "测试用户",
		Content:   "你好，世界",
		Timestamp: now,
		MsgType:   "text",
	}

	if msg.ID != "msg_001" {
		t.Errorf("ID = %q, want %q", msg.ID, "msg_001")
	}
	if msg.GroupID != "group_123" {
		t.Errorf("GroupID = %q, want %q", msg.GroupID, "group_123")
	}
	if msg.UserID != "user_456" {
		t.Errorf("UserID = %q, want %q", msg.UserID, "user_456")
	}
	if msg.Nickname != "测试用户" {
		t.Errorf("Nickname = %q, want %q", msg.Nickname, "测试用户")
	}
	if msg.Content != "你好，世界" {
		t.Errorf("Content = %q, want %q", msg.Content, "你好，世界")
	}
	if msg.MsgType != "text" {
		t.Errorf("MsgType = %q, want %q", msg.MsgType, "text")
	}
	if !msg.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", msg.Timestamp, now)
	}
}

// 创建一条图片消息，MsgType="image"
func TestNewImageMessage(t *testing.T) {
	msg := &Message{
		ID:      "msg_002",
		GroupID: "group_123",
		UserID:  "user_789",
		Content: "[图片]",
		MsgType: "image",
	}

	if msg.MsgType != "image" {
		t.Errorf("MsgType = %q, want %q", msg.MsgType, "image")
	}
	if msg.Content != "[图片]" {
		t.Errorf("Content = %q, want %q", msg.Content, "[图片]")
	}
}

// 消息默认 Compacted=false
func TestMessageDefaultCompactedFalse(t *testing.T) {
	msg := &Message{
		ID:      "msg_003",
		GroupID: "group_123",
		UserID:  "user_456",
		Content: "测试内容",
		MsgType: "text",
	}

	if msg.Compacted {
		t.Error("新创建的消息 Compacted 应该为 false")
	}
}

// 标记消息为已压缩后 Compacted=true
func TestMessageMarkCompacted(t *testing.T) {
	msg := &Message{
		ID:      "msg_004",
		GroupID: "group_123",
		UserID:  "user_456",
		Content: "测试内容",
		MsgType: "text",
	}

	msg.Compacted = true

	if !msg.Compacted {
		t.Error("标记为已压缩后 Compacted 应该为 true")
	}
}

// 消息内容为空字符串时也能创建（空消息）
func TestMessageEmptyContent(t *testing.T) {
	msg := &Message{
		ID:      "msg_005",
		GroupID: "group_123",
		UserID:  "user_456",
		Content: "",
		MsgType: "text",
	}

	if msg.Content != "" {
		t.Errorf("Content = %q, want 空字符串", msg.Content)
	}
	// 验证其他字段仍然正确
	if msg.ID != "msg_005" {
		t.Errorf("ID = %q, want %q", msg.ID, "msg_005")
	}
	if msg.MsgType != "text" {
		t.Errorf("MsgType = %q, want %q", msg.MsgType, "text")
	}
}
