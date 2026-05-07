package vm

import (
	"fmt"

	"github.com/carbon-os/compute-image/vm/alpine"
	"github.com/carbon-os/compute-image/vm/arch"
	"github.com/carbon-os/compute-image/vm/debian"
	"github.com/carbon-os/compute-image/vm/fedora"
	"github.com/carbon-os/compute-image/vm/opensuse"
	"github.com/carbon-os/compute-image/vm/rocky"
	"github.com/carbon-os/compute-image/vm/ubuntu"
	"github.com/carbon-os/compute-image/vm/almalinux"
	"github.com/carbon-os/compute-image/vm/centos"
)

// Provider is implemented by each distro sub-package.
type Provider interface {
	DefaultRegistry() string
	BuildURL(reg, version, arch string) string
	// Validate returns a descriptive error for any unsupported version or arch
	// before a download is attempted.
	Validate(version, arch string) error
}

type Ref struct {
	Image    string
	Registry string
	Arch     string
	Dir      string
}

type Image struct {
	Image   string
	Paths   Paths
	OutPath string
}

type Paths struct {
	Dir   string
	Disk  string
	Cache string
}

// — registry —

var providers = map[string]Provider{}

func Register(name string, p Provider) { providers[name] = p }

func lookup(name string) (Provider, bool) {
	p, ok := providers[name]
	return p, ok
}

func DefaultRegistry(name string) (string, error) {
	p, ok := lookup(name)
	if !ok {
		return "", fmt.Errorf("no default registry for %q — supply Registry in vm.Ref", name)
	}
	return p.DefaultRegistry(), nil
}

func buildImageURL(reg, name, version, arch string) string {
	if p, ok := lookup(name); ok {
		return p.BuildURL(reg, version, arch)
	}
	return fmt.Sprintf("https://%s/%s/%s/%s-%s-%s.qcow2", reg, name, version, name, version, arch)
}

// Validate looks up the named provider and checks version+arch before any I/O.
func Validate(name, version, canonicalArch string) error {
	p, ok := lookup(name)
	if !ok {
		return fmt.Errorf("unknown distro %q", name)
	}
	return p.Validate(version, canonicalArch)
}

// — adapter —

func adapt(
	defaultReg func() string,
	buildURL func(string, string, string) string,
	validate func(version, arch string) error,
) Provider {
	return &fnProvider{
		defaultRegistry: defaultReg,
		buildURL:        buildURL,
		validate:        validate,
	}
}

type fnProvider struct {
	defaultRegistry func() string
	buildURL        func(reg, version, arch string) string
	validate        func(version, arch string) error
}

func (p *fnProvider) DefaultRegistry() string                   { return p.defaultRegistry() }
func (p *fnProvider) BuildURL(reg, version, arch string) string { return p.buildURL(reg, version, arch) }
func (p *fnProvider) Validate(version, arch string) error       { return p.validate(version, arch) }

// — wiring —

func init() {
	Register("alpine",   adapt(alpine.DefaultRegistry,   alpine.BuildURL,   alpine.Validate))
	Register("debian",   adapt(debian.DefaultRegistry,   debian.BuildURL,   debian.Validate))
	Register("ubuntu",   adapt(ubuntu.DefaultRegistry,   ubuntu.BuildURL,   ubuntu.Validate))
	Register("fedora",   adapt(fedora.DefaultRegistry,   fedora.BuildURL,   fedora.Validate))
	Register("opensuse", adapt(opensuse.DefaultRegistry, opensuse.BuildURL, opensuse.Validate))
	Register("rocky",    adapt(rocky.DefaultRegistry,    rocky.BuildURL,    rocky.Validate))
	Register("arch",     adapt(arch.DefaultRegistry,     arch.BuildURL,     arch.Validate))
	Register("almalinux", adapt(almalinux.DefaultRegistry, almalinux.BuildURL, almalinux.Validate))
	Register("centos",    adapt(centos.DefaultRegistry,    centos.BuildURL,    centos.Validate))
}