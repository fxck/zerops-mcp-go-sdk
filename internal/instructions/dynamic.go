package instructions

import (
	"fmt"
	"os"
	"strings"
)

// GetDynamicInstructions returns client-specific instructions
func GetDynamicInstructions(clientName, clientVersion string) string {
	// Detect the AI model based on client name
	var modelType string
	switch {
	case strings.Contains(strings.ToLower(clientName), "claude"):
		modelType = "claude"
	case strings.Contains(strings.ToLower(clientName), "chatgpt") || strings.Contains(strings.ToLower(clientName), "openai"):
		modelType = "chatgpt"
	case strings.Contains(strings.ToLower(clientName), "gemini") || strings.Contains(strings.ToLower(clientName), "google"):
		modelType = "gemini"
	case strings.Contains(strings.ToLower(clientName), "cursor"):
		modelType = "cursor"
	case strings.Contains(strings.ToLower(clientName), "copilot"):
		modelType = "copilot"
	default:
		modelType = "generic"
	}

	// Always log what was detected
	fmt.Fprintf(os.Stderr, "\n=== INSTRUCTION CUSTOMIZATION ===\n")
	if clientName != "" {
		fmt.Fprintf(os.Stderr, "Client Name: %s\n", clientName)
		fmt.Fprintf(os.Stderr, "Client Version: %s\n", clientVersion)
		fmt.Fprintf(os.Stderr, "Detected Model Type: %s\n", modelType)
		fmt.Fprintf(os.Stderr, "Instructions: Customized for %s\n", modelType)
	} else {
		fmt.Fprintf(os.Stderr, "Client Name: Not provided\n")
		fmt.Fprintf(os.Stderr, "Detected Model Type: generic\n")
		fmt.Fprintf(os.Stderr, "Instructions: Using generic instructions\n")
	}
	fmt.Fprintf(os.Stderr, "=================================\n\n")

	// Base instructions
	base := `
# Zerops MCP Instructions

## CRITICAL: Always Start with Knowledge Base
BEFORE creating anything, search the knowledge base for proven configurations:
- knowledge_search("technology stack") - Find recipes and patterns
- knowledge_get("recipe/name") - Get complete configurations
- 159+ templates available (Laravel, Django, Next.js, etc.)
`

	// Model-specific instructions
	var specific string
	switch modelType {
	case "claude":
		specific = `
## Optimized for Claude
You're excellent at understanding complex requirements and finding patterns.
- Use knowledge_search extensively to find the best solutions
- Your analytical skills are perfect for debugging Zerops deployments
- Feel free to explain architectural decisions when relevant
- Trust your code generation abilities - they work well with Zerops YAML
`

	case "chatgpt":
		specific = `
## Optimized for ChatGPT
Your step-by-step approach works perfectly with Zerops workflows.
- Break deployments into clear, sequential steps
- Always verify service types with knowledge_get before using them
- Use your systematic approach for troubleshooting
- Double-check YAML syntax as Zerops requires precise formatting
`

	case "gemini":
		specific = `
## Optimized for Gemini
Your cloud expertise aligns well with Zerops architecture.
- Leverage your understanding of distributed systems
- Consider scalability patterns (HA vs NON_HA modes)
- Use your optimization skills for resource configuration
- Apply Google Cloud best practices where applicable
`

	case "cursor":
		specific = `
## Optimized for Cursor
Your code-first approach works great with Zerops.
- Use knowledge_search to find code examples quickly
- Zerops YAML files are in the project root (zerops.yml, zerops-project-import.yml)
- Autocomplete works well with service type strings from knowledge base
- Use project_import for infrastructure-as-code workflows
`

	case "copilot":
		specific = `
## Optimized for GitHub Copilot
Your context awareness helps with Zerops configurations.
- Check existing zerops.yml files in the repository
- Use knowledge_get for service configuration templates
- Zerops integrates well with GitHub Actions for CI/CD
- Follow repository patterns for consistency
`

	default:
		specific = `
## General AI Assistant Guidelines
- Always verify service types with the knowledge base
- Use knowledge_search and knowledge_get before creating services
- Follow the examples in the knowledge base closely
- Ask for clarification if requirements are unclear
`
	}

	// Common workflow instructions
	common := `
## Optimal Workflow

1. **Search Knowledge First**
   knowledge_search("python postgresql") - Find matching recipes
   knowledge_get("recipe/python") - Get full YAML configuration

2. **Create Project Structure**
   - zerops-project-import.yml - Infrastructure definition
   - zerops.yml - Deployment configuration
   IMPORTANT: Setup name in zerops.yml MUST match service hostname

3. **Execute in Order**
   project_create(name, region) - Create project
   project_import(project_id, yaml) - Import all services
   deploy_push(project_id, service_id) - Deploy code

## Key Requirements

- Git: Must have at least one commit before deploying
- Service IDs: Use the 22-char ID from service_list, not the name
- Hostnames: Alphanumeric only (no hyphens)
- Recipe services: MUST use standard names (adminer, mailpit, s3browser)

## Common Service Types

Always verify with KB, but common patterns:
- Databases: postgresql@16, mysql@8, mongodb@7
- Runtimes: nodejs@20, python@3.11, php@8.3
- Cache/Queue: redis@7, rabbitmq@3.12, keydb@6

## Error Recovery

serviceStackTypeNotFound error:
1. Use knowledge_search to find correct type
2. Check hostname has no special characters
3. Verify mode is "HA" or "NON_HA" for databases

Remember: Knowledge base has 159+ working recipes - use them!`

	return base + specific + common
}