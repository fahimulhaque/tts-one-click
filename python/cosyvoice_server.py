"""
CosyVoice3 FastAPI server.

Exposes:
  GET  /health        — liveness + capability probe
  GET  /voices        — list available built-in voices
  POST /tts           — JSON body → audio/wav stream
  POST /clone         — multipart (audio file + text) → audio/wav stream

Runs on 127.0.0.1:8001 by default (pass --port to override).

CosyVoice3 (Fun-CosyVoice3-0.5B-2512) is a zero-shot model with no fixed
speaker embeddings. /tts uses inference_zero_shot() with the bundled
reference audio; /clone uses inference_zero_shot() with the uploaded audio.

NOTE: CosyVoice is imported at module level with a try/except so that:
  1. The module can be loaded in test environments where cosyvoice is not
     installed (ImportError is swallowed; CosyVoice is set to None).
  2. unittest.mock.patch('cosyvoice_server.CosyVoice') can replace the
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

# Module-level names for test patching via unittest.mock.patch.
try:
    from cosyvoice.cli.cosyvoice import CosyVoice, AutoModel  # type: ignore[import]
except (ImportError, ModuleNotFoundError):
    CosyVoice = None  # type: ignore[assignment,misc]
    AutoModel = None  # type: ignore[assignment,misc]

# ---------------------------------------------------------------------------
# Application state
# ---------------------------------------------------------------------------
app = FastAPI(title="CosyVoice3 TTS Server")
_model = None
_whisper_model = None  # lazy-loaded on first /clone call without transcript

_MODEL_ID = "FunAudioLLM/Fun-CosyVoice3-0.5B-2512"
# AutoModel only calls snapshot_download when the path doesn't already exist.
# Prefer the local cache directory so we survive offline restarts / API glitches.
_MODEL_LOCAL = os.path.expanduser(
    "~/.cache/modelscope/hub/FunAudioLLM/Fun-CosyVoice3-0.5B-2512"
)

# CosyVoice3 is a zero-shot model — it always needs a reference audio.
# We ship a sample wav (cloned alongside CosyVoice source) that acts as the
# default voice when the user doesn't upload their own reference.
_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_DEFAULT_REF_WAV = os.path.join(_SCRIPT_DIR, "..", "CosyVoice", "asset", "zero_shot_prompt.wav")
# Transcript of the bundled reference audio (as required by inference_zero_shot).
_DEFAULT_REF_TEXT = "You are a helpful assistant.<|endofprompt|>希望你以后能够做的比我还好呦。"

# CosyVoice3 LLM assertion: <|endofprompt|> must appear in prompt_text.
_PROMPT_PREFIX = "You are a helpful assistant.<|endofprompt|>"


@app.on_event("startup")
async def load_model() -> None:
    """Load the CosyVoice model at startup.

    Uses AutoModel so it detects cosyvoice.yaml / cosyvoice2.yaml /
    cosyvoice3.yaml automatically — avoids hardcoding the wrong class.
    Skipped when AutoModel is None (package not installed / test env).
    """
    global _model
    if AutoModel is not None:
        model_dir = _MODEL_LOCAL if os.path.isdir(_MODEL_LOCAL) else _MODEL_ID
        _model = AutoModel(model_dir=model_dir)


# ---------------------------------------------------------------------------
# Request/response schemas
# ---------------------------------------------------------------------------
class TTSRequest(BaseModel):
    text: str
    speed: float = 1.0
    voice: str = "default"  # reserved for future multi-voice support; currently ignored


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def _audio_to_wav_bytes(wav: np.ndarray, sr: int) -> bytes:
    buf = io.BytesIO()
    sf.write(buf, wav, sr, format="WAV")
    buf.seek(0)
    return buf.read()


def _ensure_wav_file(audio_bytes: bytes) -> str:
    """Write audio bytes to a 16 kHz mono WAV temp file, converting if needed.

    Decoding priority:
      1. soundfile  — WAV / FLAC / OGG Vorbis / AIFF (zero extra deps)
      2. PyAV (av)  — WebM / MP4 / MP3 / Opus / AAC (bundled FFmpeg codecs,
                      no system ffmpeg binary required; pip install av)
      3. ffmpeg CLI — last-resort subprocess fallback

    Returns the path of the temp file; caller must os.unlink() it.
    """
    # 1. soundfile — covers all libsndfile-supported formats.
    try:
        data, sr = sf.read(io.BytesIO(audio_bytes), dtype="float32", always_2d=True)
        import torch as _torch
        import torchaudio as _ta
        speech = _torch.from_numpy(data.T).mean(dim=0, keepdim=True)  # [1, T]
        if sr != 16000:
            speech = _ta.transforms.Resample(sr, 16000)(speech)
        tmp = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
        sf.write(tmp.name, speech.squeeze(0).numpy(), 16000, format="WAV")
        return tmp.name
    except Exception:
        pass

    # 2. PyAV — handles WebM/Opus/MP4/AAC from browser MediaRecorder without
    #    needing a system ffmpeg binary (codecs are bundled inside the av wheel).
    try:
        import av as _av
        container = _av.open(io.BytesIO(audio_bytes))
        resampler = _av.AudioResampler(format="fltp", layout="mono", rate=16000)
        frames = []
        for frame in container.decode(audio=0):
            for r in resampler.resample(frame):
                frames.append(r.to_ndarray())
        if not frames:
            raise ValueError("no audio frames decoded")
        import numpy as _np
        pcm = _np.concatenate(frames, axis=1).squeeze(0).astype("float32")
        tmp = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
        sf.write(tmp.name, pcm, 16000, format="WAV")
        return tmp.name
    except ImportError:
        pass
    except Exception:
        pass

    # 3. ffmpeg CLI — last resort.
    import subprocess
    src = tempfile.NamedTemporaryFile(suffix=".audio", delete=False)
    dst = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
    dst.close()
    try:
        src.write(audio_bytes)
        src.flush()
        src.close()
        result = subprocess.run(
            ["ffmpeg", "-y", "-i", src.name, "-ar", "16000", "-ac", "1", "-f", "wav", dst.name],
            capture_output=True,
        )
        if result.returncode != 0:
            raise HTTPException(status_code=400, detail="Could not decode uploaded audio")
    except FileNotFoundError:
        raise HTTPException(
            status_code=422,
            detail="Unsupported audio format. Upload WAV, FLAC, OGG, WebM or MP3. "
                   "If using MP3/WebM, run: pip install av",
        )
    finally:
        os.unlink(src.name)
    return dst.name


def _transcribe_wav(wav_path: str) -> str:
    """Return a transcript of the audio at wav_path using Whisper tiny.

    Loads audio via soundfile (no ffmpeg required) and passes a numpy array
    directly to whisper.transcribe so the ffmpeg file-loading path is bypassed.
    Returns empty string on any failure so callers fall back gracefully.
    """
    global _whisper_model
    try:
        import ssl as _ssl
        import whisper as _whisper

        if _whisper_model is None:
            # Bypass corporate SSL proxies for the one-time ~72 MB model download.
            _orig = _ssl._create_default_https_context  # type: ignore[attr-defined]
            _ssl._create_default_https_context = _ssl._create_unverified_context  # type: ignore[attr-defined]
            try:
                _whisper_model = _whisper.load_model("tiny")
            finally:
                _ssl._create_default_https_context = _orig

        # Load audio with soundfile (no ffmpeg) and resample to 16 kHz float32.
        data, sr = sf.read(wav_path, dtype="float32", always_2d=False)
        if data.ndim > 1:
            data = data.mean(axis=1)  # stereo → mono
        if sr != 16000:
            import torch as _torch
            import torchaudio as _ta
            t = _torch.from_numpy(data).unsqueeze(0)
            data = _ta.transforms.Resample(sr, 16000)(t).squeeze(0).numpy()

        result = _whisper_model.transcribe(data, fp16=False)
        return result.get("text", "").strip()
    except Exception:
        return ""


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------
@app.get("/health")
def health():
    """Liveness + capability probe."""
    return {"status": "ok", "model": "cosyvoice", "gpu": _gpu}


@app.get("/voices")
def list_voices():
    """Return list of available built-in voices.

    CosyVoice3 is a zero-shot model — it has no fixed speaker embeddings.
    We return the bundled reference audio as the only built-in voice.
    """
    return {"voices": ["default"]}


@app.post("/tts")
async def tts(req: TTSRequest):
    """Generate speech from text using the built-in default voice.

    CosyVoice3 is a zero-shot model: /tts uses inference_zero_shot() with a
    bundled reference audio.  For a custom voice use /clone instead.
    Returns audio/wav at the model's native sample rate.
    """
    if _model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    if not os.path.exists(_DEFAULT_REF_WAV):
        raise HTTPException(status_code=500, detail="Default reference audio not found; re-run install.sh")
    output = next(_model.inference_zero_shot(req.text, _DEFAULT_REF_TEXT, _DEFAULT_REF_WAV, speed=req.speed))
    wav = output["tts_speech"].numpy().flatten()
    wav_bytes = _audio_to_wav_bytes(wav, _model.sample_rate)
    return StreamingResponse(io.BytesIO(wav_bytes), media_type="audio/wav")


@app.post("/clone")
async def clone(
    text: str = Form(...),
    audio: UploadFile = File(...),
    transcript: Optional[str] = Form(""),
):
    """Voice-clone: generate speech in the style of the uploaded reference audio.

    Returns audio/wav at the model's native sample rate.
    """
    if _model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    ref_bytes = await audio.read()
    # Convert uploaded audio (WAV / WebM / MP4 / OGG…) to 16 kHz WAV.
    ref_wav_path = _ensure_wav_file(ref_bytes)
    # CosyVoice3 needs a real transcript of the reference audio for good voice
    # quality.  If the user didn't supply one, auto-transcribe with Whisper tiny.
    ref_transcript = (transcript or "").strip()
    if not ref_transcript:
        ref_transcript = _transcribe_wav(ref_wav_path)
    # The LLM asserts <|endofprompt|> is present in the prompt_text.
    prompt_text = _PROMPT_PREFIX + ref_transcript
    try:
        output = next(_model.inference_zero_shot(text, prompt_text, ref_wav_path))
    finally:
        os.unlink(ref_wav_path)
    wav = output["tts_speech"].numpy().flatten()
    wav_bytes = _audio_to_wav_bytes(wav, _model.sample_rate)
    return StreamingResponse(io.BytesIO(wav_bytes), media_type="audio/wav")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------
if __name__ == "__main__":
    import uvicorn

    parser = argparse.ArgumentParser(description="CosyVoice3 TTS server")
    parser.add_argument("--port", type=int, default=8001, help="Port to listen on")
    args = parser.parse_args()
    uvicorn.run(app, host="127.0.0.1", port=args.port)
