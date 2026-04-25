package compute_image

import (
	"fmt"
	"os"
	"path/filepath"
)

// Pull downloads and prepares an image described by ref.
// Pass a ContainerRef to get a *ContainerImage, or a VMRef to get a *VMImage.
func Pull(ref Ref) (any, error) {
	switch r := ref.(type) {
	case ContainerRef:
		r.Cache = resolveCache(r.Cache)
		return pullContainer(r)
	case VMRef:
		r.Cache = resolveCache(r.Cache)
		return pullVM(r)
	default:
		return nil, fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

// resolveCache returns cache if non-empty, otherwise the default location.
func resolveCache(cache string) string {
	if cache != "" {
		return cache
	}
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local = "."
	}
	return filepath.Join(local, "carbon", "cache")
}