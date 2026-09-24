import { useState, useEffect } from 'react'

const LEVEL_STYLES = {
  critical: { chip: "bg-crit/15 text-crit border-crit/40", bar: "bg-crit" },
  high:     { chip: "bg-high/15 text-high border-high/40", bar: "bg-high" },
  medium:   { chip: "bg-med/15 text-med border-med/40",   bar: "bg-med" },
  low:      { chip: "bg-low/15 text-low border-low/40",   bar: "bg-low" },
}
function levelStyle(level) {
  return LEVEL_STYLES[level] || { chip: "bg-white/5 text-muted border-border", bar: "bg-muted" }
}

function useTypewriter(text, speed = 10) {
  const [shown, setShown] = useState("")
  useEffect(() => {
    setShown("")
    if (!text) return
    let i = 0
    const id = setInterval(() => {
      i += 2
      setShown(text.slice(0, i))
      if (i >= text.length) clearInterval(id)
    }, speed)
    return () => clearInterval(id)
  }, [text, speed])
  return shown
}

function Panel({ title, children }) {
  return (
    <div className="border border-border rounded p-4 bg-surface break-inside-avoid">
      <div className="font-mono text-[10px] text-faint uppercase tracking-widest mb-3">{title}</div>
      {children}
    </div>
  )
}

function BattleReport({ scanId, onBack }) {
  const [scan, setScan] = useState(null)

  async function loadScan() {
    const res = await fetch(`http://localhost:8500/scans/${scanId}`)
    setScan(await res.json())
  }

  useEffect(() => { loadScan() }, [scanId])

  // Whenever the scan is analyzing, keep a poll running; clean it up when the
  // status leaves "analyzing" or the component unmounts. This also resumes
  // polling automatically when re-entering a scan that's mid-analysis.
  useEffect(() => {
    if (scan?.analysis_status !== "analyzing") return
    const timer = setInterval(loadScan, 2000)
    return () => clearInterval(timer)
  }, [scan?.analysis_status])

  async function handleAnalyze() {
    const provider = localStorage.getItem("fortify_provider") || "ollama"
    const apiKey = localStorage.getItem("fortify_gemini_key") || ""

    const res = await fetch(`http://localhost:8500/scans/${scanId}/analyze`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider, api_key: apiKey }),
    })
    if (!res.ok) { alert(`Analysis failed to start (${res.status})`); return }

    // flip to "analyzing" — the useEffect above sees this and starts polling
    setScan((prev) => ({ ...prev, analysis_status: "analyzing" }))
  }

  if (!scan) return <div className="font-mono text-muted">Loading…</div>

  const results = scan.results || {}
  const analysis = scan.analysis
  const status = scan.analysis_status

  return (
    <div className="max-w-6xl">
      {/* header — nav row hidden when printing the PDF */}
      <div className="flex items-center justify-between mb-4 print:hidden">
        <button onClick={onBack} className="font-mono text-xs text-muted hover:text-text">← Back to Command</button>
        <button onClick={() => window.print()} className="font-mono text-xs text-accent hover:brightness-110">Export PDF ↓</button>
      </div>
      <div className="flex items-baseline gap-3">
        <h1 className="font-display text-3xl">Battle Report</h1>
        <span className="font-mono text-sm text-muted">{scan.target_url}</span>
      </div>
      <div className="font-mono text-xs text-muted mt-1">{scan.scan_type} · {scan.status}</div>
      {/* print-only: timestamp + branding for the exported report */}
      <div className="hidden print:block font-mono text-xs text-faint mt-1">
        Generated {new Date().toLocaleString()} · Fortify
      </div>

      {/* two columns on screen; stacked into one column for the printed page */}
      <div className="mt-6 grid grid-cols-3 gap-6 items-start print:grid-cols-1 print:gap-4">

        {/* LEFT — AI analysis */}
        <div className="col-span-2">
          {status === "analyzing" ? (
            <div className="flex items-center gap-3 border border-border rounded p-4 bg-surface">
              <div className="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin" />
              <span className="font-mono text-sm text-muted animate-pulse">
                Consulting the war council… the model is thinking.
              </span>
            </div>
          ) : status === "failed" ? (
            <div className="border border-crit/40 rounded p-4 bg-surface">
              <div className="font-mono text-sm text-crit mb-2">
                {analysis?.error || "Analysis failed — the model errored or was unreachable."}
                </div>
              <button
                onClick={handleAnalyze}
                className="bg-accent text-bg font-semibold px-4 py-1.5 rounded hover:brightness-110 print:hidden"
              >
                Retry
              </button>
            </div>
          ) : analysis ? (
            <AnalysisPanel analysis={analysis} />
          ) : (
            <button
              onClick={handleAnalyze}
              disabled={scan.status !== "completed"}
              className="bg-accent text-bg font-semibold px-5 py-2 rounded hover:brightness-110 disabled:opacity-40 print:hidden"
            >
              Analyze with AI
            </button>
          )}
        </div>

        {/* RIGHT — raw scan details, stacked */}
        <div className="space-y-4">
          {results.tls && (
            <Panel title="TLS">
              <div className="font-mono text-xs space-y-1">
                <div>version: <span className="text-text">{results.tls.tls_version || "—"}</span></div>
                <div>cert valid: <span className={results.tls.cert_valid ? "text-low" : "text-crit"}>{String(results.tls.cert_valid)}</span></div>
                <div>cert expired: <span className={results.tls.cert_expired ? "text-crit" : "text-low"}>{String(results.tls.cert_expired)}</span></div>
                <div>cipher: <span className="text-muted break-all">{results.tls.cipher_suite || "—"}</span></div>
              </div>
            </Panel>
          )}

          {results.headers && (
            <Panel title="Headers">
              <div className="font-mono text-xs space-y-2">
                <div>
                  <div className="text-faint">missing:</div>
                  {results.headers.missing_headers?.length
                    ? results.headers.missing_headers.map((h) => <div key={h} className="text-high">· {h}</div>)
                    : <div className="text-low">none</div>}
                </div>
                <div>
                  <div className="text-faint">leaky:</div>
                  {Object.keys(results.headers.leaky_headers || {}).length
                    ? Object.entries(results.headers.leaky_headers).map(([k, v]) => {
                        const disc = results.headers.version_disclosures?.[k]
                        return (
                          <div key={k} className="break-all">
                            <span className={disc ? "text-high" : "text-med"}>· {k}: {v}</span>
                            {disc && <span className="text-high"> → discloses {disc.software} {disc.version}</span>}
                          </div>
                        )
                      })
                    : <div className="text-low">none</div>}
                </div>
              </div>
            </Panel>
          )}

          {results.status && (
            <Panel title="Sensitive paths">
              <div className="font-mono text-xs space-y-1">
                {Object.entries(results.status).map(([path, info]) => (
                  <div key={path} className="flex justify-between gap-4">
                    <span className="text-muted break-all">{path}</span>
                    <span className={info.exposed ? "text-high" : "text-faint"}>{info.status_code}</span>
                  </div>
                ))}
              </div>
            </Panel>
          )}

          {["sqli", "sqli_boolean", "xss", "path_traversal"].map((check) =>
            results[check] ? (
              <Panel key={check} title={check.replace("_", " ")}>
                <div className="font-mono text-xs">
                  <div className={results[check].vulnerable ? "text-crit" : "text-low"}>
                    {results[check].vulnerable ? "VULNERABLE" : "not vulnerable"}
                  </div>
                  {results[check].findings?.map((f, i) => (
                    <div key={i} className="text-muted mt-1">· param "{f.parameter}"</div>
                  ))}
                </div>
              </Panel>
            ) : null
          )}
        </div>

      </div>
    </div>
  )
}

