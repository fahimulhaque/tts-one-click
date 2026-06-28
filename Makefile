.PHONY: help build test clean frontend dev

BINARY=tts-server
GO_SRC=$(shell find cmd internal pkg -name '*.go')

help:
	@echo "TTS-One-Click"
	@echo "  make install    - Full installation (interactive)"
	@echo "  make build      - Build Go server"
	@echo "  make frontend   - Build React frontend"
	@echo "  make test       - Run all tests"
	@echo "  make dev        - Start dev server (hot reload)"
	@echo "  make clean      - Remove build artifacts"

install:
	@bash scripts/install.sh

build: $(GO_SRC)
	go build -o $(BINARY) ./cmd/server

frontend:
	cd web && npm ci && npm run build

test:
	go test ./... -v
	cd web && npm test -- --run
	cd python && python -m pytest tests/ -v

clean:
	rm -f $(BINARY)
	rm -rf web/dist

dev:
	@echo "Start Python server first: cd python && python chatterbox_server.py"
	go run ./cmd/server --dev
