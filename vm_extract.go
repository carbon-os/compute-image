package compute_image

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/partition/gpt"
	ext4 "github.com/masahiro331/go-ext4-filesystem/ext4"
)

// extractVM finds the Linux ext4 partition inside a raw disk image and
// extracts its contents to outDir.
func extractVM(rawPath, outDir string) error {
	offset, size, err := findExt4Partition(rawPath)
	if err != nil {
		return err
	}
	logf("    ext4 partition: offset=%d size=%d", offset, size)

	rawFile, err := os.Open(rawPath)
	if err != nil {
		return err
	}
	defer rawFile.Close()

	partReader := io.NewSectionReader(rawFile, offset, size)
	filesystem, err := ext4.NewFS(*partReader, nil)
	if err != nil {
		return fmt.Errorf("open ext4: %w", err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	var extracted, skipped int
	err = fs.WalkDir(filesystem, "/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logf("    [WARN] %s: %v", path, err)
			return nil
		}
		dst := filepath.Join(outDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		// Skip symlinks and device files
		if d.Type()&fs.ModeSymlink != 0 || d.Type()&fs.ModeDevice != 0 {
			skipped++
			return nil
		}
		src, err := filesystem.Open(path)
		if err != nil {
			skipped++
			return nil
		}
		defer src.Close()

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		f, err := os.Create(dst)
		if err != nil {
			skipped++
			return nil
		}
		defer f.Close()
		if _, err := io.Copy(f, src); err != nil {
			skipped++
			return nil
		}
		extracted++
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk ext4: %w", err)
	}
	logf("    extracted=%d skipped=%d", extracted, skipped)
	return nil
}

func findExt4Partition(rawPath string) (offset, size int64, err error) {
	d, err := diskfs.Open(rawPath, diskfs.WithSectorSize(512))
	if err != nil {
		return 0, 0, fmt.Errorf("open disk: %w", err)
	}
	table, err := d.GetPartitionTable()
	if err != nil {
		return 0, 0, fmt.Errorf("read partition table: %w", err)
	}
	gptTable, ok := table.(*gpt.Table)
	if !ok {
		return 0, 0, fmt.Errorf("not a GPT disk")
	}
	const sector = int64(512)
	for _, p := range gptTable.Partitions {
		if p.Start == 0 {
			continue
		}
		if p.Type == gpt.LinuxFilesystem {
			return int64(p.Start) * sector, int64(p.End-p.Start+1) * sector, nil
		}
	}
	return 0, 0, fmt.Errorf("no Linux filesystem partition found")
}