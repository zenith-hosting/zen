package zen

import "context"

type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}

type ssrRequest struct {
	Mode   string `json:"mode,omitempty"`
	URL    string `json:"url"`
	Page   string `json:"page,omitempty"`
	Island string `json:"island,omitempty"`
	Props  any    `json:"props"`
}

type ssrResponse struct {
	HTML string `json:"html"`
	Head string `json:"head"`
}
