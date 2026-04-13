# Terraform Data Source Generator Workflow

This skill generates complete Terraform data sources for Fyre API resources.

## Usage

Switch to the `tf-datasource-gen` mode and provide:
1. The resource name (e.g., "user", "stencil", "cluster")
2. The API route/operation (e.g., "/user/{user_identifier}", "GetUserDetails")

Example:
```
Switch to tf-datasource-gen mode and create a data source for the user resource using the /user/{user_identifier} endpoint
```

## Workflow Steps

### 1. Prerequisites Check
- Verify the client library has the necessary types in `internal/client/client.gen.go`
- Check for the API operation method (e.g., `GetUserDetailsWithResponse`)
- Verify response types are complete and accurate
- **If incomplete**: Switch to `fyre-api-updater` mode first to update the client library

### 2. Analyze API Structure
- Review the OpenAPI spec in `internal/client/api.yaml`
- Identify the operation ID and response schema
- Note required vs optional parameters
- Understand the response structure (nested objects, arrays, etc.)

### 3. Generate Data Source Implementation
Create `internal/provider/data_source_<name>.go` with:
- Proper copyright header
- Data source struct implementing `datasource.DataSource`
- Model structs for the data and nested objects
- `Metadata()` method setting the type name
- `Schema()` method defining all attributes
- `Configure()` method receiving provider data
- `Read()` method implementing the API call and mapping

Key patterns:
- Site parameter: `Optional: true, Computed: true` with default to 'svl'
- Use `types.StringValue()`, `types.Int64Value()`, etc. for conversions
- Handle nil pointers safely
- Use `types.ObjectValueFrom()` for nested objects
- Use `types.ListValue()` for arrays
- Initialize with Null values, populate if data exists

### 4. Generate Test File
Create `internal/provider/data_source_<name>_test.go` with:
- `TestAccDataSource<Name>` function
- Credential check and skip logic (FYRE_USERNAME, FYRE_API_KEY)
- **External resource requirements**: Use environment variables for test data
  - For VM identifiers: Use `FYRE_TEST_VM_ID` environment variable
  - Skip test if required environment variable is not set
  - Never hardcode specific resource IDs that may not exist
- Test configuration with provider setup
- Comprehensive attribute checks using `TestCheckResourceAttrSet`

**Test Pattern for External Resources:**
```go
vmID := os.Getenv("FYRE_TEST_VM_ID")
if vmID == "" {
    t.Skip("FYRE_TEST_VM_ID must be set for acceptance tests")
}
```

### 5. Register Data Source
Update `internal/provider/provider.go`:
- Add `NewDataSource<Name>()` to the `DataSources()` method
- Follow the existing pattern

### 6. Test and Verify
```bash
# Run the acceptance test
go test -v ./internal/provider -run TestAccDataSource<Name>

# Generate documentation
make generate
```

## Example: User Data Source

For a user data source using `/user/{user_identifier}`:

1. Check client has `GetUserDetailsWithResponse` method
2. Create `data_source_user.go` with:
   - `DataSourceUser` struct
   - `DataSourceUserModel` with fields from API response
   - Schema with `user_identifier` (required) and `site` (optional)
   - Read method calling `GetUserDetailsWithResponse`
3. Create `data_source_user_test.go` with test cases
4. Register in provider.go
5. Test with `go test -v ./internal/provider -run TestAccDataSourceUser`

## Reference Files

- Style reference: `internal/provider/data_source_quota.go`
- Test reference: `internal/provider/data_source_user_test.go`
- Client library: `internal/client/client.gen.go`
- OpenAPI spec: `internal/client/api.yaml`

## Common Issues

### Client Library Missing Types
**Solution**: Use `fyre-api-updater` mode to update the OpenAPI spec and regenerate the client.

### Nested Object Mapping
**Solution**: Create separate model structs for nested objects and use `types.ObjectValueFrom()` with proper `AttrTypes`.

### Optional Fields
**Solution**: Initialize with `types.<Type>Null()`, then conditionally set with `types.<Type>Value()` if data exists.

### List/Array Fields
**Solution**: Create a slice of `attr.Value`, populate it, then use `types.ListValue()` with the element type.

## Tips

- Always follow the quota data source as a style reference
- Use `tflog.Debug()` for debugging API responses
- Handle all error cases properly
- Ensure MarkdownDescription is clear and helpful
- Test with real credentials before committing
- Documentation is auto-generated - don't create docs manually
