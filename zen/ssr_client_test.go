package zen

import (
	"context"
	"testing"
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

func TestSSRClientInterfaceCapturesRenderRequest(t *testing.T) {
	client := &fakeSSRClient{
		res: ssrResponse{HTML: "<main>Hello</main>"},
	}

	res, err := client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{"title": "Hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.req.Page != "Home" {
		t.Fatalf("expected page Home, got %q", client.req.Page)
	}
	if res.HTML != "<main>Hello</main>" {
		t.Fatalf("expected rendered html, got %q", res.HTML)
	}
}

func TestProcessSSRClientRendersThroughWorker(t *testing.T) {
	client, err := newProcessSSRClient([]string{
		"node",
		"../js/ssr-worker.mjs",
		"--entry",
		"../js/fixtures/entry-server-ok.mjs",
	})
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	defer client.Close()

	res, err := client.Render(context.Background(), ssrRequest{
		URL:  "/",
		Page: "Home",
		Props: map[string]string{
			"title": "Hello",
		},
	})
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}

	if res.HTML != `<main data-page="Home">Hello</main>` {
		t.Fatalf("unexpected html: %s", res.HTML)
	}
}

func TestProcessSSRClientReturnsWorkerError(t *testing.T) {
	client, err := newProcessSSRClient([]string{
		"node",
		"../js/ssr-worker.mjs",
		"--entry",
		"../js/fixtures/entry-server-error.mjs",
	})
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	defer client.Close()

	_, err = client.Render(context.Background(), ssrRequest{
		URL:   "/",
		Page:  "Home",
		Props: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected render error")
	}
}
