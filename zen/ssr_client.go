package zen

import "context"

type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}

type ssrRequest struct {
	URL   string `json:"url"`
	Page  string `json:"page"`
	Props any    `json:"props"`
}

type ssrResponse struct {
	HTML string `json:"html"`
	Head string `json:"head"`
}
