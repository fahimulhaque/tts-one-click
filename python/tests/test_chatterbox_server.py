import os
import sys
import unittest.mock as mock

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

# Mock the heavy model load before importing server.
# chatterbox_server exposes ChatterboxTTS as a module-level name (via a
# guarded try/except import), so mock.patch can replace it before the startup
# event fires.
with mock.patch("chatterbox_server.ChatterboxTTS") as m:
    m.from_pretrained.return_value = mock.MagicMock()
    from chatterbox_server import app

client_module = __import__("fastapi.testclient", fromlist=["TestClient"])
TestClient = client_module.TestClient

client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["model"] == "chatterbox"
    assert "gpu" in data


def test_tts_missing_text():
    resp = client.post("/tts", json={})
    assert resp.status_code == 422


def test_tts_text_too_long():
    resp = client.post("/tts", json={"text": "x" * 501, "speed": 1.0})
    assert resp.status_code == 400
