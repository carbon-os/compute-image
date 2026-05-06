//go:build windows

package compute_image

import (
	"fmt"
	"os"
	"path/filepath"

	wincontainer "github.com/carbon-os/compute-image/container/windows"
	"github.com/carbon-os/compute-image/registry"
	"github.com/carbon-os/compute-image/vm"
)

func Pull(ref any) (any, error) {
	switch r := ref.(type) {
	case ContainerRef:
		return wincontainer.Pull(wincontainer.Ref{Image: r.Image, Dir: r.Dir})
	case VMRef:
		return vm.Pull(vm.Ref{Image: r.Image, Registry: r.Registry, Arch: r.Arch, Dir: r.Dir})
	default:
		return nil, fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

func Remove(ref any) error {
	switch r := ref.(type) {
	case ContainerRef:
		return wincontainer.Remove(wincontainer.Ref{Image: r.Image, Dir: r.Dir})
	case VMRef:
		return vm.Remove(vm.Ref{Image: r.Image, Registry: r.Registry, Arch: r.Arch, Dir: r.Dir})
	default:
		return fmt.Errorf("compute-image: unknown ref type %T", ref)
	}
}

func ResolveContainerPaths(ref ContainerRef) (ContainerPaths, error) {
	p, err := wincontainer.ResolvePaths(wincontainer.Ref{Image: ref.Image, Dir: ref.Dir})
	if err != nil {
		return ContainerPaths{}, err
	}
	return ContainerPaths{Dir: p.Dir, Base: p.Base, Scratch: p.Scratch, Cache: p.Cache}, nil
}

func ResolveVMPaths(ref VMRef) (VMPaths, error) {
	p, err := vm.ResolvePaths(vm.Ref{Image: ref.Image, Registry: ref.Registry, Arch: ref.Arch, Dir: ref.Dir})
	if err != nil {
		return VMPaths{}, err
	}
	return VMPaths{Dir: p.Dir, Disk: p.Disk, Cache: p.Cache}, nil
}

func ContainerPathsFromImage(img any) ContainerPaths {
	r := img.(*wincontainer.Image)
	return ContainerPaths{Dir: r.Paths.Dir, Base: r.Paths.Base, Scratch: r.Paths.Scratch, Cache: r.Paths.Cache}
}

func VMPathsFromImage(img any) VMPaths {
	r := img.(*vm.Image)
	return VMPaths{Dir: r.Paths.Dir, Disk: r.Paths.Disk, Cache: r.Paths.Cache}
}

func RemoveAll(dir string) error {
	return wincontainer.RemoveAll(dir)
}

func HumanBytes(b int64) string {
	return registry.HumanBytes(b)
}

func DefaultRootDir() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local = "."
	}
	return filepath.Join(local, "carbon")
}