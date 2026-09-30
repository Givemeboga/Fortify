import { useState } from 'react'
import { Icon } from './icons/Icons'

const SEV = {
  critical: { c: "#FF4D4D", icon: "i-sev-crit", label: "CRITICAL" },
  high:     { c: "#F97316", icon: "i-sev-high", label: "HIGH" },
  medium:   { c: "#F5C518", icon: "i-sev-med",  label: "MEDIUM" },
  low:      { c: "#3DD68C", icon: "i-sev-low",  label: "LOW" },
}
const COLS = "minmax(0,2.4fr) 110px 140px 140px 180px 40px"

function formatTime(iso) {
  const d = new Date(iso)
  const p = (n) => String(n).padStart(2, "0")
  return `${p(d.getDate())}/${p(d.getMonth() + 1)}/${d.getFullYear()} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

// `heading` kept for compatibility; the Siege Log page supplies its own <h1>.
function SiegeLog({ scans, onDelete, onSelect }) {
  const [page, setPage] = useState(0)
  const pageSize = 10
  const totalPages = Math.max(1, Math.ceil(scans.length / pageSize))
  const start = page * pageSize
  const pageScans = scans.slice(start, start + pageSize)

  return (
    <div className="flex flex-col" style={{ gap: 16 }}>
      {scans.length === 0 ? (
        <div className="panel-stone font-mono" style={{ padding: "18px 20px" }}>
          <div className="text-text" style={{ fontSize: 14 }}>The watch is quiet.</div>
          <div className="text-muted" style={{ fontSize: 12, marginTop: 4 }}>No scans yet — enter a URL above to run your first patrol.</div>
        </div>
      ) : (
        <div className="panel-stone" style={{ padding: "0 20px 6px" }}>
          <div className="grid items-center font-mono text-muted" style={{ gridTemplateColumns: COLS, height: 40, fontSize: 10, letterSpacing: ".16em", borderBottom: "1px solid rgba(130,160,210,.14)" }}>
            <span>TARGET</span><span>MODE</span><span>STATUS</span><span>SEVERITY</span><span>STARTED</span><span />
          </div>
          {pageScans.map((r) => {
            const done = r.status === "completed"
            const s = done ? SEV[r.analysis?.overall_risk?.level] : null
            return (
              <div
                key={r.id}
                onClick={() => onSelect(r.id)}
                className="grid items-center font-mono cursor-pointer"
                style={{ gridTemplateColumns: COLS, height: 46, fontSize: 13, borderBottom: "1px solid rgba(130,160,210,.08)" }}
              >
                <span className="text-text truncate">{r.target_url}</span>
                <span className="text-muted">{r.scan_type}</span>
                <span className="flex items-center text-muted" style={{ gap: 8 }}>
                  <span style={{ color: done ? "#3DD68C" : "#8A97A8", lineHeight: 0 }}>
                    <Icon id={done ? "i-seal" : "i-hourglass"} size={13} />
                  </span>
                  {r.status}
                </span>
                <span>
                  {s ? (
                    <span style={{ display: "inline-flex", alignItems: "center", gap: 6, height: 24, padding: "0 9px 0 7px", fontSize: 11, fontWeight: 700, letterSpacing: ".08em", color: s.c, background: `${s.c}1a`, border: `1px solid ${s.c}73` }}>
                      <span style={{ color: s.c, lineHeight: 0 }}><Icon id={s.icon} size={13} /></span>{s.label}
                    </span>
                  ) : (
                    <span className="text-faint">—</span>
                  )}
                </span>
                <span className="text-muted">{formatTime(r.created_at)}</span>
                <span
                  title="Delete scan"
                  onClick={(e) => { e.stopPropagation(); onDelete(r.id) }}
                  className="flex items-center justify-center"
                  style={{ width: 36, height: 36, color: "#8A97A8" }}
                >
                  <Icon id="i-strike" size={15} />
                </span>
              </div>
            )
          })}
        </div>
      )}

      {/* pagination */}
      {totalPages > 1 && (
        <div className="flex items-center font-mono text-muted" style={{ gap: 16, fontSize: 12 }}>
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="flex items-center"
            style={{ height: 36, padding: "0 14px", gap: 8, borderRadius: 2, backgroundColor: "#131A28", backgroundImage: "var(--tex-stone)", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)", opacity: page === 0 ? 0.4 : 1 }}
          >
            <span style={{ transform: "scaleX(-1)", lineHeight: 0 }}><Icon id="i-arrow-r" size={13} /></span>Prev
          </button>
          <span style={{ color: "#EAF1F8" }}>Page {page + 1} of {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="flex items-center"
            style={{ height: 36, padding: "0 14px", gap: 8, borderRadius: 2, backgroundColor: "#131A28", backgroundImage: "var(--tex-stone)", color: "#EAF1F8", border: "1px solid rgba(130,160,210,.22)", boxShadow: "inset 0 1px 0 rgba(234,241,248,.06), 0 2px 0 rgba(0,0,0,.45)", opacity: page >= totalPages - 1 ? 0.4 : 1 }}
          >
            Next<span style={{ lineHeight: 0 }}><Icon id="i-arrow-r" size={13} /></span>
          </button>
        </div>
      )}
    </div>
  )
}

export default SiegeLog
