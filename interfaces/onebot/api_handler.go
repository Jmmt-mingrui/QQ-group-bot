package onebot

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// APIHandler 发送 OneBot v11 API 请求
// 通过 io.Writer 接口可脱离真实 WebSocket 连接进行测试
type APIHandler struct {
	conn io.Writer // WebSocket 连接
}

// NewAPIHandler 创建 APIHandler
func NewAPIHandler(conn io.Writer) *APIHandler {
	return &APIHandler{conn: conn}
}

// ---- API 请求结构 ----

// sendGroupMsgParams send_group_msg API 参数
type sendGroupMsgParams struct {
	GroupID int64  `json:"group_id"`
	Message string `json:"message"`
}

// apiRequest OneBot API 请求通用结构
type apiRequest struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
}

// SendGroupMessage 发送群消息
func (h *APIHandler) SendGroupMessage(groupID string, message string) error {
	gid, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的群号: %w", err)
	}

	req := apiRequest{
		Action: "send_group_msg",
		Params: sendGroupMsgParams{
			GroupID: gid,
			Message: message,
		},
	}
	return h.writeJSON(req)
}

// ReplyToUser 回复群内指定用户（在消息前添加 @ 前缀）
func (h *APIHandler) ReplyToUser(groupID, userID, message string) error {
	gid, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的群号: %w", err)
	}

	req := apiRequest{
		Action: "send_group_msg",
		Params: sendGroupMsgParams{
			GroupID: gid,
			Message: fmt.Sprintf("[CQ:at,qq=%s] %s", userID, message),
		},
	}
	return h.writeJSON(req)
}

// writeJSON 将请求序列化为 JSON 并写入连接
func (h *APIHandler) writeJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = h.conn.Write(data)
	return err
}
