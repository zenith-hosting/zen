package zen

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const (
	htmlContentType      = "text/html; charset=utf-8"
	pageIdentifierPrefix = "zen-page-"
)

type Renderer struct {
	config           Config
	ssr              ssrClient
	manifest         viteManifest
	islandIdentifier atomic.Uint64
}

type Response struct {
	Status      int
	ContentType string
	Body        []byte
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

func (r *Renderer) RenderPage(ctx context.Context, url, page string, props any, options ...RenderOption) (Response, error) {
	opts := renderOptions{
		Title:  r.config.DefaultTitle,
		Status: http.StatusOK,
	}

	for _, option := range options {
		option(&opts)
	}

	if r.ssr == nil {
		return Response{}, errors.New("zen: renderer has no SSR client; configure RenderURL or inject an SSR client in tests")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	res, err := r.ssr.Render(ctx, ssrRequest{
		Mode:             "page",
		URL:              url,
		Page:             page,
		IdentifierPrefix: pageIdentifierPrefix,
		Props:            props,
	})
	if err != nil {
		return Response{}, err
	}

	hydrationJSON, err := serializeHydrationData(hydrationData{
		Page:             page,
		IdentifierPrefix: pageIdentifierPrefix,
		Props:            props,
	})
	if err != nil {
		return Response{}, err
	}

	assets, devScripts, err := r.clientEntryAssets()
	if err != nil {
		return Response{}, err
	}

	// Inline the compiled CSS into a <style> instead of a <link> when asked (prod
	// only) — reads the same manifest-resolved files zen would otherwise link, so
	// there's no extra render-blocking CSS request on the critical path.
	var compiledCSS string
	if opts.InlineStyles && !r.config.Dev {
		compiledCSS, err = r.readStyles(assets.Styles)
		if err != nil {
			return Response{}, err
		}
		assets.Styles = nil
	}

	moduleScriptURLs := append(devScripts, assets.Scripts...)

	doc := renderDocument(documentInput{
		Title:            opts.Title,
		HeadHTML:         res.Head,
		BaseElements:     opts.Base,
		MetaElements:     opts.Meta,
		LinkElements:     opts.Link,
		HeadScripts:      opts.Script,
		CustomCSS:        opts.Style,
		AppHTML:          res.HTML,
		HydrationJSON:    hydrationJSON,
		StylesheetURLs:   assets.Styles,
		CompiledCSS:      compiledCSS,
		ModuleScriptURLs: moduleScriptURLs,
	})

	return Response{
		Status:      opts.Status,
		ContentType: htmlContentType,
		Body:        []byte(doc),
	}, nil
}

func (r *Renderer) RenderIsland(ctx context.Context, url, island string, props any, options ...RenderOption) (Response, error) {
	opts := renderOptions{
		Status: http.StatusOK,
	}

	for _, option := range options {
		option(&opts)
	}

	if r.ssr == nil {
		return Response{}, errors.New("zen: renderer has no SSR client; configure RenderURL or inject an SSR client in tests")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	identifierPrefix := fmt.Sprintf("zen-island-%d-", r.islandIdentifier.Add(1))
	res, err := r.ssr.Render(ctx, ssrRequest{
		Mode:             "island",
		URL:              url,
		Island:           island,
		IdentifierPrefix: identifierPrefix,
		Props:            props,
	})
	if err != nil {
		return Response{}, err
	}

	hydrationJSON, err := serializeHydrationData(hydrationData{
		Island:           island,
		IdentifierPrefix: identifierPrefix,
		Props:            props,
	})
	if err != nil {
		return Response{}, err
	}

	fragment := renderIslandFragment(islandFragmentInput{
		Island:        island,
		HTML:          res.HTML,
		HydrationJSON: hydrationJSON,
	})

	return Response{
		Status:      opts.Status,
		ContentType: htmlContentType,
		Body:        []byte(fragment),
	}, nil
}

func (r *Renderer) clientEntryAssets() (clientAssets, []string, error) {
	if r.config.Dev {
		return clientAssets{}, []string{
			r.config.viteURL + "/@vite/client",
			r.config.viteURL + "/.zen/entries/react-refresh.mjs",
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
