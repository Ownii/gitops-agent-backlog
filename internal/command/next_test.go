package command

import (
	"path/filepath"
	"testing"

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
}
