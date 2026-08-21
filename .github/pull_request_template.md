## Summary

- What changed?
- Why is it needed?

## Validation

- [ ] `go mod verify`
- [ ] `go mod tidy -diff`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] `golangci-lint run ./...`
- [ ] `go generate ./...` produces no uncommitted changes

## Checklist

- [ ] Tests cover the changed behavior
- [ ] Documentation and changelog are updated when applicable
- [ ] No credentials, state files, generated binaries, or vendored dependencies are committed
