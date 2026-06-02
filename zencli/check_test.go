package zencli

import "testing"

func TestRequiredTools(t *testing.T) {
	got := requiredTools()

	want := []string{"go", "node", "pnpm"}

	if len(got) != len(want) {
		t.Fatalf("expected %d tools, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected tool %q at index %d, got %q", want[i], i, got[i])
		}
	}
}
