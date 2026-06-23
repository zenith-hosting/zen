package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/zenith-hosting/zen"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())

	dev := os.Getenv("ZEN_ENV") != "prod"

	port := ":3000"
	if dev {
		port = ":30001"
	}

	cfg := zen.Config{
		Dev:           dev,
		DefaultTitle:  "Zen App",
		RenderTimeout: 5 * time.Second,
	}

	renderer, err := zen.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Close()

	app.Get("/assets/*", renderer.Static())

	app.Get("/islands/counter", func(c fiber.Ctx) error {
		return renderer.RenderIsland(c, "Counter", map[string]any{
			"count": 0,
		})
	})

	app.Post("/contact", func(c fiber.Ctx) error {
		name := c.FormValue("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("name is required")
		}

		return c.Redirect().To("/")
	})

	app.Get("/*", func(c fiber.Ctx) error {
		return renderer.RenderPage(c, "App", map[string]any{
			"url": c.OriginalURL(),
		}, zen.WithTitle("Zen App"))
	})

	log.Fatal(app.Listen(port, fiber.ListenConfig{
		DisableStartupMessage: true,
	}))
}
