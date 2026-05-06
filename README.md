# Carbon Compute Image

Carbon Compute Image is the image preparation package for Carbon Compute. It
has one job: given an image reference, produce the on-disk paths that
`compute-container` or a VM runtime needs to do its job.

It has no concept of container lifecycle, process execution, or networking —
those are the responsibility of `carbon-os/compute-container`.

## Overview

`compute-image` accepts an image reference, handles downloading, caching,
format conversion, and layer preparation, and returns a ready-to-use result.
The API is designed to be driven by both AI agents and human operators.

## Usage

```go
// Pull a Windows container image
img, err := compute_image.Pull(compute_image.ContainerRef{
    Image: "mcr.microsoft.com/windows/nanoserver:ltsc2022",
    Dir:   `C:\images`, // optional
})
result := img.(*compute_image.ContainerImage)
// result.BaseLayer and result.Scratch are ready for compute_container.ImageMount
// result.Paths has all resolved on-disk locations

// Pull a Linux VM image
img, err := compute_image.Pull(compute_image.VMRef{
    Image:    "ubuntu:22.04",
    Registry: "cloud-images.ubuntu.com",
    Dir:      `C:\images`, // optional
})
result := img.(*compute_image.VMImage)
// result.OutPath is the prepared disk image
// result.Paths has all resolved on-disk locations

// Resolve paths without pulling — useful for scripting and inspection
paths, err := compute_image.ResolveContainerPaths(compute_image.ContainerRef{
    Image: "mcr.microsoft.com/windows/nanoserver:ltsc2022",
})
// paths.Dir, paths.Base, paths.Scratch, paths.Cache
```

## API

### Pull

```go
func Pull(ref Ref) (any, error)
```

Pull is the single entry point. Pass a `ContainerRef` or `VMRef` and get back
a `*ContainerImage` or `*VMImage` respectively. All downloading, caching,
conversion, and layer preparation happens inside.

### Ref types

| Type | Description |
|---|---|
| `ContainerRef` | OCI container image from any registry |
| `VMRef` | VM disk image from a cloud image registry |

### ContainerRef fields

| Field | Description |
|---|---|
| `Image` | Full image reference, e.g. `mcr.microsoft.com/windows/nanoserver:ltsc2022` |
| `Dir` | Root directory for all image data — defaults to `%LOCALAPPDATA%\carbon` |

### VMRef fields

| Field | Description |
|---|---|
| `Image` | Image reference in `name:version` form, e.g. `ubuntu:22.04` |
| `Registry` | Registry hostname, e.g. `cloud-images.ubuntu.com` |
| `Dir` | Root directory for all image data — defaults to `%LOCALAPPDATA%\carbon` |

### Result types

| Type | Fields | Ready for |
|---|---|---|
| `ContainerImage` | `BaseLayer`, `Scratch`, `Paths` | `compute_container.ImageMount` |
| `VMImage` | `OutPath`, `Paths` | VM runtime |

### Path resolution

All paths are derived from `Dir` and the image reference. You never need to
construct them manually.

**Container layout:**
```
<Dir>\
  cache\                              ← raw downloaded layer tarballs (by digest)
  <registry>\<repo>\<tag>\
    base\                             ← read-only HCS base layer
    scratch\                          ← writable scratch layer
```

**VM layout:**
```
<Dir>\
  cache\                              ← raw downloaded qcow2 images
  <registry>\<name>\<version>\
    disk.raw                          ← prepared disk image
```

### Inspect paths without pulling

```go
func ResolveContainerPaths(ref ContainerRef) (ContainerPaths, error)
func ResolveVMPaths(ref VMRef) (VMPaths, error)
```

Both functions resolve and return all paths without downloading or importing
anything. Useful for scripting, pre-flight checks, or passing paths to other
tools without having to construct them yourself.

## CLI

`image-cli` is a command-line tool for driving the package directly.

```
cd cmd/cli
go build -o image-cli.exe .

image-cli pull container <image> [--dir <path>]
image-cli pull vm        <image> --registry <host> [--dir <path>]
image-cli info container <image> [--dir <path>]
image-cli info vm        <image> --registry <host> [--dir <path>]
image-cli ls             [--dir <path>]
image-cli rm             <image> [--dir <path>]
```

### Examples

```powershell
# Pull a Windows container image
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022

# Pull to a specific directory
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022 --dir C:\images

# Pull a Ubuntu VM image
image-cli pull vm ubuntu:22.04 --registry cloud-images.ubuntu.com

# Inspect paths and status without pulling
image-cli info container mcr.microsoft.com/windows/nanoserver:ltsc2022

# List cached layers and images
image-cli ls

# Remove a cached image
image-cli rm ubuntu:22.04
```

The `base` and `scratch` paths produced by `pull container` can be passed
directly to `compute-container`:

```powershell
container-cli run `
    --base    $env:LOCALAPPDATA\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\base `
    --scratch $env:LOCALAPPDATA\carbon\mcr.microsoft.com\windows\nanoserver\ltsc2022\scratch `
    -- cmd.exe
```

Or use `image-cli info` to get the exact paths for your machine without
constructing them by hand.

## Container image pipeline

```
registry  →  manifest fetch + auth  →  layer download  →  cache
cache     →  ociwclayer import      →  base layer dir
base      →  CreateSandboxLayer     →  scratch dir + layerchain.json
```

Layers are cached by digest. Re-running `pull container` for the same image is
a no-op if all layers are already cached.

## VM image pipeline

```
registry  →  qcow2 download  →  cache
cache     →  qcow2 → raw conversion
raw       →  GPT parse → ext4 partition locate → extract to out dir
```

Downloaded qcow2 images are cached by name and version. Format conversion and
extraction always run against the cached source.

## Platforms

| Platform | Container support | VM support |
|----------|-------------------|------------|
| Windows  | ✓ (HCS / ociwclayer) | ✓ |
| Linux    | planned | ✓ |

Platform selection for container import is automatic at compile time via Go
build tags. The `Pull` API is identical across platforms.

## Architecture

```
compute-image/
  image.go                  // Pull, Ref interface, resolveDir, ResolveContainerPaths, ResolveVMPaths
  types.go                  // ContainerRef, VMRef, ContainerImage, VMImage, ContainerPaths, VMPaths
  log.go                    // logf

  container.go              // pullContainer — orchestrates registry + import
  container_registry.go     // manifest fetch, auth, layer download, cache
  container_import.go       // ociwclayer import, scratch, layerchain.json (windows)
  container_privileges.go   // SeBackupPrivilege / SeRestorePrivilege (windows)

  vm.go                     // pullVM — orchestrates fetch + convert + extract
  vm_registry.go            // cloud image fetch, cache
  vm_convert.go             // qcow2 → raw
  vm_extract.go             // GPT + ext4 extraction

  cmd/
    image-cli/
      main.go               // image-cli binary
      README.md
```

## Scope

This package does not:
- Run containers or VMs
- Manage container lifecycle or networking
- Interact with `compute-container` directly

Image consumption is the responsibility of `carbon-os/compute-container` for
containers, and the VM runtime for VM images. This package only produces the
paths they need.