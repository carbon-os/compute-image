package registry

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile fetches url and writes it to destPath with progress output.
func DownloadFile(url, destPath string) error {
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
	_, err = io.Copy(out, &ProgressReader{R: resp.Body, Total: resp.ContentLength})
	return err
}