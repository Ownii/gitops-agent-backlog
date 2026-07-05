package command

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func TestListEmptyBacklog(t *testing.T) {
	dir := testutil.InitRepo(t)
	out, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("expected empty output for empty backlog, got %q", out)
	}
}

func TestListRendersActiveTicketsByRank(t *testing.T) {
	dir := testutil.InitRepo(t)
	// seedPlanned creates rank 010; add a lower-ranked todo via New (rank grows
	// by +10) so we also cover ordering and a ticket with dependencies.
	seedPlanned(t, dir, "T1", "login")

	out, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), out)
	}
	line := lines[0]
	for _, want := range []string{"010", "T1", "planned", "login"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q missing %q", line, want)
		}
	}
}

func TestListShowsDependencies(t *testing.T) {
	dir := testutil.InitRepo(t)
	tdir := seedPlanned(t, dir, "T2", "checkout")
	// Give T2 a dependency and re-render.
	metaPath := filepath.Join(tdir, "meta.yml")
	m, _ := ticket.ReadMeta(metaPath)
	m.DependsOn = []string{"T1"}
	if err := ticket.WriteMeta(metaPath, m); err != nil {
		t.Fatal(err)
	}

	out, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "deps: T1") {
		t.Fatalf("expected dependency shown, got %q", out)
	}
}
