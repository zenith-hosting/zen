package zen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type Renderer struct {
	config   Config
	ssr      ssrClient
	manifest viteManifest
}

type RenderOption func(*renderOptions)

type renderOptions struct {
	Title        string
	Status       int
	Base         []headElement
	Meta         []headElement
	Link         []headElement
	Script       []headElement
	Style        string
	InlineStyles bool
}

func WithTitle(title string) RenderOption {
	return func(opts *renderOptions) {
		opts.Title = title
	}
}

// WithInlineStyles emits this render's stylesheet(s) inline in a <style> tag rather
// than as a render-blocking <link>, collapsing the CSS request out of the critical
// path. Production only: in dev the CSS is injected by the Vite client, so a dev
// render ignores this and keeps the normal dev asset injection.
func WithInlineStyles() RenderOption {
	return func(opts *renderOptions) {
		opts.InlineStyles = true
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

	// Inline the compiled CSS into a <style> instead of a <link> when asked (prod
	// only) — reads the same manifest-resolved files zen would otherwise link, so
	// there's no extra render-blocking CSS request on the critical path.
	var inlineCSS string
	if opts.InlineStyles && !r.config.Dev {
		inlineCSS, err = r.readStyles(assets.Styles)
		if err != nil {
			return err
		}
		assets.Styles = nil
	}

	template, err := r.documentTemplate()
	if err != nil {
		return err
	}

	doc, err := renderDocumentTemplate(template, documentInput{
		Title:         opts.Title,
		Head:          res.Head,
		Base:          opts.Base,
		Meta:          opts.Meta,
		Link:          opts.Link,
		Script:        opts.Script,
		Style:         opts.Style,
		DataElementID: r.config.DataElementID,
		HTML:          res.HTML,
		HydrationJSON: hydrationJSON,
		Styles:        assets.Styles,
		InlineCSS:     inlineCSS,
		Scripts:       assets.Scripts,
		DevScripts:    devScripts,
	})
	if err != nil {
		if strings.TrimSpace(r.config.DocumentPath) == "" {
			return err
		}
		return fmt.Errorf("zen: invalid document template %s: %w", r.config.DocumentPath, err)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.Status(opts.Status).SendString(doc)
}

func (r *Renderer) documentTemplate() (string, error) {
	if strings.TrimSpace(r.config.DocumentPath) == "" {
		return defaultDocumentTemplate, nil
	}

	raw, err := os.ReadFile(r.config.DocumentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("zen: missing document template %s; restore it or set Config.DocumentPath", r.config.DocumentPath)
		}
		return "", fmt.Errorf("zen: read document template %s: %w", r.config.DocumentPath, err)
	}

	return string(raw), nil
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

// readStyles reads the given manifest CSS asset paths (e.g. "/assets/app.<hash>.css")
// from the production client-dist directory and returns their concatenated contents,
// for inlining into a <style> tag. The paths are the same href values zen otherwise
// emits as <link rel="stylesheet">.
func (r *Renderer) readStyles(styles []string) (string, error) {
	var b strings.Builder

	for _, href := range styles {
		rel := strings.TrimPrefix(href, "/")
		raw, err := os.ReadFile(filepath.Join(r.config.clientDist, rel))
		if err != nil {
			return "", fmt.Errorf("zen: read inline stylesheet %s: %w", href, err)
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.Write(raw)
	}

	return b.String(), nil
}

func (r *Renderer) Close() error {
	return nil
}
