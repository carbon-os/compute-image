// img-2-tar.go
// Extracts the filesystem from a bootable .img file into a .tar archive.
// Supports ext4 (via go-ext4) and FAT32/ISO9660 (via go-diskfs).
// No external tools, no root required.
//
// Usage:
//   go run img-2-tar.go <image.img> <output.tar> [options]
//
// Dependencies:
//   go get github.com/diskfs/go-diskfs
//   go get github.com/dsoprea/go-ext4

package main

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/filesystem"
	ext4lib "github.com/dsoprea/go-ext4"
)

// =============================================================================
// Partition offset helper
// =============================================================================

// getPartitionOffsetBytes returns the byte offset of a partition.
// go-diskfs GPT partitions return bytes from GetStart(); MBR returns sectors.
func getPartitionOffsetBytes(start int64, tableType string, sectorSize int64) int64 {
	if tableType == "gpt" {
		return start // already bytes
	}
	return start * sectorSize // MBR: sectors → bytes
}

// =============================================================================
// ext4 extraction (via go-ext4 + io.SectionReader at partition offset)
// =============================================================================

// extractExt4 walks an ext4 filesystem at the given byte offset inside imgFile
// and writes all entries into tw.
func extractExt4(tw *tar.Writer, imgFile *os.File, offset, partSize int64, verbose bool) error {
	sr := io.NewSectionReader(imgFile, offset, partSize)

	if _, err := sr.Seek(ext4lib.Superblock0Offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to superblock: %w", err)
	}

	sb, err := ext4lib.NewSuperblockWithReader(sr)
	if err != nil {
		return fmt.Errorf("read ext4 superblock: %w", err)
	}
	fmt.Printf("  ext4 volume: %q  block size: %d\n", sb.VolumeName(), sb.BlockSize())

	bgdl, err := ext4lib.NewBlockGroupDescriptorListWithReadSeeker(sr, sb)
	if err != nil {
		return fmt.Errorf("read block group descriptors: %w", err)
	}

	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "./",
		Mode:     0755,
	}); err != nil {
		return fmt.Errorf("tar root dir header: %w", err)
	}

	rootBgd, err := bgdl.GetWithAbsoluteInode(ext4lib.InodeRootDirectory)
	if err != nil {
		return fmt.Errorf("get root inode bgd: %w", err)
	}

	dw, err := ext4lib.NewDirectoryWalk(sr, rootBgd, ext4lib.InodeRootDirectory)
	if err != nil {
		return fmt.Errorf("create directory walk: %w", err)
	}

	for {
		fullPath, de, err := dw.Next()
		if err != nil {
			if errors.Is(err, io.EOF) || err.Error() == "EOF" {
				break
			}
			fmt.Fprintf(os.Stderr, "  [warn] walk error at %q: %v\n", fullPath, err)
			continue
		}

		tarName  := "./" + fullPath
		inodeNum := int(de.Data().Inode)

		bgd, err := bgdl.GetWithAbsoluteInode(inodeNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [warn] get bgd for inode %d (%s): %v\n", inodeNum, fullPath, err)
			continue
		}
		inode, err := ext4lib.NewInodeWithReadSeeker(bgd, sr, inodeNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [warn] read inode %d (%s): %v\n", inodeNum, fullPath, err)
			continue
		}

		mode    := os.FileMode(uint32(inode.Data().IMode) & 0xFFF)
		modTime := inode.ModificationTime()
		uid     := int(inode.Data().IUid)
		gid     := int(inode.Data().IGid)

		switch {
		case de.IsDirectory():
			hdr := &tar.Header{
				Typeflag: tar.TypeDir,
				Name:     tarName + "/",
				Mode:     int64(mode),
				ModTime:  modTime,
				Uid:      uid,
				Gid:      gid,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("tar dir header %q: %w", tarName, err)
			}
			if verbose {
				fmt.Printf("  d  %s/\n", fullPath)
			}

		case de.IsSymbolicLink():
			linkTarget, err := readExt4Symlink(sr, inode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [warn] symlink %q: %v\n", fullPath, err)
				continue
			}
			hdr := &tar.Header{
				Typeflag: tar.TypeSymlink,
				Name:     tarName,
				Linkname: linkTarget,
				Mode:     int64(mode),
				ModTime:  modTime,
				Uid:      uid,
				Gid:      gid,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("tar symlink header %q: %w", tarName, err)
			}
			if verbose {
				fmt.Printf("  l  %s -> %s\n", fullPath, linkTarget)
			}

		case de.IsRegular():
			size := int64(inode.Size())
			hdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     tarName,
				Size:     size,
				Mode:     int64(mode),
				ModTime:  modTime,
				Uid:      uid,
				Gid:      gid,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("tar file header %q: %w", tarName, err)
			}
			en := ext4lib.NewExtentNavigatorWithReadSeeker(sr, inode)
			ir := ext4lib.NewInodeReader(en)
			n, err := io.Copy(tw, io.LimitReader(ir, size))
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [warn] copy %q (%d/%d bytes): %v\n", fullPath, n, size, err)
			} else if verbose {
				fmt.Printf("  f  %s (%d bytes)\n", fullPath, n)
			}

		default:
			if verbose {
				fmt.Printf("  -  %s [%s — skipped]\n", fullPath, de.TypeName())
			}
		}
	}
	return nil
}

