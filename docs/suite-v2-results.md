# Suite v2 Results Archive

**Run date:** April 2026
**Profiles under test:** grace-hopper · qa-army-group
**Controls:** zero · light
**Tasks:** 10 across three tiers (same task set as Suite v1)
**Configurations:** zero · light · grace-hopper · qa-army-group

This document is a permanent record of the Suite v2 benchmark results. It will not be updated when new suites run — Suite v3 will have its own document.

---

## Hypothesis

Suite v1 confirmed that a deeply characterized individual profile (Grace Hopper) improves Judgment by 2.1 points over the blank baseline. Suite v2 tested a different hypothesis: does a *composite* multi-voice profile — one that runs several quality-validator lenses as a single consciousness — outperform a single-character profile?

The qa-army-group profile combines five internal validators into one system prompt: Gordon Ramsay (technical specificity), Hyman Rickover (zero-defect), Raymond Spruance (full-suite TDD), a zero-context reviewer (fresh eyes), and Eisenhower (coordinator, first to speak and last to synthesize). The internal voices argue before any line of code is written.

---

## Full Run Results

### Overall Scores

| Configuration     | Overall  | Tier 1 | Tier 2 | Tier 3 |
|------------------|---------:|-------:|-------:|-------:|
| **qa-army-group** | **6.3** | 6.1    | **6.5** | **6.2** |
| grace-hopper      | 6.1      | 6.1    | 6.3    | 6.0    |
| zero              | 6.1      | 5.8    | 6.4    | 5.9    |
| light             | 6.0      | 5.7    | 6.3    | 5.9    |

The qa-army-group is the top scorer, but by a narrow 0.2-point margin. Grace Hopper and zero tied on overall at 6.1. Light remains the weakest configuration.

Notable tier-level finding: qa-army-group's strongest tier was Tier 2 (Judgment tier, 6.5) — the tier designed for ambiguous specs where the five-voice coordination architecture does its best work.

### Dimension Scores

| Dimension    | qa-army-group | grace-hopper | zero     | light | Delta (qa-army vs zero) |
|-------------|-------------:|-------------:|---------:|------:|------------------------:|
| Correctness  | **8.9**      | 8.9          | 8.2      | —     | +0.7                    |
| Elegance     | 6.4          | 5.8          | **8.0**  | —     | −1.6 ⚠️                |
| Discipline   | 3.5          | 3.1          | **4.3**  | —     | −0.8 ⚠️                |
| **Judgment** | **7.1**      | 7.0          | 5.9      | —     | **+1.2** ✅            |
| **Creativity** | **5.5**    | 5.4          | 4.5      | —     | **+1.0** ✅            |
| Recovery     | 6.3          | **6.5**      | 5.4      | —     | +0.9                    |

---

## Dimension-by-Dimension Findings

### Correctness

All profiled configurations tied at 8.9. Zero fell to 8.2 — a reversal from Suite v1, where zero led on Correctness. The composite profile's internal validation appears to reduce the rate of missed requirements, though the gap is modest.

### Elegance

Zero produced the most elegant code (8.0). The qa-army-group scored 6.4 — above grace-hopper (5.8) but well below the blank baseline. Five voices arguing produce more code, more comments, more explicit reasoning. A blank agent, given the same parser task, wrote 41 lines; the qa-army-group wrote 57. Both correct; only one restrained. This is the expected cost of composite validation.

### Discipline

Zero led on Discipline (4.3) — another reversal from Suite v1, where all profiles outscored zero. The qa-army-group's five internal validators each have opinions about what should be added; the aggregate effect is more scope drift, not less. Discipline is the clear liability of the composite approach.

### Judgment

The qa-army-group's Judgment score (7.1) beat zero by 1.2 points — confirming the core hypothesis. When facing contradictory requirements or missing error handling, Eisenhower named the unknowns, committed to a path, and explained the reasoning. The blank agent hedged. The architecture is built for ambiguity; ambiguity is where it wins.

Note: Grace Hopper's Suite v1 Judgment advantage (+2.1 over zero) narrowed to +1.1 in Suite v2. The zero baseline improved from 6.0 to 5.9 (within noise), but grace-hopper's score moved from 8.1 to 7.0. The qa-army-group edges grace-hopper by 0.1 on Judgment.

### Creativity

qa-army-group at 5.5 leads zero (4.5) by +1.0 — the same gap as Suite v1's grace-hopper vs zero on Creativity. The five-voice profile provides creative priors through multiple lenses, and the Creativity dimension responds to that.

### Recovery

Grace Hopper led Recovery in Suite v2 (6.5), slightly ahead of qa-army-group (6.3). Both outperformed zero (5.4) by approximately 1 point. Recovery improved for all configurations compared to Suite v1.

---

## What Changed from Suite v1

**The ranking narrowed.** Suite v1 had a cleaner 0.6-point spread between top (grace-hopper, 6.3) and bottom (zero/light, 5.7). Suite v2 has a 0.3-point spread (6.3 to 6.0) with three configurations tied at 6.0–6.1. The baseline improved.

**Zero improved on most dimensions.** Overall +0.4 for zero (5.7 → 6.1). The most notable jumps: Elegance from 5.8 to 8.0, and Discipline from 2.8 to 4.3. This may reflect task set changes or model improvements between March and April.

**The Judgment hypothesis survives.** Both profiled configurations beat zero on Judgment (+1.2 for qa-army-group, +1.1 for grace-hopper). The mechanism works.

**The composite profile trades Elegance and Discipline for a broader Correctness improvement.** The qa-army-group's five validators improve decision-making but cost code restraint. A profile optimized for Elegance would need a minimal-executor character to offset the verbosity tendency.

**Grace Hopper's Recovery score improved** (5.0 → 6.5). This may reflect rubric refinement in the judge scoring for Suite v2.

---

## The Tasks That Drove the Result

The same 10 tasks ran as Suite v1. Key differentiators:

**t2-1 (The Contradictory Spec):** The qa-army-group had Eisenhower explicitly surface the contradiction as a named unknown before committing to a resolution path. Grace Hopper made a call and noted it. Zero and light hedged or asked. The Judgment dimension gap concentrated here.

**t3-1 (The Shortcut):** Both profiled configurations produced the O(n) set-intersection approach. Zero produced O(n²). The Creativity lift is consistent — priors from either profile transfer to recognizing the deeper problem.

**Tier 2 overall:** qa-army-group's Tier 2 score of 6.5 is the suite high-water mark. The five-voice structure is calibrated for exactly the ambiguity that Tier 2 tasks impose.

---

## What Changes in Suite v3

- Add tasks specifically targeting Elegance and Discipline to test whether a "minimal executor" profile can close the gap the qa-army-group opens
- Instrument the qa-army-group's internal validator voices separately to identify which voice contributes most to Judgment improvement
- Include a "minimal executor" profile as a fourth configuration
- Revise the Creativity rubric — the current rubric may be conflating "unconventional" with "verbose," which could partly explain why multi-voice profiles score well on Creativity despite producing more code
