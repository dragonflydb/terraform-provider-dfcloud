package resource_model

import (
	"context"
	"testing"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// planDatastore returns a minimal planned datastore, shaped like the value
// Terraform hands to Create when only the required attributes are configured.
func planDatastore() Datastore {
	return Datastore{
		Name: types.StringValue("test"),
		Location: DatastoreLocation{
			Provider:          types.StringValue("aws"),
			Region:            types.StringValue("eu-west-1"),
			AvailabilityZones: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("euw1-az2")}),
		},
		Tier: DatastoreTier{
			Memory:          types.Int64Value(3_000_000_000),
			PerformanceTier: types.StringValue("dev"),
			Replicas:        types.Int64Unknown(),
		},
		Dragonfly: types.ObjectUnknown(map[string]attr.Type{
			"cache_mode": types.BoolType,
			"tls":        types.BoolType,
			"bullmq":     types.BoolType,
			"sidekiq":    types.BoolType,
			"memcached":  types.BoolType,
			"acl_rules":  types.ListType{ElemType: types.StringType},
		}),
		FlagPolicyID: types.StringNull(),
	}
}

func TestIntoDatastoreConfigFlagPolicyID(t *testing.T) {
	tests := []struct {
		name  string
		value types.String
		want  string
	}{
		{name: "set", value: types.StringValue("fp-1"), want: "fp-1"},
		{name: "null", value: types.StringNull(), want: ""},
		{name: "unknown", value: types.StringUnknown(), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := planDatastore()
			in.FlagPolicyID = tt.value

			got := IntoDatastoreConfig(in).Config.FlagPolicyID
			if got != tt.want {
				t.Fatalf("Config.FlagPolicyID = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDatastoreFromConfigFlagPolicyID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want types.String
	}{
		{name: "set", id: "fp-1", want: types.StringValue("fp-1")},
		{name: "unset", id: "", want: types.StringNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Datastore
			d.FromConfig(context.Background(), &dfcloud.Datastore{
				ID: "ds-1",
				Config: dfcloud.DatastoreConfig{
					Name:         "test",
					FlagPolicyID: tt.id,
				},
			})

			if !d.FlagPolicyID.Equal(tt.want) {
				t.Fatalf("FlagPolicyID = %s, want %s", d.FlagPolicyID, tt.want)
			}
		})
	}
}

// TestDatastoreFlagPolicyIDRoundTrip checks that a configured flag policy
// survives the plan -> API -> state conversion unchanged, so applying a
// configuration produces no diff on the next plan.
func TestDatastoreFlagPolicyIDRoundTrip(t *testing.T) {
	plan := planDatastore()
	plan.FlagPolicyID = types.StringValue("fp-1")

	api := IntoDatastoreConfig(plan)

	var state Datastore
	state.FromConfig(context.Background(), api)

	if !state.FlagPolicyID.Equal(plan.FlagPolicyID) {
		t.Fatalf("FlagPolicyID after round trip = %s, want %s", state.FlagPolicyID, plan.FlagPolicyID)
	}
}
