import { useState } from 'react'

const STATUS_COLORS = {
  completed: "text-low",     // green
  running:   "text-accent",  // blue
  pending:   "text-muted",   // grey
  failed:    "text-crit",    // red
}
function statusColor(status) {
  return STATUS_COLORS[status] || "text-muted"
}

function formatTime(iso) {
  return new Date(iso).toLocaleString()
}

function SiegeLog({ scans, onDelete }) {
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
        <div className="font-mono text-sm text-faint py-6">
          The watch is quiet — launch your first patrol.
        </div>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="font-mono text-[10px] text-faint uppercase tracking-widest text-left border-b border-border">
              <th className="py-2 pr-6 font-normal">Target</th>
              <th className="py-2 pr-6 font-normal">Mode</th>
              <th className="py-2 pr-6 font-normal">Status</th>
              <th className="py-2 pr-6 font-normal">Started</th>
              <th className="py-2 font-normal"></th>
            </tr>
          </thead>
          <tbody>
            {pageScans.map((scan) => (
              <tr key={scan.id} className="border-b border-border/50 hover:bg-surface-2">
                <td className="py-2 pr-6 font-mono text-text whitespace-nowrap">{scan.target_url}</td>
                <td className="py-2 pr-6 font-mono text-muted">{scan.scan_type}</td>
                <td className={`py-2 pr-6 font-mono ${statusColor(scan.status)}`}>{scan.status}</td>
                <td className="py-2 pr-6 font-mono text-muted whitespace-nowrap">{formatTime(scan.created_at)}</td>
                <td className="py-2">
                  <button
                    onClick={() => onDelete(scan.id)}
                    className="text-faint hover:text-crit font-mono"
                    title="Delete scan"
                  >
                    ✕
                  </button>
                </td>
              </tr>
            ))}
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