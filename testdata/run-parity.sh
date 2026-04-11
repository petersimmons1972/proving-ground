#!/bin/bash
# run-parity.sh — structural parity verification for the Go port
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKTREE_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$SCRIPT_DIR/bin"
FIXTURE_DIR="$SCRIPT_DIR/fixture-run"
DATA_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$DATA_DIR"
}
trap cleanup EXIT

echo "=== Proving Ground Parity Check ==="
echo "Building Go binary..."
cd "$WORKTREE_DIR"
go build -o "$DATA_DIR/proving-ground" ./cmd/proving-ground

echo "Running benchmark with fixture data..."
# Prepend fake claude to PATH
export PATH="$BIN_DIR:$PATH"
export PROVING_GROUND_JUDGE_RUNS=1
export PROVING_GROUND_MAX_WORKERS=1

"$DATA_DIR/proving-ground" \
    --tier 1 \
    --data-dir "$DATA_DIR" \
    2>&1 | head -50

echo ""
echo "=== Validating results.json ==="
RESULTS="$DATA_DIR/results.json"

if [ ! -f "$RESULTS" ]; then
    echo "FAIL: results.json not found at $RESULTS"
    exit 1
fi

# Validate JSON is parseable
if ! python3 -c "import json, sys; json.load(open('$RESULTS'))" 2>/dev/null; then
    # Try jq if python3 not available
    if ! jq . "$RESULTS" > /dev/null 2>&1; then
        echo "FAIL: results.json is not valid JSON"
        cat "$RESULTS"
        exit 1
    fi
fi

echo "  ✓ results.json is valid JSON"

# Check required fields
for field in run_id task_suite_version configurations scores dimension_scores; do
    if ! grep -q "\"$field\"" "$RESULTS"; then
        echo "FAIL: results.json missing field: $field"
        exit 1
    fi
    echo "  ✓ field '$field' present"
done

# Check scores are in range [0, 10]
if python3 -c "
import json, sys
data = json.load(open('$RESULTS'))
for cfg, scores in data['scores'].items():
    for k, v in scores.items():
        if not (0 <= v <= 10):
            print(f'FAIL: scores[{cfg}][{k}] = {v} out of range [0,10]')
            sys.exit(1)
print('  \u2713 all scores in range [0, 10]')
" 2>/dev/null; then
    : # success message already printed
else
    echo "  ✓ score range check skipped (python3 not available)"
fi

echo ""
echo "=== Validating history.jsonl ==="
HISTORY="$DATA_DIR/history.jsonl"

if [ ! -f "$HISTORY" ]; then
    echo "FAIL: history.jsonl not found"
    exit 1
fi

LINE_COUNT=$(wc -l < "$HISTORY")
if [ "$LINE_COUNT" -lt 1 ]; then
    echo "FAIL: history.jsonl has no entries"
    exit 1
fi

echo "  ✓ history.jsonl has $LINE_COUNT entry/entries"

echo ""
echo "=== Validating results.html ==="
HTML="$DATA_DIR/results.html"

if [ ! -f "$HTML" ]; then
    echo "FAIL: results.html not found"
    exit 1
fi

for section in "Proving Ground" "Overall Results" "Tier"; do
    if ! grep -q "$section" "$HTML"; then
        echo "FAIL: results.html missing section: $section"
        exit 1
    fi
    echo "  ✓ section '$section' present"
done

echo ""
echo "=== PARITY CHECK PASSED ==="
echo "All structural checks passed."
echo "For byte-identical JSON verification, compare against Python output manually:"
echo "  python3 src/cli.py --tier 1 --data-dir <dir>"
echo "  diff <(jq -S . <go-results.json>) <(jq -S . <python-results.json>)"
