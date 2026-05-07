package alpine

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the Alpine nocloud qcow2 download URL.
//
// version must be a full patch string such as "3.20.10" or "3.23.4".
// Bare branch refs like "3.21" have no corresponding file on the mirror.
//
// Firmware is arch-dependent:
//   - x86_64  → bios  (nocloud_alpine-{v}-x86_64-bios-tiny-r0.qcow2)
//   - aarch64 → uefi  (nocloud_alpine-{v}-aarch64-uefi-tiny-r0.qcow2)
func BuildURL(reg, version, arch string) string {
	a := toAlpineArch(arch)
	fw := toAlpineFirmware(arch)
	return fmt.Sprintf(downloadPath, reg, majorMinor(version), version, a, fw)
}

// Validate checks that:
//  1. arch is supported
//  2. version is a full patch string (e.g. "3.21.7"), not a bare branch ("3.21")
//  3. the version's branch is a supported Alpine release
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("alpine: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	// Require a full patch version — bare branch refs produce 404s.
	if !hasPatch(version) {
		return fmt.Errorf(
			"alpine: version %q is a branch ref, not a patch release — "+
				"supply the full version, e.g. %q",
			version, version+".0",
		)
	}
	branch := majorMinor(version)
	if !ValidBranches[branch] {
		return fmt.Errorf(
			"alpine: unsupported version %q (branch %q) — valid branches: 3.20, 3.21, 3.22, 3.23, 3.24",
			version, branch,
		)
	}
	return nil
}

// majorMinor extracts "major.minor" from a version string.
// "3.23.4" → "3.23".  Falls back to the full string if fewer than two dots.
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

// hasPatch returns true when the version string contains at least two dots,
// i.e. it is a full patch release like "3.21.7" rather than a branch "3.21".
func hasPatch(v string) bool {
	return strings.Count(v, ".") >= 2
}

// toAlpineArch maps canonical arch names to Alpine's filename convention.
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

// toAlpineFirmware returns the firmware segment for the nocloud image filename.
// x86_64 images use BIOS; aarch64 images are UEFI-only — the bios variant
// simply does not exist for arm64.
func toAlpineFirmware(arch string) string {
	if arch == "arm64" {
		return "uefi"
	}
	return "bios"
}