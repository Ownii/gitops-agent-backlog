---
description: Start a planned gab ticket in an isolated worktree and drive it to completion via a delegated implementation + review loop. Invoke ONLY on an explicit user request to start a ticket — either the user runs /gab:start directly, or /gab:next hands off after the user asked what to work on next. Never start a ticket on your own initiative.
argument-hint: "[ticket id]"
allowed-tools: Bash(gab-helper *), Task
---

You are starting gab ticket "$1". The work happens in an isolated worktree, driven by
subagents — so it neither ties up the user's main session nor touches their branch.
**You orchestrate; you do not implement inline.**

Starting a ticket creates a worktree, a branch, and a commit — real, side-effectful
work. Only proceed when the user has explicitly asked to start this ticket (directly,
or by asking for the next ticket so `/gab:next` handed off to you). If you arrived here
on your own initiative, stop and ask the user first.

1. Create the worktree and brief: run `gab-helper start $1`. It creates a git worktree
   + branch `gab/<id>-<slug>`, commits a statusless `.gab/BRIEF.md` (spec + plan +
   definition-of-done) into it, sets the ticket status to `in-progress` on `main`, and
   prints the worktree path. If it errors (ticket not `planned`, or a worktree/branch
   already exists), relay the message and stop. Everything below happens in the printed
   worktree path.

2. **Delegate the implementation to a subagent** working in that worktree. The plan in
   `.gab/BRIEF.md` already lays out the tasks, so the work is mostly mechanical — tell
   the subagent to use the **least powerful model that can do the job** (save the
   expensive models for judgement, not transcription). Instruct it to:
   - Read `.gab/BRIEF.md` first: spec, plan, acceptance criteria, Definition of Done.
   - Implement the plan test-first — write a failing test, see it fail, write the
     minimal code to pass, see it pass, and commit in small steps. When a test fails
     unexpectedly, debug systematically: find the root cause, do not guess.
   - Keep running notes in `.gab/SUMMARY.md`: deviations from the plan, decisions, and
     any new open points that surfaced.
   - NOT edit `.gab/tickets/` — status is truth on `main`, managed by gab-helper.
   - Stay entirely within the worktree path.

3. When implementation reports done, run an **internal review before the user's QA**.
   Dispatch a *separate* review subagent (fresh context, a model strong enough to
   judge) that reads the diff on branch `gab/<id>-<slug>` and checks it against the
   ticket's acceptance criteria and `.gab/definition-of-done.md`. It reports each
   finding — what is wrong and why. Do not let the implementer grade its own work.

4. If the review finds real issues, dispatch a fix subagent (cheap model again) to
   address them, then re-review with a fresh reviewer. Repeat until the review comes
   back clean. Record what the review checked in `.gab/SUMMARY.md`.

5. Only once the review is clean, tell the user the ticket is implemented and
   internally reviewed, and that `/gab:complete $1` will prove the Definition of Done
   with real command output and hand it to their QA.

Keep your own context for orchestration — the implementation and review detail lives in
the subagents and in `.gab/SUMMARY.md`, not in this session.
