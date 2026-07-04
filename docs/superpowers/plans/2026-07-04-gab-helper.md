# gab-helper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `gab-helper`, a small deterministic Go CLI that owns the git/filesystem state of a `.gab/` backlog — the 5 verbs `new`, `start`, `complete`, `done`, `next`.

**Architecture:** A single static Go binary. Real `git` is invoked via `exec.Command` (never reimplemented). `.gab/` files are the source of truth on the `main` worktree; the binary locates the main worktree from any working directory and writes truth there. Ticket ordering and identity live in folder names (`<rank>-<id>-<slug>`); status/dependencies live in each ticket's `meta.yml`. The binary contains no product judgement — only deterministic state transitions the AI must not do freehand.

**Tech Stack:** Go 1.22+, standard library (`os/exec`, `flag`, `path/filepath`, `regexp`), and `gopkg.in/yaml.v3` for `meta.yml`.

## Global Constraints

- **Language:** Go 1.22+; build a static binary via `go build -o bin/gab-helper ./cmd/gab-helper`.
- **Module path:** `github.com/Ownii/gitops-agent-backlog` (matches the GitHub remote).
- **Git:** shell out to the installed `git` binary via `exec.Command`. Do NOT use `go-git`.
- **Only dependency:** `gopkg.in/yaml.v3`. No CLI framework — dispatch on `os.Args` with `flag` per subcommand.
- **Truth on `main`:** all `.gab/` mutations are written to and committed in the worktree that has `main` checked out. Feature worktrees never mutate `.gab/tickets/`.
- **Status vocabulary:** `todo | planned | in-progress | to-verify`. `done` is NOT a status value — a ticket is done when its folder lives under `.gab/done/`.
- **Folder name format:** `<rank>-<id>-<slug>`, rank zero-padded to 3 digits (`010`), id `T<n>`, slug kebab-case. Sorting is done numerically on the parsed rank, not by string.
- **Determinism only:** the binary generates NO prose content (`spec.md`/`plan.md`/`summary.md` bodies are written by the agent). It scaffolds files, moves them, sets `meta.yml` fields, and runs git.
- **Push is best-effort:** if an `origin` remote exists, push; otherwise skip with a printed notice (never fail because there is no remote).

---

## File Structure

```text
go.mod
go.sum
cmd/gab-helper/
  main.go                 # arg dispatch → command package; usage text; exit codes
internal/gitx/
  gitx.go                 # Run(dir, args...) wrapper; HasRemote
  gitx_test.go
internal/repo/
  repo.go                 # Discover(cwd) → main worktree; .gab path helpers; worktree paths
  repo_test.go
internal/ticket/
  ticket.go               # Meta struct + Read/Write; Folder parse/format; status consts
  ticket_test.go
internal/backlog/
  backlog.go              # Load; NextRank; NextID; Next (selection + cycle detection)
  backlog_test.go
internal/command/
  command.go              # shared helpers (find ticket dir by id, concat brief)
  new.go        new_test.go
  start.go      start_test.go
  complete.go   complete_test.go
  done.go       done_test.go
  next.go       next_test.go
internal/testutil/
  testutil.go             # temp git repo + bare "origin" + worktree helpers for tests
```

Responsibilities: `gitx` is the only place that calls `git`. `repo` is the only place that knows where `.gab/` lives and how to find worktrees. `ticket` owns the on-disk ticket format. `backlog` is pure selection logic over loaded tickets. `command` composes the three lower layers into the 5 verbs. `main` only parses args.

---

### Task 1: Module scaffold and command dispatch

**Files:**
- Create: `go.mod`
- Create: `cmd/gab-helper/main.go`
- Test: `cmd/gab-helper/main_test.go`

**Interfaces:**
- Produces: `func dispatch(args []string, stdout, stderr io.Writer) int` in `main` package — takes CLI args (without program name), returns process exit code. `main()` calls `os.Exit(dispatch(os.Args[1:], os.Stdout, os.Stderr))`.

- [ ] **Step 1: Initialize the module**

Run:
```bash
cd /Library/Repos/Privat/gitops-agent-backlog
go mod init github.com/Ownii/gitops-agent-backlog
go get gopkg.in/yaml.v3@v3.0.1
```
Expected: `go.mod` created with `module github.com/Ownii/gitops-agent-backlog` and a `require gopkg.in/yaml.v3 v3.0.1` line; `go.sum` populated.

- [ ] **Step 2: Write the failing test**

`cmd/gab-helper/main_test.go`:
```go
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDispatchNoArgsShowsUsageAndFails(t *testing.T) {
	var out, errOut bytes.Buffer
	code := dispatch(nil, &out, &errOut)
	if code == 0 {
		t.Fatalf("expected non-zero exit for no args, got 0")
	}
	if !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("expected usage text on stderr, got %q", errOut.String())
	}
}

func TestDispatchUnknownCommandFails(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := dispatch([]string{"frobnicate"}, &out, &errOut); code == 0 {
		t.Fatalf("expected non-zero exit for unknown command")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./cmd/gab-helper/ -run TestDispatch -v`
Expected: FAIL — `undefined: dispatch`.

- [ ] **Step 4: Write minimal implementation**

`cmd/gab-helper/main.go`:
```go
package main

import (
	"fmt"
	"io"
	"os"
)

const usage = `usage: gab-helper <command> [args]

commands:
  new <slug>        scaffold a new ticket folder (status: todo)
  start <id>        create worktree + brief, set status in-progress
  complete <id>     flow summary back to main, set status to-verify, push
  done <id>         squash-merge, archive to done/, remove worktree
  next              print the id of the next ready ticket
`

func dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	// command cases are wired up in later tasks
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func main() {
	os.Exit(dispatch(os.Args[1:], os.Stdout, os.Stderr))
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./cmd/gab-helper/ -run TestDispatch -v`
Expected: PASS (both tests).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum cmd/gab-helper/main.go cmd/gab-helper/main_test.go
git commit -m "feat(helper): module scaffold and command dispatch"
```

---

### Task 2: `gitx` — the git command wrapper

**Files:**
- Create: `internal/gitx/gitx.go`
- Create: `internal/testutil/testutil.go`
- Test: `internal/gitx/gitx_test.go`

**Interfaces:**
- Produces:
  - `func Run(dir string, args ...string) (string, error)` — runs `git args...` in `dir`, returns trimmed stdout; on failure returns an error including stderr.
  - `func HasRemote(dir, name string) bool` — true if remote `name` exists.
  - `testutil.InitRepo(t *testing.T) string` — creates a temp git repo with an initial commit on branch `main`, returns its path.

- [ ] **Step 1: Write the test helper**

`internal/testutil/testutil.go`:
```go
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// git runs a git command in dir and fails the test on error.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// InitRepo creates a temp repo with one commit on branch main.
func InitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-b", "main")
	git(t, dir, "config", "user.email", "test@example.com")
	git(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "initial")
	return dir
}

