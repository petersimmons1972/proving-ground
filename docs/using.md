# Using Proving Ground

![Mark 3 Console installed in a receiver room of an East Coast Chain Home radar station, 1939–1945](assets/posters/poster-using-launch.jpg)
*Source: Royal Air Force Radar 1939–1945, Mark 3 Console Receiver Room, RAF official photographer, 1939–1945 — Imperial War Museums, IWM CH 15178 (Public Domain)*

![Execution Pipeline](assets/using-pipeline.svg)

## What Do You Need to Get Started?

- **Docker** (or Go if you prefer to build from source)
- **Anthropic API key** with access to Claude
- **Your agent profile** — a `.txt` or `.md` file describing your agent's personality, working style, and known failure modes. This is the variable you're testing. See [interpreting results](interpreting.md) for what makes a good profile.

A full run costs roughly **$15–40** depending on model pricing. Budget accordingly. You can run individual tiers for a cheaper partial result.

---

## How Do You Run It with Docker?

First, build the image:

```bash
docker build -t provingground .
```

Then run it:

```bash
docker run \
  --env-file .env \
  -v ./data:/data \
  provingground
```

This runs all ten tasks with all three configurations (zero, light, and any profiles you've placed in `data/profiles/`). When it finishes, open `data/results.html`.

To run a single tier:

```bash
docker run \
  --env-file .env \
  -v ./data:/data \
  provingground --tier 1
```

(Both `--env-file .env` and `-e ANTHROPIC_API_KEY=your-key` work; `--env-file` avoids typing the key directly into your shell history.)

Tier 1 is the cheapest run (~$4–10) and gives you a quick read on your profile's correctness and elegance characteristics. Tier 2 is where judgment scores appear. Tier 3 is the most expensive and produces the most interesting data.

### Using docker-compose

A `docker-compose.yml` is included for convenience. Set your API key in the environment before running:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
docker-compose run proving-ground --tier 1
```

To run all tiers: `docker-compose run proving-ground` (no `--tier` flag runs all tiers).

> **Security**: The container runs as a non-root user and uses Docker's default seccomp profile, which blocks dangerous syscalls including `ptrace`, `perf_event_open`, and `kexec_load`. This workload does not require network access — add `--network none` if isolation is desired.

---

## How Do You Provide Your Own Profile?

Place your profile file in the `data/profiles/` directory before running:

```
data/
  profiles/
    your-profile.txt   ← this file becomes a configuration called "your-profile"
```

> **Note:** Create this directory first if it doesn't exist: `mkdir -p data/profiles`

The filename (without extension) becomes the configuration label in the results. If you name it `my-agent.txt`, the results will show a `my-agent` column alongside `zero` and `light`.

**Format**: Plain text. No special syntax required. The entire file contents become the system prompt.

**What to include**: This is not a rigid template. Write what you'd actually want an agent to carry into every task. Effective profiles tend to include:

- Perspective on ambiguity (ask, or decide and document?)
- How to handle scope pressure (gold-plating tendency, or tight executor?)
- What the agent values when tradeoffs appear
- Known failure modes and how to compensate for them

The Grace Hopper profile that shipped with Suite v1 is 184 lines covering her life, her work patterns, her philosophy on moving fast versus cleaning up, and specific operational instructions derived from her historical behavior. Most of it is narrative. The narrative is not decoration — it establishes priors that surface when the tasks require judgment.

You don't need 184 lines. But you need more than three sentences.

---

## How Do You Build and Run Without Docker?

**Option 1 — Build locally (recommended):**

```bash
git clone https://github.com/petersimmons1972/proving-ground.git
cd proving-ground
go build ./cmd/proving-ground

export ANTHROPIC_API_KEY=sk-ant-...
./proving-ground --tier all --data-dir ./data
```

Run from the repo root and the binary finds its assets (`tasks/`, `profiles/`, `prompts/`, `templates/`) automatically.

**Option 2 — `go install`:**

```bash
go install github.com/petersimmons1972/proving-ground/cmd/proving-ground@latest
```

When installed this way, the binary lands in `$GOPATH/bin` with no assets alongside it. You must tell it where to find them using explicit flags:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
REPO=/path/to/cloned/proving-ground

proving-ground \
  --tasks-dir     $REPO/tasks \
  --profiles-dir  $REPO/profiles \
  --prompts-dir   $REPO/prompts \
  --templates-dir $REPO/templates \
  --data-dir      ./data \
  --tier all
```

You'll need Claude Code CLI installed and on your PATH. The benchmark was built with `@anthropic-ai/claude-code@1.0.3`. Install a specific version: `npm install -g @anthropic-ai/claude-code@1.0.3`. Check your installed version: `claude --version`.

Claude Code runs in headless `--print` mode. The `--dangerously-skip-permissions` flag is **not** used — Claude Code enforces its standard permission system. If tasks produce empty output, the agent may need explicit permission grants.

---

## Configuration

Beyond `--tier` and `--data-dir`, three environment variables tune benchmark behavior:

| Variable | Default | Effect |
|----------|---------|--------|
| `PROVING_GROUND_MODEL` | CLI default | Claude model for both runner and judge. Set to pin model for reproducibility (e.g. `claude-3-5-sonnet-20241022`). |
| `PROVING_GROUND_JUDGE_RUNS` | `5` | Judge invocations per task. Reduce to `1` for fast iteration; increase for lower variance. |
| `PROVING_GROUND_MAX_WORKERS` | `4` | Parallel task workers. Reduce to `1` for debugging; increase to run faster (higher API cost). |

Set via `--env-file .env` or `-e VAR=value` in Docker, or `export VAR=value` for the native binary.

---

## What Happens During a Run?

Each task executes independently with each configuration. For a full run with one custom profile, that's 30 executions (10 tasks × 3 configurations). Each execution:

1. Creates an isolated working directory under `data/runs/<timestamp>/<task-id>/<config>/`
2. Launches Claude Code with the task spec
3. Captures the full session transcript and all files written
4. Scores the output: automated metrics first (pytest, LOC, complexity, scope), then LLM-as-judge
5. Judge scoring runs 5 times by default (configurable via `PROVING_GROUND_JUDGE_RUNS`); the median across runs is taken for each dimension to reduce variance

You'll see progress logged to stdout. Expect 45–90 minutes for a full run.

---

## What Files Does It Produce?

After a run, `data/` contains:

| File                   | What it is                                                         |
|------------------------|--------------------------------------------------------------------|
| `results.html`         | Full interactive results page — open this in a browser             |
| `results.json`         | Machine-readable scores in the same structure                      |
| `history.jsonl`        | One line per completed run, all-time — never overwritten           |
| `runs/<timestamp>/`    | Full agent outputs for every task × config pair                    |

**The history file is append-only.** Every run adds a line; nothing is removed. This gives you a longitudinal record as you iterate on your profile.

---

## How Do You Actually Improve Your Profile?

This is what the tool is for. Run the benchmark. Read the results. Find where your profile is weak. Revise. Run again.

That loop is the point. The benchmark gives you a controlled way to test whether a change to your profile made the agent better — or just different. Without a repeatable measurement, you're guessing. With it, you're iterating.

The Judgment and Creativity dimensions respond most visibly to profile changes. Correctness is largely noise — any capable model solves well-specified tasks regardless of what personality you layer on top. Recovery is the hardest to move; it requires the agent to recognize that it's in trouble and adapt mid-task, which is closer to a character property than a learned behavior.

When you find a low dimension score, look at the profile content that addresses that area. If Judgment is low and your profile is thin on decision-making priors, that's the lever. If Discipline is low and your profile is a "thorough engineer" character, that's the profile working as designed — not a defect in the benchmark.

The history file lets you track progress across runs. Don't delete it. Every run is a data point.

---

## Which Tier Should You Run?

| Tier        | Tasks | What it tests                    | Approximate cost |
|-------------|-------|----------------------------------|-----------------|
| `--tier 1`  | 3     | Craft, execution, correctness    | $4–10           |
| `--tier 2`  | 3     | Judgment, ambiguity, scope       | $6–15           |
| `--tier 3`  | 4     | Pressure, coordination, recovery | $8–20           |

`--tier all` (or omitting `--tier` entirely) runs all three tiers sequentially. These are equivalent — `all` is the default.

Tier 2 is the best single-tier run if you're trying to evaluate whether your profile improves decision-making. Tier 1 results are relatively unsurprising. Tier 3 results are the most interesting and the most expensive.
