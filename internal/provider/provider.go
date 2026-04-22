// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure FyreProvider satisfies various provider interfaces.
var (
	_ provider.Provider                       = &FyreProvider{}
	_ provider.ProviderWithFunctions          = &FyreProvider{}
	_ provider.ProviderWithEphemeralResources = &FyreProvider{}
	_ provider.ProviderWithActions            = &FyreProvider{}
)

// FyreProvider defines the provider implementation.
type FyreProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// FyreProviderModel describes the provider data model.
type FyreProviderModel struct {
	Username            types.String `tfsdk:"username"`
	APIKey              types.String `tfsdk:"api_key"`
	DefaultSite         types.String `tfsdk:"site"`
	DefaultProductGroup types.Int64  `tfsdk:"product_group_id"`
}

// FyreProviderData is the data passed to data sources and resources.
type FyreProviderData struct {
	Client                 *client.ClientWithResponses
	DefaultSite            string
	DefaultProductGroupID  *int64
}

func (p *FyreProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fyre"
	resp.Version = p.version
}

func (p *FyreProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Fyre provider is used to interact with IBM Fyre cloud resources.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Fyre username for authentication. Can also be set via FYRE_USERNAME environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Fyre API key for authentication. Can also be set via FYRE_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Default Fyre site location. Can be either 'svl' or 'rtp'. Default is 'svl'. Can also be set via FYRE_SITE environment variable.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"rtp", "svl"}...),
				},
			},
			"product_group_id": schema.Int64Attribute{
				MarkdownDescription: "Default product group ID for resources and data sources. Can be overridden at the resource/data source level. Can also be set via FYRE_PRODUCT_GROUP_ID environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *FyreProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FyreProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get username from config or environment variable
	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("FYRE_USERNAME")
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"Username must be provided via the 'username' attribute or FYRE_USERNAME environment variable.",
		)
	}

	// Get API key from config or environment variable
	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("FYRE_API_KEY")
	}

	// Validate credentials
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"API key must be provided via the 'api_key' attribute or FYRE_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get site from config or environment variable, default to svl
	site := data.DefaultSite.ValueString()
	if site == "" {
		site = os.Getenv("FYRE_SITE")
	}
	if site == "" {
		site = "svl"
	}

	// Get product_group_id from config or environment variable (optional, no default)
	var productGroupID *int64
	if !data.DefaultProductGroup.IsNull() && !data.DefaultProductGroup.IsUnknown() {
		pgID := data.DefaultProductGroup.ValueInt64()
		productGroupID = &pgID
	} else {
		// Check environment variable if not in config
		if envPGID := os.Getenv("FYRE_PRODUCT_GROUP_ID"); envPGID != "" {
			var pgID int64
			if _, err := fmt.Sscanf(envPGID, "%d", &pgID); err == nil {
				productGroupID = &pgID
			}
		}
	}

	// Create basic auth request editor
	basicAuthEditor := func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(username, apiKey)
		return nil
	}

	// Create client with basic auth
	fyreClient, err := client.NewClientWithResponses(
		client.ServerUrlSVLSanJoseProductionServer,
		client.WithRequestEditorFn(basicAuthEditor),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Fyre API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Fyre API client: %s", err.Error()),
		)
		return
	}

	// Create provider data
	providerData := &FyreProviderData{
		Client:                fyreClient,
		DefaultSite:           site,
		DefaultProductGroupID: productGroupID,
	}

	// Make the client available to data sources and resources
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *FyreProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceVM,
	}
}

func (p *FyreProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *FyreProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataSourceClusterDetails,
		NewDataSourceClusters,
		NewDataSourceQuota,
		NewDataSourceStencils,
		NewDataSourceUser,
		NewDataSourceVMCheckHostname,
		NewDataSourceVMDetails,
		NewDataSourceVMOSAvailable,
		NewDataSourceVMSnapshots,
		NewDataSourceVMStatus,
	}
}

func (p *FyreProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *FyreProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FyreProvider{
			version: version,
		}
	}
}
