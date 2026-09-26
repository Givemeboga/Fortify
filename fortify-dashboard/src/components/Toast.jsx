import { Icon } from './icons/Icons'

const KIND = {
  dispatch: { icon: "i-pennant", anim: "anim-flutter", color: "text-accent", title: "Dispatch arrived" },
  critical: { icon: "i-brazier", anim: "anim-flick",   color: "text-crit",   title: "A signal fire is lit" },
}

export default function Toast({ toasts, onDismiss }) {
  if (!toasts.length) return null
  return (
    <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-2 print:hidden">
      {toasts.map((t) => {
        const k = KIND[t.kind] || KIND.dispatch
        return (
          <div
            key={t.id}
            className="panel-iron flex items-center gap-3 pl-3 pr-3 py-2.5 min-w-[240px] max-w-[320px]"
            style={{ animation: "rise 0.3s ease both" }}
          >
            <span className={`${k.anim} shrink-0`}>
              <Icon id={k.icon} size={22} className={k.color} />
            </span>
            <div className="flex-1 min-w-0">
              <div className="font-mono text-sm text-text">{k.title}</div>
              {t.detail && <div className="font-mono text-[11px] text-muted truncate">{t.detail}</div>}
            </div>
            <button
              onClick={() => onDismiss(t.id)}
              className="text-faint hover:text-text shrink-0"
              aria-label="Dismiss"
            >
              <Icon id="i-strike" size={14} />
            </button>
          </div>
        )
      })}
    </div>
  )
}
