import { useState, useEffect } from 'react'
import Sidebar from './components/Sidebar'
import ScanForm from './components/ScanForm'
import SiegeLog from './components/SiegeLog'

function App() {
  const [scans, setScans] = useState([])   // the list of scans, shared state

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
      <Sidebar />
      <main className="flex-1 p-8">
        <h1 className="font-display text-3xl">Command</h1>
        <span className="font-mono text-xs text-muted tracking-widest uppercase">// perimeter control</span>

        {/* pass loadScans down so ScanForm can refresh the list after a scan */}
        <ScanForm onScanStarted={loadScans} />

        {/* pass the scans down so SiegeLog can display them */}
        <SiegeLog scans={scans} onDelete={handleDelete} />
      </main>
    </div>
  )
}

export default App