package alpine

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the Alpine nocloud qcow2 download URL.
// If version is a bare branch (e.g. "3.20"), it is resolved to the latest
// known patch release via LatestPatch before building the URL.
func BuildURL(reg, version, arch string) string {
	version = resolve(version)
	a := toAlpineArch(arch)
	fw := toAlpineFirmware(arch)
	return fmt.Sprintf(downloadPath, reg, majorMinor(version), version, a, fw)
}

// Validate checks arch and version. Bare branch refs are accepted and
// silently resolved to their latest patch release.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("alpine: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	branch := majorMinor(version)
	if !ValidBranches[branch] {
		return fmt.Errorf(
			"alpine: unsupported version %q — valid versions: 3.20, 3.21, 3.22, 3.23, 3.24",
			version,
		)
	}
	return nil
}

// resolve expands a bare branch ref to its latest known patch release.
// "3.21" → "3.21.7".  Full versions pass through unchanged.
func resolve(version string) string {
	if !hasPatch(version) {
		if patch, ok := LatestPatch[version]; ok {
			return patch
		}
	}
	return version
}

// majorMinor extracts "major.minor" from a version string.
// "3.23.4" → "3.23". Falls back to the full string if fewer than two dots.
func majorMinor(v string) string {
	dots := 0
	for i, c := range v {
		if c == '.' {
			dots++
			if dots == 2 {
				return v[:i]
			}
		}
	}
	return v
}

// hasPatch returns true when the version contains at least two dots.
func hasPatch(v string) bool {
	return strings.Count(v, ".") >= 2
}

func toAlpineArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}

// toAlpineFirmware returns the firmware segment for the image filename.
// x86_64 images are bios; aarch64 images are uefi-only.
func toAlpineFirmware(arch string) string {
	if arch == "arm64" {
		return "uefi"
	}
	return "bios"
}