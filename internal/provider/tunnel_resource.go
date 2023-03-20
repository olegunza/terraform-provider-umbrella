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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TunnelResource{}
var _ resource.ResourceWithImportState = &TunnelResource{}

func NewTunnelResource() resource.Resource {
	return &TunnelResource{}
}

// ExampleResource defines the resource implementation.
type TunnelResource struct {
	client *umbrella.Client
}

// ExampleResourceModel describes the resource data model.
type TunnelResourceModel struct {
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

type TunnelClientResourceModel struct {
	DeviceType     types.String `tfsdk:"device_type"`
	Authentication types.Object `tfsdk:"authentication"`
}

func (o TunnelClientResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"device_type":    types.StringType,
		"authentication": types.ObjectType{AttrTypes: AuthAttrTypes()},
	}
}

type TunnelTransResourceModel struct {
	Protocol types.String `tfsdk:"protocol"`
}

func (o TunnelTransResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"protocol": types.StringType,
	}
}

type TunnelMetaResourceModel struct {
}

type TunnelAuthResourceModel struct {
	Type       types.String `tfsdk:"type"`
	Parameters types.Object `tfsdk:"parameters"`
}

func AuthAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":       types.StringType,
		"parameters": types.ObjectType{AttrTypes: ParamsAttrTypes()},
	}
}

