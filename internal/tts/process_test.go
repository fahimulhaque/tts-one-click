package tts_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fahimulhaque/tts-one-click/internal/config"
	"github.com/fahimulhaque/tts-one-click/internal/tts"
	"go.uber.org/zap"
)

func TestManagerBaseURL(t *testing.T) {
	cfg := &config.Config{PythonPort: 8001}
	m := tts.NewManager(cfg, zap.NewNop())
	if m.BaseURL() != "http://127.0.0.1:8001" {
		t.Errorf("unexpected base URL: %s", m.BaseURL())
	}
}

func TestManagerWaitReady(t *testing.T) {
	// Start a test HTTP server simulating a healthy Python server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	cfg := &config.Config{PythonPort: 0} // port unused; we test waitReady directly
	m := tts.NewManager(cfg, zap.NewNop())
	err := m.WaitReady(ts.URL+"/health", 3)
	if err != nil {
		t.Fatalf("expected ready, got error: %v", err)
	}
}
