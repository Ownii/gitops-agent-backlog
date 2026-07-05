# gab — GitOps Agent Backlog

**A local-first, git-native task backlog built for AI coding agents.**

Your backlog lives in your repo as plain files, is the single source of truth (no Jira, no
GitHub Issues required), and each ticket is worked autonomously in an isolated git worktree —
planned, implemented, and verified by an agent, with you at the approval gates.

> **Status: early design phase.** The concept and architecture are specified
> ([design doc](docs/superpowers/specs/2026-07-04-gab-core-design.md)); implementation has
> not started yet. Ideas, feedback, and issues are very welcome.

---

## Why

Working with AI coding agents today has a few recurring frictions:

- **Platform lock-in.** Tickets live in Jira or GitHub — your workflow is bound to a platform
  and a network connection.
- **Context lives elsewhere.** The agent has to reach out to an external system to learn what
  to do, instead of finding the task right next to the code.
- **Resource blocking.** An agent working directly in your main checkout ties up your branch
  and your terminal.

`gab` takes the opposite stance: **the backlog is just files in your repo.** Platform-independent,
versioned alongside your code, diffable, and readable by any agent. Push it to GitHub/GitLab if
you want a cloud copy — but the truth stays local.

## Principles

- **Local-first & platform-independent** — `.gab/` in your repo is the single source of truth.
- **Single-player** — a personal, offline work-queue for one developer and their agents, not a
  team PM tool.
- **Truth on `main`, worktree is the workspace** — status lives on `main`; a feature branch only
  carries a read-only brief and its output, never a second copy of the truth.
- **Portable core, thin adapters** — the real product is a small CLI plus a file convention. The
  Claude Code plugin is just the first adapter; Cursor, Codex, Gemini, and others follow as thin
  manifests.
- **A deliberately "dumb" helper** — the `gab-helper` binary does *only* what must be
  deterministic (git & filesystem state) so the agent can't get it wrong. All judgement stays
  with the agent.

## How it works

```text
[ MAIN ]
  /gab:new         → create a ticket (spec + acceptance criteria)        status: todo
  /gab:plan <id>   → agent explores the repo and writes an impl plan     status: planned
  /gab:next        → pick the next ready ticket and start it

        │  /gab:start <id>
        ▼
[ ISOLATED WORKTREE ]
  · git worktree + branch created; a self-contained brief is committed   status: in-progress
  · TDD loop + subagents implement; a summary of deviations is written
  · the agent must satisfy definition-of-done.md + all acceptance criteria

        │  /gab:complete <id>
        ▼
[ MAIN / QA ]
  · summary flows back to main; branch pushed                            status: to-verify
  · you review the code and behaviour locally

        │  /gab:done <id>
        ▼
  squash-merge → ticket archived to done/ → worktree & branch removed    (done)
```

## Layout in your repo

```text
.gab/
├── definition-of-done.md    # global bar every agent run must meet
├── tickets/                 # the active backlog (stays short — done tickets move out)
│   └── 020-T9-oauth-login/
│       ├── meta.yml         # status, id, priority, depends_on, branch
│       ├── spec.md          # what & why + acceptance criteria
│       ├── plan.md          # implementation plan
│       └── summary.md       # what actually happened, open points (flows back)
└── done/                    # archived, completed tickets
```

- **A ticket is a folder** named `<rank>-<id>-<slug>` — ordering lives in the filename prefix
  (reorder with a single `git mv`), while the stable `id` is what `depends_on` references.
- **Selecting the next ticket** is a cheap filesystem scan: the first `planned` ticket by rank
  whose dependencies are all in `done/`.

## Commands

| Command | Runs on | What it does |
| :--- | :--- | :--- |
| `/gab:new` | main | Brainstorm a ticket: spec + acceptance criteria (`status: todo`) |
| `/gab:plan <id>` | main | Agent writes the implementation plan (`status: planned`) |
| `/gab:next` | main | Pick the next ready ticket and start it |
| `/gab:start <id>` | → worktree | Create worktree + brief, begin the implementation loop |
| `/gab:complete <id>` | worktree → main | Verify done-criteria, push, flow summary back (`to-verify`) |
| `/gab:done <id>` | main | After your QA: squash-merge, archive to `done/`, clean up |

## Portability

The core is agent-neutral by design:

- **`gab-helper`** — a single static Go binary (zero runtime dependency) that handles the
  deterministic git/filesystem work. Any agent that can run a shell can use `gab`.
- **`.gab/` convention** — plain files, readable by anything.
- **Thin per-agent adapters** — Claude Code first (`plugin.json` + skills), then Cursor / Codex /
  Gemini as small manifests. `AGENTS.md` in your project serves as a universal hook, and an
  optional MCP server can later expose `gab`'s operations as native tools to any MCP-capable
  agent.

## Roadmap

**MVP (in design):**
- The six commands, the `.gab/` layout, and the `gab-helper` Go binary
- Dependency gating, manual reordering, single-player / local-first
- Claude Code adapter

**Later:**
- A local, offline web UI to visualise the backlog
- Turning open points from a summary into new tickets
- Additional agent adapters (Cursor, Codex, Gemini) and an MCP server

**Non-goals:** team/multi-user coordination, a cloud/SaaS platform, replacing your issue tracker
for cross-team work. `gab` is intentionally a personal, local tool.

## Try it

`gab` ships as a Claude Code plugin (this repo). Build the helper with
`go build -o bin/gab-helper ./cmd/gab-helper`, enable the plugin, and drive the
loop with `/gab:new → /gab:plan → /gab:next → /gab:complete → /gab:done`. See
[docs/gab-usage.md](docs/gab-usage.md) for setup and the full walkthrough.

## Documentation

- [Core design spec](docs/superpowers/specs/2026-07-04-gab-core-design.md)
- [Product concept](docs/vision.md)
- [Workflow](docs/flow.md)

## License

MIT (planned).
