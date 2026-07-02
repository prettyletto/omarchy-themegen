# Release Checks

Run these before publishing binaries or tags. All checks must pass.

## Automated (AFK)

```bash
# Formatting
go fmt ./...

# Static analysis
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...

# Build
go build ./...

# Tests (does not mutate user config, does not call omarchy theme set)
go test ./... -count=1
```

## Manual (HITL)

```bash
# Install
go install .

# Version check (works without Omarchy or magick)
omarchy-themegen --version
omarchy-themegen --version --json

# Image validation (requires magick)
omarchy-themegen --image <valid-opaque-image>.png --name release-check --direction 1 --yes

# Recipe export
omarchy-themegen --image <image>.png --name release-check --direction 1 --recipe /tmp/test.recipe.json --yes

# Recipe replay
omarchy-themegen --image <image>.png --name release-replay --replay /tmp/test.recipe.json --yes

# Reproducible archive
omarchy-themegen --image <image>.png --name release-repro --direction 1 --reproducible --yes

# Apply (explicit confirmation required)
omarchy theme set <name>

# Rollback
omarchy theme set <known-good-theme>
```

## Requirements

- `magick` (ImageMagick) must be on PATH for image validation and generation.
- Network access is only required for `go install` and `staticcheck` download.
- Generation is fully offline.
- No `omarchy theme set` is ever called by tests.
