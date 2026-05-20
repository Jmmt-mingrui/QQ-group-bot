package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qq-group-bot/domain/llm"
)

// 编译时检查：OpenAIProvider 实现了 domain/llm.Provider
var _ llm.Provider = (*OpenAIProvider)(nil)

func TestNewOpenAIProvider(t *testing.T) {
	p := NewOpenAIProvider("https://api.openai.com", "sk-test", "gpt-4")
	if p == nil {
		t.Fatal("NewOpenAIProvider() 返回 nil")
	}
	if p.baseURL != "https://api.openai.com" {
		t.Errorf("baseURL = %q, want %q", p.baseURL, "https://api.openai.com")
	}
	if p.apiKey != "sk-test" {
		t.Errorf("apiKey = %q, want %q", p.apiKey, "sk-test")
	}
	if p.model != "gpt-4" {
		t.Errorf("model = %q, want %q", p.model, "gpt-4")
	}
}

func TestChatCompletion_Success(t *testing.T) {
	// 启动 mock HTTP 服务器，返回合法 OpenAI 格式响应
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("Authorization = %q, want Bearer sk-test", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"你好！"}}]}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider(srv.URL, "sk-test", "gpt-4")
	resp, err := p.ChatCompletion([]llm.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("ChatCompletion() 返回错误: %v", err)
	}
	if resp != "你好！" {
		t.Errorf("response = %q, want %q", resp, "你好！")
	}
}

func TestChatCompletion_RequestBody(t *testing.T) {
	// 捕获请求体，验证 model 和 messages 字段
	var reqBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider(srv.URL, "sk-test", "gpt-4")
	msgs := []llm.Message{
		{Role: "system", Content: "你是一个助手"},
		{Role: "user", Content: "你好"},
	}
	_, _ = p.ChatCompletion(msgs)

	if reqBody == nil {
		t.Fatal("未能捕获请求体")
	}

	// 验证 model
	model, ok := reqBody["model"].(string)
	if !ok || model != "gpt-4" {
		t.Errorf("model = %v, want gpt-4", reqBody["model"])
	}

	// 验证 messages
	messages, ok := reqBody["messages"].([]any)
	if !ok {
		t.Fatal("messages 字段不存在或类型错误")
	}
	if len(messages) != 2 {
		t.Fatalf("messages 长度 = %d, want 2", len(messages))
	}

	msg0 := messages[0].(map[string]any)
	if msg0["role"] != "system" || msg0["content"] != "你是一个助手" {
		t.Errorf("messages[0] = %v", msg0)
	}
	msg1 := messages[1].(map[string]any)
	if msg1["role"] != "user" || msg1["content"] != "你好" {
		t.Errorf("messages[1] = %v", msg1)
	}
}

func TestChatCompletion_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider(srv.URL, "sk-test", "gpt-4")
	_, err := p.ChatCompletion([]llm.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("期望错误，但得到 nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("错误信息应包含状态码 500，实际: %v", err)
	}
}

func TestChatCompletion_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`这不是合法 JSON`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider(srv.URL, "sk-test", "gpt-4")
	_, err := p.ChatCompletion([]llm.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("期望 JSON 解析错误，但得到 nil")
	}
}
