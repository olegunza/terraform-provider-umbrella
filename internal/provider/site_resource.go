package umbrellaprovider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SiteResource{}
var _ resource.ResourceWithImportState = &SiteResource{}

func NewSiteResource() resource.Resource {
	return &SiteResource{}
}

// ExampleResource defines the resource implementation.
type SiteResource struct {
	client *umbrella.Client
}

// ExampleResourceModel describes the resource data model.
type SiteResourceModel struct {
	SiteId      types.Int64  `tfsdk:"site_id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	OriginId    types.Int64  `tfsdk:"origin_id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Name        types.String `tfsdk:"name"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ID          types.Int64  `tfsdk:"id"`
}

func (r *SiteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (r *SiteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Site resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the Site",
				Computed:            true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"origin_id": schema.Int64Attribute{
				MarkdownDescription: "The origin ID of the Site",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the Site is default or not",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Site",
				Required:            true,
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the Site was modified",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the Site was created",
				Computed:            true,
			},
		},
	}
}

func (r *SiteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SiteResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	siteItem := umbrella.Site{
		Name: data.Name.ValueString(),
	}

	site, err := r.client.CreateSite(siteItem, &r.client.Token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Site",
			"Could not create Site, unexpected error: "+err.Error(),
		)
		return
	}

	data.SiteId = types.Int64Value(int64(site.Siteid))
	data.Name = types.StringValue(site.Name)
	data.OriginId = types.Int64Value(site.Originid)
	data.IsDefault = types.BoolValue(site.Isdefault)
	data.ModifiedAt = types.StringValue(site.Modifiedat)
	data.CreatedAt = types.StringValue(site.Createdat)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.ID = types.Int64Value(int64(site.Siteid))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SiteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	site, err := r.client.GetSite(data.SiteId.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella Site",
			"Could not read Umbrella Site ID "+strconv.FormatInt(data.SiteId.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	data.Name = types.StringValue(site.Name)
	data.OriginId = types.Int64Value(site.Originid)
	data.IsDefault = types.BoolValue(site.Isdefault)
	data.ModifiedAt = types.StringValue(site.Modifiedat)
	data.CreatedAt = types.StringValue(site.Createdat)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.ID = types.Int64Value(int64(site.Siteid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SiteResourceModel
	var statedata *SiteResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &statedata)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//siteid, _ := strconv.Atoi(data.SiteId.ValueString())
	siteItem := umbrella.Site{
		Name: data.Name.ValueString(),
	}

	_, err := r.client.UpdateSite(statedata.SiteId.ValueInt64(), siteItem, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Umbrella Site"+strconv.FormatInt(statedata.SiteId.ValueInt64(), 10),
			"Could not update order, unexpected error: "+err.Error(),
		)
		return
	}

	site, err := r.client.GetSite(statedata.SiteId.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella Site",
			"Could not read Umbrella Site ID "+strconv.FormatInt(statedata.SiteId.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	data.SiteId = types.Int64Value(int64(site.Siteid))
	data.Name = types.StringValue(site.Name)
	data.OriginId = types.Int64Value(site.Originid)
	data.IsDefault = types.BoolValue(site.Isdefault)
	data.ModifiedAt = types.StringValue(site.Modifiedat)
	data.CreatedAt = types.StringValue(site.Createdat)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.ID = types.Int64Value(int64(site.Siteid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *SiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	siteid, _ := strconv.ParseInt(req.ID, 10, 64)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteid)...)
	//resource.ImportStatePassthroughID(ctx, path.Root("site_id"), req, resp)
}
