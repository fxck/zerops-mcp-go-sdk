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

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zerops-go/sdk"
)

const knowledgeAPIURL = "https://kbapi-167b-8080.prg1.zerops.app"

// SearchRequest represents the search API request
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchResult represents individual search result
type SearchResult struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Score   float64  `json:"score"`
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Count   int            `json:"count"`
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

// KnowledgeResponse represents the full knowledge content from the API
type KnowledgeResponse struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// RegisterKnowledge registers knowledge base tools that use the external API
func RegisterKnowledge(server *mcp.Server, client *sdk.Handler) {
	registerKnowledgeSearch(server)
	registerKnowledgeGet(server)
}

func registerKnowledgeSearch(server *mcp.Server) {
	type SearchArgs struct {
		Query string `json:"query" mcp:"Search terms: framework names (laravel, django), services (nodejs, postgresql), or features (database, cache)"`
		Limit *int   `json:"limit,omitempty" mcp:"Number of results to return (default: 10, max: 20)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "knowledge_search",
		Description: "Search Zerops knowledge base API for services, recipes, and deployment patterns",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		limit := 10
		if args.Limit != nil && *args.Limit > 0 && *args.Limit <= 20 {
			limit = *args.Limit
		}

		// Call the API
		searchReq := SearchRequest{
			Query: args.Query,
			Limit: limit,
		}

		jsonData, err := json.Marshal(searchReq)
		if err != nil {
			return errorResult(fmt.Errorf("failed to prepare request: %w", err)), nil
		}

		httpClient := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "POST", knowledgeAPIURL+"/api/v1/search", bytes.NewBuffer(jsonData))
		if err != nil {
			return errorResult(fmt.Errorf("failed to create request: %w", err)), nil
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return textResult(fmt.Sprintf("Error: Knowledge base API is unavailable\n\n"+
				"The API at %s is not responding.\n"+
				"Please try again later or contact support.", knowledgeAPIURL)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return textResult(fmt.Sprintf("Error: API returned status %d\n\n%s", resp.StatusCode, string(body))), nil
		}

		var searchResp SearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		if len(searchResp.Results) == 0 {
			return textResult(fmt.Sprintf("No results found for: %s\n\n"+
				"Try different search terms:\n"+
				"  • Framework names: laravel, django, nextjs\n"+
				"  • Service types: nodejs, postgresql, redis\n"+
				"  • Features: database, cache, email", args.Query)), nil
		}

		message := fmt.Sprintf("Found %d result(s) for: %s\n\n", searchResp.Count, args.Query)

		for i, result := range searchResp.Results {
			message += fmt.Sprintf("%d. %s\n", i+1, formatName(result.Name))
			message += fmt.Sprintf("   ID: %s\n", result.ID)
			message += fmt.Sprintf("   Type: %s\n", result.Type)

			if result.Summary != "" {
				message += fmt.Sprintf("   Summary: %s\n", result.Summary)
			}

			if len(result.Tags) > 0 {
				message += fmt.Sprintf("   Tags: %s\n", strings.Join(result.Tags, ", "))
			}

			message += fmt.Sprintf("   Relevance: %.0f%%\n\n", result.Score*100)
		}

		message += "Use 'knowledge_get' with the ID to retrieve full content."

		return textResult(message), nil
	})
}

func registerKnowledgeGet(server *mcp.Server) {
	type GetArgs struct {
		ID string `json:"id" mcp:"Knowledge ID (format: {type}/{name}, e.g., 'recipe/laravel' or 'service/nodejs')"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "knowledge_get",
		Description: "Get full content of a specific knowledge item from the API by its semantic ID",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[GetArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		if !strings.Contains(args.ID, "/") {
			return textResult(fmt.Sprintf("Invalid ID format: %s\n\n"+
				"Expected format: {type}/{name}\n"+
				"Examples:\n"+
				"  • service/nodejs\n"+
				"  • recipe/laravel\n"+
				"  • patterns/nextjs", args.ID)), nil
		}

		// Call the API
		httpClient := &http.Client{Timeout: 10 * time.Second}
		url := fmt.Sprintf("%s/api/v1/knowledge/%s", knowledgeAPIURL, args.ID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return errorResult(fmt.Errorf("failed to create request: %w", err)), nil
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return textResult(fmt.Sprintf("Error: Knowledge base API is unavailable\n\n"+
				"The API at %s is not responding.\n"+
				"Please try again later or contact support.", knowledgeAPIURL)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return textResult(fmt.Sprintf("Knowledge not found: %s\n\n"+
				"Use 'knowledge_search' to find available content.", args.ID)), nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return textResult(fmt.Sprintf("Error: API returned status %d\n\n%s", resp.StatusCode, string(body))), nil
		}

		var knowledge KnowledgeResponse
		if err := json.NewDecoder(resp.Body).Decode(&knowledge); err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		// Format the content nicely
		var prettyJSON strings.Builder
		var temp interface{}
		if err := json.Unmarshal(knowledge.Content, &temp); err == nil {
			prettyBytes, _ := json.MarshalIndent(temp, "", "  ")
			prettyJSON.Write(prettyBytes)
		} else {
			prettyJSON.Write(knowledge.Content)
		}

		message := fmt.Sprintf("Knowledge: %s\n", formatName(knowledge.Name))
		message += fmt.Sprintf("Type: %s\n", knowledge.Type)
		message += fmt.Sprintf("ID: %s\n\n", knowledge.ID)
		message += "Content:\n"
		message += "```json\n"
		message += prettyJSON.String()
		message += "\n```"

		return textResult(message), nil
	})
}

// Helper function to format names
func formatName(name string) string {
	// Convert kebab-case or snake_case to Title Case
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, " ")
}
