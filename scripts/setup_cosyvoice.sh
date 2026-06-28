#!/usr/bin/env bash
# scripts/setup_cosyvoice.sh
set -euo pipefail

VENV="${1:-.venv}"

echo "Setting up CosyVoice3 environment..."
python3 -m venv --clear "$VENV"
# shellcheck source=/dev/null
source "$VENV/bin/activate"

# Upgrade pip/wheel; pin setuptools>=68,<82 (Python 3.12 compat + torch requirement)
pip install --upgrade pip wheel --quiet
pip install "setuptools>=68,<82" --quiet

# Install PyTorch and onnxruntime — GPU variants on Linux+CUDA, CPU elsewhere
if command -v nvidia-smi &>/dev/null; then
  pip install torch torchaudio --index-url "https://download.pytorch.org/whl/cu121" --quiet
  pip install "onnxruntime-gpu>=1.17.0" --quiet
else
  pip install torch torchaudio --index-url "https://download.pytorch.org/whl/cpu" --quiet
  pip install "onnxruntime>=1.17.0" --quiet
fi

pip install -r python/requirements_cosyvoice.txt --quiet

# Clone CosyVoice3 from source.
# The repo has no setup.py/pyproject.toml, so pip install -e . doesn't work.
# Instead: install its deps, then add the checkout to sys.path via a .pth file.
if [ ! -d "CosyVoice" ]; then
  git clone --depth 1 https://github.com/FunAudioLLM/CosyVoice.git
fi

# Pre-install numpy so old-style setup.py packages that import it at build time
# (openai-whisper, pyworld, etc.) can find it under --no-build-isolation.
pip install "numpy>=1.26.0" --quiet

# Install CosyVoice's own requirements with --no-build-isolation so packages
# with old-style setup.py (openai-whisper, conformer…) can find setuptools/
# pkg_resources from our pinned venv instead of pip's isolated build env.
# - Strip --extra-index-url lines (we already set up torch; avoid CUDA index on CPU)
# - Skip grpcio* (gRPC serving, not needed for our FastAPI server)
# - Skip deepspeed / lightning (training-only; deepspeed is Linux-only anyway)
# Platform markers in requirements.txt handle onnxruntime-gpu vs onnxruntime.
grep -vE "^--extra-index-url|^grpcio|^deepspeed|^lightning|^pyworld|^gradio|^fastapi-cli" CosyVoice/requirements.txt \
  | pip install --no-build-isolation -r /dev/stdin --quiet || true

# Patch processor.py: top-level `import pyworld` fails on Python 3.12 because
# pyworld 0.3.4 uses a removed C API. pyworld is only called inside compute_f0
# (training data pipeline) and is never needed during inference. Make it optional.
sed -i.bak 's/^import pyworld as pw$/try:\n    import pyworld as pw\nexcept ImportError:\n    pw = None  # training-only; not needed for inference/' \
  CosyVoice/cosyvoice/dataset/processor.py

# Patch file_utils.py: torchaudio 2.x removed backend= from torchaudio.load() and
# now requires torchcodec (not available on macOS). Use soundfile directly instead.
python3 - <<'PYEOF'
import pathlib
p = pathlib.Path('CosyVoice/cosyvoice/utils/file_utils.py')
src = p.read_text()
old = (
    'def load_wav(wav, target_sr, min_sr=16000):\n'
    '    speech, sample_rate = torchaudio.load(wav, backend=\'soundfile\')\n'
)
new = (
    'def load_wav(wav, target_sr, min_sr=16000):\n'
    '    # torchaudio 2.x removed backend= and requires torchcodec; use soundfile directly.\n'
    '    import soundfile as _sf\n'
    '    data, sample_rate = _sf.read(wav, dtype=\'float32\', always_2d=True)\n'
    '    speech = torch.from_numpy(data.T)  # [channels, time]\n'
)
if old in src:
    p.write_text(src.replace(old, new, 1))
    print('Patched file_utils.py')
else:
    print('file_utils.py already patched or pattern not found — skipping')
PYEOF

# Register the CosyVoice checkout on Python's path
PYVER=$(python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
echo "$(pwd)/CosyVoice" > "$VENV/lib/python${PYVER}/site-packages/cosyvoice.pth"

# Install matcha-tts — bundled as a git submodule in CosyVoice/third_party/Matcha-TTS
# but that dir is empty after a shallow clone. Install from PyPI instead.
# Cython is required at metadata-generation time by matcha-tts's build backend.
pip install Cython --quiet
pip install matcha-tts --no-build-isolation --quiet

echo "CosyVoice3 setup complete."
