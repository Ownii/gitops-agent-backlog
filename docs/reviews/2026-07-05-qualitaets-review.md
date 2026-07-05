# Review: Vision/Konzept vs. Implementierung — Verbesserungspotential

**Datum:** 2026-07-05
**Methode:** Abgleich von `docs/vision.md`, `docs/flow.md`, `README.md` und dem Design-Spec
(`docs/superpowers/specs/2026-07-04-gab-core-design.md`) gegen die tatsächliche Implementierung
(Go-Helper, Skills, Plugin-Manifeste). Claude-Code-Plugin-Verhalten wurde gegen die offizielle
Doku verifiziert (PATH-Verhalten von `bin/`, Skill-Frontmatter, `disable-model-invocation`,
`allowed-tools`-Semantik). `go test ./...` grün, `go vet` sauber.

**Nur Dokumentation — keine Änderungen umgesetzt.**

---

## Zusammenfassung (Top-Findings)

| # | Finding | Schwere |
| :-- | :--- | :--- |
| B1 | Kein Release-/CI-Setup: Binary ist gitignored, Marketplace-Install liefert kein `gab-helper` | Hoch |
| B2 | `/gab:next` kann `/gab:start` nie automatisch aufrufen (`disable-model-invocation` blockiert SlashCommand) — Kern-USP „autonom starten" faktisch tot | Hoch |
| C1 | `start`/`complete` committen auf `main` und reißen dabei alles mit, was der User dort bereits gestaged hatte | Hoch |
| C2 | `complete` prüft nicht, ob es im richtigen Worktree/Branch läuft — von `main` aus gestartet geht ein Ticket ohne Arbeit und ohne Summary auf `to-verify` | Hoch |
| A1 | README behauptet „implementation has not started yet" — falsch; `vision.md`/`flow.md` beschreiben eine Hook-Architektur, die es nie gab | Hoch (Außenwirkung) |
| A3 | LICENSE-Datei fehlt (plugin.json deklariert MIT) | Hoch (Distribution) |

Der Rest ist Mittel/Niedrig — Details unten.

---

## A. Dokumentations-Drift (Konzept ↔ Realität)

