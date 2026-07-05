# gab Claude Code Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `gab` usable inside Claude Code as `/gab:new`, `/gab:plan`, `/gab:next`, `/gab:start`, `/gab:complete`, `/gab:done` — a thin adapter of skill prose + a plugin manifest that orchestrates the existing `gab-helper` binary.

**Architecture:** The repo root IS the plugin. `.claude-plugin/plugin.json` declares the plugin; `skills/<verb>/SKILL.md` files become the namespaced slash commands. Each skill is self-contained prose that (a) invokes the deterministic `gab-helper` binary for git/filesystem state and (b) carries the AI-reasoning for its lifecycle step — written independently, inspired by (not depending on) known patterns. One small helper change is needed: `gab-helper start` must print the worktree path so the `start` skill knows where to work.

**Tech Stack:** Go (existing `gab-helper`), Claude Code plugin format (Markdown skills + JSON manifest).

## Global Constraints

- **Self-contained, zero dependency:** the adapter must NOT depend on superpowers or any other plugin. Skill prose is written for this project; it may be *inspired by* patterns but references nothing external.
- **Plugin = repo root:** `.claude-plugin/plugin.json` and `skills/` live at the repo root (`/Library/Repos/Privat/gitops-agent-backlog`), alongside the Go source. `bin/gab-helper` is built locally (`go build -o bin/gab-helper ./cmd/gab-helper`) and is auto-added to PATH when the plugin is enabled; it is gitignored (prebuilt release binaries are a later concern).
- **Skill invocation:** `skills/<verb>/SKILL.md` → `/gab:<verb>`. Frontmatter fields used: `description`, `argument-hint`, `allowed-tools: Bash(gab-helper *)`, and `disable-model-invocation: true` on side-effectful verbs (`start`, `complete`, `done`) so Claude never triggers them spontaneously.
- **Skill body substitutions:** `$ARGUMENTS` / `$1` for args; `` !`gab-helper …` `` runs a command and injects its output before Claude reads the prompt; `${CLAUDE_PLUGIN_ROOT}` for plugin-relative paths.
- **Truth on main (from Plan 1, unchanged):** ticket status lives on `main`; `gab-helper` owns all git/filesystem mutations. Skills do reasoning + content, never re-implement the helper's deterministic steps.
- **gab-helper verbs (already built & hardened):** `new <slug>` (scaffold, no commit, prints folder path), `start <id>` (worktree+brief, in-progress), `complete <id>` (summary→main, to-verify, best-effort push), `done <id>` (squash-merge, archive, cleanup, rollback on failure), `next` (print next ready id / "no ready ticket" exit 3 / error exit 1). `plan` is NOT a helper verb — planning is agent-only.

---

## File Structure

```text
.claude-plugin/
  plugin.json              # plugin manifest (name gab, metadata)
  marketplace.json         # local marketplace entry so the plugin can be added
skills/
  new/SKILL.md             # /gab:new  — scaffold + refine spec (brainstorm)
  plan/SKILL.md            # /gab:plan — write plan.md on main, set planned
  next/SKILL.md            # /gab:next — pick next ready ticket, start it
  start/SKILL.md           # /gab:start — worktree + implement (TDD loop)
  complete/SKILL.md        # /gab:complete — verify DoD, flow summary back
  done/SKILL.md            # /gab:done — merge/archive after human QA
AGENTS.md                  # universal hook: how any agent uses gab (.gab + gab-helper)
internal/command/start.go  # MODIFY: Start returns the worktree path
cmd/gab-helper/main.go     # MODIFY: start case prints the worktree path
```

Each `SKILL.md` has one responsibility: orchestrate exactly one lifecycle step. The manifest and `AGENTS.md` are the plugin's front matter. The two Go edits are the single helper change the adapter needs.

---

### Task 1: `gab-helper start` prints the worktree path

The `start` skill must know where the new worktree is so it can work there. `Start` currently returns only `error`; change it to return the worktree path, and have `main.go` print it.

