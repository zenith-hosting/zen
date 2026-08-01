package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zenith-hosting/zen"
)

func main() {
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

	log.Fatal(http.ListenAndServe(port, routes(renderer)))
}

func routes(renderer *zen.Renderer) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(renderer.AssetsDir()))))
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		response, err := renderer.RenderIsland(r.Context(), r.URL.RequestURI(), "User", map[string]any{"name": name})
		send(w, response, err)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RequestURI()
		props := map[string]any{"url": url}

		if r.URL.Path == "/" {
			counter, err := renderer.RenderIsland(r.Context(), url, "Counter", map[string]any{"count": 0})
			if err != nil {
				send(w, zen.Response{}, err)
				return
			}
			props["counter"] = string(counter.Body)

			if name := strings.TrimSpace(r.URL.Query().Get("name")); name != "" {
				user, err := renderer.RenderIsland(r.Context(), url, "User", map[string]any{"name": name})
				if err != nil {
					send(w, zen.Response{}, err)
					return
				}
				props["user"] = string(user.Body)
			}
		}

		response, err := renderer.RenderPage(r.Context(), url, "App", props, zen.WithTitle("Zen App"))
		send(w, response, err)
	})

	return mux
}

func send(w http.ResponseWriter, response zen.Response, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", response.ContentType)
	w.WriteHeader(response.Status)
	_, _ = w.Write(response.Body)
}
