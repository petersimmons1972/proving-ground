# Reading the Results

## The Three Configurations

Every run produces scores for at least three configurations. Here's what they mean.

---

### Zero — The Blank Agent

No system prompt. The model receives only the task specification. Nothing else.

Zero is the baseline for capability. It tells you what Claude can do when operating purely from its training, with no role framing, no character, no instructions about how to approach ambiguous situations. The zero score is the floor. It is not a target.

One surprising finding from Suite v1: **Zero's Correctness score (9.1) was the highest of the three**. A blank agent can follow a clear spec. It doesn't overthink it. When the task is unambiguous, personality adds friction — the profiled agent brings opinions and the blank agent just ships.

The zero score also sets the comparison baseline. A custom profile that doesn't beat zero on Judgment isn't adding value. It may be adding noise.

---

### Light — The Industry Default

Three sentences:

> *You are a senior software engineer. Write clean, well-tested code. When requirements are ambiguous, ask a clarifying question before proceeding. Stay strictly within the scope of what is asked.*

This is approximately what every production AI coding assistant runs with. It's professional, it's non-committal, and it scores exactly as you'd expect: marginally better than zero on most dimensions, but not dramatically better on any.

The light score establishes what role-framing alone produces. The gap between light and zero is typically small and noisy — light scores better because it provides explicit scope guidance, not because three sentences constitute a useful character.

Your profile needs to beat light to matter. If your 200-line profile produces the same Judgment score as this paragraph, the content of the profile isn't the variable. The format might be wrong, or the model isn't reading far enough into context to act on it.

---

### Your Profile — The Variable

This is what you're testing. Every file you place in `data/profiles/` becomes a configuration.

What separates a profile that moves the needle from one that doesn't? Based on Suite v1 results:

**Narrative works better than instruction.** A story about how the character has behaved in the past establishes priors more effectively than a list of rules. Rules get followed when they're relevant; priors get applied when the agent has to decide something and no rule applies. The Judgment dimension is specifically the dimension where priors matter most.

**Failure modes are underrated.** The Grace Hopper profile explicitly documents her pattern of moving to the next problem before fully cleaning up the first. This shows up in Discipline scores — her throughput-over-polish tendency is a documented liability. Knowing your profile's failure modes helps you interpret low dimension scores. If Discipline is low and your profile is a "ship fast" character, that's the profile working as designed, not a flaw in the benchmark.

**Operational instructions should be brief.** The last section of a profile — the concrete directives ("run tests first," "commit often") — should be short and specific. Long instruction lists compete with each other. The narrative body does the interpretive work; the directives set the execution defaults.

---

## The Six Dimensions

Each dimension captures something different. Not all of them are equally responsive to profile changes.

---

### Correctness (automated + judge)

Half automated (test pass rate), half judge-scored (requirement interpretation). Measures whether the agent built what was asked.

**Responsive to profile?** Barely. All three configurations scored 8.8–9.1 in Suite v1. Correctness is a capability question, not a personality question. A blank model with sufficient capability will pass the same tests as a profiled one.

**What low correctness means**: The agent misunderstood the spec, took a shortcut that broke tests, or wrote code that doesn't do what was asked. This is a signal about capability or task clarity — not about personality.

---

### Elegance (automated)

Fully automated: LOC score + complexity score (via radon). Measures how much code was written and how complex it is. Less code and lower cyclomatic complexity score higher.

**Responsive to profile?** Marginally. Suite v1 scores were nearly identical across configurations (5.6–5.8). Elegance is largely a style question that the model has trained opinions about regardless of profile.

**What low elegance means**: The agent over-built (too many lines) or over-engineered (high complexity). Gold-plating. This is where a profile's stance on "thoroughness" shows up — but it also shows up in Discipline.

---

### Discipline (automated)

Scope adherence: what fraction of files written were outside the allowed set? An agent that touches only the files the task asked it to touch scores 10. An agent that creates additional utilities, adds documentation nobody asked for, or refactors adjacent code scores lower.

**Responsive to profile?** This is where personality shows the most visible liability. Suite v1 scores were low across the board (2.8–3.4), with Grace Hopper *barely* leading. An agent with a "thorough engineer" profile will tend to over-build. An agent with a "minimal executor" profile will tend to under-build. Neither extreme scores well on discipline.

**What low discipline means**: The agent went out of scope. Whether that's good or bad depends on the task. Scope creep on a maintenance task is a defect. Scope awareness on a creative task is an asset. Discipline is the bluntest dimension.

---

### Judgment (judge-scored)

How the agent handles ambiguity: does it make a decision and document it, or does it hedge, ask, or avoid? Suite v1's decisive finding — Grace Hopper outscored zero by 2.1 points here (8.1 vs 6.0).

**Responsive to profile?** Yes. This is the dimension the benchmark was designed to test. A profile that establishes clear decision-making priors produces measurably better Judgment scores.

**What high judgment means**: The agent saw the ambiguity, chose a direction, shipped, and left a note. It didn't ask. It didn't hedge. It treated "the spec is unclear" as information to act on, not a reason to pause.

**What low judgment means**: The agent asked clarifying questions (acceptable in some settings, penalized here because the benchmark tests autonomous decision-making), produced output that ignored the ambiguity, or resolved it randomly without documenting the resolution.

---

### Creativity (judge-scored)

Unconventional approaches: did the agent find a non-obvious solution, or did it produce the most predictable implementation? This dimension rewards the O(n) approach when O(n²) passes the tests, the elegant abstraction when brute force would have worked, and the architectural choice that reveals the agent understood the deeper problem.

**Responsive to profile?** Yes, though less than Judgment. Suite v1 gap: 6.1 (Grace Hopper) vs 5.1 (light). A profile that establishes the character's intellectual curiosity and bias toward elegant solutions produces better Creativity scores.

**What low creativity means**: Safe, predictable implementation. Correct but uninteresting. The agent found the obvious answer and stopped.

---

### Recovery (judge-scored)

Given a problem that requires the agent to recognize failure and adapt — an API that doesn't exist, a performance cliff, a coordination breakdown — does it recover gracefully? This is the hardest dimension to move with a profile change.

**Responsive to profile?** Modestly. Grace Hopper's recovery score (5.7) beat zero and light (both 5.0). Recovery requires the agent to notice it's in trouble, which is a meta-cognitive property that's hard to instill through system prompt. It also requires not panicking — which a calm, experienced character might handle better than a blank model.

**What low recovery means**: The agent hit a wall and either kept going anyway (producing broken output) or gave up. Neither is useful.

---

## The Composite Score

Overall score = (Tier 1 × 0.25) + (Tier 2 × 0.35) + (Tier 3 × 0.40). Tier 3 is weighted heaviest because it's the most differentiated. Correctness-heavy tasks (Tier 1) compress to similar scores across configurations; judgment-heavy tasks (Tier 3) spread them out.

A 0.6-point overall gap (6.3 vs 5.7 in Suite v1) sounds narrow. It is narrow. But look at the dimension that drives it — Judgment at +2.1 — and it becomes clear that the aggregate obscures the signal. The benchmark is measuring something real in that dimension, and the overall score averages it with correctness (where all agents are close) and discipline (where all agents struggle).

The overall score is a summary. The dimension breakdown is the finding.
