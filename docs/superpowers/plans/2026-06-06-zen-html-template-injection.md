# Zen HTML Template Injection Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let Zen assemble full page responses by injecting SSR output, hydration data, renderer head content, and Vite assets into the developer-owned `frontend/index.html` document template.

**Architecture:** Keep the HTTP renderer protocol unchanged: the Node renderer still returns `ssrResponse{HTML, Head}` and Go still sends the final Fiber response. Replace the current hard-coded full document builder in `document.go` with a slot-based renderer that reads `Config.DocumentPath`, while leaving `RenderIsland` on its existing fragment path.

**Tech Stack:** Go, Fiber, Vite, Preact, Tailwind, Node HTTP renderer, Zen CLI starter templates.

---

I'm using the writing-plans skill to create the implementation plan.

**Plan save path:** `docs/superpowers/plans/2026-06-06-zen-html-template-injection.md`

## Current Codebase State

This plan is based on the repository state on 2026-06-06.

- Root library files are at repo root, not under a `zen/` package directory.
- `render.go` already has `Render`, `RenderPage`, and `RenderIsland`.
- `RenderPage` currently calls `renderDocument(documentInput{...})` after receiving `ssrResponse{HTML, Head}`.
- `RenderIsland` currently calls `renderIslandFragment(islandFragmentInput{...})`; this feature should not turn islands into full documents.
- `document.go` currently builds full HTML with `strings.Builder`; it has no template loader or slot renderer.
- `documentInput` currently has `Title`, `AppElementID`, `DataElementID`, `HTML`, `HydrationJSON`, `Styles`, `Scripts`, and `DevScripts`; it does not pass `Head` through yet.
- `Config` currently has `ProjectRoot`, `AppElementID`, `DataElementID`, `DefaultTitle`, private `clientDist`, and private `manifest`; it does not have a document template path.
- `zen.config.json` supports `frontendDir`, `devRendererPort`, and `prodRendererPort`; use `frontendDir` to default the template path.
- Starter files are embedded from `zencli/internal/zencli/init_template`.
- There are two checked-in examples with Vite `index.html` files: `examples/basic/frontend/index.html` and `examples/todo/frontend/index.html`.

## Scope Check

This feature has one subsystem: full-page document assembly. It touches root renderer configuration, root document rendering, the Fiber page render path, and starter/example `index.html` files.

This plan does **not** implement:

- a metadata DSL
- a route system
- Vite `transformIndexHtml`
- HTML parsing or mutation of arbitrary tags
- custom page layouts
- renderer protocol changes
- island document templating
- generated frontend types

## File Structure

```text
.
  config.go
  config_test.go

  document.go
  document_test.go

  render.go
  render_test.go

  examples/
    basic/frontend/index.html
    todo/frontend/index.html

  zencli/
    internal/zencli/
      init_templates_test.go
      init_template/frontend/index.html
```

## Responsibility Map

| File | Responsibility |
|---|---|
| `config.go` | Add public `DocumentPath` and default it from `ProjectRoot` + `frontendDir` + `index.html`. |
| `document.go` | Replace hard-coded page document assembly with required slot replacement while keeping island fragment rendering. |
| `render.go` | Load the configured template for `RenderPage`, pass `ssrResponse.Head`, and return actionable template errors. |
| `examples/basic/frontend/index.html` | Demonstrate the slot convention in an existing app. |
| `examples/todo/frontend/index.html` | Keep the todo example compatible with the new renderer. |
| `zencli/internal/zencli/init_template/frontend/index.html` | Ensure `zen init` creates a runnable template-based app. |

## Template Slots

Use these exact required slots in `frontend/index.html`:

```html
<!--zen:title-->
<!--zen:head-->
<!--zen:styles-->
<!--zen:app-->
<!--zen:data-->
<!--zen:scripts-->
```

| Slot | Injected content |
|---|---|
| `<!--zen:title-->` | Escaped page title text from `WithTitle` or `Config.DefaultTitle`. |
| `<!--zen:head-->` | Raw `ssrResponse.Head` returned by the Node renderer. |
| `<!--zen:styles-->` | Production CSS `<link>` tags from the current Vite manifest asset model. Empty in dev unless styles are present. |
| `<!--zen:app-->` | Raw SSR HTML from `ssrResponse.HTML`. |
| `<!--zen:data-->` | Go-generated safe hydration JSON `<script>` using `Config.DataElementID`. |
| `<!--zen:scripts-->` | Dev Vite module scripts or production client entry scripts. |

The template should keep the current frontend contract:

```html
<div id="app"><!--zen:app--></div>
```

