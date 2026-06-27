package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/client"
)

var (
	_ datasource.DataSource              = (*workspaceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspaceDataSource)(nil)
)

// NewWorkspaceDataSource is the plane_workspace data source factory.
func NewWorkspaceDataSource() datasource.DataSource {
	return &workspaceDataSource{}
}

type workspaceDataSource struct {
	client *client.Client
}

type workspaceDataSourceModel struct {
	Slug types.String `tfsdk:"slug"`
	ID   types.String `tfsdk:"id"`
}

func (d *workspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *workspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resolve a Plane workspace by slug.\n\n" +
			"The self-hosted v1 API exposes no workspace endpoint, so this data source " +
			"is intentionally thin: it validates the slug is reachable and derives the " +
			"workspace UUID from the workspace's projects. `id` is null when the workspace " +
			"has no projects.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				MarkdownDescription: "Workspace slug.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Workspace UUID, derived from the workspace's projects. " +
					"Null when the workspace has no projects.",
				Computed: true,
			},
		},
	}
}

func (d *workspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *workspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg workspaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projects, err := d.client.ListProjects(ctx, cfg.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	cfg.ID = types.StringNull()
	if len(projects) > 0 && projects[0].Workspace != "" {
		cfg.ID = types.StringValue(projects[0].Workspace)
	} else {
		resp.Diagnostics.AddWarning(
			"Workspace UUID unavailable",
			"The workspace has no projects, so its UUID could not be derived. `id` is null.",
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
