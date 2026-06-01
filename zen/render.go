package zen

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type Renderer struct {
	config   Config
	ssr      ssrClient
	manifest viteManifest
}

type RenderOption func(*renderOptions)

type renderOptions struct {
	Title  string
	Status int
}

func WithTitle(title string) RenderOption {
	return func(opts *renderOptions) {
		opts.Title = title
	}
}

func WithStatus(status int) RenderOption {
	return func(opts *renderOptions) {
		opts.Status = status
	}
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

		client, err := newProcessSSRClient(cfg.SSRCommand)
		if err != nil {
			return nil, err
		}
		r.ssr = client
	}

	return r, nil
}

func (r *Renderer) Render(c fiber.Ctx, page string, props any, options ...RenderOption) error {
	opts := renderOptions{
		Title:  r.config.DefaultTitle,
		Status: fiber.StatusOK,
	}

	for _, option := range options {
		option(&opts)
	}

	ctx := context.Background()
	if userCtx := c.Context(); userCtx != nil {
		ctx = userCtx
	}

	res, err := r.ssr.Render(ctx, ssrRequest{
		URL:   c.OriginalURL(),
		Page:  page,
		Props: props,
	})
	if err != nil {
		return err
	}

	hydrationJSON, err := serializeHydrationData(hydrationData{
		Page:  page,
		Props: props,
	})
	if err != nil {
		return err
	}

	assets := clientAssets{}
	devScripts := []string(nil)

	if r.config.Dev {
		devScripts = []string{
			r.config.ViteURL + "/@vite/client",
			r.config.ViteURL + "/src/entry-client.tsx",
		}
	} else {
		assets, err = r.manifest.clientAssets("src/entry-client.tsx")
		if err != nil {
			return err
		}
	}

	doc := renderDocument(documentInput{
		Title:         opts.Title,
		AppElementID:  r.config.AppElementID,
		DataElementID: r.config.DataElementID,
		HTML:          res.HTML,
		HydrationJSON: hydrationJSON,
		Styles:        assets.Styles,
		Scripts:       assets.Scripts,
		DevScripts:    devScripts,
	})

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.Status(opts.Status).SendString(doc)
}
