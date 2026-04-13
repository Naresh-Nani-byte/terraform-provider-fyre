# Terraform Data Source Generator Skill

A Bob Shell custom mode for generating complete Terraform data sources for Fyre API resources.

## Quick Start

```bash
# Switch to the skill mode
bob> switch to tf-datasource-gen mode

# Generate a data source
bob> Create a data source for the stencil resource using the ListStencils operation
```

## What This Skill Does

This skill automates the complete workflow for creating a new Terraform data source:

1. ✅ Verifies the client library has necessary types
2. ✅ Generates the data source implementation file
3. ✅ Generates comprehensive test file
4. ✅ Registers the data source with the provider
5. ✅ Follows all Terraform and Go best practices

## Prerequisites

- The Fyre API endpoint must be defined in `internal/client/api.yaml`
- The client library must be generated with `make generate`
- If the client library is incomplete, use `fyre-api-updater` mode first

## Usage Examples

### Example 1: Simple Data Source (User)

```
Switch to tf-datasource-gen mode and create a data source for the user resource using the GetUserDetails operation at /user/{user_identifier}
```

This generates:
- `internal/provider/user_data_source.go`
- `internal/provider/user_data_source_test.go`
- Updates `internal/provider/provider.go`

### Example 2: List Data Source (Stencils)

```
Switch to tf-datasource-gen mode and create a data source for listing stencils using the ListStencils operation
```

This handles list responses with proper array mapping.

### Example 3: Complex Nested Data Source (Cluster Details)

```
Switch to tf-datasource-gen mode and create a data source for cluster details using GetClusterDetails at /cluster/{cluster_identifier}
```

This handles nested objects and complex response structures.

## Generated File Structure

```
internal/provider/
├── <name>_data_source.go       # Main implementation
├── <name>_data_source_test.go  # Acceptance tests
└── provider.go                  # Updated with new data source
```

## Key Features

### Automatic Site Parameter Handling
Most data sources get a `site` parameter that:
- Is optional and computed
- Defaults to provider's configured site or 'svl'
- Follows the established pattern

Confirm that site is in the OpenAPI spec in internal/client/api.yaml.

### Proper Null Handling
The skill generates code that:
- Initializes all fields with Null values
- Conditionally populates fields if data exists
- Handles nil pointers safely
- Make all fields Optional.

### Comprehensive Testing
Generated tests include:
- Credential checks with skip logic
- Provider configuration
- Verification of all computed attributes
- Follows acceptance testing best practices

### Idiomatic Go Code
All generated code:
- Follows terraform-plugin-framework patterns
- Uses proper naming conventions
- Includes comprehensive error handling
- Has helpful logging statements

## Testing Generated Data Sources

```bash
# Set credentials
export FYRE_USERNAME="your-username"
export FYRE_API_KEY="your-api-key"

# Run acceptance test
TF_ACC=1 go test -v ./internal/provider -run TestAcc<Name>DataSource

# Generate documentation
make generate
```

## Common Patterns

### Required Input Parameter
```go
"resource_id": schema.StringAttribute{
    MarkdownDescription: "The resource identifier",
    Required:            true,
},
```

### Optional Site Parameter
```go
"site": schema.StringAttribute{
    MarkdownDescription: "Site location (svl or rtp). Defaults to 'svl' or inherits from provider configuration.",
    Optional:            true,
    Computed:            true,
},
```

### Computed Output Field
```go
"name": schema.StringAttribute{
    MarkdownDescription: "The resource name",
    Computed:            true,
    Optional:            true,
},
```

### List of Strings
```go
"platforms": schema.ListAttribute{
    MarkdownDescription: "List of available platforms",
    Computed:            true,
    Optional:            true,
    ElementType:         types.StringType,
},
```

### Nested Object
```go
"details": schema.SingleNestedAttribute{
    MarkdownDescription: "Detailed information",
    Computed:            true,
    Attributes: map[string]schema.Attribute{
        "field1": schema.StringAttribute{
            Computed: true,
            Optional: true,
        },
    },
},
```

## Troubleshooting

### Client Library Missing Types

**Problem**: The skill reports that the client library doesn't have the necessary types.

**Solution**:
```bash
# Switch to the API updater mode
bob> switch to fyre-api-updater mode

# Update the client library for the endpoint
bob> Update the client library for the /user/{user_identifier} endpoint
```

### Test Failures

**Problem**: Acceptance tests fail with authentication errors.

**Solution**: Ensure `FYRE_USERNAME` and `FYRE_API_KEY` environment variables are set correctly.

**Problem**: Tests fail with "attribute not found" errors.

**Solution**: The API response structure may have changed. Use `fyre-api-updater` mode to verify and update the OpenAPI spec.

### Compilation Errors

**Problem**: Generated code doesn't compile.

**Solution**: Check that:
1. The client library is up to date (`make generate`)
2. All imports are correct
3. Type conversions match the client library types

## Integration with Other Skills

### With fyre-api-updater

When the client library is incomplete:
1. Use `fyre-api-updater` to update the OpenAPI spec
2. Run `make generate` to regenerate the client
3. Then use `tf-datasource-gen` to create the data source

### Workflow Example

```bash
# Step 1: Update API spec if needed
bob> switch to fyre-api-updater mode
bob> Update the client library for /stencil/list endpoint

# Step 2: Generate data source
bob> switch to tf-datasource-gen mode
bob> Create a data source for listing stencils

# Step 3: Test
bob> switch to code mode
bob> Run the acceptance test for the stencils data source
```

## Reference Files

- **Style Guide**: `internal/provider/quota_data_source.go`
- **Test Guide**: `internal/provider/quota_data_source_test.go`
- **Client Library**: `internal/client/client.gen.go`
- **OpenAPI Spec**: `internal/client/api.yaml`
- **Workflow**: `.bob/rules-tf-datasource-gen/workflow.md`
- **Examples**: `.bob/rules-tf-datasource-gen/example-template.md`

## Best Practices

1. **Always verify client library first** - Check that the API operation exists and has complete types
2. **Follow the quota pattern** - Use it as the reference for style and structure
3. **Test with real credentials** - Always run acceptance tests before committing
4. **Handle all nil cases** - Initialize with Null, populate conditionally
5. **Add helpful descriptions** - MarkdownDescription should be clear and informative
6. **Use proper logging** - Add tflog.Debug for API responses, tflog.Trace for flow

## Contributing

When enhancing this skill:
1. Update the custom mode configuration in `.bob/custom_modes.yaml`
2. Update workflow documentation in `.bob/rules-tf-datasource-gen/workflow.md`
3. Add examples to `.bob/rules-tf-datasource-gen/example-template.md`
4. Test with multiple data source types (simple, list, nested)

## Support

For issues or questions:
1. Check the example template for common patterns
2. Review the quota data source implementation
3. Consult the workflow documentation
4. Use the fyre-api-updater skill if client library issues arise
