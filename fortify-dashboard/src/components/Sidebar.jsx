/* global __APP_VERSION__ */  // injected from package.json by Vite (vite.config.js)
import { Watchtower, Scroll } from './icons/Icons'

const OPERATIONS = [
  { label: "Command", view: "command", badge: null, icon: Watchtower },
]
const COUNSEL = [
  { label: "Settings", view: "settings", badge: null, icon: Scroll },
]

function Sidebar({ view, onNavigate, provider }) {
  const renderItem = (item) => {
    const Icon = item.icon
    return (
      <div
        key={item.label}
        onClick={() => onNavigate(item.view)}
        className={`flex justify-between items-center px-2 py-1.5 rounded cursor-pointer ${
          view === item.view ? "bg-surface-2 text-text" : "text-muted hover:text-text hover:bg-surface-2"
        }`}
      >
        <span className="flex items-center gap-2.5 text-sm">
          {Icon && <Icon size={16} className="shrink-0" />}
          {item.label}
        </span>
        {item.badge && (
          <span className="font-mono text-[10px] text-accent bg-accent/20 px-2 py-0.5 rounded-full">{item.badge}</span>
        )}
      </div>
    )
  }

  return (
    <aside className="w-60 shrink-0 border-r border-border flex flex-col print:hidden">
      {/* wordmark */}
      <div className="p-4 border-b border-border flex items-center gap-3">
        <img src="/logo.png" alt="Fortify" className="w-10 h-10 rounded-full shrink-0" />
        <div>
          <div className="font-display text-2xl text-text leading-none">Fortify</div>
          <div className="font-mono text-[10px] text-accent tracking-widest">THE KEEP · v{__APP_VERSION__}</div>
        </div>
      </div>

      {/* nav */}
      <nav className="flex-1 p-3 space-y-4">
        <div>
          <div className="font-mono text-[10px] text-faint tracking-widest uppercase px-2 mb-2">Operations</div>
          {OPERATIONS.map(renderItem)}
        </div>
        <div>
          <div className="font-mono text-[10px] text-faint tracking-widest uppercase px-2 mb-2">Counsel</div>
          {COUNSEL.map(renderItem)}
        </div>
      </nav>

      {/* provider footer — reflects the active AI provider */}
      <div className="p-3 border-t border-border font-mono text-[10px] text-faint leading-relaxed">
        {provider === "gemini" ? (
          <>
            <div><span className="text-med">●</span> gemini · cloud</div>
            <div>google api · data leaves</div>
          </>
        ) : (
          <>
            <div><span className="text-low">●</span> ollama · llama3.1:8b</div>
            <div>127.0.0.1:11434 · nothing leaves</div>
          </>
        )}
      </div>
    </aside>
  )
}

export default Sidebar