package zencli

import "testing"

func TestBuildPlanIncludesFrontendThenGoBuild(t *testing.T) {
	cfg := Config{
		FrontendDir: "frontend",
		BinaryPath:  "./bin/app",
	}

	steps := buildPlan(cfg)

	if len(steps) != 2 {
		t.Fatalf("expected 2 build steps, got %d", len(steps))
	}

	if steps[0].Name != "pnpm" {
		t.Fatalf("expected frontend build first, got %q", steps[0].Name)
	}

	if steps[1].Name != "sh" {
		t.Fatalf("expected shell go build second, got %q", steps[1].Name)
	}
}