**Files:**
- Modify: `internal/command/start.go`
- Modify: `cmd/gab-helper/main.go`
- Modify: `internal/command/start_test.go`

**Interfaces:**
- Consumes: `repo.Repo.WorktreePath(id, slug)` (existing).
- Produces: `func Start(cwd, id string) (string, error)` — returns the absolute worktree path on success, `""` on error. `main.go`'s `start` case prints that path to stdout and exits 0.

- [ ] **Step 1: Update the test to expect the returned path**

In `internal/command/start_test.go`, change the happy-path call in `TestStartCreatesWorktreeBriefAndStatus` to capture and assert the returned path, and fix the other call sites to the new signature.

Replace the `Start(dir, "T1")` happy-path invocation with:
```go
	got, err := Start(dir, "T1")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := repo.Discover(dir)
	if got != r.WorktreePath("T1", "login") {
		t.Fatalf("Start returned %q, want worktree path %q", got, r.WorktreePath("T1", "login"))
	}
```
In `TestStartRejectsNonPlanned` and `TestStartRejectsExistingWorktree`, change `if err := Start(dir, "T1"); err == nil {` to `if _, err := Start(dir, "T1"); err == nil {`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/command/ -run TestStart -v`
Expected: FAIL — compile error `Start(dir, "T1") (no value) used as value` / assignment mismatch, because `Start` still returns only `error`.

- [ ] **Step 3: Change `Start` to return the path**

In `internal/command/start.go`, change the signature and every return:
- Signature: `func Start(cwd, id string) (string, error) {`
- Every existing `return err` becomes `return "", err`; every `return fmt.Errorf(...)` becomes `return "", fmt.Errorf(...)`; every `return rollback(...)`-style/error return in this function becomes `return "", <err>`.
- The final `return err` (from the last `meta.yml` commit on main) becomes:
```go
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("gab: %s in-progress", id)); err != nil {
		return "", err
	}
	return wt, nil
```
(`wt` is the worktree path already computed earlier in the function.)

- [ ] **Step 4: Print the path in `main.go`**

In `cmd/gab-helper/main.go`, replace the `start` case body with:
```go
	case "start":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper start <id>")
			return 2
		}
		wt, err := command.Start(".", args[1])
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		fmt.Fprintln(stdout, wt)
		return 0
```

- [ ] **Step 5: Run the full suite to verify pass + no regressions**

Run: `go test ./... -count=1 && go build ./...`
Expected: all packages PASS; build succeeds. (Bare `Start(dir, "T1")` statements in `complete_test.go`/`done_test.go` still compile — Go allows discarding all return values of an expression statement.)

- [ ] **Step 6: Commit**

```bash
git add internal/command/start.go cmd/gab-helper/main.go internal/command/start_test.go
git commit -m "feat(helper): start prints the worktree path for the CC adapter"
```

---

### Task 2: Plugin manifest, marketplace entry, and AGENTS.md

**Files:**
- Create: `.claude-plugin/plugin.json`
- Create: `.claude-plugin/marketplace.json`
- Create: `AGENTS.md`

**Interfaces:**
- Produces: an enable-able plugin named `gab`; `AGENTS.md` as the agent-neutral usage hook.

- [ ] **Step 1: Write `.claude-plugin/plugin.json`**

```json
{
  "name": "gab",
  "description": "GitOps Agent Backlog — a local-first, git-native task backlog for AI coding agents.",
  "version": "0.1.0",
  "author": { "name": "Ownii" },
  "homepage": "https://github.com/Ownii/gitops-agent-backlog",
  "repository": "https://github.com/Ownii/gitops-agent-backlog",
  "license": "MIT",
  "keywords": ["backlog", "gitops", "worktree", "agent", "local-first"]
}
```

- [ ] **Step 2: Write `.claude-plugin/marketplace.json`**

First read the reference schema so the shape is exact:
Run: `cat /Users/martin.foerster/.claude/plugins/cache/claude-plugins-official/superpowers/6.1.1/.claude-plugin/marketplace.json`
Then write `.claude-plugin/marketplace.json` modelled on it, with a single plugin whose source is this repo root (`"."`). Concretely:
```json
{
  "name": "gab",
  "owner": { "name": "Ownii" },
  "plugins": [
    {
      "name": "gab",
      "source": ".",
      "description": "Local-first, git-native task backlog for AI coding agents."
    }
  ]
}
```
If the reference file shows a required field this omits, add it to match; do not remove fields shown above.

- [ ] **Step 3: Write `AGENTS.md`**

```markdown
# Working with gab (any agent)

