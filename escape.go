package zen

import (
	"bytes"
	"encoding/json"
)

type hydrationData struct {
	Page   string `json:"page,omitempty"`
	Island string `json:"island,omitempty"`
	Props  any    `json:"props"`
}

func serializeHydrationData(data hydrationData) (string, error) {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(data); err != nil {
		return "", err
	}

	out := buf.String()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}

	return out, nil
}
