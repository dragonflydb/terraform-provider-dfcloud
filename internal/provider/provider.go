package provider

import (
	"context"

	dfcloud "github.com/dragonflydb/dfcloud/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DragonflyDBCloudProvider struct {
	version string
}

type ProviderSchema struct {
	ApiKey  types.String `tfsdk:"api_key"`
	ApiHost types.String `tfsdk:"api_host"`
}

func NewDragonflyDBCloudProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DragonflyDBCloudProvider{
			version: version,
		}
	}
}

func (p DragonflyDBCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dfcloud_"
	resp.Version = p.version
}

func (p DragonflyDBCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Dragonfly Cloud API key. This can also be set via the DFCLOUD_API_KEY environment variable.",
			},
			"api_host": schema.StringAttribute{
				Optional:    true,
				Description: "The URL of the Dragonfly Cloud API.",
			},
		},
		Description: `The Dragonfly Cloud provider is used to interact with resources supported by Dragonfly Cloud.

The provider needs to be configured with the proper credentials before it can be used.`,
	}
}

func (p DragonflyDBCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderSchema
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var options []dfcloud.ClientOption
	if config.ApiKey.ValueString() != "" {
		options = append(options, dfcloud.WithAPIKey(config.ApiKey.ValueString()))

	} else {
		options = append(options, dfcloud.WithAPIKeyFromEnv())
	}

	if config.ApiHost.ValueString() != "" {
		options = append(options, dfcloud.WithAPIHost(config.ApiHost.ValueString()))
	}

	client, err := dfcloud.NewClient(options...)
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	resp.ResourceData = client
}

func (p DragonflyDBCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Provider specific implementation
	}
}

func (p DragonflyDBCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatastoreResource,
		NewNetworkResource,
		NewConnectionResource,
	}
}

var _ provider.Provider = &DragonflyDBCloudProvider{}