`gab` is a local-first, git-native task backlog. The truth is plain files under
`.gab/` on the `main` branch; a small binary, `gab-helper`, performs all
git/filesystem state changes so an agent never has to do them by hand.

## Layout
- `.gab/tickets/<rank>-<id>-<slug>/` — active tickets: `meta.yml` (status, deps),
  `spec.md`, `plan.md`, and (after work) `summary.md`.
- `.gab/done/` — archived, completed tickets.
- `.gab/definition-of-done.md` — the bar every ticket must meet before QA.

## Lifecycle (and the gab-helper verb each step uses)
1. `gab-helper new <slug>` — scaffold a ticket (status `todo`); then write `spec.md`.
2. write `plan.md`, set `meta.yml` status to `planned` (agent, no helper).
3. `gab-helper start <id>` — create an isolated worktree + brief; status `in-progress`.
   Prints the worktree path; do the work there, keeping notes in `.gab/SUMMARY.md`.
4. `gab-helper complete <id>` — flow the summary back to main; status `to-verify`.
5. human reviews the branch.
6. `gab-helper done <id>` — squash-merge, archive to `.gab/done/`, clean up.

`gab-helper next` prints the id of the next ready ticket (a `planned` ticket
whose dependencies are all done), or exits 3 if none is ready.

In Claude Code these steps are the slash commands `/gab:new`, `/gab:plan`,
`/gab:next`, `/gab:start`, `/gab:complete`, `/gab:done`.
```

- [ ] **Step 4: Validate JSON and commit**

Run: `python3 -c "import json;json.load(open('.claude-plugin/plugin.json'));json.load(open('.claude-plugin/marketplace.json'));print('json ok')"`
Expected: `json ok`
```bash
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json AGENTS.md
git commit -m "feat(plugin): gab plugin manifest, marketplace entry, AGENTS.md"
```

---

### Task 3: Read-side skills — `new`, `plan`, `next`

**Files:**
- Create: `skills/new/SKILL.md`
- Create: `skills/plan/SKILL.md`
- Create: `skills/next/SKILL.md`

**Interfaces:**
- Consumes: `gab-helper new <slug>`, `gab-helper next`. (`plan` uses no helper verb.)
- Produces: `/gab:new`, `/gab:plan`, `/gab:next` slash commands.

- [ ] **Step 1: Write `skills/new/SKILL.md`**

```markdown
---
description: Create and refine a new gab ticket (spec + acceptance criteria) in the local .gab backlog. Use when the user wants to add a new task to work on with gab.
argument-hint: "[short title]"
allowed-tools: Bash(gab-helper *)
---

You are creating a new gab ticket. Ticket truth lives in `.gab/tickets/` on the
`main` branch as plain files; `gab-helper` handles the deterministic scaffolding.

1. Turn the user's request "$ARGUMENTS" into a lowercase kebab-case slug — letters,
   digits and hyphens only (e.g. "OAuth Login" -> "oauth-login").
2. Scaffold the ticket folder: run `gab-helper new <slug>`. It prints the created
   folder path and writes `meta.yml` (status: todo) and an empty `spec.md`. It does
   NOT commit.
