package resource_model

import (
	"context"
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
	AccountID types.String `tfsdk:"account_id"`
	VPCID     types.String `tfsdk:"vpc_id"`
	Region    types.String `tfsdk:"region"`
}

func IntoPeerConfig(in PeerConfigModel) dfcloud.PeerConfig {
	return dfcloud.PeerConfig{
		AccountID: in.AccountID.ValueString(),
		VPCID:     in.VPCID.ValueString(),
		Region:    in.Region.ValueString(),
	}
}

func IntoConnectionConfig(in Connection) *dfcloud.ConnectionConfig {
	return &dfcloud.ConnectionConfig{
		Name:      in.Name.ValueString(),
		NetworkID: in.NetworkID.ValueString(),
		Peer:      IntoPeerConfig(*in.Peer),
	}
}

func FromConnectionConfig(in *dfcloud.Connection) *Connection {
	return &Connection{
		ConnectionID: types.StringValue(in.ID),
		Name:         types.StringValue(in.Config.Name),
		NetworkID:    types.StringValue(in.Config.NetworkID),
		Peer: &PeerConfigModel{
			AccountID: types.StringValue(in.Config.Peer.AccountID),
			VPCID:     types.StringValue(in.Config.Peer.VPCID),
			Region:    types.StringValue(in.Config.Peer.Region),
		},
		Status:       types.StringValue(string(in.Status)),
		StatusDetail: types.StringValue(in.StatusDetail),
		PeerConnID:   types.StringValue(in.PeerConnectionID),
	}
}

func WaitUntilConnectionStatus(ctx context.Context, client *dfcloud.Client, id string, status dfcloud.ConnectionStatus) (*dfcloud.Connection, error) {
	for {
		conn, err := client.GetConnection(ctx, id)
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
