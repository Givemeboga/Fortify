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

// Neutral fleet context — counts every scan, analyzed or not.
function opsSummary(scans) {
  const targets = new Set(scans.map((s) => s.target_url))
  const reports = scans.filter((s) => s.analysis_status === "completed").length
  return { patrols: scans.length, targets: targets.size, reports }
}

// Vulnerability classes, read straight from each scan's raw `results`.
// `test(results)` is true when that scan tripped the check.
const VULN_CLASSES = [
  { key: "sqli",      label: "SQL Injection",        icon: "i-vial",       test: (r) => r.sqli?.vulnerable || r.sqli_boolean?.vulnerable },
  { key: "xss",       label: "Cross-Site Scripting", icon: "i-herald",     test: (r) => r.xss?.vulnerable },
  { key: "traversal", label: "Path Traversal",       icon: "i-postern",    test: (r) => r.path_traversal?.vulnerable },
  { key: "tls",       label: "Weak TLS",             icon: "i-seal-crack", test: (r) => r.tls && (r.tls.cert_valid === false || r.tls.cert_expired === true) },
  { key: "headers",   label: "Missing Headers",      icon: "i-ward",       test: (r) => (r.headers?.missing_headers?.length || 0) > 0 },
]

// How many scans tripped each vulnerability class across the fleet.
function vulnBreakdown(scans) {
  const counts = Object.fromEntries(VULN_CLASSES.map((v) => [v.key, 0]))
  for (const s of scans) {
    const r = s.results
    if (!r) continue
    for (const v of VULN_CLASSES) if (v.test(r)) counts[v.key] += 1
  }
  return counts
}

export default function Overview({ scans, onSelect, onViewAll }) {
  const counts = rollup(scans)
  const ops = opsSummary(scans)
  const vulns = vulnBreakdown(scans)

  return (
    <>
      <h1 className="font-display text-3xl">Command</h1>
      <span className="font-mono text-xs text-muted tracking-widest uppercase">// state of the keep</span>

      {/* Operations summary — neutral fleet context */}
      <div className="mt-8 grid grid-cols-3 gap-3">
        {[
          { label: "Patrols run", value: ops.patrols },
          { label: "Targets watched", value: ops.targets },
          { label: "Reports drafted", value: ops.reports },
        ].map((c) => (
          <div key={c.label} className="panel-iron px-4 py-3">
            <div className="font-display text-3xl leading-none text-text">{c.value}</div>
            <div className="font-mono text-[11px] font-medium uppercase tracking-widest text-muted mt-1">{c.label}</div>
          </div>
        ))}
      </div>

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

      {/* Vulnerability breakdown — what kinds of holes exist across the fleet */}
      <div className="mt-8">
        <h2 className="font-display text-xl mb-3">Threats sighted</h2>
        <div className="panel-stone divide-y divide-border/60">
          {VULN_CLASSES.map((v) => {
            const n = vulns[v.key]
            const active = n > 0
            return (
              <div key={v.key} className={`flex items-center justify-between px-4 py-2.5 ${active ? "" : "opacity-40"}`}>
                <span className="flex items-center gap-2.5">
                  <Icon id={v.icon} size={18} className={active ? "text-high" : "text-faint"} />
                  <span className={`font-mono text-sm ${active ? "text-text" : "text-faint"}`}>{v.label}</span>
                </span>
                <span className={`font-display text-xl ${active ? "text-text" : "text-faint"}`}>{n}</span>
              </div>
            )
          })}
        </div>
      </div>

      {/* Recent activity — compact preview, full list one click away */}
      <div className="mt-10 flex items-baseline justify-between">
        <h2 className="font-display text-xl">Recent activity</h2>
        <button onClick={onViewAll} className="font-mono text-xs text-accent hover:brightness-110">
          View all →
        </button>
      </div>
      <SiegeLog scans={scans} onSelect={onSelect} heading={null} limit={3} />
    </>
  )
}
