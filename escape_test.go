package zen

import (
	"strings"
	"testing"
)

func TestSerializeDataEscapesScriptBreakout(t *testing.T) {
	payload := hydrationData{
		Page: "Home",
		Props: map[string]string{
			"message": "</script><script>alert(1)</script>",
		},
	}

	got, err := serializeHydrationData(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "</script>") {
		t.Fatalf("serialized payload contains raw script close: %s", got)
	}
	if !strings.Contains(got, `\u003c/script\u003e`) {
		t.Fatalf("serialized payload did not escape less-than characters: %s", got)
	}
}

func TestSerializeDataIncludesPageAndProps(t *testing.T) {
	payload := hydrationData{
		Page:             "Home",
		IdentifierPrefix: "zen-page-",
		Props: map[string]string{
			"title": "Hello",
		},
	}

	got, err := serializeHydrationData(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, `"page":"Home"`) {
		t.Fatalf("serialized payload missing page: %s", got)
	}
	if !strings.Contains(got, `"title":"Hello"`) {
		t.Fatalf("serialized payload missing props: %s", got)
	}
	if !strings.Contains(got, `"identifierPrefix":"zen-page-"`) {
		t.Fatalf("serialized payload missing identifier prefix: %s", got)
	}
}
