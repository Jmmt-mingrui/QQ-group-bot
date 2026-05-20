package llm

import "testing"

func TestMessageCreation(t *testing.T) {
	m := Message{Role: "user", Content: "你好"}
	if m.Role != "user" {
		t.Errorf("Role = %s, want user", m.Role)
	}
	if m.Content != "你好" {
		t.Errorf("Content = %s, want 你好", m.Content)
	}
}

func TestMessageRoles(t *testing.T) {
	roles := []string{"system", "user", "assistant"}
	for _, r := range roles {
		m := Message{Role: r, Content: "test"}
		if m.Role != r {
			t.Errorf("Role = %s, want %s", m.Role, r)
		}
	}
}

// mockProvider 用于测试 Provider 接口
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) ChatCompletion(messages []Message) (string, error) {
	return m.response, m.err
}

func TestProviderInterface(t *testing.T) {
	p := &mockProvider{response: "你好，有什么可以帮你？"}
	resp, err := p.ChatCompletion([]Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "你好，有什么可以帮你？" {
		t.Errorf("response = %s", resp)
	}
}

func TestProviderInterfaceError(t *testing.T) {
	p := &mockProvider{err: ErrProviderUnavailable}
	_, err := p.ChatCompletion(nil)
	if err != ErrProviderUnavailable {
		t.Errorf("expected ErrProviderUnavailable, got %v", err)
	}
}