function AnalysisPanel({ analysis }) {
  const risk = analysis.overall_risk || {}
  const rs = levelStyle(risk.level)
  const typedSummary = useTypewriter(analysis.summary)

  return (
    <div className="border border-border rounded-lg bg-surface overflow-hidden">
      {/* header */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-border bg-surface-2">
        <span className="font-mono text-[10px] text-faint uppercase tracking-widest">Model Notes</span>
        <div className="flex items-center gap-2">
          <span className="font-mono text-[10px] text-faint uppercase tracking-widest">Overall Risk</span>
          <span className={`font-mono text-xs font-bold uppercase px-2.5 py-1 rounded border ${rs.chip}`}>
            {risk.level} · {risk.score}
          </span>
        </div>
      </div>

      <div className="p-5">
        <p className="text-sm text-muted leading-relaxed min-h-[3rem]">
          {typedSummary}<span className="text-accent animate-pulse print:hidden">▍</span>
        </p>

        <div className="mt-5 space-y-2.5">
          {analysis.findings?.map((f, i) => {
            const fs = levelStyle(f.severity?.level)
            return (
              <div
                key={i}
                className="flex gap-3 rounded border border-border bg-bg/40 p-3 break-inside-avoid"
                style={{ animation: "rise 0.4s ease both", animationDelay: `${300 + i * 140}ms` }}
              >
                <div className={`w-0.5 rounded ${fs.bar}`} />
                <div className="flex-1">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className={`font-mono text-[10px] font-bold uppercase px-1.5 py-0.5 rounded border ${fs.chip}`}>
                      {f.severity?.level} · {f.severity?.score}
                    </span>
                    <span className="text-sm text-text font-medium">{f.vulnerability}</span>
                  </div>
                  <div className="text-xs text-muted mt-1.5 leading-relaxed">{f.explanation}</div>
                  <div className="text-xs text-low mt-1.5 leading-relaxed">→ {f.remediation}</div>
                </div>
              </div>
            )
          })}
        </div>

        {analysis.priority_order?.length > 0 && (
          <div className="mt-5 border-t border-border pt-4">
            <div className="font-mono text-[10px] text-faint uppercase tracking-widest mb-2">Priority Order</div>
            <ol className="space-y-1">
              {analysis.priority_order.map((p, i) => (
                <li key={i} className="flex gap-2 text-xs text-muted">
                  <span className="font-mono text-faint">{String(i + 1).padStart(2, "0")}</span>
                  <span>{p}</span>
                </li>
              ))}
            </ol>
          </div>
        )}

        <div className="mt-5 text-[11px] text-faint italic">{analysis.disclaimer}</div>
      </div>
    </div>
  )
}

export default BattleReport