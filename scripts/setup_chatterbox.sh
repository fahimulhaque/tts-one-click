#!/usr/bin/env bash
# scripts/setup_chatterbox.sh
set -euo pipefail

VENV="${1:-.venv}"

echo "Setting up Chatterbox-Turbo environment..."
python3 -m venv --clear "$VENV"
# shellcheck source=/dev/null
source "$VENV/bin/activate"

# Upgrade pip/wheel; pin setuptools>=68,<82 (Python 3.12 compat + torch requirement)
pip install --upgrade pip wheel --quiet
pip install "setuptools>=68,<82" --quiet

# Install PyTorch with CUDA if available, else CPU
if command -v nvidia-smi &>/dev/null; then
  CUDA_VER=$(nvidia-smi | grep -oP 'CUDA Version: \K[\d.]+' | cut -d. -f1,2 || echo "12.1")
  if echo "$CUDA_VER" | grep -q "^11"; then
    TORCH_INDEX="https://download.pytorch.org/whl/cu118"
  else
    TORCH_INDEX="https://download.pytorch.org/whl/cu121"
  fi
  pip install torch torchaudio --index-url "$TORCH_INDEX" --quiet
else
  pip install torch torchaudio --index-url "https://download.pytorch.org/whl/cpu" --quiet
fi

# Install numpy first — pkuseg (a chatterbox-tts dep) imports numpy at build time,
# so it must be present before chatterbox-tts resolves its dependency tree.
pip install "numpy>=1.26.0" --quiet

# Install chatterbox-tts without build isolation so pkuseg can find numpy during its build
pip install chatterbox-tts --no-build-isolation --quiet

pip install -r python/requirements_chatterbox.txt --quiet

echo "Chatterbox-Turbo setup complete."
