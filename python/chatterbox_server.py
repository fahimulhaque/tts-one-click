"""
Chatterbox-Turbo FastAPI server.

Exposes:
  GET  /health        — liveness + capability probe
  POST /tts           — JSON body → audio/wav stream
  POST /clone         — multipart (audio file + text) → audio/wav stream

Runs on 127.0.0.1:8001 by default (pass --port to override).

NOTE: ChatterboxTTS is imported at module level with a try/except so that:
  1. The module can be loaded in test environments where chatterbox-tts is not
     installed (ImportError is swallowed; ChatterboxTTS is set to None).
  2. unittest.mock.patch('chatterbox_server.ChatterboxTTS') can replace the
     name before the startup event fires, because the name exists in the module
     namespace regardless of whether the package is installed.
"""
import argparse
import io
import os
import tempfile
from typing import Optional

import numpy as np
import soundfile as sf
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.responses import StreamingResponse
from pydantic import BaseModel

# ---------------------------------------------------------------------------
# Optional heavy dependencies — guarded so the module is import-safe in test
# environments that lack GPU packages.
# ---------------------------------------------------------------------------
try:
    import torch
    _gpu: bool = torch.cuda.is_available()
except ImportError:
    _gpu = False

# Module-level name so unittest.mock.patch('chatterbox_server.ChatterboxTTS')
# has a real attribute to patch at test time.
try:
    from chatterbox.tts import ChatterboxTTS  # type: ignore[import]
except ImportError:
    ChatterboxTTS = None  # type: ignore[assignment,misc]

# ---------------------------------------------------------------------------
# Application state
# ---------------------------------------------------------------------------
app = FastAPI(title="Chatterbox-Turbo TTS Server")
_model = None


@app.on_event("startup")
async def load_model() -> None:
    """Load the Chatterbox model at startup.

    Skipped when ChatterboxTTS is None (package not installed / test env).
    """
    global _model
    if ChatterboxTTS is not None:
        device = "cuda" if _gpu else "cpu"
        _model = ChatterboxTTS.from_pretrained(device=device)


# ---------------------------------------------------------------------------
# Request/response schemas
# ---------------------------------------------------------------------------
class TTSRequest(BaseModel):
    text: str
    speed: float = 1.0
    tags: Optional[str] = None


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def _audio_to_wav_bytes(wav: np.ndarray, sr: int) -> bytes:
    buf = io.BytesIO()
    sf.write(buf, wav, sr, format="WAV")
    buf.seek(0)
    return buf.read()


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------
@app.get("/health")
def health():
    """Liveness + capability probe."""
    return {"status": "ok", "model": "chatterbox", "gpu": _gpu}


@app.post("/tts")
async def tts(req: TTSRequest):
    """Generate speech from text.

    Text is capped at 500 characters (Chatterbox-Turbo limit).
    Returns audio/wav at 24 kHz (Chatterbox native sample rate).
    """
    if len(req.text) > 500:
        raise HTTPException(
            status_code=400, detail="Text exceeds 500 character limit"
        )
    if _model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    text = req.text
    if req.tags:
        text = f"{req.tags} {text}"
    # generate() returns torch.Tensor [1, N]; sample rate is always model.sr (24000)
    wav_tensor = _model.generate(text)
    wav = wav_tensor.squeeze(0).cpu().numpy()
    wav_bytes = _audio_to_wav_bytes(wav, _model.sr)
    return StreamingResponse(io.BytesIO(wav_bytes), media_type="audio/wav")


@app.post("/clone")
async def clone(
    text: str = Form(...),
    audio: UploadFile = File(...),
    transcript: Optional[str] = Form(None),
):
    """Voice-clone: generate speech in the style of the uploaded reference audio.

    Returns audio/wav at the model's native sample rate.
    """
    if _model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    ref_bytes = await audio.read()
    # audio_prompt_path requires a real file path, not a buffer
    tmp = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
    try:
        tmp.write(ref_bytes)
        tmp.flush()
        tmp.close()
        wav_tensor = _model.generate(text, audio_prompt_path=tmp.name)
    finally:
        os.unlink(tmp.name)
    wav = wav_tensor.squeeze(0).cpu().numpy()
    wav_bytes = _audio_to_wav_bytes(wav, _model.sr)
    return StreamingResponse(io.BytesIO(wav_bytes), media_type="audio/wav")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------
if __name__ == "__main__":
    import uvicorn

    parser = argparse.ArgumentParser(description="Chatterbox-Turbo TTS server")
    parser.add_argument("--port", type=int, default=8001, help="Port to listen on")
    args = parser.parse_args()
    uvicorn.run(app, host="127.0.0.1", port=args.port)
