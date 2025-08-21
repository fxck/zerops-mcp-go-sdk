package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterKnowledgeBase registers knowledge base tool
func RegisterKnowledgeBase() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "knowledge_base",
		Description: `Provides comprehensive YAML examples and deployment patterns for specific runtimes.

AVAILABLE RUNTIMES:
- Web: nodejs, python, go, php, rust
- Databases: postgresql, mariadb, mongodb
- Cache: redis, valkey, keydb
- Storage: elasticsearch, objectstorage
- Web servers: nginx, static

RETURNS:
- Import YAML examples (for import_services)
- Deployment YAML examples (for zerops.yml)
- Runtime-specific tips and best practices
- Common configuration patterns

EXAMPLES PROVIDED:
- Basic single-service setup
- Multi-service applications with databases
- Production configurations with scaling
- High-availability setups

WHEN TO USE:
- Before importing services to get correct YAML format
- Learning Zerops configuration patterns
- Setting up common application stacks
- Getting runtime-specific recommendations

FOLLOW-UP ACTIONS:
- Copy YAML examples for import_services tool
- Adapt examples for your specific needs
- Check get_service_types for latest versions`,
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
		return getGeneralKnowledge(runtime), nil
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

func getGeneralKnowledge(runtime string) interface{} {
	return map[string]interface{}{
		"runtime": runtime,
		"message": fmt.Sprintf("Use the Node.js example as reference pattern for '%s'", runtime),
		"reference": "Check 'nodejs' in knowledge_base for complete zerops.yml example",
		"available_runtimes": []string{
			"nodejs", "python", "go", "php",
			"postgresql", "mariadb", "mongodb",
			"redis", "valkey", "keydb",
			"elasticsearch", "rabbitmq",
			"nginx", "static", "objectstorage",
		},
		"general_pattern": `services:
  - hostname: service-name
    type: runtime@version
    mode: NON_HA  # or HA
    enableSubdomainAccess: true  # for web services
    minContainers: 1
    maxContainers: 3`,
		"deployment_reference": "Adapt the Node.js zerops.yml pattern - change runtime, build commands, and start commands for your specific technology",
		"tips": []string{
			"Use knowledge_base('nodejs') to see complete zerops.yml example",
			"Adapt Node.js pattern for your runtime",
			"Check 'get_service_types' for available types",
			"Use mode: HA for production databases",
			"Enable subdomain for web services",
		},
	}
}