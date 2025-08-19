package tools

// ValidServiceTypes contains all valid Zerops service type strings
// This is used for validation and error messages
var ValidServiceTypes = map[string]bool{
	// Databases
	"postgresql@16":    true,
	"postgresql@15":    true,
	"postgresql@14":    true,
	"mariadb@11":       true,
	"mariadb@10.6":     true,
	"mongodb@7":        true,
	"mongodb@6":        true,
	"valkey@7":         true,
	"keydb@6":          true,
	"elasticsearch@8":  true,
	"couchbase@7":      true,
	"qdrant@1":         true,
	
	// Runtimes - PHP (NOT php-apache!)
	"php@8.3":          true,
	"php@8.2":          true,
	"php@8.1":          true,
	
	// Runtimes - Node.js
	"nodejs@20":        true,
	"nodejs@18":        true,
	"nodejs@16":        true,
	
	// Runtimes - Python
	"python@3.12":      true,
	"python@3.11":      true,
	"python@3.10":      true,
	
	// Runtimes - Go
	"go@1":             true,
	"go@1.21":          true,
	"go@1.20":          true,
	
	// Runtimes - Others
	"java@17":          true,
	"java@11":          true,
	"dotnet@8":         true,
	"dotnet@6":         true,
	"ruby@3":           true,
	"rust@1":           true,
	
	// Message Queues
	"rabbitmq@3.12":    true,
	"kafka@3":          true,
	"nats@2":           true,
	
	// Search
	"meilisearch@1":    true,
	"typesense@26":     true,
	
	// Static
	"static@1":         true,
	
	// Object Storage
	"object-storage@1": true,
}

// GetServiceTypeSuggestion returns a suggestion for an invalid service type
func GetServiceTypeSuggestion(invalidType string) string {
	// Common mistakes and their corrections
	corrections := map[string]string{
		"php-apache":    "php (use php@8.3 with appropriate run configuration)",
		"php-nginx":     "php (use php@8.3 with appropriate run configuration)",
		"redis":         "valkey (Redis-compatible, use valkey@7)",
		"mysql":         "mariadb (MySQL-compatible, use mariadb@11)",
		"postgres":      "postgresql (use postgresql@16)",
		"mongo":         "mongodb (use mongodb@7)",
		"elastic":       "elasticsearch (use elasticsearch@8)",
	}
	
	// Check for exact match in corrections
	for mistake, correction := range corrections {
		if invalidType == mistake || invalidType == mistake+"@8.3" {
			return correction
		}
	}
	
	// Check if it's a version issue (e.g., mongodb@7.0 instead of mongodb@7)
	if idx := len(invalidType) - 2; idx > 0 && invalidType[idx] == '.' {
		// Remove minor version (e.g., @7.0 -> @7)
		suggestion := invalidType[:idx]
		if ValidServiceTypes[suggestion] {
			return "use '" + suggestion + "' (no minor version needed)"
		}
	}
	
	return ""
}