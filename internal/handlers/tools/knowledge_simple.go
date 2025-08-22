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
		Description: `Provides comprehensive service import YAML examples and configuration patterns.

QUERY TYPES:
- "service_import" - Get service import YAML patterns for databases, storage, runtime services
- "runtime_name" (nodejs, python, go, php) - Get complete zerops.yml examples with dev/prod setups
- "database_patterns" - Get database and storage service configurations
- "autoscaling" - Get vertical/horizontal autoscaling configurations

RETURNS:
- Complete service import YAML with all parameters
- Runtime-specific zerops.yml with dev/prod setups  
- Database, cache, storage service patterns
- Autoscaling and mount configurations
- Environment variables and secrets patterns`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"runtime": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Query type - use 'service_import' for service import patterns, 'database_patterns' for databases/storage, 'autoscaling' for scaling configs, or specific runtime name (nodejs, python, go, php) for zerops.yml examples",
					"enum": []interface{}{
						"service_import", "database_patterns", "autoscaling",
						"nodejs", "node", "bun", "deno", "golang", "go", "rust",
						"python", "dotnet", "java", "php", "elixir", "gleam", "ruby",
						"postgresql", "mariadb", "clickhouse", "valkey", "keydb",
						"elasticsearch", "meilisearch", "typesense", "qdrant",
						"kafka", "nats", "nginx", "static", "object-storage",
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
	case "service_import":
		return getServiceImportPatterns(), nil
	case "database_patterns":
		return getDatabasePatterns(), nil
	case "autoscaling":
		return getAutoscalingPatterns(), nil
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

func getServiceImportPatterns() interface{} {
	return map[string]interface{}{
		"title": "Complete Service Import YAML Patterns",
		"description": "Examples use current versions from get_service_types. Always verify with get_service_types for latest versions.",
		"patterns": map[string]interface{}{
			"basic_runtime": `# Basic runtime service (Node.js, Python, Go, PHP)
services:
  - hostname: app                    # REQUIRED: alphanumeric only, max 25 chars
    type: nodejs@22                  # REQUIRED: from get_service_types
    startWithoutCode: true           # CRITICAL for dev services
    enableSubdomainAccess: true      # Enable public access
    minContainers: 1                 # Horizontal scaling
    maxContainers: 3
    buildFromGit: https://github.com/myorg/myapp
    priority: 1                      # Higher = created first`,
			
			"dev_stage_pair": `# Development + Stage service pair (recommended)
services:
  - hostname: apidev
    type: nodejs@22
    startWithoutCode: true           # CRITICAL: Allows manual dev control
    enableSubdomainAccess: true
  - hostname: apistage  
    type: nodejs@22                  # Auto-starts for production`,
			
			"with_secrets": `# Service with environment secrets
services:
  - hostname: app
    type: nodejs@22
    startWithoutCode: true
    envSecrets:
      SECRET_KEY: mySecretValue
      API_TOKEN: abc123
    dotEnvSecrets: |
      DATABASE_URL=postgresql://user:pass@db:5432/myapp
      REDIS_URL=redis://cache:6379`,
			
			"database_services": `# Database and storage services (import FIRST)
services:
  - hostname: db
    type: postgresql@17              # Latest PostgreSQL
    mode: NON_HA                     # or HA for high availability
  - hostname: cache
    type: valkey@7.2                 # Modern Redis alternative
    mode: NON_HA
  - hostname: storage
    type: object-storage
    objectStorageSize: 5             # Size in GB
    objectStoragePolicy: public-read # or private, public-write, etc.`,
			
			"with_autoscaling": `# Service with vertical autoscaling
services:
  - hostname: app
    type: nodejs@22
    startWithoutCode: true
    verticalAutoscaling:
      minCpu: 1                      # Min virtual CPUs
      maxCpu: 4                      # Max virtual CPUs  
      cpuMode: SHARED                # SHARED or DEDICATED
      minRam: 1                      # Min RAM in GB
      maxRam: 8                      # Max RAM in GB
      minDisk: 1                     # Min disk in GB
      maxDisk: 20                    # Max disk in GB`,
			
			"with_mounts": `# Service with shared storage mounts
services:
  - hostname: app
    type: php@8.3
    buildFromGit: https://github.com/myorg/myapp
    mount:
      - sharedstorage1               # Mount existing shared storage
      - sharedstorage2`,
			
			"complete_example": `# Complete example with all options
services:
  - hostname: webapi
    type: nodejs@22
    mode: NON_HA
    startWithoutCode: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/myorg/webapp
    priority: 2
    minContainers: 2
    maxContainers: 6
    envSecrets:
      JWT_SECRET: myJwtSecret
      DATABASE_PASSWORD: dbPassword
    dotEnvSecrets: |
      NODE_ENV=development
      LOG_LEVEL=debug
    verticalAutoscaling:
      minCpu: 1
      maxCpu: 3
      minRam: 2
      maxRam: 6
    mount:
      - uploads`,
		},
		"field_reference": map[string]interface{}{
			"hostname": "REQUIRED: Unique service identifier, alphanumeric only, max 25 chars",
			"type": "REQUIRED: Service type and version (from get_service_types)",
			"mode": "HA or NON_HA (default: NON_HA)",
			"startWithoutCode": "CRITICAL for dev services - prevents auto-start",
			"enableSubdomainAccess": "Enable public access via Zerops subdomain",
			"buildFromGit": "GitHub/GitLab repository URL for one-time build",
			"priority": "Creation order (higher = created first)",
			"minContainers/maxContainers": "Horizontal autoscaling (1-10)",
			"envSecrets": "Secret environment variables (hidden in GUI)",
			"dotEnvSecrets": ".env format environment variables",
			"objectStorageSize": "Storage size in GB (for objectstorage type)",
			"objectStoragePolicy": "private, public-read, public-write, public-read-write, custom",
			"verticalAutoscaling": "CPU, RAM, disk scaling configuration",
			"mount": "List of shared storage services to mount",
		},
		"workflow_order": []string{
			"1. ALWAYS run get_service_types first to get exact service names and versions",
			"2. Import databases/storage FIRST (postgresql, valkey, object-storage)",
			"3. Monitor completion with get_process_status", 
			"4. Import runtime services with startWithoutCode: true",
			"5. Monitor completion with get_process_status",
			"6. Deploy hello-world pattern to validate pipeline",
			"7. Begin real development only after validation",
		},
	}
}

func getDatabasePatterns() interface{} {
	return map[string]interface{}{
		"title": "Database and Storage Service Patterns",
		"description": "Complete configuration examples for databases, caches, and storage services",
		"databases": map[string]interface{}{
			"postgresql": `# PostgreSQL database
services:
  - hostname: db
    type: postgresql@17              # Latest stable version
    mode: NON_HA                     # or HA for production`,
			
			"mariadb": `# MariaDB database
services:
  - hostname: db
    type: mariadb@11
    mode: NON_HA`,
			
			"clickhouse": `# ClickHouse database
services:
  - hostname: analytics
    type: clickhouse@25.3
    mode: NON_HA`,
			
		},
		"caches": map[string]interface{}{
			"valkey": `# Valkey (Redis alternative)
services:
  - hostname: cache
    type: valkey@7
    mode: NON_HA`,
			
			"keydb": `# KeyDB (multi-master Redis)
services:
  - hostname: cache
    type: keydb@6
    mode: NON_HA`,
		},
		"search_engines": map[string]interface{}{
			"elasticsearch": `# Elasticsearch
services:
  - hostname: search
    type: elasticsearch@8.16
    mode: NON_HA`,
			
			"meilisearch": `# Meilisearch
services:
  - hostname: search
    type: meilisearch@1
    mode: NON_HA`,
			
			"typesense": `# Typesense
services:
  - hostname: search
    type: typesense@0
    mode: NON_HA`,
		},
		"vector_databases": map[string]interface{}{
			"qdrant": `# Qdrant vector database
services:
  - hostname: vectors
    type: qdrant@1
    mode: NON_HA`,
		},
		"message_brokers": map[string]interface{}{
			"kafka": `# Apache Kafka
services:
  - hostname: events
    type: kafka@3.8
    mode: NON_HA`,
			
			"nats": `# NATS messaging
services:
  - hostname: messaging
    type: nats@2.10
    mode: NON_HA`,
		},
		"storage": map[string]interface{}{
			"object_storage": `# Object storage (S3-compatible)
services:
  - hostname: storage
    type: object-storage
    objectStorageSize: 10            # Size in GB
    objectStoragePolicy: private     # Access policy`,
			
			"shared_storage": `# Shared storage
services:
  - hostname: shared
    type: shared-storage`,
			
			"elasticsearch": `# Elasticsearch
services:
  - hostname: search
    type: elasticsearch@8.16
    mode: NON_HA`,
		},
		"complete_stack": `# Complete managed services stack
services:
  # Primary database
  - hostname: db
    type: postgresql@17
    mode: NON_HA
    priority: 10                     # Create first
  
  # Cache layer  
  - hostname: cache
    type: valkey@7
    mode: NON_HA
    priority: 9
  
  # File storage
  - hostname: storage
    type: objectstorage
    objectStorageSize: 50
    objectStoragePolicy: public-read
    priority: 8
  
  # Search engine
  - hostname: search
    type: elasticsearch@8
    mode: NON_HA
    priority: 7
  
  # Message broker
  - hostname: events
    type: kafka@3.8
    mode: NON_HA
    priority: 6`,
		"environment_variables": map[string]interface{}{
			"description": "Auto-generated environment variables available to other services",
			"postgresql": []string{
				"db_connectionString",
				"db_hostname", 
				"db_port",
				"db_user",
				"db_password",
			},
			"objectstorage": []string{
				"storage_hostname",
				"storage_accessKeyId", 
				"storage_secretAccessKey",
				"storage_bucketName",
			},
			"usage": "Access from other services using ${servicename_variablename} syntax",
		},
	}
}

func getAutoscalingPatterns() interface{} {
	return map[string]interface{}{
		"title": "Autoscaling Configuration Patterns",
		"description": "Vertical and horizontal autoscaling examples for services",
		"vertical_autoscaling": map[string]interface{}{
			"basic": `# Basic vertical autoscaling
services:
  - hostname: app
    type: nodejs@22
    verticalAutoscaling:
      minCpu: 1                      # Minimum virtual CPUs
      maxCpu: 4                      # Maximum virtual CPUs
      minRam: 1                      # Minimum RAM in GB
      maxRam: 8                      # Maximum RAM in GB`,
			
			"advanced": `# Advanced vertical autoscaling with thresholds
services:
  - hostname: app
    type: nodejs@22
    verticalAutoscaling:
      minCpu: 2
      maxCpu: 8
      cpuMode: DEDICATED             # SHARED or DEDICATED
      minRam: 2
      maxRam: 16
      minDisk: 5                     # Minimum disk space in GB
      maxDisk: 50                    # Maximum disk space in GB
      startCpuCoreCount: 2           # Initial CPU cores
      minFreeCpuCores: 0.5           # Min free CPU before scaling
      minFreeCpuPercent: 20          # Min free CPU percentage
      minFreeRamGB: 1                # Min free RAM in GB
      minFreeRamPercent: 15          # Min free RAM percentage`,
		},
		"horizontal_autoscaling": map[string]interface{}{
			"basic": `# Basic horizontal autoscaling
services:
  - hostname: app
    type: nodejs@22
    minContainers: 2                 # Minimum containers
    maxContainers: 6                 # Maximum containers (max 10)`,
			
			"load_balanced": `# Load-balanced service with scaling
services:
  - hostname: api
    type: nodejs@22
    enableSubdomainAccess: true
    minContainers: 3                 # Always have 3 instances
    maxContainers: 10                # Scale up to 10 under load`,
		},
		"combined_scaling": `# Both vertical and horizontal autoscaling
services:
  - hostname: webapp
    type: nodejs@22
    enableSubdomainAccess: true
    # Horizontal scaling
    minContainers: 2
    maxContainers: 8
    # Vertical scaling  
    verticalAutoscaling:
      minCpu: 1
      maxCpu: 4
      cpuMode: SHARED
      minRam: 2
      maxRam: 8
      minDisk: 5
      maxDisk: 20`,
		"scaling_tips": []string{
			"Start with conservative limits and increase based on monitoring",
			"Use horizontal scaling for stateless applications",
			"Use vertical scaling for CPU/memory intensive tasks", 
			"DEDICATED CPU mode for predictable performance",
			"SHARED CPU mode for cost optimization",
			"Monitor scaling events in Zerops dashboard",
			"Set appropriate free resource thresholds to prevent constant scaling",
		},
	}
}