# Example: Creating a VM Resource

This example demonstrates how to use the `tf-gen` skill to create a complete Terraform resource for managing Fyre VMs with full CRUD operations.

## Step 1: Switch to the Mode

```
Switch to tf-gen mode and create a resource for managing VMs using the CreateVM, GetVM, UpdateVM, and DeleteVM operations
```

## Step 2: Verify Client Library

The skill will check if `internal/client/client.gen.go` has:
- `CreateVMWithResponse` method (POST)
- `GetVMWithResponse` method (GET)
- `UpdateVMWithResponse` method (PUT/PATCH)
- `DeleteVMWithResponse` method (DELETE)
- Proper request and response types

If missing or incomplete, it will recommend using `fyre-api-updater` mode first.

## Step 3: Generated Files

### File 1: `internal/provider/resource_vm.go`

```go
// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ResourceVM{}
var _ resource.ResourceWithImportState = &ResourceVM{}

// NewResourceVM creates a new VM resource.
func NewResourceVM() resource.Resource {
	return &ResourceVM{}
}

// ResourceVM defines the resource implementation.
type ResourceVM struct {
	client      *client.ClientWithResponses
	defaultSite string
}

// VMModel describes the resource data model.
type VMModel struct {
	ID          types.String `tfsdk:"id"`
	Site        types.String `tfsdk:"site"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Platform    types.String `tfsdk:"platform"`
	Memory      types.Int64  `tfsdk:"memory"`
	CPU         types.Int64  `tfsdk:"cpu"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	// Add other fields as needed
}

// Metadata returns the resource type name.
func (r *ResourceVM) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

// Schema defines the schema for the resource.
func (r *ResourceVM) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Fyre VM resource with full lifecycle support.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the VM",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VM",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the VM",
				Optional:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform/OS for the VM",
				Required:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Memory allocation in GB",
				Required:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: "Number of CPU cores",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the VM",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the VM was created",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ResourceVM) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FyreProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FyreProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
	r.defaultSite = providerData.DefaultSite
}

// Create creates the resource and sets the initial Terraform state.
func (r *ResourceVM) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine site
	site := data.Site.ValueString()
	if site == "" {
		site = r.defaultSite
	}

	// Build API request
	createReq := client.CreateVMJSONRequestBody{
		Name:        data.Name.ValueString(),
		Platform:    data.Platform.ValueString(),
		Memory:      int(data.Memory.ValueInt64()),
		CPU:         int(data.CPU.ValueInt64()),
		Site:        &site,
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		createReq.Description = &desc
	}

	tflog.Debug(ctx, "Creating VM", map[string]any{
		"name":     createReq.Name,
		"platform": createReq.Platform,
	})

	// Call API
	createResp, err := r.client.CreateVMWithResponse(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create VM: %s", err),
		)
		return
	}

	if createResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", createResp.StatusCode(), string(createResp.Body)),
		)
		return
	}

	if createResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse create response. Body: %s", string(createResp.Body)),
		)
		return
	}

	// Map response to state
	data.ID = types.StringValue(createResp.JSON201.ID)
	data.Site = types.StringValue(site)
	
	if createResp.JSON201.Status != nil {
		data.Status = types.StringValue(*createResp.JSON201.Status)
	} else {
		data.Status = types.StringNull()
	}

	if createResp.JSON201.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*createResp.JSON201.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}

	tflog.Trace(ctx, "created VM resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ResourceVM) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API
	readResp, err := r.client.GetVMWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read VM: %s", err),
		)
		return
	}

	if readResp.StatusCode() == 404 {
		// VM no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	if readResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", readResp.StatusCode(), string(readResp.Body)),
		)
		return
	}

	if readResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse read response. Body: %s", string(readResp.Body)),
		)
		return
	}

	// Update state with latest data
	if readResp.JSON200.Name != nil {
		data.Name = types.StringValue(*readResp.JSON200.Name)
	}

	if readResp.JSON200.Description != nil {
		data.Description = types.StringValue(*readResp.JSON200.Description)
	} else {
		data.Description = types.StringNull()
	}

	if readResp.JSON200.Platform != nil {
		data.Platform = types.StringValue(*readResp.JSON200.Platform)
	}

	if readResp.JSON200.Memory != nil {
		data.Memory = types.Int64Value(int64(*readResp.JSON200.Memory))
	}

	if readResp.JSON200.CPU != nil {
		data.CPU = types.Int64Value(int64(*readResp.JSON200.CPU))
	}

	if readResp.JSON200.Status != nil {
		data.Status = types.StringValue(*readResp.JSON200.Status)
	}

	tflog.Trace(ctx, "read VM resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ResourceVM) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API request
	updateReq := client.UpdateVMJSONRequestBody{
		Name:     data.Name.ValueString(),
		Memory:   int(data.Memory.ValueInt64()),
		CPU:      int(data.CPU.ValueInt64()),
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		updateReq.Description = &desc
	}

	tflog.Debug(ctx, "Updating VM", map[string]any{
		"id":   data.ID.ValueString(),
		"name": updateReq.Name,
	})

	// Call API
	updateResp, err := r.client.UpdateVMWithResponse(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update VM: %s", err),
		)
		return
	}

	if updateResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", updateResp.StatusCode(), string(updateResp.Body)),
		)
		return
	}

	if updateResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse update response. Body: %s", string(updateResp.Body)),
		)
		return
	}

	// Update state with response data
	if updateResp.JSON200.Status != nil {
		data.Status = types.StringValue(*updateResp.JSON200.Status)
	}

	tflog.Trace(ctx, "updated VM resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ResourceVM) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting VM", map[string]any{
		"id": data.ID.ValueString(),
	})

	// Call API
	deleteResp, err := r.client.DeleteVMWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete VM: %s", err),
		)
		return
	}

	if deleteResp.StatusCode() != 204 && deleteResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("API returned status %d: %s", deleteResp.StatusCode(), string(deleteResp.Body)),
		)
		return
	}

	tflog.Trace(ctx, "deleted VM resource")
}

// ImportState imports an existing resource into Terraform state.
func (r *ResourceVM) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### File 2: `internal/provider/resource_vm_test.go`

```go
// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceVM(t *testing.T) {
	if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
		t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVMConfig("test-vm-initial", "ubuntu-20.04", 8, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "name", "test-vm-initial"),
					resource.TestCheckResourceAttr("fyre_vm.test", "platform", "ubuntu-20.04"),
					resource.TestCheckResourceAttr("fyre_vm.test", "memory", "8"),
					resource.TestCheckResourceAttr("fyre_vm.test", "cpu", "4"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "id"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "site"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "status"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "created_at"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResourceVMConfig("test-vm-updated", "ubuntu-20.04", 16, 8),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fyre_vm.test", "name", "test-vm-updated"),
					resource.TestCheckResourceAttr("fyre_vm.test", "memory", "16"),
					resource.TestCheckResourceAttr("fyre_vm.test", "cpu", "8"),
					resource.TestCheckResourceAttrSet("fyre_vm.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "fyre_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceVMConfig(name, platform string, memory, cpu int) string {
	return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

resource "fyre_vm" "test" {
  name        = %[1]q
  platform    = %[2]q
  memory      = %[3]d
  cpu         = %[4]d
  description = "Terraform acceptance test VM"
}
`, name, platform, memory, cpu)
}
```

### File 3: Update `internal/provider/provider.go`

Add to the `Resources()` method:

```go
func (p *FyreProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceVM,  // Add this line
	}
}
```

## Step 4: Test

```bash
# Run the acceptance test
export FYRE_USERNAME="your-username"
export FYRE_API_KEY="your-api-key"
TF_ACC=1 go test -v ./internal/provider -run TestAccResourceVM

# Generate documentation
make generate
```

## Key Patterns Demonstrated

### 1. Naming Conventions
- **Function**: `NewResource<Name>()` (e.g., `NewResourceVM`)
- **Struct**: `Resource<Name>` (e.g., `ResourceVM`)
- **Model**: `<Name>Model` (e.g., `VMModel`)
- **Nested Models**: `<NestedName>Model` (e.g., `NetworkModel`)

### 2. Interface Implementation
```go
var _ resource.Resource = &ResourceVM{}
var _ resource.ResourceWithImportState = &ResourceVM{}
```

### 3. Attribute Types
- **Required**: User must provide (name, platform, memory, cpu)
- **Optional**: User can provide (description)
- **Computed**: Provider sets (id, status, created_at)
- **Optional + Computed**: User can provide or API computes (site)

### 4. Plan Modifiers
```go
PlanModifiers: []planmodifier.String{
    stringplanmodifier.UseStateForUnknown(),
}
```
- Prevents unnecessary resource replacement
- Used for computed fields that shouldn't trigger updates

### 5. CRUD Operations

**Create**: `req.Plan` → API POST → `resp.State`
- Read desired state from plan
- Call API to create resource
- Write created resource to state

**Read**: `req.State` → API GET → `resp.State`
- Read current state
- Call API to get latest data
- Update state with latest data
- Handle 404 by removing from state

**Update**: `req.Plan` → API PUT/PATCH → `resp.State`
- Read desired state from plan
- Call API to update resource
- Write updated resource to state

**Delete**: `req.State` → API DELETE → (no state)
- Read current state
- Call API to delete resource
- No state write (resource removed)

### 6. Import Support
```go
func (r *ResourceVM) ImportState(ctx, req, resp) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```
- Allows importing existing resources
- Uses ID as the import identifier

### 7. Testing Patterns

**Multiple Test Steps:**
1. Create with initial values
2. Update with new values
3. Import existing resource

**Assertions:**
- `TestCheckResourceAttr`: Verify expected values
- `TestCheckResourceAttrSet`: Verify computed values exist

### 8. Error Handling
- Check HTTP status codes
- Handle 404 in Read (remove from state)
- Validate response parsing
- Provide clear error messages

## Common Variations

### For Immutable Fields

If a field cannot be updated (requires resource replacement):

```go
"platform": schema.StringAttribute{
    Required: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
```

### For Nested Objects

```go
// In model
Config types.Object `tfsdk:"config"`

// Nested model
type ConfigModel struct {
    Setting1 types.String `tfsdk:"setting1"`
    Setting2 types.Int64  `tfsdk:"setting2"`
}

// In schema
"config": schema.SingleNestedAttribute{
    Required: true,
    Attributes: map[string]schema.Attribute{
        "setting1": schema.StringAttribute{
            Required: true,
        },
        "setting2": schema.Int64Attribute{
            Required: true,
        },
    },
},

// In Create/Update
var configModel ConfigModel
resp.Diagnostics.Append(data.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{})...)
if resp.Diagnostics.HasError() {
    return
}

createReq.Config = &client.Config{
    Setting1: configModel.Setting1.ValueString(),
    Setting2: int(configModel.Setting2.ValueInt64()),
}
```

### For List Fields

```go
// In model
Tags types.List `tfsdk:"tags"`

// In schema
"tags": schema.ListAttribute{
    Optional:    true,
    ElementType: types.StringType,
},

// In Create/Update
var tags []string
resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
if resp.Diagnostics.HasError() {
    return
}

createReq.Tags = &tags
```

### For Resources Without Update Support

If the API doesn't support updates:

```go
func (r *ResourceVM) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    resp.Diagnostics.AddError(
        "Update Not Supported",
        "This resource does not support updates. To modify, please destroy and recreate the resource.",
    )
}
```

Or use `RequiresReplace()` on all mutable fields to force recreation.

## Differences from Data Sources

| Aspect | Data Source | Resource |
|--------|-------------|----------|
| **Methods** | Read only | Create, Read, Update, Delete |
| **State Source** | Config | Plan (Create/Update), State (Read/Delete) |
| **Attributes** | Mostly Computed | Required, Optional, Computed |
| **Plan Modifiers** | Rare | Common |
| **Import** | N/A | ImportState method |
| **Testing** | Single step | Multiple steps |
| **State Management** | Read-only | Full lifecycle |

## Tips

- Always implement all CRUD methods, even if some are no-ops
- Use plan modifiers to prevent unnecessary replacements
- Handle 404 in Read by removing from state
- Test both create and update paths
- Include ImportState testing when possible
- Use clear, descriptive error messages
- Log important operations with tflog
- Follow existing resource patterns in the codebase