// readExt4Symlink reads the symlink target from an ext4 inode.
// Short symlinks (<=60 bytes) are stored inline in IBlock; longer ones use extents.
func readExt4Symlink(sr io.ReadSeeker, inode *ext4lib.Inode) (string, error) {
	size := inode.Size()
	if size == 0 {
		return "", nil
	}
	// Inline symlinks fit entirely inside the IBlock field (max 60 bytes).
	if size <= 60 {
		raw := inode.Data().IBlock[:]
		return strings.TrimRight(string(raw[:size]), "\x00"), nil
	}
	// Long symlink — data stored in extents.
	en := ext4lib.NewExtentNavigatorWithReadSeeker(sr, inode)
	ir := ext4lib.NewInodeReader(en)
	buf, err := io.ReadAll(io.LimitReader(ir, int64(size)))
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(buf), "\x00"), nil
}

// isExt4 checks the ext4 superblock magic at the given partition byte offset.
func isExt4(imgFile *os.File, offset, partSize int64) bool {
	sr := io.NewSectionReader(imgFile, offset, partSize)
	if _, err := sr.Seek(ext4lib.Superblock0Offset, io.SeekStart); err != nil {
		return false
	}
	_, err := ext4lib.NewSuperblockWithReader(sr)
	return err == nil
}

// =============================================================================
// FAT32 / ISO9660 extraction (via go-diskfs filesystem.FileSystem)
// =============================================================================

func walkFAT(tw *tar.Writer, fs filesystem.FileSystem, fsPath string, verbose bool) error {
	entries, err := fs.ReadDir(fsPath)
	if err != nil {
		return fmt.Errorf("readdir %q: %w", fsPath, err)
	}
	for _, entry := range entries {
		import_path := fsPath + "/" + entry.Name()
		if fsPath == "/" {
			import_path = "/" + entry.Name()
		}
		tarName := strings.TrimPrefix(import_path, "/")

		info, err := entry.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [warn] stat %q: %v\n", tarName, err)
			continue
		}
		mode := info.Mode()

		if info.IsDir() {
			hdr := &tar.Header{
				Typeflag: tar.TypeDir,
				Name:     tarName + "/",
				Mode:     int64(mode.Perm()),
				ModTime:  info.ModTime(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if verbose {
				fmt.Printf("  d  %s/\n", tarName)
			}
			if err := walkFAT(tw, fs, import_path, verbose); err != nil {
				return err
			}
		} else {
			hdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     tarName,
				Size:     info.Size(),
				Mode:     int64(mode.Perm()),
				ModTime:  info.ModTime(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			f, err := fs.OpenFile(import_path, os.O_RDONLY)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [warn] open %q: %v\n", tarName, err)
				continue
			}
			n, err := io.Copy(tw, f)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [warn] copy %q (%d bytes): %v\n", tarName, n, err)
			} else if verbose {
				fmt.Printf("  f  %s (%d bytes)\n", tarName, n)
			}
		}
	}
	return nil
}

// =============================================================================
// Entry point
// =============================================================================

