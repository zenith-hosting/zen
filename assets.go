package zen

import "path/filepath"

func (r *Renderer) AssetsDir() string {
	return filepath.Join(r.config.clientDist, "assets")
}
