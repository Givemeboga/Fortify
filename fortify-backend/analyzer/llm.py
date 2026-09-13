import requests

OLLAMA_URL = "http://localhost:11434/api/generate"
MODEL = "llama3.1:8b"

def get_llm_response(prompt: str) -> str:
    payload = {
        "model": MODEL,
        "prompt": prompt,
        "format": "json",
        "stream": False,
    }

    response = requests.post(OLLAMA_URL, json=payload, timeout=120)

    data = response.json()
    return data["response"]