type TunnelAuthParamsResourceModel struct {
	Id         types.String `tfsdk:"id"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	Secret     types.String `tfsdk:"secret"`
	IdPrefix   types.String `tfsdk:"id_prefix"`
}

func ParamsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"modified_at": types.StringType,
		"secret":      types.StringType,
		"id_prefix":   types.StringType,
	}
}
func (r *TunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tunnel"
}

func (r *TunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Tunnel resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the Site",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The Uri of the Tunnel",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Tunnel",
				Required:            true,
			},
			"site_origin_id": schema.Int64Attribute{
				MarkdownDescription: "The origin ID of the Site",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,

				Attributes: map[string]schema.Attribute{
					"device_type": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"authentication": schema.SingleNestedAttribute{
						Computed: true,
						Optional: true,

						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed: true,
								Optional: true,
							},
							"parameters": schema.SingleNestedAttribute{
								Computed: true,
								Optional: true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},

								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed: true,
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"modified_at": schema.StringAttribute{
										Computed: true,
									},
									"secret": schema.StringAttribute{
										Computed: true,
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"id_prefix": schema.StringAttribute{
										Computed: true,
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
			"transport": schema.SingleNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},

				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},

			"service_type": schema.StringAttribute{
				MarkdownDescription: "Specifies whether the Site is default or not",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"modified_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the Site was modified",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time (ISO8601 timestamp) when the Site was created",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_cidrs": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			//"meta": schema.SingleNestedAttribute{
			//	Computed:   true,
			//	Optional:   true,
			//	Attributes: map[string]schema.Attribute{},
			//},
		},
	}
}
func buildTunnelItem(data TunnelResourceModel, client TunnelClientResourceModel, auth TunnelAuthResourceModel, parameters TunnelAuthParamsResourceModel, transport TunnelTransResourceModel, networkcidrs []string) umbrella.NetworkTunnel {

	tunnelItem := umbrella.NetworkTunnel{Name: data.Name.ValueString(),
		Client: umbrella.TunnelClient{
			Authentication: umbrella.TunnelAuth{
				Type: auth.Type.ValueString(),
				Parameters: umbrella.TunnelAuthParams{
					Id: parameters.IdPrefix.ValueString(),
				},
			},
		},
		ServiceType:  data.ServiceType.ValueString(),
		NetworkCIDRs: networkcidrs,
	}

	if !client.DeviceType.IsNull() && !client.DeviceType.IsUnknown() {
		tunnelItem.Client.DeviceType = client.DeviceType.ValueString()

	}
	if !parameters.Secret.IsNull() && !parameters.Secret.IsUnknown() && !parameters.Secret.Equal(types.StringValue("")) {
		//tunnelItem.Client.Authentication.Parameters.Secret = parameters.Secret.ValueString()

	}
	if !data.SiteOriginId.IsNull() && !data.SiteOriginId.IsUnknown() && !data.SiteOriginId.Equal(types.Int64Value(0)) {
		tunnelItem.SiteOriginId = data.SiteOriginId.ValueInt64()

	}

	if !transport.Protocol.IsNull() && !transport.Protocol.IsUnknown() && !transport.Protocol.Equal(types.StringValue("")) {
		tunnelItem.Transport.Protocol = transport.Protocol.ValueString()

	}
	return tunnelItem
}

func (r *TunnelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TunnelResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//https://discuss.hashicorp.com/t/plan-mylist-elementsas-vs-elements/46455
	var networkcidrs []string

	if !data.NetworkCidrs.IsNull() && !data.NetworkCidrs.IsUnknown() {
		resp.Diagnostics.Append(data.NetworkCidrs.ElementsAs(ctx, &networkcidrs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//Doing conversion of ObjectType from parent TunnelResourceModel Transport to Go struct TunnelTransResourceModel to handle Unknown and Null values properly
	//https://discuss.hashicorp.com/t/thoughts-on-framework-v0-15-0/46277
	//https://discuss.hashicorp.com/t/nested-types-object-and-redundant-attrtypes/45927
	var transport TunnelTransResourceModel
	resp.Diagnostics.Append(data.Transport.As(ctx, &transport, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client TunnelClientResourceModel
	resp.Diagnostics.Append(data.Client.As(ctx, &client, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var auth TunnelAuthResourceModel
	resp.Diagnostics.Append(client.Authentication.As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var parameters TunnelAuthParamsResourceModel
	resp.Diagnostics.Append(auth.Parameters.As(ctx, &parameters, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	if resp.Diagnostics.HasError() {
		return
	}

	tunnelItem := umbrella.NetworkTunnel{
		Name:         data.Name.ValueString(),
		SiteOriginId: data.SiteOriginId.ValueInt64(),
		Client: umbrella.TunnelClient{
			DeviceType: client.DeviceType.ValueString(),
			Authentication: umbrella.TunnelAuth{
				Type: auth.Type.ValueString(),
				Parameters: umbrella.TunnelAuthParams{
					Id:     parameters.IdPrefix.ValueString(),
					Secret: parameters.Secret.ValueString(),
				},
			},
		},
		ServiceType:  data.ServiceType.ValueString(),
		NetworkCIDRs: networkcidrs,
		Transport: umbrella.TunnelTrans{
			Protocol: transport.Protocol.ValueString(),
		},
	}
	if !data.NetworkCidrs.IsNull() && !data.NetworkCidrs.IsUnknown() {
		tunnelItem.NetworkCIDRs = networkcidrs
	}

	tunnel, err := r.client.CreateTunnel(tunnelItem, &r.client.Token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Tunnel",
			"Could not create Tunnel, unexpected error: "+err.Error(),
		)
		return
	}

	cidrs, _ := types.ListValueFrom(ctx, types.StringType, tunnel.NetworkCIDRs)
	//idprefix := strings.Split(tunnel.Client.Authentication.Parameters.Id, "@")[0]

	//Doing conversion from API NetworkTunnel struct to ObjectType TunnelTransResourceModel which we then assign to TunnelResourceModel data.Transport
	transport.Protocol = types.StringValue(tunnel.Transport.Protocol)
	trans, _ := types.ObjectValueFrom(ctx, transport.attrTypes(), transport)

	tflog.Trace(ctx, "Before parameters and auth conversions")

	parameters.Id = types.StringValue(tunnel.Client.Authentication.Parameters.Id)
	parameters.ModifiedAt = types.StringValue(tunnel.Client.Authentication.Parameters.ModifiedAt)
	parameters.Secret = types.StringValue(tunnel.Client.Authentication.Parameters.Secret)
	//parameters.IdPrefix = types.StringValue(idprefix)

	params, _ := types.ObjectValueFrom(ctx, ParamsAttrTypes(), parameters)

	tflog.Trace(ctx, "Params done")

	auth.Type = types.StringValue(tunnel.Client.Authentication.Type)
	auth.Parameters = params

	tflog.Trace(ctx, "Assigned auth")

	auti, _ := types.ObjectValueFrom(ctx, AuthAttrTypes(), auth)

	client.DeviceType = types.StringValue(tunnel.Client.DeviceType)
	client.Authentication = auti

	clienti, _ := types.ObjectValueFrom(ctx, client.attrTypes(), client)

	tflog.Trace(ctx, "Before assigning to data")

	//To overcome discrepancy between Create and Update when Update returns more detailed uri
	uri := tunnel.Uri + "/" + strconv.FormatInt(tunnel.Id, 10)
	data.Uri = types.StringValue(uri)

	data.Name = types.StringValue(tunnel.Name)
	data.SiteOriginId = types.Int64Value(tunnel.SiteOriginId)
	data.Client = clienti
	data.Transport = trans
	data.ServiceType = types.StringValue(tunnel.ServiceType)
	data.NetworkCidrs = cidrs
	//data.Meta = nil
	data.ModifiedAt = types.StringValue(tunnel.ModifiedAt)
	data.CreatedAt = types.StringValue(tunnel.CreatedAt)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.Id = types.Int64Value(int64(tunnel.Id))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TunnelResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var stateclient TunnelClientResourceModel
	resp.Diagnostics.Append(data.Client.As(ctx, &stateclient, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var stateauth TunnelAuthResourceModel
	resp.Diagnostics.Append(stateclient.Authentication.As(ctx, &stateauth, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var stateparameters TunnelAuthParamsResourceModel
	resp.Diagnostics.Append(stateauth.Parameters.As(ctx, &stateparameters, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	tunnel, err := r.client.GetTunnel(data.Id.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella Tunnel",
			"Could not read Umbrella Tunnel ID "+strconv.FormatInt(data.Id.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	cidrs, _ := types.ListValueFrom(ctx, types.StringType, tunnel.NetworkCIDRs)

	var transport TunnelTransResourceModel
	transport.Protocol = types.StringValue(tunnel.Transport.Protocol)
	trans, _ := types.ObjectValueFrom(ctx, transport.attrTypes(), transport)

	var parameters TunnelAuthParamsResourceModel
	parameters.Id = types.StringValue(tunnel.Client.Authentication.Parameters.Id)
	parameters.ModifiedAt = types.StringValue(tunnel.Client.Authentication.Parameters.ModifiedAt)
	parameters.Secret = stateparameters.Secret
	parameters.IdPrefix = stateparameters.IdPrefix

	params, _ := types.ObjectValueFrom(ctx, ParamsAttrTypes(), parameters)

	var auth TunnelAuthResourceModel
	auth.Type = types.StringValue(tunnel.Client.Authentication.Type)
	auth.Parameters = params

	auti, _ := types.ObjectValueFrom(ctx, AuthAttrTypes(), auth)

	var client TunnelClientResourceModel
	client.DeviceType = types.StringValue(tunnel.Client.DeviceType)
	client.Authentication = auti

	clienti, _ := types.ObjectValueFrom(ctx, client.attrTypes(), client)

	data.Uri = types.StringValue(tunnel.Uri)
	data.Name = types.StringValue(tunnel.Name)
	data.SiteOriginId = types.Int64Value(tunnel.SiteOriginId)
	data.Client = clienti
	data.Transport = trans
	data.ServiceType = types.StringValue(tunnel.ServiceType)
	data.NetworkCidrs = cidrs
	//data.Meta = nil
	data.ModifiedAt = types.StringValue(tunnel.ModifiedAt)
	data.CreatedAt = types.StringValue(tunnel.CreatedAt)

	data.Id = types.Int64Value(int64(tunnel.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TunnelResourceModel
	var statedata *TunnelResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &statedata)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var networkcidrs []string
	resp.Diagnostics.Append(data.NetworkCidrs.ElementsAs(ctx, &networkcidrs, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//Doing conversion of ObjectType from parent TunnelResourceModel Transport to Go struct TunnelTransResourceModel to handle Unknown and Null values properly
	//https://discuss.hashicorp.com/t/thoughts-on-framework-v0-15-0/46277
	//https://discuss.hashicorp.com/t/nested-types-object-and-redundant-attrtypes/45927
	var transport TunnelTransResourceModel
	resp.Diagnostics.Append(data.Transport.As(ctx, &transport, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client TunnelClientResourceModel
	resp.Diagnostics.Append(data.Client.As(ctx, &client, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var auth TunnelAuthResourceModel
	resp.Diagnostics.Append(client.Authentication.As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var parameters TunnelAuthParamsResourceModel
	resp.Diagnostics.Append(auth.Parameters.As(ctx, &parameters, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	if resp.Diagnostics.HasError() {
		return
	}

	var stateclient TunnelClientResourceModel
	resp.Diagnostics.Append(statedata.Client.As(ctx, &stateclient, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var stateauth TunnelAuthResourceModel
	resp.Diagnostics.Append(stateclient.Authentication.As(ctx, &stateauth, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	var stateparameters TunnelAuthParamsResourceModel
	resp.Diagnostics.Append(stateauth.Parameters.As(ctx, &stateparameters, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true, UnhandledNullAsEmpty: true})...)

	//siteid, _ := strconv.Atoi(data.SiteId.ValueString())
	tunnelItem := buildTunnelItem(*data, client, auth, parameters, transport, networkcidrs)

	_, err := r.client.UpdateTunnel(statedata.Id.ValueInt64(), tunnelItem, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Umbrella Tunnel"+strconv.FormatInt(statedata.Id.ValueInt64(), 10),
			"Could not update tunnel, unexpected error: "+err.Error(),
		)
		return
	}

	tunnel, err := r.client.GetTunnel(statedata.Id.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Umbrella Tunnel",
			"Could not read Umbrella Tunnel ID "+strconv.FormatInt(statedata.Id.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	cidrs, _ := types.ListValueFrom(ctx, types.StringType, tunnel.NetworkCIDRs)

	transport.Protocol = types.StringValue(tunnel.Transport.Protocol)
	trans, _ := types.ObjectValueFrom(ctx, transport.attrTypes(), transport)

	parameters.Id = types.StringValue(tunnel.Client.Authentication.Parameters.Id)
	parameters.ModifiedAt = types.StringValue(tunnel.Client.Authentication.Parameters.ModifiedAt)
	parameters.Secret = stateparameters.Secret

	params, _ := types.ObjectValueFrom(ctx, ParamsAttrTypes(), parameters)

	auth.Type = types.StringValue(tunnel.Client.Authentication.Type)
	auth.Parameters = params

	auti, _ := types.ObjectValueFrom(ctx, AuthAttrTypes(), auth)

	client.DeviceType = types.StringValue(tunnel.Client.DeviceType)
	client.Authentication = auti

	clienti, _ := types.ObjectValueFrom(ctx, client.attrTypes(), client)

	data.Uri = types.StringValue(tunnel.Uri)
	data.Name = types.StringValue(tunnel.Name)
	data.SiteOriginId = types.Int64Value(tunnel.SiteOriginId)
	data.Client = clienti
	data.Transport = trans
	data.ServiceType = types.StringValue(tunnel.ServiceType)
	data.NetworkCidrs = cidrs
	//data.Meta = nil
	data.ModifiedAt = types.StringValue(tunnel.ModifiedAt)
	data.CreatedAt = types.StringValue(tunnel.CreatedAt)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	data.Id = types.Int64Value(int64(tunnel.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TunnelResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTunnel(data.Id.ValueInt64(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Umbrella Tunnel",
			"Could not delete Tunnel, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *TunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	id, _ := strconv.ParseInt(req.ID, 10, 64)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	//resource.ImportStatePassthroughID(ctx, path.Root("site_id"), req, resp)
}
