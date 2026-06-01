package zen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL + "/__zen/render",
		Timeout:   time.Second,
	})

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

func TestHTTPSSRClientReturnsRendererError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(httpRendererErrorResponse{
			Error: httpRendererError{
				Message: "Unknown page: Admin",
				Stack:   "Error: Unknown page: Admin",
			},
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

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

func TestHTTPSSRClientHandlesParallelRequests(t *testing.T) {
	var count atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)

		var req ssrRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(ssrResponse{
			HTML: "<main>" + req.Page + "</main>",
		})
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

	errs := make(chan error, 25)

	for i := 0; i < 25; i++ {
		go func() {
			_, err := client.Render(context.Background(), ssrRequest{
				URL:   "/",
				Page:  "Home",
				Props: map[string]string{},
			})
			errs <- err
		}()
	}

	for i := 0; i < 25; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("parallel render failed: %v", err)
		}
	}

	if count.Load() != 25 {
		t.Fatalf("expected 25 requests, got %d", count.Load())
	}
}

func TestHTTPSSRClientRespectsContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newHTTPSSRClient(httpSSRClientConfig{
		RenderURL: server.URL,
		Timeout:   time.Second,
	})

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
