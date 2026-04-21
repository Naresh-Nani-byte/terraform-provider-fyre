# Terraform Generator (tf-gen)

A comprehensive Bob Shell skill for generating Terraform provider components for the Fyre API.

## Overview

The `tf-gen` skill automates the creation of complete, production-ready Terraform provider components following HashiCorp's best practices and the terraform-plugin-framework patterns. It supports both:

- **Data Sources**: Read-only resources for querying Fyre API data
- **Resources**: Full CRUD lifecycle management for Fyre infrastructure

## When to Use This Skill

### Use for Data Sources When:
- You need to query existing Fyre resources (VMs, clusters, users, etc.)
- The operation is read-only (no modifications)
- You want to reference existing infrastructure in Terraform configurations
- You need to fetch configuration data or metadata

**Examples:**
- Fetching user details
- Listing available VM platforms
- Getting cluster information
- Checking VM status

### Use for Resources When:
- You need to create, update, or delete Fyre infrastructure
- You want full lifecycle management through Terraform
- You need to manage resource state over time
- You want to support resource import

**Examples:**
- Creating and managing VMs
- Managing clusters
- Configuring network settings
- Managing user permissions

## Quick Start

### Creating a Data Source

```
Switch to tf-gen mode and create a data source for the user resource using the GetUserDetails operation
```

### Creating a Resource

```
Switch to tf-gen mode and create a resource for managing VMs using the CreateVM, GetVM, UpdateVM, and DeleteVM operations
```

## What This Skill Does

### For Data Sources:
1. ✅ Verifies client library has necessary types
2. ✅ Generates `data_source_<name>.go` with complete implementation
3. ✅ Generates `data_source_<name>_test.go` with acceptance tests
4. ✅ Registers data source in `provider.go`
5. ✅ Creates example Terraform configuration
6. ✅ Adds to Enos test scenarios (optional)

### For Resources:
1. ✅ Verifies client library has CRUD operation methods
2. ✅ Generates `resource_<name>.go` with full CRUD implementation
3. ✅ Generates `resource_<name>_test.go` with multi-step tests
4. ✅ Implements import support
5. ✅ Registers resource in `provider.go`
6. ✅ Creates example Terraform configuration
7. ✅ Adds proper plan modifiers and state management

## Prerequisites

Before using this skill, ensure:

1. **Client Library is Complete**: The OpenAPI spec must be accurate and the client generated
   - If not, use `fyre-api-updater` mode first to update the spec
   - Run `make generate` to regenerate the client

2. **API Operations Exist**:
   - For data sources: GET operation method (e.g., `GetUserDetailsWithResponse`)
   - For resources: CREATE, GET, UPDATE, DELETE operation methods

3. **Test Environment**:
   - `FYRE_USERNAME` and `FYRE_API_KEY` environment variables set
   - Access to Fyre API for acceptance testing

## File Structure

```
internal/provider/
├── data_source_<name>.go       # Data source implementation
├── data_source_<name>_test.go  # Data source tests
├── resource_<name>.go          # Resource implementation
├── resource_<name>_test.go     # Resource tests
└── provider.go                 # Registration

examples/
├── data-sources/<name>/main.tf # Data source example
└── resources/<name>/main.tf    # Resource example

enos/
├── modules/datasources/main.tf # Enos data source tests
└── enos-scenario-fyre.hcl      # Enos scenario config
```

## Key Differences

| Aspect | Data Source | Resource |
|--------|-------------|----------|
| **File Prefix** | `data_source_` | `resource_` |
| **Struct Name** | `DataSource<Name>` | `Resource<Name>` |
| **Methods** | Read only | Create, Read, Update, Delete |
| **State Source** | Config | Plan (Create/Update), State (Read/Delete) |
| **Attributes** | Mostly Computed | Required, Optional, Computed |
| **Plan Modifiers** | Rare | Common (UseStateForUnknown) |
| **Import Support** | N/A | ImportState method |
| **Testing** | Single step | Multiple steps (Create, Update, Import) |
| **Registration** | DataSources() | Resources() |

## Naming Conventions

### Data Sources
```go
// Function
func NewDataSourceUser() datasource.DataSource { ... }

// Struct
type DataSourceUser struct { ... }

// Model
type UserModel struct { ... }

// Nested Model
type DevelopmentModel struct { ... }
```

### Resources
```go
// Function
func NewResourceVM() resource.Resource { ... }

// Struct
type ResourceVM struct { ... }

// Model
type VMModel struct { ... }

// Nested Model
type NetworkModel struct { ... }
```

### Terraform Names
- Data source: `fyre_<name>` (e.g., `fyre_user`, `fyre_vm_details`)
- Resource: `fyre_<name>` (e.g., `fyre_vm`, `fyre_cluster`)

## Common Patterns

