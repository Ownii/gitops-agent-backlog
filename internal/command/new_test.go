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

func TestNewRejectsInvalidSlug(t *testing.T) {
	dir := testutil.InitRepo(t)
	if _, err := New(dir, "My Feature"); err == nil {
		t.Fatal("expected error for invalid slug \"My Feature\"")
	}
	if _, err := New(dir, "Bad_Slug"); err == nil {
		t.Fatal("expected error for invalid slug \"Bad_Slug\"")
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".gab", "tickets"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no ticket folders created, got %v", entries)
	}
}
