package manifest

// Manifest represents the aigogo.json configuration (v2)
type Manifest struct {
	Schema       string        `json:"$schema,omitempty"`
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Description  string        `json:"description,omitempty"`
	Author       string        `json:"author,omitempty"`
	Language     Language      `json:"language"`
	Dependencies *Dependencies `json:"dependencies,omitempty"`
	Files        FileSpec      `json:"files"`
	Metadata     Metadata      `json:"metadata,omitempty"`
	AI           *AISpec       `json:"ai,omitempty"`
}

// Language specifies the programming language and version requirements
type Language struct {
	Name    string `json:"name"`              // Required: python, javascript, go, rust
	Runtime string `json:"runtime,omitempty"` // Optional: cpython, pypy, node, deno, bun
	Version string `json:"version"`           // Required: >=3.8,<4.0
}

// Dependencies specifies runtime and development dependencies
type Dependencies struct {
	Runtime []Dependency `json:"runtime,omitempty"`
	Dev     []Dependency `json:"dev,omitempty"`
}

// Dependency represents a single package dependency
type Dependency struct {
	Package  string `json:"package"`
	Version  string `json:"version"`
	Optional bool   `json:"optional,omitempty"`
}

// FileSpec defines which files to include/exclude
type FileSpec struct {
	Include interface{} `json:"include,omitempty"` // string "auto" or []string patterns
	Exclude []string    `json:"exclude,omitempty"`
}

// Metadata holds optional package metadata
type Metadata struct {
	License  string            `json:"license,omitempty"`
	Homepage string            `json:"homepage,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// AISpec provides metadata for AI agent discovery and usage
type AISpec struct {
	Summary      string   `json:"summary"`           // What this snippet does (one sentence)
	Capabilities []string `json:"capabilities"`      // List of actions the code can perform
	Usage        string   `json:"usage,omitempty"`   // Example import and call
	Inputs       string   `json:"inputs,omitempty"`  // Description of expected inputs
	Outputs      string   `json:"outputs,omitempty"` // Description of return values
}

// GetIncludePatterns returns include patterns as string slice
func (f *FileSpec) GetIncludePatterns() ([]string, bool) {
	if f.Include == nil {
		return nil, false
	}

	// Check if it's "auto"
	if str, ok := f.Include.(string); ok {
		if str == "auto" {
			return nil, true // true = auto-discovery
		}
		return []string{str}, false
	}

	// Check if it's array
	if arr, ok := f.Include.([]interface{}); ok {
		patterns := make([]string, len(arr))
		for i, v := range arr {
			patterns[i] = v.(string)
		}
		return patterns, false
	}

	return nil, false
}

// SupportedLanguages returns list of supported language names
func SupportedLanguages() []string {
	return []string{"python", "javascript", "go", "rust"}
}

// ValidateLanguage checks if language name is supported
func ValidateLanguage(name string) bool {
	for _, lang := range SupportedLanguages() {
		if lang == name {
			return true
		}
	}
	return false
}
