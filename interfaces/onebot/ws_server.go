// Package onebot — WebSocket 服务器，接收 NapCatQQ 连接
package onebot

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSServer WebSocket 服务器
type WSServer struct {
	addr    string
	path    string
	token   string
	handler *EventHandler
}

// NewWSServer 创建 WebSocket 服务器
func NewWSServer(addr, path, token string, handler *EventHandler) *WSServer {
	return &WSServer{
		addr:    addr,
		path:    path,
		token:   token,
		handler: handler,
	}
}

// Start 启动 WebSocket 服务器，阻塞直到出错
func (s *WSServer) Start() error {
	http.HandleFunc(s.path, s.handleWebSocket)
	if s.path != "/" {
		http.HandleFunc("/", s.handleWebSocket)
	}
	log.Printf("[onebot] WebSocket 服务器启动于 %s%s (同时监听 /)", s.addr, s.path)
	return http.ListenAndServe(s.addr, nil)
}

// handleWebSocket 处理 WebSocket 升级和连接
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 验证 access_token
	if s.token != "" {
		token := r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token != s.token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[onebot] WebSocket 升级失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[onebot] NapCatQQ 已连接: %s", r.RemoteAddr)

	// 创建 API 句柄并注入到 EventHandler
	api := NewAPIHandler(&wsWriter{conn: conn})
	s.handler.SetAPI(api)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[onebot] 连接异常关闭: %v", err)
			} else {
				log.Printf("[onebot] 连接断开")
			}
			break
		}

		if len(raw) == 0 {
			continue
		}

		if err := s.handler.HandleEvent(raw); err != nil {
			log.Printf("[onebot] 事件处理错误: %v", err)
		}
	}
}

// wsWriter 实现 io.Writer，用 WebSocket 发送消息
type wsWriter struct {
	conn *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
