import { useState } from 'react'

// Selected-card accent per provider: local/ollama = azure, cloud/gemini = orange
// (the "data leaves the machine" option reads as the louder choice).
const TONE = {
  ollama: { solid: "#2FA4FF", ring: "rgba(47,164,255,.22)", glow: "rgba(47,164,255,.6)" },
  gemini: { solid: "#F97316", ring: "rgba(249,115,22,.22)", glow: "rgba(249,115,22,.6)" },
}

// Counsel — AI provider settings (handoff 03). Two provider cards act as a
// radio group; Gemini reveals a key field. Save persists to localStorage.
function Settings({ onProviderSaved }) {
  const [provider, setProvider] = useState(localStorage.getItem("fortify_provider") || "ollama")
  const [apiKey, setApiKey] = useState(localStorage.getItem("fortify_gemini_key") || "")
  const [saved, setSaved] = useState(false)

  function save() {
    localStorage.setItem("fortify_provider", provider)
    localStorage.setItem("fortify_gemini_key", apiKey)
    onProviderSaved(provider)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const ProviderCard = ({ value, title, desc }) => {
    const on = provider === value
    const t = TONE[value] || TONE.ollama
    return (
      <div
        onClick={() => setProvider(value)}
        className="panel-stone relative flex flex-col cursor-pointer"
        style={{
          width: 380, padding: "20px 22px", gap: 10,
          boxShadow: on
            ? `0 0 0 1px ${t.solid}, 0 0 20px ${t.ring}, inset 1px 1px 0 rgba(234,241,248,.09), inset -1px -1px 0 rgba(0,0,0,.75)`
            : undefined,
        }}
      >
        {on && <div style={{ position: "absolute", inset: 0, background: "var(--rivets)", pointerEvents: "none" }} />}
        <div className="flex items-center font-mono text-text" style={{ gap: 10, fontSize: 14, letterSpacing: ".12em" }}>
          {on ? (
            <span style={{ width: 14, height: 14, borderRadius: "50%", background: t.solid, boxShadow: `0 0 8px ${t.glow}` }} />
          ) : (
            <span style={{ width: 10, height: 10, borderRadius: "50%", border: "2px solid #8A97A8" }} />
          )}
          {title}
        </div>
        <div style={{ fontSize: 13, lineHeight: 1.55, color: "#8A97A8", textWrap: "pretty" }}>{desc}</div>
      </div>
    )
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        <h1 className="font-display font-semibold leading-none text-text" style={{ fontSize: 44 }}>Counsel</h1>
        <span className="font-mono text-muted" style={{ fontSize: 12, letterSpacing: ".16em" }}>// WHO READS YOUR FINDINGS</span>
      </div>

      <div className="flex flex-col" style={{ gap: 14 }}>
        <div className="font-mono text-muted" style={{ fontSize: 10, letterSpacing: ".16em" }}>AI ANALYSIS PROVIDER</div>
        <div className="grid" style={{ gridTemplateColumns: "repeat(2,380px)", gap: 12 }}>
          <ProviderCard value="ollama" title="LOCAL · OLLAMA" desc="Runs a model on your machine. Scan data never leaves the host. Slower, fully private." />
          <ProviderCard value="gemini" title="CLOUD · GEMINI" desc="Faster and stronger, but scan results are sent to Google's servers. Bring your own API key." />
        </div>

        {provider === "gemini" && (
          <div style={{ maxWidth: 772 }}>
            <label className="font-mono text-faint block" style={{ fontSize: 10, letterSpacing: ".16em", marginBottom: 6 }}>GEMINI API KEY</label>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="AIza…"
              className="w-full font-mono text-text"
              style={{
                height: 46, padding: "0 14px", fontSize: 14, borderRadius: 2,
                backgroundColor: "#0A0E16", border: "1px solid rgba(130,160,210,.22)",
                boxShadow: "inset 0 2px 4px rgba(0,0,0,.6)", outline: "none",
              }}
            />
          </div>
        )}

        <div className="flex items-center" style={{ gap: 12, marginTop: 6 }}>
          <button
            onClick={save}
            className="btn-primary flex items-center"
            style={{ height: 46, padding: "0 28px", fontSize: 15 }}
          >
            Save
          </button>
          {saved && <span className="font-mono text-low" style={{ fontSize: 12 }}>✓ saved</span>}
        </div>
      </div>
    </>
  )
}

export default Settings