// AddBareOrigin creates a bare repo and wires it as origin of dir.
func AddBareOrigin(t *testing.T, dir string) string {
	t.Helper()
	bare := t.TempDir()
	git(t, bare, "init", "--bare", "-b", "main")
	git(t, dir, "remote", "add", "origin", bare)
	return bare
}
```

- [ ] **Step 2: Write the failing test**

`internal/gitx/gitx_test.go`:
```go
package gitx

import (
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestRunReturnsStdout(t *testing.T) {
	dir := testutil.InitRepo(t)
	out, err := Run(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if out != "main" {
		t.Fatalf("expected branch main, got %q", out)
	}
}

func TestRunErrorIncludesStderr(t *testing.T) {
	dir := testutil.InitRepo(t)
	if _, err := Run(dir, "cat-file", "-p", "deadbeef"); err == nil {
		t.Fatal("expected error for bad object")
	}
}

func TestHasRemote(t *testing.T) {
	dir := testutil.InitRepo(t)
	if HasRemote(dir, "origin") {
		t.Fatal("no origin expected yet")
	}
	testutil.AddBareOrigin(t, dir)
	if !HasRemote(dir, "origin") {
		t.Fatal("origin expected after AddBareOrigin")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/gitx/ -v`
Expected: FAIL — `undefined: Run` / `undefined: HasRemote`.

- [ ] **Step 4: Write minimal implementation**

`internal/gitx/gitx.go`:
```go
package gitx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes `git args...` in dir and returns trimmed stdout.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// HasRemote reports whether a remote with the given name is configured.
func HasRemote(dir, name string) bool {
	out, err := Run(dir, "remote")
	if err != nil {
		return false
	}
	for _, r := range strings.Fields(out) {
		if r == name {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/gitx/ -v`
Expected: PASS (all three tests).

- [ ] **Step 6: Commit**

```bash
git add internal/gitx/ internal/testutil/
git commit -m "feat(helper): git command wrapper and test harness"
```

---

### Task 3: `repo` — locate the main worktree and `.gab` paths

**Files:**
- Create: `internal/repo/repo.go`
- Test: `internal/repo/repo_test.go`

**Interfaces:**
- Consumes: `gitx.Run`.
- Produces:
  - `type Repo struct { Main string }` — `Main` is the absolute path of the worktree with `main` checked out (the directory that contains `.gab/`).
  - `func Discover(cwd string) (*Repo, error)` — works from the main checkout OR any feature worktree.
  - `func (r *Repo) GabDir() string`, `TicketsDir()`, `DoneDir()`, `DoDPath()` (`.gab/definition-of-done.md`).
  - `func (r *Repo) WorktreePath(id, slug string) string` — deterministic feature-worktree location: `<parent-of-Main>/.gab-worktrees/<id>-<slug>`.

- [ ] **Step 1: Write the failing test**

`internal/repo/repo_test.go`:
```go
package repo

import (
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestDiscoverFromMain(t *testing.T) {
	dir := testutil.InitRepo(t)
	r, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	// EvalSymlinks because macOS temp dirs are symlinked (/var → /private/var).
	got, _ := filepath.EvalSymlinks(r.Main)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("Main = %q, want %q", got, want)
	}
	if filepath.Base(r.GabDir()) != ".gab" {
		t.Fatalf("GabDir = %q", r.GabDir())
	}
}

func TestDiscoverFromFeatureWorktreeFindsMain(t *testing.T) {
	dir := testutil.InitRepo(t)
	wt := filepath.Join(t.TempDir(), "feature")
	if _, err := gitx.Run(dir, "worktree", "add", "-b", "gab/T1-x", wt, "main"); err != nil {
		t.Fatal(err)
	}
	r, err := Discover(wt) // discover from inside the feature worktree
	if err != nil {
		t.Fatal(err)
	}
	got, _ := filepath.EvalSymlinks(r.Main)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("Main = %q, want main checkout %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/repo/ -v`
Expected: FAIL — `undefined: Discover`.

- [ ] **Step 3: Write minimal implementation**

`internal/repo/repo.go`:
```go
package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
)

type Repo struct {
	Main string // worktree path that has `main` checked out
}

// Discover finds the main worktree starting from any working directory
// inside the repository (the main checkout or a feature worktree).
func Discover(cwd string) (*Repo, error) {
	main, err := mainWorktree(cwd)
	if err != nil {
		return nil, err
	}
	return &Repo{Main: main}, nil
}

// mainWorktree parses `git worktree list --porcelain` and returns the path
// of the worktree checked out on refs/heads/main.
func mainWorktree(cwd string) (string, error) {
	out, err := gitx.Run(cwd, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	var curPath string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			curPath = strings.TrimPrefix(line, "worktree ")
		case line == "branch refs/heads/main":
			return curPath, nil
		}
	}
	return "", fmt.Errorf("no worktree checked out on branch main found from %s", cwd)
}

func (r *Repo) GabDir() string     { return filepath.Join(r.Main, ".gab") }
func (r *Repo) TicketsDir() string { return filepath.Join(r.GabDir(), "tickets") }
func (r *Repo) DoneDir() string    { return filepath.Join(r.GabDir(), "done") }
func (r *Repo) DoDPath() string    { return filepath.Join(r.GabDir(), "definition-of-done.md") }

// WorktreePath is the deterministic location for a ticket's feature worktree.
func (r *Repo) WorktreePath(id, slug string) string {
	return filepath.Join(filepath.Dir(r.Main), ".gab-worktrees", id+"-"+slug)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/repo/ -v`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add internal/repo/
git commit -m "feat(helper): repo discovery and .gab path helpers"
```

---

### Task 4: `ticket` — meta.yml and folder-name format

**Files:**
- Create: `internal/ticket/ticket.go`
- Test: `internal/ticket/ticket_test.go`

**Interfaces:**
- Produces:
  - Status constants: `StatusTodo`, `StatusPlanned`, `StatusInProgress`, `StatusToVerify` (string values `"todo"`, `"planned"`, `"in-progress"`, `"to-verify"`).
  - `type Meta struct` with yaml tags: `ID`, `Title`, `Status`, `Priority` (omitempty), `DependsOn []string` (omitempty), `Branch` (omitempty).
  - `func ReadMeta(path string) (*Meta, error)`, `func WriteMeta(path string, m *Meta) error`.
  - `type Folder struct { Rank int; ID, Slug, Name string }`.
  - `func ParseFolder(name string) (Folder, error)`, `func FormatFolder(rank int, id, slug string) string`.

- [ ] **Step 1: Write the failing test**

`internal/ticket/ticket_test.go`:
```go
package ticket

import (
	"path/filepath"
	"testing"
)

func TestFormatAndParseFolder(t *testing.T) {
	name := FormatFolder(20, "T9", "oauth-login")
	if name != "020-T9-oauth-login" {
		t.Fatalf("FormatFolder = %q", name)
	}
	f, err := ParseFolder(name)
	if err != nil {
		t.Fatal(err)
	}
	if f.Rank != 20 || f.ID != "T9" || f.Slug != "oauth-login" {
		t.Fatalf("ParseFolder = %+v", f)
	}
}

func TestParseFolderRejectsBadNames(t *testing.T) {
	for _, bad := range []string{"nope", "10-T9-x", "020-9-x", "020-T9-"} {
		if _, err := ParseFolder(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestWriteThenReadMeta(t *testing.T) {
	p := filepath.Join(t.TempDir(), "meta.yml")
	in := &Meta{ID: "T9", Title: "OAuth", Status: StatusPlanned, DependsOn: []string{"T4"}}
	if err := WriteMeta(p, in); err != nil {
		t.Fatal(err)
	}
	out, err := ReadMeta(p)
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "T9" || out.Status != StatusPlanned || len(out.DependsOn) != 1 || out.DependsOn[0] != "T4" {
		t.Fatalf("round-trip mismatch: %+v", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ticket/ -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Write minimal implementation**

`internal/ticket/ticket.go`:
```go
package ticket

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	StatusTodo       = "todo"
	StatusPlanned    = "planned"
	StatusInProgress = "in-progress"
	StatusToVerify   = "to-verify"
)

type Meta struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Priority  string   `yaml:"priority,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
	Branch    string   `yaml:"branch,omitempty"`
}

func ReadMeta(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &m, nil
}

func WriteMeta(path string, m *Meta) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

type Folder struct {
	Rank int
	ID   string
	Slug string
	Name string
}

var folderRe = regexp.MustCompile(`^(\d{3})-(T\d+)-([a-z0-9]+(?:-[a-z0-9]+)*)$`)

func ParseFolder(name string) (Folder, error) {
	m := folderRe.FindStringSubmatch(name)
	if m == nil {
		return Folder{}, fmt.Errorf("invalid ticket folder name %q", name)
	}
	rank, _ := strconv.Atoi(m[1])
	return Folder{Rank: rank, ID: m[2], Slug: m[3], Name: name}, nil
}

func FormatFolder(rank int, id, slug string) string {
	return fmt.Sprintf("%03d-%s-%s", rank, id, slug)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ticket/ -v`
Expected: PASS (all three tests).

- [ ] **Step 5: Commit**

```bash
git add internal/ticket/
git commit -m "feat(helper): ticket meta.yml and folder-name format"
```

---

### Task 5: `backlog` — load, id/rank allocation, selection, cycle detection

**Files:**
- Create: `internal/backlog/backlog.go`
- Test: `internal/backlog/backlog_test.go`

**Interfaces:**
- Consumes: `repo.Repo`, `ticket.*`.
- Produces:
  - `type Ticket struct { Folder ticket.Folder; Meta *ticket.Meta; Dir string }` — `Dir` is the absolute ticket folder path (under `tickets/`).
  - `func Load(r *repo.Repo) (active []Ticket, doneIDs map[string]bool, err error)` — active tickets sorted by rank asc; `doneIDs` = ids present under `done/`.
  - `func NextRank(active []Ticket) int` — highest active rank + 10 (10 if empty).
  - `func NextID(active []Ticket, doneIDs map[string]bool) string` — `T<max+1>` across active and done (`T1` if empty).
  - `func Next(active []Ticket, doneIDs map[string]bool) (chosen *Ticket, blocked []string, err error)` — first `planned` ticket by rank whose `depends_on` are all in `doneIDs`; `blocked` explains skipped planned tickets; `err` is non-nil only on a dependency cycle among active tickets.
  - `func Find(active []Ticket, id string) (*Ticket, bool)`.

- [ ] **Step 1: Write the failing test**

`internal/backlog/backlog_test.go`:
```go
package backlog

import (
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func mk(rank int, id, status string, deps ...string) Ticket {
	return Ticket{
		Folder: ticket.Folder{Rank: rank, ID: id},
		Meta:   &ticket.Meta{ID: id, Status: status, DependsOn: deps},
	}
}

func TestNextPicksLowestRankReadyPlanned(t *testing.T) {
	active := []Ticket{
		mk(10, "T1", ticket.StatusInProgress),
		mk(20, "T2", ticket.StatusPlanned, "T9"), // blocked: T9 not done
		mk(30, "T3", ticket.StatusPlanned),       // ready
	}
	done := map[string]bool{}
	chosen, blocked, err := Next(active, done)
	if err != nil {
		t.Fatal(err)
	}
	if chosen == nil || chosen.Meta.ID != "T3" {
		t.Fatalf("chosen = %+v", chosen)
	}
	if len(blocked) != 1 {
		t.Fatalf("expected 1 blocked, got %v", blocked)
	}
}

func TestNextReadyWhenDepDone(t *testing.T) {
	active := []Ticket{mk(20, "T2", ticket.StatusPlanned, "T9")}
	chosen, _, err := Next(active, map[string]bool{"T9": true})
	if err != nil {
		t.Fatal(err)
	}
	if chosen == nil || chosen.Meta.ID != "T2" {
		t.Fatalf("expected T2 ready, got %+v", chosen)
	}
}

func TestNextDetectsCycle(t *testing.T) {
	active := []Ticket{
		mk(10, "T1", ticket.StatusPlanned, "T2"),
		mk(20, "T2", ticket.StatusPlanned, "T1"),
	}
	if _, _, err := Next(active, map[string]bool{}); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestNextNoneReadyReturnsNilNoError(t *testing.T) {
	active := []Ticket{mk(10, "T1", ticket.StatusTodo)}
	chosen, _, err := Next(active, map[string]bool{})
	if err != nil || chosen != nil {
		t.Fatalf("expected (nil,nil), got (%+v,%v)", chosen, err)
	}
}

func TestNextIDAndRank(t *testing.T) {
	active := []Ticket{mk(10, "T1", ticket.StatusTodo), mk(20, "T3", ticket.StatusTodo)}
	if got := NextRank(active); got != 30 {
		t.Fatalf("NextRank = %d, want 30", got)
	}
	if got := NextID(active, map[string]bool{"T5": true}); got != "T6" {
		t.Fatalf("NextID = %s, want T6", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/backlog/ -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Write minimal implementation**

`internal/backlog/backlog.go`:
```go
package backlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

type Ticket struct {
	Folder ticket.Folder
	Meta   *ticket.Meta
	Dir    string
}

// Load reads active tickets (sorted by rank) and the set of done ticket ids.
func Load(r *repo.Repo) ([]Ticket, map[string]bool, error) {
	active, err := loadDir(r.TicketsDir(), true)
	if err != nil {
		return nil, nil, err
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Folder.Rank < active[j].Folder.Rank })

	doneTickets, err := loadDir(r.DoneDir(), false)
	if err != nil {
		return nil, nil, err
	}
	done := map[string]bool{}
	for _, d := range doneTickets {
		done[d.Folder.ID] = true
	}
	return active, done, nil
}

// loadDir lists ticket folders in dir. When readMeta is true, meta.yml is loaded.
func loadDir(dir string, readMeta bool) ([]Ticket, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Ticket
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		f, perr := ticket.ParseFolder(e.Name())
		if perr != nil {
			continue // ignore non-ticket folders
		}
		td := Ticket{Folder: f, Dir: filepath.Join(dir, e.Name())}
		if readMeta {
			m, merr := ticket.ReadMeta(filepath.Join(td.Dir, "meta.yml"))
			if merr != nil {
				return nil, merr
			}
			td.Meta = m
		}
		out = append(out, td)
	}
	return out, nil
}

func NextRank(active []Ticket) int {
	max := 0
	for _, t := range active {
		if t.Folder.Rank > max {
			max = t.Folder.Rank
		}
	}
	return max + 10
}

func NextID(active []Ticket, doneIDs map[string]bool) string {
	max := 0
	consider := func(id string) {
		if n, err := strconv.Atoi(strings.TrimPrefix(id, "T")); err == nil && n > max {
			max = n
		}
	}
	for _, t := range active {
		consider(t.Folder.ID)
	}
	for id := range doneIDs {
		consider(id)
	}
	return "T" + strconv.Itoa(max+1)
}

func Find(active []Ticket, id string) (*Ticket, bool) {
	for i := range active {
		if active[i].Meta != nil && active[i].Meta.ID == id || active[i].Folder.ID == id {
			return &active[i], true
		}
	}
	return nil, false
}

// Next returns the first ready planned ticket by rank, an explanation of any
// blocked planned tickets, and a non-nil error only on a dependency cycle.
func Next(active []Ticket, doneIDs map[string]bool) (*Ticket, []string, error) {
	if err := detectCycle(active); err != nil {
		return nil, nil, err
	}
	var blocked []string
	for i := range active {
		t := &active[i]
		if t.Meta.Status != ticket.StatusPlanned {
			continue
		}
		missing := unmetDeps(t.Meta.DependsOn, doneIDs)
		if len(missing) == 0 {
			return t, blocked, nil
		}
		blocked = append(blocked, fmt.Sprintf("%s blocked on %s", t.Meta.ID, strings.Join(missing, ", ")))
	}
	return nil, blocked, nil
}

func unmetDeps(deps []string, doneIDs map[string]bool) []string {
	var missing []string
	for _, d := range deps {
		if !doneIDs[d] {
			missing = append(missing, d)
		}
	}
	return missing
}

// detectCycle reports a cycle in depends_on edges among active tickets.
// Dependencies pointing outside the active set (e.g. to done tickets) are ignored.
func detectCycle(active []Ticket) error {
	index := map[string]*Ticket{}
	for i := range active {
		index[active[i].Meta.ID] = &active[i]
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var visit func(id string) error
	visit = func(id string) error {
		color[id] = gray
		for _, dep := range index[id].Meta.DependsOn {
			if _, ok := index[dep]; !ok {
				continue // dep not among active tickets
			}
			switch color[dep] {
			case gray:
				return fmt.Errorf("dependency cycle involving %s and %s", id, dep)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}
	for id := range index {
		if color[id] == white {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/backlog/ -v`
Expected: PASS (all five tests).

- [ ] **Step 5: Commit**

```bash
git add internal/backlog/
git commit -m "feat(helper): backlog load, allocation, selection, cycle detection"
```

---

### Task 6: `new` command and `.gab` scaffolding

**Files:**
- Create: `internal/command/command.go`
- Create: `internal/command/new.go`
- Test: `internal/command/new_test.go`
- Modify: `cmd/gab-helper/main.go` (wire the `new` case)

**Interfaces:**
- Consumes: `repo`, `backlog`, `ticket`.
- Produces:
  - `func EnsureGab(r *repo.Repo) error` (in `command.go`) — create `tickets/`, `done/`, and a default `definition-of-done.md` if absent.
  - `func TicketDirByID(r *repo.Repo, id string) (string, ticket.Folder, error)` (in `command.go`) — locate an active ticket folder by id.
  - `func New(cwd, slug string) (string, error)` (in `new.go`) — scaffold the folder, return the created folder path. Does NOT commit (the agent fills `spec.md` then commits).

- [ ] **Step 1: Write the failing test**

`internal/command/new_test.go`:
```go
package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestNewScaffoldsFirstTicket(t *testing.T) {
	dir := testutil.InitRepo(t)
	got, err := New(dir, "oauth-login")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "010-T1-oauth-login" {
		t.Fatalf("folder = %q", got)
	}
	if _, err := os.Stat(filepath.Join(got, "spec.md")); err != nil {
		t.Fatalf("spec.md missing: %v", err)
	}
	m, err := ticket.ReadMeta(filepath.Join(got, "meta.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "T1" || m.Status != ticket.StatusTodo {
		t.Fatalf("meta = %+v", m)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gab", "definition-of-done.md")); err != nil {
		t.Fatalf("DoD not scaffolded: %v", err)
	}
}

func TestNewIncrementsIDAndRank(t *testing.T) {
	dir := testutil.InitRepo(t)
	if _, err := New(dir, "first"); err != nil {
		t.Fatal(err)
	}
	got, err := New(dir, "second")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "020-T2-second" {
		t.Fatalf("second folder = %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/command/ -run TestNew -v`
Expected: FAIL — `undefined: New`.

- [ ] **Step 3: Write shared helpers**

`internal/command/command.go`:
```go
package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

const defaultDoD = `# Definition of Done

Every ticket must satisfy these before moving to ` + "`to-verify`" + `:

- [ ] All automated tests pass.
- [ ] Linting/formatting is clean.
- [ ] No leftover TODOs related to this ticket.
- [ ] Public behaviour is documented where it changed.

Edit this file to match your project. The worktree agent must show evidence
(actual command output) that these are met before completing.
`

// EnsureGab creates the .gab skeleton and a default DoD if missing.
func EnsureGab(r *repo.Repo) error {
	for _, d := range []string{r.GabDir(), r.TicketsDir(), r.DoneDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(r.DoDPath()); os.IsNotExist(err) {
		if err := os.WriteFile(r.DoDPath(), []byte(defaultDoD), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// TicketDirByID returns the active ticket folder path and parsed folder for an id.
func TicketDirByID(r *repo.Repo, id string) (string, ticket.Folder, error) {
	entries, err := os.ReadDir(r.TicketsDir())
	if err != nil {
		return "", ticket.Folder{}, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		f, perr := ticket.ParseFolder(e.Name())
		if perr == nil && f.ID == id {
			return filepath.Join(r.TicketsDir(), e.Name()), f, nil
		}
	}
	return "", ticket.Folder{}, fmt.Errorf("no active ticket with id %s", id)
}
```

- [ ] **Step 4: Write minimal implementation**

`internal/command/new.go`:
```go
package command

import (
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

const specTemplate = `## Spec

<what & why>

## Acceptance Criteria

- [ ] 
`

// New scaffolds a new ticket folder and returns its path. It does not commit.
func New(cwd, slug string) (string, error) {
	r, err := repo.Discover(cwd)
	if err != nil {
		return "", err
	}
	if err := EnsureGab(r); err != nil {
		return "", err
	}
	active, doneIDs, err := backlog.Load(r)
	if err != nil {
		return "", err
	}
	id := backlog.NextID(active, doneIDs)
	rank := backlog.NextRank(active)
	folder := ticket.FormatFolder(rank, id, slug)
	dir := filepath.Join(r.TicketsDir(), folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	m := &ticket.Meta{ID: id, Title: slug, Status: ticket.StatusTodo}
	if err := ticket.WriteMeta(filepath.Join(dir, "meta.yml"), m); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(specTemplate), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}
```

- [ ] **Step 5: Wire the command in main.go**

In `cmd/gab-helper/main.go`, add the import and replace the `switch` with a `new` case:
```go
import (
	"fmt"
	"io"
	"os"

	"github.com/Ownii/gitops-agent-backlog/internal/command"
)
```
```go
	case "new":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper new <slug>")
			return 2
		}
		dir, err := command.New(".", args[1])
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		fmt.Fprintln(stdout, dir)
		return 0
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/command/ -run TestNew -v && go build ./...`
Expected: PASS; build succeeds.

- [ ] **Step 7: Commit**

```bash
git add internal/command/command.go internal/command/new.go internal/command/new_test.go cmd/gab-helper/main.go
git commit -m "feat(helper): new command and .gab scaffolding"
```

---

### Task 7: `start` command — worktree, brief, status

**Files:**
- Create: `internal/command/start.go`
- Test: `internal/command/start_test.go`
- Modify: `cmd/gab-helper/main.go` (wire the `start` case)

**Interfaces:**
- Consumes: `repo`, `gitx`, `ticket`, `TicketDirByID`.
- Produces: `func Start(cwd, id string) error` — creates worktree + branch `gab/<id>-<slug>`, commits `.gab/BRIEF.md` (concat of `spec.md` + `plan.md` + `definition-of-done.md`) into it, and on the main worktree sets `meta.status=in-progress`, `meta.branch=<branch>` and commits `meta.yml`.

Precondition: ticket status must be `planned`.

- [ ] **Step 1: Write the failing test**

`internal/command/start_test.go`:
```go
package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

// seedPlanned creates a committed planned ticket with spec.md + plan.md.
func seedPlanned(t *testing.T, dir, id, slug string) string {
	t.Helper()
	r, _ := repo.Discover(dir)
	if err := EnsureGab(r); err != nil {
		t.Fatal(err)
	}
	tdir := filepath.Join(r.TicketsDir(), ticket.FormatFolder(10, id, slug))
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	ticket.WriteMeta(filepath.Join(tdir, "meta.yml"), &ticket.Meta{ID: id, Title: slug, Status: ticket.StatusPlanned})
	os.WriteFile(filepath.Join(tdir, "spec.md"), []byte("## Spec\nlogin\n"), 0o644)
	os.WriteFile(filepath.Join(tdir, "plan.md"), []byte("## Plan\nstep 1\n"), 0o644)
	gitx.Run(dir, "add", "-A")
	gitx.Run(dir, "commit", "-m", "seed "+id)
	return tdir
}

func TestStartCreatesWorktreeBriefAndStatus(t *testing.T) {
	dir := testutil.InitRepo(t)
	tdir := seedPlanned(t, dir, "T1", "login")

	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// status flipped on main
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusInProgress || m.Branch != "gab/T1-login" {
		t.Fatalf("meta after start = %+v", m)
	}
	// worktree exists with a committed BRIEF.md
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	brief, err := os.ReadFile(filepath.Join(wt, ".gab", "BRIEF.md"))
	if err != nil {
		t.Fatalf("BRIEF.md missing in worktree: %v", err)
	}
	if !contains(string(brief), "login") || !contains(string(brief), "Plan") {
		t.Fatalf("brief missing content: %s", brief)
	}
	// branch is committed (brief commit present)
	if _, err := gitx.Run(wt, "rev-parse", "gab/T1-login"); err != nil {
		t.Fatalf("branch not found: %v", err)
	}
}

func TestStartRejectsNonPlanned(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	// force status back to todo
	r, _ := repo.Discover(dir)
	tdir, _, _ := TicketDirByID(r, "T1")
	ticket.WriteMeta(filepath.Join(tdir, "meta.yml"), &ticket.Meta{ID: "T1", Status: ticket.StatusTodo})
	if err := Start(dir, "T1"); err == nil {
		t.Fatal("expected error starting a non-planned ticket")
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (func() bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}()) }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/command/ -run TestStart -v`
Expected: FAIL — `undefined: Start`.

- [ ] **Step 3: Write minimal implementation**

`internal/command/start.go`:
```go
package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Start(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, folder, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	metaPath := filepath.Join(tdir, "meta.yml")
	m, err := ticket.ReadMeta(metaPath)
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusPlanned {
		return fmt.Errorf("ticket %s is %q, must be %q to start", id, m.Status, ticket.StatusPlanned)
	}

	branch := "gab/" + folder.ID + "-" + folder.Slug
	wt := r.WorktreePath(folder.ID, folder.Slug)
	if err := os.MkdirAll(filepath.Dir(wt), 0o755); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "worktree", "add", "-b", branch, wt, "main"); err != nil {
		return err
	}

	// Materialize the statusless brief and commit it on the branch.
	brief, err := buildBrief(tdir, r.DoDPath())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(wt, ".gab"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(wt, ".gab", "BRIEF.md"), brief, 0o644); err != nil {
		return err
	}
	if _, err := gitx.Run(wt, "add", ".gab/BRIEF.md"); err != nil {
		return err
	}
	if _, err := gitx.Run(wt, "commit", "-m", "gab: brief for "+id); err != nil {
		return err
	}

	// Set truth on main.
	m.Status = ticket.StatusInProgress
	m.Branch = branch
	if err := ticket.WriteMeta(metaPath, m); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "add", metaPath); err != nil {
		return err
	}
	_, err = gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("gab: %s in-progress", id))
	return err
}

// buildBrief concatenates spec.md, plan.md and the global DoD into one file.
func buildBrief(ticketDir, dodPath string) ([]byte, error) {
	var b []byte
	appendFile := func(path, heading string) error {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		b = append(b, []byte("<!-- "+heading+" -->\n")...)
		b = append(b, data...)
		b = append(b, '\n')
		return nil
	}
	if err := appendFile(filepath.Join(ticketDir, "spec.md"), "spec"); err != nil {
		return nil, err
	}
	if err := appendFile(filepath.Join(ticketDir, "plan.md"), "plan"); err != nil {
		return nil, err
	}
	if err := appendFile(dodPath, "definition-of-done"); err != nil {
		return nil, err
	}
	return b, nil
}
```

- [ ] **Step 4: Wire the command in main.go**

Add to the `switch`:
```go
	case "start":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper start <id>")
			return 2
		}
		if err := command.Start(".", args[1]); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		return 0
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/command/ -run TestStart -v && go build ./...`
Expected: PASS; build succeeds.

- [ ] **Step 6: Commit**

```bash
git add internal/command/start.go internal/command/start_test.go cmd/gab-helper/main.go
git commit -m "feat(helper): start command (worktree, brief, in-progress)"
```

---

### Task 8: `complete` command — flow summary back, to-verify, push

**Files:**
- Create: `internal/command/complete.go`
- Test: `internal/command/complete_test.go`
- Modify: `cmd/gab-helper/main.go` (wire the `complete` case)

**Interfaces:**
- Consumes: `repo`, `gitx`, `ticket`, `TicketDirByID`.
- Produces: `func Complete(cwd, id string) error` — run from the feature worktree. Requires a clean worktree (no uncommitted changes). Copies `<worktree>/.gab/SUMMARY.md` (if present) to `<main>/.gab/tickets/<...>/summary.md`, sets `meta.status=to-verify` on main and commits, then pushes the branch when an `origin` remote exists.

Precondition: ticket status must be `in-progress`.

- [ ] **Step 1: Write the failing test**

`internal/command/complete_test.go`:
```go
package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestCompleteFlowsSummaryAndSetsToVerify(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")

	// Agent does work: writes summary + a source file, commits in the worktree.
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("done\n"), 0o644)
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("## Summary\nno deviations\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "implement login")

	// complete is invoked from the worktree
	if err := Complete(wt, "T1"); err != nil {
		t.Fatal(err)
	}

	tdir, _, _ := TicketDirByID(r, "T1")
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusToVerify {
		t.Fatalf("status = %q, want to-verify", m.Status)
	}
	if _, err := os.Stat(filepath.Join(tdir, "summary.md")); err != nil {
		t.Fatalf("summary not flowed back: %v", err)
	}
}

func TestCompleteRejectsDirtyWorktree(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "dirty.txt"), []byte("x"), 0o644) // uncommitted
	if err := Complete(wt, "T1"); err == nil {
		t.Fatal("expected error for dirty worktree")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/command/ -run TestComplete -v`
Expected: FAIL — `undefined: Complete`.

- [ ] **Step 3: Write minimal implementation**

`internal/command/complete.go`:
```go
package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Complete(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, _, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	metaPath := filepath.Join(tdir, "meta.yml")
	m, err := ticket.ReadMeta(metaPath)
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusInProgress {
		return fmt.Errorf("ticket %s is %q, must be %q to complete", id, m.Status, ticket.StatusInProgress)
	}
	if m.Branch == "" {
		return fmt.Errorf("ticket %s has no branch recorded", id)
	}

	// The feature worktree (cwd) must be clean.
	status, err := gitx.Run(cwd, "status", "--porcelain")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("worktree has uncommitted changes; commit before completing:\n%s", status)
	}

	// Flow summary.md back to the truth on main (if the agent wrote one).
	src := filepath.Join(cwd, ".gab", "SUMMARY.md")
	if data, rerr := os.ReadFile(src); rerr == nil {
		if err := os.WriteFile(filepath.Join(tdir, "summary.md"), data, 0o644); err != nil {
			return err
		}
		if _, err := gitx.Run(r.Main, "add", filepath.Join(tdir, "summary.md")); err != nil {
			return err
		}
	} else if !os.IsNotExist(rerr) {
		return rerr
	}

	// Set status to-verify on main and commit.
	m.Status = ticket.StatusToVerify
	if err := ticket.WriteMeta(metaPath, m); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "add", metaPath); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("gab: %s to-verify", id)); err != nil {
		return err
	}

	// Best-effort push of the feature branch.
	if gitx.HasRemote(cwd, "origin") {
		if _, err := gitx.Run(cwd, "push", "-u", "origin", m.Branch); err != nil {
			return err
		}
	} else {
		fmt.Printf("gab: no origin remote; skipped push of %s\n", m.Branch)
	}
	return nil
}
```

- [ ] **Step 4: Wire the command in main.go**

Add to the `switch`:
```go
	case "complete":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper complete <id>")
			return 2
		}
		if err := command.Complete(".", args[1]); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		return 0
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/command/ -run TestComplete -v && go build ./...`
Expected: PASS; build succeeds.

- [ ] **Step 6: Commit**

```bash
git add internal/command/complete.go internal/command/complete_test.go cmd/gab-helper/main.go
git commit -m "feat(helper): complete command (summary flow-back, to-verify, push)"
```

---

### Task 9: `done` command — squash-merge, archive, remove worktree

**Files:**
- Create: `internal/command/done.go`
- Test: `internal/command/done_test.go`
- Modify: `cmd/gab-helper/main.go` (wire the `done` case)

**Interfaces:**
- Consumes: `repo`, `gitx`, `ticket`, `TicketDirByID`.
- Produces: `func Done(cwd, id string) error` — run from the main checkout. Squash-merges the feature branch into `main`, discards any `.gab/` changes carried by the branch (main owns `.gab/` truth), commits the code, moves the ticket folder from `tickets/` to `done/` and commits, then removes the worktree and deletes the branch.

Precondition: ticket status must be `to-verify`.

- [ ] **Step 1: Write the failing test**

`internal/command/done_test.go`:
```go
package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestDoneMergesArchivesAndCleansUp(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("feature\n"), 0o644)
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("ok\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "impl")
	Complete(wt, "T1")

	if err := Done(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// code merged into main
	if _, err := os.Stat(filepath.Join(dir, "app.txt")); err != nil {
		t.Fatalf("feature file not merged: %v", err)
	}
	// ticket archived, not active
	if _, _, err := TicketDirByID(r, "T1"); err == nil {
		t.Fatal("ticket should no longer be active")
	}
	if _, err := os.Stat(filepath.Join(r.DoneDir(), "010-T1-login")); err != nil {
		t.Fatalf("ticket not archived to done/: %v", err)
	}
	// main's .gab was not polluted by the branch's BRIEF.md
	if _, err := os.Stat(filepath.Join(dir, ".gab", "BRIEF.md")); !os.IsNotExist(err) {
		t.Fatalf("BRIEF.md leaked into main .gab")
	}
	// worktree + branch removed
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree not removed")
	}
	if _, err := gitx.Run(dir, "rev-parse", "--verify", "gab/T1-login"); err == nil {
		t.Fatal("branch should be deleted")
	}
	_ = ticket.StatusToVerify // keep ticket import used
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/command/ -run TestDone -v`
Expected: FAIL — `undefined: Done`.

- [ ] **Step 3: Write minimal implementation**

`internal/command/done.go`:
```go
package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Done(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, folder, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	m, err := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusToVerify {
		return fmt.Errorf("ticket %s is %q, must be %q for done", id, m.Status, ticket.StatusToVerify)
	}
	if m.Branch == "" {
		return fmt.Errorf("ticket %s has no branch recorded", id)
	}

	// Squash-merge the branch into main (staged, not committed).
	if _, err := gitx.Run(r.Main, "merge", "--squash", m.Branch); err != nil {
		return err
	}
	// main owns .gab truth: drop any .gab changes the branch carried (e.g. BRIEF.md).
	if _, err := gitx.Run(r.Main, "reset", "-q", "--", ".gab"); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "checkout", "--", ".gab"); err != nil {
		return err
	}
	// Remove any brief file that was newly added by the branch (untracked after reset).
	_ = os.Remove(filepath.Join(r.GabDir(), "BRIEF.md"))
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("feat: %s (%s)", m.Title, id)); err != nil {
		return err
	}

	// Archive the ticket folder to done/.
	dest := filepath.Join(r.DoneDir(), folder.Name)
	if _, err := gitx.Run(r.Main, "mv", tdir, dest); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("chore(gab): archive %s", id)); err != nil {
		return err
	}

	// Remove the worktree and delete the branch.
	wt := r.WorktreePath(folder.ID, folder.Slug)
	if _, err := gitx.Run(r.Main, "worktree", "remove", "--force", wt); err != nil {
		return err
	}
	if _, err := gitx.Run(r.Main, "branch", "-D", m.Branch); err != nil {
		return err
	}
	return nil
}
```

- [ ] **Step 4: Wire the command in main.go**

Add to the `switch`:
```go
	case "done":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: gab-helper done <id>")
			return 2
		}
		if err := command.Done(".", args[1]); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		return 0
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/command/ -run TestDone -v && go build ./...`
Expected: PASS; build succeeds.

- [ ] **Step 6: Commit**

```bash
git add internal/command/done.go internal/command/done_test.go cmd/gab-helper/main.go
git commit -m "feat(helper): done command (squash-merge, archive, cleanup)"
```

---

### Task 10: `next` command, full build, and README

**Files:**
- Create: `internal/command/next.go`
- Test: `internal/command/next_test.go`
- Modify: `cmd/gab-helper/main.go` (wire the `next` case)
- Create: `helper/README.md`

**Interfaces:**
- Consumes: `repo`, `backlog`.
- Produces: `func Next(cwd string) (id string, blocked []string, err error)` — returns the next ready ticket id, or `""` with a `blocked` explanation. Cycle → error.
- The `next` case prints the id to stdout and exits 0 when one is ready; prints the blocked explanation to stderr and exits 3 when nothing is ready; exits 1 on error (e.g. cycle).

- [ ] **Step 1: Write the failing test**

`internal/command/next_test.go`:
```go
package command

import (
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestNextReturnsReadyID(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login") // planned, no deps → ready
	id, blocked, err := Next(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "T1" {
		t.Fatalf("id = %q, blocked = %v", id, blocked)
	}
}

func TestNextBlockedByDependency(t *testing.T) {
	dir := testutil.InitRepo(t)
	tdir := seedPlanned(t, dir, "T1", "login")
	// add a dependency on a non-existent/undone ticket
	metaPath := filepath.Join(tdir, "meta.yml")
	m, _ := ticket.ReadMeta(metaPath)
	m.DependsOn = []string{"T9"}
	ticket.WriteMeta(metaPath, m)

	id, blocked, err := Next(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "" || len(blocked) == 0 {
		t.Fatalf("expected blocked, got id=%q blocked=%v", id, blocked)
	}
	_ = repo.Repo{} // keep repo import used
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/command/ -run TestNext -v`
Expected: FAIL — `undefined: Next`.

- [ ] **Step 3: Write minimal implementation**

`internal/command/next.go`:
```go
package command

import (
	"github.com/Ownii/gitops-agent-backlog/internal/backlog"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
)

func Next(cwd string) (string, []string, error) {
	r, err := repo.Discover(cwd)
	if err != nil {
		return "", nil, err
	}
	active, doneIDs, err := backlog.Load(r)
	if err != nil {
		return "", nil, err
	}
	chosen, blocked, err := backlog.Next(active, doneIDs)
	if err != nil {
		return "", nil, err
	}
	if chosen == nil {
		return "", blocked, nil
	}
	return chosen.Meta.ID, blocked, nil
}
```

- [ ] **Step 4: Wire the command in main.go**

Add to the `switch`:
```go
	case "next":
		id, blocked, err := command.Next(".")
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if id == "" {
			fmt.Fprintln(stderr, "no ready ticket")
			for _, b := range blocked {
				fmt.Fprintln(stderr, "  -", b)
			}
			return 3
		}
		fmt.Fprintln(stdout, id)
		return 0
```

- [ ] **Step 5: Run the full test suite and build**

Run: `go test ./... && go build -o bin/gab-helper ./cmd/gab-helper`
Expected: all packages PASS; `bin/gab-helper` produced.

- [ ] **Step 6: Smoke-test the binary end-to-end**

Run (in a throwaway git repo):
```bash
tmp=$(mktemp -d); cd "$tmp"; git init -b main -q; git commit --allow-empty -m init -q
/Library/Repos/Privat/gitops-agent-backlog/bin/gab-helper new login
/Library/Repos/Privat/gitops-agent-backlog/bin/gab-helper next   # prints nothing ready (status todo) → exit 3
```
Expected: `new` prints the folder path `…/.gab/tickets/010-T1-login`; `next` prints "no ready ticket" (the ticket is `todo`, not `planned`).

- [ ] **Step 7: Write `helper/README.md`**

`helper/README.md`:
```markdown
# gab-helper

The deterministic core of `gab`: a small Go CLI that owns the git/filesystem
state of a `.gab/` backlog. It contains no product judgement — it scaffolds,
moves, and commits files and runs git so an agent doesn't have to do those
error-prone steps freehand.

## Build

    go build -o bin/gab-helper ./cmd/gab-helper

## Commands

    gab-helper new <slug>     scaffold a ticket folder (status: todo)
    gab-helper start <id>     create worktree + brief, set in-progress
    gab-helper complete <id>  flow summary back to main, set to-verify, push
    gab-helper done <id>      squash-merge, archive to done/, remove worktree
    gab-helper next           print the id of the next ready ticket

Exit codes: 0 success · 1 error · 2 usage · 3 next found nothing ready.
```

- [ ] **Step 8: Add a `.gitignore` for the build output**

Append to `.gitignore` (create if absent):
```
/bin/
/.gab-worktrees/
```

- [ ] **Step 9: Commit**

```bash
git add internal/command/next.go internal/command/next_test.go cmd/gab-helper/main.go helper/README.md .gitignore
git commit -m "feat(helper): next command, full build, and helper README"
```

---

## Self-Review

**1. Spec coverage** (against `docs/superpowers/specs/2026-07-04-gab-core-design.md`):
- §3 Storage layout (`.gab/`, `tickets/`, `done/`, `definition-of-done.md`) → Task 6 `EnsureGab`. ✓
- §4 Ticket schema (`meta.yml`, folder name, `spec.md`) → Tasks 4, 6. ✓ (`plan.md`/`summary.md` are written by the agent; the helper only reads/relocates them — Tasks 7, 8.)
- §5 Ordering & selection (rank scan, planned + deps-in-`done/`) → Task 5. ✓
- §6 Lifecycle transitions (`new`→todo, `start`→in-progress, `complete`→to-verify, `done`→archived) → Tasks 6–9. ✓ (`plan` is agent-only per §8, correctly no helper verb.)
- §7 Git mechanic (statusless brief committed on branch, truth on main, `.gab` dropped on squash-merge) → Tasks 7, 9. ✓
- §8 Helper scope = exactly the 5 verbs, no content generation → Tasks 6–10. ✓
- §14 Go + `exec.Command` git + `yaml.v3` → Global Constraints, Task 2. ✓
- §13 Edge cases: `next` nothing-ready message → Task 10; cycle detection → Task 5. ✓ Parallel worktrees fall out of per-ticket worktree paths (`WorktreePath`). ✓

**2. Placeholder scan:** No "TBD"/"handle errors appropriately"/"similar to Task N". Every code step shows complete code; every test step shows real assertions and exact `go test` commands with expected PASS/FAIL. ✓

**3. Type consistency:** `Meta`, `Folder`, `Ticket` structs and the `Status*` constants are defined once (Tasks 4/5) and referenced with the same field/method names throughout. Command signatures are consistent: `New(cwd, slug)`, `Start(cwd, id)`, `Complete(cwd, id)`, `Done(cwd, id)`, `Next(cwd)`. `repo.Repo.WorktreePath(id, slug)` is used identically in Tasks 7, 8, 9. `gitx.Run(dir, args...)` / `gitx.HasRemote(dir, name)` signatures match all call sites. ✓

Note carried to Plan 2 (adapter): the branch's inherited stale `.gab/tickets/<id>/meta.yml` is never written in the worktree and is discarded on `done`'s squash-merge, so no divergence reaches `main`. The agent writes its running notes to `<worktree>/.gab/SUMMARY.md` (not the ticket folder); the Claude Code skills in Plan 2 must instruct exactly that path.
