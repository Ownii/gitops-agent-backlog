# Design: `gab` (GitOps Agent Backlog) — Kern

**Datum:** 2026-07-04
**Status:** Design (genehmigt, vor Implementierungsplan)

## 1. Kontext & Ziel

`gab` ist ein local-first, git-natives Backlog, das im Repo lebt und von KI-Coding-Agenten
gelesen und fortgeschrieben wird. Jedes Ticket wird in einem isolierten Git-Worktree
bearbeitet. Der Kern ist agent-neutral; die Anbindung an einen konkreten Agenten (zuerst
Claude Code) ist eine dünne Verpackung.

Das Produkt ist kein Ersatz für ein Team-PM-Tool und keine Cloud-Plattform. Es ist eine
**persönliche, offline Agent-Arbeitsqueue** für den Solo-/Small-Team-Entwickler, der mit
KI-Agenten am eigenen Produkt arbeitet.

## 2. Leitprinzipien

- **Local-first / plattformunabhängig:** `.gab/` im Repo ist die *einzige Wahrheit*.
  Es besteht kein Zwang zu Jira/GitHub. Ein Push zu einem Remote ist optionales Backup,
  kein zweites Wahrheitsregister.
- **Single-Player:** bewusst kein Multi-User-Koordinationswerkzeug. Damit entfallen
  Merge-Konflikte auf Status-Feldern.
- **Wahrheit auf `main`, Worktree = Arbeitsplatz:** Der Backlog (Status inbegriffen) lebt
  auf `main`. Ein Feature-Branch trägt nur einen statuslosen Lese-Brief und erzeugten
  Output — nie eine zweite Statuswahrheit.
- **Portabler Kern, dünne Adapter:** Der eigentliche Kern ist eine CLI (`gab-helper`) plus
  eine Datei-Konvention (`.gab/`) plus portable Instruktions-Prosa. Pro Agent gibt es nur
  ein kleines Manifest/Shim.
- **Minimaler, "dummer" Helper:** `gab-helper` macht *ausschließlich* das, was
  deterministisch passieren muss, damit eine AI es nicht verkackt. Er ist keine Engine und
  enthält keine fachliche Logik.

## 3. Storage-Layout

Alles im Repo, Wahrheit auf `main`:

```text
.gab/
├── definition-of-done.md    # global; jeder Worktree-Agent muss ihn erfüllen
├── tickets/                 # aktiver Backlog (kurz, weil done rausfliegt)
│   ├── 010-T4-auth-setup/
│   │   ├── meta.yml         # Maschinen-State
│   │   ├── spec.md          # WAS/WARUM + Acceptance Criteria (aus /gab:new)
│   │   ├── plan.md          # Implementation-Plan (aus /gab:plan)
│   │   └── summary.md       # Rückfluss nach Impl (Abweichungen, offene Punkte)
│   └── 020-T9-oauth-login/
└── done/                    # erledigte Ticket-Ordner wandern hierher
```

- **Ticket = Ordner**, nicht Datei. Ordnername: `<rank>-<id>-<slug>`.
- **`rank`** (10er-Gaps): Reihenfolge, per `git mv` änderbar; Platz zum Einschieben (`015`).
- **`id`** (`T4`): stabile Identität für `depends_on`; ändert sich nie.
- **`slug`**: menschenlesbar.
- **`done/`**: Beim Abschluss wird der Ticket-Ordner hierher verschoben. Hält den aktiven
  Backlog schlank und macht "ist T-4 fertig?" zum reinen Datei-Existenz-Check.

### `definition-of-done.md` (global)

Prosa (Checkliste), die *jeder* Worktree-Agent erfüllen muss, bevor er auf `to-verify`
setzt — z.B. "alle Tests grün, Lint clean, keine offenen TODOs, Doku aktualisiert".
Bewusst tech-stack-agnostisch und an *einem* Ort statt pro Ticket. Der DoD schreibt den
konkreten Testlauf selbst vor und ersetzt so den früher angedachten `verify`-Befehl.

## 4. Ticket-Schema

### `meta.yml` (Maschinen-State, getrennt von der Prosa)

```yaml
id: T9
title: OAuth-Login via Google
status: planned        # todo | planned | in-progress | to-verify
                       # (done wird NICHT hier gesetzt; done = Ordner liegt in done/)
priority: high         # optionales Label fürs Diskutieren/Filtern, KEIN Sortierschlüssel
depends_on: [T4]       # IDs; startbar nur wenn alle referenzierten im done/-Ordner liegen
branch: gab/T9-oauth-login   # beim Start gesetzt
```

### `spec.md`
WAS & WARUM plus Acceptance Criteria als Checkliste. Vom Menschen bzw. via `/gab:new`
geschrieben, danach im Prinzip stabil.

