# Development Guide

## Prerequisites

- Go 1.24+
- Git
- Make
- GCC

## Building

```bash
go build -o taws ./cmd/taws
```

## Testing

```bash
go test ./...
```

## Project Structure

- `cmd/` - CLI entry points
- `internal/` - Private packages
- `pkg/` - Public packages
- `templates/` - Configuration templates
- `tests/` - Test files
