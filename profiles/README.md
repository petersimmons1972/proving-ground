# `profiles/` (repo root)

This directory holds the **built-in control profiles** that ship with
Proving Ground. They are reference baselines used for cross-profile
comparison in every report:

| File         | Role                                                              |
|--------------|-------------------------------------------------------------------|
| `zero.txt`   | Minimal system prompt — the "no profile" floor                    |
| `light.txt`  | A small, broadly-applicable nudge — the "any profile" floor       |

## When to put a file here vs `data/profiles/`

| Path                | Purpose                              | Tracked in git? |
|---------------------|--------------------------------------|-----------------|
| `profiles/`         | Repo-shipped controls (this dir)     | Yes             |
| `data/profiles/`    | YOUR profiles to evaluate            | No (user data)  |

You should **never edit `profiles/`** unless you are contributing a new
control baseline upstream. Put profiles you want to benchmark in
`data/profiles/` — that's where the CLI looks for user-supplied prompts.

See issue #47.
