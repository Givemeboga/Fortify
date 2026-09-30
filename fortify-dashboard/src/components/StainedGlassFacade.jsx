/**
 * Stained-glass stat façade for Command (handoff 01). A fixed 1128×400 gothic
 * triptych: three pointed-arch windows with quarry glass, tracery and rose
 * windows, a glowing central "stat pane" per window, and a stone framework.
 * The three stats render as HTML over the SVG (crisp ~7:1 contrast).
 *
 * The quarry panes are a grid of blue lights at random opacities (spec set
 * {.10,.14,.18,.24,.30}); we generate them deterministically so the glass is
 * lively without hand-placing ~200 rects.
 */
const OPAC = [0.1, 0.14, 0.18, 0.24, 0.3]
function panes(x0, y0, cols, rows, cw, ch, seed) {
  let s = ""
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      const o = OPAC[(seed + c * 7 + r * 13 + c * r) % OPAC.length]
      s += `<rect x="${x0 + c * cw}" y="${y0 + r * ch}" width="${cw}" height="${ch}" fill-opacity="${o}"></rect>`
    }
  }
  return s
}

const ARCHES = "M34 380V215A176.3 176.3 0 0 1 189 40A176.3 176.3 0 0 1 344 215V380Z M374 380V215A207.8 207.8 0 0 1 564 8A207.8 207.8 0 0 1 754 215V380Z M784 380V215A176.3 176.3 0 0 1 939 40A176.3 176.3 0 0 1 1094 215V380Z"

