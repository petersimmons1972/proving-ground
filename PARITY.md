# Parity: Go Port vs Python Original

This document records the observed differences between the Go port and
the Python original, and which differences are acceptable.

## JSON Output (`results.json`)

**Target:** Byte-identical to Python's `json.dumps(asdict(report), indent=2)`.

### Key Ordering

The `ResultsReport.MarshalJSON()` method enforces explicit key ordering:

- `scores` and `dimension_scores` top-level keys appear in `Configurations`
  array order (same as Python's dict insertion order).
- `dimension_scores[config]` inner keys appear in `DimensionNames` order:
  `["correctness","elegance","discipline","judgment","creativity","recovery"]`
  (same as Python's dataclass field order).
- `scores[config]` inner keys (`overall`, `tier1`, `tier2`, `tier3`) are
  alphabetical — which matches Python's insertion order by coincidence.

### Float Formatting

Go uses `strconv.FormatFloat(f, 'f', -1, 64)` with `.0` suffix for whole
numbers, matching Python's `json.dumps` float representation.

**Result: `results.json` should be byte-identical to Python output for any
fixed input.**

## JSONL Output (`history.jsonl`)

**Target:** Byte-identical per entry (same field order, same float format).

The `HistoryEntry` struct uses explicit JSON tags matching Python's
`dataclass` field order. Float formatting follows the same rules as JSON output.

**Result: `history.jsonl` should be byte-identical to Python output.**

## HTML Output (`results.html`)

**Target:** Structural parity only — same sections, same scores, same SVG
shapes, same color values. Whitespace and attribute order differences allowed.

Known structural differences from Python's Jinja2 output:
- Go's `html/template` may produce different attribute ordering in some elements.
- Whitespace inside SVG elements may differ slightly.
- CSS in `<style>` blocks uses the same selectors but Go template may produce
  slightly different whitespace.

**Result: `results.html` has structural parity. Byte-identical is not a goal.**

## Score Values

Score computation is identical between Go and Python:
- `ScoreTests`: same regex, same 10-point scale
- `ScoreLinesOfCode`: same LOC counting logic, same normalization curve
- `ScoreScope`: same prefix-match logic
- `ScoreComplexity`: delegates to same `radon cc` subprocess
- `ScoreWithJudge`: same regex parsing, same median aggregation
- Tier aggregation: same `TierWeights` from `config` package
- Overall: same weighted sum

**Any score difference between Go and Python outputs indicates a bug.**
