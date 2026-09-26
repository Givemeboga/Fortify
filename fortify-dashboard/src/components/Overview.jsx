import ScanForm from './ScanForm'
import SiegeLog from './SiegeLog'
import { Icon } from './icons/Icons'

// Severity levels in display order — mirrors the SeverityStrip on the Battle
// Report so the fleet rollup reads as the same "war-room tally". `color` is a
// literal token class (Tailwind can't see runtime-built class names).
const LEVELS = [
  { key: "critical", label: "Critical", color: "text-crit", icon: "i-sev-crit" },
  { key: "high",     label: "High",     color: "text-high", icon: "i-sev-high" },
  { key: "medium",   label: "Medium",   color: "text-med",  icon: "i-sev-med" },
  { key: "low",      label: "Low",      color: "text-low",  icon: "i-sev-low" },
]

// Tally how many *analyzed* scans sit at each severity level.
// A scan counts only once its analysis has completed and produced an overall
// risk — pending / running / failed / never-analyzed scans have no verdict yet,
// so they contribute nothing. `level in counts` also drops any unexpected level.
function rollup(scans) {
  const counts = { critical: 0, high: 0, medium: 0, low: 0 }
  for (const s of scans) {
    if (s.analysis_status !== "completed") continue
    const level = s.analysis?.overall_risk?.level
    if (level in counts) counts[level] += 1
  }
  return counts
}

export default function Overview({ scans, onScanStarted, onSelect, onViewAll }) {
  const counts = rollup(scans)

  return (
    <>
      <h1 className="font-display text-3xl">Command</h1>
      <span className="font-mono text-xs text-muted tracking-widest uppercase">// perimeter control</span>

      {/* Primary action — kept immediate on the landing page */}
      <ScanForm onScanStarted={onScanStarted} />

      {/* Fleet severity rollup — same card treatment as the Battle Report's SeverityStrip */}
      <div className="mt-8 grid grid-cols-4 gap-3">
        {LEVELS.map((l) => {
          const n = counts[l.key] ?? 0
          const active = n > 0
          return (
            <div
              key={l.key}
              className={`panel-stone flex items-center gap-3 px-4 py-3 ${active ? "" : "opacity-40"}`}
            >
              <Icon id={l.icon} size={26} className={active ? l.color : "text-faint"} />
              <div>
                <div className={`font-display text-4xl leading-none ${active ? l.color : "text-muted"}`}>{n}</div>
                <div className="font-mono text-[11px] font-medium uppercase tracking-widest text-muted mt-1">{l.label}</div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Recent activity — compact preview, full list one click away */}
      <div className="mt-10 flex items-baseline justify-between">
        <h2 className="font-display text-xl">Recent activity</h2>
        <button onClick={onViewAll} className="font-mono text-xs text-accent hover:brightness-110">
          View all →
        </button>
      </div>
      <SiegeLog scans={scans} onSelect={onSelect} heading={null} limit={5} />
    </>
  )
}
