#!/usr/bin/env bash
# scripts/install.sh — TTS-One-Click one-step installer
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Colors
bold()  { echo -e "\033[1m$*\033[0m"; }
green() { echo -e "\033[32m$*\033[0m"; }
blue()  { echo -e "\033[34m$*\033[0m"; }
red()   { echo -e "\033[31m$*\033[0m"; }

bold ""
bold "╔═══════════════════════════════════╗"
bold "║      TTS One-Click Installer      ║"
bold "╚═══════════════════════════════════╝"
echo ""

# Step 1: Check prerequisites
blue "Step 1/6: Checking prerequisites..."
# shellcheck source=scripts/check_prereqs.sh
source scripts/check_prereqs.sh
if [ "${CHECK_FAILED:-0}" -eq 1 ]; then
  red "Prerequisites check failed. Fix the issues above and re-run."
  exit 1
fi
green "Prerequisites OK"
echo ""

# Step 2: Model selection
blue "Step 2/6: Select TTS model"
echo ""
echo "  [1] Chatterbox-Turbo  (350M params, English, paralinguistic tags, fastest)"
echo "  [2] CosyVoice3        (multilingual, zero-shot voice cloning)"
echo ""

# Support non-interactive mode via MODEL env var
if [ -n "${MODEL:-}" ]; then
  CHOICE="$MODEL"
else
  read -r -p "Enter choice (1/2): " CHOICE
fi

case "$CHOICE" in
  1|chatterbox)
    MODEL_NAME="chatterbox"
    SETUP_SCRIPT="scripts/setup_chatterbox.sh"
    ;;
  2|cosyvoice)
    MODEL_NAME="cosyvoice"
    SETUP_SCRIPT="scripts/setup_cosyvoice.sh"
    ;;
  *)
    red "Invalid choice. Run again and enter 1 or 2."
    exit 1
    ;;
esac

green "Selected: $MODEL_NAME"
echo ""

# Step 3: Python environment + model deps
blue "Step 3/6: Setting up Python environment..."
bash "$SETUP_SCRIPT" .venv
green "Python environment ready"
echo ""

# Step 4: Write config
blue "Step 4/6: Writing config..."
cat > config.yaml <<EOF
model: $MODEL_NAME
server_port: 8080
python_port: 8001
venv_path: .venv
dev_mode: false
EOF
green "Config written to config.yaml"
echo ""

# Step 5: Build Go server
blue "Step 5/6: Building Go server..."
go build -o tts-server ./cmd/server
green "Go server built"
echo ""

# Step 6: Build frontend
blue "Step 6/6: Building frontend..."
cd web && npm ci --silent && npm run build --silent && cd ..
green "Frontend built"
echo ""

bold "═══════════════════════════════════════════"
green "  Installation complete!"
bold "═══════════════════════════════════════════"
echo ""
echo "  Start the server:"
echo "    ./tts-server"
echo ""
echo "  Then open: http://localhost:8080"
echo ""