const SVG = `<svg width="1128" height="400" viewBox="0 0 1128 400" style="position:absolute;inset:0;overflow:visible">
<defs><clipPath id="kw-c0"><path d="M48 380V215A160.1 160.1 0 0 1 189 56A160.1 160.1 0 0 1 330 215V380Z"></path></clipPath><clipPath id="kw-c1"><path d="M388 380V215A191.6 191.6 0 0 1 564 24A191.6 191.6 0 0 1 740 215V380Z"></path></clipPath><clipPath id="kw-c2"><path d="M798 380V215A160.1 160.1 0 0 1 939 56A160.1 160.1 0 0 1 1080 215V380Z"></path></clipPath><radialGradient id="kw-glow" cx="50%" cy="68%" r="60%"><stop offset="0" stop-color="#2FA4FF" stop-opacity=".28"></stop><stop offset="1" stop-color="#2FA4FF" stop-opacity="0"></stop></radialGradient><radialGradient id="kw-core" cx="50%" cy="45%" r="62%"><stop offset="0" stop-color="#EAF1F8" stop-opacity=".42"></stop><stop offset=".6" stop-color="#EAF1F8" stop-opacity=".08"></stop><stop offset="1" stop-color="#0A0E16" stop-opacity=".15"></stop></radialGradient></defs>
<g fill="#131A28" fill-opacity=".5" stroke="rgba(0,0,0,.55)" stroke-width="2"><path d="M34 380V215A176.3 176.3 0 0 1 189 40A176.3 176.3 0 0 1 344 215V380Z"></path><path d="M374 380V215A207.8 207.8 0 0 1 564 8A207.8 207.8 0 0 1 754 215V380Z"></path><path d="M784 380V215A176.3 176.3 0 0 1 939 40A176.3 176.3 0 0 1 1094 215V380Z"></path></g>
<g clip-path="url(#kw-c0)"><rect x="48" y="56" width="282" height="324" fill="#0A0E16"></rect><g fill="#2FA4FF" stroke="#0A0E16" stroke-width="1.4">${panes(48, 56, 6, 12, 47, 28, 0)}</g><rect x="48" y="56" width="282" height="324" fill="url(#kw-glow)"></rect></g>
<g fill="#2FA4FF" fill-opacity=".42" stroke="#0A0E16" stroke-width="2"><circle cx="189" cy="118" r="37" fill="none"></circle><path d="M172 101A17 17 0 1 1 206 101A17 17 0 1 1 206 135A17 17 0 1 1 172 135A17 17 0 1 1 172 101Z"></path></g>
<g fill="none" stroke-linecap="round" filter="url(#ink)"><g stroke="#0A0E16" stroke-width="7"><path d="M48 380V215A160.1 160.1 0 0 1 189 56A160.1 160.1 0 0 1 330 215V380Z" stroke-width="9"></path><path d="M189 169V380" stroke-width="10"></path><path d="M52 380V221A70 70 0 0 1 120.5 151A70 70 0 0 1 189 221V380"></path><path d="M189 380V221A70 70 0 0 1 257.5 151A70 70 0 0 1 326 221V380"></path></g><g stroke="#131A28" stroke-width="5"><path d="M48 380V215A160.1 160.1 0 0 1 189 56A160.1 160.1 0 0 1 330 215V380Z" stroke-width="7"></path><path d="M189 169V380" stroke-width="8"></path><path d="M52 380V221A70 70 0 0 1 120.5 151A70 70 0 0 1 189 221V380"></path><path d="M189 380V221A70 70 0 0 1 257.5 151A70 70 0 0 1 326 221V380"></path></g></g>
<path d="M91 358V245C91 229 150 222 189 203C228 222 287 229 287 245V358Z" fill="none" stroke="#131A28" stroke-width="9"></path>
<path d="M91 358V245C91 229 150 222 189 203C228 222 287 229 287 245V358Z" fill="#2FA4FF" style="filter:drop-shadow(0 0 14px rgba(47,164,255,.45))"></path>
<path d="M91 358V245C91 229 150 222 189 203C228 222 287 229 287 245V358Z" fill="url(#kw-core)"></path>
<g fill="#0A0E16" fill-opacity=".18"><path d="M91 328V310L109 328Z"></path><path d="M287 328V310L269 328Z"></path></g>
<path d="M91 328H287M91 310L109 328M287 310L269 328" fill="none" stroke="#0A0E16" stroke-width="2.4"></path>
<path d="M91 358V245C91 229 150 222 189 203C228 222 287 229 287 245V358Z" fill="none" stroke="#0A0E16" stroke-width="3.2"></path>
<g clip-path="url(#kw-c1)"><rect x="388" y="24" width="352" height="356" fill="#0A0E16"></rect><g fill="#2FA4FF" stroke="#0A0E16" stroke-width="1.4">${panes(388, 24, 8, 13, 44, 28, 3)}</g><rect x="388" y="24" width="352" height="356" fill="url(#kw-glow)"></rect></g>
<g fill="#2FA4FF" fill-opacity=".42" stroke="#0A0E16" stroke-width="2"><circle cx="478" cy="169" r="21" fill="none"></circle><path d="M469 160A9 9 0 1 1 487 160A9 9 0 1 1 487 178A9 9 0 1 1 469 178A9 9 0 1 1 469 160Z"></path><circle cx="650" cy="169" r="21" fill="none"></circle><path d="M641 160A9 9 0 1 1 659 160A9 9 0 1 1 659 178A9 9 0 1 1 641 178A9 9 0 1 1 641 160Z"></path><circle cx="564" cy="92" r="37" fill="none"></circle><path d="M547 75A17 17 0 1 1 581 75A17 17 0 1 1 581 109A17 17 0 1 1 547 109A17 17 0 1 1 547 75Z"></path></g>
<g fill="none" stroke-linecap="round" filter="url(#ink)"><g stroke="#0A0E16" stroke-width="7"><path d="M388 380V215A191.6 191.6 0 0 1 564 24A191.6 191.6 0 0 1 740 215V380Z" stroke-width="9"></path><path d="M564 151V380" stroke-width="11"></path><path d="M392 380V219A96.6 96.6 0 0 1 478 123A96.6 96.6 0 0 1 564 219V380"></path><path d="M478 197V380" stroke-width="8"></path><path d="M397 380V231A56.3 56.3 0 0 1 437.5 177A56.3 56.3 0 0 1 478 231V380" stroke-width="6"></path><path d="M478 380V231A56.3 56.3 0 0 1 518.5 177A56.3 56.3 0 0 1 559 231V380" stroke-width="6"></path><path d="M564 380V219A96.6 96.6 0 0 1 650 123A96.6 96.6 0 0 1 736 219V380"></path><path d="M650 197V380" stroke-width="8"></path><path d="M569 380V231A56.3 56.3 0 0 1 609.5 177A56.3 56.3 0 0 1 650 231V380" stroke-width="6"></path><path d="M650 380V231A56.3 56.3 0 0 1 690.5 177A56.3 56.3 0 0 1 731 231V380" stroke-width="6"></path></g><g stroke="#131A28" stroke-width="5"><path d="M388 380V215A191.6 191.6 0 0 1 564 24A191.6 191.6 0 0 1 740 215V380Z" stroke-width="7"></path><path d="M564 151V380" stroke-width="9"></path><path d="M392 380V219A96.6 96.6 0 0 1 478 123A96.6 96.6 0 0 1 564 219V380"></path><path d="M478 197V380" stroke-width="6"></path><path d="M397 380V231A56.3 56.3 0 0 1 437.5 177A56.3 56.3 0 0 1 478 231V380" stroke-width="4"></path><path d="M478 380V231A56.3 56.3 0 0 1 518.5 177A56.3 56.3 0 0 1 559 231V380" stroke-width="4"></path><path d="M564 380V219A96.6 96.6 0 0 1 650 123A96.6 96.6 0 0 1 736 219V380"></path><path d="M650 197V380" stroke-width="6"></path><path d="M569 380V231A56.3 56.3 0 0 1 609.5 177A56.3 56.3 0 0 1 650 231V380" stroke-width="4"></path><path d="M650 380V231A56.3 56.3 0 0 1 690.5 177A56.3 56.3 0 0 1 731 231V380" stroke-width="4"></path></g></g>
<path d="M446 358V245C446 229 517 222 564 203C611 222 682 229 682 245V358Z" fill="none" stroke="#131A28" stroke-width="9"></path>
<path d="M446 358V245C446 229 517 222 564 203C611 222 682 229 682 245V358Z" fill="#2FA4FF" style="filter:drop-shadow(0 0 14px rgba(47,164,255,.45))"></path>
<path d="M446 358V245C446 229 517 222 564 203C611 222 682 229 682 245V358Z" fill="url(#kw-core)"></path>
<g fill="#0A0E16" fill-opacity=".18"><path d="M446 328V310L464 328Z"></path><path d="M682 328V310L664 328Z"></path></g>
<path d="M446 328H682M446 310L464 328M682 310L664 328" fill="none" stroke="#0A0E16" stroke-width="2.4"></path>
<path d="M446 358V245C446 229 517 222 564 203C611 222 682 229 682 245V358Z" fill="none" stroke="#0A0E16" stroke-width="3.2"></path>
<g clip-path="url(#kw-c2)"><rect x="798" y="56" width="282" height="324" fill="#0A0E16"></rect><g fill="#2FA4FF" stroke="#0A0E16" stroke-width="1.4">${panes(798, 56, 6, 12, 47, 28, 1)}</g><rect x="798" y="56" width="282" height="324" fill="url(#kw-glow)"></rect></g>
<g fill="#2FA4FF" fill-opacity=".42" stroke="#0A0E16" stroke-width="2"><circle cx="939" cy="118" r="37" fill="none"></circle><path d="M922 101A17 17 0 1 1 956 101A17 17 0 1 1 956 135A17 17 0 1 1 922 135A17 17 0 1 1 922 101Z"></path></g>
<g fill="none" stroke-linecap="round" filter="url(#ink)"><g stroke="#0A0E16" stroke-width="7"><path d="M798 380V215A160.1 160.1 0 0 1 939 56A160.1 160.1 0 0 1 1080 215V380Z" stroke-width="9"></path><path d="M939 169V380" stroke-width="10"></path><path d="M802 380V221A70 70 0 0 1 870.5 151A70 70 0 0 1 939 221V380"></path><path d="M939 380V221A70 70 0 0 1 1007.5 151A70 70 0 0 1 1076 221V380"></path></g><g stroke="#131A28" stroke-width="5"><path d="M798 380V215A160.1 160.1 0 0 1 939 56A160.1 160.1 0 0 1 1080 215V380Z" stroke-width="7"></path><path d="M939 169V380" stroke-width="8"></path><path d="M802 380V221A70 70 0 0 1 870.5 151A70 70 0 0 1 939 221V380"></path><path d="M939 380V221A70 70 0 0 1 1007.5 151A70 70 0 0 1 1076 221V380"></path></g></g>
<path d="M841 358V245C841 229 900 222 939 203C978 222 1037 229 1037 245V358Z" fill="none" stroke="#131A28" stroke-width="9"></path>
<path d="M841 358V245C841 229 900 222 939 203C978 222 1037 229 1037 245V358Z" fill="#2FA4FF" style="filter:drop-shadow(0 0 14px rgba(47,164,255,.45))"></path>
<path d="M841 358V245C841 229 900 222 939 203C978 222 1037 229 1037 245V358Z" fill="url(#kw-core)"></path>
<g fill="#0A0E16" fill-opacity=".18"><path d="M841 328V310L859 328Z"></path><path d="M1037 328V310L1019 328Z"></path></g>
<path d="M841 328H1037M841 310L859 328M1037 310L1019 328" fill="none" stroke="#0A0E16" stroke-width="2.4"></path>
<path d="M841 358V245C841 229 900 222 939 203C978 222 1037 229 1037 245V358Z" fill="none" stroke="#0A0E16" stroke-width="3.2"></path>
<g fill="#131A28" fill-opacity=".75" stroke="rgba(0,0,0,.5)" stroke-width="1.2"><rect x="350" y="215" width="18" height="165"></rect><path d="M341 215H377L372 203H346Z"></path><rect x="341" y="199" width="36" height="5"></rect><rect x="344" y="380" width="30" height="20"></rect><rect x="760" y="215" width="18" height="165"></rect><path d="M751 215H787L782 203H756Z"></path><rect x="751" y="199" width="36" height="5"></rect><rect x="754" y="380" width="30" height="20"></rect><rect x="14" y="26" width="14" height="354"></rect><path d="M5 26H37L32 14H10Z"></path><rect x="5" y="10" width="32" height="5"></rect><rect x="8" y="380" width="26" height="20"></rect><rect x="1100" y="26" width="14" height="354"></rect><path d="M1091 26H1123L1118 14H1096Z"></path><rect x="1091" y="10" width="32" height="5"></rect><rect x="1094" y="380" width="26" height="20"></rect><rect x="0" y="0" width="1128" height="10"></rect><rect x="4" y="10" width="1120" height="5"></rect><rect x="34" y="380" width="1060" height="8"></rect></g>
</svg>`

