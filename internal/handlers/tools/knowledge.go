package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterKnowledge registers knowledge base tools in the global registry
func RegisterKnowledge() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "knowledge_search",
		Description: "ALWAYS USE THIS FIRST! Search Zerops knowledge base for service types, recipes, and configurations. Returns exact service type strings needed for import.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search: service names (mongodb, postgresql), frameworks (laravel, django), or 'list services' for all available",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to return (default: 10, max: 20)",
				},
			},
			"required": []string{"query"},
		},
		Handler: handleKnowledgeSearch,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "knowledge_get",
		Description: "Get EXACT configuration for a service or complete recipe. Use this to get the correct 'type' string for project_import. Returns working YAML examples.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Knowledge ID from search results. Format: 'services/mongodb' or 'recipe/laravel-minimal'",
				},
			},
			"required": []string{"id"},
		},
		Handler: handleKnowledgeGet,
	})
}

func handleKnowledgeSearch(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return shared.ErrorResponse("Search query is required"), nil
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 && l <= 20 {
		limit = int(l)
	}

	// Call the API
	searchReq := SearchRequest{
		Query: query,
		Limit: limit,
	}

	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to prepare request: %v", err)), nil
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", knowledgeAPIURL+"/api/v1/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to create request: %v", err)), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return shared.TextResponse(fmt.Sprintf("Error: Knowledge base API is unavailable\n\n"+
			"The API at %s is not responding.\n"+
			"Please try again later or contact support.", knowledgeAPIURL)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return shared.TextResponse(fmt.Sprintf("Error: API returned status %d\n\n%s", resp.StatusCode, string(body))), nil
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	if len(searchResp.Results) == 0 {
		return shared.TextResponse(fmt.Sprintf("No results found for: %s\n\n"+
			"Try different search terms:\n"+
			"  • Framework names: laravel, django, nextjs\n"+
			"  • Service types: nodejs, postgresql, redis\n"+
			"  • Features: database, cache, email", query)), nil
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Found %d result(s) for: %s\n\n", searchResp.Count, query))

	for i, result := range searchResp.Results {
		message.WriteString(fmt.Sprintf("%d. %s\n", i+1, formatName(result.Name)))
		message.WriteString(fmt.Sprintf("   ID: %s\n", result.ID))
		message.WriteString(fmt.Sprintf("   Type: %s\n", result.Type))

		if result.Summary != "" {
			message.WriteString(fmt.Sprintf("   Summary: %s\n", result.Summary))
		}

		if len(result.Tags) > 0 {
			message.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(result.Tags, ", ")))
		}

		message.WriteString(fmt.Sprintf("   Relevance: %.0f%%\n\n", result.Score*100))
	}

	message.WriteString("Use 'knowledge_get' with the ID to retrieve full content.")

	return shared.TextResponse(message.String()), nil
}

func handleKnowledgeGet(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return shared.ErrorResponse("Knowledge ID is required"), nil
	}

	if !strings.Contains(id, "/") {
		return shared.TextResponse(fmt.Sprintf("Invalid ID format: %s\n\n"+
			"Expected format: {type}/{name}\n"+
			"Examples:\n"+
			"  • service/nodejs\n"+
			"  • recipe/laravel\n"+
			"  • patterns/nextjs", id)), nil
	}

	// Call the API
	httpClient := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/v1/knowledge/%s", knowledgeAPIURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return shared.TextResponse(fmt.Sprintf("Error: Knowledge base API is unavailable\n\n"+
			"The API at %s is not responding.\n"+
			"Please try again later or contact support.", knowledgeAPIURL)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return shared.TextResponse(fmt.Sprintf("Knowledge not found: %s\n\n"+
			"Use 'knowledge_search' to find available content.", id)), nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return shared.TextResponse(fmt.Sprintf("Error: API returned status %d\n\n%s", resp.StatusCode, string(body))), nil
	}

	var knowledge KnowledgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&knowledge); err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	// Format the content nicely
	var prettyJSON strings.Builder
	// Content is already an interface{}, so just marshal it directly
	if prettyBytes, err := json.MarshalIndent(knowledge.Content, "", "  "); err == nil {
		prettyJSON.Write(prettyBytes)
	} else {
		// If it's not JSON-able, convert to string
		prettyJSON.WriteString(fmt.Sprintf("%v", knowledge.Content))
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Knowledge: %s\n", formatName(knowledge.Name)))
	message.WriteString(fmt.Sprintf("Type: %s\n", knowledge.Type))
	message.WriteString(fmt.Sprintf("ID: %s\n\n", knowledge.ID))
	
	// If it's a recipe, extract and show ready-to-use YAML
	if strings.HasPrefix(knowledge.ID, "recipe/") {
		yamlExample := extractYAMLFromRecipe(knowledge.Content)
		if yamlExample != "" {
			message.WriteString("READY TO USE YAML for project_import:\n")
			message.WriteString("```yaml\n")
			message.WriteString(yamlExample)
			message.WriteString("\n```\n\n")
		}
		
		// Also provide important recipe information
		if recipe, ok := knowledge.Content.(map[string]interface{}); ok {
			if sourceRecipe, ok := recipe["sourceRecipe"].(string); ok {
				message.WriteString(fmt.Sprintf("Recipe GitHub: https://github.com/zeropsio/%s\n", sourceRecipe))
			}
			// Check if there's a specific note about the recipe
			message.WriteString("\nIMPORTANT for recipes:\n")
			message.WriteString("- Use the EXACT YAML shown above\n")
			message.WriteString("- The buildFromGit field is REQUIRED for utility services:\n")
			message.WriteString("  * Adminer: https://github.com/zeropsio/recipe-adminer\n")
			message.WriteString("  * Adminerevo: https://github.com/zeropsio/recipe-adminerevo\n")
			message.WriteString("  * Mailpit: https://github.com/zeropsio/recipe-mailpit\n")
			message.WriteString("  * S3Browser: https://github.com/zeropsio/recipe-s3browser\n")
			message.WriteString("- These are pre-built Zerops recipes, NOT the original tool repos\n\n")
			message.WriteString("TO USE RECIPE WITH CUSTOM HOSTNAME:\n")
			message.WriteString("```yaml\nservices:\n")
			message.WriteString("  - hostname: yourcustomname  # Use ANY hostname you want\n")
			message.WriteString("    type: php@8.3\n")
			message.WriteString("    buildFromGit: https://github.com/zeropsio/recipe-adminer\n")
			message.WriteString("```\n")
			message.WriteString("Just change the hostname - keep the buildFromGit URL!\n\n")
		}
	}
	
	// If it's a service, show the exact type string to use
	if strings.HasPrefix(knowledge.ID, "services/") {
		serviceType := extractServiceType(knowledge.Content)
		if serviceType != "" {
			message.WriteString(fmt.Sprintf("EXACT TYPE TO USE: %s\n\n", serviceType))
			message.WriteString("Example YAML:\n```yaml\nservices:\n")
			message.WriteString(fmt.Sprintf("  - hostname: %s\n", strings.Split(serviceType, "@")[0]))
			message.WriteString(fmt.Sprintf("    type: %s\n", serviceType))
			if strings.Contains(prettyJSON.String(), "\"modes\"") {
				message.WriteString("    mode: NON_HA  # or HA for high availability\n")
			}
			message.WriteString("```\n\n")
		}
	}
	
	message.WriteString("Full Content:\n")
	message.WriteString("```json\n")
	message.WriteString(prettyJSON.String())
	message.WriteString("\n```")

	return shared.TextResponse(message.String()), nil
}

