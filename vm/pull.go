package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carbon-os/compute-image/internal/logf"
	"github.com/carbon-os/compute-image/registry"
)

// Pull downloads and converts a VM image described by ref.
func Pull(ref Ref) (*Image, error) {
	if ref.Arch == "" {
		return nil, fmt.Errorf("vm: Arch must be set (e.g. \"amd64\", \"arm64\")")
	}

	name, version, err := ParseRef(ref.Image)
	if err != nil {
		return nil, fmt.Errorf("vm: parse ref: %w", err)
	}

	// Validate version and arch before touching the network.
	if err := Validate(name, version, ref.Arch); err != nil {
		return nil, fmt.Errorf("vm: %w", err)
	}

	if ref.Registry == "" {
		reg, err := DefaultRegistry(name)
		if err != nil {
			return nil, fmt.Errorf("vm: %w", err)
		}
		ref.Registry = reg
		logf.Logf("[*] Using default registry: %s", ref.Registry)
	}

	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return nil, fmt.Errorf("vm: resolve paths: %w", err)
	}

	for _, d := range []string{paths.Cache, paths.Dir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("vm: create dir %s: %w", d, err)
		}
	}

	logf.Logf("[*] Pulling VM image: %s (%s) from %s", ref.Image, ref.Arch, ref.Registry)
	logf.Logf("    Dir: %s", paths.Dir)

	cachedQcow2, err := fetchVMImage(ref.Image, ref.Registry, ref.Arch, paths.Cache)
	if err != nil {
		return nil, err
	}

	logf.Logf("[*] Converting qcow2 → raw...")
	rawPath := strings.TrimSuffix(cachedQcow2, ".qcow2") + ".raw"
	if err := convertQcow2ToRaw(cachedQcow2, rawPath); err != nil {
		return nil, fmt.Errorf("vm: convert qcow2: %w", err)
	}

	logf.Logf("[*] Converting raw → VHD: %s", paths.Disk)
	if err := convertRawToVHD(rawPath, paths.Disk); err != nil {
		return nil, fmt.Errorf("vm: convert vhd: %w", err)
	}
	os.Remove(rawPath)

	logf.Logf("[+] VM image ready.")
	logf.Logf("    Disk: %s", paths.Disk)

	return &Image{Image: ref.Image, Paths: paths, OutPath: paths.Disk}, nil
}

// Remove deletes all on-disk data for ref. Cache files are left intact.
func Remove(ref Ref) error {
	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("vm: remove dir: %w", err)
	}
	logf.Logf("[+] Removed: %s", paths.Dir)
	return nil
}

// ResolvePaths returns the fully resolved paths for ref without pulling anything.
func ResolvePaths(ref Ref) (Paths, error) {
	if ref.Arch == "" {
		return Paths{}, fmt.Errorf("vm: Arch must be set (e.g. \"amd64\", \"arm64\")")
	}
	ref.Dir = resolveDir(ref.Dir)
	return resolvePaths(ref)
}

// ParseRef splits "name:version" into its components.
// Version defaults to "latest" if omitted.
func ParseRef(image string) (name, version string, err error) {
	if image == "" {
		return "", "", fmt.Errorf("image ref must not be empty")
	}
	if i := strings.LastIndexByte(image, ':'); i >= 0 {
		name, version = image[:i], image[i+1:]
	} else {
		name, version = image, "latest"
	}
	if name == "" {
		return "", "", fmt.Errorf("image ref %q: name must not be empty", image)
	}
	if version == "" {
		return "", "", fmt.Errorf("image ref %q: version must not be empty", image)
	}
	return name, version, nil
}

// fetchVMImage downloads the qcow2 into cacheDir and returns the local path.
// Returns immediately on a cache hit.
func fetchVMImage(image, reg, arch, cacheDir string) (string, error) {
	name, version, err := ParseRef(image)
	if err != nil {
		return "", err
	}

	url := buildImageURL(reg, name, version, arch)
	dest := filepath.Join(cacheDir, filepath.Base(url))

	if registry.IsCacheValid(dest) {
		logf.Logf("[*] Using cached image: %s", filepath.Base(dest))
		return dest, nil
	}

	logf.Logf("[*] Downloading: %s", url)
	tmp := dest + ".tmp"
	if err := registry.DownloadFile(url, tmp); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("fetch vm image: %w", err)
	}
	fmt.Println()

	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("fetch vm image: finalise: %w", err)
	}
	return dest, nil
}

func resolvePaths(ref Ref) (Paths, error) {
	name, version, err := ParseRef(ref.Image)
	if err != nil {
		return Paths{}, err
	}
	imageDir := filepath.Join(ref.Dir, ref.Registry, name, version, ref.Arch)
	return Paths{
		Dir:   imageDir,
		Disk:  filepath.Join(imageDir, "disk.vhd"),
		Cache: filepath.Join(ref.Dir, "cache"),
	}, nil
}

func resolveDir(dir string) string {
	if dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "carbon")
	}
	return ".carbon"
}