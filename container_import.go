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

func importLayers(ctx context.Context, tarPaths []string, baseDir, scratchDir string) error {
	// Wipe and recreate base and scratch — always start clean
	for _, d := range []string{baseDir, scratchDir} {
		if _, err := os.Stat(d); err == nil {
			logf("[*] Cleaning: %s", d)
			if err := os.RemoveAll(d); err != nil {
				return fmt.Errorf("compute-image: clean %s: %w", d, err)
			}
		}
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("compute-image: create base dir: %w", err)
	}

	// Import each layer in order
	logf("[*] Importing %d layer(s) into: %s", len(tarPaths), baseDir)
	for _, tarPath := range tarPaths {
		logf("    -> %s", filepath.Base(tarPath))
		if err := importOneTar(ctx, tarPath, baseDir); err != nil {
			return fmt.Errorf("compute-image: import %s: %w", filepath.Base(tarPath), err)
		}
	}
	logf("[+] Base layer imported.")

	// Create scratch (writable overlay)
	logf("[*] Creating scratch layer: %s", scratchDir)
	di := hcsshim.DriverInfo{
		HomeDir: filepath.Dir(baseDir),
		Flavour: 1, // windowsfilter
	}
	scratchID := filepath.Base(scratchDir)
	if err := hcsshim.CreateSandboxLayer(di, scratchID, baseDir, []string{baseDir}); err != nil {
		return fmt.Errorf("compute-image: create scratch: %w", err)
	}

	// Write layerchain.json so HCS knows the parent chain
	chain, _ := json.Marshal([]string{baseDir})
	lcPath := filepath.Join(scratchDir, "layerchain.json")
	if err := os.WriteFile(lcPath, chain, 0644); err != nil {
		return fmt.Errorf("compute-image: write layerchain.json: %w", err)
	}

	logf("[+] Scratch layer ready.")
	return nil
}

func importOneTar(ctx context.Context, tarPath, destDir string) error {
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

	// nil parent paths = base layer
	_, err = ociwclayer.ImportLayerFromTar(ctx, gzr, destDir, nil)
	return err
}

// scratchLayerExists reports whether an already-prepared scratch exists and
// its layerchain.json references the given baseDir.
func scratchLayerExists(scratchDir, baseDir string) bool {
	data, err := os.ReadFile(filepath.Join(scratchDir, "layerchain.json"))
	if err != nil {
		return false
	}
	var chain []string
	if err := json.Unmarshal(data, &chain); err != nil || len(chain) == 0 {
		return false
	}
	return strings.EqualFold(chain[0], baseDir)
}