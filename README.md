# qmdverify

A command-line tool for checking QMD file compatibility against reMarkable device firmware versions.

`qmdverify` uploads `.qmd` files to a [rm-qmd-verify](https://github.com/rmitchellscott/rm-qmd-verify) server and displays compatibility results in a matrix format.

## Features

- Upload and check `.qmd` files for compatibility
- Matrix view showing compatibility across device types and OS versions
- List available device types and OS versions
- Verbose mode for detailed error information
- Configurable server endpoint via environment variable
- Cross-platform support

## Installation

### Homebrew (macOS/Linux)

```bash
brew install rmitchellscott/tap/qmdverify
```

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/rmitchellscott/rm-qmd-verify-cli/releases).

#### Supported Platforms

- **Windows**: x86_64, ARM64
- **Linux**: x86_64, ARM64, ARMv7
- **macOS**: x86_64 (Intel), ARM64 (Apple Silicon)

### Install from Source

```bash
go install github.com/rmitchellscott/rm-qmd-verify-cli/cmd/qmdverify@latest
```

## Usage

### Check QMD File Compatibility

Upload and check a `.qmd` file:

```bash
qmdverify myfile.qmd
```

Or use the explicit `check` command:

```bash
qmdverify check myfile.qmd
```

Show detailed error messages with the `--verbose` flag:

```bash
qmdverify myfile.qmd --verbose
qmdverify check myfile.qmd -v
```

### List Available Hashtables

Display all available device types and OS versions:

```bash
qmdverify list
```

### Version Information

Show CLI and server versiosn:

```bash
qmdverify version
```

## Configuration

### Server Endpoint

By default, `qmdverify` connects to `http://qmdverify.scottlabs.io`. Override this using the `QMDVERIFY_HOST` environment variable:

```bash
export QMDVERIFY_HOST=https://qmdverify.example.com
qmdverify myfile.qmd
```

Or set it inline:

```bash
QMDVERIFY_HOST=https://qmdverify.example.com qmdverify myfile.qmd
```

## Examples

### Check Command

```bash
$ qmdverify template.qmd
Uploading template.qmd to http://qmdverify.scottlabs.io...


         reMarkable QMD Verifier

                  rm1   rm2   rmpp rmppm 
─────────────────────────────────────────
 3.22.4.2          —     ✗     ✓     ✓   
 3.22.0.65         —     —     ✓     ✓   
 3.22.0.64         ✗     ✗     ✓     —   
 3.21.0.79         —     —     —     ✓   
 3.20.0.92         ✗     ✗     ✓     —   


Summary: 12 checked | 7 compatible | 5 incompatible
```

The matrix displays:
- `✓` Compatible (green)
- `✗` Incompatible (red)
- `—` No data available

Exit code is 0 if all devices are compatible, 1 if any incompatibilities are found.

### Verbose Mode

```bash
$ qmdverify template.qmd --verbose
Uploading template.qmd to http://qmdverify.scottlabs.io...


         reMarkable QMD Verifier

                  rm1   rm2   rmpp rmppm 
─────────────────────────────────────────
 3.22.4.2          —     ✗     ✓     ✓   
 3.22.0.65         —     —     ✓     ✓   
 3.22.0.64         ✗     ✗     ✓     —   
 3.21.0.79         —     —     —     ✓   
 3.20.0.92         ✗     ✗     ✓     —   

Error Details:
  • 3.22.4.2  (rm2): Cannot resolve hash 1121852971369147487
  • 3.22.0.64 (rm1): Cannot resolve hash 1121852971369147487
  • 3.22.0.64 (rm2): Cannot resolve hash 1121852971369147487
  • 3.20.0.92 (rm1): Cannot resolve hash 1121852971369147487
  • 3.20.0.92 (rm2): Cannot resolve hash 1121852971369147487


Summary: 12 checked | 7 compatible | 5 incompatible
```

### List Command

```bash
$ qmdverify list
Fetching hashtables from http://qmdverify.scottlabs.io...

                    
Available Hashtables
                    

 Device    OS Version  Hashtable                Entries   
─────────────────────────────────────────────────────────────
 rmpp      3.20.0.92   3.20.0.92-rmpp           11061     
 rmppm     3.21.0.79   3.21.0.79-rmppm          11614     
 rm1       3.22.0.64   3.22.0.64-rm1            11587     
 rm2       3.22.0.64   3.22.0.64-rm2            11587     
 rmpp      3.22.0.64   3.22.0.64-rmpp           11217     
 rmppm     3.22.0.65   3.22.0.65-rmppm          11123     
 rmpp      3.22.4.2    3.22.4.2-rmpp            11232 
```

### Version Command

```bash
$ qmdverify version
qmdverify CLI
  Version: v1.0.0

Server (https://qmdverify.example.com)
  Version: v2.1.0
```

## Development

### Build

```bash
go build -o qmdverify ./cmd/qmdverify
```

### Build with Version Info

```bash
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o qmdverify ./cmd/qmdverify
```

### Test Release Build

```bash
goreleaser release --snapshot --clean
```

## Requirements

- Go 1.21 or later (for building from source)
- Access to a running [rm-qmd-verify](https://github.com/rmitchellscott/rm-qmd-verify) server

## License

MIT
