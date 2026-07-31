package main

import (
	"log"
	"net/http"
	"os"
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

	mux.HandleFunc("GET /islands/counter", func(w http.ResponseWriter, r *http.Request) {
		response, err := renderer.RenderIsland(r.Context(), r.URL.RequestURI(), "Counter", map[string]any{
			"count": 0,
		})
		send(w, response, err)
	})

	mux.HandleFunc("POST /contact", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("name") == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RequestURI()
		response, err := renderer.RenderPage(r.Context(), url, "App", map[string]any{
			"url": url,
		}, zen.WithTitle("Zen App"))
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
