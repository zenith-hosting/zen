package zen

import "context"

type ssrClient interface {
	Render(ctx context.Context, req ssrRequest) (ssrResponse, error)
}

type ssrRequest struct {
	Mode             string `json:"mode,omitempty"`
	Page             string `json:"page,omitempty"`
	Island           string `json:"island,omitempty"`
	IdentifierPrefix string `json:"identifierPrefix,omitempty"`
	Props            any    `json:"props"`
}

type ssrResponse struct {
	HTML string `json:"html"`
	Head string `json:"head"`
}
