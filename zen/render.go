package zen

type Renderer struct {
	config   Config
	ssr      ssrClient
	manifest viteManifest
}

func New(config Config) (*Renderer, error) {
	cfg := config.withDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	r := &Renderer{
		config: cfg,
	}

	if !cfg.Dev {
		manifest, err := readManifest(cfg.Manifest)
		if err != nil {
			return nil, err
		}
		r.manifest = manifest
	}

	return r, nil
}
