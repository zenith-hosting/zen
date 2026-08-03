package zen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPSSRClientRendersPage(t *testing.T) {
	var received ssrRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/__zen/render" {
			t.Fatalf("expected /__zen/render, got %s", r.URL.Path)
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(ssrResponse{
			HTML: "<main>Hello</main>",
			Head: "<title>Hello</title>",
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(server.URL+"/__zen/render", time.Second)

	res, err := client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{"title": "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}

	if received.Page != "Home" {
		t.Fatalf("expected page Home, got %q", received.Page)
	}

	if res.HTML != "<main>Hello</main>" {
		t.Fatalf("expected rendered html, got %q", res.HTML)
	}

	if res.Head != "<title>Hello</title>" {
		t.Fatalf("expected head html, got %q", res.Head)
	}
}

func TestHTTPSSRClientSerializesIslandMode(t *testing.T) {
	var received ssrRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(ssrResponse{
			HTML: "<button>Count 0</button>",
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(server.URL, time.Second)

	_, err := client.Render(context.Background(), ssrRequest{
		Mode:   "island",
		URL:    "/counter",
		Island: "Counter",
		Props:  map[string]int{"count": 0},
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}

	if received.Mode != "island" {
		t.Fatalf("expected island mode, got %q", received.Mode)
	}
	if received.Island != "Counter" {
		t.Fatalf("expected island Counter, got %q", received.Island)
	}
	if received.Page != "" {
		t.Fatalf("expected empty page for island render, got %q", received.Page)
	}
}

func TestHTTPSSRClientReturnsRendererError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(httpRendererErrorResponse{
			Error: httpRendererError{
				Message: "Unknown page: Admin",
			},
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(server.URL, time.Second)

	_, err := client.Render(context.Background(), ssrRequest{
		URL:   "/admin",
		Page:  "Admin",
		Props: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected render error")
	}

	if !strings.Contains(err.Error(), "Unknown page: Admin") {
		t.Fatalf("expected renderer error message, got %v", err)
	}
}

func TestHTTPSSRClientRespectsContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newHTTPSSRClient(server.URL, time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	_, err := client.Render(ctx, ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
