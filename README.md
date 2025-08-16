# Yuki - Package Manager for Zig

Yuki is a comprehensive package manager for Zig projects written in Go. 
It provides modern dependency management with GitHub integration, semantic versioning, and a complete CLI interface similar to Cargo or npm. 
The tool manages Zig project dependencies, builds, tests, and provides package discovery through GitHub's API.

## Features

### 🚀 Project Management
- **`yuki init`** - Initialize new projects with interactive setup
- **`yuki build`** - Compile projects with dependencies
- **`yuki test`** - Run tests with dependencies
- **`yuki run`** - Compile and execute projects
- **`yuki clean`** - Clean build artifacts and dependencies

### 📦 Dependency Management
- **`yuki add <pkg>@<version>`** - Add dependencies with semantic versioning
- **`yuki install`** - Install all dependencies from manifest
- **`yuki update [pkg]`** - Update dependencies to latest compatible versions
- **`yuki remove <pkg>`** - Remove dependencies
- **`yuki sync`** - Verify dependency consistency
- **`yuki list`** - Display dependency tree
- **`yuki why <pkg>`** - Explain dependency requirements

### 🔍 Discovery & Information
- **`yuki search <query>`** - Search GitHub for Zig packages
- **`yuki info <pkg>`** - Display detailed package information
- **`yuki outdated`** - Show packages with available updates

### 🛠️ Utilities & Configuration
- **`yuki doctor`** - Diagnose environment issues
- **`yuki config`** - Manage global configuration
- **`yuki cache`** - Manage local package cache

## Installation

### Prerequisites
- [Go 1.21+](https://golang.org/dl/)
- [Zig 0.12.0+](https://ziglang.org/download/)
- [Git](https://git-scm.com/)

### Build from Source
```bash
git clone https://github.com/DitzDev/yuki-zpm.git
cd yuki-zpm
go build -o yuki ./cmd/yuki/main.go
```

## TODO
- [] Registry
- [] Add more features

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License
