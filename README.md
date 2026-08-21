# Terraform Provider for Azure Resource Graph (`azresourcegraph`)

[![Tests](https://github.com/nikhil-pandey/terraform-provider-azresourcegraph/actions/workflows/test.yml/badge.svg)](https://github.com/nikhil-pandey/terraform-provider-azresourcegraph/actions/workflows/test.yml)

This repository is a maintained fork of [`tiwood/terraform-provider-azresourcegraph`](https://github.com/tiwood/terraform-provider-azresourcegraph). It preserves the original provider's history while continuing development and releases under the `nikhil-pandey` namespace.

The provider queries [Azure Resource Graph](https://learn.microsoft.com/azure/governance/resource-graph/overview) and exposes results for use with other Terraform providers and modules.

## Requirements

- Terraform 1.0 or newer
- Go 1.26 or newer for development; the repository selects Go 1.26.7 locally and CI tracks the latest Go 1.26 patch

## Usage

```terraform
terraform {
  required_providers {
    azresourcegraph = {
      source = "nikhil-pandey/azresourcegraph"
    }
  }
}

provider "azresourcegraph" {}

data "azresourcegraph_query" "resource_ids" {
  query = "Resources | project id"
}

output "resource_ids" {
  value = jsondecode(data.azresourcegraph_query.resource_ids.result)
}
```

## Authentication

Azure Default Credential is enabled by default. Its chain includes environment credentials, workload identity, managed identity, Azure CLI, Azure Developer CLI, and Azure PowerShell. `AZURE_TOKEN_CREDENTIALS` can restrict the chain.

An optional `tenant_id` is passed to Azure Default Credential as its default tenant:

```terraform
provider "azresourcegraph" {
  tenant_id = var.tenant_id
}
```

Service-principal authentication takes precedence when all three values are set:

```terraform
provider "azresourcegraph" {
  tenant_id     = var.tenant_id
  client_id     = var.client_id
  client_secret = var.client_secret
}
```

Provider settings can also be sourced from `AZRGRAPH_TENANT_ID`, `AZRGRAPH_CLIENT_ID`, `AZRGRAPH_CLIENT_SECRET`, and `AZRGRAPH_USE_AZURE_DEFAULT_CREDENTIAL`.

## Development

```sh
go mod download
go mod verify
go test -race -cover ./...
go vet ./...
go generate ./...
```

`go generate ./...` requires Terraform on `PATH` and updates the generated Registry documentation in `docs/`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the complete workflow and [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Releases

Tags matching `v*` run GoReleaser and publish signed Terraform Registry artifacts. Maintainers must configure the `GPG_PRIVATE_KEY` and `PASSPHRASE` repository secrets before creating a release tag.

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).