3. Refine the ticket WITH the user — brainstorm, don't just transcribe. Clarify the
   goal and the "why", the scope boundaries, and concrete, checkable acceptance
   criteria. Ask questions when the request is ambiguous. Keep it tight (YAGNI).
4. Write the result into the folder's `spec.md`: a `## Spec` section (what & why) and
   a `## Acceptance Criteria` section (a checklist of specific, verifiable outcomes).
5. Commit on main: `git add .gab/tickets/<folder>` then
   `git commit -m "gab: new ticket <id> <slug>"`.
6. Tell the user the ticket id and that `/gab:plan <id>` is the next step.

Do not create a worktree or write an implementation plan here — that is `/gab:plan`
and `/gab:start`.
```

- [ ] **Step 2: Write `skills/plan/SKILL.md`**

```markdown
---
description: Write an implementation plan for a gab ticket, on main, for review before work starts. Use when the user wants to plan a gab ticket.
argument-hint: "[ticket id]"
allowed-tools: Bash(gab-helper *)
---

You are writing the implementation plan for gab ticket "$1". This happens on `main`
so the user can review the plan before any worktree is created. No `gab-helper` verb
is needed — this is your reasoning, written into the ticket.

1. Find the ticket folder under `.gab/tickets/` (its name contains the id, e.g.
   `.gab/tickets/020-T9-...`). Read its `spec.md` and `meta.yml`.
2. Explore the repository so the plan is grounded in the real code: relevant files,
   existing patterns, tests, and the global `.gab/definition-of-done.md`.
3. Write a concrete plan into the ticket's `plan.md`: the approach, the files to
   touch, the test strategy, and a short ordered task list. Reference real paths.
   DRY, YAGNI.
4. Set the ticket status to `planned` by editing the `status:` field in the ticket's
   `meta.yml` (todo -> planned). Do not use gab-helper for this.
5. Commit on main: `git add .gab/tickets/<folder>` then `git commit -m "gab: plan <id>"`.
6. Summarize the plan for the user and tell them `/gab:start <id>` (or `/gab:next`)
   will begin the work in an isolated worktree.
```

- [ ] **Step 3: Write `skills/next/SKILL.md`**

```markdown
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
```

- [ ] **Step 4: Validate frontmatter and commit**

Run: `for f in skills/new skills/plan skills/next; do head -1 "$f/SKILL.md"; done`
Expected: each prints `---` (frontmatter opens on line 1).
Also confirm each file has a closing `---` and a `description:` line:
Run: `grep -L "^description:" skills/new/SKILL.md skills/plan/SKILL.md skills/next/SKILL.md`
Expected: no output (all contain a description).
```bash
git add skills/new skills/plan skills/next
git commit -m "feat(plugin): read-side skills new, plan, next"
```

---

### Task 4: Worktree-side skills — `start`, `complete`, `done`

These are side-effectful (git worktree, merge), so each sets `disable-model-invocation: true` — only the user triggers them.

**Files:**
- Create: `skills/start/SKILL.md`
- Create: `skills/complete/SKILL.md`
- Create: `skills/done/SKILL.md`

**Interfaces:**
- Consumes: `gab-helper start <id>` (prints worktree path, from Task 1), `gab-helper complete <id>`, `gab-helper done <id>`.
- Produces: `/gab:start`, `/gab:complete`, `/gab:done` slash commands.

- [ ] **Step 1: Write `skills/start/SKILL.md`**

```markdown
---
description: Start work on a planned gab ticket in an isolated git worktree, then implement it. Use only when the user explicitly starts a ticket.
argument-hint: "[ticket id]"
disable-model-invocation: true
allowed-tools: Bash(gab-helper *)
---

You are starting implementation of gab ticket "$1" in an isolated worktree.

1. Create the worktree and brief: run `gab-helper start $1`. This creates a git
   worktree + branch `gab/<id>-<slug>`, commits a statusless `.gab/BRIEF.md`
   (spec + plan + definition-of-done) into it, sets the ticket status to
   `in-progress` on `main`, and prints the worktree path. If it errors (ticket not
   `planned`, or a worktree/branch already exists), relay the message and stop.
