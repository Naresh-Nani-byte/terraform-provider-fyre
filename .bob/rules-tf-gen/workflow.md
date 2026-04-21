# Terraform Generator Workflow

This mode generates complete Terraform data sources and resources for Fyre API endpoints.

## Usage

Specify whether you want a data source (read-only) or resource (full CRUD):

**Data Source:**
```
Create a data source for the user resource using the GetUserDetails operation
```

**Resource:**
```
Create a resource for managing VMs using the CreateVM, GetVM, UpdateVM, and DeleteVM operations
```

## Type Determination

- **Data Source**: Read-only operations, single Read method, computed attributes
- **Resource**: Full CRUD lifecycle (Create, Read, Update, Delete), state management

If unclear from the request, ask the user to clarify which type is needed.

---

# Data Source Workflow

## Workflow Steps

### 1. Prerequisites Check
- Verify `internal/client/client.gen.go` has the necessary types
- Check for the API operation method (e.g., `GetUserDetailsWithResponse`)
- Verify response types are complete and accurate
- **If incomplete**: Switch to `fyre-api-updater` mode first

### 2. Analyze API Structure
- Review OpenAPI spec in `internal/client/api.yaml`
- Identify operation ID and response schema
- Note required vs optional parameters
- Understand response structure (nested objects, arrays, etc.)

### 3. Generate Data Source Implementation

Create `internal/provider/data_source_<name>.go` with:

**CRITICAL Naming Conventions:**
- **Function**: `NewDataSource<Name>()` (e.g., `NewDataSourceVMSnapshots`, `NewDataSourceUser`)
- **Struct**: `DataSource<Name>` (e.g., `DataSourceVMSnapshots`, `DataSourceUser`)
- **Model**: `<Name>Model` (e.g., `UserModel`, `VMDetailsModel`) - NO DataSource prefix
- **Nested Models**: `<NestedName>Model` (e.g., `DevelopmentModel`, `SnapshotModel`) - NO DataSource prefix

**Examples of Correct Naming:**
```go
// ✅ CORRECT
func NewDataSourceUser() datasource.DataSource { return &DataSourceUser{} }
type DataSourceUser struct { ... }
type UserModel struct { ... }
type DevelopmentModel struct { ... }  // nested model
type UserDevelopmentModel struct { ... }  // nested model when conflicts arise

// ❌ WRONG - Do NOT use these patterns
func NewUserDataSource() datasource.DataSource { ... }  // Wrong order
type UserDataSource struct { ... }  // Wrong order
type UserDataSourceModel struct { ... }  // Unnecessary suffix
type DataSourceDevelopmentModel struct { ... }  // Unnecessary prefix
```

**CRITICAL Attribute Naming (must be consistent across all data sources):**
- VM identifier: `vm_id` (NOT `vm_identifier`)
- User identifier: `user_id` (NOT `user_identifier`)
- Cluster identifier: `cluster_id` (NOT `cluster_identifier`)
- **Always check existing data sources for the correct attribute name**

**Required Components:**
- Copyright header: `// Copyright IBM Corp. 2021, 2026`
- Data source struct implementing `datasource.DataSource`
- Model structs for data and nested objects
- `Metadata()` method setting type name
- `Schema()` method defining all attributes
- `Configure()` method receiving provider data
- `Read()` method implementing API call and mapping

**Key Implementation Patterns:**
- Site parameter: `Optional: true, Computed: true`, defaults to 'svl'
- Type conversions: `types.StringValue()`, `types.Int64Value()`, etc.
- Nil handling: Always check pointers before accessing
- Nested objects: Use `types.ObjectValueFrom()` with proper `AttrTypes`
- Arrays: Use `types.ListValue()` with element type
- Initialize all fields with Null values, populate conditionally
- Always use modern Go syntax and features (i.e. `any` in-place of `interface{}`)
- Always write GoDoc compatible comments on _all_ public functions or methods. (They always start with a capital letter)
- **CRITICAL**: If API response includes `request_id`, the operation is asynchronous and MUST poll for completion using `pollRequestStatus()` helper

### 4. Generate Test File

Create `internal/provider/data_source_<name>_test.go` with:

**Test Function:**
- Name: `TestAccDataSource<Name>` (e.g., `TestAccDataSourceUser`, `TestAccDataSourceVMDetails`)
- Credential check: Skip if `FYRE_USERNAME` or `FYRE_API_KEY` not set

