package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestConfigureDiagnosticsDoNotAccumulate(t *testing.T) {
	p := New("test")()
	data := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"use_azure_default_credential": false,
	})

	for i := 0; i < 2; i++ {
		_, diagnostics := p.ConfigureContextFunc(context.Background(), data)
		if len(diagnostics) != 1 {
			t.Fatalf("configure call %d returned %d diagnostics, want 1", i+1, len(diagnostics))
		}
	}
}

func TestConfigureRejectsIncompleteServicePrincipal(t *testing.T) {
	testCases := map[string]map[string]interface{}{
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
	p := New("test")()
	data := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
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