### A1 — README ist veraltet und widerspricht sich selbst
[README.md:9-11](../../README.md#L9-L11) sagt *„Status: early design phase … implementation has
not started yet"*. Tatsächlich existieren Helper, sechs Skills und eine vollständige Testsuite.
Gleichzeitig sagt der „Try it"-Abschnitt (Z. 127-132), man könne das Plugin bauen und benutzen.
Für ein Open-Source-Projekt mit „Developer Marketing"-Kanal (Vision §5) ist ein widersprüchliches
README das Erste, was Vertrauen kostet.
**Empfehlung:** Status-Block ersetzen (z. B. „MVP implemented, pre-release"), Roadmap-Abschnitt
„MVP (in design)" → „MVP (implemented)".

### A2 — `vision.md` und `flow.md` beschreiben eine verworfene Architektur
Beide Dokumente beschreiben `TaskCreated`/`TaskCompleted`-Lifecycle-Hooks, `hooks/hooks.json`
und nur *einen* Slash-Command (`/gab:next`). Das Design-Spec korrigiert das explizit (§6:
*„Die in flow.md angenommene TaskCreated-Kausalkette existiert so nicht"*), aber die Quelle des
Irrtums wurde nie aktualisiert. Auch das Storage-Layout in der Vision (`.gab/tickets/` als
flache Dateien) entspricht nicht mehr dem Ordner-pro-Ticket-Modell.
**Empfehlung:** Entweder beide Dokumente auf den Ist-Stand heben (6 Commands, Helper-Trigger
statt Hooks, Ticket=Ordner) oder klar als „historisches Konzept, superseded by design spec"
markieren. Aktuell verwirren drei Dokumente mit drei Wahrheiten jeden neuen Leser (und jeden
Agenten, der das Repo als Kontext liest — das Produkt wirbt ausgerechnet mit „Agent-Optimized").

### A3 — LICENSE fehlt
`plugin.json` deklariert `"license": "MIT"`, README sagt „MIT (planned)", eine LICENSE-Datei
gibt es nicht. Ohne sie ist der Code formal *nicht* Open Source — blockiert Distribution und
Community-Beiträge (Vision §5).

---

## B. Distribution & Claude-Code-Adapter

### B1 — Binary-Distribution ist der fehlende Kern des Onboardings
Verifiziert: Claude Code legt das `bin/`-Verzeichnis eines aktivierten Plugins automatisch auf
den PATH — die Mechanik der Skills (`gab-helper …` als Bare-Command) ist also korrekt. **Aber**
`/bin/` ist gitignored ([.gitignore:1](../../.gitignore#L1)): Wer das Repo als Marketplace
installiert, bekommt kein Binary und braucht eine Go-Toolchain plus manuellen Build-Schritt.
Das Design-Spec §14 benennt genau diesen Preis („Build-/Release-Pipeline nötig, z. B.
goreleaser + CI; der Plugin-Adapter liefert/holt das passende Binary") — nichts davon existiert.
Es gibt überhaupt keine CI (auch kein `go test`/`vet`/lint auf Push).
**Empfehlung (priorisiert):**
1. CI-Workflow: `go test ./...`, `go vet`, Lint.
2. Release-Pipeline (goreleaser): Binaries pro OS/Arch.
3. Bootstrap im Plugin: z. B. ein Skill-/Setup-Hinweis oder Script, das das passende Release-
   Binary nach `bin/` holt (oder dokumentierter `go build`-Fallback als Zweitweg).

### B2 — Die „Autonomie" von `/gab:next` ist strukturell kaputt
[skills/start/SKILL.md](../../skills/start/SKILL.md) setzt `disable-model-invocation: true`.
Verifiziert gegen die Doku: Das blockiert **jede** modell-initiierte Invokation — auch über das
SlashCommand-Tool aus `/gab:next` heraus. Der next-Skill kennt das bereits und fällt auf „sag
dem User, er soll `/gab:start <id>` selbst tippen" zurück. Damit ist `/gab:next` de facto nur
noch eine Anzeige („dieses Ticket wäre dran"), während Vision (USP: „startet autonom das
nächste priorisierte Ticket") und README („pick the next ready ticket **and start it**") mehr
versprechen.
**Optionen (Entscheidung nötig):**
- (a) `disable-model-invocation` bei `start` entfernen und den Schutz in die Beschreibung/Prosa
  verlagern („nur nach explizitem User-Auftrag") — `/gab:next` *ist* ein expliziter Auftrag.
- (b) Den Start-Flow im next-Skill inline nachziehen (Duplikation, Drift-Gefahr).
- (c) Versprechen zurücknehmen und `/gab:next` offiziell als reine Auswahl dokumentieren.
Empfehlung: (a) — der User-Intent ist bei `/gab:next` genauso explizit wie bei `/gab:start`.

### B3 — `allowed-tools` deckt die tatsächlichen Skill-Anweisungen nicht
Verifiziert: `allowed-tools` ist Pre-Approval, keine Restriktion — funktional bricht nichts,
aber die UX leidet. Die Skills verlangen Kommandos, die nicht pre-approved sind und daher
prompten:
- `new`/`plan`: `git add` + `git commit` auf main ([skills/new/SKILL.md:78-79](../../skills/new/SKILL.md#L78-L79)).
- `complete`: DoD-Beweise (Testsuite, Lint — beliebige Projekt-Kommandos) plus Commits.
**Empfehlung:** Gezielt ergänzen, z. B. `Bash(git add:*)`, `Bash(git commit:*)` — bewusst eng
halten; die DoD-Kommandos sind projektspezifisch und dürfen weiter prompten.

### B4 — `done` ohne Schutz gegen falschen Ausführungsort
`/gab:complete` sagt explizit „Run this from inside the ticket's worktree"; `/gab:done` sagt
nichts über den Ort. Läuft `gab-helper done` aus dem Ticket-Worktree heraus, versucht der
Helper am Ende `git worktree remove --force` auf das Verzeichnis, in dem der Prozess selbst
steht — bestenfalls eine Warnung, schlimmstenfalls eine Shell in einem gelöschten cwd.
**Empfehlung:** Helper-Guard (cwd unterhalb des zu entfernenden Worktrees → Fehler mit
Anweisung) und einen Satz im done-Skill („run on main").

---

## C. Helper-Robustheit (Korrektheit)

### C1 — `start`/`complete` committen fremde Staged-Changes auf `main` mit
[start.go:76-81](../../internal/command/start.go#L76-L81) und
[complete.go:61-66](../../internal/command/complete.go#L61-L66) machen `git add <pfad>` gefolgt
von `git commit -m …` auf main. `git commit` committet **den gesamten Index** — hatte der User
auf main etwas gestaged (das Produkt ermutigt paralleles Arbeiten, während Agenten in Worktrees
laufen!), landet es unbemerkt im `gab: T9 in-progress`-Commit. `done` prüft sauber auf ein
cleanes main; `start`/`complete` prüfen gar nichts.
**Empfehlung:** Pathspec-Commit (`git commit -m … -- <paths>`) — chirurgisch und ohne den
User zu blockieren. Testfall: staged fremde Datei auf main → darf nicht im gab-Commit landen.

### C2 — `complete` validiert den Ausführungskontext nicht
[complete.go](../../internal/command/complete.go) prüft nur „cwd ist clean" und „Status ist
in-progress". Es prüft **nicht**, dass cwd auf `meta.branch` steht. Von main aus ausgeführt:
main ist clean → Check besteht; `.gab/SUMMARY.md` existiert nicht → wird *silent* übersprungen
(by design für „Agent schrieb keine Summary", hier aber falsch-positiv); Ticket geht auf
`to-verify`, obwohl womöglich kein einziger Commit auf dem Branch liegt.
**Empfehlung:** Guard `git rev-parse --abbrev-ref HEAD` (in cwd) `== meta.branch`, sonst Fehler
mit dem Worktree-Pfad. Optional zusätzlich: warnen, wenn der Branch außer dem Brief-Commit
keine Commits enthält.

### C3 — `start` ist nicht atomar
[start.go:48-82](../../internal/command/start.go#L48-L82): Schlägt nach `git worktree add`
irgendetwas fehl (Brief-Build, Commit, Meta-Write), bleiben Worktree + Branch zurück, der Status
bleibt `planned`, und die Existenz-Guards blockieren jeden Retry mit „remove … and retry" —
manuelle Aufräumarbeit genau dort, wo der Helper laut Design existiert, „damit eine AI es nicht
verkackt". `done` macht es mit seinem Rollback vor.
**Empfehlung:** Best-effort-Rollback (worktree remove + branch -D) bei Fehlern nach dem
`worktree add`.

### C4 — Unbekannte `depends_on`-IDs blockieren still für immer
[backlog.go:127-135](../../internal/backlog/backlog.go#L127-L135): Eine ID, die weder aktiv
noch in `done/` existiert (Tippfehler, gelöschtes Ticket), gilt schlicht als „unmet" — das
Ticket ist dauerhaft blockiert, die Meldung `T9 blocked on T99` verrät nicht, dass T99 gar
nicht existiert. Der plan-Skill verlangt zwar existierende IDs, aber der Helper (die
deterministische Instanz) validiert nicht.
**Empfehlung:** `next` unterscheidet „blocked on T4 (in progress)" von „blocked on T99
(**unknown id — typo?**)". Betrifft Edge-Case-Anspruch aus Design §13.

### C5 — Rank-Überlauf macht Tickets still unsichtbar
[ticket.go:58](../../internal/ticket/ticket.go#L58) verlangt exakt 3 Ziffern (`^(\d{3})-`),
[backlog.go:71-79](../../internal/backlog/backlog.go#L71-L79) vergibt `max(aktiv)+10`. In einem
Backlog, der nie ganz leerläuft, wächst der Rank monoton; ab 1000 erzeugt `new` einen Ordner,
den `ParseFolder` nicht mehr erkennt → das Ticket verschwindet aus `Load` (loadDir ignoriert
unparsbare Ordner **still**). Silent data loss, wenn auch erst nach ~100 überlappenden Tickets.
**Empfehlung:** Regex auf `\d{3,}` lockern (Sortierung läuft über den geparsten Int, nicht
lexikographisch — kein Bruch) + Testfall. Zusätzlich: `loadDir` könnte unparsbare Ordner als
Warnung melden statt still zu schlucken.

### C6 — Worktree-Ablageort: Code und `.gitignore` widersprechen sich
[repo.go:51-53](../../internal/repo/repo.go#L51-L53) legt Worktrees als **Sibling des Repos**
an (`<parent>/.gab-worktrees/<id>-<slug>`), `.gitignore` ignoriert `/.gab-worktrees/` **im**
Repo — einer von beiden irrt. Außerdem enthält der Pfad keinen Repo-Namen: zwei gab-Repos im
selben Elternordner mit gleicher `id`+`slug`-Kombination kollidieren.
**Empfehlung:** Ort bewusst entscheiden und dokumentieren; Repo-Name in den Pfad aufnehmen
(`<parent>/.gab-worktrees/<repo>/<id>-<slug>`); `.gitignore`-Eintrag entsprechend anpassen
oder entfernen.

### C7 — `main` ist hart verdrahtet
[repo.go:27-43](../../internal/repo/repo.go#L27-L43) findet ausschließlich einen Worktree auf
`refs/heads/main`; [start.go:48](../../internal/command/start.go#L48) branched hart von `main`.
Repos mit `master`/`trunk` scheitern an jeder Operation; ebenso der Fall, dass der User im
Haupt-Checkout gerade einen anderen Branch ausgecheckt hat (z. B. mitten in der QA eines
gab-Branches) — dann findet `Discover` kein main und *alle* gab-Kommandos fallen mit einer
generischen Meldung um. „Truth on main" ist Designprinzip, aber der Branch-*Name* muss es nicht
sein.
**Empfehlung:** Default-Branch erkennen (`init.defaultBranch`, `origin/HEAD`) oder als
`.gab/config` konfigurierbar machen; mindestens die Fehlermeldung konkretisieren („main ist in
keinem Worktree ausgecheckt — checke main im Haupt-Checkout aus").

### C8 — Kleineres
- **Warnungen auf stdout statt stderr:** `complete`/`done` drucken Warnungen via `fmt.Printf`
  ([complete.go:71-74](../../internal/command/complete.go#L71-L74),
  [done.go:108-111](../../internal/command/done.go#L108-L111)). Die Skills/Agenten parsen
  stdout als Nutzausgabe — Warnungen gehören auf stderr.
- **`meta.yml`-Kommentare gehen verloren:** jeder Helper-Write (`yaml.Marshal`) zerstört
  Nutzer-Kommentare; das Design-Beispiel (§4) zeigt kommentiertes YAML. Entweder dokumentieren,
  dass `meta.yml` maschinen-owned ist, oder kommentarerhaltend schreiben (yaml.v3 Node-API).
- **Toter Code:** `backlog.Find` ([backlog.go:97-104](../../internal/backlog/backlog.go#L97-L104))
  wird produktiv nirgends verwendet.
- **`done`-Rollback-Residuen:** Nach einem Rollback bleiben die durch den Squash-Merge ins
  Arbeitsverzeichnis gelangten untracked `.gab`-Dateien (BRIEF/SUMMARY) liegen (`reset --hard`
  entfernt keine untracked Files). Kosmetisch, aber verwirrend.
- **ID-Recycling:** Wird ein Ticket-Ordner manuell gelöscht (statt archiviert), vergibt
  `NextID` die ID neu — `depends_on`-Referenzen zeigen dann auf das falsche Ticket. Edge-Case,
  ggf. nur dokumentieren.

---

## D. Produktqualität / Lücken jenseits des Codes

### D1 — Kein `gab-helper list`
Der new-Skill verlangt „look at existing tickets under `.gab/`", `/gab:plan` muss Ordner
suchen, der User hat keinerlei Backlog-Übersicht außer `ls`. Ein read-only `list`-Verb
(rank, id, status, deps, title — eine Zeile pro Ticket) wäre billig, stärkt das Kernversprechen
„Agent-Optimized Backlog" (Vision: Context Efficiency) und ist die natürliche Vorstufe des
späteren `/gab:board`. Bewusst nicht MVP-Scope — aber das beste Preis-Leistungs-Feature auf
der Liste.

### D2 — Env-Setup im Worktree fehlt (bekannte, aber undokumentierte Lücke)
Design §11 sieht optional `hooks.json` (WorktreeCreate → `npm install` o. ä.) vor. Ohne das
startet jeder Implementierungs-Subagent in einem Worktree ohne installierte Dependencies und
muss das selbst merken. Sollte mindestens im start-Skill als Hinweis an den Subagenten stehen
(„richte die Projekt-Dependencies ein, bevor du testest") oder als bewusste Lücke in der
Roadmap auftauchen.

### D3 — Erfolgsmetriken der Vision sind unerhoben
Vision §8 definiert Context Efficiency, Developer Interruption Rate, Task Success Rate — nichts
davon wird (auch nur manuell) erfasst. Für den MVP okay, aber als bewusste Entscheidung
festhalten; sonst bleibt §8 Deko.

### D4 — `.superpowers/`-Artefakte im Repo-Root
Der Ordner ist zwar gitignored-ähnlich strukturiert (eigene `.gitignore`), aber Review-Diffs
und Task-Reports liegen versioniert im Repo. Für ein Projekt mit Public-GitHub-Anspruch:
bewusst entscheiden, ob interne Prozess-Artefakte Teil der Außendarstellung sein sollen.

---

## Was gut ist (und so bleiben sollte)

- **Design-Treue:** Die Trennung „dummer Helper / denkende Skills" ist konsequent umgesetzt;
  der Helper enthält tatsächlich null Fachlogik.
- **`done` ist vorbildlich:** atomare Phase mit Rollback auf Start-SHA, bewusster Umgang mit
  untracked Files, Best-effort-Cleanup erst nach dem Point-of-no-return — genau dieses Niveau
  verdienen auch `start`/`complete` (→ C1-C3).
- **Statusloser Brief + Truth-on-main** funktionieren wie spezifiziert; kein `merge=ours`-
  Gefrickel nötig.
- **Testsuite:** 19 Tests über alle Verben, inkl. Dirty-State-Fällen; `next`-Fehlermeldungen
  erklären Blockaden statt zu schweigen (Design §13 erfüllt).
- **Skills-Prosa:** neue Tickets „interpretation-free" zu erzwingen (one question at a time,
  Ambiguity-Self-Review) und Implementer/Reviewer strikt zu trennen ist stärker als das, was
  die Vision versprach.

## Empfohlene Reihenfolge

1. **Vertrauen & Korrektheit:** C1, C2 (Helper-Guards), A1, A3 (README, LICENSE) — klein, hohe Wirkung.
2. **Produkt-Kernversprechen:** B2 (next→start-Autonomie entscheiden), B1 (CI + Release-Binaries).
3. **Robustheit:** C3, C4, C6, C7, B3, B4.
4. **Politur & Ausbau:** C5, C8, D1 (`list`), D2, A2 (vision/flow aktualisieren).
