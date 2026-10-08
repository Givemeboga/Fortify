import { useState } from 'react'

// Accent per provider: local/ollama = azure, cloud = orange (the "data
// leaves the machine" options read as the louder choice), custom = violet.
const TONE = {
  ollama:    { solid: "#2FA4FF", ring: "rgba(47,164,255,.22)", glow: "rgba(47,164,255,.6)" },
  gemini:    { solid: "#F97316", ring: "rgba(249,115,22,.22)", glow: "rgba(249,115,22,.6)" },
  openai:    { solid: "#F97316", ring: "rgba(249,115,22,.22)", glow: "rgba(249,115,22,.6)" },
  anthropic: { solid: "#F97316", ring: "rgba(249,115,22,.22)", glow: "rgba(249,115,22,.6)" },
  custom:    { solid: "#A78BFA", ring: "rgba(167,139,250,.22)", glow: "rgba(167,139,250,.6)" },
}

const PROVIDERS = [
  { value: "ollama", title: "LOCAL · OLLAMA", desc: "Runs a model on your machine. Scan data never leaves the host. Slower, fully private.", needsKey: false, showModel: true, modelLabel: "MODEL (OPTIONAL)", modelPlaceholder: "llama3.1:8b" },
  { value: "gemini", title: "CLOUD · GEMINI", desc: "Google's models via API. Faster and stronger, but scan results are sent to Google's servers. Bring your own key.", needsKey: true, keyLabel: "GEMINI API KEY", keyPlaceholder: "AIza…", showModel: true, modelLabel: "MODEL (OPTIONAL)", modelPlaceholder: "gemini-3.6-flash" },
  { value: "openai", title: "CLOUD · OPENAI", desc: "OpenAI chat models via API. Scan results are sent to OpenAI's servers. Bring your own key.", needsKey: true, keyLabel: "OPENAI API KEY", keyPlaceholder: "sk-…", showModel: true, modelLabel: "MODEL (OPTIONAL)", modelPlaceholder: "gpt-4o-mini" },
  { value: "anthropic", title: "CLOUD · ANTHROPIC", desc: "Claude models via API. Scan results are sent to Anthropic's servers. Bring your own key.", needsKey: true, keyLabel: "ANTHROPIC API KEY", keyPlaceholder: "sk-ant-…", showModel: true, modelLabel: "MODEL (OPTIONAL)", modelPlaceholder: "claude-sonnet-4-5" },
  { value: "custom", title: "CUSTOM · OPENAI-COMPATIBLE", desc: "Any OpenAI-compatible endpoint — Groq, Together, OpenRouter, Mistral, LM Studio, … Privacy depends on the endpoint you point at.", needsKey: true, keyLabel: "API KEY (IF REQUIRED)", keyPlaceholder: "key…", showModel: true, modelLabel: "MODEL", modelPlaceholder: "e.g. llama-3.1-70b-versatile", showBaseURL: true },
]

const keyOf = (p) => `fortify_key_${p}`
const modelOf = (p) => `fortify_model_${p}`

