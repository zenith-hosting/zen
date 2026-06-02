package zencli

import "testing"

func TestStartPlanUsesProductionRendererAndBinary(t *testing.T) {
	cfg := Config{
		FrontendDir:      "frontend",
		ProdRendererPort: 4174,
		BinaryPath:       "./bin/app",
	}

	plan := startPlan(cfg)

	if len(plan) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(plan))
	}

	if plan[0].Name != "renderer" {
		t.Fatalf("expected renderer first, got %q", plan[0].Name)
	}

	if plan[0].Command.Args[0] != ".zen/renderers/prod-renderer.mjs" {
		t.Fatalf("expected production renderer, got %#v", plan[0].Command.Args)
	}

	if plan[1].Name != "app" {
		t.Fatalf("expected app second, got %q", plan[1].Name)
	}

	if plan[1].Command.Env[0] != "ZEN_ENV=prod" {
		t.Fatalf("expected prod env, got %#v", plan[1].Command.Env)
	}
}
