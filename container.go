package compute_image

import (
	"context"
	"fmt"
	"os"
)

func pullContainer(ref ContainerRef) (*ContainerImage, error) {
	ctx := context.Background()

	if err := enablePrivileges(); err != nil {
		return nil, fmt.Errorf("compute-image: privileges: %w", err)
	}

	if err := os.MkdirAll(ref.Cache, 0755); err != nil {
		return nil, fmt.Errorf("compute-image: create cache dir: %w", err)
	}

	registry, repo, tag, err := parseImageRef(ref.Image)
	if err != nil {
		return nil, fmt.Errorf("compute-image: parse ref: %w", err)
	}

	logf("[*] Pulling container image: %s", ref.Image)

	// Resolve manifest list → platform-specific manifest
	manifestBytes, err := fetchManifestWithAuth(registry, repo, tag, "")
	if err != nil {
		return nil, fmt.Errorf("compute-image: fetch manifest: %w", err)
	}
	digest := resolveDigest(manifestBytes, tag)
	manifestBytes, err = fetchManifestWithAuth(registry, repo, digest, "")
	if err != nil {
		return nil, fmt.Errorf("compute-image: fetch image manifest: %w", err)
	}
	layers, err := parseManifestLayers(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("compute-image: parse manifest: %w", err)
	}
	logf("[+] Manifest resolved. %d layer(s).", len(layers))

	// Download layers to cache
	cachedTars, err := downloadLayers(registry, repo, layers, ref.Cache)
	if err != nil {
		return nil, err
	}

	// Import into HCS layer + scratch
	if err := importLayers(ctx, cachedTars, ref.BaseDir, ref.Scratch); err != nil {
		return nil, err
	}

	logf("[+] Container image ready.")
	logf("    Base: %s", ref.BaseDir)
	logf("    Scratch: %s", ref.Scratch)

	return &ContainerImage{
		BaseLayer: ref.BaseDir,
		Scratch:   ref.Scratch,
	}, nil
}