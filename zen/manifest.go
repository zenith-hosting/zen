package zen

import (
	"encoding/json"
	"errors"
	"os"
)

type viteManifest map[string]viteManifestEntry

type viteManifestEntry struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

type clientAssets struct {
	Scripts []string
	Styles  []string
}

func readManifest(path string) (viteManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest viteManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (m viteManifest) clientAssets(entryName string) (clientAssets, error) {
	entry, ok := m[entryName]
	if !ok {
		return clientAssets{}, errors.New("zen: Vite manifest entry not found: " + entryName)
	}

	var out clientAssets

	seenScripts := map[string]bool{}
	for _, importName := range entry.Imports {
		importEntry, ok := m[importName]
		if !ok || importEntry.File == "" {
			continue
		}
		path := "/" + importEntry.File
		if !seenScripts[path] {
			out.Scripts = append(out.Scripts, path)
		}
		seenScripts[path] = true
	}

	if entry.File != "" {
		path := "/" + entry.File
		if !seenScripts[path] {
			out.Scripts = append(out.Scripts, path)
		}
	}

	seenStyles := map[string]bool{}
	for _, css := range entry.CSS {
		path := "/" + css
		if !seenStyles[path] {
			out.Styles = append(out.Styles, path)
		}
		seenStyles[path] = true
	}

	return out, nil
}
