package zen

import (
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func (r *Renderer) Static() fiber.Handler {
	return static.New(filepath.Join(r.config.clientDist, "assets"))
}
