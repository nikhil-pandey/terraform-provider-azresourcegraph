package provider

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"tenant_id": {
					Description: "The tenant ID used for service-principal authentication or as the default tenant for Azure Default Credential. This can also be sourced from the `AZRGRAPH_TENANT_ID` environment variable.",
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZRGRAPH_TENANT_ID", nil),
				},
				"client_id": {
					Description: "The client ID used for service-principal authentication. This can also be sourced from the `AZRGRAPH_CLIENT_ID` environment variable.",
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZRGRAPH_CLIENT_ID", nil),
				},
				"client_secret": {
					Description: "The client secret used for service-principal authentication. This can also be sourced from the `AZRGRAPH_CLIENT_SECRET` environment variable.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("AZRGRAPH_CLIENT_SECRET", nil),
				},
				"use_azure_default_credential": {
					Description: "Use Azure Default Credential for authentication. The default chain tries environment credentials, workload identity, managed identity, Azure CLI, Azure Developer CLI, and Azure PowerShell, in that order. The `AZURE_TOKEN_CREDENTIALS` environment variable can restrict the chain. Complete service-principal credentials take precedence. This can also be sourced from the `AZRGRAPH_USE_AZURE_DEFAULT_CREDENTIAL` environment variable. Defaults to true.",
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AZRGRAPH_USE_AZURE_DEFAULT_CREDENTIAL", true),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"azresourcegraph_query": dataSourceQuery(),
			},
			ResourcesMap: map[string]*schema.Resource{},
		}

		p.ConfigureContextFunc = configure

		return p
	}
}

type resourceGraphClient interface {
	Resources(context.Context, armresourcegraph.QueryRequest, *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error)
}

type clients struct {
	resourceGraph resourceGraphClient
}

func configure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	tenantID := d.Get("tenant_id").(string)
	useDefaultCredential := d.Get("use_azure_default_credential").(bool)

	hasClientID := clientID != ""
	hasClientSecret := clientSecret != ""
	if hasClientID != hasClientSecret || (hasClientID && tenantID == "") {
		return nil, diag.Errorf("client_id, client_secret, and tenant_id must all be set to use service-principal authentication")
	}

	var (
		credential azcore.TokenCredential
		err        error
	)

	switch {
	case hasClientID:
		credential, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	case useDefaultCredential:
		credential, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			TenantID: tenantID,
		})
	default:
		return nil, diag.Errorf("no authentication credentials provided")
	}
	if err != nil {
		return nil, diag.Errorf("unable to create Azure credential: %v", err)
	}

	clientFactory, err := armresourcegraph.NewClientFactory(credential, nil)
	if err != nil {
		return nil, diag.Errorf("unable to create Resource Graph client factory: %v", err)
	}

	return &clients{resourceGraph: clientFactory.NewClient()}, nil
}