```markdown
## Spec
<Was & Warum>

## Acceptance Criteria
- [ ] Nutzer kann sich mit Google-Account einloggen
- [ ] Abgelehnte Zustimmung → saubere Fehlermeldung
```

### `plan.md`
Implementation-Plan, vom Agenten via `/gab:plan` geschrieben (auf `main`).

### `summary.md`
Entsteht *im Worktree* während der Implementierung. Hält Abweichungen vom Plan und neu
entstandene offene Punkte fest und **fließt bei `/gab:complete` zurück nach `main`**.
Im MVP nur festhalten; die Konvertierung offener Punkte in neue Tickets ist ein späteres
Feature.

## 5. Reihenfolge & Auswahl

- **Reihenfolge** lebt im Ordner-Prefix (`rank`), nicht im Datei-Inhalt. Umsortieren =
  `git mv` (meist genau ein Rename dank 10er-Gaps). Kein Content-Read/-Write nötig.
- **`/gab:next`-Auswahl:** Scanne `tickets/` in `rank`-Reihenfolge; nimm das erste Ticket
  mit `status: planned`, dessen `depends_on`-IDs *alle* im `done/`-Ordner liegen. Weil
  erledigte Tickets aus dem Backlog rausfliegen, ist die aktive Liste kurz und der
  Top-down-Scan liest nur wenige `meta.yml`.
- `priority` beeinflusst die Auswahl **nicht** — der `rank` ist die Wahrheit
  (wie "nach oben ziehen" in einem Jira-Backlog).

## 6. Lifecycle & Commands

Zustände: `todo → planned → in-progress → to-verify → done` (done = Ordner in `done/`).

```text
[ MAIN ]
/gab:new            → Ordner anlegen, spec.md (+ Acceptance), meta.status=todo
/gab:plan <id>      → Agent liest spec, exploriert Repo, schreibt plan.md,
                      meta.status=planned            → Mensch reviewt plan.md
/gab:next           → wählt erstes startbares Ticket (siehe §5), ruft /gab:start

        │  /gab:start <id>
        ▼
[ ISOLATED WORKTREE ]  (gab-helper: git worktree add, Branch gab/<id>-<slug>)
  1. Brief-Commit: spec.md + plan.md + definition-of-done.md in den Worktree
  2. meta.status=in-progress (auf main geschrieben), meta.branch gesetzt
  3. TDD- & Subagent-Loop implementiert; schreibt laufend summary.md
  4. Agent erfüllt definition-of-done.md + hakt Acceptance Criteria ab

        │  /gab:complete <id>
        ▼
[ MAIN / QA ]
  1. DoD-Check bestanden → commit & push branch
  2. Rückfluss auf main: summary.md → Ticket-Ordner, meta.status=to-verify
  3. Human QA: prüft Code & Verhalten lokal auf dem Branch

        │  /gab:done <id>   (nach Freigabe)
        ▼
  Squash-Merge in main → Ticket-Ordner nach done/ verschieben → Worktree/Branch entfernen
```

- **Trigger = Command ruft `gab-helper`** — *nicht* ein Lifecycle-Hook. (Die in flow.md
  angenommene `TaskCreated`-Kausalkette existiert so nicht: Slash-Commands sind
  Prompt-Templates und feuern kein Task-Lifecycle-Event.)
- **`/gab:done`** ist das explizite QA-Gate-Ende; `gab` merged nicht autonom.

## 7. Git-Mechanik: Brief rein, Status/Output zurück

- **Kanonisches Ticket** (nur `main`): Spec + Plan + `meta.yml` mit Status. Die eine Wahrheit.
- **Brief** (im Worktree, als Opening-Commit des Feature-Branches): ein self-contained
  Snapshot aus `spec.md` + `plan.md` + `definition-of-done.md` — **ohne** Status-Felder.
  Read-only Auftragszettel; macht die PR selbstdokumentierend. Weil kein Status enthalten
  ist, kann der Brief beim Merge nie mit der `main`-Wahrheit kollidieren.
- **`.gab/tickets/`** wird im Feature-Branch nie angefasst → kein `merge=ours`-Gefrickel,
  keine stale Statuskopie.
- **Status-Writes** gehen *immer* auf `main`, auch aus dem Worktree heraus (der Helper
  schreibt `meta.yml` auf main). `summary.md` ist der Gegenverkehr: im Branch erzeugt, bei
  `/gab:complete` nach `main` zurückgeschrieben.
- **Squash-Merge** bei `/gab:done` hält die `main`-History sauber (auch der Brief-Commit
  bleibt draußen).

## 8. `gab-helper` — Scope (bewusst minimal)

Der Helper macht *nur* das, was schiefgeht, wenn eine AI es freihändig macht. Enge,
deterministische Git/FS-Verben, keine fachliche Logik, keine Inhaltserzeugung.

