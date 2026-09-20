import os
import requests
from pathlib import Path
from dotenv import load_dotenv

# load .env from the backend dir (fortify-backend/.env)
load_dotenv(Path(__file__).parent.parent / ".env")

OLLAMA_URL = os.getenv("OLLAMA_URL", "http://localhost:11434/api/generate")
OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", "llama3.1:8b")
GEMINI_MODEL = "gemini-3.6-flash"
GEMINI_URL = f"https://generativelanguage.googleapis.com/v1beta/models/{GEMINI_MODEL}:generateContent"


def get_llm_response(prompt: str, provider: str | None = None, api_key: str | None = None) -> str:
    # provider comes from the caller (frontend later) or falls back to .env, default ollama
    provider = provider or os.getenv("LLM_PROVIDER", "ollama")
    if provider == "gemini":
        return _gemini_response(prompt, api_key or os.getenv("GEMINI_API_KEY"))
    return _ollama_response(prompt)


def _ollama_response(prompt: str) -> str:
    payload = {"model": OLLAMA_MODEL, "prompt": prompt, "format": "json", "stream": False}
    response = requests.post(OLLAMA_URL, json=payload, timeout=120)
    return response.json()["response"]


def _gemini_response(prompt: str, api_key: str) -> str:
    if not api_key:
        raise ValueError("Gemini selected but no API key provided (set GEMINI_API_KEY)")
    headers = {"x-goog-api-key": api_key}
    payload = {
        "contents": [{"parts": [{"text": prompt}]}],
        "generationConfig": {"responseMimeType": "application/json"},
    }
    response = requests.post(GEMINI_URL, headers=headers, json=payload, timeout=120)
    data = response.json()
    if "error" in data:                       # ← surface the real API error
        raise RuntimeError(f"Gemini API error: {data['error'].get('message')}")
    return data["candidates"][0]["content"]["parts"][0]["text"]