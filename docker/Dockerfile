# Stage 1: Python deps + model (pre-download not included; mount or pull at runtime)
FROM python:3.11-slim AS python-base
WORKDIR /app
RUN apt-get update && apt-get install -y git ffmpeg libsndfile1 && rm -rf /var/lib/apt/lists/*
COPY python/requirements_chatterbox.txt python/requirements_cosyvoice.txt ./python/
ARG MODEL=chatterbox
RUN if [ "$MODEL" = "chatterbox" ]; then \
      pip install --no-cache-dir torch torchaudio --index-url https://download.pytorch.org/whl/cpu && \
      pip install --no-cache-dir -r python/requirements_chatterbox.txt; \
    else \
      pip install --no-cache-dir torch torchaudio --index-url https://download.pytorch.org/whl/cpu && \
      pip install --no-cache-dir -r python/requirements_cosyvoice.txt; \
    fi

# Stage 2: Go build
FROM golang:1.25-alpine AS go-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 go build -o tts-server ./cmd/server

# Stage 3: Frontend build
FROM node:20-alpine AS frontend-build
WORKDIR /app
COPY web/package*.json ./
RUN npm ci --silent
COPY web/ ./
RUN npm run build

# Stage 4: Final image
FROM python:3.11-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ffmpeg libsndfile1 && rm -rf /var/lib/apt/lists/*

# Copy Python env from stage 1
COPY --from=python-base /usr/local/lib/python3.11 /usr/local/lib/python3.11
COPY --from=python-base /usr/local/bin /usr/local/bin

# Copy Go binary
COPY --from=go-build /app/tts-server ./

# Copy frontend
COPY --from=frontend-build /app/dist ./web/dist

# Copy app files
COPY python/ ./python/
COPY config.yaml ./

ARG MODEL=chatterbox
ENV MODEL=$MODEL

# Override venv_path for container (deps are in system Python, not a venv)
RUN sed -i 's|venv_path: .venv|venv_path: /usr/local|' config.yaml

EXPOSE 8080
CMD ["./tts-server", "--config", "config.yaml"]
