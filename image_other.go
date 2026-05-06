//go:build !windows

package compute_image

import (
	"fmt"
	"os"
	"path/filepath"

	linuxcontainer "github.com/carbon-os/compute-image/container/linux"
	"github.com/carbon-os/compute-image/registry"
	"github.com/carbon-os/compute-image/vm"
)

func Pull(ref any) (any, error) {
	switch r := ref.(type) {
	case ContainerRef:
		return linuxcontainer.Pull(linuxcontainer.Ref{Image: r.Image, Dir: r.Dir})
	case VMRef:
		return vm.Pull(vm.Ref{Image: r.Image, Registry: r.Registry, Arch: r.Arch, Dir: r.Dir})
	default:
		return nil, fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

func Remove(ref any) error {
	switch r := ref.(type) {
	case ContainerRef:
		return linuxcontainer.Remove(linuxcontainer.Ref{Image: r.Image, Dir: r.Dir})
	case VMRef:
		return vm.Remove(vm.Ref{Image: r.Image, Registry: r.Registry, Arch: r.Arch, Dir: r.Dir})
	default:
		return fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

func ResolveContainerPaths(ref ContainerRef) (ContainerPaths, error) {
	p, err := linuxcontainer.ResolvePaths(linuxcontainer.Ref{Image: ref.Image, Dir: ref.Dir})
	if err != nil {
		return ContainerPaths{}, err
	}
	return ContainerPaths{Dir: p.Dir, Layers: p.Layers, Upper: p.Upper, Work: p.Work, Cache: p.Cache}, nil
}

func ResolveVMPaths(ref VMRef) (VMPaths, error) {
	p, err := vm.ResolvePaths(vm.Ref{Image: ref.Image, Registry: ref.Registry, Arch: ref.Arch, Dir: ref.Dir})
	if err != nil {
		return VMPaths{}, err
	}
	return VMPaths{Dir: p.Dir, Disk: p.Disk, Cache: p.Cache}, nil
}

func ContainerPathsFromImage(img any) ContainerPaths {
	r := img.(*linuxcontainer.Image)
	return ContainerPaths{Dir: r.Paths.Dir, Layers: r.Paths.Layers, Upper: r.Paths.Upper, Work: r.Paths.Work, Cache: r.Paths.Cache}
}

func VMPathsFromImage(img any) VMPaths {
	r := img.(*vm.Image)
	return VMPaths{Dir: r.Paths.Dir, Disk: r.Paths.Disk, Cache: r.Paths.Cache}
}

func RemoveAll(dir string) error {
	return os.RemoveAll(dir)
}

func HumanBytes(b int64) string {
	return registry.HumanBytes(b)
}

func DefaultRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".carbon"
	}
	return filepath.Join(home, ".local", "share", "carbon")
}