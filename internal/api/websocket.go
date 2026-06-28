package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsRequest struct {
	Text  string  `json:"text"`
	Model string  `json:"model"`
	Speed float64 `json:"speed"`
	Voice string  `json:"voice"`
}

// WebSocketTTS upgrades the connection to WebSocket, reads a JSON synthesis
// request, and streams the Python TTS response as binary frames.
func (h *Handlers) WebSocketTTS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var req wsRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`)) //nolint:errcheck
		return
	}
	if req.Text == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"text required"}`)) //nolint:errcheck
		return
	}

	payload := fmt.Sprintf(`{"text":%q,"speed":%f,"voice":%q}`,
		req.Text, speedOrDefault(req.Speed), req.Voice)
	resp, err := http.Post(h.pythonURL+"/tts", "application/json", //nolint:noctx
		strings.NewReader(payload))
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"python unreachable"}`)) //nolint:errcheck
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				break
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			break
		}
	}
}
