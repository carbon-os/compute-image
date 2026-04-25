package compute_image

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// fetchVMImage downloads the qcow2 image for the given ref from the registry,
// caches it, and returns the local path.
func fetchVMImage(image, registry, cacheDir string) (string, error) {
	name, version, err := parseVMRef(image)
	if err != nil {
		return "", err
	}

	url := buildVMImageURL(registry, name, version)
	fileName := fmt.Sprintf("%s-%s.qcow2", name, version)
	cachedPath := filepath.Join(cacheDir, fileName)

	if isCacheValid(cachedPath) {
		logf("    [CACHE HIT] %s", fileName)
		return cachedPath, nil
	}

	logf("    [CACHE MISS] Downloading %s...", url)
	if err := downloadFile(url, cachedPath); err != nil {
		os.Remove(cachedPath)
		return "", fmt.Errorf("compute-image: download vm image: %w", err)
	}
	fmt.Println()
	logf("    [+] Download complete.")
	return cachedPath, nil
}

// parseVMRef splits "ubuntu:22.04" into ("ubuntu", "22.04").
func parseVMRef(image string) (name, version string, err error) {
	parts := strings.SplitN(image, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid vm image ref %q, expected name:version", image)
	}
	return parts[0], parts[1], nil
}

// defaultVMRegistry returns a well-known registry for common VM image names.
// Returns an error if the name is not recognized, so the caller can ask the
// user to supply --registry explicitly.
func defaultVMRegistry(name string) (string, error) {
	switch name {
	case "ubuntu":
		return "cloud-images.ubuntu.com", nil
	case "debian":
		return "cloud.debian.org", nil
	default:
		return "", fmt.Errorf("no default registry for %q — supply --registry", name)
	}
}

// buildVMImageURL constructs the download URL for a well-known VM image.
// Extend this as more registries/distros are supported.
func buildVMImageURL(registry, name, version string) string {
	switch name {
	case "ubuntu":
		// e.g. https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
		return fmt.Sprintf("https://%s/releases/%s/release/%s-%s-server-cloudimg-amd64.img",
			registry, version, name, version)
	default:
		// Generic fallback: registry/name/version/name-version.qcow2
		return fmt.Sprintf("https://%s/%s/%s/%s-%s.qcow2", registry, name, version, name, version)
	}
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, &progressReader{r: resp.Body, total: resp.ContentLength})
	return err
}