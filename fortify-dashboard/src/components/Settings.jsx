import { useState } from 'react'

function Settings({ onProviderSaved }) {
  const [provider, setProvider] = useState(localStorage.getItem("fortify_provider") || "ollama")
  const [apiKey, setApiKey] = useState(localStorage.getItem("fortify_gemini_key") || "")
  const [saved, setSaved] = useState(false)

  function save() {
    localStorage.setItem("fortify_provider", provider)
    localStorage.setItem("fortify_gemini_key", apiKey)
    onProviderSaved(provider)   // notify App so the sidebar footer updates
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const Option = ({ value, title, desc }) => (
    <div
      onClick={() => setProvider(value)}
      className={`flex-1 p-4 cursor-pointer ${
        provider === value ? "panel-iron" : "panel-stone"
      }`}
    >
      <div className="flex items-center gap-2">
        <div className={`w-3 h-3 rounded-full border-2 ${provider === value ? "border-accent bg-accent" : "border-muted"}`} />
        <span className="font-mono text-sm uppercase tracking-wider text-text">{title}</span>
      </div>
      <p className="text-xs text-muted mt-2 leading-relaxed">{desc}</p>
    </div>
  )

  return (
    <div className="max-w-2xl">
      <h1 className="font-display text-3xl">Settings</h1>
      <span className="font-mono text-xs text-muted tracking-widest uppercase">// counsel</span>

      <div className="mt-8">
        <div className="font-mono text-[10px] text-faint uppercase tracking-widest mb-3">AI Analysis Provider</div>
        <div className="flex gap-3">
          <Option value="ollama" title="Local · Ollama" desc="Runs a model on your machine. Scan data never leaves the host. Slower, fully private." />
          <Option value="gemini" title="Cloud · Gemini" desc="Faster and stronger, but scan results are sent to Google's servers. Bring your own API key." />
        </div>

        {provider === "gemini" && (
          <div className="mt-4 border border-high/50 rounded p-4 bg-high/5">
            <div className="font-mono text-xs text-high uppercase tracking-wider mb-2">Data leaves the machine</div>
            <p className="text-xs text-muted mb-3">
              Cloud analysis sends the scan's findings to Google. Only enable this for targets whose data you're comfortable sharing.
            </p>
            <label className="font-mono text-[10px] text-faint uppercase tracking-widest">Gemini API key</label>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="AIza…"
              className="w-full mt-1 bg-surface border border-border rounded px-3 py-2 font-mono text-sm text-text focus:outline-none focus:border-accent"
            />
          </div>
        )}

        <div className="flex items-center gap-3 mt-5">
          <button onClick={save} className="bg-accent text-bg font-semibold px-5 py-2 rounded hover:brightness-110">
            Save
          </button>
          {saved && <span className="font-mono text-xs text-low">✓ saved</span>}
        </div>
      </div>
    </div>
  )
}

export default Settings