package provider

import (
	"context"
	"fmt"

	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FlagPolicyDataSource struct {
	client *dfcloud.Client
}

type flagPolicyDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Flags       []types.String `tfsdk:"flags"`
	CreatedAt   types.Int64    `tfsdk:"created_at"`
}

func NewFlagPolicyDataSource() datasource.DataSource {
	return &FlagPolicyDataSource{}
}

func (d *FlagPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "dfcloud_flag_policy"
}

func (d *FlagPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Dragonfly flag policy by name. Flag policies are created by Dragonfly Cloud operators and can be applied to datastores.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the flag policy.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the flag policy to look up.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the flag policy.",
				Computed:            true,
			},
			"flags": schema.ListAttribute{
				MarkdownDescription: "The Dragonfly flags applied by the policy.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "The timestamp when the flag policy was created.",
				Computed:            true,
			},
		},
	}
}

func (d *FlagPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dfcloud.Client)
	if !ok {
		resp.Diagnostics.AddError("failed to get provider", "failed to get provider")
		return
	}

	d.client = client
}

func (d *FlagPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config flagPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListFlagPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to list flag policies", err.Error())
		return
	}

	var found *dfcloud.FlagPolicy
	for _, p := range policies {
		if p.Name == config.Name.ValueString() {
			found = p
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("flag policy not found", fmt.Sprintf("no flag policy named %q", config.Name.ValueString()))
		return
	}

	state := flagPolicyDataSourceModel{
		ID:          types.StringValue(found.ID),
		Name:        types.StringValue(found.Name),
		Description: types.StringNull(),
		Flags:       make([]types.String, 0, len(found.Flags)),
		CreatedAt:   types.Int64Value(found.CreatedAt),
	}
	if found.Description != nil {
		state.Description = types.StringValue(*found.Description)
	}
	for _, f := range found.Flags {
		state.Flags = append(state.Flags, types.StringValue(f))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
