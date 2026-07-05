# Using gab in Claude Code

## One-time setup
1. Build the helper: `go build -o bin/gab-helper ./cmd/gab-helper`
2. Add this repo as a plugin marketplace and enable `gab` in Claude Code
   (`/plugin marketplace add <path-to-this-repo>` then enable `gab`), so
   `bin/gab-helper` is on PATH and the `/gab:*` commands are available.

## The loop
- `/gab:new <title>` — create and refine a ticket (spec + acceptance criteria).
- `/gab:plan <id>` — write the implementation plan; review it on `main`.
- `/gab:next` — pick the next ready ticket and start it, or
  `/gab:start <id>` — start a specific ticket in an isolated worktree.
- (implement in the worktree; the agent keeps notes in `.gab/SUMMARY.md`)
- `/gab:complete <id>` — verify the Definition of Done, flow the summary back,
  mark `to-verify`.
- review the branch yourself (human QA), then
- `/gab:done <id>` — squash-merge, archive, and clean up.

Everything is local files under `.gab/`; push `main` when you want a cloud copy.
