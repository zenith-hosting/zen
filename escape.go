package zen

import "encoding/json"

type hydrationData struct {
	Page             string `json:"page,omitempty"`
	Island           string `json:"island,omitempty"`
	IdentifierPrefix string `json:"identifierPrefix,omitempty"`
	Props            any    `json:"props"`
}

func serializeHydrationData(data hydrationData) (string, error) {
	raw, err := json.Marshal(data)
	return string(raw), err
}
