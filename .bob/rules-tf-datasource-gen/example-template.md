# Example: Creating a User Data Source

This example demonstrates how to use the `tf-datasource-gen` skill to create a complete Terraform data source for the Fyre User API.

## Step 1: Switch to the Mode

```
Switch to tf-datasource-gen mode and create a data source for the user resource using the GetUserDetails operation
```

## Step 2: Verify Client Library

The skill will check if `internal/client/client.gen.go` has:
- `GetUserDetailsWithResponse` method
- `UserDetails` response type
- Proper parameter types

If missing or incomplete, it will recommend using `fyre-api-updater` mode first.

## Step 3: Generated Files

### File 1: `internal/provider/data_source_user.go`

```go
// Copyright IBM Corp. 2021, 2026
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

var _ datasource.DataSource = &DataSourceUser{}

func NewDataSourceUser() datasource.DataSource {
	return &DataSourceUser{}
}

type DataSourceUser struct {
	client      *client.ClientWithResponses
	defaultSite string
}

type UserModel struct {
	ID            types.String `tfsdk:"id"`
	Site          types.String `tfsdk:"site"`
	FullName      types.String `tfsdk:"full_name"`
	Email         types.String `tfsdk:"email"`
	Authenticated types.Bool   `tfsdk:"authenticated"`
	// Add other fields from UserDetails response
}

func (d *DataSourceUser) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *DataSourceUser) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches details about the authenticated Fyre user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The user identifier",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "User's full name",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User's email address",
				Computed:            true,
			},
			"authenticated": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is authenticated",
				Computed:            true,
			},
			// Add other attributes...
		},
	}
}

func (d *DataSourceUser) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceUser) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request parameters
	site := client.GetUserDetailsParamsSite(data.Site.ValueString())
	if site == "" {
		site = client.GetUserDetailsParamsSite(d.defaultSite)
	}

	// Call API
	userResp, err := d.client.GetUserDetailsWithResponse(ctx, &client.GetUserDetailsParams{
		Site: &site,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read user details: %s", err),
		)
		return
	}

	if userResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", userResp.StatusCode(), string(userResp.Body)),
		)
		return
	}

	if userResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse user response. Body: %s", string(userResp.Body)),
		)
		return
	}

	// Map to Terraform state
	data.ID = types.StringValue("user")
	data.Site = types.StringValue(string(site))

	if userResp.JSON200.FullName != nil {
		data.FullName = types.StringValue(*userResp.JSON200.FullName)
	} else {
		data.FullName = types.StringNull()
	}

	if userResp.JSON200.Email != nil {
		data.Email = types.StringValue(*userResp.JSON200.Email)
	} else {
		data.Email = types.StringNull()
	}

	if userResp.JSON200.Authenticated != nil {
		data.Authenticated = types.BoolValue(*userResp.JSON200.Authenticated)
	} else {
		data.Authenticated = types.BoolNull()
	}

	// Map other fields...

	tflog.Trace(ctx, "read user data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### File 2: `internal/provider/data_source_user_test.go`

```go
// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "site"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "full_name"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "email"),
					resource.TestCheckResourceAttrSet("data.fyre_user.test", "authenticated"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig() string {
	return `
provider "fyre" {
  site = "svl"
}

data "fyre_user" "test" {
}
`
}
```

### File 3: Update `internal/provider/provider.go`

Add to the `DataSources()` method:

```go
func (p *FyreProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataSourceUser,  // Add this line
		NewDataSourceQuota,
	}
}
```

## Step 4: Test

```bash
# Run the acceptance test
export FYRE_USERNAME="your-username"
export FYRE_API_KEY="your-api-key"
go test -v ./internal/provider -run TestAccDataSourceUser

# Generate documentation
make generate
```

## Key Patterns Demonstrated

1. **Naming Conventions**:
   - Function: `NewDataSource<Name>()` (e.g., `NewDataSourceUser`)
   - Struct: `DataSource<Name>` (e.g., `DataSourceUser`)
   - Model: `<Name>Model` (e.g., `UserModel`)
   - Nested Models: `<NestedName>Model` (e.g., `DevelopmentModel`)

2. **Site Parameter**: Optional, computed, defaults to provider's site or 'svl'
3. **Nil Handling**: Check for nil pointers before accessing values
4. **Type Conversion**: Use `types.StringValue()`, `types.BoolValue()`, etc.
5. **Error Handling**: Check HTTP status, parse errors, nil responses
6. **Logging**: Use `tflog.Trace()` for debugging
7. **Testing**: Skip if credentials not set, verify all computed attributes

## Common Variations

### For List Fields

```go
// In model
Platforms types.List `tfsdk:"platforms"`

// In schema
"platforms": schema.ListAttribute{
	MarkdownDescription: "List of available platforms",
	Computed:            true,
	ElementType:         types.StringType,
},

// In Read method
if response.Platforms != nil && len(*response.Platforms) > 0 {
	platformsList := make([]attr.Value, 0, len(*response.Platforms))
	for _, platform := range *response.Platforms {
		platformsList = append(platformsList, types.StringValue(platform))
	}
	listValue, diags := types.ListValue(types.StringType, platformsList)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Platforms = listValue
	}
}
```

### For Nested Objects

```go
// In model
Details types.Object `tfsdk:"details"`

// Nested model
type DetailsModel struct {
	Field1 types.String `tfsdk:"field1"`
	Field2 types.Int64  `tfsdk:"field2"`
}

// In schema
"details": schema.SingleNestedAttribute{
	MarkdownDescription: "Detailed information",
	Computed:            true,
	Attributes: map[string]schema.Attribute{
		"field1": schema.StringAttribute{
			Computed: true,
		},
		"field2": schema.Int64Attribute{
			Computed: true,
		},
	},
},

// In Read method
detailsModel := DetailsModel{
	Field1: types.StringNull(),
	Field2: types.Int64Null(),
}

if response.Details != nil {
	if response.Details.Field1 != nil {
		detailsModel.Field1 = types.StringValue(*response.Details.Field1)
	}
	if response.Details.Field2 != nil {
		detailsModel.Field2 = types.Int64Value(int64(*response.Details.Field2))
	}
}

detailsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
	"field1": types.StringType,
	"field2": types.Int64Type,
}, detailsModel)
resp.Diagnostics.Append(diags...)
if !resp.Diagnostics.HasError() {
	data.Details = detailsObj
}
```