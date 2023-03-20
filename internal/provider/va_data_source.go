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
var _ datasource.DataSource = &VADataSource{}
var _ datasource.DataSourceWithConfigure = &VADataSource{}

func NewVADataSource() datasource.DataSource {
	return &VADataSource{}
}

type VADataSource struct {
	client *umbrella.Client
}

// VADataSourceModel describes the data source data model.
type VADataSourceModel struct {
	Vas []VAModel `tfsdk:"vas"`
}

type VAModel struct {
	ID             types.Int64     `tfsdk:"id"`
	OriginId       types.Int64     `tfsdk:"origin_id"`
	SiteId         types.Int64     `tfsdk:"site_id"`
	CreatedAt      types.String    `tfsdk:"created_at"`
	Health         types.String    `tfsdk:"health"`
	ModifiedAt     types.String    `tfsdk:"modified_at"`
	Name           types.String    `tfsdk:"name"`
	StateUpdatedAt types.String    `tfsdk:"state_updated_at"`
	Type           types.String    `tfsdk:"type"`
	IsUpgradable   types.Bool      `tfsdk:"is_upgradable"`
	Settings       VASettingsModel `tfsdk:"settings"`
	State          VAStateModel    `tfsdk:"state"`
	LastUpdated    types.String    `tfsdk:"last_updated"`
}

type VASettingsModel struct {
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

type VAStateModel struct {
	ConnectedToConnector       types.String `tfsdk:"connected_to_connector"`
	HasLocalDomainConfigured   types.String `tfsdk:"has_local_domain_configured"`
	QueryFailureRateAcceptable types.String `tfsdk:"query_failure_rate_acceptable"`
	ReceivedInternalDNSQueries types.String `tfsdk:"received_internal_dns_queries"`
	RedundantWithinSite        types.String `tfsdk:"redundant_within_site"`
	Syncing                    types.String `tfsdk:"syncing"`
}

func (d *VADataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_va"
}

func (d *VADataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "VA data source",
		Attributes: map[string]schema.Attribute{
			"vas": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"origin_id": schema.Int64Attribute{
							MarkdownDescription: "The origin ID of the Virtual Appliance",
							Computed:            true,
						},
						"site_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the Site",
							Computed:            true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"health": schema.StringAttribute{
							MarkdownDescription: "A description of the health of the virtual appliance",
							Computed:            true,
						},
						"is_upgradable": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether you can upgrade the Virtual Appliance (VA) to the latest VA version",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the Virtual Appliance",
							Computed:            true,
						},
						"modified_at": schema.StringAttribute{
							MarkdownDescription: "The date and time (ISO8601 timestamp) when the VA was modified",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "The date and time (ISO8601 timestamp) when the VA was created",
							Computed:            true,
						},
						"state_updated_at": schema.StringAttribute{
							MarkdownDescription: "The date and time (ISO8601 timestamp) when the state was updated",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the virtual appliance",
							Computed:            true,
						},
						"settings": schema.SingleNestedAttribute{
							Computed: true,
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
				},
			},
		},
	}
}

func (d *VADataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VADataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VADataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vas, err := d.client.GetVAs(&d.client.Token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Umbrella VAs",
			err.Error(),
		)
		return
	}

	for _, va := range vas {

		if va.Type == "virtual_appliance" {
			//Converting Go slice to TF ListType
			domains, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.Domains)
			internalips, _ := types.ListValueFrom(ctx, types.StringType, va.Settings.InternalIps)
			vaState := VAModel{
				ID:             types.Int64Value(int64(va.OriginId)),
				Name:           types.StringValue(va.Name),
				OriginId:       types.Int64Value(va.OriginId),
				IsUpgradable:   types.BoolValue(va.IsUpgradable),
				ModifiedAt:     types.StringValue(va.ModifiedAt),
				CreatedAt:      types.StringValue(va.CreatedAt),
				SiteId:         types.Int64Value(int64(va.SiteId)),
				Health:         types.StringValue(va.Health),
				StateUpdatedAt: types.StringValue(va.StateUpdatedAt),
				Type:           types.StringValue(va.Type),
				Settings: VASettingsModel{
					Uptime:            types.Int64Value(va.Settings.Uptime),
					ExternalIp:        types.StringValue(va.Settings.ExternalIp),
					HostType:          types.StringValue(va.Settings.HostType),
					LastSyncTime:      types.StringValue(va.Settings.LastSyncTime),
					UpgradeError:      types.StringValue(va.Settings.UpgradeError),
					Version:           types.StringValue(va.Settings.Version),
					IsDnscryptEnabled: types.BoolValue(va.Settings.IsDnscryptEnabled),
					Domains:           domains,
					InternalIps:       internalips,
				},
				State: VAStateModel{
					ConnectedToConnector:       types.StringValue(va.State.ConnectedToConnector),
					HasLocalDomainConfigured:   types.StringValue(va.State.HasLocalDomainConfigured),
					QueryFailureRateAcceptable: types.StringValue(va.State.QueryFailureRateAcceptable),
					ReceivedInternalDNSQueries: types.StringValue(va.State.ReceivedInternalDNSQueries),
					RedundantWithinSite:        types.StringValue(va.State.RedundantWithinSite),
					Syncing:                    types.StringValue(va.State.Syncing),
				},
			}
			data.Vas = append(data.Vas, vaState)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")
	fmt.Println(data)
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