| Verb | Tut (deterministisch) |
| :--- | :--- |
| `new <slug>` | nächste `id` + `rank` vergeben (kollisionsfrei), Ordner anlegen, `meta.yml` (status=todo) + leere `spec.md` scaffolden |
| `start <id>` | `git worktree add` + Branch `gab/<id>-<slug>`; Brief materialisieren & committen; `meta.status=in-progress` auf main; `meta.branch` setzen |
| `complete <id>` | `summary.md` → `main`-Ticket zurückschreiben; `meta.status=to-verify` auf main; Branch pushen |
| `done <id>` | Squash-Merge in main; Ticket-Ordner → `done/`; Worktree + Branch entfernen |
| `next` | Backlog nach `rank` scannen; erstes startbares Ticket-`id` ausgeben; Dep-Zyklen erkennen & melden |

**Ausdrücklich NICHT im Helper** (bleibt beim Agenten): `spec.md`/`plan.md`/`summary.md`
inhaltlich schreiben, Acceptance abhaken, `definition-of-done.md` interpretieren,
Status-Flips während man auf `main` ist (z.B. `/gab:plan` → `planned` per Datei-Edit —
kein Cross-Tree-Risiko, daher kein Helper nötig).

## 9. Skills & Inspirationen (Reasoning vs. Tooling)

`gab` ist self-contained: es bringt seine eigenen Skills mit und benötigt **keine externe
Abhängigkeit** (kein superpowers-Install). Wir schreiben die Skills eigenständig neu und
lassen uns nur von bewährten Mustern *inspirieren* — Muster/Ideen sind nicht
lizenzpflichtig (MIT/Copyright schützt den konkreten Text, nicht das Konzept). Credit an
superpowers (© 2025 Jesse Vincent, MIT) als Inspiration ist gute Praxis, aber keine Pflicht.

Jeder Command trennt sauber: **deterministisches Tooling (`gab-helper`)** vs.
**AI-Reasoning (eigene Skill-Prosa, inspiriert von etablierten Mustern)**.

| Command | Deterministisch (`gab-helper`) | AI-Reasoning (inspiriert von) |
| :--- | :--- | :--- |
| `/gab:new` (Refine) | ID/Rank vergeben, Ordner scaffolden | brainstorming → `spec.md` + Acceptance Criteria |
| `/gab:plan` | — (Status-Flip auf main per Datei-Edit) | writing-plans → `plan.md`, kennt `.gab`-Pfade & DoD |
| `/gab:start` | Worktree + Branch, Brief committen, `status=in-progress` | TDD + subagent-driven + systematic-debugging; Brief lesen, `summary.md` mitschreiben |
| `/gab:complete` | `summary.md` → main, `status=to-verify`, push | verification-before-completion + interner Review |
| `/gab:done` | Squash-Merge, → `done/`, Worktree/Branch entfernen | finishing-a-development-branch; offene Punkte behandeln |

**Vier Muster, die `gab`s Lücken füllen:**
1. **verification-before-completion → das DoD-Gate.** Der Agent liefert *Evidenz* (echter
   Befehls-Output), bevor er auf `to-verify` schaltet — nicht Selbsteinschätzung. Fest in
   `/gab:complete` verdrahtet.
2. **Interner Review-Agent im Worktree vor `to-verify`** (inspiriert von
   requesting-code-review): prüft gegen Spec + Acceptance + DoD und lässt fixen, bevor die
   menschliche QA drankommt.
3. **systematic-debugging** im TDD-Loop von `/gab:start`: bei roten Tests diszipliniert
   debuggen statt rumprobieren.
4. **finishing-a-development-branch** in `/gab:done`: Disziplin um Merge/Cleanup und der
   Ort, um übrig gebliebene offene Punkte aus `summary.md` bewusst zu behandeln.

**Worktree-Disziplin → Automatik:** superpowers hat `using-git-worktrees` als *manuelle*
Disziplin. `gab` befördert das in Tooling — `gab-helper` legt Worktrees deterministisch an,
der Agent muss nicht daran denken (und kann es damit nicht verkacken).

**`gab`s eigene Erfindung (nicht aus superpowers):** das **`summary.md`-Zurückschreiben in
die Wahrheit** — Learnings/Abweichungen fließen zurück zur Source of Truth. Das ist `gab`s
Alleinstellungsmerkmal.

## 10. Portabilität: ein Kern, viele Adapter

Vorbild ist die Struktur von `superpowers`: eine agent-neutrale Wissens-/Logikbasis plus
dünne Adapter pro Agent (je ein kleines Manifest, das auf denselben Kern zeigt).

- **Universeller Kern (einmal):**
  - `gab-helper` CLI — die deterministische Git/FS-Engine. Von jedem Agenten per Shell
    aufrufbar; funktioniert sogar ganz ohne Plugin (`gab-helper next`).
  - `.gab/`-Layout + `definition-of-done.md` + `meta.yml` — nur Dateien.
  - Instruktions-Prosa als *eine* portable Markdown-Basis.
