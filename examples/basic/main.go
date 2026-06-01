package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/zenith/zen/zen"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())

	dev := os.Getenv("ZEN_ENV") != "prod"

	cfg := zen.Config{
		Dev:           dev,
		ViteURL:       "http://localhost:5173",
		RenderURL:     "http://localhost:5173/__zen/render",
		ClientDist:    "./frontend/dist/client",
		Manifest:      "./frontend/dist/client/.vite/manifest.json",
		DefaultTitle:  "Zen Basic Example",
		RenderTimeout: 5 * time.Second,
	}

	if !dev {
		cfg.RenderURL = "http://127.0.0.1:4174/__zen/render"
	}

	renderer, err := zen.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Close()

	if !dev {
		app.Get("/assets/*", renderer.Static())
	}

	app.Get("/", func(c fiber.Ctx) error {
		return renderer.Render(c, "Home", map[string]any{
			"title": "Zen Basic Example",
			"body":  "Fiber route, Preact page, Vite renderer. No pipe slop.",
		}, zen.WithTitle("Home"))
	})

	app.Get("/users/:id", func(c fiber.Ctx) error {
		id := c.Params("id")

		return renderer.Render(c, "User", map[string]any{
			"id": id,
		}, zen.WithTitle("User "+id))
	})

	app.Post("/contact", func(c fiber.Ctx) error {
		name := c.FormValue("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("name is required")
		}

		return c.Redirect().To("/")
	})

	log.Fatal(app.Listen(":3000"))
}
