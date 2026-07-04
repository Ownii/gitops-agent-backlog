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
