// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSourceVMStatus{}

// NewDataSourceVMStatus creates a new instance of the VM status data source.
// This factory function is called by the provider to instantiate the data source.
func NewDataSourceVMStatus() datasource.DataSource {
	return &DataSourceVMStatus{}
}

type DataSourceVMStatus struct {
	client      *client.ClientWithResponses
	defaultSite string
}

type VMStatusModel struct {
	ID          types.String `tfsdk:"id"`
	VMID        types.String `tfsdk:"vm_id"`
	IP          types.String `tfsdk:"ip"`
	FQDN        types.String `tfsdk:"fqdn"`
	Site        types.String `tfsdk:"site"`
	LastOSState types.String `tfsdk:"last_os_state"`
	Status      types.String `tfsdk:"status"`
}

// Metadata sets the data source type name for the VM status data source.
// The type name is used in Terraform configurations as "fyre_vm_status".
func (d *DataSourceVMStatus) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_status"
}

// Schema defines the structure and attributes of the VM status data source.
// It specifies optional vm_id, ip, and fqdn parameters (at least one required),
// optional site parameter, and computed attributes including last OS state and status.
func (d *DataSourceVMStatus) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the current status of a Fyre VM. You must provide at least one identifier: vm_id, ip, or fqdn. The data source will try each non-null identifier until it successfully retrieves the VM status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier used to retrieve the VM status (same as the successful lookup field)",
				Computed:            true,
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "VM identifier (format: x-xxxxxxx). At least one of vm_id, ip, or fqdn must be provided.",
				Optional:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP address of the VM. At least one of vm_id, ip, or fqdn must be provided.",
				Optional:            true,
			},
			"fqdn": schema.StringAttribute{
				MarkdownDescription: "Fully Qualified Domain Name of the VM (must be in DNS). At least one of vm_id, ip, or fqdn must be provided.",
				Optional:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"last_os_state": schema.StringAttribute{
				MarkdownDescription: "The last known operating system state of the VM (e.g., running, stopped)",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status message about any requests in progress",
				Computed:            true,
			},
		},
	}
}

// Configure initializes the VM status data source with the Fyre API client
// and default site configuration from the provider. This method is called by the
// framework during provider initialization.
func (d *DataSourceVMStatus) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the current status of a Fyre VM from the API. It accepts vm_id,
// ip, or fqdn as identifiers (at least one required) and tries each in order until
// successful. Returns the VM's last OS state and any status messages about requests
// in progress.
func (d *DataSourceVMStatus) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMStatusModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one identifier is provided
	if data.VMID.IsNull() && data.IP.IsNull() && data.FQDN.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"At least one of vm_id, ip, or fqdn must be provided",
		)
		return
	}

	// Prepare site parameter
	site := client.GetVMStatusParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.GetVMStatusParamsSite(d.defaultSite)
	}

	// Try each identifier in order until one succeeds
	identifiers := []struct {
		value string
		name  string
	}{
		{data.VMID.ValueString(), "vm_id"},
		{data.IP.ValueString(), "ip"},
		{data.FQDN.ValueString(), "fqdn"},
	}

	var statusResp *client.GetVMStatusResponse
	var err error
	var successfulIdentifier string

	for _, identifier := range identifiers {
		if identifier.value == "" {
			continue
		}

		tflog.Debug(ctx, fmt.Sprintf("Trying to fetch VM status using %s: %s", identifier.name, identifier.value))

		statusResp, err = d.client.GetVMStatusWithResponse(ctx, identifier.value, &client.GetVMStatusParams{
			Site: &site,
		})
		if err != nil {
			tflog.Debug(ctx, fmt.Sprintf("Error fetching VM status with %s: %s", identifier.name, err))
			continue
		}

		if statusResp.StatusCode() == 200 && statusResp.JSON200 != nil {
			successfulIdentifier = identifier.value
			tflog.Debug(ctx, fmt.Sprintf("Successfully fetched VM status using %s", identifier.name))
			break
		}

		tflog.Debug(ctx, fmt.Sprintf("Failed to fetch VM status with %s, status: %d", identifier.name, statusResp.StatusCode()))
	}

	if successfulIdentifier == "" {
		resp.Diagnostics.AddError(
			"VM Not Found",
			fmt.Sprintf("Unable to find VM using any of the provided identifiers. Last error: %v", err),
		)
		return
	}

	if statusResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse VM status response. Body: %s", string(statusResp.Body)),
		)
		return
	}

	// Map response to Terraform state
	vmStatus := statusResp.JSON200

	data.ID = types.StringValue(successfulIdentifier)
	data.Site = types.StringValue(string(site))

	// Set the identifier fields based on what was used
	if data.VMID.ValueString() == successfulIdentifier {
		data.VMID = types.StringValue(successfulIdentifier)
	} else if data.IP.ValueString() == successfulIdentifier {
		data.IP = types.StringValue(successfulIdentifier)
	} else if data.FQDN.ValueString() == successfulIdentifier {
		data.FQDN = types.StringValue(successfulIdentifier)
	}

	if vmStatus.LastOsState != nil {
		data.LastOSState = types.StringValue(*vmStatus.LastOsState)
	} else {
		data.LastOSState = types.StringNull()
	}

	if vmStatus.Status != nil {
		data.Status = types.StringValue(*vmStatus.Status)
	} else {
		data.Status = types.StringNull()
	}

	tflog.Trace(ctx, "read vm_status data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
