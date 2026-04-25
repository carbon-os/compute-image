package compute_image

import (
	"fmt"
	"os"
)

func pullVM(ref VMRef) (*VMImage, error) {
	paths, err := resolveVMPaths(ref)
	if err != nil {
		return nil, fmt.Errorf("compute-image: resolve paths: %w", err)
	}

	for _, d := range []string{paths.Cache, paths.Dir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("compute-image: create dir %s: %w", d, err)
		}
	}

	logf("[*] Pulling VM image: %s from %s", ref.Image, ref.Registry)
	logf("    Dir: %s", paths.Dir)

	cachedQcow2, err := fetchVMImage(ref.Image, ref.Registry, paths.Cache)
	if err != nil {
		return nil, err
	}

	logf("[*] Converting to raw...")
	rawPath := cachedQcow2[:len(cachedQcow2)-len(".qcow2")] + ".raw"
	if err := convertQcow2ToRaw(cachedQcow2, rawPath); err != nil {
		return nil, fmt.Errorf("compute-image: convert qcow2: %w", err)
	}

	logf("[*] Extracting to: %s", paths.Disk)
	if err := extractVM(rawPath, paths.Dir); err != nil {
		return nil, fmt.Errorf("compute-image: extract vm: %w", err)
	}

	logf("[+] VM image ready.")
	logf("    Disk: %s", paths.Disk)

	return &VMImage{
		Image:   ref.Image,
		Paths:   paths,
		OutPath: paths.Disk,
	}, nil
}