package zencli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrefixLines(t *testing.T) {
	var out bytes.Buffer

	err := prefixLines(strings.NewReader("one\ntwo\n"), &out, "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	want := "[app] one\n[app] two\n"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
