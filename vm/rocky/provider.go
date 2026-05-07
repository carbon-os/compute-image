package rocky

import (
	"fmt"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the GenericCloud qcow2 download URL.
// version can be a major ("9") or a full patch version ("9.7") — only the
// major segment is used since Rocky publishes a stable .latest. symlink.
//
// Rocky 10 changed the image naming convention: the bare "GenericCloud" name
// was replaced with "GenericCloud-Base" and "GenericCloud-LVM" variants.
// We always target -Base as the closest equivalent to the pre-10 image.
func BuildURL(reg, version, arch string) string {
	ra := toRockyArch(arch)
	major := majorVersion(version)
	path := downloadPath
	if major == "10" {
		path = downloadPathV10
	}
	return fmt.Sprintf(path, reg, major, ra, major, ra)
}

// Validate checks that version resolves to a supported Rocky Linux major release
// and that arch is valid. Both bare majors ("9") and patch strings ("9.7") are accepted.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("rocky: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	major := majorVersion(version)
	if !ValidMajors[major] {
		return fmt.Errorf(
			"rocky: unsupported version %q (major %q) — valid majors: 8, 9, 10",
			version, major,
		)
	}
	return nil
}

func majorVersion(v string) string {
	if i := strings.IndexByte(v, '.'); i > 0 {
		return v[:i]
	}
	return v
}

func toRockyArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}