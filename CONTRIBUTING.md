# 🤝 Contributing to TTS One-Click

We love your contributions! Whether you are fixing bugs, proposing new features, adding documentation, or integrating new Text-to-Speech models, this guide will help you get your development environment set up quickly.

---

## 📋 Table of Contents
1. [Development Environment Setup](#-development-environment-setup)
2. [Project Structure](#-project-structure)
3. [Adding a New TTS Model Engine](#-adding-a-new-tts-model-engine)
4. [Testing & Quality Checks](#-testing--quality-checks)
5. [Pull Request Checklist](#-pull-request-checklist)

---

## 🛠️ Development Environment Setup

Follow these steps to configure your machine for active code development:

### 1. Clone the Repository
```bash
git clone https://github.com/fahimulhaque/tts-one-click.git
cd tts-one-click
```

### 2. Set Up the Go Backend
Download module dependencies:
```bash
go mod download
```

### 3. Set Up the Python Inference Worker
Create a virtual environment and verify package installations:
```bash
# Create local virtualenv
python3 -m venv .venv
source .venv/bin/activate

# Upgrade pip tools
pip install --upgrade pip setuptools wheel
```

### 4. Set Up the React Frontend
Install Node packages and run the frontend build watcher:
```bash
cd web
npm install
npm run dev # Start React dev server on http://localhost:5173
```

---

## 📂 Project Structure

Here is a map of the repository structure to guide your edits:

```
├── cmd/
│   └── server/             # Go Gateway server entry point
├── internal/
│   ├── api/                # HTTP & WebSockets handlers
│   ├── config/             # Config loaders
│   └── tts/                # Go sub-process runner control
├── pkg/
│   └── logger/             # Lightweight logging wrapper
├── python/
│   ├── chatterbox_server.py # Python adapter for Chatterbox model
│   └── cosyvoice_server.py # Python adapter for CosyVoice3 model
├── scripts/                # Helper installation shell scripts
├── web/                    # Single-Page React Application (Vite + TS)
└── config.yaml             # Local runtime settings (git ignored)
```

---

## 🎨 Adding a New TTS Model Engine

Adding another state-of-the-art TTS model (e.g., F5-TTS, VITS) is straightforward. Follow this step-by-step layout:

### Step 1: Create a Python Server Adapter
Implement a new FastAPI application wrapper in `python/<model_name>_server.py`. The server MUST listen on port `8001` and expose the following endpoints:

*   **`GET /health`**
    Checks if model checkpoints are loaded successfully.
    *Response:* `{"status": "ok", "model": "<model_name>", "gpu": true}`

*   **`POST /tts`**
    Processes text payload and returns a streamed binary WAV audio response.
    *Payload Schema:*
    ```json
    {
      "text": "Text to synthesize",
      "speed": 1.0,
      "voice": "default_voice"
    }
    ```
    *Response:* `audio/wav` chunk stream.

*   **`POST /clone`**
    Zero-shot cloning endpoint parsing `multipart/form-data`.
    *Form-Data parameters:*
    *   `audio` (file binary)
    *   `text` (string)
    *   `transcript` (optional string)
    *Response:* `audio/wav` binary stream.

### Step 2: Define Python Dependencies
Create `python/requirements_<model_name>.txt` containing the necessary PyTorch, Hugging Face transformers, and sound libraries required by the model.

### Step 3: Write the Model Installer Script
Create `scripts/setup_<model_name>.sh` to handle downloading the checkpoints, installing custom model wrappers, or preparing specific OS compile libraries.

### Step 4: Hook into the Install CLI
Edit `scripts/install.sh` to include your new model as a menu selection option, running your setup script when selected.

---

## 🧪 Testing & Quality Checks

Before committing your changes or opening a PR, ensure all tests pass:

### Go Unit Tests
Run backend tests to verify request validation, logging, and proxy routing:
```bash
go test -v ./...
```

### React Web Tests
Verify UI components, parsing, and context providers:
```bash
cd web
npm run test -- --run
```

### Shell Script Linting
Run `shellcheck` on all automated installers to guarantee bash script compatibility:
```bash
shellcheck scripts/*.sh
```

---

## 📝 Pull Request Checklist

When submitting your contribution, ensure that:
- [ ] Your code compiles correctly on at least two platform environments.
- [ ] New APIs, flags, or configuration variables are documented in [INSTALL.md](INSTALL.md).
- [ ] All Unit Tests pass locally (`go test` and React testing suite).
- [ ] The code is formatted correctly (`go fmt ./...`, `npm run lint`).
- [ ] The branch is rebased onto the latest `main` commit.
