# Proving Ground

**Does giving an AI agent a personality make it better at its job?**

We ran the same ten tasks with three agents — a blank slate, a lightly-prompted agent, and a fully realized character profile — scored every output across six dimensions, and tallied the results.

**Yes. But not where you'd expect.**

---

![Benchmark Apparatus](docs/assets/readme-apparatus.svg)

## Suite v1 Results — March 2026

![Overall Scores](docs/assets/suite-v1-scores.svg)

The profiled agent outscored the blank agent by **0.6 points overall** (6.3 vs 5.7). The number is narrow. The dimension breakdown is not.

![Dimension Analysis](docs/assets/suite-v1-dimensions.svg)

**Correctness was nearly identical across all three agents** — 9.0, 8.8, 9.1. A blank AI can follow a spec. Personality adds almost nothing there.

**Judgment diverged by 2.1 points.** When the spec was ambiguous and the agent had to decide, the profiled agent made better calls. The blank agent hedged or asked; the profiled agent committed. This is the decisive dimension.

**Creativity diverged by nearly a point** (6.1 vs 5.1/5.3). Under open-ended prompts, the profiled agent generated more interesting solutions. The blank agent generated correct ones.

**Discipline was low across the board** — all three agents over-built and under-stayed-in-scope. This is a finding about current AI agents generally, not about personality. A profile didn't fix it.

→ **[Full narrative report](https://www.petersimmons.com/proving_ground.html)** — complete wartime technical brief with task-level analysis

---

## Quick Start

```bash
docker run \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  -v ./data:/data \
  provingground
```

Place your agent profile in `data/profiles/your-profile.txt` before running. It becomes a configuration alongside `zero` and `light`. **[Full usage guide →](docs/using.md)**

**STATUS: experimental**

---

## What It Measures

Ten tasks across three tiers of increasing difficulty:

- **Craft** — solo execution where personality shows in code elegance
- **Judgment** — ambiguous specs where personality shows in decision-making
- **Pressure** — multi-agent coordination and creative problem-solving under strain

Scored across six dimensions: Correctness, Elegance, Discipline, Judgment, Creativity, Recovery.

## How It Works

1. Provide your API key and optionally your agent profile
2. The benchmark runs each task three times: blank agent, light prompt, your profile
3. Automated metrics + LLM-as-judge score every dimension
4. A single HTML results page shows exactly where your agent excels and where it falls short
5. Run again after improving your profile — history tracking shows your progress over time

---

## Documentation

| Document | What it covers |
|----------|---------------|
| **[Why this exists](docs/why.md)** | The hypothesis, the argument for and against personality, what the benchmark is designed to settle |
| **[How to use it](docs/using.md)** | Running with Docker, providing your own profile, cost estimates, output files |
| **[Reading the results](docs/interpreting.md)** | The three configurations, what each dimension actually measures, how to tell if your profile is working |
| **[Suite v1 archive](docs/suite-v1-results.md)** | Full preserved results from the first run, including both historical runs, task-level analysis, and what changes in Suite v2 |
| **[Benchmark design](docs/plans/2026-03-30-benchmark-design.md)** | Full architecture and task specifications |
