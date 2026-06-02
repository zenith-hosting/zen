package zencli

import (
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
		"frontend/src/entry-client.tsx",
		"frontend/src/entry-server.tsx",
		"frontend/src/pages/Home.tsx",
		"frontend/src/pages/User.tsx",
	}

	for _, path := range required {
		if _, ok := files[path]; !ok {
			t.Fatalf("starter files missing %s", path)
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
