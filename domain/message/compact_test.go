package message

import "testing"

// 空消息列表返回空字符串
func TestCompactMessagesEmpty(t *testing.T) {
	// nil 切片
	result := CompactMessages(nil)
	if result != "" {
		t.Errorf("CompactMessages(nil) = %q, want 空字符串", result)
	}

	// 空切片
	result = CompactMessages([]*Message{})
	if result != "" {
		t.Errorf("CompactMessages([]) = %q, want 空字符串", result)
	}
}

// 单条消息返回该消息内容
func TestCompactMessagesSingle(t *testing.T) {
	msg := &Message{
		ID:       "msg_001",
		GroupID:  "group_123",
		UserID:   "user_456",
		Nickname: "测试用户",
		Content:  "你好，世界",
		MsgType:  "text",
	}

	result := CompactMessages([]*Message{msg})
	expected := "测试用户: 你好，世界"
	if result != expected {
		t.Errorf("CompactMessages(single) = %q, want %q", result, expected)
	}
}

// 多条消息返回换行拼接的内容
func TestCompactMessagesMultiple(t *testing.T) {
	msgs := []*Message{
		{
			ID:       "msg_001",
			GroupID:  "group_123",
			UserID:   "user_456",
			Nickname: "用户A",
			Content:  "第一条消息",
			MsgType:  "text",
		},
		{
			ID:       "msg_002",
			GroupID:  "group_123",
			UserID:   "user_789",
			Nickname: "用户B",
			Content:  "第二条消息",
			MsgType:  "text",
		},
	}

	result := CompactMessages(msgs)
	expected := "用户A: 第一条消息\n用户B: 第二条消息"
	if result != expected {
		t.Errorf("CompactMessages(multiple) = %q, want %q", result, expected)
	}
}

// 包含 @消息 时保留 @信息
func TestCompactMessagesWithAt(t *testing.T) {
	msg := &Message{
		ID:       "msg_001",
		GroupID:  "group_123",
		UserID:   "user_456",
		Nickname: "测试用户",
		Content:  "@张三 你好呀",
		MsgType:  "at",
	}

	result := CompactMessages([]*Message{msg})
	expected := "测试用户: @张三 你好呀"
	if result != expected {
		t.Errorf("CompactMessages(at) = %q, want %q", result, expected)
	}
}
