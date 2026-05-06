//go:build windows

package windows

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Microsoft/hcsshim"
	"github.com/carbon-os/compute-image/internal/logf"
	"golang.org/x/sys/windows"
)

// Remove destroys the HCS layers for ref and deletes the image directory.
func Remove(ref Ref) error {
	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return err
	}

	// Destroy scratch before base — scratch holds a reference to base in its
	// layer chain, so order matters.
	for _, layerDir := range []string{paths.Scratch, paths.Base} {
		if _, err := os.Stat(layerDir); os.IsNotExist(err) {
			continue
		}
		di := hcsshim.DriverInfo{HomeDir: filepath.Dir(layerDir), Flavour: 1}
		if err := hcsshim.DestroyLayer(di, filepath.Base(layerDir)); err != nil {
			// Non-fatal: log and continue — the layer may already be gone.
			logf.Logf("[!] DestroyLayer %s: %v", filepath.Base(layerDir), err)
		}
	}

	stripAttributes(paths.Dir)
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("container/windows: remove dir: %w", err)
	}
	logf.Logf("[+] Removed: %s", paths.Dir)
	return nil
}

// RemoveAll releases HCS layer locks and strips restrictive file attributes
// across dir, then calls os.RemoveAll. Use instead of os.RemoveAll for any
// directory that may contain HCS layer data.
func RemoveAll(dir string) error {
	prepareRemove(dir)
	return os.RemoveAll(dir)
}

// prepareRemove releases HCS layer locks and strips restrictive file attributes.
func prepareRemove(dir string) {
	destroyHCSLayers(dir)
	stripAttributes(dir)
}

// destroyHCSLayers walks dir and calls hcsshim.DestroyLayer on every HCS
// layer directory it finds.
func destroyHCSLayers(dir string) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if isHCSLayer(path) {
			di := hcsshim.DriverInfo{HomeDir: filepath.Dir(path), Flavour: 1}
			if err := hcsshim.DestroyLayer(di, filepath.Base(path)); err != nil {
				logf.Logf("[!] DestroyLayer %s: %v", filepath.Base(path), err)
			}
		}
		return nil
	})
}

// isHCSLayer reports whether dir is an HCS layer directory.
// Base layers contain Files\ or Hives\; scratch layers contain layerchain.json.
func isHCSLayer(dir string) bool {
	for _, marker := range []string{"Files", "Hives", "layerchain.json"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// stripAttributes resets every file under dir to FILE_ATTRIBUTE_NORMAL.
// Belt-and-suspenders after DestroyLayer: HCS marks layer files read-only
// and system, which causes RemoveAll to fail if any slip through.
func stripAttributes(dir string) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		p, _ := windows.UTF16PtrFromString(path)
		windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_NORMAL)
		return nil
	})
}