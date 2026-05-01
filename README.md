# GMCore Workspace Tool

Internal development tool for GMCore framework management.

## Installation

Download the binary for your platform from the releases page:

| Platform | Architecture | File |
|----------|-------------|------|
| Linux | amd64 | `gmcore-linux-amd64` |
| Linux | arm64 | `gmcore-linux-arm64` |
| macOS | amd64 | `gmcore-darwin-amd64` |
| macOS | arm64 | `gmcore-darwin-arm64` |
| Windows | amd64 | `gmcore-windows-amd64.exe` |

## Usage

```bash
# Create SDK release (auto-increment version)
gmcore release minor   # 1.0.0 → 1.1.0
gmcore release major   # 1.0.0 → 2.0.0
gmcore release bugfix  # 1.0.0 → 1.0.1

# Or specify exact version
gmcore release v1.2.3

# Build framework tarball locally
gmcore build-framework v1.0.0

# Show version
gmcore version
```

## Requirements

- GitHub token set via `GITHUB_TOKEN` environment variable (for release command)
- Access to `gmcorenet/sdk` repository
