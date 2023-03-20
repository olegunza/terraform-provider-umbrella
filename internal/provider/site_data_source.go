package umbrellaprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SiteDataSource{}
var _ datasource.DataSourceWithConfigure = &SiteDataSource{}

func NewSiteDataSource() datasource.DataSource {
	return &SiteDataSource{}
}

type SiteDataSource struct {
	client *umbrella.Client
}

// ExampleDataSourceModel describes the data source data model.
type SiteDataSourceModel struct {
	Sites []SitesModel `tfsdk:"sites"`
}

type SitesModel struct {
	SiteId      types.Int64  `tfsdk:"site_id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	OriginId    types.Int64  `tfsdk:"origin_id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Name        types.String `tfsdk:"name"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ID          types.Int64  `tfsdk:"id"`
}

func (d *SiteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (d *SiteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Site data source",
		Attributes: map[string]schema.Attribute{
			"sites": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
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
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *SiteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*umbrella.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SiteDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sites, err := d.client.GetSites(&d.client.Token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Umbrella Sites",
			err.Error(),
		)
		return
	}

	for _, site := range sites {
		siteState := SitesModel{
			ID:         types.Int64Value(int64(site.Siteid)),
			Name:       types.StringValue(site.Name),
			OriginId:   types.Int64Value(site.Originid),
			IsDefault:  types.BoolValue(site.Isdefault),
			ModifiedAt: types.StringValue(site.Modifiedat),
			CreatedAt:  types.StringValue(site.Createdat),
			SiteId:     types.Int64Value(int64(site.Siteid)),
		}
		data.Sites = append(data.Sites, siteState)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
