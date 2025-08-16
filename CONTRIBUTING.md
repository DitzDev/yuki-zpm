# Contributing to Yuki

Thank you for your interest in contributing to **Yuki**!  
We welcome issues, feature requests, bug reports, and pull requests.

## Development Setup

1. Install prerequisites:
   - Go 1.21+
   - Zig 0.12.0+
   - Git

2. Clone the repository:
```bash
$ git clone https://github.com/DitzDev/yuki-zpm.git
$ cd yuki-zpm
```

3. Build the Yuki Package Manager CLI:
```
$ go build -o yuki ./cmd/yuki/main.go
```

## Contribution Workflow
1. Fork the repository
2. Create a new feature branch:
```
$ git checkout -b feature/my-feature
```

3. Make changes and commit:
```
$ git commit -m "feat: add my new feature"
```

4. Push to your fork and submit a Pull Request

## Commit Message Guidelines
We use Conventional Commits:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `chore:` for maintenance tasks
- `test:` for adding tests
- `refactor:` for code changes that don’t add features or fix bugs

Example:
```
feat(deps): add support for Zig 0.13.0
```

## Code Style
- Follow Go conventions (go fmt, golangci-lint)
- Keep CLI UX consistent with Cargo/npm semantics
- Add unit tests for new functionality

## Reporting Issues
- Please use the GitHub Issues page.
- Include reproduction steps, expected vs actual behavior
- Mention OS, Go, and Zig versions

Thank you for helping improve Yuki!