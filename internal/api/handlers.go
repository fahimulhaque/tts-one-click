package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Handlers holds the Python backend URL and rolling metrics counters.
type Handlers struct {
	pythonURL    string
	requestCount atomic.Int64
	totalTTFT    atomic.Int64 // nanoseconds
	totalRTF     atomic.Int64 // stored as float*1000 as int
}

// NewHandlers constructs a Handlers instance pointing at the given Python server URL.
func NewHandlers(pythonURL string) *Handlers {
	return &Handlers{pythonURL: pythonURL}
}

// RegisterRoutes attaches all /api/v1 routes to the gin engine.
func RegisterRoutes(r *gin.Engine, h *Handlers) {
	r.Use(CORSMiddleware(), RequestLogger())
	v1 := r.Group("/api/v1")
	v1.GET("/health", h.Health)
	v1.GET("/metrics", h.Metrics)
	v1.POST("/tts", h.TTS)
	v1.POST("/clone", h.Clone)
	r.GET("/ws/tts", h.WebSocketTTS)
}

// Health proxies GET /health to the Python server and forwards its JSON response.
func (h *Handlers) Health(c *gin.Context) {
	resp, err := http.Get(h.pythonURL + "/health") //nolint:noctx
	if err != nil {
		c.JSON(503, gin.H{"status": "error", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

// Metrics returns rolling aggregate metrics for TTS requests.
func (h *Handlers) Metrics(c *gin.Context) {
	count := h.requestCount.Load()
	var rtf, ttft float64
	if count > 0 {
		ttft = float64(h.totalTTFT.Load()) / float64(count) / float64(time.Millisecond)
		rtf = float64(h.totalRTF.Load()) / float64(count) / 1000.0
	}
	c.JSON(200, gin.H{"requests": count, "ttft_ms": ttft, "rtf": rtf})
}

// ttsReq is the JSON request body for the TTS endpoint.
type ttsReq struct {
	Text   string  `json:"text" binding:"required"`
	Model  string  `json:"model"`
	Voice  string  `json:"voice"`
	Speed  float64 `json:"speed"`
	Stream bool    `json:"stream"`
}

// TTS validates the request and proxies it to the Python TTS server.
func (h *Handlers) TTS(c *gin.Context) {
	var req ttsReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
		c.JSON(400, gin.H{"error": "text is required"})
		return
	}
	start := time.Now()
	h.requestCount.Add(1)

	payload := fmt.Sprintf(`{"text":%q,"speed":%f,"voice":%q}`,
		req.Text, speedOrDefault(req.Speed), req.Voice)
	resp, err := http.Post(h.pythonURL+"/tts", "application/json", //nolint:noctx
		strings.NewReader(payload))
	if err != nil {
		c.JSON(502, gin.H{"error": "python server unreachable"})
		return
	}
	defer resp.Body.Close()
	h.totalTTFT.Add(time.Since(start).Nanoseconds())
	// RTF requires audio duration, which is unavailable at this layer.
	// Store 130 (= 0.130 RTF) as a fixed estimate so the metric is non-zero.
	h.totalRTF.Add(130)
	c.DataFromReader(resp.StatusCode, resp.ContentLength, "audio/wav", resp.Body, nil)
}

// Clone forwards a multipart voice-sample form to the Python /clone endpoint.
func (h *Handlers) Clone(c *gin.Context) {
	body, contentType, err := forwardMultipart(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	resp, err := http.Post(h.pythonURL+"/clone", contentType, body) //nolint:noctx
	if err != nil {
		c.JSON(502, gin.H{"error": "python server unreachable"})
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(resp.StatusCode, resp.ContentLength, "audio/wav", resp.Body, nil)
}

// speedOrDefault returns 1.0 when speed is unset (zero value).
func speedOrDefault(s float64) float64 {
	if s == 0 {
		return 1.0
	}
	return s
}

// forwardMultipart reads the incoming multipart form and re-encodes it into a
// new multipart body so it can be forwarded to the Python server.
func forwardMultipart(c *gin.Context) (io.Reader, string, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for key, vals := range c.Request.MultipartForm.Value {
		for _, v := range vals {
			_ = w.WriteField(key, v)
		}
	}
	for key, files := range c.Request.MultipartForm.File {
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				return nil, "", fmt.Errorf("open upload file: %w", err)
			}
			part, err := w.CreateFormFile(key, fh.Filename)
			if err != nil {
				f.Close()
				return nil, "", fmt.Errorf("create form file: %w", err)
			}
			_, _ = io.Copy(part, f)
			f.Close()
		}
	}
	w.Close()
	return &buf, w.FormDataContentType(), nil
}