// Counsel — AI provider settings. Provider cards act as a radio group;
// cloud/custom providers reveal key (+ model / base URL) fields. Everything
// persists to localStorage; the Battle Report sends it per request (BYOK).
function Settings({ onProviderSaved }) {
  const [provider, setProvider] = useState(localStorage.getItem("fortify_provider") || "ollama")
  const [apiToken, setApiToken] = useState(localStorage.getItem("fortify_api_token") || "")
  const [keys, setKeys] = useState(() => Object.fromEntries(
    PROVIDERS.filter((p) => p.needsKey).map((p) => {
      // migrate the pre-multi-provider gemini key storage
      const legacy = p.value === "gemini" ? localStorage.getItem("fortify_gemini_key") : null
      return [p.value, localStorage.getItem(keyOf(p.value)) || legacy || ""]
    })
  ))
  const [models, setModels] = useState(() => Object.fromEntries(
    PROVIDERS.filter((p) => p.showModel).map((p) => [p.value, localStorage.getItem(modelOf(p.value)) || ""])
  ))
  const [baseURL, setBaseURL] = useState(localStorage.getItem("fortify_base_url") || "")
  const [saved, setSaved] = useState(false)

  const meta = PROVIDERS.find((p) => p.value === provider)

  function save() {
    localStorage.setItem("fortify_provider", provider)
    localStorage.setItem("fortify_api_token", apiToken)
    for (const [p, k] of Object.entries(keys)) localStorage.setItem(keyOf(p), k)
    for (const [p, m] of Object.entries(models)) localStorage.setItem(modelOf(p), m)
    localStorage.setItem("fortify_base_url", baseURL)
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

  const fieldStyle = {
    height: 46, padding: "0 14px", fontSize: 14, borderRadius: 2,
    backgroundColor: "#0A0E16", border: "1px solid rgba(130,160,210,.22)",
    boxShadow: "inset 0 2px 4px rgba(0,0,0,.6)", outline: "none",
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        <h1 className="font-display font-semibold leading-none text-text" style={{ fontSize: 44 }}>Counsel</h1>
        <span className="font-mono text-muted" style={{ fontSize: 12, letterSpacing: ".16em" }}>// WHO READS YOUR FINDINGS</span>
      </div>

      <div className="flex flex-col" style={{ gap: 14 }}>
        <div className="font-mono text-muted" style={{ fontSize: 10, letterSpacing: ".16em" }}>DASHBOARD → API ACCESS</div>
        <div style={{ maxWidth: 772 }}>
          <label className="font-mono text-faint block" style={{ fontSize: 10, letterSpacing: ".16em", marginBottom: 6 }}>API TOKEN (WHEN THE SERVER REQUIRES ONE)</label>
          <input
            type="password"
            value={apiToken}
            onChange={(e) => setApiToken(e.target.value)}
            placeholder="leave blank for local dev without FORTIFY_API_TOKEN"
            className="w-full font-mono text-text"
            style={fieldStyle}
          />
        </div>

        <div className="font-mono text-muted" style={{ fontSize: 10, letterSpacing: ".16em" }}>AI ANALYSIS PROVIDER</div>
        <div className="grid" style={{ gridTemplateColumns: "repeat(2,380px)", gap: 12 }}>
          {PROVIDERS.map((p) => (
            <ProviderCard key={p.value} value={p.value} title={p.title} desc={p.desc} />
          ))}
        </div>

        {meta.showBaseURL && (
          <div style={{ maxWidth: 772 }}>
            <label className="font-mono text-faint block" style={{ fontSize: 10, letterSpacing: ".16em", marginBottom: 6 }}>ENDPOINT BASE URL</label>
            <input
              type="text"
              value={baseURL}
              onChange={(e) => setBaseURL(e.target.value)}
              placeholder="https://api.groq.com/openai/v1"
              className="w-full font-mono text-text"
              style={fieldStyle}
            />
          </div>
        )}

        {meta.needsKey && (
          <div style={{ maxWidth: 772 }}>
            <label className="font-mono text-faint block" style={{ fontSize: 10, letterSpacing: ".16em", marginBottom: 6 }}>{meta.keyLabel}</label>
            <input
              type="password"
              value={keys[provider] || ""}
              onChange={(e) => setKeys({ ...keys, [provider]: e.target.value })}
              placeholder={meta.keyPlaceholder}
              className="w-full font-mono text-text"
              style={fieldStyle}
            />
          </div>
        )}

        {meta.showModel && (
          <div style={{ maxWidth: 772 }}>
            <label className="font-mono text-faint block" style={{ fontSize: 10, letterSpacing: ".16em", marginBottom: 6 }}>{meta.modelLabel}</label>
            <input
              type="text"
              value={models[provider] || ""}
              onChange={(e) => setModels({ ...models, [provider]: e.target.value })}
              placeholder={`${meta.modelPlaceholder} (blank = server default)`}
              className="w-full font-mono text-text"
              style={fieldStyle}
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
