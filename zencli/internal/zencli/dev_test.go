package zencli

import "testing"

func TestDevPlanRunsTailwindWatchRendererAndAir(t *testing.T) {
	cfg := Config{
		AirCommand:      "go tool air -c .air.toml",
		FrontendDir:     "frontend",
		DevRendererPort: 5173,
	}

	plan := devPlan(cfg)

	if len(plan) != 3 {
		t.Fatalf("expected 3 processes, got %d", len(plan))
	}

	if plan[0].Name != "tailwind" {
		t.Fatalf("expected tailwind first, got %q", plan[0].Name)
	}

	if plan[0].Command.Name != "pnpm" {
		t.Fatalf("expected pnpm tailwind command, got %q", plan[0].Command.Name)
	}

	if plan[0].Command.Dir != "frontend" {
		t.Fatalf("expected tailwind command to run in frontend, got %q", plan[0].Command.Dir)
	}

	if plan[0].Command.Args[0] != "tailwind:watch" {
		t.Fatalf("expected tailwind watch script, got %#v", plan[0].Command.Args)
	}

	if plan[1].Name != "renderer" {
		t.Fatalf("expected renderer second, got %q", plan[1].Name)
	}

	if plan[1].Command.Args[0] != ".zen/renderers/dev-renderer.mjs" {
		t.Fatalf("expected dev renderer, got %#v", plan[1].Command.Args)
	}

	if plan[2].Name != "app" {
		t.Fatalf("expected app third, got %q", plan[2].Name)
	}

	if plan[2].Command.Env[0] != "ZEN_ENV=dev" {
		t.Fatalf("expected dev env, got %#v", plan[2].Command.Env)
	}
}