**External Resource Requirements:**
- Use environment variables prefixed with `FYRE_ACC_`
- Example: `FYRE_ACC_VM_ID` for VM identifiers
- Skip test if required environment variable not set
- **Never hardcode specific resource IDs**
- **Always scan existing tests for `FYRE_ACC_` variables to reuse**
- Where possible, only require a single VM. Use the API client for tests that need VM properties.
- Use `github.com/stretchr/testify/require` for test assertions

**Test Pattern:**
```go
vmID := os.Getenv("FYRE_ACC_VM_ID")
if vmID == "" {
    t.Skip("FYRE_ACC_VM_ID must be set for acceptance tests")
}
```

**Test Configuration:**
- Provider setup
- Data source configuration using environment variables
- Comprehensive attribute checks with `TestCheckResourceAttrSet`

### 5. Register Data Source

Update `internal/provider/provider.go`:
- Add `NewDataSource<Name>()` to `DataSources()` method
- Follow existing pattern
- Always order entries alphabetically

### 6. Test and Verify

```bash
# Run acceptance test
TF_ACC=1 go test -v ./internal/provider -run TestAccDataSource<Name>

# Generate documentation
make generate

# Optional: Test with enos (ALWAYS ask user first)
cd enos && enos scenario run fyre use:dev
```

### 7. Create Example Configuration

Create `examples/data-sources/<name>/main.tf`:

```hcl
terraform {
  required_providers {
    fyre = {
      source = "hashicorp-forge/fyre"
    }
  }
}

provider "fyre" {
  site = "svl"
}

data "fyre_<name>" "example" {
  # Required attributes
}

output "<name>_info" {
  value = data.fyre_<name>.example
}
```

### 8. Add to Enos Test Scenario

Update `enos/modules/datasources/main.tf`:
```hcl
data "fyre_<name>" "test" {
  # Test attributes
}

output "<name>" {
  value = data.fyre_<name>.test
}
```
- Ensure `data` and `output` blocks are ordered by their <name>

Update `enos/enos-scenario-fyre.hcl`:
```hcl
output "<name>" {
  value = step.test_datasources.<name>
}
```
- Ensure `output` blocks are ordered by their <name>

---

# Resource Workflow

## Workflow Steps

### 1. Prerequisites Check
- Verify `internal/client/client.gen.go` has the necessary CRUD operation methods:
  - `Create<Name>WithResponse` (POST operation)
  - `Get<Name>WithResponse` (GET operation)
  - `Update<Name>WithResponse` (PUT/PATCH operation)
  - `Delete<Name>WithResponse` (DELETE operation)
- Verify request and response types are complete
- **If incomplete**: Switch to `fyre-api-updater` mode first

