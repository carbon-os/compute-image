package arch

import "fmt"

func DefaultRegistry() string { return DefaultReg }

// BuildURL constructs the Arch Linux cloud image download URL.
// version is ignored — Arch is a rolling release with a single "latest" path.
func BuildURL(reg, _ /* version */, arch string) string {
	return fmt.Sprintf(downloadPath, reg, toArchArch(arch))
}

// Validate checks that version is "latest" and arch is amd64.
// Arch Linux only officially publishes x86_64 cloud images via arch-boxes.
func Validate(version, arch string) error {
	if !ValidArches[arch] {
		return fmt.Errorf(
			"arch: unsupported arch %q — only \"amd64\" (x86_64) is officially published; "+
				"see https://geo.mirror.pkgbuild.com/images/latest/",
			arch,
		)
	}
	if !ValidVersions[version] {
		return fmt.Errorf(
			"arch: unsupported version %q — Arch Linux is a rolling release; use \"latest\"",
			version,
		)
	}
	return nil
}

func toArchArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return arch
	}
}