// domain/llm/provider.go — LLM Provider 接口
package llm

import "errors"

// ErrProviderUnavailable LLM 服务不可用
var ErrProviderUnavailable = errors.New("llm provider unavailable")

// Message 聊天消息
type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// Provider LLM 调用抽象接口
type Provider interface {
	ChatCompletion(messages []Message) (string, error)
}
