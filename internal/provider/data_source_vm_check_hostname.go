// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSourceVMCheckHostname{}

// NewDataSourceVMCheckHostname creates a new instance of the VM check hostname data source.
// This factory function is called by the provider to instantiate the data source.
func NewDataSourceVMCheckHostname() datasource.DataSource {
	return &DataSourceVMCheckHostname{}
}

type DataSourceVMCheckHostname struct {
	client      *client.ClientWithResponses
	defaultSite string
}

type VMCheckHostnameModel struct {
	ID          types.String `tfsdk:"id"`
	Hostname    types.String `tfsdk:"hostname"`
	Site        types.String `tfsdk:"site"`
	Status      types.String `tfsdk:"status"`
	Details     types.String `tfsdk:"details"`
	FQDN        types.String `tfsdk:"fqdn"`
	OwningUser  types.Int64  `tfsdk:"owning_user"`
	Owner       types.Object `tfsdk:"owner"`
	VMID        types.String `tfsdk:"vm_id"`
	IsAvailable types.Bool   `tfsdk:"is_available"`
}

type OwnerModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
}

// Metadata sets the data source type name for the VM check hostname data source.
// The type name is used in Terraform configurations as "fyre_vm_check_hostname".
func (d *DataSourceVMCheckHostname) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_check_hostname"
}

// Schema defines the structure and attributes of the VM check hostname data source.
// It specifies the required hostname parameter, optional site parameter, and computed
// attributes including availability status, owner information, and FQDN details.
func (d *DataSourceVMCheckHostname) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks if a hostname is available for use in the Fyre environment. Hostnames already in DNS will not be available.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The hostname being checked",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname to check for availability",
				Required:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the hostname check: 'success' if available, 'warning' if in use",
				Computed:            true,
			},
			"details": schema.StringAttribute{
				MarkdownDescription: "Human-readable message about hostname availability",
				Computed:            true,
			},
			"fqdn": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name (only present when hostname is available)",
				Computed:            true,
				Optional:            true,
			},
			"owning_user": schema.Int64Attribute{
				MarkdownDescription: "User ID of current owner (only present when hostname is in use)",
				Computed:            true,
				Optional:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "Owner details (only present when hostname is in use)",
				Computed:            true,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Owner user ID",
						Computed:            true,
						Optional:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "Owner username",
						Computed:            true,
						Optional:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Owner email address",
						Computed:            true,
						Optional:            true,
					},
				},
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "VM ID of the VM using this hostname (only present when hostname is in use)",
				Computed:            true,
				Optional:            true,
			},
			"is_available": schema.BoolAttribute{
				MarkdownDescription: "Convenience boolean indicating if the hostname is available (true) or in use (false)",
				Computed:            true,
			},
		},
	}
}

// Configure initializes the VM check hostname data source with the Fyre API client
// and default site configuration from the provider. This method is called by the
// framework during provider initialization.
func (d *DataSourceVMCheckHostname) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FyreProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FyreProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
	d.defaultSite = providerData.DefaultSite
}

// Read checks if a hostname is available for use in the Fyre environment by querying
// the Fyre API. It returns availability status, FQDN if available, or owner information
// if the hostname is already in use. Hostnames already in DNS will not be available.
func (d *DataSourceVMCheckHostname) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMCheckHostnameModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request parameters
	site := client.CheckHostnameParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.CheckHostnameParamsSite(d.defaultSite)
	}

	hostname := client.Hostname(data.Hostname.ValueString())

	// Call API
	checkResp, err := d.client.CheckHostnameWithResponse(ctx, hostname, &client.CheckHostnameParams{
		Site: &site,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to check hostname availability: %s", err),
		)
		return
	}

	if checkResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", checkResp.StatusCode(), string(checkResp.Body)),
		)
		return
	}

	if checkResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse hostname check response. Body: %s", string(checkResp.Body)),
		)
		return
	}

	tflog.Debug(ctx, "hostname check response", map[string]any{
		"status":  checkResp.JSON200.Status,
		"details": checkResp.JSON200.Details,
	})

	// Map to Terraform state
	data.ID = types.StringValue(data.Hostname.ValueString())
	data.Site = types.StringValue(string(site))

	// Status and details are always present
	if checkResp.JSON200.Status != nil {
		data.Status = types.StringValue(string(*checkResp.JSON200.Status))
		data.IsAvailable = types.BoolValue(*checkResp.JSON200.Status == "success")
	} else {
		data.Status = types.StringNull()
		data.IsAvailable = types.BoolNull()
	}

	if checkResp.JSON200.Details != nil {
		data.Details = types.StringValue(*checkResp.JSON200.Details)
	} else {
		data.Details = types.StringNull()
	}

	// FQDN is only present when hostname is available (status: success)
	if checkResp.JSON200.Fqdn != nil {
		data.FQDN = types.StringValue(*checkResp.JSON200.Fqdn)
	} else {
		data.FQDN = types.StringNull()
	}

	// Owner information is only present when hostname is in use (status: warning)
	if checkResp.JSON200.OwningUser != nil {
		data.OwningUser = types.Int64Value(int64(*checkResp.JSON200.OwningUser))
	} else {
		data.OwningUser = types.Int64Null()
	}

	if checkResp.JSON200.VmId != nil {
		data.VMID = types.StringValue(*checkResp.JSON200.VmId)
	} else {
		data.VMID = types.StringNull()
	}

	// Handle owner nested object
	ownerModel := OwnerModel{
		ID:       types.Int64Null(),
		Username: types.StringNull(),
		Email:    types.StringNull(),
	}

	if checkResp.JSON200.Owner != nil {
		if checkResp.JSON200.Owner.Id != nil {
			ownerModel.ID = types.Int64Value(int64(*checkResp.JSON200.Owner.Id))
		}
		if checkResp.JSON200.Owner.Username != nil {
			ownerModel.Username = types.StringValue(*checkResp.JSON200.Owner.Username)
		}
		if checkResp.JSON200.Owner.Email != nil {
			ownerModel.Email = types.StringValue(*checkResp.JSON200.Owner.Email)
		}
	}

	ownerObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":       types.Int64Type,
		"username": types.StringType,
		"email":    types.StringType,
	}, ownerModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Owner = ownerObj

	tflog.Trace(ctx, "read vm_check_hostname data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
