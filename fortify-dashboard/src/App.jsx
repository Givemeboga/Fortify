import { useState, useEffect, useRef } from 'react'
import Sidebar from './components/Sidebar'
import Settings from './components/Settings'
import ScanForm from './components/ScanForm'
import SiegeLog from './components/SiegeLog'
import BattleReport from './components/BattleReport'
import IconSprite from './components/icons/IconSprite'
import Toast from './components/Toast'

function useHashRoute() {
  const [hash, setHash] = useState(window.location.hash || "#/")
  useEffect(() => {
    const onChange = () => setHash(window.location.hash || "#/")
    window.addEventListener("hashchange", onChange)
    return () => window.removeEventListener("hashchange", onChange)
  }, [])
  return hash
}

function App() {
  const [scans, setScans] = useState([])   // the list of scans, shared state
  const hash = useHashRoute()                  // current hash route
  let view = "command"                          // default view
  let selectedScanId = null                     // default: no scan selected
  if (hash === "#/settings") {
    view = "settings"
  } else if (hash.startsWith("#/scans/")) {
    view = "battle-report"
    selectedScanId = hash.slice("#/scans/".length)  // extract the scan ID from the hash
  }
  function navigate(to) {
    window.location.hash = to
  }
  const [provider, setProvider] = useState(localStorage.getItem("fortify_provider") || "ollama")  // active AI provider (shared)
  const [toasts, setToasts] = useState([])
  const prevStatus = useRef({})     // scan id → last-seen scan status
  const prevAnalysis = useRef({})   // scan id → last-seen analysis_status

  function dismissToast(id) {
    setToasts((t) => t.filter((x) => x.id !== id))
  }

  // fetch all scans from the backend
  async function loadScans() {
    const res = await fetch("http://localhost:8500/scans")
    const data = await res.json()
    const fresh = []
    for (const s of data) {
      // scan finished → pennant ("Dispatch arrived")
      const prev = prevStatus.current[s.id]
      if (prev && prev !== "completed" && s.status === "completed") {
        fresh.push({ id: `${s.id}-scan-${Date.now()}`, kind: "dispatch", detail: s.target_url })
      }
      prevStatus.current[s.id] = s.status

      // analysis finished with a critical verdict → brazier ("A signal fire is lit").
      // Fire on any pre-completed → completed transition (analyzing/pending/failed/null),
      // not just the narrow "analyzing" window — a 3s poll can skip that window, and an
      // HMR/refresh remount reseeds mid-flight. `prevA === undefined` is the first-mount
      // baseline, so already-analyzed scans present on load never re-alert.
      const prevA = prevAnalysis.current[s.id]
      if (prevA !== undefined && prevA !== "completed" && s.analysis_status === "completed"
          && s.analysis?.overall_risk?.level === "critical") {
        fresh.push({ id: `${s.id}-crit-${Date.now()}`, kind: "critical", detail: s.target_url })
      }
      prevAnalysis.current[s.id] = s.analysis_status
    }
    if (fresh.length) {
      setToasts((t) => [...t, ...fresh])
      // Auto-dismiss after 8s so an alert doesn't linger into the next scan;
      // the manual ✕ still works, and dismissing an already-gone id is a no-op.
      for (const f of fresh) {
        setTimeout(() => dismissToast(f.id), 8000)
      }
    }
    setScans(data)
  }

async function handleDelete(id) {
  await fetch(`http://localhost:8500/scans/${id}`, { method: "DELETE" })
  loadScans()   // refresh the list after deleting
}

  // run loadScans once when App first mounts
useEffect(() => {
  loadScans()                                     // fetch immediately on mount
  const interval = setInterval(loadScans, 3000)   // then re-fetch every 3 seconds
  return () => clearInterval(interval)            // cleanup: stop the interval
}, [])

  return (
    <>
    <IconSprite />
    <Toast toasts={toasts} onDismiss={dismissToast} />
    <div className="flex min-h-screen bg-bg text-text">
      <Sidebar view={view} onNavigate={(v) => navigate(v === "settings" ? "#/settings" : "#/")} provider={provider} />
      <main className="flex-1 p-8">
        {view === "settings" ? (
          <Settings onProviderSaved={setProvider} />
        ) : selectedScanId !== null ? (
          <BattleReport scanId={selectedScanId} onBack={() => navigate("#/")} />
        ) : (
          <>
            <h1 className="font-display text-3xl">Command</h1>
            <span className="font-mono text-xs text-muted tracking-widest uppercase">// perimeter control</span>
            <ScanForm onScanStarted={loadScans} />
            <SiegeLog scans={scans} onDelete={handleDelete} onSelect={(id) => navigate(`#/scans/${id}`)} />
          </>
        )}
      </main>
    </div>
    </>
  )
}


export default App