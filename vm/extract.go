package vm

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/carbon-os/diskimg"
)

// bootVolume is the subset of diskimg's Volume interface needed for extraction.
// Defined locally so we are not coupled to unexported diskimg internals.
type bootVolume interface {
	ReadDir(string) ([]fs.DirEntry, error)
	Open(string) (fs.File, error)
	Stat(string) (fs.FileInfo, error)
	Unmount() error
}

// extractKernelAndInitrd attaches diskPath (read-only intent), mounts
// cfg.Partition, locates the kernel and initrd via glob patterns, and writes
// them to outDir as "vmlinuz" and "initrd". The source image is never modified.
func extractKernelAndInitrd(diskPath string, cfg BootConfig, outDir string) error {
	img, err := diskimg.Attach(diskPath)
	if err != nil {
		return fmt.Errorf("attach %s: %w", diskPath, err)
	}
	// Defers execute LIFO: Unmount fires before Detach, which is required.
	defer img.Detach("") //nolint:errcheck

	vol, err := img.Mount(cfg.Partition)
	if err != nil {
		return fmt.Errorf("mount partition %d of %s: %w", cfg.Partition, diskPath, err)
	}
	defer vol.Unmount() //nolint:errcheck

	entries, err := vol.ReadDir(cfg.BootDir)
	if err != nil {
		return fmt.Errorf("readdir %q on partition %d: %w", cfg.BootDir, cfg.Partition, err)
	}

	kernelFile, err := findBootFile(vol, cfg.BootDir, entries, cfg.KernelGlob)
	if err != nil {
		return fmt.Errorf("kernel not found (glob %q in %q, partition %d): %w",
			cfg.KernelGlob, cfg.BootDir, cfg.Partition, err)
	}
	initrdFile, err := findBootFile(vol, cfg.BootDir, entries, cfg.InitrdGlob)
	if err != nil {
		return fmt.Errorf("initrd not found (glob %q in %q, partition %d): %w",
			cfg.InitrdGlob, cfg.BootDir, cfg.Partition, err)
	}

	if err := copyFromVolume(vol,
		path.Join(cfg.BootDir, kernelFile),
		filepath.Join(outDir, "vmlinuz"),
	); err != nil {
		return fmt.Errorf("extract kernel: %w", err)
	}
	if err := copyFromVolume(vol,
		path.Join(cfg.BootDir, initrdFile),
		filepath.Join(outDir, "initrd"),
	); err != nil {
		return fmt.Errorf("extract initrd: %w", err)
	}
	return nil
}

// findBootFile returns the name of the first entry in dir that:
//   - matches glob (path.Match semantics),
//   - does not contain "rescue" in its name,
//   - resolves to a regular file after following symlinks.
//
// Symlink following via vol.Stat handles distros such as openSUSE Leap 16.0
// where the versioned vmlinuz-* entry is itself a symlink.
func findBootFile(vol bootVolume, dir string, entries []fs.DirEntry, glob string) (string, error) {
	for _, e := range entries {
		name := e.Name()
		if strings.Contains(name, "rescue") {
			continue
		}
		matched, err := path.Match(glob, name)
		if err != nil || !matched {
			continue
		}
		// Stat follows symlinks; skip if the resolved target is not a regular file.
		info, err := vol.Stat(path.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		return name, nil
	}
	return "", fmt.Errorf("no entry matches %q", glob)
}

// copyFromVolume streams src out of vol and writes it to the host path dst.
// dst is removed on write error so no partial file is left behind.
func copyFromVolume(vol bootVolume, src, dst string) error {
	f, err := vol.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, f); err != nil {
		os.Remove(dst)
		return fmt.Errorf("copy %s: %w", src, err)
	}
	return nil
}