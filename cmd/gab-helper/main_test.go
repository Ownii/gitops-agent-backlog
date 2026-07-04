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
