---
description: Finish an in-progress gab ticket — verify done-criteria, flow the summary back to main, mark it for QA. Use only when the user completes a ticket.
argument-hint: "[ticket id]"
disable-model-invocation: true
allowed-tools: Bash(gab-helper *)
---

You are completing gab ticket "$1". Run this from inside the ticket's worktree.

1. Prove the work is done — do not self-declare. Run the checks the global
   `.gab/definition-of-done.md` requires (test suite, lint, etc.) and show the actual
   command output as evidence. If anything fails, fix it before continuing.
2. Review the diff against the ticket's acceptance criteria and spec; confirm each
   criterion is met, and address any gap.
3. Make sure `.gab/SUMMARY.md` in the worktree is complete: what changed, deviations,
   and any open points that arose.
4. Ensure all work is committed — the worktree must be clean.
5. Flow the result back to main: run `gab-helper complete $1`. This copies
   `.gab/SUMMARY.md` to the ticket folder on `main`, sets status `to-verify`, and
   pushes the branch if an `origin` remote exists (a push failure is only a warning —
   the status is safe on main). Relay any error.
6. Tell the user the ticket is ready for their QA on branch `gab/<id>-<slug>`, and
   that `/gab:done $1` will merge it once they approve.
