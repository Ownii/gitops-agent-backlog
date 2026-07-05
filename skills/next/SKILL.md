---
description: Pick the next ready gab ticket from the backlog and start it. Use when the user asks what to work on next or to start the next ticket.
allowed-tools: Bash(gab-helper *)
---

Pick the next ready ticket and start it.

1. Ask the helper for the next ready ticket: run `gab-helper next`.
   - On success it prints a ticket id (e.g. `T9`).
   - If it prints "no ready ticket" (exit code 3), relay the blocked reasons it lists
     and stop — nothing is startable (either nothing is `planned`, or candidates are
     blocked by unfinished dependencies).
   - On any other error (e.g. a dependency cycle) relay it and stop.
2. If a ready id was printed, start it by invoking `/gab:start <id>` for that id.
