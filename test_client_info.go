package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	// Test initialize request to see what Claude sends
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "Claude Desktop",
				"version": "1.0.0",
			},
		},
	}

	jsonData, _ := json.MarshalIndent(request, "", "  ")
	fmt.Println("Example initialize request from Claude:")
	fmt.Println(string(jsonData))

	// Send to local server if running
	resp, err := http.Post(
		"http://localhost:8080/",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Println("\nNo server running at localhost:8080")
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("\nServer response:")
	responseJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(responseJSON))
}