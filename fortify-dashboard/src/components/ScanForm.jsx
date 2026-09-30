import { useState } from 'react'
import { Icon } from './icons/Icons'

function ScanForm({ onScanStarted }) {
  const [url, setUrl] = useState("")
  const [scanType, setScanType] = useState("passive")
  const [consent, setConsent] = useState(false)

  async function handleScan() {
    if (!url) return
    if (scanType === "active" && !consent) return
    const res = await fetch("http://localhost:8500/scan", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url, scan_type: scanType }),
    })
    if (!res.ok) return
    await res.json()
    onScanStarted()
  }

  const Mode = ({ value, icon, label, tone }) => {
    const on = scanType === value
    return (
      <button
        onClick={() => setScanType(value)}
        className="flex items-center font-mono"
        style={{
          height: 38, padding: "0 18px", gap: 10, borderRadius: 2, fontSize: 12, letterSpacing: ".14em",
          ...(on
            ? { backgroundColor: tone, color: "#0A0E16", border: `1px solid ${tone}`, boxShadow: "inset 0 1px 0 rgba(234,241,248,.45)" }
            : { backgroundColor: "#131A28", backgroundImage: "var(--tex-stone)", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)", boxShadow: "inset 0 1px 0 rgba(234,241,248,.06), 0 2px 0 rgba(0,0,0,.45)" }),
        }}
      >
        <span style={{ color: on ? "#0A0E16" : "#8A97A8", lineHeight: 0 }}><Icon id={icon} size={15} /></span>{label}
      </button>
    )
  }

  return (
    <div className="flex flex-col" style={{ gap: 14 }}>
      <div className="font-mono" style={{ fontSize: 12, fontWeight: 700, letterSpacing: ".16em", color: "#2FA4FF" }}>// WALK THE PERIMETER</div>

      {/* input + scan button */}
      <div className="flex" style={{ gap: 12 }}>
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://target.example"
          className="flex-1 font-mono text-text"
          style={{
            height: 46, padding: "0 14px", fontSize: 14, borderRadius: 2,
            backgroundColor: "#0A0E16", border: "1px solid rgba(130,160,210,.22)",
            boxShadow: "inset 0 2px 4px rgba(0,0,0,.6)", outline: "none",
          }}
        />
        <button
          onClick={handleScan}
          className="btn-primary flex items-center"
          style={{ height: 46, padding: "0 26px", gap: 10, fontSize: 15 }}
        >
          <span style={{ lineHeight: 0 }}><Icon id="i-pennant" size={16} /></span>Scan
        </button>
      </div>

      {/* passive / active toggle */}
      <div className="flex" style={{ gap: 10 }}>
        <Mode value="passive" icon="i-lantern" label="PASSIVE · THE WATCH" tone="#2FA4FF" />
        <Mode value="active" icon="i-swords" label="ACTIVE · THE SIEGE" tone="#F97316" />
      </div>

      {/* consent gate — only when active is selected */}
      {scanType === "active" && (
        <div style={{ maxWidth: 640, border: "1px solid rgba(249,115,22,.5)", borderRadius: 2, padding: 12, background: "rgba(249,115,22,.05)" }}>
          <div className="flex items-center font-mono" style={{ gap: 8, fontSize: 12, letterSpacing: ".08em", color: "#F97316", marginBottom: 8 }}>
            <span style={{ lineHeight: 0 }}><Icon id="i-portcullis" size={14} /></span>SIEGE MODE · OPT-IN
          </div>
          <label className="flex items-start" style={{ gap: 8, fontSize: 13, color: "#8A97A8", cursor: "pointer" }}>
            <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} style={{ marginTop: 3 }} />
            <span>Active scans send injection and traversal payloads. I own this target or hold written authorization to test it.</span>
          </label>
        </div>
      )}
    </div>
  )
}

export default ScanForm