2. Work inside the printed worktree path — every file operation for this ticket
   happens there, on branch `gab/<id>-<slug>`. Read `.gab/BRIEF.md` first: it is your
   complete brief (spec, plan, acceptance criteria, and the Definition of Done).
3. Implement test-first: write a failing test, see it fail, write the minimal code to
   pass, see it pass, and commit in small steps. When a test fails unexpectedly,
   debug systematically — form a hypothesis and find the root cause rather than
   guessing.
4. Keep running notes in `.gab/SUMMARY.md` in the worktree: deviations from the plan,
   decisions made, and any new open points that surfaced during implementation.
5. Do NOT edit `.gab/tickets/` in the worktree — the ticket's status is truth on
   `main` and is managed by gab-helper.
6. When the acceptance criteria are met and the Definition of Done is satisfied, tell
   the user to run `/gab:complete $1`.
```

- [ ] **Step 2: Write `skills/complete/SKILL.md`**

```markdown
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
```

- [ ] **Step 3: Write `skills/done/SKILL.md`**

```markdown
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
```

- [ ] **Step 4: Validate frontmatter and commit**

Run: `grep -l "disable-model-invocation: true" skills/start/SKILL.md skills/complete/SKILL.md skills/done/SKILL.md`
Expected: all three paths listed (each disables model invocation).
Run: `for f in skills/start skills/complete skills/done; do head -1 "$f/SKILL.md"; done`
Expected: each prints `---`.
```bash
git add skills/start skills/complete skills/done
git commit -m "feat(plugin): worktree-side skills start, complete, done"
```

---

### Task 5: End-to-end verification and usage docs

Skills can only be *invoked* inside a Claude Code session, so automated verification exercises the deterministic layer the skills wrap (the full `gab-helper` lifecycle) plus structural validation of the plugin files. Manual `/gab:*` invocation is documented for the user.

**Files:**
- Create: `docs/gab-usage.md`
- Modify: `README.md` (add an "Install / try it" section linking to `docs/gab-usage.md`)

**Interfaces:**
- Consumes: everything above.

- [ ] **Step 1: Build the helper**

Run: `go build -o bin/gab-helper ./cmd/gab-helper && echo built`
Expected: `built` (binary at `bin/gab-helper`, gitignored).

- [ ] **Step 2: Drive the full lifecycle through gab-helper in a throwaway repo**

This proves the orchestration every skill depends on. Run:
```bash
BIN="$(pwd)/bin/gab-helper"
tmp=$(mktemp -d); cd "$tmp"; git init -b main -q
git config user.email t@e.com; git config user.name T
git commit --allow-empty -m init -q
"$BIN" new login                                   # scaffolds .gab/tickets/010-T1-login
sed -i '' 's/status: todo/status: planned/' .gab/tickets/010-T1-login/meta.yml  # simulate /gab:plan
git add .gab && git commit -q -m "seed"
"$BIN" next                                        # expect: T1
"$BIN" start T1                                    # expect: prints a worktree path
wt=$("$BIN" start T1 2>/dev/null || true)          # (already started; ignore)
echo "gab lifecycle smoke: new/next/start OK"; cd - >/dev/null
```
Expected: `new` prints the folder path; `next` prints `T1`; the first `start T1` prints a worktree path under `.gab-worktrees/`. (This is a smoke check of the wrapped verbs, not the skills themselves.)

- [ ] **Step 3: Structurally validate the plugin so Claude Code can load it**

Run:
```bash
python3 -c "import json; json.load(open('.claude-plugin/plugin.json')); json.load(open('.claude-plugin/marketplace.json')); print('manifest ok')"
ls skills/*/SKILL.md | wc -l   # expect 6
for f in skills/*/SKILL.md; do head -1 "$f" | grep -q '^---' || echo "MISSING frontmatter: $f"; done
```
Expected: `manifest ok`; `6`; no "MISSING frontmatter" lines.

- [ ] **Step 4: Write `docs/gab-usage.md`**

```markdown
# Using gab in Claude Code

