package zencli

import "testing"

func TestModeFromEnvDefaultsToDev(t *testing.T) {
	got := modeFromEnv([]string{})

	if got != ModeDev {
		t.Fatalf("expected dev, got %q", got)
	}
}

func TestModeFromEnvAcceptsProductionNames(t *testing.T) {
	for _, value := range []string{"prod", "production"} {
		got := modeFromEnv([]string{"ZEN_ENV=" + value})

		if got != ModeProd {
			t.Fatalf("expected prod for %q, got %q", value, got)
		}
	}
}

func TestModeFromEnvAcceptsDevNames(t *testing.T) {
	for _, value := range []string{"dev", "development"} {
		got := modeFromEnv([]string{"ZEN_ENV=" + value})

		if got != ModeDev {
			t.Fatalf("expected dev for %q, got %q", value, got)
		}
	}
}