// The three stat panes' HTML label geometry (left, width) — number strip and
// label strip stacked, both in near-black for contrast against the blue glass.
const PANES = [
  { left: 91, width: 196 },
  { left: 446, width: 236 },
  { left: 841, width: 196 },
]

function StatText({ left, width, n, label }) {
  return (
    <>
      <div
        style={{
          position: "absolute", left, width, top: 238, height: 90,
          display: "flex", alignItems: "center", justifyContent: "center",
          fontFamily: "'Grenze Gotisch', serif", fontWeight: 700, fontSize: 64, lineHeight: 1, color: "#0A0E16",
        }}
      >{n}</div>
      <div
        style={{
          position: "absolute", left, width, top: 328, height: 30,
          display: "flex", alignItems: "center", justifyContent: "center",
          fontFamily: "'JetBrains Mono', monospace", fontWeight: 700, fontSize: 11,
          letterSpacing: ".18em", whiteSpace: "nowrap", color: "#0A0E16",
        }}
      >{label}</div>
    </>
  )
}

export default function StainedGlassFacade({ stats }) {
  return (
    <div style={{ position: "relative", width: 1128, height: 400, flex: "none", alignSelf: "center", marginTop: -12 }}>
      <div
        style={{
          position: "absolute", inset: 0, backgroundImage: "var(--tex-stone)",
          opacity: 0.35, clipPath: `path('${ARCHES}')`,
        }}
      />
      <div dangerouslySetInnerHTML={{ __html: SVG }} />
      {stats.map((s, i) => (
        <StatText key={i} left={PANES[i].left} width={PANES[i].width} n={s.n} label={s.label} />
      ))}
    </div>
  )
}
