package main

import (
	"flag"
	"fmt"
	"os"

	compute_image "github.com/carbon-os/compute-image"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "pull":
		if len(os.Args) < 3 {
			usage()
		}
		runPull(os.Args[2], os.Args[3:])
	case "info":
		if len(os.Args) < 3 {
			usage()
		}
		runInfo(os.Args[2], os.Args[3:])
	case "ls":
		runLs(os.Args[2:])
	case "rm":
		runRm(os.Args[2:])
	default:
		usage()
	}
}

func runPull(imageType string, args []string) {
	switch imageType {
	case "container":
		fs := flag.NewFlagSet("pull container", flag.ExitOnError)
		dir := fs.String("dir", "", "root directory for image data (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "usage: image-cli pull container <image> [--dir <path>]")
			os.Exit(1)
		}

		img, err := compute_image.Pull(compute_image.ContainerRef{
			Image: fs.Arg(0),
			Dir:   *dir,
		})
		if err != nil {
			fatal(err)
		}
		result := img.(*compute_image.ContainerImage)
		fmt.Printf("\n[+] Container image ready.\n")
		fmt.Printf("    Dir:     %s\n", result.Paths.Dir)
		fmt.Printf("    Base:    %s\n", result.Paths.Base)
		fmt.Printf("    Scratch: %s\n", result.Paths.Scratch)
		fmt.Printf("    Cache:   %s\n", result.Paths.Cache)

	case "vm":
		fs := flag.NewFlagSet("pull vm", flag.ExitOnError)
		registry := fs.String("registry", "", "VM image registry hostname (required)")
		dir := fs.String("dir", "", "root directory for image data (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 || *registry == "" {
			fmt.Fprintln(os.Stderr, "usage: image-cli pull vm <image> --registry <host> [--dir <path>]")
			os.Exit(1)
		}

		img, err := compute_image.Pull(compute_image.VMRef{
			Image:    fs.Arg(0),
			Registry: *registry,
			Dir:      *dir,
		})
		if err != nil {
			fatal(err)
		}
		result := img.(*compute_image.VMImage)
		fmt.Printf("\n[+] VM image ready.\n")
		fmt.Printf("    Dir:   %s\n", result.Paths.Dir)
		fmt.Printf("    Disk:  %s\n", result.Paths.Disk)
		fmt.Printf("    Cache: %s\n", result.Paths.Cache)

	default:
		fmt.Fprintf(os.Stderr, "unknown image type %q — expected 'container' or 'vm'\n", imageType)
		os.Exit(1)
	}
}

func runInfo(imageType string, args []string) {
	switch imageType {
	case "container":
		fs := flag.NewFlagSet("info container", flag.ExitOnError)
		dir := fs.String("dir", "", "root directory for image data (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "usage: image-cli info container <image> [--dir <path>]")
			os.Exit(1)
		}

		ref := compute_image.ContainerRef{Image: fs.Arg(0), Dir: *dir}
		paths, err := compute_image.ResolveContainerPaths(ref)
		if err != nil {
			fatal(err)
		}

		fmt.Printf("Image:   %s\n", fs.Arg(0))
		fmt.Printf("Dir:     %s\n", paths.Dir)
		fmt.Printf("Base:    %s\n", paths.Base)
		fmt.Printf("Scratch: %s\n", paths.Scratch)
		fmt.Printf("Cache:   %s\n", paths.Cache)
		fmt.Printf("Status:  %s\n", containerStatus(paths))

	case "vm":
		fs := flag.NewFlagSet("info vm", flag.ExitOnError)
		registry := fs.String("registry", "", "VM image registry hostname (required)")
		dir := fs.String("dir", "", "root directory for image data (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 || *registry == "" {
			fmt.Fprintln(os.Stderr, "usage: image-cli info vm <image> --registry <host> [--dir <path>]")
			os.Exit(1)
		}

		ref := compute_image.VMRef{Image: fs.Arg(0), Registry: *registry, Dir: *dir}
		paths, err := compute_image.ResolveVMPaths(ref)
		if err != nil {
			fatal(err)
		}

		fmt.Printf("Image:    %s\n", fs.Arg(0))
		fmt.Printf("Registry: %s\n", *registry)
		fmt.Printf("Dir:      %s\n", paths.Dir)
		fmt.Printf("Disk:     %s\n", paths.Disk)
		fmt.Printf("Cache:    %s\n", paths.Cache)
		fmt.Printf("Status:   %s\n", vmStatus(paths))

	default:
		fmt.Fprintf(os.Stderr, "unknown image type %q — expected 'container' or 'vm'\n", imageType)
		os.Exit(1)
	}
}

