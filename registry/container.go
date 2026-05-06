package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/carbon-os/compute-image/internal/logf"
)

// LayerDescriptor describes a single image layer from a manifest.
type LayerDescriptor struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
}

type manifestList struct {
	Manifests []struct {
		Digest   string `json:"digest"`
		Platform struct {
			Architecture string `json:"architecture"`
			Os           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}

type imageManifest struct {
	Layers []LayerDescriptor `json:"layers"`
}

type tokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

// ParseImageRef splits "registry/repo:tag" into its components.
// Registry defaults to "index.docker.io" if omitted.
func ParseImageRef(image string) (registry, repo, tag string, err error) {
	tag = "latest"
	if i := strings.LastIndex(image, ":"); i > strings.LastIndex(image, "/") {
		tag = image[i+1:]
		image = image[:i]
	}
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 2 && strings.ContainsAny(parts[0], ".:") {
		registry = parts[0]
		repo = parts[1]
	} else {
		registry = "index.docker.io"
		repo = image
	}
	if registry == "" || repo == "" {
		err = fmt.Errorf("invalid image ref %q", image)
	}
	return
}

// FetchManifest fetches the raw manifest bytes for registry/repo:ref,
// handling token auth transparently.
func FetchManifest(registry, repo, ref string) ([]byte, error) {
	return fetchManifestWithAuth(registry, repo, ref, "")
}

// ResolveWindowsDigest resolves an amd64/windows platform digest from a
// manifest list. Falls back to fallback if not a manifest list or no match.
func ResolveWindowsDigest(manifestBytes []byte, fallback string) string {
	return resolvePlatformDigest(manifestBytes, "amd64", "windows", fallback)
}

// ResolveLinuxDigest resolves an amd64/linux platform digest from a manifest list.
func ResolveLinuxDigest(manifestBytes []byte, fallback string) string {
	return resolvePlatformDigest(manifestBytes, "amd64", "linux", fallback)
}

func resolvePlatformDigest(manifestBytes []byte, arch, os_, fallback string) string {
	var ml manifestList
	if json.Unmarshal(manifestBytes, &ml) == nil {
		for _, m := range ml.Manifests {
			if m.Platform.Architecture == arch && m.Platform.Os == os_ {
				return m.Digest
			}
		}
	}
	return fallback
}

// ParseManifestLayers returns the layer descriptors from an image manifest.
func ParseManifestLayers(manifestBytes []byte) ([]LayerDescriptor, error) {
	var m imageManifest
	if err := json.Unmarshal(manifestBytes, &m); err != nil {
		return nil, err
	}
	var out []LayerDescriptor
	for _, l := range m.Layers {
		if strings.Contains(l.MediaType, "tar") || strings.Contains(l.MediaType, "layer") {
			out = append(out, l)
		}
	}
	return out, nil
}

// DownloadLayers downloads each layer to cacheDir and returns the local paths.
func DownloadLayers(reg, repo string, layers []LayerDescriptor, cacheDir string) ([]string, error) {
	var paths []string
	for i, layer := range layers {
		digestHex := strings.TrimPrefix(layer.Digest, "sha256:")
		cachedPath := filepath.Join(cacheDir, digestHex+".tar.gz")
		logf.Logf("    -> Layer %d/%d: %s", i+1, len(layers), layer.Digest)

		if IsCacheValid(cachedPath) {
			logf.Logf("       [CACHE HIT]")
		} else {
			logf.Logf("       [CACHE MISS] Downloading...")
			if err := downloadBlob(reg, repo, layer.Digest, cachedPath, layer.Size); err != nil {
				return nil, fmt.Errorf("registry: download layer %s: %w", layer.Digest, err)
			}
			fmt.Println()
			if err := VerifyDigest(cachedPath, layer.Digest); err != nil {
				os.Remove(cachedPath)
				return nil, fmt.Errorf("registry: digest mismatch %s: %w", layer.Digest, err)
			}
			logf.Logf("       [+] Digest OK")
		}
		paths = append(paths, cachedPath)
	}
	return paths, nil
}

func fetchManifestWithAuth(reg, repo, ref, token string) ([]byte, error) {
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", reg, repo, ref)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return io.ReadAll(resp.Body)
	}
	if resp.StatusCode == 401 && token == "" {
		if auth := resp.Header.Get("Www-Authenticate"); auth != "" {
			tok, err := performTokenHandshake(auth)
			if err != nil {
				return nil, err
			}
			return fetchManifestWithAuth(reg, repo, ref, tok)
		}
	}
	return nil, fmt.Errorf("manifest %s: HTTP %s", ref, resp.Status)
}

func performTokenHandshake(authHeader string) (string, error) {
	var realm, service, scope string
	for _, part := range strings.Split(authHeader, ",") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := kv[0], strings.Trim(kv[1], "\"")
		switch {
		case strings.HasSuffix(key, "realm"):
			realm = val
		case key == "service":
			service = val
		case key == "scope":
			scope = val
		}
	}
	resp, err := http.Get(fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var t tokenResponse
	json.NewDecoder(resp.Body).Decode(&t)
	if t.Token != "" {
		return t.Token, nil
	}
	return t.AccessToken, nil
}

func downloadBlob(reg, repo, digest, destPath string, expectedSize int64) error {
	resp, err := http.Get(fmt.Sprintf("https://%s/v2/%s/blobs/%s", reg, repo, digest))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	total := expectedSize
	if total == 0 {
		total = resp.ContentLength
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, &ProgressReader{R: resp.Body, Total: total})
	return err
}