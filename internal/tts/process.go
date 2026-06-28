package tts

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fahimulhaque/tts-one-click/internal/config"
	"go.uber.org/zap"
)

type Manager struct {
	cfg *config.Config
	log *zap.Logger
	cmd *exec.Cmd
}

func NewManager(cfg *config.Config, log *zap.Logger) *Manager {
	return &Manager{cfg: cfg, log: log}
}

func (m *Manager) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", m.cfg.PythonPort)
}

// FreePort kills any process currently listening on the given TCP port.
// It uses lsof on macOS/Linux. Errors are logged but not fatal — the
// subsequent bind attempt will surface any real conflict.
func FreePort(port int, log *zap.Logger) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		log.Info("freeing port: killing existing process", zap.Int("port", port), zap.Int("pid", pid))
		proc.Kill() //nolint:errcheck
	}
	time.Sleep(500 * time.Millisecond) // give the OS a moment to release the port
}

func (m *Manager) Start() error {
	script := "chatterbox_server.py"
	if m.cfg.Model == "cosyvoice" {
		script = "cosyvoice_server.py"
	}
	FreePort(m.cfg.PythonPort, m.log)
	python := filepath.Join(m.cfg.VenvPath, "bin", "python")
	m.cmd = exec.Command(python, filepath.Join("python", script),
		"--port", fmt.Sprintf("%d", m.cfg.PythonPort))
	m.cmd.Dir = "."
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("start python server: %w", err)
	}
	m.log.Info("python server starting", zap.Int("port", m.cfg.PythonPort))
	return m.WaitReady(m.BaseURL()+"/health", 120)
}

func (m *Manager) WaitReady(healthURL string, retries int) error {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(healthURL) //nolint:noctx
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("python server not ready after %d attempts", retries)
}

func (m *Manager) Stop() {
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
	}
}
