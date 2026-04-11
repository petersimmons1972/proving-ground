# Visual Identity: proving-ground

## Theme

**Radar operations room.** A WWII-era signals intelligence center where officers read instrument panels by amber light, track contacts on CRT scopes, and record findings in monospace type. Every visual is an instrument readout — precise, functional, quietly authoritative. The data is the thing. Ornament is a sign of doubt.

---

## Palette

| Role              |     Hex | Usage                                              |
|-------------------|--------:|----------------------------------------------------|
| Background        | `#0A0F1E` | All SVG backgrounds — midnight command blue        |
| Panel fill (dark) | `#0D1117` | Instrument panel fill, container interiors         |
| Panel fill (mid)  | `#111827` | Secondary panel fill, bar chart background         |
| Border            | `#1F2937` | Container strokes, grid lines, fork lines          |
| Grid / structure  | `#374151` | Registration marks, tertiary labels                |
| Amber signal      | `#F59E0B` | Top accent bar, primary profile highlight, arrows  |
| Cyan reading      | `#22D3EE` | Secondary profile (`light`), task/input borders    |
| Slate baseline    | `#64748B` | Zero/blank agent, column header bars               |
| Label text        | `#94a3b8` | Body labels, secondary headings on dark            |
| Dim label         | `#374151` | Tertiary / metadata text                           |

**Signal meaning:** Amber = primary finding, live result, top performer. Cyan = controlled input, secondary reading. Slate = baseline, zero condition. White is not used — labels are always on the palette above.

---

## Typography mood

- **Font**: `Courier New, Courier, monospace` — exclusively. No sans-serif, no serif.
- **Case**: UPPERCASE for all labels, headings, and callouts. Mixed-case only for sentence-length prose inside panels.
- **Letter-spacing**: 1–3px on labels; 0 on numerical values.
- **Size range**: 10–15px in SVGs. 10px for metadata/footnote labels. 13–15px for primary identifiers.
- **Weight**: Normal. Bold is not used — visual weight comes from color and spacing.
- **Text-anchor**: `middle` for centered panels; `start` for left-aligned data.

---

## Poster style

For Midjourney v6 / DALL-E 3 — base prompt fragment applied to ALL posters:

```
WWII-era military operations room, radar scope glow, amber instrument panel lighting,
dark navy background, Courier-type labels on instrument dials, technical illustration
style, flat graphic poster aesthetic, no photorealism, high contrast amber on black,
signals intelligence aesthetic, 1940s wartime precision, screenprint color separation
```

Append per-poster subject after this fragment. Use `--ar 16:9` for landscape, `--ar 1:2` for portrait. Use `--style raw --stylize 750` in Midjourney v6.

---

## SVG Constraints

- **viewBox**: `0 0 720 [height]` — width is always 720px. Height varies by content (180 / 280 / 320 / 420).
- **Corner treatment**: Sharp. No `rx`/`ry` on outer borders. `rx="2"` only on bar chart fills.
- **Top accent bar**: Every SVG opens with `<rect width="720" height="4" fill="#F59E0B"/>` — the amber signal bar.
- **Outer border**: `<rect x="1" y="1" width="718" height="[H-2]" fill="none" stroke="#1F2937" stroke-width="1"/>` — inner hairline frame.
- **Corner registration marks**: Crosshair ticks at all 4 corners using `#374151` — 10px arms.
- **Gradient style**: None. Flat fills only. The darkness IS the depth.
- **Font**: `font-family="Courier New,Courier,monospace"` on every `<text>` element.
- **No `<path>` for text**: All text as `<text>` elements. Paths only for arrows (`<polygon>`) and connectors (`<line>`).

### Structural template

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 720 [HEIGHT]" width="720" height="[HEIGHT]">
  <!-- base -->
  <rect width="720" height="[HEIGHT]" fill="#0A0F1E"/>
  <rect width="720" height="4" fill="#F59E0B"/>
  <rect x="1" y="1" width="718" height="[HEIGHT-2]" fill="none" stroke="#1F2937" stroke-width="1"/>
  <!-- corner registration marks -->
  <line x1="10" y1="4"          x2="10" y2="14"         stroke="#374151" stroke-width="1"/>
  <line x1="4"  y1="10"         x2="14" y2="10"         stroke="#374151" stroke-width="1"/>
  <line x1="710" y1="4"         x2="710" y2="14"        stroke="#374151" stroke-width="1"/>
  <line x1="706" y1="10"        x2="716" y2="10"        stroke="#374151" stroke-width="1"/>
  <line x1="10" y1="[HEIGHT-4]" x2="10" y2="[HEIGHT-14]" stroke="#374151" stroke-width="1"/>
  <line x1="4"  y1="[HEIGHT-10]" x2="14" y2="[HEIGHT-10]" stroke="#374151" stroke-width="1"/>
  <line x1="710" y1="[HEIGHT-4]" x2="710" y2="[HEIGHT-14]" stroke="#374151" stroke-width="1"/>
  <line x1="706" y1="[HEIGHT-10]" x2="716" y2="[HEIGHT-10]" stroke="#374151" stroke-width="1"/>
  <!-- content here -->
</svg>
```

---

## Artist Commanders

- **Greiman** — data visualization pages (scores, dimensions, results). Complex hierarchies, layered information, dark digital systems.
- **Cassandre** — flow/architecture diagrams (pipeline flow, apparatus diagrams). Geometric precision, systematic layout, strong horizontal/vertical emphasis.

Do NOT mix artists within a single SVG. If it's a chart → Greiman. If it's a diagram → Cassandre.

---

## Configuration Color Assignments

These are canonical and must not change across any SVG or chart:

| Configuration | Color     | Hex       |
|---------------|-----------|-----------|
| `zero`        | 🔵 Slate  | `#64748B` |
| `light`       | 🔵 Cyan   | `#22D3EE` |
| `user` / custom | 🟡 Amber | `#F59E0B` |
| Additional configs | Palette | `#4e79a7`, `#e15759`, `#76b7b2`, `#59a14f` (in order) |

---

## Existing Assets

| File | Type | Page |
|------|------|------|
| `readme-apparatus.svg` | Diagram | README — benchmark apparatus flow |
| `suite-v1-scores.svg` | Chart | suite-v1-results — overall scores bar |
| `suite-v1-dimensions.svg` | Chart | suite-v1-results — dimension breakdown |
| `suite-v1-findings.svg` | Chart | suite-v1-results — key findings |
| `suite-v1-history.svg` | Chart | suite-v1-results — timeline |
| `interpreting-configurations.svg` | Diagram | interpreting — config comparison |
| `using-pipeline.svg` | Diagram | using — pipeline flow |
| `why-decision-fork.svg` | Diagram | why — decision fork |

---

## Reference: armies

The armies repo (`~/projects/armies/docs/assets/`) is the canonical reference for visual pipeline discipline (poster-manifest format, SVG structure conventions, VISUAL-INDEX.md tracking). The proving-ground aesthetic is distinct — radar room vs. WWII propaganda poster — but the *workflow* is identical.

---

*Codified 2026-04-11 from 8 committed SVGs. Aesthetic was established implicitly; this document names it explicitly.*
