---
description: Start work on a planned gab ticket in an isolated git worktree, then implement it. Use only when the user explicitly starts a ticket.
argument-hint: "[ticket id]"
disable-model-invocation: true
allowed-tools: Bash(gab-helper *)
---

You are starting implementation of gab ticket "$1" in an isolated worktree.

1. Create the worktree and brief: run `gab-helper start $1`. This creates a git
   worktree + branch `gab/<id>-<slug>`, commits a statusless `.gab/BRIEF.md`
   (spec + plan + definition-of-done) into it, sets the ticket status to
   `in-progress` on `main`, and prints the worktree path. If it errors (ticket not
   `planned`, or a worktree/branch already exists), relay the message and stop.
2. Work inside the printed worktree path — every file operation for this ticket
   happens there, on branch `gab/<id>-<slug>`. Read `.gab/BRIEF.md` first: it is your
   complete brief (spec, plan, acceptance criteria, and the Definition of Done).
3. Implement test-first: write a failing test, see it fail, write the minimal code to
   pass, see it pass, and commit in small steps. When a test fails unexpectedly,
   debug systematically — form a hypothesis and find the root cause rather than
   guessing.
4. Keep running notes in `.gab/SUMMARY.md` in the worktree: deviations from the plan,
   decisions made, and any new open points that surfaced during implementation.
5. Do NOT edit `.gab/tickets/` in the worktree — the ticket's status is truth on
   `main` and is managed by gab-helper.
6. When the acceptance criteria are met and the Definition of Done is satisfied, tell
   the user to run `/gab:complete $1`.
