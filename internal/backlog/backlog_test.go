package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
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

func TestLoad(t *testing.T) {
	// Initialize a test repo
	dir := testutil.InitRepo(t)
	r, err := repo.Discover(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create .gab/tickets directory structure
	ticketsDir := r.TicketsDir()
	if err := os.MkdirAll(ticketsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create active tickets with different ranks
	t1Dir := filepath.Join(ticketsDir, "010-T1-one")
	if err := os.MkdirAll(t1Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t1Meta := &ticket.Meta{ID: "T1", Title: "First", Status: ticket.StatusPlanned}
	if err := ticket.WriteMeta(filepath.Join(t1Dir, "meta.yml"), t1Meta); err != nil {
		t.Fatal(err)
	}

	t2Dir := filepath.Join(ticketsDir, "020-T2-two")
	if err := os.MkdirAll(t2Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t2Meta := &ticket.Meta{ID: "T2", Title: "Second", Status: ticket.StatusPlanned}
	if err := ticket.WriteMeta(filepath.Join(t2Dir, "meta.yml"), t2Meta); err != nil {
		t.Fatal(err)
	}

	// Create a non-ticket folder (should be ignored)
	notesDir := filepath.Join(ticketsDir, "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .gab/done directory with a done ticket
	doneDir := r.DoneDir()
	if err := os.MkdirAll(doneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t5Dir := filepath.Join(doneDir, "005-T5-old")
	if err := os.MkdirAll(t5Dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Load tickets
	active, done, err := Load(r)
	if err != nil {
		t.Fatal(err)
	}

	// Verify active tickets are sorted by rank ascending
	if len(active) != 2 {
		t.Fatalf("expected 2 active tickets, got %d", len(active))
	}
	if active[0].Folder.ID != "T1" || active[0].Folder.Rank != 10 {
		t.Fatalf("expected first ticket to be T1 (rank 10), got %s (rank %d)", active[0].Folder.ID, active[0].Folder.Rank)
	}
	if active[1].Folder.ID != "T2" || active[1].Folder.Rank != 20 {
		t.Fatalf("expected second ticket to be T2 (rank 20), got %s (rank %d)", active[1].Folder.ID, active[1].Folder.Rank)
	}

	// Verify meta was loaded
	if active[0].Meta.ID != "T1" {
		t.Fatalf("expected T1 meta to be loaded, got %+v", active[0].Meta)
	}

	// Verify done tickets
	if !done["T5"] {
		t.Fatalf("expected T5 to be in done set, got %v", done)
	}
	if len(done) != 1 {
		t.Fatalf("expected 1 done ticket, got %d", len(done))
	}
}

func TestLoadMissingTicketsDir(t *testing.T) {
	// Initialize a test repo
	dir := testutil.InitRepo(t)
	r, err := repo.Discover(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Don't create tickets directory - should handle gracefully
	active, done, err := Load(r)
	if err != nil {
		t.Fatalf("expected Load to handle missing tickets dir, got error: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("expected empty active tickets, got %d", len(active))
	}
	if len(done) != 0 {
		t.Fatalf("expected empty done set, got %v", done)
	}
}

func TestNextFlagsUnknownDependency(t *testing.T) {
	// T2 depends on T4 (a real in-progress ticket) and T99 (neither active nor
	// done — a typo or a deleted ticket). Both leave T2 blocked, but the message
	// must call out the unknown id so it does not read as a normal wait.
	active := []Ticket{
		mk(10, "T4", ticket.StatusInProgress),
		mk(20, "T2", ticket.StatusPlanned, "T4", "T99"),
	}
	chosen, blocked, err := Next(active, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if chosen != nil {
		t.Fatalf("expected nothing ready, got %+v", chosen)
	}
	if len(blocked) != 1 {
		t.Fatalf("expected 1 blocked entry, got %v", blocked)
	}
	if !strings.Contains(blocked[0], "T99") || !strings.Contains(blocked[0], "unknown") {
		t.Fatalf("expected T99 flagged as unknown, got %q", blocked[0])
	}
	if strings.Contains(blocked[0], "T4 (unknown") {
		t.Fatalf("T4 is a known in-progress ticket and must NOT be flagged unknown, got %q", blocked[0])
	}
}
