# Claude MCP Tools - Improvement Recommendations

## Analysis Summary
After comparing the working v0.1.4 with the current version and analyzing the knowledge base, here are the key findings and recommendations to help Claude work better with fewer trial/error attempts.

## Key Issues Identified

### 1. **Service Type Discovery Problem**
**Current Issue:** Claude has to guess service types (e.g., `mongodb@7` vs `mongodb@7.0`) leading to multiple failures before success.

**Evidence from your session:**
- Failed: `type: mongodb@7` 
- Failed: `type: valkey@7`
- Success: Only after simplifying to single service imports

**Solution:** The knowledge base already contains all valid service types in `/knowledge/data/services/`. We need to expose this better.

### 2. **YAML Structure Validation**
**Current Issue:** No pre-validation of YAML structure before sending to API, resulting in cryptic errors like `[400][serviceStackTypeNotFound]`.

**The knowledge base shows correct format:**
```yaml
services:
  - hostname: postgresql16  # Note: no special chars in hostname
    type: postgresql@16
    mode: NON_HA
```

### 3. **Missing Recipe/Pattern System**
The knowledge base contains 100+ ready-to-use patterns (Laravel, Django, Next.js, etc.) but these aren't being utilized effectively.

## Recommended Improvements

### 1. Add Service Type Listing Tool
```go
// New tool: service_types_list
func handleServiceTypesList(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
    // Return all valid service types with versions
    return shared.TextResponse(`Available Service Types:

Databases:
- postgresql@16, postgresql@15, postgresql@14 (modes: HA, NON_HA)
- mariadb@11, mariadb@10.6 (modes: HA, NON_HA)
- mongodb@7, mongodb@6 (modes: HA, NON_HA)
- valkey@7 (modes: HA, NON_HA)
- keydb@6 (modes: HA, NON_HA)

Runtimes:
- php@8.3, php@8.2 (NOT php-apache or php-nginx!)
- nodejs@20, nodejs@18
- python@3.11, python@3.10
- go@1, go@1.21
...`), nil
}
```

### 2. Add YAML Validation Tool
```go
// New tool: yaml_validate
func handleYAMLValidate(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
    yamlContent := args["yaml"].(string)
    
    // Parse and validate structure
    var config ZeropsConfig
    err := yaml.Unmarshal([]byte(yamlContent), &config)
    
    // Check service types against known list
    for _, service := range config.Services {
        if !isValidServiceType(service.Type) {
            return shared.ErrorResponse(fmt.Sprintf(
                "Invalid service type '%s'. Use 'service_types_list' to see valid types.\n" +
                "Did you mean: %s?", 
                service.Type, 
                suggestSimilar(service.Type)
            )), nil
        }
    }
    
    return shared.TextResponse("YAML is valid and ready for import"), nil
}
```

### 3. Add Recipe Import Tool
```go
// New tool: recipe_import
func handleRecipeImport(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
    recipeName := args["recipe"].(string)  // e.g., "laravel-minimal"
    projectID := args["project_id"].(string)
    
    // Load recipe from knowledge base
    recipe := loadRecipe(recipeName)
    
    // Generate YAML from recipe
    yaml := generateYAMLFromRecipe(recipe)
    
    // Import to project
    return handleProjectImport(ctx, client, map[string]interface{}{
        "project_id": projectID,
        "yaml": yaml,
    })
}
```

### 4. Improve Error Messages
```go
// Before:
"Failed to parse response: [400][serviceStackTypeNotFound] Service stack Type not found."

// After:
"Service type not found.

PROBLEM: 'php-apache@8.3' is not a valid service type
SOLUTION: php (use php@8.3 with appropriate run configuration)

To fix:
1. Use knowledge_search('php') to find the correct service
2. Use knowledge_get('services/php') to get exact type string
3. Common services:
   - PHP: use 'php@8.3' (NOT php-apache or php-nginx)
   - PostgreSQL: use 'postgresql@16' (NOT postgres)
   - Redis-compatible: use 'valkey@7' (NOT redis)"
```

### 5. Add Tool Instructions in Description
```go
shared.GlobalRegistry.Register(&shared.ToolDefinition{
    Name: "project_import",
    Description: `Import services using YAML. 

IMPORTANT: Before using this tool:
1. Use 'service_types_list' to verify service types
2. Use 'yaml_validate' to check your YAML
3. Hostnames must be alphanumeric (no hyphens/underscores)

Example that works:
services:
  - hostname: postgresql16
    type: postgresql@16
    mode: NON_HA`,
    // ...
})
```

### 6. Add Common Patterns as Examples
```go
// In knowledge_search response, include working examples:
"Found recipe: laravel-minimal
Ready-to-use YAML:
```yaml
services:
  - hostname: app
    type: php-apache@8.3
    enableSubdomainAccess: true
  - hostname: db
    type: postgresql@16
    mode: NON_HA
```
Use this directly with project_import tool."
```

## Implementation Priority

1. **High Priority** (Immediate impact on reducing errors):
   - Add `service_types_list` tool
   - Improve error messages with specific guidance
   - Add validation before API calls

2. **Medium Priority** (Better UX):
   - Add `yaml_validate` tool
   - Include working examples in tool descriptions
   - Add `recipe_import` for common stacks

3. **Low Priority** (Nice to have):
   - Auto-suggest corrections for common typos
   - Batch import capabilities
   - Service dependency checking

## Example Improved Workflow

**Current (many failures):**
```
1. Try import with guessed types → Fail
2. Try different format → Fail
3. Check knowledge → Partial help
4. Try simpler format → Fail
5. Try one service → Success
6. Add services one by one → Mixed results
```

**Improved (fewer failures):**
```
1. service_types_list → See all valid types
2. recipe_search "laravel" → Get working template
3. yaml_validate → Confirm it's correct
4. project_import → Success first time
```

## Testing Recommendations

1. Create integration tests with common failure scenarios
2. Add examples of working YAML for each service type
3. Document the exact API requirements in tool descriptions
4. Add a "dry-run" mode for import validation

## Conclusion

The main issue is that Claude lacks visibility into:
1. Valid service types and versions
2. Correct YAML structure requirements
3. Working examples and patterns

By adding discovery tools (`service_types_list`), validation tools (`yaml_validate`), and better error messages, we can reduce trial-and-error attempts from 5-10 down to 1-2.

The knowledge base already contains all the necessary information - we just need to make it more accessible through dedicated tools and better guidance.