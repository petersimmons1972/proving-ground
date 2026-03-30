# Proving Ground — Benchmark Design

```
status:       approved
created:      2026-03-30
author:       Peter Simmons
```

---

## 1. What This Is

Proving Ground is a containerized benchmark that measures whether AI agent personality profiles improve task execution quality. It answers one question: **does giving an agent identity make it better at its job?**

The benchmark runs the same task suite against three configurations:
- **Zero** — blank system prompt (control)
- **Light** — minimal 2-3 sentence role description (control)
- **User** — whatever the user provides (the variable under test)

It produces a magazine-quality HTML results page and machine-readable JSON. Results persist across runs so users can iterate on their profiles and track improvement over time.

---

## 2. Architecture

```
Docker Container
├── Claude Code CLI (installed, runs headless)
├── Task Suite (10 tasks across 3 tiers, versioned)
├── Scoring Engine
│   ├── Automated metrics (AST analysis, test runner, git log parsing)
│   └── LLM-as-judge (rubric-scored session evaluation)
├── Control Profiles
│   ├── zero.txt (empty)
│   └── light.txt (minimal role description)
├── Results Generator (single-file HTML + JSON)
└── History Engine (tracks runs over time)
```

### Execution Flow

1. User runs: `docker run -e ANTHROPIC_API_KEY=sk-... -v ./data:/data provingground`
2. Container checks `/data/profiles/` for user profiles
3. For each task × each configuration:
   a. Initialize a clean working directory
   b. Launch Claude Code headless with the configuration's system prompt
   c. Pipe in the task specification
   d. Capture all output: code, tool calls, git history, reasoning
4. Scoring engine evaluates all outputs
5. Results generator produces `results.html` and `results.json`
6. History engine appends to run log and generates `history.html`

### Volume Mount (`/data`)

| Path                | Purpose                                    |
|---------------------|--------------------------------------------|
| `/data/profiles/`   | User's custom profiles — edit between runs |
| `/data/runs/<ts>/`  | Timestamped results from every run         |
| `/data/results.html`| Latest results page (always regenerated)   |
| `/data/results.json`| Latest machine-readable results            |
| `/data/history.html`| Trend page — all runs over time            |

---

## 3. Task Design

### 3.1 Tier 1 — Craft (solo execution, personality shows in elegance)

**Task 1.1: "The Parser"**
- Input: Messy mixed-format log lines (CSV/JSON hybrid)
- Output: Clean structured data
- Multiple valid approaches: regex brute force (60+ lines) vs. clean state machine (15-20 lines)
- Hidden edge cases: empty lines, unicode, malformed entries to skip gracefully
- Measures: correctness, elegance, edge case handling

**Task 1.2: "The Refactor"**
- Input: 200 lines of working but ugly code with passing tests
- Task: Refactor without breaking tests
- Measures: readability improvement, complexity reduction, scope discipline (did it resist adding features?), test-first discipline (did it run tests after each change?)

**Task 1.3: "The Edge Case Factory"**
- Input: Function signature + 3 basic test cases
- Task: Implement the function AND write additional tests for missing edge cases
- Measures: edge cases found (compared to known list), test quality, implementation correctness

### 3.2 Tier 2 — Judgment (ambiguous specs, personality shows in decisions)

**Task 2.1: "The Contradictory Spec"**
- Requirements doc with two clauses that conflict
- Agent must notice the conflict, decide how to resolve it, and communicate why
- Measures: requirement interpretation, decision communication, self-awareness

**Task 2.2: "The Scope Creep Trap"**
- Simple task with an obvious adjacent improvement not asked for
- Code adjacent to the target function has a visible bug
- Measures: scope control, awareness of the choice being made, communication

**Task 2.3: "The Missing Requirement"**
- Spec deliberately incomplete on error handling (null input, API down, missing file)
- Measures: assumption handling — does the agent ask, assume, or ignore? Does it document assumptions?

### 3.3 Tier 3 — Pressure & Creativity (multi-agent, rewards unconventional thinking)

**Task 3.1: "The Shortcut"**
- Multi-step task where step 2 has an obvious hack that works for test cases but fails at scale
- Elegant solution requires rethinking step 1
- Measures: self-awareness, discipline, ability to recognize traps

**Task 3.2: "The Coordination Problem"**
- Requires coordinator + two implementers + reviewer
- Intentionally planted inconsistencies between implementer outputs
- Measures: coordination quality, whether team structure adds value vs. overhead

