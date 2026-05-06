package opensuse

import (
	"fmt"
	"strconv"
	"strings"
)

func DefaultRegistry() string { return DefaultReg }

func BuildURL(reg, version, arch string) string {
	oa := toOpenSUSEArch(arch)
	switch {
	case version == "tumbleweed":
		return fmt.Sprintf(tumbleweedPath, reg, oa)
	case majorGte16(version):
		return fmt.Sprintf(leap16Path, reg, version, version, oa)
	default:
		return fmt.Sprintf(leap15Path, reg, version, version, oa)
	}
}

// Validate checks that version is a supported openSUSE release and arch is valid.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf("opensuse: unsupported arch %q — valid arches: amd64, arm64", arch)
	}
	if !ValidVersions[version] {
		return fmt.Errorf(
			"opensuse: unsupported version %q — valid versions: tumbleweed, 15.6, 16.0",
			version,
		)
	}
	return nil
}

func majorGte16(v string) bool {
	major := v
	if i := strings.IndexByte(v, '.'); i > 0 {
		major = v[:i]
	}
	n, err := strconv.Atoi(major)
	if err != nil {
		return false
	}
	return n >= 16
}

func toOpenSUSEArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}