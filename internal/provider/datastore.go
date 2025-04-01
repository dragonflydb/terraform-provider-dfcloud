package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/dragonflydb/terraform-provider-dfcloud/internal/resource_model"
	dfcloud "github.com/dragonflydb/terraform-provider-dfcloud/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewDatastoreResource is a helper function to simplify the provider implementation.
func NewDatastoreResource() resource.Resource {
	return &datastoreResource{}
}

// datastoreResource is the resource implementation.
type datastoreResource struct {
	client *dfcloud.Client
}

// Metadata returns the resource type name.
func (r *datastoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "datastore"
}

// Schema defines the schema for the resource.
func (r *datastoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dragonfly datastore resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the datastore.",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "The timestamp when the datastore was created.",
				Computed:            true,
			},
			"disable_pass_key": schema.BoolAttribute{
				MarkdownDescription: "Disable the passkey for the datastore.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			// password cant be set by a user
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for the datastore.",
				Optional:            false,
				Required:            false,
				Computed:            true,
				Sensitive:           true,
			},
			"addr": schema.StringAttribute{
				MarkdownDescription: "The address of the datastore.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the datastore.",
				Required:            true,
			},
			"location": schema.SingleNestedAttribute{
				MarkdownDescription: "The location configuration for the datastore.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"provider": schema.StringAttribute{
						MarkdownDescription: "The provider for the datastore location.",
						Required:            true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: "The region for the datastore location.",
						Required:            true,
					},
					"availability_zones": schema.ListAttribute{
						MarkdownDescription: "The availability zones for the datastore location.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"tier": schema.SingleNestedAttribute{
				MarkdownDescription: "The tier configuration for the datastore.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"max_memory_bytes": schema.Int64Attribute{
						MarkdownDescription: "The maximum memory (in bytes) for the datastore.",
						Required:            true,
					},
					"performance_tier": schema.StringAttribute{
						MarkdownDescription: "The performance tier for the datastore.",
						Required:            true,
					},
					"replicas": schema.Int64Attribute{
						MarkdownDescription: "The number of replicas for the datastore.",
						Optional:            true,
					},
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network the datastore should be placed into.",
				Optional:            true,
			},
			"dragonfly": schema.SingleNestedAttribute{
				MarkdownDescription: "Dragonfly-specific configuration.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cache_mode": schema.BoolAttribute{
						MarkdownDescription: "Enable cache mode for memory management.",
						Optional:            true,
						Computed:            true,
					},
					"bullmq": schema.BoolAttribute{
						MarkdownDescription: "Enable BullMQ compatibility.",
						Optional:            true,
						Computed:            true,
					},
					"tls": schema.BoolAttribute{
						MarkdownDescription: "Enable TLS.",
						Optional:            true,
						Computed:            true,
					},
					"sidekiq": schema.BoolAttribute{
						MarkdownDescription: "Enable Sidekiq compatibility.",
						Optional:            true,
						Computed:            true,
					},
					"memcached": schema.BoolAttribute{
						MarkdownDescription: "Enable Memcached protocol.",
						Optional:            true,
						Computed:            true,
					},
					"acl_rules": schema.ListAttribute{
						MarkdownDescription: "List of ACL rules.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *datastoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dfcloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *dfcloud.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create a new resource.
func (r *datastoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_model.Datastore
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	datastore := resource_model.IntoDatastoreConfig(plan)
	if datastore == nil {
		resp.Diagnostics.AddError("Configuration Error", "Failed to create datastore configuration")
		return
	}

	respDatastore, err := r.client.CreateDatastore(ctx, &datastore.Config)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Datastore", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	respDatastore, err = resource_model.WaitForDatastoreStatus(ctx, r.client, respDatastore.ID, dfcloud.DatastoreStatusActive)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Datastore", err.Error())
		return
	}

	tflog.Info(ctx, "created datastore", map[string]interface{}{
		"datastore_id": respDatastore.ID,
		"status":       respDatastore.Status,
	})

	plan.FromConfig(respDatastore)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *datastoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_model.Datastore
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	respDatastore, err := r.client.GetDatastore(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Datastore", err.Error())
		return
	}

	if respDatastore.Status == dfcloud.DatastoreStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Info(ctx, "read datastore", map[string]interface{}{
		"datastore_id": respDatastore.ID,
		"status":       respDatastore.Status,
	})

	state.FromConfig(respDatastore)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *datastoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state resource_model.Datastore
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retreive datastore to check if it is active
	respDatastore, err := r.client.GetDatastore(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Datastore", err.Error())
	}

	if respDatastore.Status == dfcloud.DatastoreStatusUpdating || respDatastore.Status == dfcloud.DatastoreStatusPending || respDatastore.Status == dfcloud.DatastoreStatusDeleting {
		resp.Diagnostics.AddError("Error Reading Datastore", "Datastore is not active")
	}

	var plan resource_model.Datastore
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateDatastore := resource_model.IntoDatastoreConfig(plan)
	respDatastore, err = r.client.UpdateDatastore(ctx, state.ID.ValueString(), &updateDatastore.Config)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Datastore", err.Error())
		return
	}

	waitForDatastoreStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	respDatastore, err = resource_model.WaitForDatastoreStatus(waitForDatastoreStatusCtx, r.client, respDatastore.ID, dfcloud.DatastoreStatusActive)
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for Datastore Update", err.Error())
		return
	}

	tflog.Info(ctx, "updated datastore", map[string]interface{}{
		"datastore_id": respDatastore.ID,
		"status":       respDatastore.Status,
	})

	plan.FromConfig(respDatastore)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *datastoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_model.Datastore
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDatastore(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Datastore", err.Error())
	}

	waitForDatastoreStatusCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = resource_model.WaitForDatastoreStatus(waitForDatastoreStatusCtx, r.client, state.ID.ValueString(), dfcloud.DatastoreStatusDeleted)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Datastore", err.Error())
	}

	tflog.Info(ctx, "deleted datastore", map[string]interface{}{
		"datastore_id": state.ID.ValueString(),
	})
}

// ImportState imports the resource state from an external system.
func (r *datastoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	datastore, err := r.client.GetDatastore(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Datastore", err.Error())
		return
	}

	var plan resource_model.Datastore
	plan.FromConfig(datastore)
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

var (
	_ resource.Resource                = &datastoreResource{}
	_ resource.ResourceWithConfigure   = &datastoreResource{}
	_ resource.ResourceWithImportState = &datastoreResource{}
)
