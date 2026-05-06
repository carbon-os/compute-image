package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// IsCacheValid reports whether path exists and is non-empty.
func IsCacheValid(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}

// VerifyDigest checks the sha256 of path against a "sha256:<hex>" digest string.
func VerifyDigest(path, digest string) error {
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
		return fmt.Errorf("digest mismatch: want %s got %s", expected, got)
	}
	return nil
}