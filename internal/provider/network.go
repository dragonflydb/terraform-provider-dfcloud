package provider

import (
	"context"
	"errors"
	"time"

	"github.com/dragonflydb/terraform-provider-dfcloud/internal/resource_model"
	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NetworkResource struct {
	client *dfcloud.Client
}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dfcloud_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dragonfly network.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "The timestamp when the network was created.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the network.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the network.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.SingleNestedAttribute{
				MarkdownDescription: "The location configuration for the network.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"provider": schema.StringAttribute{
						MarkdownDescription: "The provider for the network location.",
						Required:            true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: "The region for the network location.",
						Required:            true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"cidr_block": schema.StringAttribute{
				MarkdownDescription: "The CIDR block for the network.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc": schema.SingleNestedAttribute{
				MarkdownDescription: "The VPC information for the network.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"resource_id": schema.StringAttribute{
						MarkdownDescription: "The resource ID of the VPC.",
						Computed:            true,
					},
					"account_id": schema.StringAttribute{
						MarkdownDescription: "The account ID of the VPC.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dfcloud.Client)
	if !ok {
		resp.Diagnostics.AddError("failed to get provider", "failed to get provider")
		return
	}

	r.client = client
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state resource_model.Network
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("failed to get plan into state", "failed to get plan into state")
		return
	}

	networkConfig := resource_model.IntoNetworkConfig(state)
	respNetwork, err := r.client.CreateNetwork(ctx, networkConfig)
	if err != nil {
		resp.Diagnostics.AddError("failed to create network", err.Error())
		return
	}

	// wait until VPC IDs are created
	waitForNetworkStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	respNetwork, err = resource_model.WaitUntilNetworkStatus(waitForNetworkStatusCtx, r.client, respNetwork.ID, dfcloud.NetworkStatusActive)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for network", err.Error())
		return
	}

	state = *resource_model.FromNetworkConfig(respNetwork)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_model.Network
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	waitForNetworkStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	respNetwork, err := resource_model.WaitUntilNetworkStatus(waitForNetworkStatusCtx, r.client, state.Id.ValueString(), dfcloud.NetworkStatusActive)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for network", err.Error())
		return
	}

	if respNetwork.Status == dfcloud.NetworkStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state = *resource_model.FromNetworkConfig(respNetwork)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// no updates allowed for networks
	resp.Diagnostics.AddError(
		"Updating a Network is not supported",
		"Updating a Network is not supported",
	)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *resource_model.Network
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetwork(ctx, state.Id.ValueString())
	if errors.Is(err, dfcloud.ErrNotFound) {
		tflog.Warn(ctx, "network is already deleted", map[string]interface{}{
			"network_id": state.Id.ValueString(),
		})
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("failed to delete network", err.Error())
		return
	}

	// wait until network is deleted
	waitForNetworkStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	_, err = resource_model.WaitUntilNetworkStatus(waitForNetworkStatusCtx, r.client, state.Id.ValueString(), dfcloud.NetworkStatusDeleted)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for network deletion", err.Error())
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	network, err := r.client.GetNetwork(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to get network", err.Error())
		return
	}

	state := resource_model.FromNetworkConfig(network)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

var (
	_ resource.Resource                = &NetworkResource{}
	_ resource.ResourceWithImportState = &NetworkResource{}
)
