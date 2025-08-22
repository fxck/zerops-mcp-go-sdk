package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterKnowledgeBase registers knowledge base tool
func RegisterKnowledgeBase() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "knowledge_base",
		Description: `Provides comprehensive YAML examples and deployment patterns for specific runtimes.

This tool provides the complete zerops.yml examples that align with the Zerops development workflow.
Includes both 'dev' and 'prod' setups for proper development and staging deployment patterns.

AVAILABLE RUNTIMES:
- Web: nodejs, python, go, php, rust
- Databases: postgresql, mariadb, mongodb
- Cache: redis, valkey, keydb
- Storage: elasticsearch, objectstorage
- Web servers: nginx, static

RETURNS:
- Complete zerops.yml with dev/prod setups
- Service import YAML examples
- Runtime-specific configuration patterns
- Development and production best practices

KEY PATTERNS:
- Dev setup: deployFiles: ./ (preserves source), start: zsc noop (manual control)
- Prod setup: deployFiles: dist (production files), start: npm start (auto-start)
- Environment variable patterns using ${variable} syntax
- Proper port and build configurations

WHEN TO USE:
- Before importing services to get correct YAML format
- Setting up development and staging environments
- Learning proper zerops.yml structure
- Understanding dev vs prod deployment differences`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"runtime": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Runtime type to get examples for",
					"enum": []interface{}{
						"nodejs", "node", "python", "go", "golang", "php", "rust",
						"postgresql", "postgres", "mariadb", "mysql", "mongodb", "mongo",
						"redis", "valkey", "keydb", "elasticsearch", "rabbitmq",
						"nginx", "static", "objectstorage",
					},
				},
			},
			"required":             []string{"runtime"},
			"additionalProperties": false,
		},
		Handler: handleKnowledgeBase,
	})

	// Load platform guide
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "load_platform_guide",
		Description: `Loads comprehensive workflow guides for different development scenarios from GitHub repository.

Fetches the latest guides from https://github.com/zeropsio/zagent-knowledge with 10-minute caching.
These guides align with the Zerops development methodology and provide detailed step-by-step workflows.

AVAILABLE GUIDES:
- fresh_project: Complete setup from scratch (databases → services → hello-world → development)
- existing_service: Most common scenario - start development on existing services
- add_services: Expand existing projects with new services

EACH GUIDE INCLUDES:
- The mandatory hello-world pattern for new services
- Proper dev/stage deployment workflows
- Environment variable management patterns
- Service restart and remount procedures
- Integration testing approaches

WHEN TO USE:
- After discovery() to determine your development path
- When starting a completely new project (fresh_project)
- When working on existing services (existing_service) 
- When adding new functionality/services (add_services)
- Need structured workflow guidance

FETCHING:
- Content fetched from GitHub zagent-knowledge repository
- 10-minute cache to reduce API calls
- Falls back to local content if GitHub unavailable`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path_type": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Type of guide to load",
					"enum": []interface{}{
						"fresh_project", "existing_service", "add_services",
					},
				},
			},
			"required":             []string{"path_type"},
			"additionalProperties": false,
		},
		Handler: handleLoadPlatformGuide,
	})
}

func handleKnowledgeBase(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	runtime, ok := args["runtime"].(string)
	if !ok || runtime == "" {
		return shared.ErrorResponse("Runtime is required"), nil
	}

	runtime = strings.ToLower(runtime)

	// Return knowledge base examples based on runtime
	switch runtime {
	case "nodejs", "node":
		return getNodejsKnowledge(), nil
	case "python":
		return getPythonKnowledge(), nil
	case "go", "golang":
		return getGoKnowledge(), nil
	case "php":
		return getPHPKnowledge(), nil
	case "postgresql", "postgres":
		return getPostgreSQLKnowledge(), nil
	case "mariadb", "mysql":
		return getMariaDBKnowledge(), nil
	case "mongodb", "mongo":
		return getMongoDBKnowledge(), nil
	case "redis", "valkey":
		return getCacheKnowledge(), nil
	default:
		return map[string]interface{}{
			"runtime": runtime,
			"message": fmt.Sprintf("Runtime '%s' not directly supported. Use Node.js pattern as reference.", runtime),
			"pattern": getNodejsKnowledge(),
		}, nil
	}
}

func getNodejsKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "Node.js",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: app
    type: nodejs@22
    enableSubdomainAccess: true
    minContainers: 1`,
			"with_database": `services:
  - hostname: app
    type: nodejs@22
    enableSubdomainAccess: true
    minContainers: 1
  - hostname: db
    type: postgresql@16
    mode: NON_HA`,
			"production": `services:
  - hostname: app
    type: nodejs@22
    enableSubdomainAccess: true
    minContainers: 2
    maxContainers: 5
    verticalAutoscaling:
      cpuMin: 1
      cpuMax: 4
      ramMin: 1
      ramMax: 4`,
		},
		"deployment_yaml": `zerops:
  - setup: prod
    build:
      base: nodejs@22
      prepareCommands:
        - npm install -g typescript
      buildCommands:
        - npm i
        - npm run build
      deployFiles:
        - ./dist
        - ./node_modules
        - ./package.json
    run:
      base: nodejs@22
      ports:
        - port: 3000
          httpSupport: true
      envVariables:
        NODE_ENV: production
        DB_NAME: db
        DB_HOST: db
        DB_USER: db
        DB_PASS: ${db_password}
      start: npm run start:prod
      healthCheck:
        httpGet:
          port: 3000
          path: /status

  - setup: dev
    build:
      base: nodejs@22
      # pre-install deps
      buildCommands:
        - npm i
      # deploy the whole source
      deployFiles: ./
    run:
      base: nodejs@22
      envVariables:
        DB_NAME: db
        DB_HOST: db
        DB_USER: db
        DB_PASS: ${db_password}
      ports:
        - port: 3000
          httpSupport: true
      # user or agent will start the dev server
      start: zsc noop`,
		"tips": []string{
			"Use nodejs@22 for latest LTS version",
			"Enable subdomain for public access",
			"Set minContainers: 2 for high availability",
			"Use environment variables for configuration",
			"Use 'prod' setup for production-like services, 'dev' for development/remote",
			"Reference database password with ${db_password}",
		},
	}
}

func getPythonKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "Python",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: app
    type: python@3.12
    enableSubdomainAccess: true
    minContainers: 1`,
			"with_database": `services:
  - hostname: app
    type: python@3.12
    enableSubdomainAccess: true
    minContainers: 1
  - hostname: db
    type: postgresql@16
    mode: NON_HA`,
		},
		"deployment_yaml": `# Adapt the Node.js pattern above for Python:
# - Use python@3.12 as base
# - Replace npm commands with pip install -r requirements.txt
# - Use appropriate Python start command (gunicorn, uvicorn, etc.)
# - Adjust port numbers and health check paths
# - Reference the Node.js example structure`,
		"note": "Full Python zerops.yml examples coming soon. Use the Node.js pattern above as reference, adapting commands and runtime.",
		"tips": []string{
			"Use python@3.12 for latest stable version",
			"Adapt Node.js zerops.yml pattern for Python",
			"Replace npm commands with pip install",
			"Use gunicorn or uvicorn for production start",
		},
	}
}

func getGoKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "Go",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: app
    type: go@1.22
    enableSubdomainAccess: true
    minContainers: 1`,
			"with_database": `services:
  - hostname: api
    type: go@1.22
    enableSubdomainAccess: true
    minContainers: 2
  - hostname: db
    type: postgresql@16
    mode: HA`,
		},
		"deployment_yaml": `# Adapt the Node.js pattern above for Go:
# - Use go@1.22 as base
# - Replace npm commands with 'go mod download' and 'go build'
# - Deploy the compiled binary
# - Use './app' or similar as start command
# - Reference the Node.js example structure`,
		"note": "Full Go zerops.yml examples coming soon. Use the Node.js pattern above as reference, adapting commands and runtime.",
		"tips": []string{
			"Use go@1.22 for latest version",
			"Adapt Node.js zerops.yml pattern for Go",
			"Build binary in buildCommands",
			"Deploy only the binary for smaller image",
		},
	}
}

func getPHPKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "PHP",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: app
    type: php@8.3
    enableSubdomainAccess: true
    minContainers: 1`,
			"laravel": `services:
  - hostname: laravel
    type: php@8.3
    enableSubdomainAccess: true
    minContainers: 1
  - hostname: db
    type: mariadb@11
    mode: NON_HA
  - hostname: cache
    type: valkey@7.2
    mode: NON_HA`,
		},
		"deployment_yaml": `# Adapt the Node.js pattern above for PHP:
# - Use php@8.3 as base
# - Replace npm commands with 'composer install --no-dev'
# - Set documentRoot: public for frameworks like Laravel
# - Use php artisan commands for Laravel
# - Reference the Node.js example structure`,
		"note": "Full PHP zerops.yml examples coming soon. Use the Node.js pattern above as reference, adapting commands and runtime.",
		"tips": []string{
			"Use php@8.3 for latest stable version",
			"Adapt Node.js zerops.yml pattern for PHP",
			"Set documentRoot: public for frameworks",
			"Use composer for dependency management",
		},
	}
}

func getPostgreSQLKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "PostgreSQL",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: db
    type: postgresql@16
    mode: NON_HA`,
			"high_availability": `services:
  - hostname: db
    type: postgresql@16
    mode: HA
    verticalAutoscaling:
      cpuMin: 2
      cpuMax: 8
      ramMin: 4
      ramMax: 16`,
		},
		"tips": []string{
			"Use postgresql@16 or @17 for latest versions",
			"Use mode: HA for production",
			"Automatic backups are included",
			"Connection pooling is built-in",
		},
	}
}

func getMariaDBKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "MariaDB",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: db
    type: mariadb@11
    mode: NON_HA`,
			"high_availability": `services:
  - hostname: db
    type: mariadb@11
    mode: HA`,
		},
		"tips": []string{
			"Use mariadb@11 for latest version",
			"Compatible with MySQL applications",
			"Automatic backups included",
		},
	}
}

func getMongoDBKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "MongoDB",
		"examples": map[string]interface{}{
			"basic": `services:
  - hostname: db
    type: mongodb@7
    mode: NON_HA`,
			"replica_set": `services:
  - hostname: db
    type: mongodb@7
    mode: HA`,
		},
		"tips": []string{
			"Use mongodb@7 for latest version",
			"HA mode provides replica set",
			"Automatic backups included",
		},
	}
}

func getCacheKnowledge() interface{} {
	return map[string]interface{}{
		"runtime": "Cache (Redis/Valkey)",
		"examples": map[string]interface{}{
			"valkey": `services:
  - hostname: cache
    type: valkey@7.2
    mode: NON_HA`,
			"redis": `services:
  - hostname: cache
    type: redis@7
    mode: NON_HA`,
			"keydb": `services:
  - hostname: cache
    type: keydb@6.3
    mode: HA`,
		},
		"tips": []string{
			"Valkey is Redis-compatible",
			"KeyDB supports multi-threading",
			"Use HA mode for production",
		},
	}
}


func handleLoadPlatformGuide(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	pathType, ok := args["path_type"].(string)
	if !ok || pathType == "" {
		return shared.ErrorResponse("Path type is required"), nil
	}

	switch pathType {
	case "fresh_project":
		return getFreshProjectGuide(), nil
	case "existing_service":
		return getExistingServiceGuide(), nil
	case "add_services":
		return getAddServicesGuide(), nil
	default:
		return shared.ErrorResponse(fmt.Sprintf("Unknown path type '%s'. Available: fresh_project, existing_service, add_services", pathType)), nil
	}
}

// Guide cache with 10-minute expiration
var (
	guideCache = make(map[string]cacheEntry)
	cacheMutex sync.RWMutex
)

type cacheEntry struct {
	content   interface{}
	timestamp time.Time
}

func getFreshProjectGuide() interface{} {
	return fetchGuideFromGitHub("fresh_project")
}

func getExistingServiceGuide() interface{} {
	return fetchGuideFromGitHub("existing_service")
}

func getAddServicesGuide() interface{} {
	return fetchGuideFromGitHub("add_services")
}

