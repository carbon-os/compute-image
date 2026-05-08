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

	rawPath := strings.TrimSuffix(cachedQcow2, ".qcow2") + ".raw"
	logf.Logf("[*] Converting qcow2 → raw...")
	if err := convertQcow2ToRaw(cachedQcow2, rawPath); err != nil {
		return nil, fmt.Errorf("vm: convert qcow2: %w", err)
	}

	logf.Logf("[*] Finalising disk image: %s", paths.Disk)
	if err := finalizeDisk(rawPath, paths.Disk); err != nil {
		return nil, fmt.Errorf("vm: finalise disk: %w", err)
	}
	os.Remove(rawPath)

	if ref.ExtractKernel {
		if err := runExtract(name, version, ref.Arch, paths); err != nil {
			return nil, err
		}
	}

	logf.Logf("[+] VM image ready.")
	logf.Logf("    Disk: %s", paths.Disk)

	return &Image{Image: ref.Image, Paths: paths, OutPath: paths.Disk}, nil
}

// runExtract resolves the BootConfig for this distro/version/arch and extracts
// vmlinuz and initrd into the same directory as the disk image.
func runExtract(name, version, arch string, paths Paths) error {
	p, ok := lookup(name)
	if !ok {
		return fmt.Errorf("vm: no extraction support for distro %q", name)
	}
	cfg, err := p.BootConfig(version, arch)
	if err != nil {
		return fmt.Errorf("vm: boot config: %w", err)
	}
	logf.Logf("[*] Extracting kernel and initrd from partition %d (%s)...",
		cfg.Partition, cfg.BootDir)
	if err := extractKernelAndInitrd(paths.Disk, cfg, paths.Dir); err != nil {
		return fmt.Errorf("vm: extract: %w", err)
	}
	logf.Logf("[+] Kernel: %s", filepath.Join(paths.Dir, "vmlinuz"))
	logf.Logf("[+] Initrd: %s", filepath.Join(paths.Dir, "initrd"))
	return nil
}

// resolvePaths uses the diskFile constant so the extension is correct per platform.
func resolvePaths(ref Ref) (Paths, error) {
	name, version, err := ParseRef(ref.Image)
	if err != nil {
		return Paths{}, err
	}
	imageDir := filepath.Join(ref.Dir, ref.Registry, name, version, ref.Arch)
	return Paths{
		Dir:   imageDir,
		Disk:  filepath.Join(imageDir, diskFile),
		Cache: filepath.Join(ref.Dir, "cache"),
	}, nil
}

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

func ResolvePaths(ref Ref) (Paths, error) {
	if ref.Arch == "" {
		return Paths{}, fmt.Errorf("vm: Arch must be set (e.g. \"amd64\", \"arm64\")")
	}
	ref.Dir = resolveDir(ref.Dir)
	return resolvePaths(ref)
}

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

func resolveDir(dir string) string {
	if dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "carbon")
	}
	return ".carbon"
}