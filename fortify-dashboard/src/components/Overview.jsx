import { Icon } from './icons/Icons'
import StainedGlassFacade from './StainedGlassFacade'

// Provider accent (matches the sidebar/toggle/card convention): ollama azure, gemini orange.
const PROVIDER_COLOR = { ollama: "#2FA4FF", gemini: "#F97316" }
const providerColor = (p) => PROVIDER_COLOR[p] || "#2FA4FF"

/* ── data helpers ─────────────────────────────────────────────────────────── */

const SEV = [
  { key: "critical", label: "CRITICAL", c: "#FF4D4D", glow: "rgba(255,77,77,.45)", icon: "i-sev-crit" },
  { key: "high",     label: "HIGH",     c: "#F97316", glow: "rgba(249,115,22,.45)", icon: "i-sev-high" },
  { key: "medium",   label: "MEDIUM",   c: "#F5C518", glow: "rgba(245,197,24,.45)", icon: "i-sev-med" },
  { key: "low",      label: "LOW",      c: "#3DD68C", glow: "rgba(61,214,140,.45)", icon: "i-sev-low" },
]
const SEV_BY = Object.fromEntries(SEV.map((s) => [s.key, s]))

// Vulnerability rows (Command spec icons): vial / scroll / ladder / key / banner.
const THREATS = [
  { key: "sqli",      name: "SQL Injection",        icon: "i-vial",   test: (r) => r.sqli?.vulnerable || r.sqli_boolean?.vulnerable },
  { key: "xss",       name: "Cross-Site Scripting", icon: "i-scroll", test: (r) => r.xss?.vulnerable },
  { key: "traversal", name: "Path Traversal",       icon: "i-ladder", test: (r) => r.path_traversal?.vulnerable },
  { key: "tls",       name: "Weak TLS",             icon: "i-key",    test: (r) => r.tls && (r.tls.cert_valid === false || r.tls.cert_expired === true) },
  { key: "headers",   name: "Missing Headers",      icon: "i-banner", test: (r) => (r.headers?.missing_headers?.length || 0) > 0 },
]

function severityCounts(scans) {
  const c = { critical: 0, high: 0, medium: 0, low: 0 }
  for (const s of scans) {
    if (s.analysis_status !== "completed") continue
    const lvl = s.analysis?.overall_risk?.level
    if (lvl in c) c[lvl] += 1
  }
  return c
}

function threatCounts(scans) {
  const c = Object.fromEntries(THREATS.map((t) => [t.key, 0]))
  for (const s of scans) {
    const r = s.results
    if (!r) continue
    for (const t of THREATS) if (t.test(r)) c[t.key] += 1
  }
  return c
}

function opsSummary(scans) {
  const targets = new Set(scans.map((s) => s.target_url))
  const reports = scans.filter((s) => s.analysis_status === "completed" && s.analysis?.model).length
  return { patrols: scans.length, targets: targets.size, reports }
}

// Per-model usage across analyses, sorted most-used first — used for both the
// "most consulted" line and the per-provider share bar.
function modelUsage(scans) {
  const map = {}
  for (const s of scans) {
    if (s.analysis_status !== "completed") continue
    const m = s.analysis?.model
    if (!m) continue
    if (!map[m]) map[m] = { model: m, provider: s.analysis?.provider, count: 0 }
    map[m].count += 1
  }
  return Object.values(map).sort((a, b) => b.count - a.count)
}

// Last 7 days of AI reports (totals per day) for the candle chart.
function reportDays(scans, dayCount = 7) {
  const byDay = {}
  for (const s of scans) {
    if (s.analysis_status !== "completed" || !s.analysis?.model) continue
    const key = new Date(s.created_at).toISOString().slice(0, 10)
    if (!byDay[key]) byDay[key] = { total: 0, providers: {} }
    byDay[key].total += 1
    const p = s.analysis?.provider || "ollama"
    byDay[key].providers[p] = (byDay[key].providers[p] || 0) + 1
  }
  const out = []
  const today = new Date()
  for (let i = dayCount - 1; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    const rec = byDay[key] || { total: 0, providers: {} }
    // the provider that produced most of that day's reports → the candle's flame colour
    const provider = Object.entries(rec.providers).sort((a, b) => b[1] - a[1])[0]?.[0] || null
    out.push({ key, total: rec.total, label: `${key.slice(8, 10)}/${key.slice(5, 7)}`, provider })
  }
  return out
}

