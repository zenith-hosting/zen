package zen

import (
	"context"
)

type fakeSSRClient struct {
	req ssrRequest
	res ssrResponse
	err error
}

func (f *fakeSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	f.req = req
	return f.res, f.err
}
