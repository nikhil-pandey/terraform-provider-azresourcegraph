GO ?= go

.PHONY: default download verify tidy fmt test test-race vet lint workflow-lint vuln generate generate-check terraform-smoke check release-check

default: check

download:
	$(GO) mod download

verify:
	$(GO) mod verify

tidy:
	$(GO) mod tidy -diff

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race -cover ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

workflow-lint:
	$(GO) tool actionlint

vuln:
	$(GO) tool govulncheck ./...

generate:
	$(GO) generate ./...

generate-check: generate
	git diff --compact-summary --exit-code

terraform-smoke:
	./scripts/terraform-provider-schema-smoke.sh

check: verify tidy test-race vet lint workflow-lint vuln

release-check: check generate-check
	goreleaser check
