package onebot

import (
	"bytes"
	"encoding/json"
	"testing"
)

// SendGroupMessage 构建正确请求
func TestSendGroupMessage_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	handler := NewAPIHandler(&buf)

	err := handler.SendGroupMessage("123", "hello")
	if err != nil {
		t.Fatalf("SendGroupMessage 失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("输出不是合法 JSON: %v", err)
	}

	// 验证 action
	action, ok := result["action"].(string)
	if !ok || action != "send_group_msg" {
		t.Errorf("action = %v，期望 send_group_msg", result["action"])
	}

	// 验证 params
	params, ok := result["params"].(map[string]interface{})
	if !ok {
		t.Fatal("缺少 params 字段")
	}

	// group_id 应为数字
	groupID, ok := params["group_id"].(float64)
	if !ok || groupID != 123 {
		t.Errorf("group_id = %v，期望 123", params["group_id"])
	}

	msg, ok := params["message"].(string)
	if !ok || msg != "hello" {
		t.Errorf("message = %v，期望 hello", params["message"])
	}
}

// SendGroupMessage 支持 @回复
func TestReplyToUser_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	handler := NewAPIHandler(&buf)

	err := handler.ReplyToUser("123", "789", "你好")
	if err != nil {
		t.Fatalf("ReplyToUser 失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("输出不是合法 JSON: %v", err)
	}

	params, ok := result["params"].(map[string]interface{})
	if !ok {
		t.Fatal("缺少 params 字段")
	}

	msg, ok := params["message"].(string)
	if !ok {
		t.Fatal("缺少 message 字段")
	}

	// 消息应包含 CQ at 码
	expectedAt := "[CQ:at,qq=789]"
	if len(msg) < len(expectedAt) || msg[:len(expectedAt)] != expectedAt {
		t.Errorf("消息开头应为 %q，实际为 %q", expectedAt, msg)
	}

	// 应包含回复内容
	expectedContent := " 你好"
	if len(msg) < len(expectedAt)+len(expectedContent) || msg[len(msg)-len(expectedContent):] != expectedContent {
		t.Errorf("消息末尾应为 %q，实际为 %q", expectedContent, msg)
	}
}
