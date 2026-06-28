package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/fahimulhaque/tts-one-click/internal/api"
)

func setupRouter(pythonURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := api.NewHandlers(pythonURL)
	api.RegisterRoutes(r, h)
	return r
}

func TestHealthProxy(t *testing.T) {
	// Mock Python server
	python := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","model":"chatterbox","gpu":false}`))
	}))
	defer python.Close()

	r := setupRouter(python.URL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("unexpected body: %v", body)
	}
}

func TestTTSValidation(t *testing.T) {
	r := setupRouter("http://127.0.0.1:19999") // nothing listening
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tts",
		strings.NewReader(`{"text":"","model":"chatterbox"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("empty text should return 400, got %d", w.Code)
	}
}

func TestMetrics(t *testing.T) {
	r := setupRouter("http://127.0.0.1:19999")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if _, ok := body["requests"]; !ok {
		t.Errorf("missing 'requests' field in metrics response: %v", body)
	}
}

func TestCORSOptions(t *testing.T) {
	r := setupRouter("http://127.0.0.1:19999")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("OPTIONS preflight should return 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header, got: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
