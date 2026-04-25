//go:build windows

package compute_image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Microsoft/hcsshim"
	"golang.org/x/sys/windows"
)

// prepareRemove releases HCS layer locks and strips restrictive file
// attributes across dir before os.RemoveAll is called.
func prepareRemove(dir string) {
	destroyHCSLayers(dir)
	stripAttributes(dir)
}

// removeContainer destroys the HCS base and scratch layers for the given ref
// and then removes the image directory.
func removeContainer(ref ContainerRef) error {
	paths, err := resolveContainerPaths(ref)
	if err != nil {
		return err
	}

	// Destroy scratch before base — scratch holds a reference to base in its
	// layer chain, so the order matters.
	for _, layerDir := range []string{paths.Scratch, paths.Base} {
		if _, err := os.Stat(layerDir); os.IsNotExist(err) {
			continue
		}
		di := hcsshim.DriverInfo{
			HomeDir: filepath.Dir(layerDir),
			Flavour: 1, // windowsfilter
		}
		if err := hcsshim.DestroyLayer(di, filepath.Base(layerDir)); err != nil {
			// Non-fatal: log and continue — the layer may already be gone.
			logf("[!] DestroyLayer %s: %v", filepath.Base(layerDir), err)
		}
	}

	stripAttributes(paths.Dir)
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("compute-image: remove container dir: %w", err)
	}
	logf("[+] Removed: %s", paths.Dir)
	return nil
}

// destroyHCSLayers walks dir and calls hcsshim.DestroyLayer on every HCS
// layer directory it finds. Used by RemoveAll before os.RemoveAll.
func destroyHCSLayers(dir string) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if isHCSLayer(path) {
			di := hcsshim.DriverInfo{
				HomeDir: filepath.Dir(path),
				Flavour: 1,
			}
			if err := hcsshim.DestroyLayer(di, filepath.Base(path)); err != nil {
				logf("[!] DestroyLayer %s: %v", filepath.Base(path), err)
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