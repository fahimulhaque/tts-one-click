# ⚙️ Installation & API Reference

This document provides system prerequisites, detailed installation instructions for supported platforms, configuration guides, and a comprehensive REST/WebSocket API specification.

---

## 📋 Table of Contents
1. [Prerequisites](#-prerequisites)
2. [Platform Installation](#-platform-installation)
   - [Ubuntu / Debian](#ubuntu--debian)
   - [macOS (Apple Silicon)](#macos-apple-silicon)
   - [Windows (WSL2)](#windows-wsl2)
   - [Docker](#docker)
3. [Advanced Configuration](#-advanced-configuration)
4. [API Reference](#-api-reference)
   - [POST /api/v1/tts](#post-apiv1tts)
   - [POST /api/v1/clone](#post-apiv1clone)
   - [GET /api/v1/health](#get-apiv1health)
   - [GET /api/v1/metrics](#get-apiv1metrics)
   - [WebSocket /ws/tts](#websocket-wstts)
5. [Troubleshooting](#-troubleshooting)

---

## 🛠️ Prerequisites

Before running the automated installer, ensure your environment meets the minimum standards:

| Dependency | Required Version | Purpose |
| :--- | :--- | :--- |
| **Python** | 3.10 or 3.11 | Python workers, model compilation & inference execution |
| **Go** | 1.21+ | Compiles the unified `tts-server` gateway binary |
| **Node.js** | 18+ | Builds and packages the React frontend app |
| **FFmpeg** | system-wide | Audiosample slicing, normalization, and stitching |
| **libsndfile** | system-wide | C-library wrapper for reading/writing audio files |
| **Free Storage** | 50GB+ | Model weights, checkpoint models, environment runtimes |
| **RAM** | 16GB+ (recommended) | General processing and local model compilation buffers |

---

## 💻 Platform Installation

### Ubuntu / Debian
Update system packages, install C/C++ build tools, Python headers, FFmpeg libraries, and compile Go:

```bash
# 1. Update packages and install dependencies
sudo apt-get update && sudo apt-get install -y \
  build-essential \
  ffmpeg \
  libsndfile1 \
  git \
  python3-dev \
  python3-pip

# 2. Install Go 1.21.11 (if not present)
wget https://go.dev/dl/go1.21.11.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.11.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 3. Clone and run installation
git clone https://github.com/fahimulhaque/tts-one-click.git
cd tts-one-click
./install.sh
```

### macOS (Apple Silicon)
Make sure you have [Homebrew](https://brew.sh/) installed, then run:

```bash
# 1. Install prerequisites via Homebrew
brew install ffmpeg go node

# 2. Clone and install
git clone https://github.com/fahimulhaque/tts-one-click.git
cd tts-one-click
./install.sh
```
> [!TIP]
> Hardware acceleration via Apple Metal Performance Shaders (MPS) is enabled automatically. You do not need to configure CUDA.

### Windows (WSL2)
Ensure you are running Windows 11 (or Windows 10 version 21H2+) with WSL2 and the Ubuntu 20.04/22.04 LTS distribution. Open your WSL2 terminal:

```bash
# 1. Update and install packages
sudo apt-get update && sudo apt-get install -y \
  build-essential \
  ffmpeg \
  libsndfile1 \
  git \
  python3-dev

# 2. Run the installer
./install.sh
```
> [!NOTE]
> For GPU acceleration under WSL2, ensure you have the official Windows NVIDIA GPU driver installed. WSL2 automatically maps CUDA runtimes into the container/VM.

### Docker
Use Docker Compose to bypass installing Go, Python, and Node.js on your host:

```bash
# Standard Chatterbox-Turbo model running
MODEL=chatterbox docker compose -f docker/docker-compose.yml up

# High-fidelity multilingual CosyVoice3 model running
MODEL=cosyvoice docker compose -f docker/docker-compose.yml up
```

---

## 🔧 Advanced Configuration

Custom server settings are loaded from `config.yaml` at boot time. This file is generated automatically by `install.sh`, but can be modified manually:

```yaml
# Gateway Server Port
server_port: 8080

# Internal Python Inference Gateway Port
python_port: 8001

# active model to load: "chatterbox" or "cosyvoice"
model_name: "chatterbox"

# Log configuration level: "debug", "info", "warn", "error"
log_level: "info"
```

---

## 🔌 API Reference

The Go gateway proxies incoming client connections, validates requests, handles WebSocket streaming frames, and monitors internal Python metrics.

---

### POST `/api/v1/tts`
Synthesizes a text string into raw audio WAV output.

#### Request Parameters
| Field | Type | Required | Default | Description |
| :--- | :--- | :---: | :--- | :--- |
| `text` | string | Yes | — | The sentence string to synthesize into speech. |
| `model` | string | No | `chatterbox` | Target inference engine (`chatterbox` \| `cosyvoice`). |
| `speed` | float | No | `1.0` | Playback speed factor (range: `0.5` to `2.0`). |
| `voice` | string | No | `中文女` | Voice token profile (specific to model settings). |
| `stream` | boolean | No | `false` | If `true`, returns a chunked Transfer-Encoding audio stream. |

#### Request Payload
```json
{
  "text": "The quick brown fox jumps over the lazy dog.",
  "model": "chatterbox",
  "speed": 1.0,
  "voice": "default",
  "stream": false
}
```

#### Example Curl Call
```bash
curl -X POST http://localhost:8080/api/v1/tts \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello world", "model": "chatterbox", "speed": 1.0}' \
  --output output.wav
```

---

### POST `/api/v1/clone`
Performs zero-shot voice cloning using an uploaded sample audio clip.

#### Request Parameters (Multipart Form-Data)
| Parameter | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `audio` | File Binary | Yes | Reference audio file (WAV or MP3, recommended duration: 3–10s). |
| `text` | String | Yes | The target text to be spoken by the cloned voice. |
| `transcript` | String | No | Literal transcript text of the reference audio to guide pronunciation. |

#### Example Curl Call
```bash
curl -X POST http://localhost:8080/api/v1/clone \
  -F "audio=@/path/to/my_voice.wav" \
  -F "text=Hello, this is my new cloned voice speaking." \
  -F "transcript=This is my voice" \
  --output cloned_output.wav
```

---

### GET `/api/v1/health`
Inspects the health status of the Gateway and downstream Python workers.

#### Response Body
```json
{
  "status": "healthy",
  "gateway": "ok",
  "python_worker": "connected",
  "loaded_model": "chatterbox",
  "timestamp": "2026-06-28T02:51:00Z"
}
```

#### Example Curl Call
```bash
curl http://localhost:8080/api/v1/health
```

---

### GET `/api/v1/metrics`
Exposes core runtime performance and latency counters.

#### Response Body
```json
{
  "request_count": 42,
  "average_rtf": 0.45,
  "average_ttft_ms": 180,
  "active_sessions": 2
}
```
*   **RTF (Real-Time Factor):** Ratio of synthesis compute time to generated audio duration. Values $< 1.0$ indicate faster-than-realtime generation.
*   **TTFT (Time to First Token):** Time latency in milliseconds before the first audio chunk is written to the response buffer.

#### Example Curl Call
```bash
curl http://localhost:8080/api/v1/metrics
```

---

### WebSocket `/ws/tts`
Allows real-time streaming text-to-speech. Lowers TTFT dramatically by generating audio chunks sequentially.

#### 1. Setup Connection
Initiate a WebSocket handshake:
`ws://localhost:8080/ws/tts`

#### 2. Send Request Frame (Text JSON)
Send a JSON payload to start the stream:
```json
{
  "text": "Streaming speech generation is incredibly low latency.",
  "model": "chatterbox",
  "speed": 1.0
}
```

#### 3. Receive Stream Frames (Binary)
The server returns raw, sequential WAV audio binary chunks.

#### 4. Close Connection
Once the text is fully processed and sent, the server sends a close frame (`1000 Normal Closure`).

---

## 🔍 Troubleshooting

#### 🛑 Error: "Python server not ready"
This means the Go gateway is running but cannot connect to the Python model worker on the designated port (default: `8001`).
*   **Check Python process:** Verify if the worker is running. Try checking logs inside `scripts/` or look at running processes (`ps aux | grep python`).
*   **Check virtualenv:** Re-run `./install.sh` to ensure all Python package dependencies were compiled inside the local `.venv`.
*   **Active Port Clash:** Run `lsof -i :8001` to check if a stale service is occupying the port.

#### 🛑 Error: CUDA Out Of Memory (OOM)
The chosen model weighs more than the available VRAM on your graphics card.
*   **Set CPU Execution Mode:** Disable CUDA to fallback to system CPU execution by launching the server with the environment override:
    `CUDA_VISIBLE_DEVICES="" ./tts-server`
*   **Adjust Batching:** Modify the inference batch parameters inside the model config files under `python/`.

#### 🛑 Error: Port 8080 is already in use
Another web service is running on the default Go HTTP port.
*   Change the listening port by modifying `server_port` in your `config.yaml` file, then launch the gateway again.
*   Find what process is using the port: `lsof -i :8080`.
