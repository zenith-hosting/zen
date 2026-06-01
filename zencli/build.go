package zencli

func buildPlan(cfg Config) []ProcessCommand {
	return []ProcessCommand{
		frontendBuildCommand(cfg),
		goBuildCommand(cfg),
	}
}
