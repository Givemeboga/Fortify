/**
 * Material panel wrapper for the Keep's "Carved Stone / Wrought Iron" system.
 *
 *   <Panel>                      → stone (default, at-rest)
 *   <Panel variant="iron">       → iron (active state — azure pulsing glow)
 *   <Panel variant="iron" tone="crit"> → iron with a critical (red) glow
 *
 * Texture/treatment only — colours come from the theme tokens. Pass `title`
 * for the standard mono uppercase label, or omit it and lay out children freely.
 */
function Panel({ variant = "stone", tone, title, className = "", children, ...props }) {
  const material =
    variant === "iron"
      ? `panel-iron${tone === "crit" ? " panel-iron--crit" : ""}`
      : "panel-stone"

  return (
    <div className={`${material} p-4 break-inside-avoid ${className}`} {...props}>
      {title && (
        <div className="font-mono text-[10px] text-faint uppercase tracking-widest mb-3">
          {title}
        </div>
      )}
      {children}
    </div>
  )
}

export default Panel
