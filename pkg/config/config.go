package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents a generic Kubernetes resource configuration
// It uses map[string]interface{} to be flexible with different resource structures
type Config struct {
	Data map[string]interface{}
}

// RuntimeContext holds runtime values that can be used as placeholders in templates
// These are standard values that are available for all test scenarios
type RuntimeContext struct {
	Namespace     string
	InstanceName  string
	// Future: Timestamp, TestID, etc.
}

// LoadConfig loads a YAML configuration file
// It returns a generic Config that can handle any Kubernetes resource structure
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &Config{Data: configData}, nil
}

// UpdateConfig updates a configuration value using dot-notation path
// Example: spec.fulcio.config.OIDCIssuers.Issuer=value
// Example: metadata.name=my-resource
// Example: metadata.labels.app=myapp
func UpdateConfig(config *Config, pathValue string) error {
	parts := strings.SplitN(pathValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid path=value format: %s", pathValue)
	}

	path := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	pathParts := strings.Split(path, ".")
	if len(pathParts) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	// Navigate through the config map using the path
	return updateNestedMap(config.Data, pathParts, value)
}

func updateNestedMap(m map[string]interface{}, path []string, value string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	key := path[0]
	if len(path) == 1 {
		// Leaf node - set the value
		m[key] = value
		return nil
	}

	// Navigate deeper
	current, exists := m[key]
	if !exists {
		// Create nested map if it doesn't exist
		m[key] = make(map[string]interface{})
		current = m[key]
	}

	nextMap, ok := current.(map[string]interface{})
	if !ok {
		// If it's not a map, replace with a new map
		// This allows overwriting non-map values with nested structures
		m[key] = make(map[string]interface{})
		nextMap = m[key].(map[string]interface{})
	}

	return updateNestedMap(nextMap, path[1:], value)
}

// ToYAML converts the config back to YAML
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c.Data)
}

// GetKind returns the kind of the Kubernetes resource
func (c *Config) GetKind() string {
	if kind, ok := c.Data["kind"].(string); ok {
		return kind
	}
	return ""
}

// GetAPIVersion returns the apiVersion of the Kubernetes resource
func (c *Config) GetAPIVersion() string {
	if apiVersion, ok := c.Data["apiVersion"].(string); ok {
		return apiVersion
	}
	return ""
}

// GetGroupVersionKind returns the GroupVersionKind from the config
// Parses apiVersion (e.g., "rhtas.redhat.com/v1alpha1") and kind (e.g., "Securesign", "CTlog")
func (c *Config) GetGroupVersionKind() (string, string, string) {
	apiVersion := c.GetAPIVersion()
	kind := c.GetKind()
	
	// Parse apiVersion: "group/version" or just "version"
	var group, version string
	if apiVersion != "" {
		parts := strings.Split(apiVersion, "/")
		if len(parts) == 2 {
			group = parts[0]
			version = parts[1]
		} else {
			version = parts[0]
		}
	}
	
	return group, version, kind
}

// GetName returns the name from metadata
func (c *Config) GetName() string {
	if metadata, ok := c.Data["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			return name
		}
	}
	return ""
}

// GetNamespace returns the namespace from metadata
func (c *Config) GetNamespace() string {
	if metadata, ok := c.Data["metadata"].(map[string]interface{}); ok {
		if namespace, ok := metadata["namespace"].(string); ok {
			return namespace
		}
	}
	return ""
}

// FindConfigFiles finds all YAML config files in a directory
func FindConfigFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml")) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// LoadConfFile loads a .conf file with key=value pairs
// Returns a map of key to value
func LoadConfFile(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read conf file: %w", err)
	}

	config := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format in conf file at line %d: %s (expected key=value)", i+1, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in conf file at line %d", i+1)
		}

		config[key] = value
	}

	return config, nil
}

