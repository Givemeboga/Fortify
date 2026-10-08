import { useState, useEffect } from 'react'
import { apiFetch } from '../api'

const INTERVALS = [
  [60, "EVERY HOUR"],
  [360, "EVERY 6 HOURS"],
  [1440, "DAILY"],
  [10080, "WEEKLY"],
]

function fmtNext(iso) {
  if (!iso) return "—"
  const d = new Date(iso)
  return d.toLocaleString()
}

function everyLabel(min) {
  if (min < 60) return `every ${min}m`
  if (min < 1440) return `every ${min / 60}h`
  return `every ${min / 1440}d`
}

// Watchtower — recurring scans. A schedule re-probes a target on an
// interval; the Battle Report diffs each run against the previous one, so
// new exposures (an opened port, a leaked secret) surface immediately.
function Schedules() {
  const [schedules, setSchedules] = useState([])
  const [url, setUrl] = useState("")
  const [scanType, setScanType] = useState("passive")
  const [ports, setPorts] = useState("none")
  // NOTE: named intervalMin — `interval`/`setInterval` would shadow the
  // browser timer used for polling below (review bug: polling died and the
  // create form sent interval_minutes as an object).
  const [intervalMin, setIntervalMin] = useState(1440)
  const [autoAnalyze, setAutoAnalyze] = useState(false)
  const [createError, setCreateError] = useState("")

  async function load() {
    const res = await apiFetch("/schedules")
    setSchedules(await res.json())
  }
  useEffect(() => {
    load()
    const t = setInterval(load, 10000)
    return () => clearInterval(t)
  }, [])

  async function create() {
    if (!url) return
    setCreateError("")
    const provider = localStorage.getItem("fortify_provider") || "ollama"
    const res = await apiFetch("/schedules", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        url, scan_type: scanType, ports,
        interval_minutes: intervalMin, auto_analyze: autoAnalyze, provider,
      }),
    })
    if (!res.ok) {
      try {
        const err = await res.json()
        setCreateError(err.detail || `server rejected the watch (${res.status})`)
      } catch {
        setCreateError(`server rejected the watch (${res.status})`)
      }
      return
    }
    setUrl("")
    load()
  }

  async function toggle(s) {
    await apiFetch(`/schedules/${s.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled: !s.enabled }),
    })
    load()
  }

  async function remove(id) {
    await apiFetch(`/schedules/${id}`, { method: "DELETE" })
    load()
  }

  const fieldStyle = {
    height: 46, padding: "0 14px", fontSize: 14, borderRadius: 2,
    backgroundColor: "#0A0E16", border: "1px solid rgba(130,160,210,.22)",
    boxShadow: "inset 0 2px 4px rgba(0,0,0,.6)", outline: "none",
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        <h1 className="font-display font-semibold leading-none text-text" style={{ fontSize: 44 }}>Watchtower</h1>
        <span className="font-mono text-muted" style={{ fontSize: 12, letterSpacing: ".16em" }}>// KEEP THE WATCH, AROUND THE CLOCK</span>
      </div>

      <div className="flex flex-col" style={{ gap: 14, maxWidth: 772 }}>
        <div className="font-mono text-muted" style={{ fontSize: 10, letterSpacing: ".16em" }}>NEW WATCH</div>
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://target.example"
          className="w-full font-mono text-text"
          style={fieldStyle}
        />
        <div className="flex items-center" style={{ gap: 10 }}>
          {[["passive", "PASSIVE"], ["active", "ACTIVE"]].map(([v, label]) => (
            <button key={v} onClick={() => setScanType(v)} className="font-mono"
              style={{
                height: 30, padding: "0 14px", borderRadius: 2, fontSize: 11, letterSpacing: ".1em",
                ...(scanType === v
                  ? { backgroundColor: "#2FA4FF", color: "#0A0E16", border: "1px solid #2FA4FF" }
                  : { backgroundColor: "#131A28", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)" }),
              }}>{label}</button>
          ))}
          <span className="font-mono text-faint" style={{ fontSize: 11, marginLeft: 8 }}>PORTS</span>
          {[["none", "NONE"], ["top100", "100"], ["top1000", "1000"], ["full", "FULL"]].map(([v, label]) => (
            <button key={v} onClick={() => setPorts(v)} className="font-mono"
              style={{
                height: 30, padding: "0 14px", borderRadius: 2, fontSize: 11, letterSpacing: ".1em",
                ...(ports === v
                  ? { backgroundColor: "#A78BFA", color: "#0A0E16", border: "1px solid #A78BFA" }
                  : { backgroundColor: "#131A28", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)" }),
              }}>{label}</button>
          ))}
        </div>
        <div className="flex items-center" style={{ gap: 10 }}>
          {INTERVALS.map(([v, label]) => (
            <button key={v} onClick={() => setIntervalMin(v)} className="font-mono"
              style={{
                height: 30, padding: "0 14px", borderRadius: 2, fontSize: 11, letterSpacing: ".1em",
                ...(intervalMin === v
                  ? { backgroundColor: "#F5C518", color: "#0A0E16", border: "1px solid #F5C518" }
                  : { backgroundColor: "#131A28", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)" }),
              }}>{label}</button>
          ))}
        </div>
        <label className="flex items-center font-mono text-muted" style={{ gap: 8, fontSize: 12, cursor: "pointer" }}>
          <input type="checkbox" checked={autoAnalyze} onChange={(e) => setAutoAnalyze(e.target.checked)} />
          Auto-analyze each run with the Counsel provider
        </label>
        <div className="font-mono text-faint" style={{ fontSize: 11, lineHeight: 1.6 }}>
          Scheduled runs execute on the server — they use server-side keys (.env), never the keys
          stored in this browser. Cloud auto-analysis needs the key on the server.
        </div>
        {createError && <div className="font-mono text-crit" style={{ fontSize: 12 }}>{createError}</div>}
        <div>
          <button onClick={create} className="btn-primary flex items-center" style={{ height: 46, padding: "0 28px", fontSize: 15 }}>
            Start watch
          </button>
        </div>
      </div>

      <div className="flex flex-col" style={{ gap: 10, marginTop: 8 }}>
        <div className="font-mono text-muted" style={{ fontSize: 10, letterSpacing: ".16em" }}>ACTIVE WATCHES ({schedules.length})</div>
        {schedules.length === 0 && <div className="font-mono text-faint" style={{ fontSize: 12 }}>No watches yet — nothing patrols while you sleep.</div>}
        {schedules.map((s) => (
          <div key={s.id} className="panel-stone flex items-center" style={{ padding: "14px 18px", gap: 16, maxWidth: 900 }}>
            <span style={{ width: 10, height: 10, borderRadius: "50%", background: s.enabled ? "#3DD68C" : "#55606F" }} />
            <div className="flex-1 min-w-0">
              <div className="font-mono text-text break-all" style={{ fontSize: 13 }}>{s.target_url}</div>
              <div className="font-mono text-faint" style={{ fontSize: 11 }}>
                {s.scan_type}{s.ports !== "none" ? ` + ports:${s.ports}` : ""} · {everyLabel(s.interval_minutes)}
                {s.auto_analyze ? " · auto-analyze" : ""} · next {fmtNext(s.next_run_at)}
              </div>
            </div>
            <button onClick={() => toggle(s)} className="font-mono text-muted hover:text-text" style={{ fontSize: 11 }}>
              {s.enabled ? "PAUSE" : "RESUME"}
            </button>
            <button onClick={() => remove(s.id)} className="font-mono text-muted hover:text-text" style={{ fontSize: 11 }}>✕</button>
          </div>
        ))}
      </div>
    </>
  )
}

export default Schedules
