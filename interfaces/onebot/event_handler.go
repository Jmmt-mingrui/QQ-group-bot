// Package onebot — OneBot v11 事件解析与分发
package onebot

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"qq-group-bot/application"
)

// GroupMessageEvent 群消息事件
type GroupMessageEvent struct {
	GroupID  string
	UserID   string
	Nickname string
	Content  string
	MsgType  string // "text" / "image" / "at"
}

// EventHandler OneBot 事件处理器
type EventHandler struct {
	chatService *application.ChatService
	askService  *application.AskService
	api         *APIHandler
}

// NewEventHandler 创建事件处理器
func NewEventHandler(chat *application.ChatService, ask *application.AskService) *EventHandler {
	return &EventHandler{
		chatService: chat,
		askService:  ask,
	}
}

// SetAPI 设置 API 句柄（连接建立后注入）
func (h *EventHandler) SetAPI(api *APIHandler) {
	h.api = api
}

// HandleEvent 处理原始事件 JSON
func (h *EventHandler) HandleEvent(rawJSON []byte) error {
	log.Printf("[onebot] 收到事件: %s", string(rawJSON))

	event, err := parseGroupMessage(rawJSON)
	if err != nil {
		log.Printf("[onebot] JSON 解析错误: %v", err)
		return err
	}
	if event == nil {
		return nil
	}

	log.Printf("[onebot] 群消息: group=%s user=%s nick=%s type=%s content=%s",
		event.GroupID, event.UserID, event.Nickname, event.MsgType, event.Content)

	// @消息 → 问答
	if event.MsgType == "at" && h.askService != nil {
		answer, err := h.askService.Ask(event.GroupID, event.Content)
		if err != nil {
			log.Printf("[onebot] 问答错误: %v", err)
			// 发送错误回复
			if h.api != nil {
				h.api.ReplyToUser(event.GroupID, event.UserID, "出错了: "+err.Error())
			}
			return err
		}
		log.Printf("[onebot] 问答回复: %s", answer)
		if h.api != nil {
			return h.api.ReplyToUser(event.GroupID, event.UserID, answer)
		}
		return nil
	}

	// 普通消息 → 保存
	if h.chatService != nil {
		return h.chatService.HandleMessage(event.GroupID, event.UserID, event.Nickname, event.Content, event.MsgType)
	}
	return nil
}

// ---- JSON 解析 ----

type onebotSegment struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type textData struct {
	Text string `json:"text"`
}

type atData struct {
	QQ string `json:"qq"`
}

type onebotEvent struct {
	PostType    string          `json:"post_type"`
	MessageType string          `json:"message_type"`
	GroupID     json.RawMessage `json:"group_id"`
	UserID      json.RawMessage `json:"user_id"`
	Sender      *onebotSender   `json:"sender"`
	Message     json.RawMessage `json:"message"`
}

type onebotSender struct {
	Nickname string `json:"nickname"`
}

// parseGroupMessage 解析 OneBot v11 消息事件
func parseGroupMessage(rawJSON []byte) (*GroupMessageEvent, error) {
	var evt onebotEvent
	if err := json.Unmarshal(rawJSON, &evt); err != nil {
		return nil, err
	}

	if evt.PostType != "message" || evt.MessageType != "group" {
		return nil, nil
	}

	segs, err := parseMessageSegments(evt.Message)
	if err != nil {
		return nil, err
	}

	result := &GroupMessageEvent{
		GroupID: parseID(evt.GroupID),
		UserID:  parseID(evt.UserID),
		MsgType: "text",
	}
	if evt.Sender != nil {
		result.Nickname = evt.Sender.Nickname
	}

	var textParts []string
	hasAt := false

	for _, seg := range segs {
		switch seg.Type {
		case "text":
			var td textData
			if err := json.Unmarshal(seg.Data, &td); err == nil && td.Text != "" {
				textParts = append(textParts, td.Text)
			}
		case "image":
			if result.MsgType == "text" {
				result.MsgType = "image"
			}
		case "at":
			hasAt = true
		}
	}

	if hasAt {
		result.MsgType = "at"
	}
	result.Content = strings.TrimSpace(strings.Join(textParts, ""))
	return result, nil
}

// parseID 解析 group_id/user_id（可能是数字或字符串）
func parseID(raw json.RawMessage) string {
	var numID int64
	if err := json.Unmarshal(raw, &numID); err == nil {
		return strconv.FormatInt(numID, 10)
	}
	var strID string
	json.Unmarshal(raw, &strID)
	return strID
}

func parseMessageSegments(raw json.RawMessage) ([]onebotSegment, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var segs []onebotSegment
	if err := json.Unmarshal(raw, &segs); err != nil {
		return nil, err
	}
	return segs, nil
}
