//go:build windows

package windows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carbon-os/compute-image/internal/logf"
	"github.com/carbon-os/compute-image/registry"
)

// Pull downloads and imports a Windows container image described by ref.
func Pull(ref Ref) (*Image, error) {
	ctx := context.Background()

	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return nil, fmt.Errorf("container/windows: resolve paths: %w", err)
	}

	if err := enablePrivileges(); err != nil {
		return nil, fmt.Errorf("container/windows: privileges: %w", err)
	}

	for _, d := range []string{paths.Cache, paths.Dir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("container/windows: create dir %s: %w", d, err)
		}
	}

	reg, repo, tag, err := registry.ParseImageRef(ref.Image)
	if err != nil {
		return nil, fmt.Errorf("container/windows: parse ref: %w", err)
	}

	logf.Logf("[*] Pulling Windows container image: %s", ref.Image)
	logf.Logf("    Dir: %s", paths.Dir)

	manifestBytes, err := registry.FetchManifest(reg, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("container/windows: fetch manifest: %w", err)
	}
	digest := registry.ResolveWindowsDigest(manifestBytes, tag)
	manifestBytes, err = registry.FetchManifest(reg, repo, digest)
	if err != nil {
		return nil, fmt.Errorf("container/windows: fetch image manifest: %w", err)
	}
	layers, err := registry.ParseManifestLayers(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("container/windows: parse manifest: %w", err)
	}
	logf.Logf("[+] Manifest resolved. %d layer(s).", len(layers))

	cachedTars, err := registry.DownloadLayers(reg, repo, layers, paths.Cache)
	if err != nil {
		return nil, err
	}

	topLayer, err := importLayers(ctx, cachedTars, paths.Base, paths.Scratch)
	if err != nil {
		return nil, err
	}

	logf.Logf("[+] Container image ready.")
	logf.Logf("    Base:    %s", paths.Base)
	logf.Logf("    Top:     %s", topLayer)
	logf.Logf("    Scratch: %s", paths.Scratch)

	return &Image{
		Image:     ref.Image,
		Paths:     paths,
		BaseLayer: topLayer,
		Scratch:   paths.Scratch,
	}, nil
}

// ResolvePaths returns the fully resolved paths for ref without pulling anything.
func ResolvePaths(ref Ref) (Paths, error) {
	ref.Dir = resolveDir(ref.Dir)
	return resolvePaths(ref)
}

func resolvePaths(ref Ref) (Paths, error) {
	reg, repo, tag, err := registry.ParseImageRef(ref.Image)
	if err != nil {
		return Paths{}, err
	}
	imageDir := filepath.Join(ref.Dir, reg, filepath.FromSlash(repo), tag)
	return Paths{
		Dir:     imageDir,
		Base:    filepath.Join(imageDir, "base"),
		Scratch: filepath.Join(imageDir, "scratch"),
		Cache:   filepath.Join(ref.Dir, "cache"),
	}, nil
}

func resolveDir(dir string) string {
	if dir != "" {
		return dir
	}
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local = "."
	}
	return filepath.Join(local, "carbon")
}