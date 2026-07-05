# Working with gab (any agent)

`gab` is a local-first, git-native task backlog. The truth is plain files under
`.gab/` on the `main` branch; a small binary, `gab-helper`, performs all
git/filesystem state changes so an agent never has to do them by hand.

## Layout
- `.gab/tickets/<rank>-<id>-<slug>/` — active tickets: `meta.yml` (status, deps),
  `spec.md`, `plan.md`, and (after work) `summary.md`.
- `.gab/done/` — archived, completed tickets.
- `.gab/definition-of-done.md` — the bar every ticket must meet before QA.

## Lifecycle (and the gab-helper verb each step uses)
1. `gab-helper new <slug>` — scaffold a ticket (status `todo`); then write `spec.md`.
2. write `plan.md`, set `meta.yml` status to `planned` (agent, no helper).
3. `gab-helper start <id>` — create an isolated worktree + brief; status `in-progress`.
   Prints the worktree path; do the work there, keeping notes in `.gab/SUMMARY.md`.
4. `gab-helper complete <id>` — flow the summary back to main; status `to-verify`.
5. human reviews the branch.
6. `gab-helper done <id>` — squash-merge, archive to `.gab/done/`, clean up.

`gab-helper next` prints the id of the next ready ticket (a `planned` ticket
whose dependencies are all done), or exits 3 if none is ready.

In Claude Code these steps are the slash commands `/gab:new`, `/gab:plan`,
`/gab:next`, `/gab:start`, `/gab:complete`, `/gab:done`.
