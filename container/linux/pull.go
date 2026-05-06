package linux

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/carbon-os/compute-image/internal/logf"
	"github.com/carbon-os/compute-image/registry"
)

// Pull downloads and unpacks a Linux container image described by ref.
// Layers are extracted into numbered subdirectories under Paths.Layers.
// The caller is responsible for mounting overlayfs if a writable view is needed.
func Pull(ref Ref) (*Image, error) {
	ref.Dir = resolveDir(ref.Dir)
	paths, err := resolvePaths(ref)
	if err != nil {
		return nil, fmt.Errorf("container/linux: resolve paths: %w", err)
	}

	for _, d := range []string{paths.Cache, paths.Layers, paths.Upper, paths.Work} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("container/linux: create dir %s: %w", d, err)
		}
	}

	reg, repo, tag, err := registry.ParseImageRef(ref.Image)
	if err != nil {
		return nil, fmt.Errorf("container/linux: parse ref: %w", err)
	}

	logf.Logf("[*] Pulling Linux container image: %s", ref.Image)
	logf.Logf("    Dir: %s", paths.Dir)

	manifestBytes, err := registry.FetchManifest(reg, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("container/linux: fetch manifest: %w", err)
	}
	digest := registry.ResolveLinuxDigest(manifestBytes, tag)
	manifestBytes, err = registry.FetchManifest(reg, repo, digest)
	if err != nil {
		return nil, fmt.Errorf("container/linux: fetch image manifest: %w", err)
	}
	layers, err := registry.ParseManifestLayers(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("container/linux: parse manifest: %w", err)
	}
	logf.Logf("[+] Manifest resolved. %d layer(s).", len(layers))

	cachedTars, err := registry.DownloadLayers(reg, repo, layers, paths.Cache)
	if err != nil {
		return nil, err
	}

	rootfs, err := unpackLayers(cachedTars, paths.Layers)
	if err != nil {
		return nil, err
	}

	logf.Logf("[+] Container image ready.")
	logf.Logf("    RootFS: %s", rootfs)

	return &Image{
		Image:  ref.Image,
		Paths:  paths,
		RootFS: rootfs,
	}, nil
}

// ResolvePaths returns the fully resolved paths for ref without pulling anything.
func ResolvePaths(ref Ref) (Paths, error) {
	ref.Dir = resolveDir(ref.Dir)
	return resolvePaths(ref)
}

// unpackLayers extracts each tar into a numbered subdirectory of layersDir
// and returns the path of the topmost layer directory.
func unpackLayers(tarPaths []string, layersDir string) (string, error) {
	logf.Logf("[*] Unpacking %d layer(s) into: %s", len(tarPaths), layersDir)

	var layerDirs []string
	for i, tarPath := range tarPaths {
		layerDir := filepath.Join(layersDir, fmt.Sprintf("%02d", i))
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return "", fmt.Errorf("container/linux: create layer dir: %w", err)
		}
		logf.Logf("    -> [%d/%d] %s", i+1, len(tarPaths), filepath.Base(tarPath))
		if err := extractTar(tarPath, layerDir); err != nil {
			return "", fmt.Errorf("container/linux: extract %s: %w", filepath.Base(tarPath), err)
		}
		layerDirs = append(layerDirs, layerDir)
	}
	return layerDirs[len(layerDirs)-1], nil
}

// extractTar decompresses and extracts a .tar.gz into destDir.
// OCI whiteout files (.wh.*) are skipped — they are applied by overlayfs at
// runtime when the layers are mounted, not during extraction.
func extractTar(tarPath, destDir string) error {
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

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip OCI whiteout markers — overlayfs handles these at mount time.
		if strings.HasPrefix(filepath.Base(hdr.Name), ".wh.") {
			continue
		}

		target := filepath.Join(destDir, filepath.Clean("/"+hdr.Name))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil && !os.IsExist(err) {
				return err
			}
		case tar.TypeLink:
			linkTarget := filepath.Join(destDir, filepath.Clean("/"+hdr.Linkname))
			os.Remove(target)
			if err := os.Link(linkTarget, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolvePaths(ref Ref) (Paths, error) {
	reg, repo, tag, err := registry.ParseImageRef(ref.Image)
	if err != nil {
		return Paths{}, err
	}
	imageDir := filepath.Join(ref.Dir, reg, filepath.FromSlash(repo), tag)
	return Paths{
		Dir:    imageDir,
		Layers: filepath.Join(imageDir, "layers"),
		Upper:  filepath.Join(imageDir, "upper"),
		Work:   filepath.Join(imageDir, "work"),
		Cache:  filepath.Join(ref.Dir, "cache"),
	}, nil
}

func resolveDir(dir string) string {
	if dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "carbon")
	}
	return ".carbon"
}