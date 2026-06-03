package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestTodoRoutesReturnJSONForEnhancedRequests(t *testing.T) {
	app := fiber.New()
	store := NewTodoStore()
	registerTodoRoutes(app, store)

	addReq := formRequest(t, http.MethodPost, "/todos", "text=write+example")
	addReq.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

	addRes, err := app.Test(addReq)
	if err != nil {
		t.Fatalf("add request failed: %v", err)
	}

	if addRes.StatusCode != fiber.StatusOK {
		t.Fatalf("expected add status 200, got %d", addRes.StatusCode)
	}

	addTodos := readTodosResponse(t, addRes)
	if len(addTodos) != 1 || addTodos[0].ID != 1 || addTodos[0].Text != "write example" {
		t.Fatalf("unexpected add response: %#v", addTodos)
	}

	deleteReq := formRequest(t, http.MethodPost, "/todos/1/delete", "")
	deleteReq.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

	deleteRes, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}

	if deleteRes.StatusCode != fiber.StatusOK {
		t.Fatalf("expected delete status 200, got %d", deleteRes.StatusCode)
	}

	deleteTodos := readTodosResponse(t, deleteRes)
	if len(deleteTodos) != 0 {
		t.Fatalf("expected empty todo list after delete, got %#v", deleteTodos)
	}
}

func TestTodoRoutesRedirectForPlainFormPosts(t *testing.T) {
	app := fiber.New()
	store := NewTodoStore()
	registerTodoRoutes(app, store)

	req := formRequest(t, http.MethodPost, "/todos", "text=write+example")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != fiber.StatusSeeOther {
		t.Fatalf("expected redirect status 303, got %d", res.StatusCode)
	}
	if location := res.Header.Get(fiber.HeaderLocation); location != "/" {
		t.Fatalf("expected redirect to /, got %q", location)
	}
}

func formRequest(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)
	return req
}

func readTodosResponse(t *testing.T, res *http.Response) []Todo {
	t.Helper()

	if contentType := res.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
		t.Fatalf("expected JSON response, got %q", contentType)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var body struct {
		Todos []Todo `json:"todos"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("failed to decode response %q: %v", raw, err)
	}

	return body.Todos
}
