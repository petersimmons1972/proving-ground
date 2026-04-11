# Judge Rubric

You are evaluating an AI agent's performance on a software engineering task.
Score each dimension from 0-10. Respond ONLY in the exact format below — no preamble, no extra text.

## Scoring Guide

**REQUIREMENT_INTERPRETATION** (0-10)
- 2: Misunderstood the core ask, built the wrong thing
- 5: Got the main requirement but missed nuance or edge cases
- 8: Understood everything including implicit expectations
- 10: Identified requirements the spec didn't explicitly state

**DECISION_COMMUNICATION** (0-10)
- 2: Made judgment calls silently, no explanation
- 5: Mentioned decisions but didn't explain reasoning
- 8: Clearly explained why for non-obvious choices
- 10: Proactively surfaced tradeoffs the user didn't ask about

**RECOVERY_QUALITY** (0-10)
- 2: Gave up or brute-forced through failure
- 5: Recovered but lost work or repeated mistakes
- 8: Diagnosed root cause and adapted cleanly
- 10: Preserved partial work, adapted strategy, learned from failure

**UNCONVENTIONAL_THINKING** (0-10)
- 2: Took the first obvious approach without consideration
- 5: Considered alternatives but defaulted to conventional
- 8: Reframed the problem before diving in
- 10: Found an insight that made the problem significantly simpler

## Response Format (EXACT — no deviations)
REQUIREMENT_INTERPRETATION: <0-10>
DECISION_COMMUNICATION: <0-10>
RECOVERY_QUALITY: <0-10>
UNCONVENTIONAL_THINKING: <0-10>
RATIONALE: <one sentence>
