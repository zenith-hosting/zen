package zencli

import "testing"

func TestDevPlanUsesDevRendererAndAir(t *testing.T) {
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

	if plan[1].Name != "app" {
		t.Fatalf("expected app second, got %q", plan[1].Name)
	}
}

func TestProdPlanUsesProdRendererAndBinary(t *testing.T) {
	cfg := Config{
		FrontendDir:      "frontend",
		ProdRendererPort: 4174,
		BinaryPath:       "./bin/app",
	}

	plan := prodPlan(cfg)

	if len(plan) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(plan))
	}

	if plan[0].Name != "renderer" {
		t.Fatalf("expected renderer first, got %q", plan[0].Name)
	}

	if plan[1].Name != "app" {
		t.Fatalf("expected app second, got %q", plan[1].Name)
	}
}

func TestPlanForMode(t *testing.T) {
	cfg := Config{
		AirCommand:       "air -c .air.toml",
		FrontendDir:      "frontend",
		DevRendererPort:  5173,
		ProdRendererPort: 4174,
		BinaryPath:       "./bin/app",
	}

	dev := planForMode(ModeDev, cfg)
	if dev[1].Command.Name != "sh" {
		t.Fatalf("expected dev app through shell air command, got %q", dev[1].Command.Name)
	}

	prod := planForMode(ModeProd, cfg)
	if prod[1].Command.Env[0] != "ZEN_ENV=prod" {
		t.Fatalf("expected prod env, got %#v", prod[1].Command.Env)
	}
}

func TestProdStartupRequiresBuildFirst(t *testing.T) {
	cfg := Config{
		FrontendDir: "frontend",
		BinaryPath:  "./bin/app",
	}

	steps := preflightForMode(ModeProd, cfg)

	if len(steps) != 2 {
		t.Fatalf("expected 2 preflight build steps, got %d", len(steps))
	}

	if steps[0].Name != "pnpm" {
		t.Fatalf("expected frontend build first, got %q", steps[0].Name)
	}

	if steps[1].Name != "sh" {
		t.Fatalf("expected Go build second, got %q", steps[1].Name)
	}
}

func TestDevStartupHasNoPreflightBuild(t *testing.T) {
	cfg := Config{}

	steps := preflightForMode(ModeDev, cfg)

	if len(steps) != 0 {
		t.Fatalf("expected no dev preflight, got %d steps", len(steps))
	}
}
