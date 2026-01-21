# Contributing to council-cli

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Clone and Build

```bash
git clone https://github.com/luuuc/council-cli.git
cd council-cli
make build
```

The binary will be in `bin/council`.

**Note:** Always use `make build` - don't run `go build` directly to the project root. Build outputs belong in `bin/`.

### Install Locally

```bash
make install
```

This copies the binary to `/usr/local/bin/council`.

### Alternative: Go Install

```bash
go install github.com/luuuc/council-cli/cmd/council@latest
```

## Running Tests

```bash
make test
```

Or run specific packages:

```bash
go test -v ./internal/config/...
go test -v ./internal/sync/...
```

## Linting

```bash
make lint
```

This uses `golangci-lint`. If not installed, the command will install it first.

## CI Checks

Run all checks before pushing:

```bash
make ci
```

This runs build, tests, and lint.

## Code Structure

```
council-cli/
├── bin/                  # Build output (gitignored)
├── cmd/council/          # CLI entry point
├── internal/
│   ├── cmd/              # Cobra command definitions
│   ├── config/           # Configuration loading/saving
│   ├── creator/          # Expert creator functionality
│   ├── detect/           # Stack detection
│   ├── expert/           # Expert data structures
│   ├── export/           # Markdown export
│   ├── fs/               # File system utilities
│   ├── mcp/              # MCP server
│   ├── prompt/           # Prompt generation
│   └── sync/             # Sync targets (claude, opencode, etc.)
├── docs/                 # Documentation
├── install.sh            # Installer script
└── Makefile              # Build commands
```

## Adding a New Sync Target

1. Add the target to `internal/sync/sync.go`:

```go
var Targets = map[string]Target{
    // ... existing targets
    "newtarget": {
        Name:     "New Target",
        Location: ".newtarget/",
        Sync:     syncNewTarget,
        Check:    func() bool { return fs.DirExists(".newtarget") },
        Clean:    []string{".newtarget/agents/"},
    },
}
```

2. Implement the sync function:

```go
func syncNewTarget(experts []*expert.Expert, cfg *config.Config, opts Options) error {
    // Generate and write files
}
```

3. Add tests in `internal/sync/sync_test.go`

## Adding Suggested Experts

Suggestions are in `internal/cmd/suggestions.yaml`. Add entries under the appropriate category:

```yaml
ruby:
  - name: "Expert Name"
    focus: "Area of expertise"
```

Categories: `go`, `ruby`, `python`, `javascript`, `elixir`, `rust`, `testing`, `design`, `general`.

## PR Guidelines

1. Run `make ci` before submitting
2. Keep changes focused and atomic
3. Write tests for new functionality
4. Update documentation if adding features

## Questions?

Open an issue at https://github.com/luuuc/council-cli/issues
