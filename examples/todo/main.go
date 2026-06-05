package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/zenith/zen/zen"
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type TodoStore struct {
	mu     sync.Mutex
	nextID int
	todos  []Todo
}

func NewTodoStore() *TodoStore {
	return &TodoStore{nextID: 1}
}

func (s *TodoStore) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todos := make([]Todo, len(s.todos))
	copy(todos, s.todos)
	return todos
}

func (s *TodoStore) Add(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.todos = append(s.todos, Todo{
		ID:   s.nextID,
		Text: text,
	})
	s.nextID++
}

func (s *TodoStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, todo := range s.todos {
		if todo.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}

	return false
}

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
		DefaultTitle:  "Zen Todo Example",
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

	store := NewTodoStore()

	if !dev {
		app.Get("/assets/*", renderer.Static())
	}

	app.Get("/", func(c fiber.Ctx) error {
		return renderer.RenderPage(c, "Todo", map[string]any{
			"todos": store.List(),
		}, zen.WithTitle("Todo List"))
	})

	app.Get("/islands/counter", func(c fiber.Ctx) error {
		return renderer.RenderIsland(c, "Counter", map[string]any{
			"count": 0,
		})
	})

	registerTodoRoutes(app, store)

	log.Fatal(app.Listen(":3000", fiber.ListenConfig{
		DisableStartupMessage: true,
	}))
}

func registerTodoRoutes(app *fiber.App, store *TodoStore) {
	app.Post("/todos", func(c fiber.Ctx) error {
		text := strings.TrimSpace(c.FormValue("text"))
		if text != "" {
			store.Add(text)
		}

		if !wantsJSON(c) {
			return c.Redirect().To("/")
		}

		return c.JSON(map[string]any{
			"todos": store.List(),
		})
	})

	app.Post("/todos/:id/delete", func(c fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err == nil {
			store.Delete(id)
		}

		if !wantsJSON(c) {
			return c.Redirect().To("/")
		}

		return c.JSON(map[string]any{
			"todos": store.List(),
		})
	})
}

func wantsJSON(c fiber.Ctx) bool {
	return strings.Contains(c.Get(fiber.HeaderAccept), fiber.MIMEApplicationJSON)
}
