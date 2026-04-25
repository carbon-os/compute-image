package compute_image

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type layerDescriptor struct {
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
	Layers []layerDescriptor `json:"layers"`
}

type tokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

// parseImageRef splits "registry/repo:tag" into its components.
// Registry defaults to "index.docker.io" if omitted.
func parseImageRef(image string) (registry, repo, tag string, err error) {
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

func resolveDigest(manifestBytes []byte, fallback string) string {
	var ml manifestList
	if json.Unmarshal(manifestBytes, &ml) == nil {
		for _, m := range ml.Manifests {
			if m.Platform.Architecture == "amd64" && m.Platform.Os == "windows" {
				return m.Digest
			}
		}
	}
	return fallback
}

func parseManifestLayers(manifestBytes []byte) ([]layerDescriptor, error) {
	var m imageManifest
	if err := json.Unmarshal(manifestBytes, &m); err != nil {
		return nil, err
	}
	var out []layerDescriptor
	for _, l := range m.Layers {
		if strings.Contains(l.MediaType, "tar") || strings.Contains(l.MediaType, "layer") {
			out = append(out, l)
		}
	}
	return out, nil
}

func fetchManifestWithAuth(registry, repo, ref, token string) ([]byte, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repo, ref)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
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
			return fetchManifestWithAuth(registry, repo, ref, tok)
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

func downloadLayers(registry, repo string, layers []layerDescriptor, cacheDir string) ([]string, error) {
	var paths []string
	for i, layer := range layers {
		digestHex := strings.TrimPrefix(layer.Digest, "sha256:")
		cachedPath := filepath.Join(cacheDir, digestHex+".tar.gz")
		logf("    -> Layer %d/%d: %s", i+1, len(layers), layer.Digest)

		if isCacheValid(cachedPath) {
			logf("       [CACHE HIT]")
		} else {
			logf("       [CACHE MISS] Downloading...")
			if err := downloadBlob(registry, repo, layer.Digest, cachedPath, layer.Size); err != nil {
				return nil, fmt.Errorf("compute-image: download layer %s: %w", layer.Digest, err)
			}
			fmt.Println()
			if err := verifyDigest(cachedPath, layer.Digest); err != nil {
				os.Remove(cachedPath)
				return nil, fmt.Errorf("compute-image: digest mismatch %s: %w", layer.Digest, err)
			}
			logf("       [+] Digest OK")
		}
		paths = append(paths, cachedPath)
	}
	return paths, nil
}

func downloadBlob(registry, repo, digest, destPath string, expectedSize int64) error {
	resp, err := http.Get(fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, repo, digest))
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
	_, err = io.Copy(out, &progressReader{r: resp.Body, total: total})
	return err
}

func isCacheValid(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}

func verifyDigest(path, digest string) error {
	expected := strings.TrimPrefix(digest, "sha256:")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != expected {
		return fmt.Errorf("want %s got %s", expected, got)
	}
	return nil
}

// progressReader wraps an io.Reader and prints a progress bar.
type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	lastTick time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.current += int64(n)
	if time.Since(pr.lastTick) > 100*time.Millisecond || err == io.EOF {
		pr.lastTick = time.Now()
		pr.printBar()
	}
	return n, err
}

func (pr *progressReader) printBar() {
	const width = 40
	pct := float64(pr.current) / float64(pr.total) * 100
	if pr.total <= 0 {
		pct = 0
	}
	filled := int(pct / 100 * width)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
	fmt.Printf("\r    [%s] %.1f%% (%d MB)", bar, pct, pr.current/1024/1024)
}