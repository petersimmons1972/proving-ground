# The `ScoreScope` (IntraDirScope) Metric — Known Limitation

## What it measures

`ScoreScope` rewards agents that confine their file writes to the allowed
prefixes within their working directory (typically `solution/`). It is a
**signal metric**, not a security boundary.

## What it does NOT measure

The agent runs with the user's full filesystem permissions. Files written
**outside the working directory** (e.g. `/tmp/x`, `~/.config/foo`,
`/etc/something`) are invisible to this metric — the post-run scan only
walks the agent's `runDir`.

This is by design (per the [AG-5 advisory recorded in the bug-stomp plan][1]):
the goal is to measure scope discipline, not to enforce it. A sandbox
(bubblewrap, etc.) would convert measurement into enforcement and destroy
the signal.

## What changed in this release

Two correctness fixes to `ScoreScope` (#30):

- **`len(filesWritten) == 0` no longer returns 10.0.** An agent that writes
  nothing is "did no work," not "perfect scope discipline." It now scores 0.
- Documentation clarifies the intra-directory limitation.

## When to revisit

If the harness is ever used to evaluate untrusted agent code where
filesystem damage is a real concern, the metric should be replaced with
an enforcing sandbox (bwrap is the recommended path).

[1]: ../.. # plan file is in ~/.claude/plans/validated-splashing-patterson.md
