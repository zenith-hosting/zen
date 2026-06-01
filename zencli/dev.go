package zencli

type NamedProcessCommand struct {
	Name    string
	Command ProcessCommand
}

func devPlan(cfg Config) []NamedProcessCommand {
	return []NamedProcessCommand{
		{
			Name:    "renderer",
			Command: devRendererCommand(cfg),
		},
		{
			Name:    "app",
			Command: airCommand(cfg),
		},
	}
}

func prodPlan(cfg Config) []NamedProcessCommand {
	return []NamedProcessCommand{
		{
			Name:    "renderer",
			Command: prodRendererCommand(cfg),
		},
		{
			Name:    "app",
			Command: goProdCommand(cfg),
		},
	}
}

func planForMode(mode Mode, cfg Config) []NamedProcessCommand {
	if mode == ModeProd {
		return prodPlan(cfg)
	}

	return devPlan(cfg)
}

func preflightForMode(mode Mode, cfg Config) []ProcessCommand {
	if mode == ModeProd {
		return buildPlan(cfg)
	}

	return nil
}
