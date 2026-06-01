package zencli

import "strings"

type Mode string

const (
	ModeDev  Mode = "dev"
	ModeProd Mode = "prod"
)

func modeFromEnv(env []string) Mode {
	value := ""

	for _, item := range env {
		if after, ok := strings.CutPrefix(item, "ZEN_ENV="); ok {
			value = after
			break
		}
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "prod", "production":
		return ModeProd
	default:
		return ModeDev
	}
}
