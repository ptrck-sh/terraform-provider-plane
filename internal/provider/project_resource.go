package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.com/ptrck-sh/terraform-provider-plane/internal/client"
)

var (
	_ resource.Resource                = (*projectResource)(nil)
	_ resource.ResourceWithConfigure   = (*projectResource)(nil)
	_ resource.ResourceWithImportState = (*projectResource)(nil)
)

// NewProjectResource is the plane_project resource factory.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *client.Client
}

type projectResourceModel struct {
	ID              types.String `tfsdk:"id"`
	WorkspaceSlug   types.String `tfsdk:"workspace_slug"`
	Name            types.String `tfsdk:"name"`
	Identifier      types.String `tfsdk:"identifier"`
	Description     types.String `tfsdk:"description"`
	ProjectLead     types.String `tfsdk:"project_lead"`
	DefaultAssignee types.String `tfsdk:"default_assignee"`
	ModuleView      types.Bool   `tfsdk:"module_view"`
	CycleView       types.Bool   `tfsdk:"cycle_view"`
	IssueViewsView  types.Bool   `tfsdk:"issue_views_view"`
	PageView        types.Bool   `tfsdk:"page_view"`
	IntakeView      types.Bool   `tfsdk:"intake_view"`
	ArchiveIn       types.Int64  `tfsdk:"archive_in"`
	CloseIn         types.Int64  `tfsdk:"close_in"`
	Timezone        types.String `tfsdk:"timezone"`
	Workspace       types.String `tfsdk:"workspace"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// These mirror server-side defaults. Omitting one in config must leave the
	// remote value untouched, so they keep prior state instead of planning as
	// unknown on every refresh.
	computedBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{
			MarkdownDescription: desc,
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A Plane project within a workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project UUID, assigned by Plane.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace_slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the workspace that owns the project. Changing this forces a new project.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name (max 255 characters).",
				Required:            true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Short project identifier used as the work-item key prefix (max 12 characters). " +
					"Changing it forces a new project: Plane allows the field in PATCH, but altering it in place can " +
					"break existing work-item references, so this provider treats it as immutable until verified safe.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description. Omit to leave whatever Plane already has; set to \"\" to clear it.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_lead": schema.StringAttribute{
				MarkdownDescription: "User UUID of the project lead.",
				Optional:            true,
			},
			"default_assignee": schema.StringAttribute{
				MarkdownDescription: "User UUID of the default work-item assignee.",
				Optional:            true,
			},
			"module_view":      computedBool("Whether the Modules feature is enabled."),
			"cycle_view":       computedBool("Whether the Cycles feature is enabled."),
			"issue_views_view": computedBool("Whether the Views feature is enabled."),
			"page_view":        computedBool("Whether the Pages feature is enabled."),
			"intake_view":      computedBool("Whether the Intake feature is enabled."),
			"archive_in": schema.Int64Attribute{
				MarkdownDescription: "Months of inactivity after which work items auto-archive (0 disables).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"close_in": schema.Int64Attribute{
				MarkdownDescription: "Months of inactivity after which work items auto-close (0 disables).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "Project timezone (IANA name).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "UUID of the owning workspace.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last-update timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = c
}

// toAPI builds an API project payload from the plan model.
func (m projectResourceModel) toAPI() client.Project {
	p := client.Project{
		Name:       m.Name.ValueString(),
		Identifier: m.Identifier.ValueString(),
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	if !m.ProjectLead.IsNull() && !m.ProjectLead.IsUnknown() {
		v := m.ProjectLead.ValueString()
		p.ProjectLead = &v
	}
	if !m.DefaultAssignee.IsNull() && !m.DefaultAssignee.IsUnknown() {
		v := m.DefaultAssignee.ValueString()
		p.DefaultAssignee = &v
	}
	if !m.ModuleView.IsNull() && !m.ModuleView.IsUnknown() {
		v := m.ModuleView.ValueBool()
		p.ModuleView = &v
	}
	if !m.CycleView.IsNull() && !m.CycleView.IsUnknown() {
		v := m.CycleView.ValueBool()
		p.CycleView = &v
	}
	if !m.IssueViewsView.IsNull() && !m.IssueViewsView.IsUnknown() {
		v := m.IssueViewsView.ValueBool()
		p.IssueViewsView = &v
	}
	if !m.PageView.IsNull() && !m.PageView.IsUnknown() {
		v := m.PageView.ValueBool()
		p.PageView = &v
	}
	if !m.IntakeView.IsNull() && !m.IntakeView.IsUnknown() {
		v := m.IntakeView.ValueBool()
		p.IntakeView = &v
	}
	if !m.ArchiveIn.IsNull() && !m.ArchiveIn.IsUnknown() {
		v := m.ArchiveIn.ValueInt64()
		p.ArchiveIn = &v
	}
	if !m.CloseIn.IsNull() && !m.CloseIn.IsUnknown() {
		v := m.CloseIn.ValueInt64()
		p.CloseIn = &v
	}
	if !m.Timezone.IsNull() && !m.Timezone.IsUnknown() {
		p.Timezone = m.Timezone.ValueString()
	}
	return p
}

// applyAPI copies an API project response into the model, preserving the
// workspace_slug (a path param the API does not echo as a slug).
func (m *projectResourceModel) applyAPI(p *client.Project) {
	m.ID = types.StringValue(p.ID)
	m.Name = types.StringValue(p.Name)
	m.Identifier = types.StringValue(p.Identifier)
	m.Description = stringPtrToValue(p.Description)
	m.ProjectLead = stringPtrToValue(p.ProjectLead)
	m.DefaultAssignee = stringPtrToValue(p.DefaultAssignee)
	m.ModuleView = boolPtrToValue(p.ModuleView)
	m.CycleView = boolPtrToValue(p.CycleView)
	m.IssueViewsView = boolPtrToValue(p.IssueViewsView)
	m.PageView = boolPtrToValue(p.PageView)
	m.IntakeView = boolPtrToValue(p.IntakeView)
	m.ArchiveIn = int64PtrToValue(p.ArchiveIn)
	m.CloseIn = int64PtrToValue(p.CloseIn)
	m.Timezone = types.StringValue(p.Timezone)
	m.Workspace = types.StringValue(p.Workspace)
	m.CreatedAt = types.StringValue(p.CreatedAt)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateProject(ctx, plan.WorkspaceSlug.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}
	plan.applyAPI(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetProject(ctx, state.WorkspaceSlug.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}
	state.applyAPI(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateProject(ctx, plan.WorkspaceSlug.ValueString(), plan.ID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}
	plan.applyAPI(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProject(ctx, state.WorkspaceSlug.ValueString(), state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting project", err.Error())
	}
}

// ImportState imports by "workspace_slug/project_id".
func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected \"workspace_slug/project_id\", got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
