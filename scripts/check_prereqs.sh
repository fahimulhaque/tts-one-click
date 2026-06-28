#!/usr/bin/env bash
# scripts/check_prereqs.sh
# Source this file; sets CHECK_FAILED=1 if any hard requirement is missing.

export CHECK_FAILED=0

red()   { echo -e "\033[31m$*\033[0m"; }
green() { echo -e "\033[32m$*\033[0m"; }
warn()  { echo -e "\033[33m$*\033[0m"; }

check_python() {
  if command -v python3 &>/dev/null; then
    VER=$(python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
    MAJOR=$(echo "$VER" | cut -d. -f1)
    MINOR=$(echo "$VER" | cut -d. -f2)
    if [ "$MAJOR" -eq 3 ] && [ "$MINOR" -ge 10 ] && [ "$MINOR" -le 11 ]; then
      green "  [OK] Python $VER"
    else
      warn "  [WARN] Python $VER detected; 3.10-3.11 recommended"
    fi
  else
    red "  [FAIL] Python 3 not found"
    CHECK_FAILED=1
  fi
}

check_cuda() {
  if command -v nvidia-smi &>/dev/null; then
    GPU_MEM=$(nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits 2>/dev/null | head -1)
    green "  [OK] NVIDIA GPU detected (${GPU_MEM}MB VRAM)"
    if [ "${GPU_MEM:-0}" -lt 4000 ]; then
      warn "  [WARN] Less than 4GB VRAM; performance may be degraded"
    fi
  else
    warn "  [WARN] No NVIDIA GPU detected; will use CPU (slower)"
  fi
}

check_ram() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    RAM_GB=$(( $(sysctl -n hw.memsize) / 1073741824 ))
  else
    RAM_GB=$(( $(grep MemTotal /proc/meminfo | awk '{print $2}') / 1048576 ))
  fi
  if [ "$RAM_GB" -ge 16 ]; then
    green "  [OK] RAM: ${RAM_GB}GB"
  else
    warn "  [WARN] ${RAM_GB}GB RAM; 16GB+ recommended"
  fi
}

check_disk() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    AVAIL_GB=$(df -g . | awk 'NR==2{print $4}')
  else
    AVAIL_GB=$(df -BG . | awk 'NR==2{gsub("G",""); print $4}')
  fi
  if [ "${AVAIL_GB:-0}" -ge 50 ]; then
    green "  [OK] Disk: ${AVAIL_GB}GB available"
  else
    warn "  [WARN] ${AVAIL_GB}GB disk; 50GB+ recommended for models"
  fi
}

check_go() {
  if command -v go &>/dev/null; then
    green "  [OK] Go $(go version | awk '{print $3}')"
  else
    red "  [FAIL] Go not found. Install from https://go.dev/dl/"
    CHECK_FAILED=1
  fi
}

check_node() {
  if command -v node &>/dev/null; then
    NODE_VER=$(node --version | tr -d 'v' | cut -d. -f1)
    if [ "$NODE_VER" -ge 18 ]; then
      green "  [OK] Node $(node --version)"
    else
      warn "  [WARN] Node $(node --version); v18+ recommended"
    fi
  else
    red "  [FAIL] Node.js not found. Install from https://nodejs.org/"
    CHECK_FAILED=1
  fi
}

check_ffmpeg() {
  if command -v ffmpeg &>/dev/null; then
    green "  [OK] ffmpeg $(ffmpeg -version 2>&1 | head -1 | awk '{print $3}')"
  else
    warn "  [WARN] ffmpeg not found — voice-clone upload will use PyAV fallback (WebM/MP3 still supported)"
    if [[ "$OSTYPE" == "darwin"* ]]; then
      warn "         To install: brew install ffmpeg"
    else
      warn "         To install: sudo apt-get install ffmpeg"
    fi
  fi
}

echo "Checking prerequisites..."
check_python
check_cuda
check_ram
check_disk
check_go
check_node
check_ffmpeg
