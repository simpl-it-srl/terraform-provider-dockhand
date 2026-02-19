package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*gitStackResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitStackResource)(nil)
	_ resource.ResourceWithImportState = (*gitStackResource)(nil)
)

func NewGitStackResource() resource.Resource {
	return &gitStackResource{}
}

type gitStackResource struct {
	client *Client
}

type gitStackModel struct {
	ID             types.String `tfsdk:"id"`
	StackName      types.String `tfsdk:"stack_name"`
	ComposePath    types.String `tfsdk:"compose_path"`
	EnvFilePath    types.String `tfsdk:"env_file_path"`
	RepositoryID   types.String `tfsdk:"repository_id"`
	EnvironmentID  types.String `tfsdk:"environment_id"`
	AutoUpdate     types.Bool   `tfsdk:"auto_update"`
	AutoUpdateCron types.String `tfsdk:"auto_update_cron"`
	WebhookEnabled types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret  types.String `tfsdk:"webhook_secret"`
}

func (r *gitStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_stack"
}

func (r *gitStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand Git stack via `/api/git/stacks`. A Git stack links a specific compose file in a repository to a named stack.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_name": schema.StringAttribute{
				Required: true,
			},
			"compose_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to the Docker Compose file within the repository (e.g. `myapp_stack/docker-compose.yaml`).",
			},
			"repository_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"env_file_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_update": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"auto_update_cron": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"webhook_secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *gitStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *gitStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git stack configuration", err.Error())
		return
	}

	created, _, err := r.client.CreateGitStack(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand git stack", err.Error())
		return
	}

	state := modelFromGitStackResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gitStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitStackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stack, status, err := r.client.GetGitStack(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand git stack", err.Error())
		return
	}

	newState := modelFromGitStackResponse(stack)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackModel
	var state gitStackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git stack configuration", err.Error())
		return
	}

	updated, _, err := r.client.UpdateGitStack(ctx, state.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand git stack", err.Error())
		return
	}

	newState := modelFromGitStackResponse(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitStackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteGitStack(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand git stack", err.Error())
		return
	}
}

func (r *gitStackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildGitStackPayload(plan gitStackModel) (gitStackPayload, error) {
	repoIDStr := plan.RepositoryID.ValueString()
	if repoIDStr == "" {
		return gitStackPayload{}, fmt.Errorf("repository_id is required")
	}
	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		return gitStackPayload{}, fmt.Errorf("repository_id must be a numeric string: %w", err)
	}

	payload := gitStackPayload{
		StackName:    plan.StackName.ValueString(),
		ComposePath:  plan.ComposePath.ValueString(),
		AutoUpdate:   plan.AutoUpdate.ValueBool(),
		WebhookEnabled: plan.WebhookEnabled.ValueBool(),
		DeployNow:    false,
		EnvVars:      []any{},
		RepositoryID: repoID,
	}

	if !plan.EnvFilePath.IsNull() && !plan.EnvFilePath.IsUnknown() {
		v := plan.EnvFilePath.ValueString()
		payload.EnvFilePath = &v
	}
	if !plan.EnvironmentID.IsNull() && !plan.EnvironmentID.IsUnknown() {
		v, err := strconv.ParseInt(plan.EnvironmentID.ValueString(), 10, 64)
		if err != nil {
			return gitStackPayload{}, fmt.Errorf("environment_id must be a numeric string: %w", err)
		}
		payload.EnvironmentID = &v
	}
	if !plan.AutoUpdateCron.IsNull() && !plan.AutoUpdateCron.IsUnknown() {
		v := plan.AutoUpdateCron.ValueString()
		payload.AutoUpdateCron = &v
	}

	return payload, nil
}

func modelFromGitStackResponse(in *gitStackResponse) gitStackModel {
	out := gitStackModel{
		ID:             types.StringValue(fmt.Sprintf("%d", in.ID)),
		StackName:      types.StringValue(in.StackName),
		ComposePath:    types.StringValue(in.ComposePath),
		RepositoryID:   types.StringValue(fmt.Sprintf("%d", in.RepositoryID)),
		AutoUpdate:     types.BoolValue(in.AutoUpdate),
		WebhookEnabled: types.BoolValue(in.WebhookEnabled),
	}

	if in.EnvFilePath != nil {
		out.EnvFilePath = types.StringValue(*in.EnvFilePath)
	} else {
		out.EnvFilePath = types.StringNull()
	}
	if in.EnvironmentID != nil {
		out.EnvironmentID = types.StringValue(fmt.Sprintf("%d", *in.EnvironmentID))
	} else {
		out.EnvironmentID = types.StringNull()
	}
	if in.AutoUpdateCron != nil {
		out.AutoUpdateCron = types.StringValue(*in.AutoUpdateCron)
	} else {
		out.AutoUpdateCron = types.StringNull()
	}
	if in.WebhookSecret != nil {
		out.WebhookSecret = types.StringValue(*in.WebhookSecret)
	} else {
		out.WebhookSecret = types.StringNull()
	}

	return out
}
