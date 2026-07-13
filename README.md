<p align="center">
  <img src="assets/FortifyLogoCircle.png" alt="Fortify Logo" width="140" />
</p>

<h1 align="center">Fortify</h1>

<p align="center">
  <b>Open-source web application security scanner with AI-assisted analysis</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/python-3.10+-blue?style=flat-square&logo=python" alt="Python 3.10+" />
  <img src="https://img.shields.io/badge/fastapi-backend-009688?style=flat-square&logo=fastapi" alt="FastAPI" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License" />
  <img src="https://img.shields.io/badge/status-active-brightgreen?style=flat-square" alt="Status" />
</p>

<p align="center">
  <img src="assets/landing-hero.png" alt="Fortify — hold the wall" width="100%" />
</p>

---

## Overview

**Fortify** is an open-source web application security tool that helps developers and security professionals identify vulnerabilities and improve web app defenses. It consists of three main components:

| Component | Description | Status |
|---|---|---|
| **Scanner** | Python module that tests web apps for common security issues (headers, TLS, misconfigurations, injections) | 🟢 Passive checks built · 🟡 Active checks planned |
| **AI Analyzer** | AI engine that reads scanner output, calculates risk levels, and gives actionable remediation suggestions | 🟡 Planned |
| **Dashboard** | Frontend interface to visualize scan results, vulnerabilities, and risk assessments | 🟡 Planned |

---

## 🏰 Landing Page

A look at the Fortify landing page — the medieval-fortress metaphor carried through the whole experience.

<p align="center">
  <img src="assets/landing-1.png" alt="Live scanner demo — point it at a gate and watch" width="90%" />
</p>

<p align="center"><i>See it defend — point Fortify at a URL and watch it probe the perimeter.</i></p>

<p align="center">
  <img src="assets/landing-3.png" alt="Three walls, one stronghold — Scanner, AI Analyzer, Dashboard" width="90%" />
</p>

<p align="center"><i>Three walls, one stronghold — the Scanner, AI Analyzer, and Dashboard.</i></p>

<p align="center">
  <img src="assets/landing-4.png" alt="A bestiary of threats — XSS, CSRF, Clickjacking, Weak TLS" width="90%" />
</p>

<p align="center"><i>A bestiary of threats — the classes of vulnerability Fortify is built to repel.</i></p>

<p align="center">
  <img src="assets/landing-5.png" alt="The siege log — point it at a URL, scan, receive the ranked fix list" width="90%" />
</p>

<p align="center"><i>The siege log — from a URL to a ranked, prioritized fix list.</i></p>

<p align="center">
  <img src="assets/landing-6.png" alt="Raise your own fort — free and open source" width="90%" />
</p>

<p align="center"><i>Free and open source — clone the pipeline and hold the walls in minutes.</i></p>

> **Note:** the landing page is a design mock; the scanner panel is illustrative. The real, wired-up dashboard arrives in Phase 4.

---

## Roadmap

Fortify is built in phases. This table reflects the **actual** current state.

| Phase | Scope | Status |
|---|---|---|
| **1 — Scanner core** | Passive checks (TLS, headers, sensitive paths) | ✅ Done |
| | Active checks — SQLi (error-based, query params) | ✅ Done |
| | Active checks — XSS, path traversal | 🚧 In progress |
| **2 — Backend + DB** | SQLite result storage (data layer) | ✅ Done |
| | FastAPI endpoints (trigger & retrieve scans) | ✅ Done |
| **3 — AI Analyzer** | Claude-powered risk scoring & remediation | ⬜ Planned |
| **4 — Dashboard** | React + Tailwind visualization | ⬜ Planned |
| **5 — Polish** | PDF export, Docker, demo | ⬜ Planned |

### What works today

The **passive scanner** is functional. It runs read-only checks against a target and returns a single structured result:

- **TLS** — protocol version, certificate expiry/validity, cipher suite
- **Headers** — missing defensive headers, present headers, leaky (version-disclosing) headers, redirect chain
- **Sensitive paths** — probes common exposed paths (`/.env`, `/.git/`, `/admin`, …) and records status codes

