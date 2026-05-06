package alpine

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

func BuildURL(reg, version, arch string) string {
	return fmt.Sprintf(downloadPath, reg, majorMinor(version), version, toAlpineArch(arch))
}

// Validate checks that version is a patch release of a supported Alpine branch
// and that arch is one of the supported canonical names.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("alpine: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	branch := majorMinor(version)
	if !ValidBranches[branch] {
		branches := "3.20, 3.21, 3.22, 3.23"
		return fmt.Errorf("alpine: unsupported version %q (branch %q) — valid branches: %s", version, branch, branches)
	}
	return nil
}

// majorMinor extracts the "major.minor" prefix from a version string.
// "3.23.4" → "3.23", "3.22.1" → "3.22".
// Falls back to the full version string if it has fewer than two dots.
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

// majorVersion is kept for backward compatibility with any callers that used it.
// Prefer majorMinor.
func majorVersion(v string) string { return majorMinor(v) }

// toAlpineArch alias — suppress "declared and not used" if old callers used
// the unexported name; the exported name is now toAlpineArch above.
var _ = strings.Contains // ensure import used if needed by future code