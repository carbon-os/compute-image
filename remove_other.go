//go:build !windows

package compute_image

import (
	"fmt"
	"os"
)

func prepareRemove(_ string) {}

func removeContainer(ref ContainerRef) error {
	paths, err := resolveContainerPaths(ref)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(paths.Dir); err != nil {
		return fmt.Errorf("compute-image: remove container dir: %w", err)
	}
	logf("[+] Removed: %s", paths.Dir)
	return nil
}