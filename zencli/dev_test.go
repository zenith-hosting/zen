package zencli

import "testing"

func TestDevPlanUsesOnlyDevRendererAndAir(t *testing.T) {
	cfg := Config{
		AirCommand:      "air -c .air.toml",
		FrontendDir:     "frontend",
		DevRendererPort: 5173,
	}

	plan := devPlan(cfg)

	if len(plan) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(plan))
	}

	if plan[0].Name != "renderer" {
		t.Fatalf("expected renderer first, got %q", plan[0].Name)
	}

	if plan[0].Command.Args[0] != ".zen/renderers/dev-renderer.mjs" {
		t.Fatalf("expected dev renderer, got %#v", plan[0].Command.Args)
	}

	if plan[1].Name != "app" {
		t.Fatalf("expected app second, got %q", plan[1].Name)
	}

	if plan[1].Command.Env[0] != "ZEN_ENV=dev" {
		t.Fatalf("expected dev env, got %#v", plan[1].Command.Env)
	}
}