### Attribute Naming
Always check existing implementations for consistency:
- VM identifier: `vm_id` (NOT `vm_identifier`)
- User identifier: `user_id` (NOT `user_identifier`)
- Cluster identifier: `cluster_id` (NOT `cluster_identifier`)

### Site Parameter
Both data sources and resources include a site parameter:
```go
"site": schema.StringAttribute{
    MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
    Optional:            true,
    Computed:            true,
}
```

### Nil Handling
Always initialize with Null, populate conditionally:
```go
data.Field = types.StringNull()
if response.Field != nil {
    data.Field = types.StringValue(*response.Field)
}
```

### Nested Objects
Use `types.ObjectValueFrom()` with proper `AttrTypes`:
```go
detailsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
    "field1": types.StringType,
    "field2": types.Int64Type,
}, detailsModel)
```

### Lists/Arrays
Use `types.ListValue()` with element type:
```go
listValue, diags := types.ListValue(types.StringType, itemsList)
```

## Testing

### Data Source Tests
```bash
TF_ACC=1 go test -v ./internal/provider -run TestAccDataSource<Name>
```

- Single test step
- Use `TestCheckResourceAttrSet` for computed values
- Use `FYRE_ACC_*` environment variables for test data

### Resource Tests
```bash
TF_ACC=1 go test -v ./internal/provider -run TestAccResource<Name>
```

- Multiple test steps (Create, Update, Import)
- Use `TestCheckResourceAttr` for expected values
- Use `TestCheckResourceAttrSet` for computed values
- Test state transitions

## Documentation

Documentation is auto-generated:
```bash
make generate
```

Do NOT manually create documentation files. The generator creates:
- `docs/data-sources/<name>.md`
- `docs/resources/<name>.md`

## Reference Files

### Data Sources
- **Style Guide**: `internal/provider/data_source_quota.go`
- **Test Guide**: `internal/provider/data_source_vm_details_test.go`
- **Example Template**: `datasource-example.md`

### Resources
- **Style Guide**: [HashiCorp Scaffolding Framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/internal/provider/example_resource.go)
- **Example Template**: `resource-example.md`

### Common
- **Workflow**: `workflow.md`
- **Client Library**: `internal/client/client.gen.go`
- **OpenAPI Spec**: `internal/client/api.yaml`

## Troubleshooting

### Client Library Missing Types
**Problem**: Generated code references types that don't exist in `client.gen.go`

**Solution**: 
1. Switch to `fyre-api-updater` mode
2. Update the OpenAPI spec with actual API responses
3. Run `make generate` to regenerate the client
4. Return to `tf-gen` mode

### Test Environment Variables
**Problem**: Tests skip due to missing environment variables

**Solution**:
```bash
export FYRE_USERNAME="your-username"
export FYRE_API_KEY="your-api-key"
export FYRE_ACC_VM_ID="existing-vm-id"  # For tests requiring existing resources
```

### Nested Object Mapping Errors
**Problem**: Type mismatch errors when mapping nested objects

**Solution**: Ensure `AttrTypes` map matches model struct fields exactly:
```go
// Model
type DetailsModel struct {
    Field1 types.String `tfsdk:"field1"`
    Field2 types.Int64  `tfsdk:"field2"`
}

// AttrTypes must match exactly
map[string]attr.Type{
    "field1": types.StringType,
    "field2": types.Int64Type,
}
```

### Resource Update Not Supported
**Problem**: API doesn't support update operations

**Solution**: Either:
1. Return error in Update method
2. Use `RequiresReplace()` plan modifier on all mutable fields

## Best Practices

1. **Follow Existing Patterns**: Check existing data sources and resources for consistency
2. **Use Modern Go**: Use `any` instead of `interface{}`, modern error handling
3. **Write GoDoc Comments**: All public functions and methods need documentation
4. **Handle Errors Gracefully**: Provide clear, actionable error messages
5. **Log Appropriately**: Use `tflog.Debug()` for API responses, `tflog.Trace()` for flow
6. **Test Thoroughly**: Run acceptance tests before committing
7. **Check Attribute Names**: Ensure consistency across all data sources/resources
8. **Initialize with Null**: Always initialize optional fields with Null values
9. **Use Plan Modifiers**: Prevent unnecessary resource replacements
10. **Support Import**: Always implement ImportState for resources

## Related Skills

- **fyre-api-updater**: Update OpenAPI specs and regenerate client library
- **code**: General code editing and refactoring
- **ask**: Search Bob Shell documentation

## Support

For questions about:
- **This skill**: Use `ask` mode and search Bob Shell documentation
- **Terraform patterns**: Refer to HashiCorp's terraform-plugin-framework documentation
- **Fyre API**: Check the OpenAPI spec at `internal/client/api.yaml`

## Version History

- **v2.0** (2026-04-14): Added resource generation support, renamed from tf-datasource-gen
- **v1.0** (2021): Initial release with data source generation only