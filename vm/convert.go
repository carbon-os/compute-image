package vm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/carbon-os/compute-image/registry"
	qcow2reader "github.com/lima-vm/go-qcow2reader"
)

func convertQcow2ToRaw(srcPath, dstPath string) error {
	if !isQcow2(srcPath) {
		return fmt.Errorf("%s is not a valid QCOW2 image", srcPath)
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := qcow2reader.Open(f)
	if err != nil {
		return fmt.Errorf("parse qcow2: %w", err)
	}
	size := img.Size()
	fmt.Printf("    Virtual size: %s\n", registry.HumanBytes(size))

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(dstPath)
		}
	}()

	return copyWithProgress(out, io.NewSectionReader(img, 0, size), size)
}

// convertRawToVHD converts a flat raw disk image to a fixed VHD by copying
// the raw bytes and appending a valid 512-byte VHD footer — pure Go, no exec.
func convertRawToVHD(rawPath, vhdPath string) error {
	os.Remove(vhdPath)
	if err := copyFile(rawPath, vhdPath); err != nil {
		return fmt.Errorf("copy raw to vhd: %w", err)
	}
	f, err := os.OpenFile(vhdPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open vhd for footer: %w", err)
	}
	defer f.Close()
	if err := writeVHDFooter(f); err != nil {
		os.Remove(vhdPath)
		return fmt.Errorf("write vhd footer: %w", err)
	}
	return nil
}

// writeVHDFooter appends a valid 512-byte fixed-VHD footer per the VHD spec.
func writeVHDFooter(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	diskSize := uint64(info.Size()) // size before the footer is the disk size
	cyl, heads, spt := vhdCHS(diskSize)

	var footer [512]byte
	copy(footer[0:8], "conectix")
	binary.BigEndian.PutUint32(footer[8:12], 0x00000002)
	binary.BigEndian.PutUint32(footer[12:16], 0x00010000)
	binary.BigEndian.PutUint64(footer[16:24], 0xFFFFFFFFFFFFFFFF) // fixed VHD data offset

	epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	binary.BigEndian.PutUint32(footer[24:28], uint32(time.Now().UTC().Sub(epoch).Seconds()))

	copy(footer[28:32], "go  ")
	binary.BigEndian.PutUint32(footer[32:36], 0x00010000)
	copy(footer[36:40], "Wi2k")

	binary.BigEndian.PutUint64(footer[40:48], diskSize)
	binary.BigEndian.PutUint64(footer[48:56], diskSize)
	binary.BigEndian.PutUint16(footer[56:58], uint16(cyl))
	footer[58] = byte(heads)
	footer[59] = byte(spt)
	binary.BigEndian.PutUint32(footer[60:64], 0x00000002) // disk type: fixed

	// UUID v4
	var uid [16]byte
	if _, err := rand.Read(uid[:]); err != nil {
		return err
	}
	uid[6] = (uid[6] & 0x0f) | 0x40
	uid[8] = (uid[8] & 0x3f) | 0x80
	copy(footer[68:84], uid[:])

	// Ones-complement checksum over all bytes except the checksum field itself.
	var sum uint32
	for i, b := range footer {
		if i < 64 || i >= 68 {
			sum += uint32(b)
		}
	}
	binary.BigEndian.PutUint32(footer[64:68], ^sum)

	_, err = f.Write(footer[:])
	return err
}

// vhdCHS computes cylinder/head/sector geometry per the VHD spec appendix.
func vhdCHS(diskSize uint64) (cylinders, heads, sectorsPerTrack uint32) {
	totalSectors := diskSize / 512
	const maxSectors = 65535 * 16 * 255
	if totalSectors > maxSectors {
		totalSectors = maxSectors
	}
	var cth uint64
	if totalSectors >= 65535*16*63 {
		sectorsPerTrack = 255
		heads = 16
		cth = totalSectors / uint64(sectorsPerTrack)
	} else {
		sectorsPerTrack = 17
		cth = totalSectors / uint64(sectorsPerTrack)
		heads = uint32((cth + 1023) / 1024)
		if heads < 4 {
			heads = 4
		}
		if cth >= uint64(heads)*1024 || heads > 16 {
			sectorsPerTrack = 31
			heads = 16
			cth = totalSectors / uint64(sectorsPerTrack)
		}
		if cth >= uint64(heads)*1024 {
			sectorsPerTrack = 63
			heads = 16
			cth = totalSectors / uint64(sectorsPerTrack)
		}
	}
	cylinders = uint32(cth / uint64(heads))
	return
}

func isQcow2(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return false
	}
	return magic[0] == 'Q' && magic[1] == 'F' && magic[2] == 'I' && magic[3] == 0xfb
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyWithProgress(dst io.Writer, src io.Reader, total int64) error {
	const chunkSize = 4 * 1024 * 1024
	buf := make([]byte, chunkSize)
	var written int64
	start := time.Now()

	for {
		n, err := src.Read(buf)
		if n > 0 {
			nw, werr := dst.Write(buf[:n])
			written += int64(nw)
			if werr != nil {
				return werr
			}
			elapsed := time.Since(start).Seconds()
			speed := float64(written) / elapsed / 1024 / 1024
			fmt.Printf("\r    %.1f%%  %s / %s  (%.1f MB/s)   ",
				float64(written)/float64(total)*100,
				registry.HumanBytes(written), registry.HumanBytes(total), speed)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}