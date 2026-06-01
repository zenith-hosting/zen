package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/zenith/zen/zen"
)

func main() {
	app := fiber.New()

	dev := os.Getenv("ZEN_ENV") != "production"

	cfg := zen.Config{
		Dev:     dev,
		ViteURL: "http://localhost:5173",
		SSRCommand: []string{
			"node",
			"./node_modules/@zenith/zen/js/ssr-worker.mjs",
			"--entry",
			"./frontend/src/entry-server.tsx",
		},
		ClientDist:   "./frontend/dist/client",
		Manifest:     "./frontend/dist/client/.vite/manifest.json",
		DefaultTitle: "Zen Basic Example",
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
			"body":  "Fiber route, Preact page, Vite build. No ceremony.",
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

		return c.Redirect("/")
	})

	log.Fatal(app.Listen(":3000"))
}
