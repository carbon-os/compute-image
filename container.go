package compute_image

import (
	"context"
	"fmt"
	"os"
)

func pullContainer(ref ContainerRef) (*ContainerImage, error) {
	ctx := context.Background()

	paths, err := resolveContainerPaths(ref)
	if err != nil {
		return nil, fmt.Errorf("compute-image: resolve paths: %w", err)
	}

	if err := enablePrivileges(); err != nil {
		return nil, fmt.Errorf("compute-image: privileges: %w", err)
	}

	for _, d := range []string{paths.Cache, paths.Dir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("compute-image: create dir %s: %w", d, err)
		}
	}

	registry, repo, tag, err := parseImageRef(ref.Image)
	if err != nil {
		return nil, fmt.Errorf("compute-image: parse ref: %w", err)
	}

	logf("[*] Pulling container image: %s", ref.Image)
	logf("    Dir: %s", paths.Dir)

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

	cachedTars, err := downloadLayers(registry, repo, layers, paths.Cache)
	if err != nil {
		return nil, err
	}

	if err := importLayers(ctx, cachedTars, paths.Base, paths.Scratch); err != nil {
		return nil, err
	}

	logf("[+] Container image ready.")
	logf("    Base:    %s", paths.Base)
	logf("    Scratch: %s", paths.Scratch)

	return &ContainerImage{
		Image:     ref.Image,
		Paths:     paths,
		BaseLayer: paths.Base,
		Scratch:   paths.Scratch,
	}, nil
}