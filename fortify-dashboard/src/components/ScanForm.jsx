import { useState } from 'react'

function ScanForm({ onScanStarted }) {          // ← FIX: destructure props with { }
  const [url, setUrl] = useState("")
  const [scanType, setScanType] = useState("passive")   // ← moved INSIDE the function
  const [consent, setConsent] = useState(false)         // ← moved INSIDE the function

  async function handleScan() {
    if (!url) return                                     // no URL → stop
    if (scanType === "active" && !consent) return        // active but not consented → stop
    const res = await fetch("http://localhost:8500/scan", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url: url, scan_type: scanType }),   // ← use state, not "passive"
    })
    if (!res.ok) return
    await res.json()
    onScanStarted()
  }

  return (
    <div className="mt-6">
      <div className="font-mono text-xs text-accent tracking-widest uppercase mb-3">
        // walk the perimeter
      </div>

      {/* input + scan button */}
      <div className="flex gap-3">
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://target.example"
          className="flex-1 bg-surface border border-border rounded px-3 py-2 font-mono text-sm text-text placeholder:text-faint focus:outline-none focus:border-accent"
        />
        <button
          onClick={handleScan}
          className="bg-accent text-bg font-semibold px-6 rounded hover:brightness-110"
        >
          Scan
        </button>
      </div>

      {/* passive / active toggle */}
      <div className="flex gap-2 mt-3">
        <button
          onClick={() => setScanType("passive")}
          className={`px-4 py-1.5 rounded font-mono text-xs uppercase tracking-wider border ${
            scanType === "passive" ? "bg-accent text-bg border-accent" : "border-border text-muted hover:text-text"
          }`}
        >
          Passive · The Watch
        </button>
        <button
          onClick={() => setScanType("active")}
          className={`px-4 py-1.5 rounded font-mono text-xs uppercase tracking-wider border ${
            scanType === "active" ? "bg-high text-bg border-high" : "border-border text-muted hover:text-text"
          }`}
        >
          Active · The Siege
        </button>
      </div>

      {/* consent gate — only when active is selected */}
      {scanType === "active" && (
        <div className="mt-3 border border-high/50 rounded p-3 bg-high/5">
          <div className="font-mono text-xs text-high uppercase tracking-wider mb-2">Siege Mode · Opt-in</div>
          <label className="flex items-start gap-2 text-sm text-muted cursor-pointer">
            <input
              type="checkbox"
              checked={consent}
              onChange={(e) => setConsent(e.target.checked)}
              className="mt-1"
            />
            <span>Active scans send injection and traversal payloads. I own this target or hold written authorization to test it.</span>
          </label>
        </div>
      )}
    </div>
  )
}

export default ScanForm