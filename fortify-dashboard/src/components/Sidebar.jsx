/* global __APP_VERSION__ */  // injected from package.json by Vite (vite.config.js)
import { useState } from 'react'
import { Icon } from './icons/Icons'

// Keep App theme — recessed nav, stone-textured aside, lantern footer.
const OPERATIONS = [
  { label: "Command", view: "command", icon: "i-keep" },
  { label: "Siege Log", view: "log", icon: "i-swords" },
  { label: "Watch", view: "watch", icon: "i-hourglass" },
]
const COUNSEL = [
  { label: "Counsel", view: "settings", icon: "i-tome" },
]

function Sidebar({ view, onNavigate, provider }) {
  const [collapsed, setCollapsed] = useState(() => {
    try { return localStorage.getItem("fortify_sidebar_collapsed") === "1" } catch { return false }
  })

  function toggle() {
    setCollapsed((c) => {
      const next = !c
      try { localStorage.setItem("fortify_sidebar_collapsed", next ? "1" : "0") } catch { /* ignore */ }
      return next
    })
  }

  const renderItem = (item) => {
    const active = view === item.view
    return (
      <div
        key={item.label}
        onClick={() => onNavigate(item.view)}
        title={collapsed ? item.label : undefined}
        className="flex items-center h-10 rounded-[2px] cursor-pointer transition-colors"
        style={{
          gap: collapsed ? 0 : 12,
          padding: collapsed ? 0 : "0 12px",
          justifyContent: collapsed ? "center" : "flex-start",
          ...(active
            ? { backgroundColor: "#0A0E16", boxShadow: "inset 0 2px 4px rgba(0,0,0,.6), inset 0 0 0 1px rgba(130,160,210,.14)" }
            : {}),
        }}
      >
        <Icon id={item.icon} size={18} className={active ? "text-accent" : "text-muted"} />
        {!collapsed && (
          <span className="font-sans text-sm font-medium" style={{ color: active ? "#EAF1F8" : "#8A97A8" }}>{item.label}</span>
        )}
      </div>
    )
  }

  return (
    <aside
      className="shrink-0 flex flex-col border-r border-border print:hidden"
      style={{
        width: collapsed ? 64 : 232,
        backgroundColor: "#0F1420",
        backgroundImage: "var(--tex-stone)",
        boxShadow: "inset -3px 0 0 #0A0E16, inset -4px 0 0 rgba(130,160,210,.08)",
        transition: "width .18s ease",
      }}
    >
      {/* brand block */}
      <div
        className="flex items-center border-b border-border"
        style={{ padding: collapsed ? "18px 0" : "18px 16px", gap: 12, justifyContent: collapsed ? "center" : "flex-start", backgroundColor: "rgba(19,26,40,.6)" }}
      >
        <img src="/logo.png" alt="Fortify" className="rounded-full shrink-0" style={{ width: 44, height: 44 }} />
        {!collapsed && (
          <div>
            <div className="font-display text-text" style={{ fontSize: 28, lineHeight: 1 }}>Fortify</div>
            <div className="font-mono text-accent whitespace-nowrap" style={{ fontSize: 10, letterSpacing: ".12em", marginTop: 3 }}>
              THE KEEP · v{__APP_VERSION__}
            </div>
          </div>
        )}
      </div>

      {/* collapse toggle */}
      <button
        onClick={toggle}
        title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        className="flex items-center border-b border-border text-muted hover:text-text"
        style={{ height: 34, padding: collapsed ? 0 : "0 14px", justifyContent: collapsed ? "center" : "flex-end" }}
      >
        <span style={{ transform: collapsed ? "none" : "scaleX(-1)", lineHeight: 0 }}><Icon id="i-arrow-r" size={14} /></span>
      </button>

      {/* nav */}
      <nav className="flex-1 flex flex-col" style={{ padding: collapsed ? "18px 8px" : "18px 10px", gap: 20 }}>
        <div className="flex flex-col gap-1">
          {!collapsed && <div className="font-mono text-muted" style={{ padding: "0 12px 6px", letterSpacing: ".16em", fontSize: 10 }}>OPERATIONS</div>}
          {OPERATIONS.map(renderItem)}
        </div>
        <div className="flex flex-col gap-1">
          {!collapsed && <div className="font-mono text-muted" style={{ padding: "0 12px 6px", letterSpacing: ".16em", fontSize: 10 }}>COUNSEL</div>}
          {COUNSEL.map(renderItem)}
        </div>
      </nav>

      {/* footer — active provider (hidden when collapsed) */}
      {!collapsed && (
        <div className="border-t border-border font-mono text-muted" style={{ padding: "14px 16px", lineHeight: 1.7, fontSize: 11 }}>
          {provider === "ollama" ? (
            <>
              <div className="flex items-center gap-2"><Icon id="i-lantern" size={13} className="text-low" />ollama · llama3.1:8b</div>
              <div>127.0.0.1:11434 · nothing leaves</div>
            </>
          ) : provider === "custom" ? (
            <>
              <div className="flex items-center gap-2"><Icon id="i-lantern" size={13} className="text-med" />custom endpoint</div>
              <div>privacy depends on endpoint</div>
            </>
          ) : (
            <>
              <div className="flex items-center gap-2"><Icon id="i-lantern" size={13} className="text-med" />{provider} · cloud</div>
              <div>api key · data leaves</div>
            </>
          )}
        </div>
      )}
    </aside>
  )
}

export default Sidebar
