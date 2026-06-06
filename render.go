package zen

import (
	"context"
	"errors"

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
		ssr: newHTTPSSRClient(httpSSRClientConfig{
			RenderURL: cfg.renderURL,
			Timeout:   cfg.RenderTimeout,
		}),
	}

	if !cfg.Dev {
		manifest, err := readManifest(cfg.manifest)
		if err != nil {
			return nil, err
		}
		r.manifest = manifest
	}

	return r, nil
}

func (r *Renderer) Render(c fiber.Ctx, page string, props any, options ...RenderOption) error {
	return r.RenderPage(c, page, props, options...)
}

func (r *Renderer) RenderPage(c fiber.Ctx, page string, props any, options ...RenderOption) error {
	opts := renderOptions{
		Title:  r.config.DefaultTitle,
		Status: fiber.StatusOK,
	}

	for _, option := range options {
		option(&opts)
	}

	if r.ssr == nil {
		return errors.New("zen: renderer has no SSR client; configure RenderURL or inject an SSR client in tests")
	}

	ctx := context.Background()
	if userCtx := c.Context(); userCtx != nil {
		ctx = userCtx
	}

	res, err := r.ssr.Render(ctx, ssrRequest{
		Mode:  "page",
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

	assets, devScripts, err := r.clientEntryAssets()
	if err != nil {
		return err
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

func (r *Renderer) RenderIsland(c fiber.Ctx, island string, props any, options ...RenderOption) error {
	opts := renderOptions{
		Status: fiber.StatusOK,
	}

	for _, option := range options {
		option(&opts)
	}

	if r.ssr == nil {
		return errors.New("zen: renderer has no SSR client; configure RenderURL or inject an SSR client in tests")
	}

	ctx := context.Background()
	if userCtx := c.Context(); userCtx != nil {
		ctx = userCtx
	}

	res, err := r.ssr.Render(ctx, ssrRequest{
		Mode:   "island",
		URL:    c.OriginalURL(),
		Island: island,
		Props:  props,
	})
	if err != nil {
		return err
	}

	hydrationJSON, err := serializeHydrationData(hydrationData{
		Island: island,
		Props:  props,
	})
	if err != nil {
		return err
	}

	assets, devScripts, err := r.clientEntryAssets()
	if err != nil {
		return err
	}

	fragment := renderIslandFragment(islandFragmentInput{
		Island:        island,
		HTML:          res.HTML,
		HydrationJSON: hydrationJSON,
		Styles:        assets.Styles,
		Scripts:       assets.Scripts,
		DevScripts:    devScripts,
	})

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.Status(opts.Status).SendString(fragment)
}

func (r *Renderer) clientEntryAssets() (clientAssets, []string, error) {
	if r.config.Dev {
		return clientAssets{}, []string{
			r.config.viteURL + "/@vite/client",
			r.config.viteURL + "/.zen/entries/entry-client.tsx",
		}, nil
	}

	assets, err := r.manifest.clientAssets(".zen/entries/entry-client.tsx")
	if err != nil {
		return clientAssets{}, nil, err
	}

	return assets, nil, nil
}

func (r *Renderer) Close() error {
	return nil
}
