---
description: Pick the next ready gab ticket from the backlog and start it. Use when the user asks what to work on next or to start the next ticket.
allowed-tools: Bash(gab-helper *), SlashCommand
---

Pick the next ready ticket and start it.

1. Ask the helper for the next ready ticket: run `gab-helper next`.
   - On success it prints a ticket id (e.g. `T9`).
   - If it prints "no ready ticket" (exit code 3), relay the message it prints
     (which explains what is blocking or that nothing is planned yet) and stop.
   - On any other error (e.g. a dependency cycle) relay it and stop.
2. If a ready id was printed, hand off to the start flow: invoke `/gab:start <id>`.
   If that command cannot be invoked automatically here (it is user-gated), tell the
   user the ready ticket id and ask them to run `/gab:start <id>` themselves.
