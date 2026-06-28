"""
Tests for cosyvoice_server.py.

CosyVoice package is NOT installed in the test environment, so all heavy
dependencies are mocked at module level using sys.modules patching.

Mock strategy:
  1. Patch sys.modules with mock cosyvoice packages so the guarded import in
     cosyvoice_server succeeds and CosyVoice becomes a MagicMock (not None).
  2. Patch cosyvoice_server.CosyVoice with a callable that returns a
     pre-configured mock model instance (mock_cosyvoice).
  3. Import the app inside those patches so the module-level state is set up
     before the TestClient triggers the startup event.
"""
import os
import sys
import unittest.mock as mock

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

# Pre-configure the mock model instance returned by CosyVoice(...)
mock_cosyvoice = mock.MagicMock()
mock_cosyvoice.list_available_spks.return_value = ["中文女", "英文男"]

# Mock cosyvoice packages to prevent ImportError, then patch the module-level
# CosyVoice name so startup sets _model = mock_cosyvoice.
with mock.patch.dict(
    "sys.modules",
    {
        "cosyvoice": mock.MagicMock(),
        "cosyvoice.cli": mock.MagicMock(),
        "cosyvoice.cli.cosyvoice": mock.MagicMock(),
    },
):
    with mock.patch("cosyvoice_server.CosyVoice", return_value=mock_cosyvoice):
        from cosyvoice_server import app
        # Import TestClient and create it here so the startup event fires while
        # CosyVoice is still patched → _model is set to mock_cosyvoice.
        from fastapi.testclient import TestClient
        client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["model"] == "cosyvoice"
    assert "gpu" in data


def test_tts_missing_text():
    resp = client.post("/tts", json={})
    assert resp.status_code == 422


def test_voices_endpoint():
    resp = client.get("/voices")
    assert resp.status_code == 200
    assert isinstance(resp.json()["voices"], list)