### 2. Analyze API Structure
- Review OpenAPI spec in `internal/client/api.yaml`
- Identify all CRUD operation IDs
- Note request body schemas (for Create/Update)
- Note response schemas (for Read)
- Understand required vs optional fields
- Identify immutable fields (can't be updated)

### 3. Generate Resource Implementation

Create `internal/provider/resource_<name>.go` with:

**CRITICAL Naming Conventions:**
- **Function**: `NewResource<Name>()` (e.g., `NewResourceVM`, `NewResourceCluster`)
- **Struct**: `Resource<Name>` (e.g., `ResourceVM`, `ResourceCluster`)
- **Model**: `<Name>Model` (e.g., `VMModel`, `ClusterModel`) - NO Resource prefix
- **Nested Models**: `<NestedName>Model` (e.g., `ConfigModel`, `NetworkModel`) - NO Resource prefix

**Examples of Correct Naming:**
```go
// ✅ CORRECT
func NewResourceVM() resource.Resource { return &ResourceVM{} }
type ResourceVM struct { ... }
type VMModel struct { ... }
type NetworkModel struct { ... }  // nested model

// ❌ WRONG - Do NOT use these patterns
func NewVMResource() resource.Resource { ... }  // Wrong order
type VMResource struct { ... }  // Wrong order
type VMResourceModel struct { ... }  // Unnecessary suffix
type ResourceNetworkModel struct { ... }  // Unnecessary prefix
```

**Required Components:**
- Copyright header: `// Copyright IBM Corp. 2021, 2026`
- Resource struct implementing `resource.Resource` and `resource.ResourceWithImportState`
- Model structs for data and nested objects
- `Metadata()` method setting type name
- `Schema()` method defining all attributes with proper Required/Optional/Computed flags
- `Configure()` method receiving provider data
- `Create()` method: reads from `req.Plan`, calls API, writes to `resp.State`
- `Read()` method: reads from `req.State`, calls API, writes to `resp.State`
- `Update()` method: reads from `req.Plan`, calls API, writes to `resp.State`
- `Delete()` method: reads from `req.State`, calls API, no state write
- `ImportState()` method for resource import support

**Key Implementation Patterns:**

**Attribute Types:**
- `Required: true` - User must provide (e.g., name, required config)
- `Optional: true` - User can provide (e.g., description, tags)
- `Computed: true` - Provider sets (e.g., id, created_at)
- `Optional: true, Computed: true` - User can provide or API computes (e.g., site with default)

**Plan Modifiers:**
```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```
- Use `UseStateForUnknown()` for computed fields to prevent unnecessary replacements
- Use `RequiresReplace()` for immutable fields that force resource recreation

**CRUD Method Patterns:**

**Create:**
```go
func (r *Resource<Name>) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data <Name>Model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build API request from plan data
    createReq := client.Create<Name>JSONRequestBody{
        Field1: data.Field1.ValueString(),
        // ... map other fields
    }

    // Call API
    createResp, err := r.client.Create<Name>WithResponse(ctx, createReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create: %s", err))
        return
    }

    if createResp.StatusCode() != 201 {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d", createResp.StatusCode()))
        return
    }

    // Map response to state
    data.ID = types.StringValue(createResp.JSON201.ID)
    // ... map other fields

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

**Read:**
```go
func (r *Resource<Name>) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data <Name>Model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call API with ID from state
    readResp, err := r.client.Get<Name>WithResponse(ctx, data.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read: %s", err))
        return
    }

    if readResp.StatusCode() == 404 {
        // Resource no longer exists, remove from state
        resp.State.RemoveResource(ctx)
        return
    }

    if readResp.StatusCode() != 200 {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d", readResp.StatusCode()))
        return
    }

    // Update state with latest data
    // ... map response fields to data

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

**Update:**
```go
func (r *Resource<Name>) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data <Name>Model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build API request from plan data
    updateReq := client.Update<Name>JSONRequestBody{
        Field1: data.Field1.ValueString(),
        // ... map other fields
    }

    // Call API
    updateResp, err := r.client.Update<Name>WithResponse(ctx, data.ID.ValueString(), updateReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update: %s", err))
        return
    }

    if updateResp.StatusCode() != 200 {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d", updateResp.StatusCode()))
        return
    }

    // Map response to state
    // ... map response fields to data

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

**Delete:**
```go
func (r *Resource<Name>) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data <Name>Model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call API
    deleteResp, err := r.client.Delete<Name>WithResponse(ctx, data.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete: %s", err))
        return
    }

    if deleteResp.StatusCode() != 204 && deleteResp.StatusCode() != 200 {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d", deleteResp.StatusCode()))
        return
    }

    // No state write needed - resource is removed
}
```

**Import:**
```go
func (r *Resource<Name>) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### 4. Generate Test File

Create `internal/provider/resource_<name>_test.go` with:

**Test Function:**
- Name: `TestAccResource<Name>` (e.g., `TestAccResourceVM`, `TestAccResourceCluster`)
- Credential check: Skip if `FYRE_USERNAME` or `FYRE_API_KEY` not set
- Multiple test steps: Create, Update (if applicable), ImportState (optional)

**Test Pattern:**
```go
func TestAccResource<Name>(t *testing.T) {
    if os.Getenv("FYRE_USERNAME") == "" || os.Getenv("FYRE_API_KEY") == "" {
        t.Skip("FYRE_USERNAME and FYRE_API_KEY must be set for acceptance tests")
    }

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccResource<Name>Config("initial-value"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("fyre_<name>.test", "field1", "initial-value"),
                    resource.TestCheckResourceAttrSet("fyre_<name>.test", "id"),
                ),
            },
            // Update and Read testing (if resource supports updates)
            {
                Config: testAccResource<Name>Config("updated-value"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("fyre_<name>.test", "field1", "updated-value"),
                    resource.TestCheckResourceAttrSet("fyre_<name>.test", "id"),
                ),
            },
            // ImportState testing (optional but recommended)
            {
                ResourceName:      "fyre_<name>.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

func testAccResource<Name>Config(field1 string) string {
    return fmt.Sprintf(`
provider "fyre" {
  site = "svl"
}

resource "fyre_<name>" "test" {
  field1 = %[1]q
}
`, field1)
}
```

**Test Requirements:**
- Use `FYRE_ACC_*` environment variables for test data
- Use `TestCheckResourceAttr` for expected values
- Use `TestCheckResourceAttrSet` for computed values
- Test both create and update paths (if applicable)
- Include ImportState test when possible