- **Dünne Adapter pro Agent:** Claude Code zuerst (`.claude-plugin/plugin.json` + `skills/`),
  danach Cursor/Codex/Gemini je als Manifest/Shim — kein Rewrite.
- **`AGENTS.md`** im Zielprojekt als universeller Hebel: Codex, Cursor, Gemini u.a. lesen
  es; ein knapper "so benutzt du gab"-Block macht viele Agenten ohne dedizierten Adapter
  bedienbar.
- **MCP als spätere Option:** `gab-helper` in einen kleinen MCP-Server wickeln → `gab_next`
  / `gab_start` werden native Tools in jedem MCP-fähigen Agenten. MCP liefert nur Tools,
  nicht die Denk-Prosa — die Markdown-Basis bleibt nötig.

## 11. Plugin-Struktur (Claude-Code-Adapter, MVP)

```text
gab/
├── .claude-plugin/
│   └── plugin.json           # Namespace "gab", registriert Commands (+ optionale Hooks)
├── skills/                   # LLM-Reasoning (Prosa; portabler Kern)
│   ├── new/SKILL.md
│   ├── plan/SKILL.md
│   ├── next/SKILL.md
│   ├── start/SKILL.md
│   ├── complete/SKILL.md
│   └── done/SKILL.md
├── hooks/
│   └── hooks.json            # optional: WorktreeCreate → Env-Setup (z.B. npm install)
└── bin/
    └── gab-helper            # deterministische git/fs-Arbeit (§8)
```

**Arbeitsteilung:** Skills = Denken (Spec brainstormen, planen, implementieren).
`gab-helper` = harter State. Diese Trennung ist zugleich die Portabilitäts-Naht: der Helper
ist der universelle Kern, die Skills sind die austauschbare Prosa.

## 12. MVP-Scope

**Im MVP:**
- Commands: `new`, `plan`, `next`, `start`, `complete`, `done`.
- Storage-Layout, `meta.yml`, `spec.md`/`plan.md`/`summary.md`, `definition-of-done.md`.
- `gab-helper` (die 5 Verben aus §8).
- `depends_on`-Gating, `rank`-Reihenfolge via `git mv`.
- Single-Player, local-first, rein Konsole, Claude-Code-Adapter.

**Bewusst später (nicht MVP):**
- Lokale Web-UI-Binary (V2) und `/gab:board`-Rendering.
- Konvertierung offener Punkte aus `summary.md` in neue `todo`-Tickets.
- Weitere Agent-Adapter (Cursor/Codex/Gemini) und MCP-Server.
- Plattform-Export (GitHub/Jira). Alles Team/SaaS ist ausgeschlossen.

## 13. Edge Cases (MVP muss sie sauber behandeln)

- **`/gab:next` findet nichts** (nichts `planned` oder alles durch Deps blockiert):
  klare Meldung, welche Tickets warum blockiert sind.
- **Zyklische `depends_on`:** `gab-helper next` erkennt Zyklen und meldet sie, statt
  endlos zu suchen.
- **Paralleles Arbeiten:** fällt strukturell heraus (jedes `/gab:start` = eigener
  Worktree); mehrere `in-progress` gleichzeitig sind erlaubt, kein Sonderfall.

## 14. Offene Implementierungs-Entscheidungen (nicht blockierend)

- **Sprache von `gab-helper`: Go** (entschieden). Ein statisches Binary, null
  Runtime-Dependency (kein Interpreter/`yq` nötig), schneller Cold-Start, triviale
  Cross-Compilation für macOS/Linux/Windows. Das passt zum Kern-Prinzip "portabel, überall
  lauffähig, von jedem Agenten aufrufbar" besser als ein Python-Script.
  - **Git-Ops:** *nicht* nachbauen — das installierte `git`-Binary via `exec.Command`
    aufrufen (`worktree add/remove`, Squash-Merge, `git mv`, Branch-Delete). Robust und
    verhaltensgleich zum echten Git. `go-git` (pure Go) bewusst *nicht* — bei Worktrees
    schwach.
  - **YAML:** `gopkg.in/yaml.v3` mit getippten Structs für `meta.yml`.
  - **Preis:** Build-/Release-Pipeline nötig (vorgebaute Binaries pro OS/Arch, z.B.
    `goreleaser` + CI); der Plugin-Adapter liefert/holt das passende Binary statt eines
    droppbaren Scripts. Für einen CLI-zentrierten Kern gerechtfertigt.
- **Merge-Detail bei `/gab:done`:** Squash ist gesetzt; Branch-Naming/Cleanup-Konventionen
  im Plan festzurren.
```
