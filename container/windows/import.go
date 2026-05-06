//go:build windows

package windows

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
	"github.com/carbon-os/compute-image/internal/logf"
)

// importLayers imports each tar into a numbered subdirectory of baseDir
// (e.g. base/00, base/01 …), builds the HCS parent chain as it goes, then
// creates a writable scratch layer on top of the whole stack.
// Returns the path of the topmost layer.
func importLayers(ctx context.Context, tarPaths []string, baseDir, scratchDir string) (string, error) {
	// Always start clean.
	for _, d := range []string{baseDir, scratchDir} {
		if _, err := os.Stat(d); err == nil {
			logf.Logf("[*] Cleaning: %s", d)
			prepareRemove(d)
			if err := os.RemoveAll(d); err != nil {
				return "", fmt.Errorf("container/windows: clean %s: %w", d, err)
			}
		}
	}
	for _, d := range []string{baseDir, scratchDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", fmt.Errorf("container/windows: create dir %s: %w", d, err)
		}
	}

	logf.Logf("[*] Importing %d layer(s) into: %s", len(tarPaths), baseDir)

	var layerDirs []string // accumulated parent chain, oldest-first
	for i, tarPath := range tarPaths {
		layerDir := filepath.Join(baseDir, fmt.Sprintf("%02d", i))
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return "", fmt.Errorf("container/windows: create layer dir: %w", err)
		}
		logf.Logf("    -> [%d/%d] %s", i+1, len(tarPaths), filepath.Base(tarPath))

		// hcsshim APIs expect the parent chain ordered newest-to-oldest.
		if err := importOneTar(ctx, tarPath, layerDir, reversePaths(layerDirs)); err != nil {
			return "", fmt.Errorf("container/windows: import %s: %w", filepath.Base(tarPath), err)
		}

		// If this layer contains a UtilityVM directory, stamp it into the
		// SystemTemplateBase.vhdx that HCS needs for Hyper-V isolation.
		uvmDir := filepath.Join(layerDir, "UtilityVM")
		if _, err := os.Stat(uvmDir); err == nil {
			logf.Logf("       [*] Processing UtilityVM image (Hyper-V isolation support)...")
			// Remove stale vhdx stubs so ProcessUtilityVMImage can write fresh ones.
			for _, vhdx := range []string{
				filepath.Join(uvmDir, "SystemTemplate.vhdx"),
				filepath.Join(uvmDir, "SystemTemplateBase.vhdx"),
			} {
				if err := os.Remove(vhdx); err != nil && !os.IsNotExist(err) {
					return "", fmt.Errorf("container/windows: remove stale vhdx %s: %w", vhdx, err)
				}
			}
			if err := hcsshim.ProcessUtilityVMImage(uvmDir); err != nil {
				return "", fmt.Errorf("container/windows: ProcessUtilityVMImage layer %02d: %w", i, err)
			}
			logf.Logf("       [+] UtilityVM image ready.")
		}
		layerDirs = append(layerDirs, layerDir)
	}

	topLayer := layerDirs[len(layerDirs)-1]
	logf.Logf("[+] Base layers imported. Top layer: %s", topLayer)

	parentChain := reversePaths(layerDirs)

	logf.Logf("[*] Creating scratch layer: %s", scratchDir)
	di := hcsshim.DriverInfo{HomeDir: filepath.Dir(scratchDir), Flavour: 1}
	if err := hcsshim.CreateSandboxLayer(di, filepath.Base(scratchDir), topLayer, parentChain); err != nil {
		return "", fmt.Errorf("container/windows: create scratch: %w", err)
	}

	chain, _ := json.Marshal(parentChain)
	if err := os.WriteFile(filepath.Join(scratchDir, "layerchain.json"), chain, 0644); err != nil {
		return "", fmt.Errorf("container/windows: write layerchain.json: %w", err)
	}
	logf.Logf("[+] Scratch layer ready.")
	return topLayer, nil
}

func reversePaths(paths []string) []string {
	reversed := make([]string, len(paths))
	for i, p := range paths {
		reversed[len(paths)-1-i] = p
	}
	return reversed
}

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
// layerchain.json references baseDir anywhere in the chain.
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