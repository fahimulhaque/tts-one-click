package config_test

import (
	"os"
	"testing"
	"github.com/fahimulhaque/tts-one-click/internal/config"
)

func TestLoadConfig(t *testing.T) {
	content := `model: chatterbox
server_port: 8080
python_port: 8001
venv_path: .venv`
	f, _ := os.CreateTemp("", "config*.yaml")
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "chatterbox" {
		t.Errorf("got model %q, want chatterbox", cfg.Model)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("got port %d, want 8080", cfg.ServerPort)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	f, _ := os.CreateTemp("", "config*.yaml")
	f.WriteString("model: cosyvoice")
	f.Close()
	defer os.Remove(f.Name())

	cfg, _ := config.Load(f.Name())
	if cfg.PythonPort != 8001 {
		t.Errorf("default python_port should be 8001, got %d", cfg.PythonPort)
	}
}
