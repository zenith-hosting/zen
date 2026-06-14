package zencli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStarterFilesIncludeRunnableProject(t *testing.T) {
	files := starterFiles()

	required := []string{
		"zen.config.json",
		".air.toml",
		"go.mod",
		"main.go",
		"package.json",
		"frontend/package.json",
		"frontend/tsconfig.json",
		"frontend/vite.config.ts",
		"frontend/index.html",
		"frontend/src/app.css",
		"frontend/src/globals.css",
		"frontend/src/islands/Counter.tsx",
		"frontend/src/pages/App.tsx",
		"frontend/src/pages/Home.tsx",
		"frontend/src/pages/User.tsx",
		"frontend/.zen/entries/entry-client.tsx",
		"frontend/.zen/entries/entry-server.tsx",
		"frontend/.zen/renderers/dev-renderer.mjs",
		"frontend/.zen/renderers/prod-renderer.mjs",
		"frontend/.zen/renderers/renderer-shared.mjs",
	}

	for _, path := range required {
		if _, ok := files[path]; !ok {
			t.Fatalf("starter files missing %s", path)
		}
	}
}

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

func TestStarterZenEntriesDiscoverPagesWithViteGlob(t *testing.T) {
	files := starterFiles()

	for _, path := range []string{
		"frontend/.zen/entries/entry-client.tsx",
		"frontend/.zen/entries/entry-server.tsx",
	} {
		contents := files[path]

		for _, want := range []string{
			`import.meta.glob`,
			`../../src/pages/**/*.tsx`,
		} {
			if !strings.Contains(contents, want) {
				t.Fatalf("%s missing %q\n%s", path, want, contents)
			}
		}
	}
}

func TestStarterZenEntriesDiscoverAndHydrateIslands(t *testing.T) {
	files := starterFiles()
	server := files["frontend/.zen/entries/entry-server.tsx"]
	client := files["frontend/.zen/entries/entry-client.tsx"]

	for _, want := range []string{
		`../../src/islands/**/*.tsx`,
		`request.mode ?? "page"`,
		`mode === "island"`,
		`Unknown island: ${request.island}`,
	} {
		if !strings.Contains(server, want) {
			t.Fatalf("entry-server.tsx missing %q\n%s", want, server)
		}
	}

	for _, want := range []string{
		`../../src/islands/**/*.tsx`,
		`[data-zen-island-root]`,
		`[data-zen-island-props]`,
		`data-zen-island`,
	} {
		if !strings.Contains(client, want) {
			t.Fatalf("entry-client.tsx missing %q\n%s", want, client)
		}
	}
}

func TestStarterTemplateSourceFilesExist(t *testing.T) {
	for path := range starterFiles() {
		sourcePath := path
		for _, templateOutput := range []string{"go.mod", "main.go"} {
			if sourcePath == templateOutput {
				sourcePath = templateOutput + ".template"
			}
		}

		fullPath := filepath.Join("init_template", sourcePath)
		if _, err := os.Stat(fullPath); err != nil {
			t.Fatalf("expected starter source file %s to exist: %v", fullPath, err)
		}
	}
}

func TestStarterMainUsesHTTPRenderer(t *testing.T) {
	files := starterFiles()
	main := files["main.go"]

	for _, want := range []string{
		`github.com/gofiber/fiber/v3`,
		`github.com/zenith-hosting/zen`,
		`renderer.RenderPage`,
		`renderer.RenderIsland`,
	} {
		if !strings.Contains(main, want) {
			t.Fatalf("main.go missing %q\n%s", want, main)
		}
	}

	for _, old := range []string{
		`ViteURL:`,
		`RenderURL:`,
		`ClientDist:`,
		`Manifest:`,
		`/__zen/render`,
	} {
		if strings.Contains(main, old) {
			t.Fatalf("main.go should not hard-code %q when zen.config.json owns renderer settings\n%s", old, main)
		}
	}
}

