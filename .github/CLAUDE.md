# CLAUDE.md

## Welcome, Claude

This repository uses **`AGENTS.md`** as the entry point to our modular technical conventions:

- The main **`AGENTS.md`** provides an overview and directory structure
- Technical standards are organized in **`.github/tech-conventions/`**:
  - **Core Development**: Go essentials, testing, documentation
  - **Version Control**: Commits, branches, pull requests, releases
  - **Infrastructure**: CI/CD, dependencies, security, workflows

> **Start with `AGENTS.md`**, then explore specific conventions in `tech-conventions/`.

---

## About This Project

`go-instantly` is a zero-dependency Go client library for the **Instantly.ai V2 API**. It is a
library, not an application: there is no binary, no CLI, and no runtime configuration file.

- Module: `github.com/mrz1836/go-instantly`
- Go version: 1.25
- Runtime dependencies: none (`testify` is used for tests only)

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `magex test` | Fast linting + unit tests |
| `magex test:coverrace` | Full CI suite with race detection |
| `magex lint` | Run 60+ linters via golangci-lint |
| `magex format:fix` | Auto-fix code formatting |
| `magex bench` | Run performance benchmarks |
| `magex deps:audit` | Scan for security vulnerabilities |
| `magex version:bump bump=patch push` | Create release tag (triggers CI) |

---

## MAGE-X Build System

This project uses **[MAGE-X](https://github.com/mrz1836/mage-x)** - a zero-config build automation
system providing 150+ commands for testing, linting, building, and releasing Go projects.

**Parameter Syntax:**
```bash
magex bench time=50ms count=3      # Quick benchmarks with timing
magex version:bump bump=minor push # Bump version and push tag
magex test:fuzz time=30s           # Fuzz tests for 30 seconds
```

**Release Process:** Use `magex version:bump bump=patch push` to create a tag. GitHub Actions
handles the actual release automatically.

---

## Development Cycle

```bash
magex test           # Fast linting + unit tests (before every commit)
magex test:race      # Unit tests with race detector
magex lint           # Run all linters
magex format:fix     # Fix formatting issues
```

---

## Testing & Benchmarks

**Unit Testing:**
```bash
magex test:unit       # Skip linting, run only tests
magex test:short      # Skip integration tests
magex test:coverrace  # Full CI suite with race detection
magex test:cover      # Unit tests with coverage report
```

Tests run against an in-repo mock HTTP router. They never contact the live Instantly.ai API.

**Fuzz Testing:**
```bash
magex test:fuzz time=30s                                # Run all fuzz tests
go test -fuzz=FuzzEmailSerialization -fuzztime=30s .    # Specific fuzz test
```

**Benchmarks:**
| Type | Command | Duration |
|------|---------|----------|
| Quick (CI) | `magex bench` or `magex benchquick` | < 5 min |
| Heavy | `magex benchheavy` | 10-30 min |

---

## go-pre-commit

Fast Go-native pre-commit hooks (17x faster than Python alternatives).

```bash
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
go-pre-commit install
```

See [go-pre-commit documentation](https://github.com/mrz1836/go-pre-commit) for configuration and usage.

---

## Code Quality Guidelines

This project uses 60+ linters via golangci-lint with strict standards.

**Essential Practices:**
- Use `0600` for sensitive files, `0750` for directories
- Always check error returns: `if err := foo(); err != nil { ... }`
- Use context-aware functions: `DialContext`, `CommandContext`
- Create static error variables and wrap with context

**Code Quality:**
- Add comments to all exported functions, types, and constants
- Use `_` for intentionally unused parameters
- Avoid redefining built-in functions (`max`, `min`, etc.)
- Pre-allocate slices when size is known: `make([]Type, 0, knownSize)`

**Library Constraints:**
- **No runtime dependencies** - the non-test source must not import third-party packages
- Every exported symbol carries a doc comment
- Requests are context-first and flow through a single central request helper

**Formatting:**
```bash
# Always use magex format:fix for code formatting
magex format:fix

# Never use gofumpt or fmt directly
```

**Common Patterns:**
- Use `fmt.Fprintf(w, format, args...)` for efficient string building
- Add `//nolint:linter // reason` only when necessary with clear explanation

**Running Linters:**
```bash
magex format:fix  # Fix formatting first
magex lint        # Run all linters
```

---

## Troubleshooting

**Environment Verification:**
```bash
gh auth status    # Check GitHub authentication
go mod verify     # Validate dependencies
govulncheck ./... # Security scan
```

---

## Checklist & Reminders

**Before Development:**
1. Read `AGENTS.md` thoroughly
2. Run `magex test` to ensure all tests pass
3. Verify GitHub authentication with `gh auth status`

**Key Rules:**
- **`AGENTS.md` is the ultimate authority** - When in doubt, refer to it first
- **Never tag releases** - Only repository code-owners handle releases
- **Security first** - Run `govulncheck` and validate external dependencies
- **Test thoroughly** - Cover both happy paths and error paths

---

## Documentation

| Document                             | Purpose                                         |
|--------------------------------------|-------------------------------------------------|
| [`README.md`](../README.md)          | Project overview and quick start                |
| [`AGENTS.md`](AGENTS.md)             | Primary authority for all development standards |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Contribution guidelines                         |
| [`examples/`](../examples/)          | Usage examples                                  |