### 5. Register Resource

Update `internal/provider/provider.go`:
- Add `NewResource<Name>()` to `Resources()` method
- Follow existing pattern
- Always order entries alphabetically

### 6. Test and Verify

```bash
# Run acceptance test
TF_ACC=1 go test -v ./internal/provider -run TestAccResource<Name>

# Generate documentation
make generate
```

### 7. Create Example Configuration

Create `examples/resources/<name>/main.tf`:

```hcl
terraform {
  required_providers {
    fyre = {
      source = "hashicorp-forge/fyre"
    }
  }
}

provider "fyre" {
  site = "svl"
}

resource "fyre_<name>" "example" {
  # Required attributes
  field1 = "value1"

  # Optional attributes
  field2 = "value2"
}

output "<name>_id" {
  value = fyre_<name>.example.id
}
```

### 8. Add to Enos Test Scenario

Create or update `enos/modules/resources/main.tf`:
```hcl
resource "fyre_<name>" "test" {
  # Test attributes using data sources where applicable
  # e.g., for VMs, use fyre_vm_os_available to get valid OS
}

output "<name>" {
  value = fyre_<name>.test
}
```
- Ensure `resource` and `output` blocks are ordered by their <name>

Update `enos/enos-scenario-fyre.hcl`:
```hcl
step "test_resources" {
  module = module.resources
}

output "<name>" {
  value = step.test_resources.<name>
}
```
- Ensure `output` blocks are ordered by their <name>

---

# Reference Files

## Data Sources
- **Style Guide**: `internal/provider/data_source_quota.go`
- **Test Guide**: `internal/provider/data_source_vm_details_test.go`

## Resources
- **Style Guide**: [HashiCorp Scaffolding Framework Example](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/internal/provider/example_resource.go)
- **Test Guide**: [HashiCorp Scaffolding Framework Test](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/internal/provider/example_resource_test.go)

## Common
- **Client Library**: `internal/client/client.gen.go`
- **OpenAPI Spec**: `internal/client/api.yaml`

---

# Common Implementation Patterns

## Nested Object Mapping
Create separate model structs and use `types.ObjectValueFrom()`:

```go
type DetailsModel struct {
    Field1 types.String `tfsdk:"field1"`
    Field2 types.Int64  `tfsdk:"field2"`
}

// In Read method
detailsModel := DetailsModel{
    Field1: types.StringNull(),
    Field2: types.Int64Null(),
}

if response.Details != nil {
    if response.Details.Field1 != nil {
        detailsModel.Field1 = types.StringValue(*response.Details.Field1)
    }
}

detailsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
    "field1": types.StringType,
    "field2": types.Int64Type,
}, detailsModel)
```

## List/Array Fields
Create slice of `attr.Value` and use `types.ListValue()`:

```go
if response.Items != nil && len(*response.Items) > 0 {
    itemsList := make([]attr.Value, 0, len(*response.Items))
    for _, item := range *response.Items {
        itemsList = append(itemsList, types.StringValue(item))
    }
    listValue, diags := types.ListValue(types.StringType, itemsList)
    resp.Diagnostics.Append(diags...)
    if !resp.Diagnostics.HasError() {
        data.Items = listValue
    }
}
```

## Optional Field Handling
Always initialize with Null, populate conditionally:

```go
data.Field = types.StringNull()
if response.Field != nil {
    data.Field = types.StringValue(*response.Field)
}
```

---

# Common Issues

## Client Library Missing Types
**Solution**: Use `fyre-api-updater` mode to update OpenAPI spec and regenerate client.

## Nested Object Mapping Errors
**Solution**: Ensure `AttrTypes` map matches model struct fields exactly.

## Test Environment Variables
**Solution**: Check existing tests for `FYRE_ACC_` variables before creating new ones.

## Resource Update Not Supported
**Solution**: If API doesn't support updates, implement Update() to return an error or use `RequiresReplace()` plan modifier on all fields.

---

# Tips

- **Data Sources**: Follow `data_source_quota.go` as style reference
- **Resources**: Follow HashiCorp scaffolding framework patterns
- Use `tflog.Debug()` for API response debugging
- Use `tflog.Trace()` for execution flow
- Handle all error cases with clear messages
- Ensure MarkdownDescription is helpful
- Test with real credentials before committing
- Documentation is auto-generated - don't create manually
- Always use modern Go syntax (e.g., `any` instead of `interface{}`)
- Write GoDoc comments for all public functions and methods
