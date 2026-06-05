package zen

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith-hosting/zen/internal/testutil"
)

func TestStaticReturnsFiberHandlerForClientDist(t *testing.T) {
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")

	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &Renderer{
		config: Config{
			ClientDist: dir,
		},
	}

	app := fiber.New()
	app.Get("/assets/*", r.Static())

	res := testutil.PerformRequest(t, app, http.MethodGet, "/assets/app.js", "")
	body := testutil.ReadBody(t, res)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
	if body != "console.log('ok')" {
		t.Fatalf("unexpected body: %s", body)
	}
}
