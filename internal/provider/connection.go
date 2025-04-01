package provider

import (
	"context"
	"time"

	"github.com/dragonflydb/terraform-provider-dfcloud/internal/resource_model"
	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectionResource struct {
	client *dfcloud.Client
}

func NewConnectionResource() resource.Resource {
	return &ConnectionResource{}
}

func (r *ConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dfcloud_connection"
}

func (r *ConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dragonfly network connection.",

		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the connection.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the connection.",
				Computed:            true,
			},
			"status_detail": schema.StringAttribute{
				MarkdownDescription: "Additional details about the connection status.",
				Computed:            true,
			},
			"peer_connection_id": schema.StringAttribute{
				MarkdownDescription: "The underlying cloud provider connection ID.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the connection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network to connect to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"peer": schema.SingleNestedAttribute{
				MarkdownDescription: "The VPC to connect to.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						MarkdownDescription: "The account ID of the target VPC.",
						Required:            true,
					},
					"vpc_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the target VPC.",
						Required:            true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: "The region of the target VPC.",
						Optional:            true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dfcloud.Client)
	if !ok {
		resp.Diagnostics.AddError("failed to get provider", "failed to get provider")
	}

	r.client = client
}

func (r *ConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state resource_model.Connection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("failed to get plan into state", "failed to get plan into state")
	}

	connConfig := resource_model.IntoConnectionConfig(state)
	respConn, err := r.client.CreateConnection(ctx, connConfig)
	if err != nil {
		resp.Diagnostics.AddError("failed to create connection", err.Error())
		return
	}

	// wait until VPC IDs are created
	waitForConnectionStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	respConn, err = resource_model.WaitUntilConnectionStatus(waitForConnectionStatusCtx, r.client, respConn.ID, dfcloud.ConnectionStatusInactive)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for connection", err.Error())
		return
	}

	state.ConnectionID = types.StringValue(respConn.ID)
	state.Status = types.StringValue(string(respConn.Status))
	state.StatusDetail = types.StringValue(respConn.StatusDetail)
	state.PeerConnID = types.StringValue(respConn.PeerConnectionID)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_model.Connection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	respConn, err := r.client.GetConnection(ctx, state.ConnectionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read connection", err.Error())
		return
	}

	if respConn.Status == dfcloud.ConnectionStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ConnectionID = types.StringValue(respConn.ID)
	state.Status = types.StringValue(string(respConn.Status))
	state.StatusDetail = types.StringValue(respConn.StatusDetail)
	state.PeerConnID = types.StringValue(respConn.PeerConnectionID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resource_model.Connection
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing connection
	err := r.client.DeleteConnection(ctx, state.ConnectionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete old connection", err.Error())
		return
	}

	// Wait for deletion
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = resource_model.WaitUntilConnectionStatus(waitCtx, r.client, state.ConnectionID.ValueString(), dfcloud.ConnectionStatusDeleted)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for connection deletion", err.Error())
		return
	}

	// Create new connection
	connConfig := resource_model.IntoConnectionConfig(plan)
	respConn, err := r.client.CreateConnection(ctx, connConfig)
	if err != nil {
		resp.Diagnostics.AddError("failed to create new connection", err.Error())
		return
	}

	// Wait for creation
	waitCtx2, cancel2 := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel2()
	respConn, err = resource_model.WaitUntilConnectionStatus(waitCtx2, r.client, respConn.ID, dfcloud.ConnectionStatusInactive)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for new connection", err.Error())
		return
	}

	// Update state
	plan.ConnectionID = types.StringValue(respConn.ID)
	plan.Status = types.StringValue(string(respConn.Status))
	plan.StatusDetail = types.StringValue(respConn.StatusDetail)
	plan.PeerConnID = types.StringValue(respConn.PeerConnectionID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *resource_model.Connection
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteConnection(ctx, state.ConnectionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete connection", err.Error())
	}

	// wait until connection is deleted
	waitForConnectionStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = resource_model.WaitUntilConnectionStatus(waitForConnectionStatusCtx, r.client, state.ConnectionID.ValueString(), dfcloud.ConnectionStatusDeleted)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for connection", err.Error())
	}

}

func (r *ConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	connection, err := r.client.GetConnection(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to get network", err.Error())
		return
	}

	state := resource_model.FromConnectionConfig(connection)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

var (
	_ resource.Resource                = &ConnectionResource{}
	_ resource.ResourceWithImportState = &ConnectionResource{}
)