// ProcessTemplate processes a template YAML file with values from a conf file
// templatePath: path to the template YAML file (e.g., "rhtas-basic-template.yaml")
// confPath: path to the conf file (e.g., "rhtas-basic-default.conf")
// outputPath: path where the processed YAML will be written (e.g., "rhtas-basic-default.yaml")
// runtimeCtx: runtime context with standard placeholders (Namespace, InstanceName, etc.)
func ProcessTemplate(templatePath, confPath, outputPath string, runtimeCtx *RuntimeContext) error {
	// Load conf file
	confValues, err := LoadConfFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to load conf file: %w", err)
	}

	// Replace runtime placeholders in conf values first
	// This allows conf files to use {{NAMESPACE}}, {{INSTANCE_NAME}}, etc.
	confValues = replaceRuntimePlaceholdersInMap(confValues, runtimeCtx)

	// Load template YAML as raw string
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// First pass: Replace runtime placeholders in raw YAML string
	// This must happen before parsing because {{PLACEHOLDER}} is not valid YAML syntax
	templateDataStr := string(templateData)
	templateDataStr = replaceRuntimePlaceholdersInString(templateDataStr, runtimeCtx)

	// Split by YAML document separator (---) to handle multi-document YAML files
	documents := splitYAMLDocuments(templateDataStr)
	var processedDocs [][]byte

	for i, docStr := range documents {
		if strings.TrimSpace(docStr) == "" {
			continue // Skip empty documents
		}

		// Parse each document as YAML to work with structured data
		var templateConfig map[string]interface{}
		if err := yaml.Unmarshal([]byte(docStr), &templateConfig); err != nil {
			return fmt.Errorf("failed to parse template YAML document %d: %w", i+1, err)
		}

		// Second pass: Replace template placeholders (like 'https://your-oidc-issuer-url')
		// The placeholder in YAML is 'https://your-oidc-issuer-url' but after unmarshaling it becomes https://your-oidc-issuer-url
		placeholder := "https://your-oidc-issuer-url"
		replaceTemplatePlaceholders(templateConfig, placeholder, confValues)

		// Convert back to YAML
		docYAML, err := yaml.Marshal(templateConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal processed YAML document %d: %w", i+1, err)
		}
		processedDocs = append(processedDocs, docYAML)
	}

	// Join all documents with --- separator
	outputData := bytes.Join(processedDocs, []byte("---\n"))

	// Write output file
	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// splitYAMLDocuments splits a multi-document YAML string into individual documents
// Documents are separated by "---" on a line by itself (optionally with leading/trailing whitespace)
func splitYAMLDocuments(content string) []string {
	// Split by "---" separator
	// We need to handle cases where --- appears at the start, middle, or end
	lines := strings.Split(content, "\n")
	var documents []string
	var currentDoc []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if this line is a document separator (--- on its own line)
		if trimmed == "---" {
			// Save current document if it has content
			if len(currentDoc) > 0 {
				docStr := strings.Join(currentDoc, "\n")
				if strings.TrimSpace(docStr) != "" {
					documents = append(documents, docStr)
				}
				currentDoc = []string{}
			}
			// Skip the separator line itself
			continue
		}
		currentDoc = append(currentDoc, line)
		
		// If this is the last line, save the current document
		if i == len(lines)-1 && len(currentDoc) > 0 {
			docStr := strings.Join(currentDoc, "\n")
			if strings.TrimSpace(docStr) != "" {
				documents = append(documents, docStr)
			}
		}
	}

	// If no separators found, return the whole content as a single document
	if len(documents) == 0 {
		return []string{content}
	}

	return documents
}

// replaceRuntimePlaceholdersInString replaces {{PLACEHOLDER}} patterns in a raw string
// This handles placeholders like {{NAMESPACE}}, {{INSTANCE_NAME}}, etc.
// This must be called before YAML parsing because {{PLACEHOLDER}} is not valid YAML syntax
func replaceRuntimePlaceholdersInString(content string, runtimeCtx *RuntimeContext) string {
	if runtimeCtx == nil {
		return content
	}

	placeholderRegex := regexp.MustCompile(`\{\{(\w+)\}\}`)

	return placeholderRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract placeholder name (e.g., "NAMESPACE" from "{{NAMESPACE}}")
		placeholderName := strings.Trim(match, "{}")
		switch placeholderName {
		case "NAMESPACE":
			return runtimeCtx.Namespace
		case "INSTANCE_NAME":
			return runtimeCtx.InstanceName
		default:
			// Unknown placeholder, leave as-is
			return match
		}
	})
}

