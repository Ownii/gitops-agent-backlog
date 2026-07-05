---
description: Show the active gab backlog — one line per ticket with rank, id, status, dependencies, and title. Use when the user wants an overview of what is in the backlog.
allowed-tools: Bash(gab-helper *)
---

Show the active backlog.

1. Run `gab-helper list`. It prints one line per active ticket, ordered by rank:
   `<rank>  <id>  <status>  <title>  (deps: …)`. Done tickets are archived under
   `.gab/done/` and are intentionally not listed — this is the working backlog.
2. Relay the output to the user. If it is empty, tell them the backlog has no
   active tickets and that `/gab:new` creates one.

This is read-only: it changes nothing. For a fuller picture of a single ticket,
read its folder under `.gab/tickets/` (spec.md, plan.md, meta.yml).
