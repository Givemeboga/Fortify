# Fortify

**Fortify** is an open-source web application security tool that helps developers and security professionals identify vulnerabilities and improve web app defenses. It consists of three main parts:

1. **Scanner**: A Python-based module that tests web applications for common security issues.  
2. **AI Analyzer**: An AI engine that reads scanner output files, calculates risk levels, and provides suggestions for improving the security of the tested web app.  
3. **Dashboard**: A frontend interface to visualize scan results, vulnerabilities, and risk assessments.

---

## Features

- Web app security scanning (headers, injections, misconfigurations)  
- Risk calculation and vulnerability analysis using AI  
- Interactive dashboard to visualize findings  
- Modular and extendable architecture  

---

## Getting Started

### Prerequisites

- WSL2 (for Windows users) or Linux/macOS  
- Python 3.10+  
- Node.js (for the dashboard)  

### Setup

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/Fortify.git
cd Fortify

# Create and activate the virtual environment
python3 -m venv fortify-venv
source fortify-venv/bin/activate

# Install Python dependencies
pip install -r requirements.txt

---

## Running

### Running the Backend

Run the backend with Uvicorn from the project root:

```bash
uvicorn fortify-backend.main:app --reload --port 8500
```

### Running the Dashboard

Start the dashboard development server:

```bash
cd fortify-dashboard
npm install
npm run dev
```

### Contributing

- Fork the repository
- Create a branch: `git checkout -b feature/your-feature`
- Commit your changes: `git commit -m "Add your feature"`
- Push to your branch: `git push origin feature/your-feature`
- Open a pull request