// replaceRuntimePlaceholdersInMap replaces {{PLACEHOLDER}} patterns in a map of strings
// This is used to process conf file values that may contain runtime placeholders
func replaceRuntimePlaceholdersInMap(values map[string]string, runtimeCtx *RuntimeContext) map[string]string {
	if runtimeCtx == nil {
		return values
	}

	placeholderRegex := regexp.MustCompile(`\{\{(\w+)\}\}`)
	result := make(map[string]string)

	for key, value := range values {
		replaced := placeholderRegex.ReplaceAllStringFunc(value, func(match string) string {
			placeholderName := strings.Trim(match, "{}")
			switch placeholderName {
			case "NAMESPACE":
				return runtimeCtx.Namespace
			case "INSTANCE_NAME":
				return runtimeCtx.InstanceName
			default:
				return match
			}
		})
		result[key] = replaced
	}

	return result
}

// replaceTemplatePlaceholders recursively replaces template placeholder values in the config structure
// It looks for the placeholder string and replaces it with values from confValues
// based on the field name (key in confValues)
// This handles static template placeholders like 'https://your-oidc-issuer-url'
func replaceTemplatePlaceholders(data interface{}, placeholder string, confValues map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			// Check if this is a string value that matches the placeholder
			if strVal, ok := val.(string); ok && strVal == placeholder {
				// Try to find replacement value by key name
				if replacement, exists := confValues[key]; exists {
					v[key] = replacement
				}
			} else {
				// Recursively process nested structures
				replaceTemplatePlaceholders(val, placeholder, confValues)
			}
		}
	case []interface{}:
		for _, item := range v {
			replaceTemplatePlaceholders(item, placeholder, confValues)
		}
	}
}

// ProcessTemplateFromPaths processes a template using scenario name and variant name
// scenarioDir: directory containing the template and conf files (e.g., "scenarios/basic")
// scenarioName: base name of the scenario (e.g., "rhtas-basic")
// variantName: variant name (e.g., "default")
// runtimeCtx: runtime context with standard placeholders (Namespace, InstanceName, etc.)
// Returns the path to the generated YAML file
func ProcessTemplateFromPaths(scenarioDir, scenarioName, variantName string, runtimeCtx *RuntimeContext) (string, error) {
	templatePath := filepath.Join(scenarioDir, scenarioName+"-template.yaml")
	confPath := filepath.Join(scenarioDir, scenarioName+"-"+variantName+".conf")
	outputPath := filepath.Join(scenarioDir, scenarioName+"-"+variantName+"-scenario.yaml")

	fmt.Printf("Processing: %s, %s, %s\n", templatePath, confPath, outputPath)

	if err := ProcessTemplate(templatePath, confPath, outputPath, runtimeCtx); err != nil {
		return "", err
	}

	return outputPath, nil
}

// ProcessScenarioTemplate processes a scenario template with the given runtime context
// It generates the final YAML configuration file from the template and conf file.
// This is a convenience function that constructs the scenario directory and base name
// from the scenario name (assumes "rhtas-{scenarioName}" naming pattern).
//
// Parameters:
//   - scenarioName: Name of the scenario (e.g., "basic", "simple")
//   - scenariosDir: Base directory containing scenario directories (e.g., "../../scenarios")
//   - namespace: Kubernetes namespace name
//   - instanceName: Securesign instance name (default: "securesign-sample")
//   - variantName: Variant name for the configuration (default: "default")
//
// Returns:
//   - configPath: Path to the generated YAML configuration file
//   - error: Any error encountered during processing
func ProcessScenarioTemplate(scenarioName, scenariosDir, namespace, instanceName, variantName string) (string, error) {
	scenarioDir := filepath.Join(scenariosDir, scenarioName)
	// Extract folder name (prefix) from scenariosDir path
	// e.g., "../../scenarios/rhtas" -> "rhtas", "../../scenarios/ctlog" -> "ctlog"
	folderName := filepath.Base(scenariosDir)
	baseName := fmt.Sprintf("%s-%s", folderName, scenarioName)

	runtimeCtx := &RuntimeContext{
		Namespace:    namespace,
		InstanceName: instanceName,
	}

	configPath, err := ProcessTemplateFromPaths(scenarioDir, baseName, variantName, runtimeCtx)
	if err != nil {
		return "", fmt.Errorf("failed to process template for scenario %s: %w", scenarioName, err)
	}

	return configPath, nil
}
