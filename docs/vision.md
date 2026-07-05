> **Historical concept — superseded by the [core design spec](superpowers/specs/2026-07-04-gab-core-design.md).**
> This canvas captured the original idea. Some details no longer match what was
> built — notably the storage layout (a ticket is a *folder* of files, not flat
> files) and the command set (six slash commands, not one). Read it for intent,
> not as the current specification.

# PRODUCT CONCEPT CANVAS: `gab` (GitOps Agent Backlog) – Claude Code Plugin

| 1. CORE PROBLEM & PAIN POINTS | 2. TARGET AUDIENCE & USERS |
| :--- | :--- |
| * **Context-Loss bei Agenten:** KI-Agenten verschwenden Token beim Suchen von Dateien, da Jira/GitHub-Tickets nicht maschinenlesbar strukturiert sind.<br>* **Resource Blocking:** Wenn Claude direkt im Hauptverzeichnis arbeitet, blockiert es das Terminal und den aktuellen Branch des Entwicklers.<br>* **Fragmentierte Workflows:** Entwickler müssen zwischen Terminal (Claude Code) und Browser (Jira/GitHub) hin- und herspringen. | * **Software Entwickler**, die Claude Code bereits intensiv im Terminal nutzen und autonome Workflows suchen.<br>* **Engineering Leads & Architekten**, die ein vollkommen lokales, Git-basiertes Task-Management (Local-First / Privacy) ohne Cloud-Zwang etablieren wollen.<br>* **Open-Source-Maintainer**, die Issues isoliert testen lassen möchten. |

| 3. VALUE PROPOSITION (USP) | 4. THE SOLUTION & CLAUDE PLUGIN CORE FEATURES |
| :--- | :--- |
| * **Claude-Native Integration:** Klinkt sich nahtlos in das offizielle Claude Code Plugin-System ein (kein eigenständiges CLI-Setup nötig).<br>* **Git-Isolated (Worktrees via Hooks):** Nutzt native Lifecycle-Hooks, um Claude-Tasks vollautomatisch in isolierte Git-Worktrees auszulagern.<br>* **Agent-Optimized Backlog:** Tickets liegen als strukturiertes Markdown/YAML direkt im Repo (`.gab/`) und füttern Claude mit exakten Datei- und Testkontexten. | * **Slash-Command (`/gab:next`):** Ein Custom Skill scannt das lokale Backlog und startet autonom das nächste priorisierte Ticket.<br>* **Lifecycle Hooks:** `TaskCreated` und `TaskCompleted` steuern die automatische Erstellung und das Bereinigen von Git-Worktrees via Plugin-Skripten.<br>* **`bin/` Executables:** Lokale Shell-Skripte im Plugin-Paket übernehmen das harte Handling der Git-Zustände.<br>* **Live Backlog-Sync:** Claude aktualisiert den Ticket-Status (YAML Frontmatter) vollautomatisch während der Bearbeitung. |

| 5. CHANNELS & DISTRIBUTION | 6. REVENUE & BUSINESS MODEL |
| :--- | :--- |
| * **Claude Code Plugin Registry:** Distribution direkt über das offizielle Plugin-Verzeichnis von Anthropic/Claude Code.<br>* **GitHub Open Source:** Der gesamte Plugin-Quellcode liegt transparent als Community-Projekt auf GitHub.<br>* **Developer Marketing:** Deep-Dive Artikel über "Advanced Multi-Tasking mit Claude Code und Git Worktrees". | * **Open-Core / Free CLI Plugin:** Das Core-Plugin für die lokale Nutzung im Terminal ist 100 % kostenlos.<br>* **Premium Team Layer (SaaS):**<br>  * Ein optionales Web-Dashboard, das die lokalen `.gab/`-Verzeichnisse eines Teams aggregiert und visualisiert.<br>  * Enterprise Security-Hooks (z.B. automatisches Secret-Scanning vor dem Starten eines Sub-Agenten). |

| 7. ARCHITECTURE & PLUGIN STRUCTURE | 8. KEY METRICS FOR SUCCESS |
| :--- | :--- |
| * **`plugin.json`:** Registriert den Namespace `gab` und die Hooks.<br>* **`skills/next/SKILL.md`:** Prompt-Instruktionen für den Ticket-Auswahl-Algorithmus.<br>* **`hooks/hooks.json`:** Abfangen von `TaskCreated` zur Ausführung der Worktree-Logik.<br>* **`bin/gab-helper`:** Bash- oder Python-Skripte für `git worktree add/remove` und Dateimanipulation.<br>* **Storage:** `.gab/tickets/` Ordner im Projekt-Repo. | * **Context Efficiency:** Reduktion verbrauchter Token durch präzise Kontext-Übergabe via YAML-Metadaten.<br>* **Developer Interruption Rate:** Wie selten ein Entwickler eingreifen muss, während Claude im Worktree arbeitet.<br>* **Task Success Rate:** Anteil der Tickets, die Claude autonom und fehlerfrei durch die Test-Pipeline bringt. |

---

## 9. PLUGIN IMPLEMENTATION DETAILS (MVP SCOPE)

### Plugin-Verzeichnisstruktur (`my-plugin/`)
```text
├── .claude-plugin/
│   └── plugin.json       # Definiert Metadaten und Hooks
├── hooks/
│   └── hooks.json        # Bindet "TaskCreated" an bin/gab-helper
├── skills/
│   └── next/
│       └── SKILL.md      # Instruktion für den Befehl /gab:next
└── bin/
    └── gab-helper        # Skript für Git-Worktree-Steuerung & YAML-Updates