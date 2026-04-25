# image-cli

Command-line tool for pulling and managing images for Carbon Compute. Handles
both Windows container images (HCS/ociwclayer) and Linux VM disk images
(qcow2 → raw → ext4).

## Usage

```
image-cli pull container <image> --base <path> --scratch <path> [--cache <path>]
image-cli pull vm        <image> --registry <host> --out <path>  [--cache <path>]
image-cli ls             [--cache <path>]
image-cli rm             <image> [--cache <path>]
```

## Examples

```powershell
# Pull a Windows container image
image-cli pull container mcr.microsoft.com/windows/nanoserver:ltsc2022 `
    --base    .\images\nanoserver\base `
    --scratch .\images\nanoserver\scratch

# Pull a Ubuntu VM image
image-cli pull vm ubuntu:22.04 `
    --registry cloud-images.ubuntu.com `
    --out .\images\ubuntu

# List cached layers
image-cli ls

# Remove a cached image
image-cli rm ubuntu:22.04
```

The `--base` and `--scratch` paths produced by `pull container` map directly
to `compute-container`'s `--base` and `--scratch` flags:

```powershell
container-cli run `
    --base    .\images\nanoserver\base `
    --scratch .\images\nanoserver\scratch `
    -- cmd.exe
```

## Windows container images

All Windows base images are pulled from `mcr.microsoft.com`. Must be run as
Administrator (required for HCS layer import privileges).

| Image | Tag | Size | Use when |
|---|---|---|---|
| `mcr.microsoft.com/windows/nanoserver` | `ltsc2022`, `ltsc2019` | ~100 MB | .NET Core, minimal footprint |
| `mcr.microsoft.com/windows/servercore` | `ltsc2025`, `ltsc2022`, `ltsc2019`, `ltsc2016` | ~1.5 GB | .NET Framework, lift-and-shift |
| `mcr.microsoft.com/windows/server` | `ltsc2025`, `ltsc2022` | ~3.1 GB | Full Windows API, GPU, IIS |
| `mcr.microsoft.com/windows` | `ltsc2022`, `ltsc2019` | ~3.4 GB | Full Windows API + GDI/graphics libs |

### Nano Server

Nano Server is the smallest image but has a significantly reduced API surface.
**PowerShell, WMI, and the Windows servicing stack are not included.** It is
suitable for .NET Core applications and modern open source frameworks only. If
your workload needs PowerShell, use Server Core instead.

### Choosing a tag

`ltsc` tags (Long-Term Servicing Channel) are the stable production choice.
The host OS version must be compatible with the container image version — you
cannot run an `ltsc2025` container on a Windows Server 2019 host.

| Host OS | Recommended tag |
|---|---|
| Windows Server 2025 | `ltsc2025` |
| Windows Server 2022 | `ltsc2022` |
| Windows Server 2019 | `ltsc2019` |
| Windows Server 2016 | `ltsc2016` |

## Cache

Downloaded layers are cached by digest and reused across pulls. The default
cache location is `%LOCALAPPDATA%\carbon\cache`. Override with `--cache`.

Re-running `pull container` for the same image is a no-op if all layers are
already cached — only the HCS import and scratch creation run again.