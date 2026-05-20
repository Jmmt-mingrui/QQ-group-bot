package message

import "strings"

// CompactMessages 将消息列表压缩为文本摘要。
// 每条消息格式为 "昵称: 内容"，多条消息之间用换行分隔。
// 如果消息列表为空或 nil，返回空字符串。
func CompactMessages(msgs []*Message) string {
	if len(msgs) == 0 {
		return ""
	}

	var b strings.Builder
	for i, msg := range msgs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(msg.Nickname)
		b.WriteString(": ")
		b.WriteString(msg.Content)
	}
	return b.String()
}
