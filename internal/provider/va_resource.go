package umbrellaprovider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SiteResource{}
var _ resource.ResourceWithImportState = &SiteResource{}

func NewVAResource() resource.Resource {
	return &VAResource{}
}

// VAResource defines the resource implementation.
type VAResource struct {
	client *umbrella.Client
}

// ExampleResourceModel describes the resource data model.
type VAResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	OriginId       types.Int64  `tfsdk:"origin_id"`
	SiteId         types.Int64  `tfsdk:"site_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Health         types.String `tfsdk:"health"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	Name           types.String `tfsdk:"name"`
	StateUpdatedAt types.String `tfsdk:"state_updated_at"`
	Type           types.String `tfsdk:"type"`
	IsUpgradable   types.Bool   `tfsdk:"is_upgradable"`
	Settings       types.Object `tfsdk:"settings"`
	State          types.Object `tfsdk:"state"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

type VAResourceSettingsModel struct {
	Uptime            types.Int64  `tfsdk:"uptime"`
	ExternalIp        types.String `tfsdk:"external_ip"`
	HostType          types.String `tfsdk:"host_type"`
	LastSyncTime      types.String `tfsdk:"last_sync_time"`
	UpgradeError      types.String `tfsdk:"upgrade_error"`
	Version           types.String `tfsdk:"version"`
	IsDnscryptEnabled types.Bool   `tfsdk:"is_dnscrypt_enabled"`
	Domains           types.List   `tfsdk:"domains"`
	InternalIps       types.List   `tfsdk:"internal_ips"`
}

type VAResourceStateModel struct {
	ConnectedToConnector       types.String `tfsdk:"connected_to_connector"`
	HasLocalDomainConfigured   types.String `tfsdk:"has_local_domain_configured"`
	QueryFailureRateAcceptable types.String `tfsdk:"query_failure_rate_acceptable"`
	ReceivedInternalDNSQueries types.String `tfsdk:"received_internal_dns_queries"`
	RedundantWithinSite        types.String `tfsdk:"redundant_within_site"`
	Syncing                    types.String `tfsdk:"syncing"`
}

func VASettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"uptime":              types.Int64Type,
		"external_ip":         types.StringType,
		"host_type":           types.StringType,
		"last_sync_time":      types.StringType,
		"upgrade_error":       types.StringType,
		"version":             types.StringType,
		"is_dnscrypt_enabled": types.BoolType,
		"domains":             types.ListType{ElemType: types.StringType},
		"internal_ips":        types.ListType{ElemType: types.StringType},
	}
}

func VAStateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"connected_to_connector":        types.StringType,
		"has_local_domain_configured":   types.StringType,
		"query_failure_rate_acceptable": types.StringType,
		"received_internal_dns_queries": types.StringType,
		"redundant_within_site":         types.StringType,
		"syncing":                       types.StringType,
	}
}

func (r *VAResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_va"
}

func (r *VAResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "VA resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"origin_id": schema.Int64Attribute{
				MarkdownDescription: "The origin ID of the Virtual Appliance",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the Site",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"health": schema.StringAttribute{
				MarkdownDescription: "A description of the health of the virtual appliance",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_upgradable": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether you can upgrade the Virtual Appliance (VA) to the latest VA version",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Virtual Appliance",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the VA was modified",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the VA was created",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state_updated_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the state was updated",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the virtual appliance",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"uptime": schema.Int64Attribute{
						Computed: true,
					},
					"external_ip": schema.StringAttribute{
						Computed: true,
					},
					"host_type": schema.StringAttribute{
						Computed: true,
					},
					"last_sync_time": schema.StringAttribute{
						Computed: true,
					},
					"upgrade_error": schema.StringAttribute{
						Computed: true,
					},
					"version": schema.StringAttribute{
						Computed: true,
					},
					"is_dnscrypt_enabled": schema.BoolAttribute{
						Computed: true,
					},
					"domains": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"internal_ips": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"state": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"connected_to_connector": schema.StringAttribute{
						Computed: true,
					},
					"has_local_domain_configured": schema.StringAttribute{
						Computed: true,
					},
					"query_failure_rate_acceptable": schema.StringAttribute{
						Computed: true,
					},
					"received_internal_dns_queries": schema.StringAttribute{
						Computed: true,
					},
					"redundant_within_site": schema.StringAttribute{
						Computed: true,
					},
					"syncing": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (r *VAResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*umbrella.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Virtual appliance creation is not supported by Umbrella API",
		"Consider to use terraform import to import existing virtual appliances",
	)
	return
}

func (r *VAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VAResourceModel
	tflog.Trace(ctx, "mumba")
	// Read Terraform prior state data into the model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Trace(ctx, "jumba")

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "tumba")

	va, err := r.client.GetVA(data.OriginId.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella Site",
			"Could not read Umbrella Site ID "+strconv.FormatInt(data.SiteId.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	data.Name = types.StringValue(va.Name)
	data.OriginId = types.Int64Value(va.OriginId)
	data.IsUpgradable = types.BoolValue(va.IsUpgradable)
	data.ModifiedAt = types.StringValue(va.ModifiedAt)
	data.CreatedAt = types.StringValue(va.CreatedAt)
	data.SiteId = types.Int64Value(va.SiteId)
	data.Health = types.StringValue(va.Health)
	data.StateUpdatedAt = types.StringValue(va.StateUpdatedAt)
	data.Type = types.StringValue(va.Type)

	var vasettings VAResourceSettingsModel

	vasettings.Uptime = types.Int64Value(va.Settings.Uptime)
	vasettings.ExternalIp = types.StringValue(va.Settings.ExternalIp)
	vasettings.HostType = types.StringValue(va.Settings.HostType)
	vasettings.LastSyncTime = types.StringValue(va.Settings.LastSyncTime)
	vasettings.UpgradeError = types.StringValue(va.Settings.UpgradeError)
	vasettings.Version = types.StringValue(va.Settings.Version)
	vasettings.IsDnscryptEnabled = types.BoolValue(va.Settings.IsDnscryptEnabled)

	domains, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.Domains)
	internalips, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.InternalIps)
	vasettings.Domains = domains
	vasettings.InternalIps = internalips

	data.Settings, _ = types.ObjectValueFrom(ctx, VASettingsAttrTypes(), vasettings)

	var vastate VAResourceStateModel

	vastate.ConnectedToConnector = types.StringValue(va.State.ConnectedToConnector)
	vastate.HasLocalDomainConfigured = types.StringValue(va.State.HasLocalDomainConfigured)
	vastate.QueryFailureRateAcceptable = types.StringValue(va.State.QueryFailureRateAcceptable)
	vastate.ReceivedInternalDNSQueries = types.StringValue(va.State.ReceivedInternalDNSQueries)
	vastate.RedundantWithinSite = types.StringValue(va.State.RedundantWithinSite)
	vastate.Syncing = types.StringValue(va.State.Syncing)

	data.State, _ = types.ObjectValueFrom(ctx, VAStateAttrTypes(), vastate)

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.ID = types.Int64Value(int64(va.OriginId))

	tflog.Trace(ctx, "Mapping ended")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VAResourceModel
	var statedata *VAResourceModel

	tflog.Trace(ctx, "Getting plan")

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	tflog.Trace(ctx, "Getting state")

	//statedata.Settings.Uptime = types.Int64Value(0)
	//statedata.State.ConnectedToConnector = types.StringValue("yes")

	resp.Diagnostics.Append(req.State.Get(ctx, &statedata)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Got state")
	//siteid, _ := strconv.Atoi(data.SiteId.ValueString())
	vaItem := umbrella.VA{
		SiteId: data.SiteId.ValueInt64(),
	}

	_, err := r.client.UpdateVA(statedata.OriginId.ValueInt64(), vaItem, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Umbrella virtual appliance"+strconv.FormatInt(statedata.OriginId.ValueInt64(), 10),
			"Could not update virtual appliance, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Updated")

	va, err := r.client.GetVA(statedata.OriginId.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella virtual appliance",
			"Could not read Umbrella VA ID "+strconv.FormatInt(statedata.OriginId.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Starting mapping")

	data.Name = types.StringValue(va.Name)
	data.OriginId = types.Int64Value(va.OriginId)
	data.IsUpgradable = types.BoolValue(va.IsUpgradable)
	data.ModifiedAt = types.StringValue(va.ModifiedAt)
	data.CreatedAt = types.StringValue(va.CreatedAt)
	data.SiteId = types.Int64Value(va.SiteId)
	data.Health = types.StringValue(va.Health)
	data.StateUpdatedAt = types.StringValue(va.StateUpdatedAt)
	data.Type = types.StringValue(va.Type)

	var vasettings VAResourceSettingsModel

	vasettings.Uptime = types.Int64Value(va.Settings.Uptime)
	vasettings.ExternalIp = types.StringValue(va.Settings.ExternalIp)
	vasettings.HostType = types.StringValue(va.Settings.HostType)
	vasettings.LastSyncTime = types.StringValue(va.Settings.LastSyncTime)
	vasettings.UpgradeError = types.StringValue(va.Settings.UpgradeError)
	vasettings.Version = types.StringValue(va.Settings.Version)
	vasettings.IsDnscryptEnabled = types.BoolValue(va.Settings.IsDnscryptEnabled)

	domains, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.Domains)
	internalips, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.InternalIps)
	vasettings.Domains = domains
	vasettings.InternalIps = internalips

	data.Settings, _ = types.ObjectValueFrom(ctx, VASettingsAttrTypes(), vasettings)

	var vastate VAResourceStateModel

	vastate.ConnectedToConnector = types.StringValue(va.State.ConnectedToConnector)
	vastate.HasLocalDomainConfigured = types.StringValue(va.State.HasLocalDomainConfigured)
	vastate.QueryFailureRateAcceptable = types.StringValue(va.State.QueryFailureRateAcceptable)
	vastate.ReceivedInternalDNSQueries = types.StringValue(va.State.ReceivedInternalDNSQueries)
	vastate.RedundantWithinSite = types.StringValue(va.State.RedundantWithinSite)
	vastate.Syncing = types.StringValue(va.State.Syncing)

	data.State, _ = types.ObjectValueFrom(ctx, VAStateAttrTypes(), vastate)

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.ID = types.Int64Value(int64(va.OriginId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *VAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SiteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSite(data.SiteId.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Umbrella Site",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *VAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	originid, _ := strconv.ParseInt(req.ID, 10, 64)

	// What is path https://developer.hashicorp.com/terraform/plugin/framework/handling-data/paths

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("origin_id"), originid)...)
	//resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("settings").AtName("uptime"), 0)...)
	//resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("state").AtName("connected_to_connector"), "yes")...)
}
