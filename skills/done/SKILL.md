---
description: Merge an approved (to-verify) gab ticket into main, archive it, and clean up. Use only when the user has QA'd and approved a ticket.
argument-hint: "[ticket id]"
disable-model-invocation: true
allowed-tools: Bash(gab-helper *)
---

You are finishing gab ticket "$1" after the user's QA approval. Run this from the
main checkout, not from inside the ticket's worktree — `done` removes that worktree
at the end, and running from within it would delete your own working directory.

1. Confirm the user has reviewed and approved the work — this is the human QA gate.
   If they have not, stop and ask them to review the branch `gab/<id>-<slug>` first.
2. Handle the open points before you close the ticket. Read the ticket's
   `summary.md` (in its folder under `.gab/tickets/`) and surface any deviations from
   the plan and any open points that surfaced during implementation. Decide their
   disposition with the user — implementation often reveals new work, and this is the
   moment to capture it consciously rather than lose it on merge. If the user wants
   follow-ups, create them now with `/gab:new` (automatically converting open points
   into new `todo` tickets is a planned later feature, not yet available).
3. Integrate and clean up: run `gab-helper done $1`. This requires a clean `main`
   worktree, then squash-merges the branch into `main`, discards the branch's `.gab`
   residue (keeping main's `.gab` truth), archives the ticket folder to `.gab/done/`,
   and removes the worktree and branch. If it fails (e.g. a merge conflict) it rolls
   `main` back to its prior state and reports the error — relay it and stop; the
   ticket stays open so the conflict can be resolved.
4. On success, tell the user the ticket is merged and archived, and remind them the
   merge is local until they push `main`.
