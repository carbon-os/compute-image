package compute_image

import (
	"fmt"
	"os"
)

func pullVM(ref VMRef) (*VMImage, error) {
	if err := os.MkdirAll(ref.Cache, 0755); err != nil {
		return nil, fmt.Errorf("compute-image: create cache dir: %w", err)
	}

	logf("[*] Pulling VM image: %s from %s", ref.Image, ref.Registry)

	cachedQcow2, err := fetchVMImage(ref.Image, ref.Registry, ref.Cache)
	if err != nil {
		return nil, err
	}

	logf("[*] Converting to raw...")
	rawPath := cachedQcow2[:len(cachedQcow2)-len(".qcow2")] + ".raw"
	if err := convertQcow2ToRaw(cachedQcow2, rawPath); err != nil {
		return nil, fmt.Errorf("compute-image: convert qcow2: %w", err)
	}

	logf("[*] Preparing output: %s", ref.Out)
	if err := extractVM(rawPath, ref.Out); err != nil {
		return nil, fmt.Errorf("compute-image: extract vm: %w", err)
	}

	logf("[+] VM image ready: %s", ref.Out)
	return &VMImage{OutPath: ref.Out}, nil
}