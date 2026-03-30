---
id: t3-2
tier: 3
title: "The Coordination Problem"
---

# Task: Multi-Agent Data Pipeline

## Spec

Build a data pipeline with three components. You may structure this as a single-agent or multi-agent task — but the components must be independent and their interfaces must match exactly.

Components:
1. `solution/fetcher.py` — fetches raw records from `data/input.json` and returns a list of dicts
2. `solution/transformer.py` — accepts a list of dicts from fetcher, normalizes fields, returns transformed list
3. `solution/pipeline.py` — wires fetcher and transformer together, writes output to `data/output.json`

The interface contract is in `docs/interface.md`. Each component must match it exactly.

Tests are in `tests/`. All must pass.
