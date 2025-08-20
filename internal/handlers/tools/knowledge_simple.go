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
		Description: "Get Zerops YAML examples and patterns for specified runtime",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"runtime": map[string]interface{}{
					"type":        "string",
					"description": "Runtime type (e.g., 'nodejs', 'python', 'go', 'php', 'postgresql')",
				},
			},
			"required": []string{"runtime"},
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
  - setup: app
    build:
      base: nodejs@22
      buildCommands:
        - npm ci
      deployFiles: ./
    run:
      initCommands:
        - npm run migrate
      start: npm start`,
		"tips": []string{
			"Use nodejs@22 for latest LTS version",
			"Enable subdomain for public access",
			"Set minContainers: 2 for high availability",
			"Use environment variables for configuration",
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
			"django": `services:
  - hostname: django
    type: python@3.12
    enableSubdomainAccess: true
    minContainers: 1
    ports:
      - port: 8000
        httpSupport: true
  - hostname: db
    type: postgresql@16
    mode: NON_HA`,
		},
		"deployment_yaml": `zerops:
  - setup: app
    build:
      base: python@3.12
      buildCommands:
        - pip install -r requirements.txt
      deployFiles: ./
    run:
      initCommands:
        - python manage.py migrate
      start: gunicorn app.wsgi:application`,
		"tips": []string{
			"Use python@3.12 for latest stable version",
			"Include requirements.txt for dependencies",
			"Use gunicorn or uvicorn for production",
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
		"deployment_yaml": `zerops:
  - setup: app
    build:
      base: go@1.22
      buildCommands:
        - go mod download
        - go build -o app ./cmd/main.go
      deployFiles:
        - app
    run:
      start: ./app`,
		"tips": []string{
			"Use go@1.22 for latest version",
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
		"deployment_yaml": `zerops:
  - setup: app
    build:
      base: php@8.3
      buildCommands:
        - composer install --no-dev
      deployFiles: ./
    run:
      initCommands:
        - php artisan migrate --force
      documentRoot: public`,
		"tips": []string{
			"Use php@8.3 for latest stable version",
			"Set documentRoot for frameworks",
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
		"message": fmt.Sprintf("No specific examples for '%s'", runtime),
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
		"tips": []string{
			"Check 'get_service_types' for available types",
			"Use mode: HA for production databases",
			"Enable subdomain for web services",
			"Set container scaling for auto-scaling",
		},
	}
}