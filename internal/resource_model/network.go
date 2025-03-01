package resource_model

import (
	"context"
	"time"

	dfcloud "github.com/dragonflydb/dfcloud/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkLocation struct {
	Provider types.String `tfsdk:"provider"`
	Region   types.String `tfsdk:"region"`
}

type Network struct {
	Id        types.String     `tfsdk:"id"`
	Name      types.String     `tfsdk:"name"`
	Location  *NetworkLocation `tfsdk:"location"`
	CidrBlock types.String     `tfsdk:"cidr_block"`
	CreatedAt types.Int64      `tfsdk:"created_at"`
	Status    types.String     `tfsdk:"status"`
	Vpc       types.Object     `tfsdk:"vpc"`
}

func IntoNetworkConfig(in Network) *dfcloud.NetworkConfig {
	return &dfcloud.NetworkConfig{
		Name: in.Name.ValueString(),
		Location: dfcloud.NetworkLocation{
			Provider: dfcloud.CloudProvider(in.Location.Provider.ValueString()),
			Region:   in.Location.Region.ValueString(),
		},
		CIDRBlock: in.CidrBlock.ValueString(),
	}
}

func FromNetworkConfig(in *dfcloud.Network) *Network {
	return &Network{
		Id:   types.StringValue(in.ID),
		Name: types.StringValue(in.Name),
		Location: &NetworkLocation{
			Provider: types.StringValue(string(in.Location.Provider)),
			Region:   types.StringValue(in.Location.Region),
		},
		CidrBlock: types.StringValue(in.CIDRBlock),
		CreatedAt: types.Int64Value(in.CreatedAt),
		Status:    types.StringValue(string(in.Status)),
		Vpc: types.ObjectValueMust(
			map[string]attr.Type{
				"resource_id": types.StringType,
				"account_id":  types.StringType,
			},
			map[string]attr.Value{
				"resource_id": types.StringValue(in.VPC.ResourceID),
				"account_id":  types.StringValue(in.VPC.AccountID),
			},
		),
	}
}

func WaitUntilNetworkStatus(ctx context.Context, client *dfcloud.Client, id string, status dfcloud.NetworkStatus) (*dfcloud.Network, error) {
	for {
		network, err := client.GetNetwork(ctx, id)
		if err != nil {
			return nil, err
		}

		if network.Status == status {
			return network, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
