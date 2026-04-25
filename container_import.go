//go:build windows

package compute_image

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Microsoft/hcsshim"
	"github.com/Microsoft/hcsshim/pkg/ociwclayer"
)

// importLayers imports each tar into its own numbered subdirectory of baseDir
// (e.g. base/00, base/01 …), builds the HCS parent chain as it goes, then
// creates a writable scratch layer on top of the whole stack.
// It returns the path of the topmost (highest-numbered) layer so the caller
// can hand it to the container runtime.
func importLayers(ctx context.Context, tarPaths []string, baseDir, scratchDir string) (string, error) {
	// Always start clean.
	for _, d := range []string{baseDir, scratchDir} {
		if _, err := os.Stat(d); err == nil {
			logf("[*] Cleaning: %s", d)
			prepareRemove(d)
			if err := os.RemoveAll(d); err != nil {
				return "", fmt.Errorf("compute-image: clean %s: %w", d, err)
			}
		}
	}

	for _, d := range []string{baseDir, scratchDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", fmt.Errorf("compute-image: create dir %s: %w", d, err)
		}
	}

	logf("[*] Importing %d layer(s) into: %s", len(tarPaths), baseDir)

	var layerDirs []string // accumulated parent chain, oldest-first
	for i, tarPath := range tarPaths {
		layerDir := filepath.Join(baseDir, fmt.Sprintf("%02d", i))
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return "", fmt.Errorf("compute-image: create layer dir: %w", err)
		}
		logf("    -> [%d/%d] %s", i+1, len(tarPaths), filepath.Base(tarPath))

		// hcsshim APIs expect the parent chain ordered from newest to oldest
		parentChain := reversePaths(layerDirs)

		if err := importOneTar(ctx, tarPath, layerDir, parentChain); err != nil {
			return "", fmt.Errorf("compute-image: import %s: %w", filepath.Base(tarPath), err)
		}

		// If this layer contains a UtilityVM directory, post-process it into
		// the SystemTemplateBase.vhdx that HCS needs for Hyper-V isolation.
		// ociwclayer extracts the raw Files\ tree but leaves the vhdx as a
		// stub — ProcessUtilityVMImage stamps it into the real UVM image.
		uvmDir := filepath.Join(layerDir, "UtilityVM")
		if _, err := os.Stat(uvmDir); err == nil {
			logf("       [*] Processing UtilityVM image (Hyper-V isolation support)...")

			// ProcessUtilityVMImage fails with "The file exists" if the vhdx
			// stubs from a previous partial import are still present.
			// Delete them before processing so it can write fresh ones.
			for _, vhdx := range []string{
				filepath.Join(uvmDir, "SystemTemplate.vhdx"),
				filepath.Join(uvmDir, "SystemTemplateBase.vhdx"),
			} {
				if err := os.Remove(vhdx); err != nil && !os.IsNotExist(err) {
					return "", fmt.Errorf("compute-image: remove stale vhdx %s: %w", vhdx, err)
				}
			}

			if err := hcsshim.ProcessUtilityVMImage(uvmDir); err != nil {
				return "", fmt.Errorf("compute-image: ProcessUtilityVMImage for layer %02d: %w", i, err)
			}
			logf("       [+] UtilityVM image ready.")
		}

		layerDirs = append(layerDirs, layerDir)
	}

	topLayer := layerDirs[len(layerDirs)-1]
	logf("[+] Base layers imported. Top layer: %s", topLayer)

	parentChain := reversePaths(layerDirs)

	logf("[*] Creating scratch layer: %s", scratchDir)
	di := hcsshim.DriverInfo{
		HomeDir: filepath.Dir(scratchDir),
		Flavour: 1,
	}
	if err := hcsshim.CreateSandboxLayer(di, filepath.Base(scratchDir), topLayer, parentChain); err != nil {
		return "", fmt.Errorf("compute-image: create scratch: %w", err)
	}

	chain, _ := json.Marshal(parentChain)
	lcPath := filepath.Join(scratchDir, "layerchain.json")
	if err := os.WriteFile(lcPath, chain, 0644); err != nil {
		return "", fmt.Errorf("compute-image: write layerchain.json: %w", err)
	}

	logf("[+] Scratch layer ready.")
	return topLayer, nil
}

// reversePaths creates a new slice with paths ordered newest-to-oldest.
func reversePaths(paths []string) []string {
	reversed := make([]string, len(paths))
	for i, p := range paths {
		reversed[len(paths)-1-i] = p
	}
	return reversed
}

// importOneTar decompresses tarPath and imports it as an HCS layer into
// destDir. parentDirs is the ordered list of parent layer directories
// (newest-to-oldest).
func importOneTar(ctx context.Context, tarPath, destDir string, parentDirs []string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	_, err = ociwclayer.ImportLayerFromTar(ctx, gzr, destDir, parentDirs)
	return err
}

// scratchLayerExists reports whether an already-prepared scratch exists whose
// layerchain.json contains baseDir anywhere in the chain.
func scratchLayerExists(scratchDir, baseDir string) bool {
	data, err := os.ReadFile(filepath.Join(scratchDir, "layerchain.json"))
	if err != nil {
		return false
	}
	var chain []string
	if err := json.Unmarshal(data, &chain); err != nil || len(chain) == 0 {
		return false
	}
	for _, entry := range chain {
		if strings.EqualFold(entry, baseDir) {
			return true
		}
	}
	return false
}