The current `entry-client.tsx` hydrates `document.getElementById("app")` and reads `document.getElementById("__ZEN_DATA__")`. Do not invent a new template API for those IDs in this feature.

---

## Chunk 1: Configuration And Template Rendering

### Task 1: Add `Config.DocumentPath`

**Files:**
- Modify: `config.go`
- Modify: `config_test.go`

- [ ] **Step 1: Write failing config tests**

Update `TestConfigWithDefaultsDevUsesZenConfigDefaults` in `config_test.go` to assert:

```go
if got.DocumentPath != filepath.Join(".", "frontend", "index.html") {
	t.Fatalf("expected default document path, got %q", got.DocumentPath)
}
```

Update `TestConfigWithDefaultsUsesZenConfigJSON` to assert both dev and prod use the configured frontend directory:

```go
if dev.DocumentPath != filepath.Join(dir, "web", "index.html") {
	t.Fatalf("expected dev document path from frontend dir, got %q", dev.DocumentPath)
}
if prod.DocumentPath != filepath.Join(dir, "web", "index.html") {
	t.Fatalf("expected prod document path from frontend dir, got %q", prod.DocumentPath)
}
```

Add a new override test:

```go
func TestConfigWithDefaultsKeepsDocumentPathOverride(t *testing.T) {
	cfg := Config{
		Dev:          true,
		DocumentPath: "custom/document.html",
	}

	got := cfg.withDefaults()

	if got.DocumentPath != "custom/document.html" {
		t.Fatalf("expected document path override, got %q", got.DocumentPath)
	}
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run:

```bash
go test ./... -run 'TestConfigWithDefaults' -v
```

Expected: FAIL because `Config.DocumentPath` is undefined.

- [ ] **Step 3: Implement the config field and default**

In `config.go`, add the public field near `ProjectRoot`:

```go
DocumentPath string
```

In `withDefaults`, after `frontendDir := filepath.Join(root, project.FrontendDir)`, add:

```go
if c.DocumentPath == "" {
	c.DocumentPath = filepath.Join(frontendDir, "index.html")
}
```

Do not add a new `zen.config.json` key in this task. `frontendDir` already gives the project-level default, and the public Go config field gives tests and advanced apps an explicit override.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run:

```bash
go test ./... -run 'TestConfigWithDefaults' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add config.go config_test.go
git commit -m "feat: configure document template path"
```

### Task 2: Replace Full Document Builder With Slot Renderer

**Files:**
- Modify: `document.go`
- Modify: `document_test.go`

- [ ] **Step 1: Replace document tests with template behavior**

Keep the package as `zen`. Replace the current full-document tests in `document_test.go` with tests for slot rendering:

```go
func TestRenderDocumentTemplateInjectsSlots(t *testing.T) {
	template := `<!doctype html>
<html lang="en">
<head>
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:styles-->
</head>
<body class="font-sans">
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>`

	doc, err := renderDocumentTemplate(template, documentInput{
		Title:         "Home <unsafe>",
		Head:          `<meta name="description" content="Hello">`,
		DataElementID: "__ZEN_DATA__",
		HTML:          `<main><h1>Hello</h1></main>`,
		HydrationJSON: `{"page":"Home","props":{"title":"Hello"}}`,
		Styles:        []string{"/assets/app.css"},
		Scripts:       []string{"/assets/entry-client.js"},
		DevScripts:    []string{"http://localhost:5173/@vite/client"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		`<title>Home &lt;unsafe&gt;</title>`,
		`<meta name="description" content="Hello">`,
		`<link rel="stylesheet" href="/assets/app.css">`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`<script id="__ZEN_DATA__" type="application/json">{"page":"Home","props":{"title":"Hello"}}</script>`,
		`<script type="module" src="http://localhost:5173/@vite/client"></script>`,
		`<script type="module" src="/assets/entry-client.js"></script>`,
		`<body class="font-sans">`,
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("document missing %q\n%s", want, doc)
		}
	}
}
```

Add a required-slot test:

```go
func TestRenderDocumentTemplateRequiresSlots(t *testing.T) {
	template := `<!doctype html><html><body><!--zen:app--></body></html>`

	_, err := renderDocumentTemplate(template, documentInput{
		Title:         "Home",
		DataElementID: "__ZEN_DATA__",
		HTML:          `<main></main>`,
		HydrationJSON: `{}`,
	})
	if err == nil {
		t.Fatal("expected missing slot error")
	}

	if !strings.Contains(err.Error(), `missing required document slot <!--zen:title-->`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

Add a default-template guard:

```go
func TestDefaultDocumentTemplateContainsRequiredSlots(t *testing.T) {
	for _, slot := range requiredDocumentSlots() {
		if !strings.Contains(defaultDocumentTemplate, slot) {
			t.Fatalf("default template missing slot %s", slot)
		}
	}
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run:

```bash
go test ./... -run 'TestRenderDocumentTemplate|TestDefaultDocumentTemplate' -v
```

Expected: FAIL because `renderDocumentTemplate`, `requiredDocumentSlots`, `defaultDocumentTemplate`, and `documentInput.Head` are not implemented.

- [ ] **Step 3: Implement slot constants and helpers**

In `document.go`, add `fmt` to imports and keep `html` and `strings`.

Add slot constants:

```go
const (
	documentSlotTitle   = "<!--zen:title-->"
	documentSlotHead    = "<!--zen:head-->"
	documentSlotStyles  = "<!--zen:styles-->"
	documentSlotApp     = "<!--zen:app-->"
	documentSlotData    = "<!--zen:data-->"
	documentSlotScripts = "<!--zen:scripts-->"
)
```

Add `Head string` to `documentInput`.

Add `defaultDocumentTemplate` using the same slot convention:

```go
const defaultDocumentTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:styles-->
</head>
<body>
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>
`
```

Add:

```go
func requiredDocumentSlots() []string {
	return []string{
		documentSlotTitle,
		documentSlotHead,
		documentSlotStyles,
		documentSlotApp,
		documentSlotData,
		documentSlotScripts,
	}
}
```

- [ ] **Step 4: Implement `renderDocumentTemplate`**

Add:

```go
func renderDocumentTemplate(template string, input documentInput) (string, error) {
	for _, slot := range requiredDocumentSlots() {
		if !strings.Contains(template, slot) {
			return "", fmt.Errorf("zen: missing required document slot %s", slot)
		}
	}

	replacements := map[string]string{
		documentSlotTitle:   html.EscapeString(input.Title),
		documentSlotHead:    input.Head,
		documentSlotStyles:  stylesheetTags(input.Styles),
		documentSlotApp:     input.HTML,
		documentSlotData:    hydrationDataScript(input.DataElementID, input.HydrationJSON),
		documentSlotScripts: scriptTags(append(append([]string{}, input.DevScripts...), input.Scripts...)),
	}

	out := template
	for slot, value := range replacements {
		out = strings.ReplaceAll(out, slot, value)
	}

	return out, nil
}
```

Extract current duplicated tag generation into helpers:

```go
func stylesheetTags(styles []string) string {
	var b strings.Builder
	for _, href := range styles {
		b.WriteString(`<link rel="stylesheet" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`">` + "\n")
	}
	return b.String()
}

