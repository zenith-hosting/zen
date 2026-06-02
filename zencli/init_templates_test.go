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
		"frontend/src/pages.ts",
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

func TestStarterTemplateSourceFilesExist(t *testing.T) {
	for path := range starterFiles() {
		sourcePath := path
		if sourcePath == "go.mod" {
			sourcePath = "go.mod.template"
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
		`github.com/zenith/zen/zen`,
		`RenderURL:`,
		`renderer.Render`,
		`/__zen/render`,
	} {
		if !strings.Contains(main, want) {
			t.Fatalf("main.go missing %q\n%s", want, main)
		}
	}
}

func TestStarterPackageScriptsUseZenCommands(t *testing.T) {
	files := starterFiles()
	pkg := files["package.json"]

	for _, want := range []string{
		`"dev": "zen dev"`,
		`"build": "zen build"`,
		`"start": "zen start"`,
		`"check": "zen check"`,
	} {
		if !strings.Contains(pkg, want) {
			t.Fatalf("package.json missing %q\n%s", want, pkg)
		}
	}
}
