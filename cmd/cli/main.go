package main

import (
	"flag"
	"fmt"
	"os"

	compute_image "github.com/carbon-os/compute-image"
)

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	switch os.Args[1] {
	case "pull":
		runPull(os.Args[2], os.Args[3:])
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
		base := fs.String("base", "", "destination for the base layer (required)")
		scratch := fs.String("scratch", "", "destination for the scratch layer (required)")
		cache := fs.String("cache", "", "cache directory (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 || *base == "" || *scratch == "" {
			fmt.Fprintln(os.Stderr, "usage: image-cli pull container <image> --base <path> --scratch <path> [--cache <path>]")
			os.Exit(1)
		}

		img, err := compute_image.Pull(compute_image.ContainerRef{
			Image:   fs.Arg(0),
			BaseDir: *base,
			Scratch: *scratch,
			Cache:   *cache,
		})
		if err != nil {
			fatal(err)
		}
		result := img.(*compute_image.ContainerImage)
		fmt.Printf("\n[+] Base:    %s\n", result.BaseLayer)
		fmt.Printf("[+] Scratch: %s\n", result.Scratch)

	case "vm":
		fs := flag.NewFlagSet("pull vm", flag.ExitOnError)
		registry := fs.String("registry", "", "VM image registry hostname (required)")
		out := fs.String("out", "", "output path for the prepared disk image (required)")
		cache := fs.String("cache", "", "cache directory (optional)")
		fs.Parse(args)

		if fs.NArg() < 1 || *registry == "" || *out == "" {
			fmt.Fprintln(os.Stderr, "usage: image-cli pull vm <image> --registry <host> --out <path> [--cache <path>]")
			os.Exit(1)
		}

		img, err := compute_image.Pull(compute_image.VMRef{
			Image:    fs.Arg(0),
			Registry: *registry,
			Out:      *out,
			Cache:    *cache,
		})
		if err != nil {
			fatal(err)
		}
		result := img.(*compute_image.VMImage)
		fmt.Printf("\n[+] VM image: %s\n", result.OutPath)

	default:
		fmt.Fprintf(os.Stderr, "unknown image type %q — expected 'container' or 'vm'\n", imageType)
		os.Exit(1)
	}
}

func runLs(args []string) {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	cache := fs.String("cache", "", "cache directory (optional)")
	fs.Parse(args)

	cacheDir := compute_image.ResolveCache(*cache)
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fmt.Printf("(empty — %s not found)\n", cacheDir)
		return
	}
	fmt.Printf("Cache: %s\n\n", cacheDir)
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
	cache := fs.String("cache", "", "cache directory (optional)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: image-cli rm <image> [--cache <path>]")
		os.Exit(1)
	}

	// rm by image ref is a best-effort prefix match on cached filenames
	cacheDir := compute_image.ResolveCache(*cache)
	ref := fs.Arg(0)
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fatal(fmt.Errorf("cache not found: %s", cacheDir))
	}

	removed := 0
	for _, e := range entries {
		if matchesRef(e.Name(), ref) {
			path := cacheDir + string(os.PathSeparator) + e.Name()
			if err := os.Remove(path); err != nil {
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

// matchesRef is a loose match: the cache filename contains any component of
// the image ref (name, version, or digest prefix).
func matchesRef(filename, ref string) bool {
	for _, part := range []string{ref, extractName(ref), extractVersion(ref)} {
		if part != "" && contains(filename, part) {
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

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) &&
		(s == sub || len(s) > 0 && stringContains(s, sub))
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
  image-cli pull container <image> --base <path> --scratch <path> [--cache <path>]
  image-cli pull vm        <image> --registry <host> --out <path> [--cache <path>]
  image-cli ls             [--cache <path>]
  image-cli rm             <image> [--cache <path>]`)
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[-] %v\n", err)
	os.Exit(1)
}