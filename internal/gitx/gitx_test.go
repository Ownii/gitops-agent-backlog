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
