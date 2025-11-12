# qmdverify

A command-line tool for checking QMD file compatibility against reMarkable device firmware versions.

`qmdverify` uploads `.qmd` files to a [rm-qmd-verify](https://github.com/rmitchellscott/rm-qmd-verify) server and displays compatibility results in a matrix format.

## Features

- Multi-file validation with automatic dependency tracking
- Directory upload with preserved structure
- Root-files-only display (filters out dependencies)
- Matrix view showing compatibility across devices and OS versions
- Filter by device, version, file name, or failure status
- List available hashtables and QML trees
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
go install github.com/rmitchellscott/rm-qmd-verify-cli@latest
```

## Usage

### Check QMD File Compatibility

Upload and check one or more `.qmd` files:

```bash
# Single file
qmdverify myfile.qmd

# Multiple files
qmdverify file1.qmd file2.qmd file3.qmd

# Entire directory (preserves structure)
qmdverify ./qmd-files/

# Explicit check command
qmdverify check myfile.qmd
```

**Note**: When uploading multiple files with dependencies (via `LOAD` statements), only root files are displayed by default. Dependencies are validated but not shown in output.

Show detailed error messages with the `--verbose` flag:

```bash
qmdverify myfile.qmd --verbose
qmdverify myfile.qmd -v
```

### Filtering Results

Filter results by device type and/or OS version to focus on specific targets.

#### Filter by Device

Check compatibility for specific device(s) using `--device` or `-d`:

```bash
# Single device
qmdverify check --device rmpp myfile.qmd
qmdverify check -d rmpp myfile.qmd

# Multiple devices
qmdverify check --device rmpp --device rmppm myfile.qmd
qmdverify check -d rm1 -d rm2 myfile.qmd
```

Valid device values: `rm1`, `rm2`, `rmpp`, `rmppm`

#### Filter by Version

Check compatibility for specific OS version(s) using `--version`. Supports prefix matching:

```bash
# Exact version
qmdverify check --version 3.22.4.2 myfile.qmd

# Version prefix (matches all 3.22.x.x versions)
qmdverify check --version 3.22 myfile.qmd

# Multiple versions
qmdverify check --version 3.22.4.2 --version 3.21.0.79 myfile.qmd
```

#### Filter by File

When validating multiple files, show only specific files using `--file` or `-f`. Supports glob patterns and substring matching:

```bash
# Glob pattern
qmdverify ./qmd-files/ --file "*.qmd"

# Substring match
qmdverify ./qmd-files/ -f toolbar -f settings

# Multiple patterns
qmdverify ./qmd-files/ --file "base/*.qmd" --file "hacks/*.qmd"
```

#### Show Only Failed Files

Display only files with incompatibilities using `--failed-only`:

```bash
qmdverify ./qmd-files/ --failed-only
```

Exit code is 1 if any incompatibilities are found, 0 otherwise.

#### Combine Filters

Combine device, version, file, and failure filters:

```bash
# Device and version
qmdverify --device rmpp --version 3.22 myfile.qmd

# Failed files for specific device
qmdverify ./qmd-files/ --failed-only --device rmpp

# Specific files, specific versions
qmdverify ./qmd-files/ -f "toolbar*" --version 3.22

# All filters combined
qmdverify ./qmd-files/ --device rmpp --version 3.22 --file "*.qmd" --failed-only
```

**Note**: Filters are applied client-side after server validation. Empty results display a warning and exit with code 0.

### List Available Resources

Display all available hashtables (device types and OS versions):

```bash
qmdverify list
```

Display available QML trees for tree-based validation:

```bash
qmdverify list trees
```

### Hashtable Conversion

Convert hashtab files to compact hashlist format:

```bash
qmdverify hashlist create hashtabs/3.22.0.64-rmpp hashlists/3.22.0.64-rmpp
```

This command strips string data from hashtab files (hash-string pairs) and outputs a compact hashlist file containing only the hashes in qmldiff's binary format.

**Note**: The input must be a valid hashtab file. If the input is already a hashlist, an error is returned.

### Version Information

Show CLI and server versions:

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

### Single File Check

```bash
$ qmdverify template.qmd
Uploading 1 files to http://qmdverify.scottlabs.io...


=== template.qmd ===


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

### Multi-File Directory Check

```bash
$ qmdverify ./rm-hacks/
Uploading 39 files to http://qmdverify.scottlabs.io...


=== zz_rmhacks.qmd ===


         reMarkable QMD Verifier

                  rm1   rm2   rmpp rmppm
─────────────────────────────────────────
 3.23.0.64         ✗     ✗     ✗     ✗
 3.22.4.2          —     ✓     ✓     ✓
 3.22.0.64         ✓     ✓     ✓     —


Summary: 15 checked | 11 compatible | 4 incompatible
```

**Note**: Only root file `zz_rmhacks.qmd` is shown. The other 38 files are dependencies loaded via `LOAD` statements and validated automatically.

### Matrix Legend

- `✓` Compatible (green)
- `✗` Incompatible (red)
- `—` No data available

Exit code is 0 if all files are compatible, 1 if any incompatibilities are found.

### Verbose Mode

```bash
$ qmdverify template.qmd --verbose
Uploading 1 files to http://qmdverify.scottlabs.io...


=== template.qmd ===


         reMarkable QMD Verifier

                  rm1   rm2   rmpp rmppm
─────────────────────────────────────────
 3.22.4.2          —     ✗     ✓     ✓
 3.22.0.65         —     —     ✓     ✓
 3.22.0.64         ✗     ✗     ✓     —
 3.21.0.79         —     —     —     ✓
 3.20.0.92         ✗     ✗     ✓     —

Error Details:
  • 3.22.4.2  (rm2): 1 dependency file has errors
  • 3.22.0.64 (rm1): Cannot resolve hash 1121852971369147487
  • 3.22.0.64 (rm2): Cannot resolve hash 1121852971369147487
  • 3.20.0.92 (rm1): Cannot resolve hash 1121852971369147487
  • 3.20.0.92 (rm2): Cannot resolve hash 1121852971369147487


Summary: 12 checked | 7 compatible | 5 incompatible
```

### Failed Files Only

```bash
$ qmdverify ./rm-hacks/ --failed-only
Uploading 39 files to http://qmdverify.scottlabs.io...


=== zz_rmhacks.qmd ===


         reMarkable QMD Verifier

                  rm1   rm2   rmpp rmppm
─────────────────────────────────────────
 3.23.0.64         ✗     ✗     ✗     ✗
 3.21.0.79         —     —     —     ✗
 3.20.0.92         ✗     ✗     ✗     —


Summary: 30 checked | 22 compatible | 8 incompatible
```

### List Hashtables

```bash
$ qmdverify list
Fetching hashtables from http://qmdverify.scottlabs.io...


Available Hashtables


 Device    OS Version  Hashtable                Entries
─────────────────────────────────────────────────────────────
 rm1       3.20.0.92   3.20.0.92-rm1            11061
 rm2       3.20.0.92   3.20.0.92-rm2            11061
 rmpp      3.20.0.92   3.20.0.92-rmpp           11061
 rmppm     3.21.0.79   3.21.0.79-rmppm          11614
 rm1       3.22.0.64   3.22.0.64-rm1            11587
 rm2       3.22.0.64   3.22.0.64-rm2            11587
 rmpp      3.22.0.64   3.22.0.64-rmpp           11217


Total Hashtables: 15
```

### List QML Trees

```bash
$ qmdverify list trees
Fetching QML trees from http://qmdverify.scottlabs.io...


Available QML Trees


 Device    OS Version    QML Files  Directory
──────────────────────────────────────────────────────
 rm1       3.20.0.92     485        3.20.0.92-rm1
 rm2       3.20.0.92     485        3.20.0.92-rm2
 rmpp      3.22.0.64     512        3.22.0.64-rmpp
 rmppm     3.22.4.2      518        3.22.4.2-rmppm


Total Trees: 17
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
go build -o qmdverify
```

### Build with Version Info

```bash
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o qmdverify
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
