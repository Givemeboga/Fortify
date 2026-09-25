/**
 * Hand-inked-style icon set for the Keep. Simple stroke paths that inherit
 * `currentColor` and stay crisp at 20–24px. Each takes { size = 20, className }.
 *
 * Swap-friendly: import only what you use.
 */

const common = {
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.6,
  strokeLinecap: "round",
  strokeLinejoin: "round",
  "aria-hidden": true,
}

// Watchtower — Command / monitoring
export function Watchtower({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M8 21V9h8v12" />
      <path d="M8 9V6h1.8v1.6h1.6V6h1.8v1.6h1.6V6h1.8v3" />
      <path d="M11 21v-4h2v4" />
    </svg>
  )
}

// Scroll — Counsel / Settings / report
export function Scroll({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <rect x="6" y="4" width="12" height="16" rx="1.5" />
      <path d="M9 9h6M9 12h6M9 15h4" />
    </svg>
  )
}

// Eye — passive scan ("The Watch")
export function Eye({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" />
      <circle cx="12" cy="12" r="2.6" />
    </svg>
  )
}

// Crossed swords — active scan ("The Siege")
export function Swords({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M6.5 17.5 17.5 6.5" />
      <path d="M17.5 17.5 6.5 6.5" />
      <path d="M15.5 4.5 19.5 8.5M4.5 15.5 8.5 19.5" />
    </svg>
  )
}

// Portcullis — consent gate
export function Portcullis({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M4 5h16M8 5v15M12 5v15M16 5v15" />
      <path d="M4 10h16M4 15h16" />
    </svg>
  )
}

// Banner — dispatch / launch scan
export function Banner({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M6 3v18" />
      <path d="M6 4h11l-2.5 3.5L17 11H6" />
    </svg>
  )
}

// Crest — severity shield (colour via currentColor / text-crit etc.)
export function Crest({ size = 20, className = "" }) {
  return (
    <svg width={size} height={size} className={className} {...common}>
      <path d="M12 3l7 2.5v5.5c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V5.5L12 3Z" />
      <path d="M12 8.5v3.5M12 15h.01" />
    </svg>
  )
}
