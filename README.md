# Proving Ground

Does giving an AI agent a personality make it better at its job?

Proving Ground is a containerized benchmark that measures the effect of agent personality profiles on task execution quality. Run the same tasks with a blank agent, a lightly prompted agent, and your own profiled agent — then see exactly where personality makes the difference.

**STATUS: experimental**

---

## Quick Start

```bash
docker run -e ANTHROPIC_API_KEY=sk-... -v ./data:/data provingground
```

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
5. Run again after improving your profile — history tracking shows your progress

## Documentation

- [Benchmark Design](docs/plans/2026-03-30-benchmark-design.md) — full architecture and task specifications
