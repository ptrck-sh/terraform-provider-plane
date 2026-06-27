// Package provider implements the Plane Terraform/OpenTofu provider using
// terraform-plugin-framework.
package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/client"
)

var _ provider.Provider = (*planeProvider)(nil)

type planeProvider struct {
	version string
}

// New returns a provider factory for the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &planeProvider{version: version}
	}
}

type providerModel struct {
	Host   types.String `tfsdk:"host"`
	APIKey types.String `tfsdk:"api_key"`
}

func (p *planeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "plane"
	resp.Version = p.version
}

func (p *planeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a self-hosted [Plane](https://plane.so) instance via its REST API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Plane instance, e.g. `https://plane.example.com`. " +
					"May also be set with the `PLANE_HOST` environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Plane personal access token, sent in the `X-API-Key` header. " +
					"May also be set with the `PLANE_API_KEY` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *planeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("PLANE_HOST")
	if !cfg.Host.IsNull() {
		host = cfg.Host.ValueString()
	}
	apiKey := os.Getenv("PLANE_API_KEY")
	if !cfg.APIKey.IsNull() {
		apiKey = cfg.APIKey.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Plane host",
			"Set the provider `host` attribute or the PLANE_HOST environment variable.",
		)
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Plane API key",
			"Set the provider `api_key` attribute or the PLANE_API_KEY environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(host, apiKey, http.DefaultClient)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *planeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewStateResource,
		NewLabelResource,
	}
}

func (p *planeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewWorkspaceDataSource,
	}
}
