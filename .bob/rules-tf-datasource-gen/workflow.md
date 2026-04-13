# Terraform Data Source Generator Workflow

This mode generates complete Terraform data sources for Fyre API resources.

## Usage

Provide the resource name and API operation:

```
Create a data source for the user resource using the GetUserDetails operation
```

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
- Always write GoDoc compatible comments on _all_ public funtions or methods. (They always start with a capital letter)

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

Update `enos/enos-scenario-fyre.hcl`:
```hcl
output "<name>" {
  value = step.test_datasources.<name>
}
```

## Reference Files

- **Style Guide**: `internal/provider/data_source_quota.go`
- **Test Guide**: `internal/provider/data_source_vm_details_test.go`
- **Client Library**: `internal/client/client.gen.go`
- **OpenAPI Spec**: `internal/client/api.yaml`

## Common Implementation Patterns

### Nested Object Mapping
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

### List/Array Fields
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

### Optional Field Handling
Always initialize with Null, populate conditionally:

```go
data.Field = types.StringNull()
if response.Field != nil {
    data.Field = types.StringValue(*response.Field)
}
```

## Common Issues

### Client Library Missing Types
**Solution**: Use `fyre-api-updater` mode to update OpenAPI spec and regenerate client.

### Nested Object Mapping Errors
**Solution**: Ensure `AttrTypes` map matches model struct fields exactly.

### Test Environment Variables
**Solution**: Check existing tests for `FYRE_ACC_` variables before creating new ones.

## Tips

- Follow `data_source_quota.go` as style reference
- Use `tflog.Debug()` for API response debugging
- Use `tflog.Trace()` for execution flow
- Handle all error cases with clear messages
- Ensure MarkdownDescription is helpful
- Test with real credentials before committing
- Documentation is auto-generated - don't create manually
