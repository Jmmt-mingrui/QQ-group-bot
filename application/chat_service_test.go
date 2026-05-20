package application

import (
	"fmt"
	"testing"
)

// ---------- 测试用例 ----------

// TestHandleMessage_Success 测试正常消息保存，Save 成功返回 nil
func TestHandleMessage_Success(t *testing.T) {
	mock := &mockMessageRepo{}
	svc := NewChatService(mock)

	err := svc.HandleMessage("group1", "user1", "nick1", "hello", "text")

	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	if len(mock.messages) != 1 {
		t.Fatalf("期望 1 条消息，得到 %d 条", len(mock.messages))
	}
}

// TestHandleMessage_EmptyContent 测试 Content 为空的消息也能正常保存（不报错）
func TestHandleMessage_EmptyContent(t *testing.T) {
	mock := &mockMessageRepo{}
	svc := NewChatService(mock)

	err := svc.HandleMessage("group1", "user1", "nick1", "", "text")

	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	if len(mock.messages) != 1 {
		t.Fatalf("期望 1 条消息，得到 %d 条", len(mock.messages))
	}
}

// TestHandleMessage_SaveError 测试 repo.Save 返回 error 时 HandleMessage 传递该错误
func TestHandleMessage_SaveError(t *testing.T) {
	mock := &mockMessageRepo{saveErr: fmt.Errorf("db error")}
	svc := NewChatService(mock)

	err := svc.HandleMessage("group1", "user1", "nick1", "hello", "text")

	if err == nil {
		t.Fatal("期望错误，但得到 nil")
	}
	if err.Error() != "db error" {
		t.Fatalf("期望 'db error'，得到: %v", err.Error())
	}
}

// TestHandleMessage_AllFields 验证传入的字段值正确传递到 Message 实体中
func TestHandleMessage_AllFields(t *testing.T) {
	mock := &mockMessageRepo{}
	svc := NewChatService(mock)

	err := svc.HandleMessage("g1", "u1", "n1", "hi", "text")

	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	msg := mock.messages[0]
	if msg.GroupID != "g1" {
		t.Errorf("GroupID 期望 g1，得到 %s", msg.GroupID)
	}
	if msg.UserID != "u1" {
		t.Errorf("UserID 期望 u1，得到 %s", msg.UserID)
	}
	if msg.Nickname != "n1" {
		t.Errorf("Nickname 期望 n1，得到 %s", msg.Nickname)
	}
	if msg.Content != "hi" {
		t.Errorf("Content 期望 hi，得到 %s", msg.Content)
	}
	if msg.MsgType != "text" {
		t.Errorf("MsgType 期望 text，得到 %s", msg.MsgType)
	}
	if msg.ID == "" {
		t.Error("ID 不应为空")
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp 不应为零值")
	}
}

// TestHandleMessage_Batch 测试连续处理 3 条消息，mock 中记录数量为 3
func TestHandleMessage_Batch(t *testing.T) {
	mock := &mockMessageRepo{}
	svc := NewChatService(mock)

	for i := 0; i < 3; i++ {
		err := svc.HandleMessage("g1", "u1", "n1", "msg", "text")
		if err != nil {
			t.Fatalf("第 %d 条消息处理失败: %v", i+1, err)
		}
	}

	if len(mock.messages) != 3 {
		t.Fatalf("期望 3 条消息，得到 %d 条", len(mock.messages))
	}
}