// Candle geometry for one day (34×120 viewBox, holder at y118) — matches the
// handoff support script: rounded-top body (+ a drip on lit days), wick, flame.
function candle(v, max) {
  const lit = v > 0
  const h = lit ? 8 + (80 * v) / max : 8
  const t = 118 - h
  return {
    body: `M11 118V${t + 3}Q11 ${t} 17 ${t}Q23 ${t} 23 ${t + 3}V118Z${lit ? `M20 ${t + 1}V${t + 9}` : ""}`,
    wick: lit ? `M17 ${t}V${t - 5}` : "",
    flame: lit ? `M17 ${t - 5}C12.5 ${t - 9} 14 ${t - 17} 17 ${t - 23}C20 ${t - 17} 21.5 ${t - 9} 17 ${t - 5}Z` : "",
    gy: t - 13,
    edge: lit ? "#8A97A8" : "#55606F",
    fo: lit ? 1 : 0,
    lc: lit ? "#EAF1F8" : "#55606F",
  }
}

function formatTime(iso) {
  const d = new Date(iso)
  const p = (n) => String(n).padStart(2, "0")
  return `${p(d.getDate())}/${p(d.getMonth() + 1)}/${d.getFullYear()} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

/* ── component ────────────────────────────────────────────────────────────── */

export default function Overview({ scans, onSelect, onViewAll }) {
  const ops = opsSummary(scans)
  const sev = severityCounts(scans)
  const threats = threatCounts(scans)
  const models = modelUsage(scans)
  const tm = models[0] || null
  const totalReports = models.reduce((sum, m) => sum + m.count, 0)
  const days = reportDays(scans)
  const maxDay = Math.max(1, ...days.map((d) => d.total))
  const recent = scans.slice(0, 5)

  const facadeStats = [
    { n: ops.patrols, label: "PATROLS RUN" },
    { n: ops.targets, label: "TARGETS WATCHED" },
    { n: ops.reports, label: "REPORTS DRAFTED" },
  ]

  return (
    <div style={{ width: "100%", maxWidth: 1128, margin: "0 auto", display: "flex", flexDirection: "column", gap: 28 }}>
      {/* page header */}
      <div className="flex flex-col gap-2">
        <h1 className="font-display font-semibold leading-none text-text" style={{ fontSize: 44 }}>Command</h1>
        <span className="font-mono text-muted" style={{ fontSize: 12, letterSpacing: ".16em" }}>// STATE OF THE KEEP</span>
      </div>

      {/* stained-glass stat façade */}
      <StainedGlassFacade stats={facadeStats} />

      {/* severity strip */}
      <div className="grid gap-3" style={{ gridTemplateColumns: "repeat(4,minmax(0,1fr))" }}>
        {SEV.map((s) => (
          <div key={s.key} className="panel-stone flex items-center" style={{ gap: 14, padding: "16px 20px" }}>
            <span style={{ color: s.c, filter: `drop-shadow(0 0 6px ${s.glow})`, lineHeight: 0 }}>
              <Icon id={s.icon} size={26} />
            </span>
            <div>
              <div style={{ fontFamily: "'Grenze Gotisch',serif", fontSize: 34, lineHeight: 1, color: s.c }}>{sev[s.key]}</div>
              <div className="font-mono text-muted" style={{ marginTop: 4, fontSize: 11, letterSpacing: ".16em" }}>{s.label}</div>
            </div>
          </div>
        ))}
      </div>

      {/* two columns: Threats sighted + The Scriptorium */}
      <div className="grid gap-6" style={{ gridTemplateColumns: "minmax(0,1fr) minmax(0,1fr)" }}>
        {/* Threats sighted */}
        <div className="flex flex-col gap-3">
          <h2 className="font-display font-semibold leading-none text-text" style={{ fontSize: 26 }}>Threats sighted</h2>
          <div className="panel-stone flex-1" style={{ padding: "6px 20px" }}>
            {THREATS.map((t) => {
              const n = threats[t.key]
              const on = n > 0
              return (
                <div
                  key={t.key}
                  className="flex items-center"
                  style={{ gap: 12, height: 52, borderBottom: "1px solid rgba(130,160,210,.08)", opacity: on ? 1 : 0.35 }}
                >
                  <span style={{ color: on ? "#F97316" : "#8A97A8", lineHeight: 0 }}><Icon id={t.icon} size={18} /></span>
                  <span className="font-mono text-text" style={{ fontSize: 14 }}>{t.name}</span>
                  <span className="font-mono text-text" style={{ marginLeft: "auto", fontSize: 14, fontWeight: 700 }}>{n}</span>
                </div>
              )
            })}
          </div>
        </div>

        {/* The Scriptorium */}
        <div className="flex flex-col gap-3">
          <h2 className="font-display font-semibold leading-none text-text" style={{ fontSize: 26 }}>The Scriptorium</h2>
          <div className="panel-stone flex-1 flex flex-col" style={{ padding: "18px 20px", gap: 6 }}>
            <div className="flex items-center font-mono text-muted" style={{ gap: 8, fontSize: 10, letterSpacing: ".16em" }}>
              <span style={{ color: "#2FA4FF", lineHeight: 0 }}><Icon id="i-quill" size={14} /></span>MOST CONSULTED
            </div>
            <div className="font-mono text-text" style={{ fontSize: 22 }}>{tm ? tm.model : "—"}</div>
            <div className="font-mono text-muted" style={{ fontSize: 13 }}>{tm ? `${tm.count} report${tm.count === 1 ? "" : "s"} drafted` : "no reports yet"}</div>

            <div className="font-mono text-muted" style={{ marginTop: 14, fontSize: 10, letterSpacing: ".16em" }}>REPORTS OVER TIME</div>
            <div className="flex" style={{ gap: 10, marginTop: 6 }}>
              {/* y-axis */}
              <div className="font-mono" style={{ height: 120, display: "flex", flexDirection: "column", justifyContent: "space-between", fontSize: 10, color: "#55606F", textAlign: "right" }}>
                <span>{maxDay}</span><span>{Math.round(maxDay / 2)}</span><span>0</span>
              </div>
              <div className="flex-1 flex flex-col">
                <div style={{ height: 120, display: "flex", alignItems: "flex-end", justifyContent: "space-between", padding: "0 8px" }}>
                  {days.map((d) => {
                    const b = candle(d.total, maxDay)
                    const c = providerColor(d.provider)
                    const glowId = d.provider === "gemini" ? "cd-glow-amber" : "cd-glow"
                    return (
                      <svg key={d.key} width="34" height="120" viewBox="0 0 34 120" style={{ overflow: "visible" }}>
                        <ellipse cx="17" cy={b.gy} rx="15" ry="20" fill={`url(#${glowId})`} opacity={b.fo} />
                        <g filter="url(#ink)" strokeLinecap="round" strokeLinejoin="round">
                          <path d={b.body} fill="#131A28" stroke={b.edge} strokeWidth="1.6" />
                          {b.wick && <path d={b.wick} fill="none" stroke="#8A97A8" strokeWidth="1.2" />}
                          {b.flame && <path d={b.flame} fill={c} stroke="#EAF1F8" strokeWidth=".8" opacity={b.fo} />}
                        </g>
                        <path d="M8 119H26" stroke="#55606F" strokeWidth="2" strokeLinecap="round" />
                      </svg>
                    )
                  })}
                </div>
                {/* stone ledge */}
                <div style={{ height: 9, backgroundColor: "#131A28", backgroundImage: "var(--tex-stone)", borderTop: "1px solid rgba(234,241,248,.1)", boxShadow: "0 2px 0 rgba(0,0,0,.55), inset 0 -1px 0 rgba(0,0,0,.5)" }} />
                <div style={{ display: "flex", justifyContent: "space-between", padding: "6px 8px 0" }}>
                  {days.map((d) => (
                    <span key={d.key} className="font-mono" style={{ width: 34, textAlign: "center", fontSize: 10, color: candle(d.total, maxDay).lc }}>{d.label}</span>
                  ))}
                </div>
              </div>
            </div>

            {/* per-model share — legend rows + segmented bar, coloured by provider */}
            <div style={{ marginTop: 14, display: "flex", flexDirection: "column", gap: 6 }}>
              {models.length === 0 && <div className="font-mono text-muted" style={{ fontSize: 12 }}>—</div>}
              {models.map((m) => {
                const pct = totalReports ? Math.round((m.count / totalReports) * 100) : 0
                const c = providerColor(m.provider)
                return (
                  <div key={m.model} className="flex items-center font-mono text-text" style={{ gap: 8, fontSize: 12 }}>
                    <span style={{ color: c, lineHeight: 0 }}><Icon id="i-quill" size={14} /></span>{m.model}
                    <span className="text-muted" style={{ marginLeft: "auto" }}>{m.count} · {pct}%</span>
                  </div>
                )
              })}
            </div>
            <div style={{ height: 12, marginTop: 8, padding: 2, boxSizing: "border-box", backgroundColor: "#0A0E16", border: "1px solid rgba(130,160,210,.24)", boxShadow: "inset 0 2px 3px rgba(0,0,0,.7)", display: "flex", gap: 2 }}>
              {models.map((m) => {
                const c = providerColor(m.provider)
                return <div key={m.model} style={{ height: "100%", width: `${totalReports ? (m.count / totalReports) * 100 : 0}%`, background: `repeating-linear-gradient(90deg,${c} 0 26px,#0A0E16 26px 28px)`, boxShadow: `0 0 10px ${c}59` }} />
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
