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

func TestParseFolderAcceptsRankOverflow(t *testing.T) {
	// Ranks grow by +10 per ticket; a long-lived backlog crosses 1000. A
	// 4-digit rank must still parse (sorting is by the parsed int), or the
	// ticket silently vanishes from Load.
	f, err := ParseFolder("1000-T5-login")
	if err != nil {
		t.Fatalf("4-digit rank rejected: %v", err)
	}
	if f.Rank != 1000 || f.ID != "T5" || f.Slug != "login" {
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
