package resource_model

import (
	"context"
	"errors"
	"fmt"
	"time"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Connection struct {
	ConnectionID types.String     `tfsdk:"connection_id"`
	Name         types.String     `tfsdk:"name"`
	NetworkID    types.String     `tfsdk:"network_id"`
	Peer         *PeerConfigModel `tfsdk:"peer"`
	Status       types.String     `tfsdk:"status"`
	StatusDetail types.String     `tfsdk:"status_detail"`
	PeerConnID   types.String     `tfsdk:"peer_connection_id"`
}

type PeerConfigModel struct {
	AccountID              types.String `tfsdk:"account_id"`
	VPCID                  types.String `tfsdk:"vpc_id"`
	Region                 types.String `tfsdk:"region"`
	AzureResourceGroup     types.String `tfsdk:"azure_resource_group"`
	AzureTenantID          types.String `tfsdk:"azure_tenant_id"`
	AzureAppObjectID       types.String `tfsdk:"azure_app_object_id"`
	AzureUseRemoteGateways types.Bool   `tfsdk:"azure_use_remote_gateways"`
}

func IntoPeerConfig(in *PeerConfigModel) dfcloud.PeerConfig {
	return dfcloud.PeerConfig{
		AccountID: in.AccountID.ValueString(),
		VPCID:     in.VPCID.ValueString(),
		Region:    in.Region.ValueString(),
		AzureConfig: dfcloud.AzureConfig{
			ResourceGroup:     in.AzureResourceGroup.ValueString(),
			TenantID:          in.AzureTenantID.ValueString(),
			AppObjectID:       in.AzureAppObjectID.ValueString(),
			UseRemoteGateways: in.AzureUseRemoteGateways.ValueBool(),
		},
	}
}

func IntoConnectionConfig(in Connection) *dfcloud.ConnectionConfig {
	return &dfcloud.ConnectionConfig{
		Name:      in.Name.ValueString(),
		NetworkID: in.NetworkID.ValueString(),
		Peer:      IntoPeerConfig(in.Peer),
	}
}

func optionalString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func FromConnectionConfig(in *dfcloud.Connection) *Connection {
	az := in.Config.Peer.AzureConfig
	useRemoteGateways := types.BoolNull()
	if az.TenantID != "" {
		useRemoteGateways = types.BoolValue(az.UseRemoteGateways)
	}
	peer := &PeerConfigModel{
		AccountID:              types.StringValue(in.Config.Peer.AccountID),
		VPCID:                  types.StringValue(in.Config.Peer.VPCID),
		Region:                 types.StringValue(in.Config.Peer.Region),
		AzureResourceGroup:     optionalString(az.ResourceGroup),
		AzureTenantID:          optionalString(az.TenantID),
		AzureAppObjectID:       optionalString(az.AppObjectID),
		AzureUseRemoteGateways: useRemoteGateways,
	}
	return &Connection{
		ConnectionID: types.StringValue(in.ID),
		Name:         types.StringValue(in.Config.Name),
		NetworkID:    types.StringValue(in.Config.NetworkID),
		Peer:         peer,
		Status:       types.StringValue(string(in.Status)),
		StatusDetail: types.StringValue(in.StatusDetail),
		PeerConnID:   types.StringValue(in.PeerConnectionID),
	}
}

func WaitUntilConnectionStatus(ctx context.Context, client *dfcloud.Client, id string, status dfcloud.ConnectionStatus) (*dfcloud.Connection, error) {
	if id == "" {
		return nil, fmt.Errorf("missing connection id")
	}
	for {
		conn, err := client.GetConnection(ctx, id)
		if errors.Is(err, dfcloud.ErrNotFound) {
			if status == dfcloud.ConnectionStatusDeleted {
				return &dfcloud.Connection{
					ID:     id,
					Status: dfcloud.ConnectionStatusDeleted,
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

		if conn.Status == status {
			return conn, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):

		}
	}
}