func scriptTags(scripts []string) string {
	var b strings.Builder
	for _, src := range scripts {
		b.WriteString(`<script type="module" src="`)
		b.WriteString(html.EscapeString(src))
		b.WriteString(`"></script>` + "\n")
	}
	return b.String()
}

func hydrationDataScript(id string, json string) string {
	var b strings.Builder
	b.WriteString(`<script id="`)
	b.WriteString(html.EscapeString(id))
	b.WriteString(`" type="application/json">`)
	b.WriteString(json)
	b.WriteString("</script>")
	return b.String()
}
```

Keep a compatibility wrapper for tests or manually constructed renderers that still call `renderDocument`:

```go
func renderDocument(input documentInput) string {
	doc, err := renderDocumentTemplate(defaultDocumentTemplate, input)
	if err != nil {
		panic(err)
	}
	return doc
}
```

Update `renderIslandFragment` to use `stylesheetTags` and `scriptTags` for styles/scripts instead of duplicating loops. Do not change the island fragment shape.

- [ ] **Step 5: Run the focused tests and verify they pass**

Run:

```bash
go test ./... -run 'TestRenderDocumentTemplate|TestDefaultDocumentTemplate|TestRenderIsland' -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add document.go document_test.go
git commit -m "feat: render documents from template slots"
```

---

## Chunk 2: Page Render Integration

### Task 3: Load `frontend/index.html` In `RenderPage`

**Files:**
- Modify: `render.go`
- Modify: `render_test.go`

- [ ] **Step 1: Write failing render tests**

Add this test to `render_test.go`:

```go
func TestRenderPageUsesConfiguredDocumentTemplate(t *testing.T) {
	dir := t.TempDir()
	documentPath := filepath.Join(dir, "index.html")
	template := `<!doctype html>
<html lang="en">
<head>
<title><!--zen:title--></title>
<!--zen:head-->
<!--zen:styles-->
</head>
<body class="custom-shell">
<div id="app"><!--zen:app--></div>
<!--zen:data-->
<!--zen:scripts-->
</body>
</html>`

	if err := os.WriteFile(documentPath, []byte(template), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &fakeSSRClient{
		res: ssrResponse{
			HTML: `<main><h1>Hello</h1></main>`,
			Head: `<meta name="description" content="From renderer">`,
		},
	}

	r := &Renderer{
		config: Config{
			Dev:           true,
			viteURL:       "http://localhost:5173",
			DocumentPath:  documentPath,
			DataElementID: "__ZEN_DATA__",
			DefaultTitle:  "Zen",
		},
		ssr: client,
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return r.RenderPage(c, "Home", map[string]string{"title": "Hello"}, WithTitle("Custom <Title>"))
	})

	res := testutil.PerformRequest(t, app, "GET", "/", "")
	body := testutil.ReadBody(t, res)

	for _, want := range []string{
		`<body class="custom-shell">`,
		`<title>Custom &lt;Title&gt;</title>`,
		`<meta name="description" content="From renderer">`,
		`<div id="app"><main><h1>Hello</h1></main></div>`,
		`http://localhost:5173/@vite/client`,
		`http://localhost:5173/.zen/entries/entry-client.tsx`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\n%s", want, body)
		}
	}
}
```

Add a helper-level missing-template test:

```go
func TestRendererDocumentTemplateReportsMissingFile(t *testing.T) {
	r := &Renderer{
		config: Config{
			DocumentPath: filepath.Join(t.TempDir(), "missing.html"),
		},
	}

	_, err := r.documentTemplate()
	if err == nil {
		t.Fatal("expected missing document template error")
	}
	if !strings.Contains(err.Error(), "zen: missing document template") {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run:

```bash
go test ./... -run 'TestRenderPageUsesConfiguredDocumentTemplate|TestRendererDocumentTemplateReportsMissingFile' -v
```

Expected: FAIL because `RenderPage` does not load `DocumentPath`, does not pass `Head`, and `documentTemplate` does not exist.

- [ ] **Step 3: Implement the template loader**

In `render.go`, add imports:

```go
"fmt"
"os"
"strings"
```

Add a method near `RenderPage`:

```go
func (r *Renderer) documentTemplate() (string, error) {
	if strings.TrimSpace(r.config.DocumentPath) == "" {
		return defaultDocumentTemplate, nil
	}

	raw, err := os.ReadFile(r.config.DocumentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("zen: missing document template %s.\n\nRun:\n  zen init", r.config.DocumentPath)
		}
		return "", fmt.Errorf("zen: read document template %s: %w", r.config.DocumentPath, err)
	}

	return string(raw), nil
}
```

The blank-path fallback keeps existing direct `Renderer{...}` tests and low-level unit tests simple. Real renderers created through `New(Config{...})` will have `DocumentPath` defaulted.

- [ ] **Step 4: Use the template in `RenderPage`**

In `RenderPage`, after `clientEntryAssets()` succeeds and before sending the response:

```go
template, err := r.documentTemplate()
if err != nil {
	return err
}

doc, err := renderDocumentTemplate(template, documentInput{
	Title:         opts.Title,
	Head:          res.Head,
	DataElementID: r.config.DataElementID,
	HTML:          res.HTML,
	HydrationJSON: hydrationJSON,
	Styles:        assets.Styles,
	Scripts:       assets.Scripts,
	DevScripts:    devScripts,
})
if err != nil {
	if strings.TrimSpace(r.config.DocumentPath) == "" {
		return err
	}
	return fmt.Errorf("zen: invalid document template %s: %w", r.config.DocumentPath, err)
}
```

Remove the old `doc := renderDocument(documentInput{...})` call from `RenderPage`.

Do not change `RenderIsland`; it should keep calling `renderIslandFragment`.

- [ ] **Step 5: Run render tests**

Run:

```bash
go test ./... -run 'TestRender' -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add render.go render_test.go
git commit -m "feat: inject page renders into index template"
```

---

## Chunk 3: Starter And Examples

### Task 4: Update `index.html` Templates

**Files:**
- Modify: `zencli/internal/zencli/init_templates_test.go`
- Modify: `zencli/internal/zencli/init_template/frontend/index.html`
- Modify: `examples/basic/frontend/index.html`
- Modify: `examples/todo/frontend/index.html`

- [ ] **Step 1: Write a starter template test**

Add to `zencli/internal/zencli/init_templates_test.go`:

```go
func TestStarterIndexHTMLContainsZenDocumentSlots(t *testing.T) {
	files := starterFiles()
	index := files["frontend/index.html"]

	for _, want := range []string{
		`<!--zen:title-->`,
		`<!--zen:head-->`,
		`<!--zen:styles-->`,
		`<div id="app"><!--zen:app--></div>`,
		`<!--zen:data-->`,
		`<!--zen:scripts-->`,
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("frontend/index.html missing %q\n%s", want, index)
		}
	}
}
```

- [ ] **Step 2: Run the CLI focused test and verify it fails**

Run:

```bash
cd zencli && go test ./internal/zencli -run 'TestStarterIndexHTMLContainsZenDocumentSlots' -v
```

Expected: FAIL because the starter `frontend/index.html` still contains the old Vite-only shell with a direct `.zen/entries/entry-client.tsx` script.

- [ ] **Step 3: Update starter `frontend/index.html`**

Replace `zencli/internal/zencli/init_template/frontend/index.html` with:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title><!--zen:title--></title>
    <!--zen:head-->
    <!--zen:styles-->
  </head>
  <body>
    <div id="app"><!--zen:app--></div>
    <!--zen:data-->
    <!--zen:scripts-->
  </body>
</html>
```

Do not keep the direct script tag. Zen now injects the dev or production client entry through `<!--zen:scripts-->`.

- [ ] **Step 4: Update checked-in examples**

Update both `examples/basic/frontend/index.html` and `examples/todo/frontend/index.html` to the same slot shape. Keep titles appropriate for the example if desired, but the `<title>` contents must be `<!--zen:title-->` so `WithTitle` and `DefaultTitle` work.

- [ ] **Step 5: Run CLI and root tests**

Run:

```bash
go test ./...
cd zencli && go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add zencli/internal/zencli/init_templates_test.go zencli/internal/zencli/init_template/frontend/index.html examples/basic/frontend/index.html examples/todo/frontend/index.html
git commit -m "chore: add zen document slots to templates"
```

---

## Chunk 4: Final Verification

### Task 5: Verify Root, CLI, Renderer, And Build Behavior

**Files:**
- No source changes expected.

- [ ] **Step 1: Run root Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run CLI tests**

Run:

```bash
cd zencli && go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run Node renderer tests**

Run from the repository root:

```bash
node --test js/renderers/*.test.mjs
node --test scripts/*.test.mjs
```

Expected: PASS. This feature should not require renderer source changes.

- [ ] **Step 4: Build the CLI**

Run:

```bash
go build -o ./bin/zen ./zencli/cmd/zen
```

Expected: PASS.

- [ ] **Step 5: Smoke-test production artifact injection**

Run:

```bash
cd examples/basic
../../bin/zen build
../../bin/zen start
```

In a separate shell:

```bash
curl -i http://127.0.0.1:3000/
```

Expected:

- Response is HTML.
- The response preserves the surrounding `frontend/index.html` shell.
- The response contains SSR page HTML.
- The response contains production client script assets from the Vite manifest.
- The response does not contain `/@vite/client`.

- [ ] **Step 6: Smoke-test dev injection**

Run:

```bash
cd examples/basic
../../bin/zen dev
```

In a separate shell:

```bash
curl -i http://127.0.0.1:3000/
```

Expected:

- Response is HTML.
- The response preserves the surrounding `frontend/index.html` shell.
- The response contains SSR page HTML.
- The response contains `http://localhost:5173/@vite/client`.
- The response contains `http://localhost:5173/.zen/entries/entry-client.tsx`.

- [ ] **Step 7: Commit final fixes if verification exposed any**

Only commit if Step 1-6 required fixes:

```bash
git add <changed-files>
git commit -m "test: verify document template injection"
```

## Implementation Notes

- `ssrResponse.Head` already exists in `ssr_client.go`; use it. Do not modify the HTTP renderer protocol.
- `serializeHydrationData` already escapes script-breakout data with `json.Encoder.SetEscapeHTML(true)`; keep using it.
- `renderDocumentTemplate` should escape only values Zen owns as text or attributes: title, style hrefs, script srcs, and data script ID. It should not escape SSR HTML or renderer head HTML.
- `Config.AppElementID` exists today, but `entry-client.tsx` currently hydrates the literal `app` ID. This plan keeps the starter and examples on `<div id="app">` and does not broaden that contract.
- Reading the template per page render is acceptable for this feature. It keeps dev edits to `frontend/index.html` visible without adding cache invalidation machinery.
- Slot errors should name the missing slot and, when `DocumentPath` is known, the template file.

Plan complete and saved to `docs/superpowers/plans/2026-06-06-zen-html-template-injection.md`. Ready to execute?
