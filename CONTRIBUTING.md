# Contributing

Thanks for helping maintain the Azure Resource Graph Terraform provider.

## Development setup

Install Go 1.26 or newer and Terraform 1.0 or newer, then run:

```sh
go mod download
go mod verify
go mod tidy -diff
go test -race -cover ./...
go vet ./...
golangci-lint run ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
go generate ./...
```

Generated documentation must be committed whenever provider schemas or examples change.

## Pull requests

- Keep each change focused and explain its user-visible behavior.
- Add tests for fixes and new behavior.
- Run the full validation sequence above.
- Update `CHANGELOG.md` for user-facing changes.
- Do not commit `vendor/`, build output, state files, credentials, or access tokens.

Dependency updates should keep `go.mod` and `go.sum` tidy and must pass `govulncheck`.

## Releases

Releases are produced from `v*` tags by `.github/workflows/release.yml`. Only maintainers with access to the configured GPG signing secrets should create release tags.
