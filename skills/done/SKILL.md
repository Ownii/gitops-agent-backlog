---
description: Merge an approved (to-verify) gab ticket into main, archive it, and clean up. Use only when the user has QA'd and approved a ticket.
argument-hint: "[ticket id]"
disable-model-invocation: true
allowed-tools: Bash(gab-helper *)
---

You are finishing gab ticket "$1" after the user's QA approval.

1. Confirm the user has reviewed and approved the work — this is the human QA gate.
   If they have not, stop and ask them to review the branch `gab/<id>-<slug>` first.
2. Integrate and clean up: run `gab-helper done $1`. This requires a clean `main`
   worktree, then squash-merges the branch into `main`, discards the branch's `.gab`
   residue (keeping main's `.gab` truth), archives the ticket folder to `.gab/done/`,
   and removes the worktree and branch. If it fails (e.g. a merge conflict) it rolls
   `main` back to its prior state and reports the error — relay it and stop; the
   ticket stays open so the conflict can be resolved.
3. On success, tell the user the ticket is merged and archived, and remind them the
   merge is local until they push `main`.
