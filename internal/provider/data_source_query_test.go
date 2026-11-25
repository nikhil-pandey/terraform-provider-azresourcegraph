package provider

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type fakeResourceGraphClient struct {
	responses  []armresourcegraph.ClientResourcesResponse
	err        error
	skipTokens []string
	requests   []armresourcegraph.QueryRequest
}

func (f *fakeResourceGraphClient) Resources(_ context.Context, request armresourcegraph.QueryRequest, _ *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
	f.requests = append(f.requests, request)
	if request.Options != nil && request.Options.SkipToken != nil {
		f.skipTokens = append(f.skipTokens, *request.Options.SkipToken)
	} else {
		f.skipTokens = append(f.skipTokens, "")
	}
	if f.err != nil {
		return armresourcegraph.ClientResourcesResponse{}, f.err
	}

	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestDoResourceQueryPaginates(t *testing.T) {
	next := "next-page"
	client := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{
			Data:      []any{map[string]any{"id": "one"}},
			SkipToken: &next,
		},
		{Data: []any{map[string]any{"id": "two"}}},
	}}

	results, err := doResourceQuery(context.Background(), client, armresourcegraph.QueryRequest{
		Options: &armresourcegraph.QueryRequestOptions{},
	})
	if err != nil {
		t.Fatalf("doResourceQuery returned an error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if !slices.Equal(client.skipTokens, []string{"", next}) {
		t.Fatalf("skip tokens = %#v, want %#v", client.skipTokens, []string{"", next})
	}
}

func TestDoResourceQueryReturnsClientError(t *testing.T) {
	client := &fakeResourceGraphClient{err: errors.New("service unavailable")}
	_, err := doResourceQuery(context.Background(), client, armresourcegraph.QueryRequest{})
	if err == nil || !strings.Contains(err.Error(), "service unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoResourceQueryRejectsUnexpectedResultFormat(t *testing.T) {
	client := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{Data: map[string]any{"unexpected": true}},
	}}

	_, err := doResourceQuery(context.Background(), client, armresourcegraph.QueryRequest{})
	if err == nil || !strings.Contains(err.Error(), "unexpected result type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDataSourceQueryReadProducesStableIDAndJSON(t *testing.T) {
	client := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{Data: []any{map[string]any{"id": "resource-one"}}},
		{Data: []any{map[string]any{"id": "resource-one"}}},
	}}
	resource := dataSourceQuery()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"query":            "Resources | project id",
		"subscription_ids": []any{"subscription-b", "subscription-a"},
	})

	var firstID string
	for i := range 2 {
		if diagnostics := dataSourceQueryRead(context.Background(), data, &clients{resourceGraph: client}); diagnostics.HasError() {
			t.Fatalf("read %d returned diagnostics: %#v", i+1, diagnostics)
		}
		if got := data.Get("result").(string); got != `[{"id":"resource-one"}]` {
			t.Fatalf("result = %q", got)
		}
		if i == 0 {
			firstID = data.Id()
			continue
		}
		if data.Id() != firstID {
			t.Fatalf("query ID changed between reads: %q != %q", data.Id(), firstID)
		}
	}
	if firstID == "" {
		t.Fatal("query ID is empty")
	}

	wantSubscriptions := []string{"subscription-a", "subscription-b"}
	gotSubscriptions := dereferenceStrings(client.requests[0].Subscriptions)
	if !slices.Equal(gotSubscriptions, wantSubscriptions) {
		t.Fatalf("subscriptions = %#v, want %#v", gotSubscriptions, wantSubscriptions)
	}

	changed := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"query":            "Resources | project name",
		"subscription_ids": []any{"subscription-a", "subscription-b"},
	})
	changedClient := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{Data: []any{}},
	}}
	if diagnostics := dataSourceQueryRead(context.Background(), changed, &clients{resourceGraph: changedClient}); diagnostics.HasError() {
		t.Fatalf("changed query returned diagnostics: %#v", diagnostics)
	}
	if changed.Id() == firstID {
		t.Fatal("different queries produced the same ID")
	}
}

func dereferenceStrings(values []*string) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = *value
	}
	return result
}

func TestDoResourceQueryRejectsRepeatedSkipToken(t *testing.T) {
	repeated := "same-page"
	client := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{Data: []any{}, SkipToken: &repeated},
		{Data: []any{}, SkipToken: &repeated},
	}}

	_, err := doResourceQuery(context.Background(), client, armresourcegraph.QueryRequest{})
	if err == nil || !strings.Contains(err.Error(), "repeated skip token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDataSourceQueryReadReturnsEmptyArrayWithoutMatches(t *testing.T) {
	client := &fakeResourceGraphClient{responses: []armresourcegraph.ClientResourcesResponse{
		{Data: []any{}},
	}}
	resource := dataSourceQuery()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"query": "Resources | where false",
	})

	if diagnostics := dataSourceQueryRead(context.Background(), data, &clients{resourceGraph: client}); diagnostics.HasError() {
		t.Fatalf("read returned diagnostics: %#v", diagnostics)
	}
	if got := data.Get("result").(string); got != "[]" {
		t.Fatalf("result = %q, want %q", got, "[]")
	}
}
