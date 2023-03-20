package umbrellaprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TunnelDataSource{}
var _ datasource.DataSourceWithConfigure = &TunnelDataSource{}

func NewTunnelDataSource() datasource.DataSource {
	return &TunnelDataSource{}
}

type TunnelDataSource struct {
	client *umbrella.Client
}
type TunnelDataSourceModel struct {
	tunnels []VAModel `tfsdk:"vas"`
}
type TunnelDataSourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Uri          types.String `tfsdk:"uri"`
	Name         types.String `tfsdk:"name"`
	SiteOriginId types.Int64  `tfsdk:"site_origin_id"`
	Client       types.Object `tfsdk:"client"`
	Transport    types.Object `tfsdk:"transport"`
	ServiceType  types.String `tfsdk:"service_type"`
	NetworkCidrs types.List   `tfsdk:"network_cidrs"`
	//Meta         *TunnelMetaResourceModel   `tfsdk:"meta"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

type TunnelClientDataSourceModel struct {
	DeviceType     types.String `tfsdk:"device_type"`
	Authentication types.Object `tfsdk:"authentication"`
}

type TunnelTransDataSourceModel struct {
	Protocol types.String `tfsdk:"protocol"`
}

type TunnelAuthDataSourceModel struct {
	Type       types.String `tfsdk:"type"`
	Parameters types.Object `tfsdk:"parameters"`
}

type TunnelAuthParamsDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	Secret     types.String `tfsdk:"secret"`
	IdPrefix   types.String `tfsdk:"id_prefix"`
}

func (d *TunnelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tunnel"
}

func (d *TunnelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "DC list data source",
		Attributes: map[string]schema.Attribute{
			"continents": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: continentsResourceAttr(),
				},
			},
		},
	}
}

func citiesResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"latitude": schema.StringAttribute{
			Computed: true,
		},
		"longitude": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"dc": schema.StringAttribute{
			Computed: true,
		},
		"range": schema.StringAttribute{
			Computed: true,
		},
		"fqdn": schema.StringAttribute{
			Computed: true,
		},
	}
}

func continentsResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed: true,
		},
		"cities": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: citiesResourceAttr(),
			},
		},
	}
}

func (d *TunnelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func typeFromAttrs(in map[string]schema.Attribute) map[string]attr.Type {
	out := map[string]attr.Type{}
	for k, v := range in {
		out[k] = v.GetType()
	}
	return out
}

func (d *TunnelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DCListDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dclist, err := d.client.GetDCs(&d.client.Token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Umbrella Dc List",
			err.Error(),
		)
		return
	}

	var dccitieslist []City
	var dccontlist []Continent

	for _, continent := range dclist.Continents {
		for _, city := range continent.Cities {
			dccity := City{
				Latitude:  types.StringValue(city.Latitude),
				Longitude: types.StringValue(city.Longitude),
				Name:      types.StringValue(city.Name),
				Dc:        types.StringValue(city.Dc),
				Range:     types.StringValue(city.Range),
				Fqdn:      types.StringValue(city.Fqdn),
			}

			dccitieslist = append(dccitieslist, dccity)
		}
		cities, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(citiesResourceAttr())}, dccitieslist)
		dccitieslist = nil
		dccontinent := Continent{
			Name:   types.StringValue(continent.Name),
			Cities: cities,
		}

		dccontlist = append(dccontlist, dccontinent)
	}
	continents, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(continentsResourceAttr())}, dccontlist)
	data.Continents = continents

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
