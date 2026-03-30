package resource_model

import (
	"context"
	"errors"
	"fmt"
	"time"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkLocation struct {
	Provider types.String `tfsdk:"provider"`
	Region   types.String `tfsdk:"region"`
}

type Network struct {
	Id            types.String     `tfsdk:"id"`
	Name          types.String     `tfsdk:"name"`
	Location      *NetworkLocation `tfsdk:"location"`
	CidrBlock     types.String     `tfsdk:"cidr_block"`
	CreatedAt     types.Int64      `tfsdk:"created_at"`
	Status        types.String     `tfsdk:"status"`
	Vpc           types.Object     `tfsdk:"vpc"`
	BYOCAccountID types.String     `tfsdk:"byoc_account_id"`
}

func IntoNetworkConfig(in Network) *dfcloud.NetworkConfig {
	cfg := &dfcloud.NetworkConfig{
		Name: in.Name.ValueString(),
		Location: dfcloud.NetworkLocation{
			Provider: dfcloud.CloudProvider(in.Location.Provider.ValueString()),
			Region:   in.Location.Region.ValueString(),
		},
		CIDRBlock: in.CidrBlock.ValueString(),
	}
	if !in.BYOCAccountID.IsNull() && !in.BYOCAccountID.IsUnknown() {
		cfg.BYOC.AccountID = in.BYOCAccountID.ValueString()
	}
	return cfg
}

func FromNetworkConfig(in *dfcloud.Network) *Network {
	n := &Network{
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
	if in.BYOC.AccountID != "" {
		n.BYOCAccountID = types.StringValue(in.BYOC.AccountID)
	} else {
		n.BYOCAccountID = types.StringNull()
	}
	return n
}

func WaitUntilNetworkStatus(ctx context.Context, client *dfcloud.Client, id string, status dfcloud.NetworkStatus) (*dfcloud.Network, error) {
	if id == "" {
		return nil, fmt.Errorf("missing network id")
	}
	for {
		network, err := client.GetNetwork(ctx, id)
		if errors.Is(err, dfcloud.ErrNotFound) {
			if status == dfcloud.NetworkStatusDeleted {
				return &dfcloud.Network{
					ID:     id,
					Status: dfcloud.NetworkStatusDeleted,
				}, nil
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
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
