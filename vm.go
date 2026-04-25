package compute_image

import (
	"fmt"
	"os"
)

func pullVM(ref VMRef) (*VMImage, error) {
	if ref.Registry == "" {
		name, _, err := parseVMRef(ref.Image)
		if err != nil {
			return nil, fmt.Errorf("compute-image: parse ref: %w", err)
		}
		reg, err := defaultVMRegistry(name)
		if err != nil {
			return nil, fmt.Errorf("compute-image: %w", err)
		}
		ref.Registry = reg
		logf("[*] Using default registry: %s", ref.Registry)
	}

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

	// qcow2 → raw (intermediate, kept in cache next to the qcow2)
	logf("[*] Converting qcow2 to raw...")
	rawPath := cachedQcow2[:len(cachedQcow2)-len(".qcow2")] + ".raw"
	if err := convertQcow2ToRaw(cachedQcow2, rawPath); err != nil {
		return nil, fmt.Errorf("compute-image: convert qcow2: %w", err)
	}

	// raw → VHD (pure Go footer append, no exec)
	logf("[*] Converting raw to VHD: %s", paths.Disk)
	if err := convertRawToVHD(rawPath, paths.Disk); err != nil {
		return nil, fmt.Errorf("compute-image: convert vhd: %w", err)
	}

	// Raw is only needed for conversion — clean it up.
	os.Remove(rawPath)

	logf("[+] VM image ready.")
	logf("    Disk: %s", paths.Disk)

	return &VMImage{
		Image:   ref.Image,
		Paths:   paths,
		OutPath: paths.Disk,
	}, nil
}