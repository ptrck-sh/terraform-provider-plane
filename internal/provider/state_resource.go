package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/client"
)

var (
	_ resource.Resource                = (*stateResource)(nil)
	_ resource.ResourceWithConfigure   = (*stateResource)(nil)
	_ resource.ResourceWithImportState = (*stateResource)(nil)
)

func NewStateResource() resource.Resource { return &stateResource{} }

type stateResource struct{ client *client.Client }

type stateResourceModel struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Color         types.String `tfsdk:"color"`
	Group         types.String `tfsdk:"group"`
	Default       types.Bool   `tfsdk:"default"`
	Project       types.String `tfsdk:"project"`
	Workspace     types.String `tfsdk:"workspace"`
}

func (r *stateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_state"
}

func (r *stateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A workflow state within a Plane project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_slug": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":  schema.StringAttribute{Required: true},
			"color": schema.StringAttribute{Required: true},
			"group": schema.StringAttribute{
				MarkdownDescription: "State group: `backlog`, `unstarted`, `started`, `completed`, `cancelled`, or `triage`. Defaults to `backlog` when omitted.",
				Optional:            true,
				Computed:            true,
			},
			"default": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"project": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *stateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData))
		return
	}
	r.client = c
}

func (m stateResourceModel) toAPI() client.State {
	return client.State{
		Name:    m.Name.ValueString(),
		Color:   m.Color.ValueString(),
		Group:   m.Group.ValueString(),
		Default: m.Default.ValueBool(),
	}
}

func (m *stateResourceModel) applyAPI(s *client.State) {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.Color = types.StringValue(s.Color)
	m.Group = types.StringValue(s.Group)
	m.Default = types.BoolValue(s.Default)
	m.Project = types.StringValue(s.Project)
	m.Workspace = types.StringValue(s.Workspace)
}

func (r *stateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateState(ctx, plan.WorkspaceSlug.ValueString(), plan.ProjectID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error creating state", err.Error())
		return
	}
	plan.applyAPI(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetState(ctx, state.WorkspaceSlug.ValueString(), state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading state", err.Error())
		return
	}
	state.applyAPI(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateState(ctx, plan.WorkspaceSlug.ValueString(), plan.ProjectID.ValueString(), plan.ID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating state", err.Error())
		return
	}
	plan.applyAPI(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteState(ctx, state.WorkspaceSlug.ValueString(), state.ProjectID.ValueString(), state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting state", err.Error())
	}
}

// ImportState imports by "workspace_slug/project_id/state_id".
func (r *stateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected \"workspace_slug/project_id/state_id\", got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}
