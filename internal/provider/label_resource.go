package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/client"
)

var (
	_ resource.Resource                = (*labelResource)(nil)
	_ resource.ResourceWithConfigure   = (*labelResource)(nil)
	_ resource.ResourceWithImportState = (*labelResource)(nil)
)

func NewLabelResource() resource.Resource { return &labelResource{} }

type labelResource struct{ client *client.Client }

type labelResourceModel struct {
	ID            types.String `tfsdk:"id"`
	WorkspaceSlug types.String `tfsdk:"workspace_slug"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Color         types.String `tfsdk:"color"`
	Parent        types.String `tfsdk:"parent"`
	Project       types.String `tfsdk:"project"`
	Workspace     types.String `tfsdk:"workspace"`
}

func (r *labelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *labelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A label within a Plane project.",
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
			"name": schema.StringAttribute{Required: true},
			"color": schema.StringAttribute{
				MarkdownDescription: "Hex color string for the label. Assigned by Plane when omitted.",
				Optional:            true,
				Computed:            true,
			},
			"parent": schema.StringAttribute{
				MarkdownDescription: "UUID of a parent label (for sub-labels).",
				Optional:            true,
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

func (r *labelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m labelResourceModel) toAPI() client.Label {
	l := client.Label{
		Name:  m.Name.ValueString(),
		Color: m.Color.ValueString(),
	}
	if !m.Parent.IsNull() && !m.Parent.IsUnknown() {
		v := m.Parent.ValueString()
		l.Parent = &v
	}
	return l
}

func (m *labelResourceModel) applyAPI(l *client.Label) {
	m.ID = types.StringValue(l.ID)
	m.Name = types.StringValue(l.Name)
	m.Color = types.StringValue(l.Color)
	m.Parent = stringPtrToValue(l.Parent)
	m.Project = types.StringValue(l.Project)
	m.Workspace = types.StringValue(l.Workspace)
}

func (r *labelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan labelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateLabel(ctx, plan.WorkspaceSlug.ValueString(), plan.ProjectID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error creating label", err.Error())
		return
	}
	plan.applyAPI(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *labelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state labelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetLabel(ctx, state.WorkspaceSlug.ValueString(), state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading label", err.Error())
		return
	}
	state.applyAPI(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *labelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan labelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateLabel(ctx, plan.WorkspaceSlug.ValueString(), plan.ProjectID.ValueString(), plan.ID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating label", err.Error())
		return
	}
	plan.applyAPI(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *labelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state labelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteLabel(ctx, state.WorkspaceSlug.ValueString(), state.ProjectID.ValueString(), state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting label", err.Error())
	}
}

// ImportState imports by "workspace_slug/project_id/label_id".
func (r *labelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected \"workspace_slug/project_id/label_id\", got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}
