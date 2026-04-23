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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSourceUserAPIKey{}

// NewDataSourceUserAPIKey creates a new user API key data source.
func NewDataSourceUserAPIKey() datasource.DataSource {
	return &DataSourceUserAPIKey{}
}

// DataSourceUserAPIKey defines the data source implementation.
type DataSourceUserAPIKey struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// UserAPIKeyModel describes the data source data model.
type UserAPIKeyModel struct {
	ID         types.String `tfsdk:"id"`
	Site       types.String `tfsdk:"site"`
	Status     types.String `tfsdk:"status"`
	Details    types.String `tfsdk:"details"`
	Expiration types.String `tfsdk:"expiration"`
}

// Metadata returns the data source type name.
func (d *DataSourceUserAPIKey) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_api_key"
}

// Schema defines the schema for the data source.
func (d *DataSourceUserAPIKey) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the expiration date for the authenticated user's API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (always set to 'user_api_key')",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Response status from the API",
				Computed:            true,
			},
			"details": schema.StringAttribute{
				MarkdownDescription: "Human-readable message about the API key",
				Computed:            true,
			},
			"expiration": schema.StringAttribute{
				MarkdownDescription: "API key expiration timestamp in format 'YYYY-MM-DD HH:MM:SS'",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSourceUserAPIKey) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

// Read refreshes the Terraform state with the latest data.
func (d *DataSourceUserAPIKey) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserAPIKeyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine site to use
	site := data.Site.ValueString()
	if site == "" {
		site = d.defaultSite
	}

	// Prepare API request parameters
	params := &client.GetUserAPIKeyParams{}
	if site != "" {
		siteParam := client.GetUserAPIKeyParamsSite(site)
		params.Site = &siteParam
	}

	tflog.Debug(ctx, "Fetching user API key expiration", map[string]any{
		"site": site,
	})

	// Call API
	apiResp, err := d.client.GetUserAPIKeyWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read user API key: %s", err),
		)
		return
	}

	if apiResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse user API key response. Body: %s", string(apiResp.Body)),
		)
		return
	}

	// Map response to Terraform state
	data.ID = types.StringValue("user_api_key")
	data.Site = types.StringValue(site)

	// Map API response fields
	if apiResp.JSON200.Status != nil {
		data.Status = types.StringValue(*apiResp.JSON200.Status)
	} else {
		data.Status = types.StringNull()
	}

	if apiResp.JSON200.Details != nil {
		data.Details = types.StringValue(*apiResp.JSON200.Details)
	} else {
		data.Details = types.StringNull()
	}

	if apiResp.JSON200.Expiration != nil {
		data.Expiration = types.StringValue(*apiResp.JSON200.Expiration)
	} else {
		data.Expiration = types.StringNull()
	}

	tflog.Trace(ctx, "read user API key data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
