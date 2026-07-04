[ MAIN WORKTREE ]
1. Brainstorming (/gab:new)  -> Erstellt Ticket/Spec in .gab/tickets/
2. Planning                  -> Agent schreibt "Implementation Plan" ins Ticket
3. Approval                  -> Entwickler gibt Plan frei

          │  (Trigger: /gab:start -> TaskCreated Hook)
          ▼
[ ISOLATED WORKTREE ]
4. Environment Setup         -> `git worktree add ...` im Hintergrund
5. TDD & Subagents Loop      -> Agent schreibt Tests -> Subagents fixen Code
6. Verification              -> Interner Test-Run erfolgreich -> Commit & Push

          │  (Trigger: /gab:complete -> TaskCompleted Hook)
          ▼
[ MAIN WORKTREE / QA-PHASE ]
7. State: "To Verify"        -> Worktree wird attached oder Branch bereitgestellt
8. Human QA                  -> Entwickler prüft Code & Verhalten lokal
9. Merge                     -> Freigabe -> Einpflegen in Main-Branch -> Done