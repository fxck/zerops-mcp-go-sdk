package tools

import (
	"strings"
	
	"github.com/zeropsio/zerops-go/dto/output"
)

// Common types used across tool implementations

// projectInfo holds project information with organization context
type projectInfo struct {
	Project output.EsProject
	OrgName string
	OrgId   string
}

// Knowledge API constants
const knowledgeAPIURL = "https://kbapi-167b-8080.prg1.zerops.app"

// SearchRequest represents a search request to the knowledge API
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SearchResponse represents the response from knowledge search
type SearchResponse struct {
	Results []struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Score    float64  `json:"score"`
		Summary  string   `json:"summary,omitempty"`
		Tags     []string `json:"tags,omitempty"`
	} `json:"results"`
	Count int `json:"count"`
}

// KnowledgeResponse represents a single knowledge item
type KnowledgeResponse struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// formatName converts kebab-case or snake_case to Title Case
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