// extractYAMLFromRecipe extracts ready-to-use YAML from recipe content
func extractYAMLFromRecipe(content interface{}) string {
	// Content is already parsed as interface{}, cast it
	recipe, ok := content.(map[string]interface{})
	if !ok {
		return ""
	}
	
	// Look for services array in the recipe
	if services, ok := recipe["services"].([]interface{}); ok {
		// Convert to YAML format
		yaml := "services:\n"
		for _, service := range services {
			if svc, ok := service.(map[string]interface{}); ok {
				yaml += fmt.Sprintf("  - hostname: %v\n", svc["hostname"])
				yaml += fmt.Sprintf("    type: %v\n", svc["type"])
				
				// Add optional fields if present
				if buildFromGit, ok := svc["buildFromGit"]; ok {
					yaml += fmt.Sprintf("    buildFromGit: %v\n", buildFromGit)
				}
				if mode, ok := svc["mode"]; ok {
					yaml += fmt.Sprintf("    mode: %v\n", mode)
				}
				if enabled, ok := svc["enableSubdomainAccess"]; ok && enabled == true {
					yaml += "    enableSubdomainAccess: true\n"
				}
				if minContainers, ok := svc["minContainers"]; ok {
					yaml += fmt.Sprintf("    minContainers: %v\n", minContainers)
				}
				if maxContainers, ok := svc["maxContainers"]; ok {
					yaml += fmt.Sprintf("    maxContainers: %v\n", maxContainers)
				}
				// Include ports if present
				if ports, ok := svc["ports"].([]interface{}); ok && len(ports) > 0 {
					yaml += "    ports:\n"
					for _, port := range ports {
						if p, ok := port.(map[string]interface{}); ok {
							yaml += fmt.Sprintf("      - port: %v\n", p["port"])
							if httpSupport, ok := p["httpSupport"]; ok && httpSupport == true {
								yaml += "        httpSupport: true\n"
							}
						}
					}
				}
			}
		}
		
		// Add a comment about using custom hostnames for recipes
		if strings.Contains(yaml, "buildFromGit") && strings.Contains(yaml, "recipe") {
			yaml += "\n# To use with custom hostname: just change the hostname field!\n"
			yaml += "# Keep the buildFromGit URL the same.\n"
		}
		
		return yaml
	}
	return ""
}

// extractServiceType extracts the exact service type string from service content
func extractServiceType(content interface{}) string {
	// Content is already parsed as interface{}, cast it
	service, ok := content.(map[string]interface{})
	if !ok {
		return ""
	}
	
	// Look for type field
	if typeField, ok := service["type"].(string); ok {
		// Look for current/recommended version
		if versions, ok := service["versions"].([]interface{}); ok {
			for _, v := range versions {
				if version, ok := v.(map[string]interface{}); ok {
					if recommended, ok := version["recommended"].(bool); ok && recommended {
						if ver, ok := version["version"].(string); ok {
							return fmt.Sprintf("%s@%s", typeField, ver)
						}
					}
				}
			}
			// If no recommended, use first current version
			for _, v := range versions {
				if version, ok := v.(map[string]interface{}); ok {
					if status, ok := version["status"].(string); ok && status == "current" {
						if ver, ok := version["version"].(string); ok {
							return fmt.Sprintf("%s@%s", typeField, ver)
						}
					}
				}
			}
		}
	}
	return ""
}

