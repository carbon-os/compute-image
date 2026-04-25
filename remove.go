package compute_image

import (
	"fmt"
	"os"
)

// RemoveAll removes the entire image root directory, properly releasing any
// HCS layer locks before deletion.
func RemoveAll(dir string) error {
	prepareRemove(dir) // platform-specific pre-delete cleanup; see remove_windows.go
	return os.RemoveAll(dir)
}

// Remove deletes all on-disk data for the given image ref — the image
// directory (base + scratch for containers, disk.vhd for VMs).
// Cache files are left intact; remove them separately if desired.
func Remove(ref Ref) error {
	switch r := ref.(type) {
	case ContainerRef:
		r.Dir = resolveDir(r.Dir)
		return removeContainer(r)
	case VMRef:
		r.Dir = resolveDir(r.Dir)
		return removeVM(r)
	default:
		return fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

func removeVM(ref VMRef) error {
	paths, err := resolveVMPaths(ref)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("compute-image: remove vm dir: %w", err)
	}
	logf("[+] Removed: %s", paths.Dir)
	return nil
}