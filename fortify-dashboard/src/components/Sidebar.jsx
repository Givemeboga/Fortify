// nav items as data
const OPERATIONS = [
  { label: "Command", badge: null },
  { label: "Siege Log", badge: 34 },
  { label: "Battle Reports", badge: null },
]

function Sidebar() {
  return (
    <aside className="w-60 shrink-0 border-r border-border flex flex-col">

      {/* wordmark */}
      <div className="p-4 border-b border-border">
        <div className="font-display text-2xl text-text">Fortify</div>
        <div className="font-mono text-[10px] text-accent tracking-widest">THE KEEP · v0.4.2</div>
      </div>

      {/* nav */}
      <nav className="flex-1 p-3">
        <div className="font-mono text-[10px] text-faint tracking-widest uppercase px-2 mb-2">
          Operations
        </div>

        {/* TODO: map over OPERATIONS -> one row per item.
            Each row: a clickable div/a with the label on the left,
            and (only if item.badge exists) the badge on the right. */}
        {OPERATIONS.map((item) => (
            <div key={item.label} className="flex justify-between items-center px-2 py-1 rounded hover:bg-bg-hover cursor-pointer">
              <span className="font-mono text-[10px] text-faint">{item.label}</span>
              {item.badge !== null && (
                <span className="font-mono text-[10px] text-accent bg-accent/20 px-2 py-1 rounded-full">
                  {item.badge}
                </span>
              )}
            </div>
        ))}

      </nav>

      {/* ollama footer — the local-LLM privacy touch */}
      <div className="p-3 border-t border-border font-mono text-[10px] text-faint leading-relaxed">
        <div><span className="text-low">●</span> ollama · llama3.1:8b</div>
        <div>127.0.0.1:11434 · nothing leaves</div>
      </div>

    </aside>
  )
}

export default Sidebar