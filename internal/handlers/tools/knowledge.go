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
		Description: "Search Zerops knowledge base API for services, recipes, and deployment patterns",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search terms: framework names (laravel, django), services (nodejs, postgresql), or features (database, cache)",
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
		Description: "Get full content of a specific knowledge item from the API by its semantic ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Knowledge ID (format: {type}/{name}, e.g., 'recipe/laravel' or 'service/nodejs')",
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
	message.WriteString("Content:\n")
	message.WriteString("```json\n")
	message.WriteString(prettyJSON.String())
	message.WriteString("\n```")

	return shared.TextResponse(message.String()), nil
}
