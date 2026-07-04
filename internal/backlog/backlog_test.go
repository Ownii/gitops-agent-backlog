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