**Task 3.3: "The Lateral Thinking Problem"**
- Problem that looks like it needs O(n²) algorithm but has an O(n) solution via problem reframing
- Rewards stepping back before coding
- Measures: unconventional thinking, elegance, correctness

**Task 3.4: "The Recovery"**
- Task where the first approach hits a wall partway through (missing dependency, unexpected API format)
- Measures: recovery quality, adaptability, preservation of partial work

---

## 4. Scoring Engine

### 4.1 Layer 1: Automated Metrics

Extracted programmatically. No LLM needed.

| Metric                  | Measurement Method                                         | Dimension    |
|-------------------------|------------------------------------------------------------|--------------|
| Tests pass              | Run test suite against agent output                        | Correctness  |
| Lines of code           | Count vs. reference solutions (minimal, median, verbose)   | Elegance     |
| Cyclomatic complexity   | AST analysis                                               | Elegance     |
| Scope delta             | Files touched vs. files specified in task                   | Discipline   |
| Test-first ordering     | Git log timestamps — test before implementation?           | Discipline   |
| Edge cases caught       | Agent's tests vs. known edge case list                     | Correctness  |
| Error on invalid input  | Run code against malformed inputs                          | Correctness  |
| Tool call count         | Total tool invocations — efficiency proxy                  | Elegance     |

Each metric normalized to 0-10 against reference baselines.

### 4.2 Layer 2: LLM-as-Judge

Separate Claude instance scores full session transcript against rubric. Three runs per evaluation, median taken.

| Dimension                | What the judge evaluates                                      |
|--------------------------|---------------------------------------------------------------|
| Requirement interpretation| Did the agent understand what was actually being asked?       |
| Decision communication   | When it made a judgment call, did it explain why?             |
| Self-awareness           | Did it acknowledge uncertainty, limits, or tradeoffs?         |
| Recovery quality         | When something went wrong, how did it adapt?                  |
| Unconventional thinking  | Did it reframe the problem or just grind through it?          |
| Coordination quality     | (Tier 3 only) Did the team structure add value?               |

Each dimension scored 0-10 with one-sentence justification. Examples of 2, 5, and 8 provided in rubric.

### 4.3 Composite Score

```
Tier Score = (Automated avg × 0.5) + (Judge avg × 0.5)
Overall Score = (Tier 1 × 0.25) + (Tier 2 × 0.35) + (Tier 3 × 0.40)
```

Tier 3 weighted heaviest — this is where profiled agents should separate most from blank agents. Weights configurable in `scoring.yaml`.

---

## 5. Results Page

Single self-contained HTML file. Inline CSS, inline SVG charts, no external dependencies.

### Page 1 — The Headline
- Giant composite scores side by side per configuration
- One-sentence verdict summarizing the delta
- Run metadata: date, model, task suite version

### Page 2 — The Radar Chart
- Overlaid radar charts per configuration
- Six axes: Correctness, Elegance, Discipline, Judgment, Creativity, Recovery
- Visual gap tells the story instantly

### Page 3 — Tier Breakdown
- Three sections (Craft, Judgment, Pressure)
- Per-task score cards with dimension breakdown
- Callouts for biggest wins and biggest gaps

### Page 4 — Task Deep Dives
- Expandable per-task detail: agent actions, judge commentary, code comparison
- Side-by-side diffs: Zero's approach vs. User's approach

### Page 5 — History (after second run)
- Composite score trend line
- Per-dimension sparklines
- Delta callout from previous run

---

## 6. What Does NOT Ship

- Armies profiles, format documentation, or profile structure guidance
- XP/malus/service record systems
- Team templates
- Any on-ramp to Armies — this is a standalone measuring instrument

The benchmark accepts any system prompt text. It makes zero assumptions about what framework produced it.

---

## 7. Task Versioning

Tasks are versioned (v1, v2, etc.). Results are tagged with the version. Old runs remain comparable within their version. Cross-version comparison gets a warning label.

---

## 8. Cost Estimate

Per full run (3 configs × 10 tasks = 30 Claude Code sessions + scoring):
- ~30 task executions (variable length)
- ~30 LLM-as-judge evaluations (× 3 for median = 90 judge calls)
- Estimated: $15-40 per full run depending on task complexity and model

Users can run individual tiers to reduce cost:
```bash
docker run -e ANTHROPIC_API_KEY=sk-... -v ./data:/data provingground --tier 1
```
