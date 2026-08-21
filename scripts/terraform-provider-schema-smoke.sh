#!/usr/bin/env bash
set -euo pipefail

repository_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
temporary_directory=$(mktemp -d)
trap 'rm -rf "$temporary_directory"' EXIT

provider_directory="$temporary_directory/plugins"
configuration_directory="$temporary_directory/configuration"
mkdir -p "$provider_directory" "$configuration_directory"

(
  cd "$repository_root"
  go build -o "$provider_directory/terraform-provider-azresourcegraph" .
)

cat >"$temporary_directory/terraform.rc" <<EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/nikhil-pandey/azresourcegraph" = "$provider_directory"
  }
  direct {}
}
EOF

cat >"$configuration_directory/main.tf" <<'EOF'
terraform {
  required_providers {
    azresourcegraph = {
      source = "nikhil-pandey/azresourcegraph"
    }
  }
}
EOF

TF_CLI_CONFIG_FILE="$temporary_directory/terraform.rc" \
  terraform -chdir="$configuration_directory" providers schema -json \
  >"$temporary_directory/schema.json"

grep -q '"azresourcegraph_query"' "$temporary_directory/schema.json"
