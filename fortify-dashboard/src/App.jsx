import { useState, useEffect } from 'react'
import Sidebar from './components/Sidebar'
import Settings from './components/Settings'
import ScanForm from './components/ScanForm'
import SiegeLog from './components/SiegeLog'
import BattleReport from './components/BattleReport'

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
  // fetch all scans from the backend
  async function loadScans() {
    const res = await fetch("http://localhost:8500/scans")
    const data = await res.json()
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
  )
}


export default App