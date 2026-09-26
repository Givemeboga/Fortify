/**
 * Icon — renders one symbol from the ported <IconSprite/> via <use>.
 * The sprite (and its shared #ink wobble filter) must be mounted once, in App.
 *
 *   <Icon id="i-tower" size={16} className="text-accent" />
 *
 * Named aliases below keep older call-sites working while pointing at the real
 * hand-inked sprite symbols.
 */
export function Icon({ id, size = 20, className = "", viewBox = "0 0 24 24" }) {
  return (
    <svg width={size} height={size} viewBox={viewBox} className={className} aria-hidden="true">
      <use href={`#${id}`} />
    </svg>
  )
}

export const Watchtower = (p) => <Icon id="i-tower" {...p} />
export const Scroll = (p) => <Icon id="i-scroll" {...p} />
export const Banner = (p) => <Icon id="i-banner" {...p} />
export const Eye = (p) => <Icon id="i-lantern" {...p} />        // "The Watch"
export const Swords = (p) => <Icon id="i-swords" {...p} />       // "The Siege"
export const Portcullis = (p) => <Icon id="i-portcullis" {...p} />
export const Crest = (p) => <Icon id="i-shield" {...p} />
