package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// clearProviderEnvironment keeps provider defaults sourced from the environment
// from changing test results on machines with Azure settings configured.
func clearProviderEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"AZRGRAPH_TENANT_ID",
		"AZRGRAPH_CLIENT_ID",
		"AZRGRAPH_CLIENT_SECRET",
		"AZRGRAPH_USE_AZURE_DEFAULT_CREDENTIAL",
		"AZURE_TOKEN_CREDENTIALS",
	} {
		// Setenv registers the original value for restoration after the test.
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}
	}
}

func TestConfigureDiagnosticsDoNotAccumulate(t *testing.T) {
	clearProviderEnvironment(t)

	p := New("test")()
	data := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"use_azure_default_credential": false,
	})

	for i := range 2 {
		_, diagnostics := p.ConfigureContextFunc(context.Background(), data)
		if len(diagnostics) != 1 {
			t.Fatalf("configure call %d returned %d diagnostics, want 1", i+1, len(diagnostics))
		}
	}
}

func TestConfigureRejectsIncompleteServicePrincipal(t *testing.T) {
	clearProviderEnvironment(t)

	testCases := map[string]map[string]any{
		"client ID only": {
			"client_id":                    "client-id",
			"use_azure_default_credential": true,
		},
		"client secret only": {
			"client_secret":                "secret",
			"use_azure_default_credential": true,
		},
		"missing tenant": {
			"client_id":                    "client-id",
			"client_secret":                "secret",
			"use_azure_default_credential": true,
		},
	}

	for name, raw := range testCases {
		t.Run(name, func(t *testing.T) {
			p := New("test")()
			data := schema.TestResourceDataRaw(t, p.Schema, raw)
			_, diagnostics := p.ConfigureContextFunc(context.Background(), data)
			if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Summary, "must all be set") {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}
		})
	}
}

func TestConfigureAllowsTenantWithDefaultCredential(t *testing.T) {
	clearProviderEnvironment(t)

	p := New("test")()
	data := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"tenant_id":                    "00000000-0000-0000-0000-000000000000",
		"use_azure_default_credential": true,
	})

	meta, diagnostics := p.ConfigureContextFunc(context.Background(), data)
	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, ok := meta.(*clients); !ok {
		t.Fatalf("unexpected provider metadata type %T", meta)
	}
}
