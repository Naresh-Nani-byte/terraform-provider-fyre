# Terraform Data Source Generator

A Bob Shell mode that generates complete Terraform data sources for Fyre API resources.

## Quick Start

```bash
# Switch to the mode
bob> switch to tf-datasource-gen mode

# Generate a data source
bob> Create a data source for the user resource using the GetUserDetails operation
```

## What It Does

Automates the complete workflow for creating a Terraform data source:

- ✅ Verifies client library has necessary types
- ✅ Generates data source implementation
- ✅ Generates comprehensive tests
- ✅ Registers with provider
- ✅ Creates example configuration
- ✅ Follows Terraform and Go best practices

## Prerequisites

- Fyre API endpoint defined in `internal/client/api.yaml`
- Client library generated with `make generate`
- If client library incomplete, use `fyre-api-updater` mode first

## Usage Examples

### Simple Data Source
```
Create a data source for the user resource using GetUserDetails at /user/{user_identifier}
```

### List Data Source
```
Create a data source for listing stencils using the ListStencils operation
```

### Complex Nested Data Source
```
Create a data source for cluster details using GetClusterDetails at /cluster/{cluster_identifier}
```

## Testing

```bash
# Set credentials
export FYRE_USERNAME="your-username"
export FYRE_API_KEY="your-api-key"

# Set any specific prerequisite identifiers (optional)
export FYRE_ACC_VM_ID=v1-8103661

# Run acceptance test
TF_ACC=1 go test -v ./internal/provider -run TestAccDataSource<Name>

# Generate documentation
make generate
```

## Troubleshooting

### Client Library Missing Types
**Solution**: Switch to `fyre-api-updater` mode to update the OpenAPI spec, then regenerate the client with `make generate`.

### Test Failures
- **Authentication errors**: Verify `FYRE_USERNAME` and `FYRE_API_KEY` are set correctly
- **Attribute not found**: API response may have changed - use `fyre-api-updater` mode to update the spec
- **Compilation errors**: Ensure client library is up to date with `make generate`

## Integration with fyre-api-updater

When the client library is incomplete:

```bash
# Step 1: Update API spec
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

- **Style Guide**: `internal/provider/data_source_quota.go`
- **Test Guide**: `internal/provider/data_source_quota_test.go`
- **Example**: `.bob/rules-tf-datasource-gen/example-template.md`
jjjjjjjjjjjjjjjkkkkkkjrules-tf-datasource-gen/workflow.md`