func usage() {
	fmt.Fprintf(os.Stderr, `
img-2-tar — extract a bootable .img filesystem into a tar archive

Usage:
  img-2-tar <image.img> <output.tar> [options]

Options:
  -p <N>   Use partition number N (1-based). Default: auto (tries all, picks first ext4).
  -v       Verbose: print every file as it is added.
  -h       Show this help.

Examples:
  img-2-tar debian.img rootfs.tar
  img-2-tar ubuntu.img rootfs.tar -p 2 -v
`)
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}

	imgPath := args[0]
	tarPath := args[1]
	partNum := 0
	verbose := false

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "-v":
			verbose = true
		case "-h", "--help":
			usage()
			os.Exit(0)
		case "-p":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: -p requires a partition number")
				os.Exit(1)
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				fmt.Fprintf(os.Stderr, "Error: invalid partition number %q\n", args[i])
				os.Exit(1)
			}
			partNum = n
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", args[i])
			usage()
			os.Exit(1)
		}
	}

	// ── Open raw image file for SectionReader use (ext4 path) ─────────────
	imgFile, err := os.Open(imgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening image file: %v\n", err)
		os.Exit(1)
	}
	defer imgFile.Close()

	// ── Open via go-diskfs for partition table parsing ─────────────────────
	fmt.Printf("Opening image: %s\n", imgPath)
	disk, err := diskfs.Open(imgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening disk image: %v\n", err)
		os.Exit(1)
	}
	defer disk.Close()

	sectorSize := disk.LogicalBlocksize
	diskSize   := disk.Size

	// ── Parse partition table ──────────────────────────────────────────────
	table, err := disk.GetPartitionTable()
	if err != nil {
		// No partition table — try raw ext4
		fmt.Println("No partition table found — trying raw ext4...")
		if !isExt4(imgFile, 0, diskSize) {
			fmt.Fprintln(os.Stderr, "Not an ext4 filesystem either. Unsupported format.")
			os.Exit(1)
		}
		writeTar(tarPath, func(tw *tar.Writer) error {
			return extractExt4(tw, imgFile, 0, diskSize, verbose)
		})
		return
	}

	tableType := table.Type()
	parts     := table.GetPartitions()
	fmt.Printf("Partition table : %s\n", tableType)
	fmt.Printf("Sector size     : %d bytes\n", sectorSize)
	fmt.Printf("Partitions      : %d\n\n", len(parts))

	for i, p := range parts {
		offsetBytes := getPartitionOffsetBytes(p.GetStart(), tableType, sectorSize)
		fmt.Printf("  [%d] offset=0x%09X  raw_size=%s\n",
			i+1, offsetBytes, humanBytes(p.GetSize()))
	}
	fmt.Println()

	// ── Select partition ───────────────────────────────────────────────────
	selectedIdx    := -1
	selectedOffset := int64(0)
	selectedSize   := int64(0)

	if partNum > 0 {
		if partNum > len(parts) {
			fmt.Fprintf(os.Stderr, "Error: partition %d does not exist (max %d)\n", partNum, len(parts))
			os.Exit(1)
		}
		selectedIdx    = partNum - 1
		selectedOffset = getPartitionOffsetBytes(parts[selectedIdx].GetStart(), tableType, sectorSize)
		selectedSize   = parts[selectedIdx].GetSize()
		fmt.Printf("Using partition %d (offset 0x%X)\n", partNum, selectedOffset)
	} else {
		// Auto-detect: find first ext4 partition
		for i, p := range parts {
			off  := getPartitionOffsetBytes(p.GetStart(), tableType, sectorSize)
			size := p.GetSize()
			if isExt4(imgFile, off, size) {
				selectedIdx    = i
				selectedOffset = off
				selectedSize   = size
				fmt.Printf("Auto-selected partition %d (ext4 detected at offset 0x%X)\n", i+1, off)
				break
			}
			fmt.Printf("  Partition %d: not ext4\n", i+1)
		}

		// Fallback: try go-diskfs FAT32 on each partition
		if selectedIdx < 0 {
			fmt.Println("No ext4 partition found — trying FAT32 via go-diskfs...")
			for i := range parts {
				n := i + 1
				fs, ferr := disk.GetFilesystem(n)
				if ferr != nil {
					fmt.Printf("  Partition %d: not readable (%v)\n", n, ferr)
					continue
				}
				fmt.Printf("Auto-selected partition %d (FAT/ISO)\n", n)
				writeTar(tarPath, func(tw *tar.Writer) error {
					return walkFAT(tw, fs, "/", verbose)
				})
				return
			}
			fmt.Fprintln(os.Stderr, "No readable filesystem found in any partition.")
			os.Exit(1)
		}
	}

	// ── Extract chosen partition ───────────────────────────────────────────
	fmt.Printf("Extracting to: %s\n", tarPath)
	writeTar(tarPath, func(tw *tar.Writer) error {
		return extractExt4(tw, imgFile, selectedOffset, selectedSize, verbose)
	})
}

// writeTar creates the output tar file and calls fn to populate it.
func writeTar(tarPath string, fn func(*tar.Writer) error) {
	out, err := os.Create(tarPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tar: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	tw := tar.NewWriter(out)
	if err := fn(tw); err != nil {
		fmt.Fprintf(os.Stderr, "Error during extraction: %v\n", err)
		os.Exit(1)
	}
	if err := tw.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing tar: %v\n", err)
		os.Exit(1)
	}

	fi, _ := os.Stat(tarPath)
	fmt.Printf("\nDone. Output: %s (%s)\n", tarPath, humanBytes(fi.Size()))
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}