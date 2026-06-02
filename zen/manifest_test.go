package zen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadManifestEntryClientAssets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	manifest := `{
		".zen/entries/entry-client.tsx": {
			"file": "assets/entry-client.abc123.js",
			"css": ["assets/app.def456.css"],
			"imports": ["_vendor.js"]
		},
		"_vendor.js": {
			"file": "assets/vendor.999.js"
		}
	}`

	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assets, err := got.clientAssets(".zen/entries/entry-client.tsx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(assets.Scripts) != 2 {
		t.Fatalf("expected 2 scripts, got %#v", assets.Scripts)
	}
	if assets.Scripts[0] != "/assets/vendor.999.js" {
		t.Fatalf("expected vendor import first, got %#v", assets.Scripts)
	}
	if assets.Scripts[1] != "/assets/entry-client.abc123.js" {
		t.Fatalf("expected client script second, got %#v", assets.Scripts)
	}
	if len(assets.Styles) != 1 || assets.Styles[0] != "/assets/app.def456.css" {
		t.Fatalf("expected css asset, got %#v", assets.Styles)
	}
}

func TestManifestMissingEntryReturnsError(t *testing.T) {
	m := viteManifest{}

	_, err := m.clientAssets(".zen/entries/entry-client.tsx")
	if err == nil {
		t.Fatal("expected error")
	}
}
