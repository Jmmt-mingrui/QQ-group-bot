// Package llm 提供 LLM 基础设施实现
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"qq-group-bot/domain/llm"
)

// OpenAIProvider 调用 OpenAI 兼容 API 的 LLM Provider
type OpenAIProvider struct {
	baseURL string
	apiKey  string
	model   string
}

// chatRequest OpenAI Chat Completion 请求体
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatMessage OpenAI 消息格式
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse OpenAI Chat Completion 响应体
type chatResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message chatMessage `json:"message"`
}

// NewOpenAIProvider 创建 OpenAI Provider 实例
func NewOpenAIProvider(baseURL, apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
	}
}

// ChatCompletion 调用 OpenAI Chat Completion API
func (p *OpenAIProvider) ChatCompletion(messages []llm.Message) (string, error) {
	// 转换消息格式
	reqMsgs := make([]chatMessage, 0, len(messages))
	for _, m := range messages {
		reqMsgs = append(reqMsgs, chatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// 构建请求体
	reqBody := chatRequest{
		Model:    p.model,
		Messages: reqMsgs,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm: 序列化请求体失败: %w", err)
	}

	// 发送 POST 请求
	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("llm: 创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("llm: HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: 读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: API 返回状态码 %d: %s", resp.StatusCode, string(respBytes))
	}

	// 解析 JSON 响应
	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("llm: 解析响应 JSON 失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llm: 响应中无 choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}
