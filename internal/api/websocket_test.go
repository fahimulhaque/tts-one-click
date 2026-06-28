package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketUpgrade(t *testing.T) {
	python := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/wav")
		w.Write([]byte("RIFF....fake audio")) //nolint:errcheck
	}))
	defer python.Close()

	r := setupRouter(python.URL)
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/tts"
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v (resp: %v)", err, resp)
	}
	defer conn.Close()

	err = conn.WriteJSON(map[string]interface{}{
		"text": "Hello", "model": "chatterbox", "speed": 1.0,
	})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if msgType != websocket.BinaryMessage || len(data) == 0 {
		t.Errorf("expected binary audio data, got type=%d len=%d", msgType, len(data))
	}
}

func TestWebSocketEmptyText(t *testing.T) {
	r := setupRouter("http://127.0.0.1:19999")
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/tts"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close()

	err = conn.WriteJSON(map[string]interface{}{"text": ""})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "text required") {
		t.Errorf("expected 'text required' error, got: %s", string(data))
	}
}

func TestWebSocketPythonUnreachable(t *testing.T) {
	// Point to a URL that is definitely not listening
	r := setupRouter("http://127.0.0.1:19999")
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/tts"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close()

	err = conn.WriteJSON(map[string]interface{}{"text": "Hello"})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "python unreachable") {
		t.Errorf("expected 'python unreachable' error, got: %s", string(data))
	}
}
