import { useState } from 'react'
import { Icon } from './icons/Icons'

// Status = low-weight dot + label (a scan's lifecycle, not its risk).
// Live/in-progress rows stay plain — operators read status fast (per handoff).
const STATUS_DOT = {
  completed: "bg-low",
  running:   "bg-accent",
  pending:   "bg-accent",
  failed:    "bg-crit",
}
function statusDot(status) {
  return STATUS_DOT[status] || "bg-muted"
}

// Severity = the louder signal (a bordered, tinted badge + shield icon).
const SEVERITY_BADGE = {
  critical: "text-crit border-crit/40 bg-crit/10",
  high:     "text-high border-high/40 bg-high/10",
  medium:   "text-med border-med/40 bg-med/10",
  low:      "text-low border-low/40 bg-low/10",
}
const SEV_ICON = { critical: "i-sev-crit", high: "i-sev-high", medium: "i-sev-med", low: "i-sev-low" }

function formatTime(iso) {
  return new Date(iso).toLocaleString()
}

function SiegeLog({ scans, onDelete, onSelect }) {
  const [page, setPage] = useState(0)          // local UI state
  const pageSize = 10
  const totalPages = Math.max(1, Math.ceil(scans.length / pageSize))
  const start = page * pageSize
  const pageScans = scans.slice(start, start + pageSize)   // just this page's rows

  return (
    <div className="mt-8">
      {/* header */}
      <div className="flex items-baseline gap-3 mb-3">
        <h2 className="font-display text-xl">Siege Log</h2>
        <span className="font-mono text-xs text-muted">{scans.length} scans · newest first</span>
      </div>

      {/* empty state */}
      {scans.length === 0 ? (
        <div className="font-mono py-6">
          <div className="text-sm text-text">The watch is quiet.</div>
          <div className="text-xs text-muted mt-1">No scans yet — enter a URL above to run your first patrol.</div>
        </div>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="font-mono text-[10px] text-faint uppercase tracking-widest text-left border-b border-border">
              <th className="py-2 pr-6 font-normal">Target</th>
              <th className="py-2 pr-6 font-normal">Mode</th>
              <th className="py-2 pr-6 font-normal">Status</th>
              <th className="py-2 pr-6 font-normal">Severity</th>
              <th className="py-2 pr-6 font-normal">Started</th>
              <th className="py-2 font-normal"></th>
            </tr>
          </thead>
          <tbody>
            {pageScans.map((scan) => {
              const level = scan.analysis?.overall_risk?.level
              return (
              <tr
                key={scan.id}
                onClick={() => onSelect(scan.id)}
                className="border-b border-border/50 hover:bg-surface-2 cursor-pointer"
              >
                <td className="py-2 pr-6 font-mono text-text whitespace-nowrap">{scan.target_url}</td>
                <td className="py-2 pr-6 font-mono text-muted">{scan.scan_type}</td>
                <td className="py-2 pr-6">
                  <span className="flex items-center gap-2 font-mono text-xs text-muted">
                    <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${statusDot(scan.status)}`} />
                    {scan.status}
                  </span>
                </td>
                <td className="py-2 pr-6">
                  {level ? (
                    <span className={`inline-flex items-center gap-1 font-mono text-[10px] font-bold uppercase px-2 py-0.5 rounded-[2px] border ${SEVERITY_BADGE[level] || "text-muted border-border"}`}>
                      <Icon id={SEV_ICON[level]} size={12} />
                      {level}
                    </span>
                  ) : (
                    <span className="font-mono text-xs text-faint">—</span>
                  )}
                </td>
                <td className="py-2 pr-6 font-mono text-muted whitespace-nowrap">{formatTime(scan.created_at)}</td>
                <td className="py-2">
                  <button
                    onClick={(e) => { e.stopPropagation(); onDelete(scan.id) }}
                    className="text-faint hover:text-crit font-mono"
                    title="Delete scan"
                  >
                    ✕
                  </button>
                </td>
              </tr>
              )
            })}
          </tbody>
        </table>
      )}

      {/* pagination controls — only when more than one page */}
      {totalPages > 1 && (
        <div className="flex items-center gap-4 mt-4 font-mono text-xs text-muted">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 border border-border rounded hover:text-text disabled:opacity-30 disabled:hover:text-muted"
          >
            ← Prev
          </button>
          <span>Page {page + 1} of {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 border border-border rounded hover:text-text disabled:opacity-30 disabled:hover:text-muted"
          >
            Next →
          </button>
        </div>
      )}
    </div>
  )
}

export default SiegeLog