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
		r.Dir = resolveDir(r.Dir)
		return pullContainer(r)
	case VMRef:
		r.Dir = resolveDir(r.Dir)
		return pullVM(r)
	default:
		return nil, fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

// resolveDir returns dir if non-empty, otherwise the default root location.
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

// ResolveContainerPaths returns the fully resolved paths for a ContainerRef
// without pulling anything. Useful for info and scripting.
func ResolveContainerPaths(ref ContainerRef) (ContainerPaths, error) {
	ref.Dir = resolveDir(ref.Dir)
	return resolveContainerPaths(ref)
}

// ResolveVMPaths returns the fully resolved paths for a VMRef
// without pulling anything. Useful for info and scripting.
func ResolveVMPaths(ref VMRef) (VMPaths, error) {
	ref.Dir = resolveDir(ref.Dir)
	return resolveVMPaths(ref)
}

func resolveContainerPaths(ref ContainerRef) (ContainerPaths, error) {
	registry, repo, tag, err := parseImageRef(ref.Image)
	if err != nil {
		return ContainerPaths{}, err
	}
	imageDir := filepath.Join(ref.Dir, registry, filepath.FromSlash(repo), tag)
	cache := filepath.Join(ref.Dir, "cache")
	return ContainerPaths{
		Dir:     imageDir,
		Base:    filepath.Join(imageDir, "base"),
		Scratch: filepath.Join(imageDir, "scratch"),
		Cache:   cache,
	}, nil
}

func resolveVMPaths(ref VMRef) (VMPaths, error) {
	name, version, err := parseVMRef(ref.Image)
	if err != nil {
		return VMPaths{}, err
	}
	imageDir := filepath.Join(ref.Dir, ref.Registry, name, version)
	cache := filepath.Join(ref.Dir, "cache")
	return VMPaths{
		Dir:   imageDir,
		Disk:  filepath.Join(imageDir, "disk.raw"),
		Cache: cache,
	}, nil
}

// HumanBytes is exported so image-cli can format sizes consistently.
func HumanBytes(b int64) string { return humanBytes(b) }