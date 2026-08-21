package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceQuery() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for querying resources managed by Azure Resource Manager.",

		ReadContext: dataSourceQueryRead,

		Schema: map[string]*schema.Schema{
			"query": {
				Description: "The query to execute.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"subscription_ids": {
				Description:   `Azure subscription ids against which to execute the query.`,
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"management_group_ids"},
				MinItems:      1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"management_group_ids": {
				Description:   `Azure management groups against which to execute the query.`,
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"subscription_ids"},
				MinItems:      1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"result": {
				Description: `The query output in raw JSON format.`,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients).resourceGraph
	query := d.Get("query").(string)
	subscriptionIDs := sortedSetValues(d, "subscription_ids")
	managementGroupIDs := sortedSetValues(d, "management_group_ids")

	format := armresourcegraph.ResultFormatObjectArray
	opts := armresourcegraph.QueryRequestOptions{
		ResultFormat: &format,
	}

	queryRequest := armresourcegraph.QueryRequest{
		Options: &opts,
		Query:   &query,
	}

	if len(subscriptionIDs) > 0 {
		queryRequest.Subscriptions = stringPointers(subscriptionIDs)
	}
	if len(managementGroupIDs) > 0 {
		queryRequest.ManagementGroups = stringPointers(managementGroupIDs)
	}

	data, err := doResourceQuery(ctx, client, queryRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return diag.Errorf("failed to marshal query results to JSON string: %v", err)
	}

	id, err := queryID(query, subscriptionIDs, managementGroupIDs)
	if err != nil {
		return diag.Errorf("failed to calculate query ID: %v", err)
	}
	d.SetId(id)
	if err := d.Set("result", string(jsonData)); err != nil {
		return diag.Errorf("failed to set query result: %v", err)
	}

	return nil
}

func doResourceQuery(ctx context.Context, client resourceGraphClient, queryRequest armresourcegraph.QueryRequest) ([]interface{}, error) {
	var results []interface{}

	for {
		resp, err := client.Resources(ctx, queryRequest, nil)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		if resp.Data != nil {
			page, ok := resp.Data.([]interface{})
			if !ok {
				return nil, fmt.Errorf("query returned unexpected result type %T", resp.Data)
			}
			results = append(results, page...)
		}
		if resp.SkipToken == nil || *resp.SkipToken == "" {
			break
		}
		if queryRequest.Options == nil {
			queryRequest.Options = &armresourcegraph.QueryRequestOptions{}
		}
		queryRequest.Options.SkipToken = resp.SkipToken
	}

	return results, nil
}

func sortedSetValues(d *schema.ResourceData, key string) []string {
	set, ok := d.GetOk(key)
	if !ok {
		return nil
	}

	values := set.(*schema.Set).List()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.(string)
	}
	sort.Strings(result)
	return result
}

func stringPointers(values []string) []*string {
	result := make([]*string, len(values))
	for i := range values {
		result[i] = &values[i]
	}
	return result
}

func queryID(query string, subscriptionIDs, managementGroupIDs []string) (string, error) {
	payload, err := json.Marshal(struct {
		Query              string   `json:"query"`
		SubscriptionIDs    []string `json:"subscription_ids,omitempty"`
		ManagementGroupIDs []string `json:"management_group_ids,omitempty"`
	}{
		Query:              query,
		SubscriptionIDs:    subscriptionIDs,
		ManagementGroupIDs: managementGroupIDs,
	})
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}
