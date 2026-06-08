# Multi-Protocol Proxy Build Scripts

> Modular bash scripts for building, installing, and running the Multi-Protocol Proxy across multiple platforms.

## Quick Start

```bash
# Interactive menu
./scripts/start.sh

# Or use specific commands
./scripts/start.sh install    # Install dependencies
./scripts/start.sh build      # Build for current platform
./scripts/start.sh run        # Run the proxy
./scripts/start.sh dev        # Build and run in dev mode
```

## Scripts Overview

| Script | Description |
|--------|-------------|
| `start.sh` | Unified entry point with interactive menu |
| `install.sh` | Installs Go dependencies |
| `build.sh` | Builds the proxy with platform detection |
| `run.sh` | Runs the built proxy |

### Library Modules (`lib/`)

| Module | Purpose |
|--------|---------|
| `colors.sh` | Terminal styling matching the Go UI theme |
| `platform.sh` | OS and architecture detection |
| `utils.sh` | Common helper functions |

## Usage

### Installing Dependencies

```bash
./scripts/install.sh
```

Downloads Go modules and verifies the build can compile.

### Building

```bash
# Build for current platform (auto-detected)
./scripts/build.sh

# Build for all supported platforms
./scripts/build.sh --all

# Build for specific platform
./scripts/build.sh --os linux --arch arm64

# Clean build
./scripts/build.sh --clean
```

#### Build Options

| Option | Description |
|--------|-------------|
| `--all` | Build for all supported platforms |
| `--os <os>` | Target OS: `linux`, `darwin`, `windows` |
| `--arch <arch>` | Target architecture: `amd64`, `arm64` |
| `--clean` | Clean build directory before building |

### Running

```bash
# Run with auto-detected platform binary
./scripts/run.sh

# Run in development mode
./scripts/run.sh --dev

# Run in production mode
./scripts/run.sh --prod

# Build first if binary missing, then run
./scripts/run.sh --build-first

# Pass arguments to the proxy
./scripts/run.sh -- --config custom.json
```

### Unified Start Script

```bash
# Show interactive menu
./scripts/start.sh

# Direct commands
./scripts/start.sh install      # Install dependencies
./scripts/start.sh build        # Build for current platform
./scripts/start.sh build-all    # Build for all platforms
./scripts/start.sh run          # Run the proxy
./scripts/start.sh dev          # Build and run (dev mode)
./scripts/start.sh info         # Show platform info
```

## Build Output

Binaries are placed in the `build/` directory with platform-specific names:

```
build/
├── multi-protocol-proxy-linux-amd64
├── multi-protocol-proxy-linux-arm64
├── multi-protocol-proxy-darwin-amd64
├── multi-protocol-proxy-darwin-arm64
└── multi-protocol-proxy-windows-amd64.exe
```

## Supported Platforms

| OS | Architecture | Binary Name |
|----|--------------|-------------|
| Linux | amd64 | `multi-protocol-proxy-linux-amd64` |
| Linux | arm64 | `multi-protocol-proxy-linux-arm64` |
| macOS | amd64 | `multi-protocol-proxy-darwin-amd64` |
| macOS | arm64 | `multi-protocol-proxy-darwin-arm64` |
| Windows | amd64 | `multi-protocol-proxy-windows-amd64.exe` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Disable colored output |
| `FORCE_COLOR` | Force colored output |
| `APP_ENV` | Environment mode (`development` or `production`) |

## Troubleshooting

### "Go is not installed"

Install Go from [go.dev/dl](https://go.dev/dl/) and ensure it's in your PATH.

### "Binary not found"

Run `./scripts/build.sh` to build the proxy first.

### Colors not displaying

- Check that your terminal supports ANSI colors
- Set `FORCE_COLOR=1` to force colors
- If colors cause issues, set `NO_COLOR=1` to disable

### Permission denied

Make scripts executable:

```bash
chmod +x scripts/*.sh scripts/lib/*.sh
```