func fetchGuideFromGitHub(pathType string) interface{} {
	cacheMutex.RLock()
	if entry, exists := guideCache[pathType]; exists {
		if time.Since(entry.timestamp) < 10*time.Minute {
			cacheMutex.RUnlock()
			return entry.content
		}
	}
	cacheMutex.RUnlock()

	// Fetch from GitHub
	baseURL := "https://raw.githubusercontent.com/zeropsio/zagent-knowledge/main"
	var fileURL string
	switch pathType {
	case "fresh_project":
		fileURL = fmt.Sprintf("%s/fresh_project.md", baseURL)
	case "existing_service":
		fileURL = fmt.Sprintf("%s/existing_service.md", baseURL)
	case "add_services":
		fileURL = fmt.Sprintf("%s/add_services.md", baseURL)
	default:
		return map[string]interface{}{
			"error": "Unknown path type",
			"available": []string{"fresh_project", "existing_service", "add_services"},
		}
	}

	// Fetch actual content from GitHub
	content, err := fetchFromURL(fileURL)
	var result interface{}
	
	if err != nil {
		// Fallback to local content on error
		result = map[string]interface{}{
			"source": "fallback",
			"error":  fmt.Sprintf("Failed to fetch from GitHub: %v", err),
			"content": getFallbackGuide(pathType),
		}
	} else {
		// Return the fetched markdown content
		result = map[string]interface{}{
			"source":    "github",
			"path_type": pathType,
			"url":       fileURL,
			"content":   content,
			"cached_at": time.Now().Format("2006-01-02 15:04:05"),
			"cache_expires": time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05"),
		}
	}

	// Cache the result
	cacheMutex.Lock()
	guideCache[pathType] = cacheEntry{
		content:   result,
		timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	return result
}

func fetchFromURL(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

func getFallbackGuide(pathType string) interface{} {
	switch pathType {
	case "fresh_project":
		return map[string]interface{}{
			"path_type": "fresh_project",
			"title":     "Complete Guide: Starting a Fresh Project",
			"workflow": []string{
				"1. discovery() - Check current project state (should be empty)",
				"2. get_service_types() - See available service types",
				"3. knowledge_base('nodejs') - Get complete YAML examples",
				"4. import_services(yaml: '...') - Create your services",
				"5. discovery() - Verify services were created",
				"6. enable_preview_subdomain(service_id: '...') - Enable public access",
				"7. get_process_status(process_id: '...') - Monitor subdomain setup",
				"8. discovery() - Get final subdomain URLs",
				"9. set_project_env() / set_service_env() - Configure environment",
				"10. get_service_logs() - Monitor application startup",
			},
		}
	case "existing_service":
		return map[string]interface{}{
			"path_type": "existing_service",
			"title":     "Guide: Working with Existing Services",
			"workflow": []string{
				"1. discovery() - See all existing services and their status",
				"2. get_service_logs(service_id: '...') - Check current service health",
				"3. set_service_env() / set_project_env() - Update configuration",
				"4. restart_service(service_id: '...') - Apply configuration changes",
				"5. get_process_status(process_id: '...') - Monitor restart progress",
				"6. scale_service(service_id: '...') - Adjust resources if needed",
				"7. remount_service(service_name: '...') - Fix SSHFS issues if needed",
				"8. discovery() - Verify final state",
			},
		}
	case "add_services":
		return map[string]interface{}{
			"path_type": "add_services",
			"title":     "Guide: Adding Services to Existing Project",
			"workflow": []string{
				"1. discovery() - See current project services",
				"2. get_service_types() - Check available service types",
				"3. knowledge_base('database_type') - Get examples for new services",
				"4. import_services(yaml: '...') - Add new services to project",
				"5. discovery() - Verify new services were created",
				"6. set_project_env() - Add shared environment variables",
				"7. restart_service() - Restart existing services to use new config",
				"8. enable_preview_subdomain() - Enable access for web services",
				"9. get_running_processes() - Monitor all operations",
				"10. discovery() - Get final project state",
			},
		}
	default:
		return map[string]interface{}{
			"error": "Unknown guide type",
		}
	}
}