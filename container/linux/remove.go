package linux

import (
	"fmt"
	"os"

	"github.com/carbon-os/compute-image/internal/logf"
)

// Remove deletes all on-disk data for ref. Cache files are left intact.
func Remove(ref Ref) error {
	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("container/linux: remove dir: %w", err)
	}
	logf.Logf("[+] Removed: %s", paths.Dir)
	return nil
}