func TestStarterUsesAppOwnedClientRouter(t *testing.T) {
	files := starterFiles()
	main := files["main.go"]
	pkg := files["frontend/package.json"]
	app := files["frontend/src/pages/App.tsx"]

	for _, want := range []string{
		`"preact-iso": "^2.12.0"`,
	} {
		if !strings.Contains(pkg, want) {
			t.Fatalf("frontend/package.json missing %q\n%s", want, pkg)
		}
	}

	for _, want := range []string{
		`app.Get("/*", func(c fiber.Ctx) error {`,
		`renderer.RenderPage(c, "App", map[string]any{`,
		`"url": c.OriginalURL(),`,
	} {
		if !strings.Contains(main, want) {
			t.Fatalf("main.go missing %q\n%s", want, main)
		}
	}

	for _, old := range []string{
		`renderer.RenderPage(c, "Home"`,
		`renderer.RenderPage(c, "User"`,
	} {
		if strings.Contains(main, old) {
			t.Fatalf("main.go should render the app shell instead of page-specific Fiber routes %q\n%s", old, main)
		}
	}

	for _, want := range []string{
		`import { ErrorBoundary, LocationProvider, Route, Router } from "preact-iso";`,
		`import { locationStub } from "preact-iso/prerender";`,
		`locationStub(props.url ?? "/");`,
		`<Route path="/" component={Home} />`,
		`<Route path="/users/:id" component={User} />`,
		`<Route default component={NotFound} />`,
	} {
		if !strings.Contains(app, want) {
			t.Fatalf("frontend/src/pages/App.tsx missing %q\n%s", want, app)
		}
	}
}

func TestStarterFrontendUsesTailwindCLI(t *testing.T) {
	files := starterFiles()
	pkg := files["frontend/package.json"]
	globals := files["frontend/src/globals.css"]
	appCSS := files["frontend/src/app.css"]
	client := files["frontend/.zen/entries/entry-client.tsx"]

	for _, want := range []string{
		`"tailwind:build": "tailwindcss -i ./src/globals.css -o ./src/app.css"`,
		`"tailwind:watch": "tailwindcss -i ./src/globals.css -o ./src/app.css --watch=always"`,
		`"build": "pnpm tailwind:build && pnpm build:client && pnpm build:server"`,
		`"@tailwindcss/cli": "^4.3.0"`,
	} {
		if !strings.Contains(pkg, want) {
			t.Fatalf("frontend/package.json missing %q\n%s", want, pkg)
		}
	}

	if !strings.Contains(globals, `@import "tailwindcss";`) {
		t.Fatalf("frontend/src/globals.css should be the Tailwind CLI input\n%s", globals)
	}

	if strings.Contains(appCSS, `@import "tailwindcss";`) {
		t.Fatalf("frontend/src/app.css should be generated output, not the Tailwind CLI input\n%s", appCSS)
	}

	if !strings.Contains(appCSS, `tailwindcss v4`) {
		t.Fatalf("frontend/src/app.css should contain generated Tailwind CSS\n%s", appCSS)
	}

	if !strings.Contains(client, `../../src/app.css`) {
		t.Fatalf("entry-client.tsx should import generated app.css\n%s", client)
	}
}

func TestStarterGoModDependsOnPublicZenModule(t *testing.T) {
	files := starterFiles()
	goMod := files["go.mod"]

	if !strings.Contains(goMod, `github.com/zenith-hosting/zen v0.0.4`) {
		t.Fatalf("go.mod missing public zen module version\n%s", goMod)
	}
}

func TestStarterPackageScriptsUseZenCommands(t *testing.T) {
	files := starterFiles()
	pkg := files["package.json"]

	for _, want := range []string{
		`"dev": "zen dev"`,
		`"build": "zen build"`,
		`"start": "zen start"`,
	} {
		if !strings.Contains(pkg, want) {
			t.Fatalf("package.json missing %q\n%s", want, pkg)
		}
	}
}