// containerStatus checks what's already on disk.
func containerStatus(paths compute_image.ContainerPaths) string {
	if dirExists(paths.Base) && dirExists(paths.Scratch) {
		return "ready"
	}
	if dirExists(paths.Cache) {
		return "cached"
	}
	return "not pulled"
}

// vmStatus checks what's already on disk.
func vmStatus(paths compute_image.VMPaths) string {
	if fileExists(paths.Disk) {
		return "ready"
	}
	if dirExists(paths.Cache) {
		return "cached"
	}
	return "not pulled"
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func runLs(args []string) {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	dir := fs.String("dir", "", "root directory for image data (optional)")
	fs.Parse(args)

	cacheDir := compute_image.ResolveContainerPaths // we just need the cache path
	_ = cacheDir

	// Resolve cache via a dummy ref — dir is the only thing that matters here
	paths, _ := compute_image.ResolveContainerPaths(compute_image.ContainerRef{Dir: *dir})
	entries, err := os.ReadDir(paths.Cache)
	if err != nil {
		fmt.Printf("(empty — %s not found)\n", paths.Cache)
		return
	}
	fmt.Printf("Cache: %s\n\n", paths.Cache)
	for _, e := range entries {
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		fmt.Printf("  %-70s  %s\n", e.Name(), compute_image.HumanBytes(size))
	}
}

func runRm(args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	dir := fs.String("dir", "", "root directory for image data (optional)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: image-cli rm <image> [--dir <path>]")
		os.Exit(1)
	}

	paths, _ := compute_image.ResolveContainerPaths(compute_image.ContainerRef{Dir: *dir})
	ref := fs.Arg(0)
	entries, err := os.ReadDir(paths.Cache)
	if err != nil {
		fatal(fmt.Errorf("cache not found: %s", paths.Cache))
	}

	removed := 0
	for _, e := range entries {
		if matchesRef(e.Name(), ref) {
			p := paths.Cache + string(os.PathSeparator) + e.Name()
			if err := os.Remove(p); err != nil {
				fmt.Fprintf(os.Stderr, "rm %s: %v\n", e.Name(), err)
			} else {
				fmt.Printf("removed: %s\n", e.Name())
				removed++
			}
		}
	}
	if removed == 0 {
		fmt.Printf("no cached files matched %q\n", ref)
	}
}

func matchesRef(filename, ref string) bool {
	for _, part := range []string{ref, extractName(ref), extractVersion(ref)} {
		if part != "" && stringContains(filename, part) {
			return true
		}
	}
	return false
}

func extractName(ref string) string {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			return ref[:i]
		}
	}
	return ref
}

func extractVersion(ref string) string {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			return ref[i+1:]
		}
	}
	return ""
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  image-cli pull container <image> [--dir <path>]
  image-cli pull vm        <image> --registry <host> [--dir <path>]
  image-cli info container <image> [--dir <path>]
  image-cli info vm        <image> --registry <host> [--dir <path>]
  image-cli ls             [--dir <path>]
  image-cli rm             <image> [--dir <path>]`)
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[-] %v\n", err)
	os.Exit(1)
}