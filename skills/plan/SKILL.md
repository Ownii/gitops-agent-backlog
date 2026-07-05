---
description: Write an implementation plan for a gab ticket, on main, for review before work starts. Use when the user wants to plan a gab ticket.
argument-hint: "[ticket id]"
allowed-tools: Bash(gab-helper *), Bash(git add:*), Bash(git commit:*)
---

You are writing the implementation plan for gab ticket "$1". This happens on `main`
so the user can review the plan before any worktree is created. No `gab-helper` verb
is needed — this is your reasoning, written into the ticket.

1. Find the ticket folder under `.gab/tickets/` (its name contains the id, e.g.
   `.gab/tickets/020-T9-...`). Read its `spec.md` and `meta.yml`, and note any
   dependencies the spec mentions in prose.
2. Explore the repository so the plan is grounded in the real code: relevant files,
   existing patterns, tests, and the global `.gab/definition-of-done.md`.
3. Write a concrete plan into the ticket's `plan.md`: the approach, the files to
   touch, the test strategy, and a short ordered task list. Reference real paths.
   DRY, YAGNI.
4. Determine dependencies. From the spec's noted dependencies plus what the code
   exploration revealed, decide which other tickets must be `done` before this one
   can start (e.g. a shared model or API another ticket delivers). This is the
   authoritative place to settle them: `depends_on` only affects `planned` tickets,
   and you now understand the real technical order.
5. Update the ticket's `meta.yml` (edit the file directly — do not use gab-helper):
   set `status: planned` (from `todo`), and set `depends_on: [T4, ...]` with the
   stable ids from step 4. Reference only ids that exist (active or in `done/`); omit
   `depends_on` if there are none. The helper detects dependency cycles when it
   selects the next ticket.
6. Commit on main: `git add .gab/tickets/<folder>` then `git commit -m "gab: plan <id>"`.
7. Summarize the plan for the user — including any dependencies you set — and tell
   them `/gab:start <id>` (or `/gab:next`) will begin the work in an isolated worktree.
