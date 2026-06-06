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
		"frontend/src/islands/Counter.tsx",
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

func TestStarterGoModDependsOnPublicZenModule(t *testing.T) {
	files := starterFiles()
	goMod := files["go.mod"]

	if !strings.Contains(goMod, `github.com/zenith-hosting/zen v0.0.1`) {
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
