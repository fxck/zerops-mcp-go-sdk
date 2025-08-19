package instructions

// GetWorkflowInstructions returns the comprehensive workflow instructions for Zerops MCP
func GetWorkflowInstructions() string {
	return `
# Zerops MCP Workflow Instructions

## CRITICAL: Always Start with Knowledge Base
BEFORE creating anything, search the knowledge base for proven configurations:
- knowledge_search("technology stack") - Find recipes and patterns
- knowledge_get("recipe/name") - Get complete configurations
- 159+ templates available (Laravel, Django, Next.js, etc.)

## Optimal Workflow

### 1. Research → 2. Plan → 3. Create → 4. Deploy

1. **Search Knowledge First**
   knowledge_search("python postgresql") - Find matching recipes
   knowledge_get("recipe/python") - Get full YAML configuration

2. **Create Project Structure Files**
   - zerops-project-import.yml - Infrastructure definition (all services)
   - zerops.yml - Deployment config (build/run commands)
   IMPORTANT: Setup name in zerops.yml MUST match service name

3. **Execute in Order**
   project_create(name, region) - Create project
   project_import(project_id, yaml) - Import all services
   git init && git add . && git commit -m "Initial" - Initialize git with commit (REQUIRED)
   deploy_push(project_id, service_id) - Deploy code (needs BOTH IDs)

## Key Requirements

- Git: Must have at least one commit (git init + git add . + git commit)
- Deployment: BOTH project_id AND service_id required
- Service IDs: 22-char strings shown as "SERVICE_ID:" in service_list output
- IMPORTANT: Use the exact SERVICE_ID value, NOT the service name (like "app" or "db")
- Setup names: Must match between zerops-project-import.yml and zerops.yml

## Common Patterns

- Web + Database: PostgreSQL/MySQL + Python/Node.js/PHP
- Microservices: Multiple services with shared database
- Full Stack: Frontend (static) + Backend (API) + Database + Cache

## Quick Examples

**Python + PostgreSQL:**
knowledge_search("python postgresql") → project_create → project_import → deploy_push

**Laravel:**
knowledge_search("laravel") → knowledge_get("recipe/laravel-jetstream") → create with PHP + MySQL + Redis

**Node.js Microservices:**
knowledge_search("nodejs") → create multiple services → deploy with matching setup names

## Available Services

**Runtimes:** nodejs, python, php, go, rust, java, dotnet, ruby
**Databases:** postgresql, mysql, mariadb, mongodb, keydb, elasticsearch  
**Tools:** rabbitmq, redis, meilisearch, object-storage
**Utility Services (require buildFromGit):** 
- Adminer (database UI): https://github.com/zeropsio/recipe-adminer
- Adminerevo (enhanced UI): https://github.com/zeropsio/recipe-adminerevo  
- Mailpit (email testing): https://github.com/zeropsio/recipe-mailpit
- S3Browser (S3 UI): https://github.com/zeropsio/recipe-s3browser

Remember: Knowledge base first, then create!
`
}
