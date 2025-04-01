package resource_model

import (
	"context"
	"time"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// Datastore maps the resource schema data.
type Datastore struct {
	ID             types.String      `tfsdk:"id"`
	Name           types.String      `tfsdk:"name"`
	NetworkId      types.String      `tfsdk:"network_id"`
	Location       DatastoreLocation `tfsdk:"location"`
	Tier           DatastoreTier     `tfsdk:"tier"`
	Dragonfly      types.Object      `tfsdk:"dragonfly"`
	CreatedAt      types.Int64       `tfsdk:"created_at"`
	Password       types.String      `tfsdk:"password"`
	Addr           types.String      `tfsdk:"addr"`
	DisablePassKey types.Bool        `tfsdk:"disable_pass_key"`
}

type DatastoreLocation struct {
	Provider          types.String `tfsdk:"provider"`
	Region            types.String `tfsdk:"region"`
	AvailabilityZones types.List   `tfsdk:"availability_zones"`
}

type DatastoreTier struct {
	Memory          types.Int64  `tfsdk:"max_memory_bytes"`
	PerformanceTier types.String `tfsdk:"performance_tier"`
	Replicas        types.Int64  `tfsdk:"replicas"`
}

func (d *Datastore) FromConfig(ctx context.Context, in *dfcloud.Datastore) {
	d.ID = types.StringValue(in.ID)
	d.Name = types.StringValue(in.Config.Name)
	d.NetworkId = types.StringNull()
	d.CreatedAt = types.Int64Value(in.CreatedAt)
	d.Location.Provider = types.StringValue(string(in.Config.Location.Provider))
	d.Location.Region = types.StringValue(in.Config.Location.Region)
	d.Location.AvailabilityZones, _ = types.ListValueFrom(ctx, types.StringType, in.Config.Location.AvailabilityZones)
	d.Addr = types.StringValue(in.Addr)
	d.Password = types.StringValue(in.Key)
	d.Tier.Memory = types.Int64Value(int64(in.Config.Tier.Memory))
	d.Tier.PerformanceTier = types.StringValue(string(in.Config.Tier.PerformanceTier))
	d.Tier.Replicas = types.Int64Value(int64(*in.Config.Tier.Replicas))

	aclRules, _ := types.ListValueFrom(ctx, types.StringType, in.Config.Dragonfly.AclRules)
	d.Dragonfly = types.ObjectValueMust(map[string]attr.Type{
		"cache_mode": types.BoolType,
		"tls":        types.BoolType,
		"bullmq":     types.BoolType,
		"sidekiq":    types.BoolType,
		"memcached":  types.BoolType,
		"acl_rules":  types.ListType{ElemType: types.StringType},
	}, map[string]attr.Value{
		"cache_mode": types.BoolPointerValue(in.Config.Dragonfly.CacheMode),
		"tls":        types.BoolPointerValue(in.Config.Dragonfly.TLS),
		"bullmq":     types.BoolPointerValue(in.Config.Dragonfly.BullMQ),
		"sidekiq":    types.BoolPointerValue(in.Config.Dragonfly.Sidekiq),
		"memcached":  types.BoolPointerValue(in.Config.Dragonfly.Memcached),
		"acl_rules":  aclRules,
	})

	if in.Config.NetworkID != "" {
		d.NetworkId = types.StringValue(in.Config.NetworkID)
	}

}

func IntoDatastoreConfig(in Datastore) *dfcloud.Datastore {
	datastore := &dfcloud.Datastore{
		ID: in.ID.ValueString(),
		Config: dfcloud.DatastoreConfig{
			Name: in.Name.ValueString(),
			Location: dfcloud.DatastoreLocation{
				Provider: dfcloud.CloudProvider(in.Location.Provider.ValueString()),
				Region:   in.Location.Region.ValueString(),
			},
			Tier: dfcloud.DatastoreTier{
				Memory:          uint64(in.Tier.Memory.ValueInt64()),
				PerformanceTier: dfcloud.PerformanceTier(in.Tier.PerformanceTier.ValueString()),
				Replicas:        lo.ToPtr(int(in.Tier.Replicas.ValueInt64())),
			},
		},
	}

	_ = in.Location.AvailabilityZones.ElementsAs(context.Background(), &datastore.Config.Location.AvailabilityZones, false)

	if !in.NetworkId.IsNull() {
		datastore.Config.NetworkID = in.NetworkId.ValueString()
	}

	if in.DisablePassKey.ValueBool() {
		datastore.Config.DisablePasskey = in.DisablePassKey.ValueBool()
	}

	if in.Dragonfly.IsNull() {
		in.Dragonfly = types.ObjectValueMust(map[string]attr.Type{
			"cache_mode": types.BoolType,
			"tls":        types.BoolType,
			"bullmq":     types.BoolType,
			"sidekiq":    types.BoolType,
			"memcached":  types.BoolType,
			"acl_rules":  types.ListType{ElemType: types.StringType},
		}, map[string]attr.Value{})
	}

	if in.Dragonfly.Attributes()["cache_mode"] != nil {
		datastore.Config.Dragonfly.CacheMode = lo.ToPtr(in.Dragonfly.Attributes()["cache_mode"].(types.Bool).ValueBool())
	}
	if in.Dragonfly.Attributes()["tls"] != nil {
		datastore.Config.Dragonfly.TLS = lo.ToPtr(in.Dragonfly.Attributes()["tls"].(types.Bool).ValueBool())
	}
	if in.Dragonfly.Attributes()["bullmq"] != nil {
		datastore.Config.Dragonfly.BullMQ = lo.ToPtr(in.Dragonfly.Attributes()["bullmq"].(types.Bool).ValueBool())
	}
	if in.Dragonfly.Attributes()["sidekiq"] != nil {
		datastore.Config.Dragonfly.Sidekiq = lo.ToPtr(in.Dragonfly.Attributes()["sidekiq"].(types.Bool).ValueBool())
	}
	if in.Dragonfly.Attributes()["memcached"] != nil {
		datastore.Config.Dragonfly.Memcached = lo.ToPtr(in.Dragonfly.Attributes()["memcached"].(types.Bool).ValueBool())
	}

	if in.Dragonfly.Attributes()["acl_rules"] != nil {
		var rules dfcloud.AclRuleArray
		for _, rule := range in.Dragonfly.Attributes()["acl_rules"].(types.List).Elements() {
			rules = append(rules, rule.String())
		}
		datastore.Config.Dragonfly.AclRules = &rules
	}

	return datastore
}

// Helper function to wait for datastore to become active.
func WaitForDatastoreStatus(ctx context.Context, client *dfcloud.Client, id string, status dfcloud.DatastoreStatus) (*dfcloud.Datastore, error) {
	for {
		datastore, err := client.GetDatastore(ctx, id)
		if err != nil {
			return nil, err
		}

		if datastore.Status == status {
			return datastore, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