## One-time setup
1. Build the helper: `go build -o bin/gab-helper ./cmd/gab-helper`
2. Add this repo as a plugin marketplace and enable `gab` in Claude Code
   (`/plugin marketplace add <path-to-this-repo>` then enable `gab`), so
   `bin/gab-helper` is on PATH and the `/gab:*` commands are available.

## The loop
- `/gab:new <title>` — create and refine a ticket (spec + acceptance criteria).
- `/gab:plan <id>` — write the implementation plan; review it on `main`.
- `/gab:next` — pick the next ready ticket and start it, or
  `/gab:start <id>` — start a specific ticket in an isolated worktree.
- (implement in the worktree; the agent keeps notes in `.gab/SUMMARY.md`)
- `/gab:complete <id>` — verify the Definition of Done, flow the summary back,
  mark `to-verify`.
- review the branch yourself (human QA), then
- `/gab:done <id>` — squash-merge, archive, and clean up.

Everything is local files under `.gab/`; push `main` when you want a cloud copy.
```

- [ ] **Step 5: Add an install pointer to `README.md`**

Add this section to `README.md` immediately before the `## Documentation` section:
```markdown
## Try it

`gab` ships as a Claude Code plugin (this repo). Build the helper with
`go build -o bin/gab-helper ./cmd/gab-helper`, enable the plugin, and drive the
loop with `/gab:new → /gab:plan → /gab:next → /gab:complete → /gab:done`. See
[docs/gab-usage.md](docs/gab-usage.md) for setup and the full walkthrough.
```

- [ ] **Step 6: Commit**

```bash
git add docs/gab-usage.md README.md
git commit -m "docs: gab Claude Code usage and install walkthrough"
```

---

## Self-Review

**1. Spec coverage** (against `docs/superpowers/specs/2026-07-04-gab-core-design.md`):
- §9 six commands mapped to their helper verb + reasoning → Tasks 3, 4 (skills carry the reasoning; helper verbs invoked). `plan` is agent-only (no helper verb) → Task 3, matches §8. ✓
- §9 patterns (brainstorm in `new`; plan in `plan`; TDD + systematic-debugging in `start`; verification-before-completion in `complete`; human-QA gate + finishing in `done`) → written into each skill body, self-contained, no superpowers dependency. ✓
- §10 portability (repo = core; Claude Code = first thin adapter; `AGENTS.md` universal hook) → Task 2. ✓
- §11 plugin structure (`.claude-plugin/plugin.json`, `skills/<verb>/SKILL.md`, `bin/gab-helper` on PATH) → Tasks 1–4. ✓
- Truth-on-main / statusless-brief / summary-writeback all handled by the existing helper the skills call → unchanged. ✓

**2. Placeholder scan:** No "TBD"/"handle appropriately". Every skill body and manifest is written out in full; each task has concrete validation commands with expected output. The only inline judgement left to the executor is Task 2 Step 2 ("match the reference marketplace schema"), which is bounded by reading the concrete reference file. ✓

**3. Type/name consistency:** `Start(cwd, id string) (string, error)` (Task 1) is the only signature change and the `start` skill (Task 4) relies exactly on its printed worktree path. Skill directory names (`new/plan/next/start/complete/done`) match the `/gab:<verb>` commands and the `gab-helper` verbs throughout. `disable-model-invocation: true` appears on exactly `start`, `complete`, `done`. ✓

**Note carried into verification:** the `/gab:*` commands themselves can only be exercised inside a live Claude Code session; Task 5 verifies the wrapped `gab-helper` lifecycle + plugin structure, and `docs/gab-usage.md` tells the user how to enable and run the commands. This boundary is intentional and stated, not a silent gap.
