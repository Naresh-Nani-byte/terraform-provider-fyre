# Fyre API Update Workflow

## Step-by-Step Process

### 1. Prepare Environment
Ensure environment variables are set:
- `FYRE_USERNAME`: Your Fyre username
- `FYRE_API_KEY`: Your Fyre API key

### 2. Make API Request
```bash
curl -s -X GET -u "$FYRE_USERNAME":"$FYRE_API_KEY" \
  -H 'Content-type: application/json' \
  https://ocpapi.svl.ibm.com/v1/<route>
```

### 3. Analyze Response Structure
Look for:
- Top-level structure (object, array, etc.)
- Field names and types
- Nested structures
- Optional fields (may not appear in all responses)
- Enum values
- **CRITICAL**: Check for `request_id` field in responses - if present, the operation is asynchronous and MUST be polled for completion

### 4. Update OpenAPI Spec
In `internal/client/api.yaml`:
- Find the endpoint definition (paths are alphabetically ordered)
- Locate the response schema reference
- Update the schema in `components/schemas`
- Add a comment: `# Updated from actual API response on YYYY-MM-DD`

**Organization Guidelines:**
- **CRITICAL**: Keep paths/routes alphabetically organized in the `paths:` section
  - When adding new routes, insert them in alphabetical order
  - When searching for existing routes, use alphabetical ordering to locate them quickly
  - More specific paths come after general ones (e.g., `/user` before `/user/api_key`)
- Keep schemas alphabetically organized within their section
- Maintain consistent formatting (2-space indentation)
- Follow OpenAPI 3.0 specification strictly
- Place new schemas in alphabetical order within their category
- Use consistent naming: PascalCase for schema names
- Group related schemas with section comments (e.g., `# Cluster Schemas`)

### 5. Regenerate Client
```bash
make generate
```

### 6. Verify Changes
Check `internal/client/client.gen.go` for:
- Correct struct definitions
- Proper field types
- JSON tags matching response fields

## Common Response Patterns

### Success Response
```json
{
  "status": "success",
  "data": { ... }
}
```

### Error Response
```json
{
  "status": "error",
  "message": "Error description"
}
```

## Tips
- Use `jq` for pretty-printing: `curl ... | jq`
- Save responses for reference: `curl ... > response.json`
- Compare multiple responses to identify optional fields
- Check for pagination in list endpoints
