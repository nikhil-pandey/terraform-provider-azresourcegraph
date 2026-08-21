terraform {
  required_version = ">= 1.0"

  required_providers {
    azresourcegraph = {
      source = "nikhil-pandey/azresourcegraph"
    }
  }
}

# Authentication using Client Credentials
provider "azresourcegraph" {
  tenant_id     = var.tenant_id
  client_id     = var.client_id
  client_secret = var.client_secret
}

# Authentication using Azure Default Credential
# 1. Environment Variables
# 2. Workload Identity
# 3. Managed Identity
# 4. Azure CLI
# 5. Azure Developer CLI
# 6. Azure PowerShell
provider "azresourcegraph" {
  use_azure_default_credential = true
}
