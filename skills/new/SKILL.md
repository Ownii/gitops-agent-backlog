---
description: Create and refine a new gab ticket (spec + acceptance criteria) in the local .gab backlog. Use when the user wants to add a new task to work on with gab.
argument-hint: "[short title]"
allowed-tools: Bash(gab-helper *)
---

You are creating a new gab ticket. Ticket truth lives in `.gab/tickets/` on the
`main` branch as plain files; `gab-helper` handles the deterministic scaffolding.

1. Turn the user's request "$ARGUMENTS" into a lowercase kebab-case slug — letters,
   digits and hyphens only (e.g. "OAuth Login" -> "oauth-login").
2. Scaffold the ticket folder: run `gab-helper new <slug>`. It prints the created
   folder path and writes `meta.yml` (status: todo) and an empty `spec.md`. It does
   NOT commit.
3. Refine the ticket WITH the user — brainstorm, don't just transcribe. Clarify the
   goal and the "why", the scope boundaries, and concrete, checkable acceptance
   criteria. Ask questions when the request is ambiguous. Keep it tight (YAGNI).
4. Write the result into the folder's `spec.md`: a `## Spec` section (what & why) and
   a `## Acceptance Criteria` section (a checklist of specific, verifiable outcomes).
5. Commit on main: `git add .gab/tickets/<folder>` then
   `git commit -m "gab: new ticket <id> <slug>"`.
6. Tell the user the ticket id and that `/gab:plan <id>` is the next step.

Do not create a worktree or write an implementation plan here — that is `/gab:plan`
and `/gab:start`.
