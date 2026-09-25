import { Crest } from './icons/Icons'

// Highest severity first — a war-room tally at the top of the Battle Report.
const LEVELS = [
  { key: "critical", label: "Critical", color: "text-crit" },
  { key: "high",     label: "High",     color: "text-high" },
  { key: "medium",   label: "Medium",   color: "text-med" },
  { key: "low",      label: "Low",      color: "text-low" },
]

function SeverityStrip({ findings = [] }) {
  const counts = findings.reduce((acc, f) => {
    const lvl = f.severity?.level
    if (lvl) acc[lvl] = (acc[lvl] || 0) + 1
    return acc
  }, {})

  return (
    <div className="grid grid-cols-4 gap-3">
      {LEVELS.map((l) => {
        const n = counts[l.key] || 0
        const active = n > 0
        return (
          <div
            key={l.key}
            className={`panel-stone flex items-center gap-3 px-4 py-3 ${active ? "" : "opacity-40"}`}
          >
            <Crest size={22} className={active ? l.color : "text-faint"} />
            <div>
              <div className={`font-display text-3xl leading-none ${active ? l.color : "text-faint"}`}>{n}</div>
              <div className="font-mono text-[10px] uppercase tracking-widest text-faint mt-0.5">{l.label}</div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default SeverityStrip