Scan results are persisted to a local **SQLite** database (`db.py`) with a full create → update → retrieve lifecycle, storing the nested result as JSON.

The **FastAPI backend** exposes this over HTTP. Scans run in the background, so a request returns immediately with an ID and the client polls for the result:

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/scan` | Validate a target URL, start a background passive scan, return the scan ID with `status: pending` |
| `GET` | `/scans` | List all scans (newest first) |
| `GET` | `/scans/{id}` | Retrieve one scan by ID (`404` if not found) |

Invalid URLs are rejected with `422` at the API boundary (Pydantic `HttpUrl` validation).

The **active scanner** (injection-based, opt-in) has its first check working:

- **SQL injection** — injects payloads into each URL query parameter and flags a parameter when the response leaks a database error signature. Reports the vulnerable parameter, the triggering payload, and the matched signature, plus scan-health counters (`requests_made`, `errors`) so a failed scan is never mistaken for a clean one.

---

## Project Structure

```
Fortify/
├── assets/                      # Static assets (logo, images)
├── fortify-backend/             # FastAPI backend & scanner logic
│   ├── main.py                  # FastAPI entry point
│   ├── db.py                    # SQLite data layer (scan persistence)
│   └── scanner/
│       ├── passive/             # Read-only checks (safe)
│       │   ├── tls.py           # TLS version, cert, cipher
│       │   ├── headers.py       # Security & leaky headers
│       │   ├── status.py        # Sensitive-path probing
│       │   └── runner.py        # Orchestrates a full passive scan
│       ├── active/              # Injection checks (opt-in)
│       │   ├── injector.py      # Injects a payload into each query param
│       │   └── sqli.py          # Error-based SQL injection detection
│       └── config/
│           ├── headers.json     # Header lists (config)
│           ├── paths.txt        # Sensitive-path wordlist
│           ├── sqli_payloads.txt # SQL injection payloads
│           └── sql_errors.txt   # DB error signatures
├── requirements.txt             # Python dependencies
├── LICENSE
└── README.md
```

---

## Getting Started

### Prerequisites

- Python 3.10+
- Node.js (for the dashboard)
- WSL2 (Windows users) or Linux/macOS

### Installation

```bash
# Clone the repository
git clone https://github.com/Givemeboga/Fortify.git
cd Fortify

# Create and activate a virtual environment
python -m venv venv
source venv/bin/activate          # Windows: venv\Scripts\activate

# Install Python dependencies
pip install -r requirements.txt
```

---

## Usage

### Run a passive scan (available now)

From inside `fortify-backend/`:

```python
from scanner.passive.runner import run_passive_scan
import json

result = run_passive_scan("https://example.com")
print(json.dumps(result, indent=2))
```

Returns a single nested dictionary with `tls`, `headers`, and `status` sections — ready to be stored, served over an API, or analyzed.

### Backend API (available now)

Run the FastAPI backend from inside `fortify-backend/`:

```bash
uvicorn main:app --reload --port 8500
```

The API is available at `http://localhost:8500`, with interactive Swagger docs at `http://localhost:8500/docs`.

Example — start a scan, then retrieve it:

```bash
# start a scan (returns an id immediately)
curl -X POST http://localhost:8500/scan \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# retrieve the result by id
curl http://localhost:8500/scans/1
```

### Dashboard (planned)

The React dashboard arrives in Phase 4.

---

## ⚠️ Legal & Ethical Use

Fortify is intended for **authorized security testing only**. Only scan systems you **own** or have **explicit written permission** to test. Unauthorized scanning of systems you do not control may be illegal under computer-misuse laws (e.g. the CFAA in the US and equivalents elsewhere). You are solely responsible for how you use this tool.

---

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository
2. Create a branch: `git checkout -b feature/your-feature`
3. Commit your changes: `git commit -m "Add your feature"`
4. Push to your branch: `git push origin feature/your-feature`
5. Open a pull request

Please make sure your code is clean and tested before submitting.

---

## License

This project is licensed under the [MIT License](LICENSE) © 2026 Youssef Ben Chaouacha.