import { Icon } from './icons/Icons'

/** Portcullis gate that grinds up and back — the loading state. */
export function GateLoader() {
  return (
    <div className="relative overflow-hidden print:hidden" style={{ width: 56, height: 58 }} aria-hidden="true">
      <Icon id="o-arch" viewBox="0 0 44 52" size={56} className="absolute inset-0 w-full h-full text-faint" />
      <div className="anim-gate absolute left-1/2 -translate-x-1/2" style={{ top: 6 }}>
        <Icon id="i-portcullis" size={40} className="text-accent" />
      </div>
    </div>
  )
}

/** Quill sweeping across four lines, ink revealing behind it — AI analysis running. */
export function QuillWriter() {
  return (
    <div className="relative print:hidden" style={{ width: 210, height: 84 }} aria-hidden="true">
      {[0, 1, 2, 3].map((n) => (
        <div
          key={n}
          className={`absolute h-[2px] bg-accent/80 anim-ink${n + 1}`}
          style={{ top: n * 22 + 16, left: 0, width: [188, 172, 190, 116][n] }}
        />
      ))}
      <span className="anim-write absolute" style={{ top: -4, left: -4 }}>
        <Icon id="i-quill" size={22} className="text-accent" />
      </span>
    </div>
  )
}
