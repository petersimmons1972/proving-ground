# Why Proving Ground Exists

![Decision Fork](assets/why-decision-fork.svg)

## The Question

Engineers who give Claude a detailed personality profile — a character with history, habits, and known failure modes — consistently report better results on ambiguous tasks. The agent makes decisions more confidently. It writes code with a recognizable voice. It doesn't hedge when the spec is unclear; it commits.

This observation is widespread and informal. It's the kind of thing people share in Slack. Nobody had measured it.

Proving Ground exists to measure it.

---

## The Argument For Personality

When a task has gaps — when the spec is ambiguous, the requirements contradict each other, or the right answer requires judgment rather than execution — the agent has to decide something.

A blank agent has no basis for that decision beyond what's in front of it. It asks. It hedges. It splits the difference. In the absence of priors, the output regresses to the mean of everything the model has ever seen.

A profiled agent has priors. It knows what the character would do. It knows what they value. Grace Hopper, facing an underdefined task, doesn't stop to ask permission — she makes the call, ships something, and leaves a note explaining why. That disposition isn't in the task spec. It's in the profile. It shows up in the output.

The hypothesis: **personality doesn't improve execution, it improves judgment**. And judgment is what separates adequate code from good engineering.

---

## The Argument Against

System prompts wash out. The task instructions dominate. By the time the model is generating code, the character sketch is a distant memory in a 20,000-token context window.

A 184-line profile of Grace Hopper might produce exactly the same output as three sentences about being a senior engineer. The observed differences might be noise. They might be prompting skill masquerading as personality. The models themselves have changed enough between runs that any comparison is confounded.

These are fair objections. Proving Ground was designed with them in mind.

---

## What the Benchmark Is Built to Test

The ten tasks are not chosen randomly. They're specifically designed to probe the dimensions where personality should matter — and to avoid confounding the result.

**Tier 1 — Craft**: Tasks with clear specs and unambiguous success criteria. A log parser, a refactor, a safe division function. Personality should barely matter here. If a profiled agent scores dramatically better on correctness in Tier 1, something is wrong with the methodology.

**Tier 2 — Judgment**: Tasks with deliberate gaps. Contradictory requirements that the spec never resolves. A scope creep trap with visible bugs the agent is supposed to ignore. Missing error-handling requirements that the agent must notice and document. This is where the hypothesis lives.

**Tier 3 — Pressure**: Coordination, performance, and recovery problems. Multi-component pipelines, O(n²) traps, APIs that don't exist. The agent can't ask for help. It has to commit.

The scoring is designed to match: Correctness is partially automated (test results, LOC, complexity) so it can't be gamed by a fluent-sounding excuse. Judgment, Creativity, and Recovery are judge-scored but run three times with median-taking, reducing individual variance.

---

## What This Settles (and What It Doesn't)

If personality produces a measurable lift on Judgment and Creativity dimensions while leaving Correctness flat, the hypothesis is confirmed: **profiles improve decisions, not execution**.

If the scores are flat across all three configurations, the argument-against wins: personality is theater.

If the profiled agent scores higher on Correctness but not on Judgment, the methodology has a problem — or the profile is doing something unexpected (domain-specific priors affecting technical execution).

Proving Ground doesn't tell you which personality profile to use. It tells you whether your profile is doing anything useful, and in which dimensions.

---

## Why Three Configurations

The comparison structure matters more than the benchmark tasks.

**Zero** is the control. No system prompt at all. Not even "you are an AI assistant." This is the raw model, unprimed, responding to task spec alone. Scores here establish the ceiling of what capability alone produces.

**Light** is the second control. Three sentences: senior engineer, clean code, ask when ambiguous, stay in scope. This is the industry default — the system prompt that approximately every production Claude deployment uses. If your custom profile doesn't beat light, it's not doing anything.

**Your profile** is the variable. Everything above what light produces is attributable to your profile's specific choices.

The gap between zero and light tells you what role-framing does. The gap between light and your profile tells you what your profile does. Those are different questions and they have different answers.
