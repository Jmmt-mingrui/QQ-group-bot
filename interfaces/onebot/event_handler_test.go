package onebot

import (
	"testing"
)

// 解析群消息事件：文本消息
func TestParseGroupMessage_Text(t *testing.T) {
	raw := `{"post_type":"message","message_type":"group","group_id":123456,"user_id":789,"message":[{"type":"text","data":{"text":"你好"}}],"sender":{"nickname":"小明"}}`

	event, err := parseGroupMessage([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if event == nil {
		t.Fatal("期望解析出事件，但返回 nil")
	}
	if event.GroupID != "123456" {
		t.Errorf("GroupID = %q，期望 %q", event.GroupID, "123456")
	}
	if event.UserID != "789" {
		t.Errorf("UserID = %q，期望 %q", event.UserID, "789")
	}
	if event.Nickname != "小明" {
		t.Errorf("Nickname = %q，期望 %q", event.Nickname, "小明")
	}
	if event.Content != "你好" {
		t.Errorf("Content = %q，期望 %q", event.Content, "你好")
	}
	if event.MsgType != "text" {
		t.Errorf("MsgType = %q，期望 %q", event.MsgType, "text")
	}
}

// 解析图片消息
func TestParseGroupMessage_Image(t *testing.T) {
	raw := `{"post_type":"message","message_type":"group","group_id":123456,"user_id":789,"message":[{"type":"image","data":{"file":"abc.jpg"}}],"sender":{"nickname":"小明"}}`

	event, err := parseGroupMessage([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if event == nil {
		t.Fatal("期望解析出事件，但返回 nil")
	}
	if event.MsgType != "image" {
		t.Errorf("MsgType = %q，期望 %q", event.MsgType, "image")
	}
}

// 解析 @消息
func TestParseGroupMessage_At(t *testing.T) {
	raw := `{"post_type":"message","message_type":"group","group_id":123456,"user_id":789,"message":[{"type":"at","data":{"qq":"123"}},{"type":"text","data":{"text":" 你好"}}],"sender":{"nickname":"小明"}}`

	event, err := parseGroupMessage([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if event == nil {
		t.Fatal("期望解析出事件，但返回 nil")
	}
	if event.MsgType != "at" {
		t.Errorf("MsgType = %q，期望 %q", event.MsgType, "at")
	}
	if event.Content != "你好" {
		t.Errorf("Content = %q，期望 %q", event.Content, "你好")
	}
}

// 非消息事件（notice）返回 nil
func TestParseGroupMessage_NonMessage(t *testing.T) {
	raw := `{"post_type":"notice","notice_type":"group_increase","group_id":123456}`

	event, err := parseGroupMessage([]byte(raw))
	if err != nil {
		t.Fatalf("解析非消息事件不应返回错误: %v", err)
	}
	if event != nil {
		t.Fatal("非消息事件应返回 nil")
	}
}

// 非法 JSON 返回错误
func TestParseGroupMessage_InvalidJSON(t *testing.T) {
	raw := `{invalid json}`

	event, err := parseGroupMessage([]byte(raw))
	if err == nil {
		t.Fatal("非法 JSON 应返回错误")
	}
	if event != nil {
		t.Fatal("非法 JSON 应返回 nil event")
	}
}

// 多个 text 段拼接为完整内容
func TestParseGroupMessage_MultipleTextSegments(t *testing.T) {
	raw := `{"post_type":"message","message_type":"group","group_id":123456,"user_id":789,"message":[{"type":"text","data":{"text":"Hello"}},{"type":"text","data":{"text":" "}},{"type":"text","data":{"text":"World"}}],"sender":{"nickname":"Test"}}`

	event, err := parseGroupMessage([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if event == nil {
		t.Fatal("期望解析出事件，但返回 nil")
	}
	if event.Content != "Hello World" {
		t.Errorf("Content = %q，期望 %q", event.Content, "Hello World")
	}
}
