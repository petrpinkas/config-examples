package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents a generic Kubernetes resource configuration
// It uses map[string]interface{} to be flexible with different resource structures
type Config struct {
	Data map[string]interface{}
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
func ProcessTemplate(templatePath, confPath, outputPath string) error {
	// Load conf file
	confValues, err := LoadConfFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to load conf file: %w", err)
	}

	// Load template YAML
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template as YAML to work with structured data
	var templateConfig map[string]interface{}
	if err := yaml.Unmarshal(templateData, &templateConfig); err != nil {
		return fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Replace placeholder values in the config structure
	// The placeholder in YAML is 'https://your-oidc-issuer-url' but after unmarshaling it becomes https://your-oidc-issuer-url
	placeholder := "https://your-oidc-issuer-url"
	replacePlaceholders(templateConfig, placeholder, confValues)

	// Convert back to YAML
	outputData, err := yaml.Marshal(templateConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal processed YAML: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// replacePlaceholders recursively replaces placeholder values in the config structure
// It looks for the placeholder string and replaces it with values from confValues
// based on the field name (key in confValues)
func replacePlaceholders(data interface{}, placeholder string, confValues map[string]string) {
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
				replacePlaceholders(val, placeholder, confValues)
			}
		}
	case []interface{}:
		for _, item := range v {
			replacePlaceholders(item, placeholder, confValues)
		}
	}
}

// ProcessTemplateFromPaths processes a template using scenario name and variant name
// scenarioDir: directory containing the template and conf files (e.g., "scenarios/basic")
// scenarioName: base name of the scenario (e.g., "rhtas-basic")
// variantName: variant name (e.g., "default")
// Returns the path to the generated YAML file
func ProcessTemplateFromPaths(scenarioDir, scenarioName, variantName string) (string, error) {
	templatePath := filepath.Join(scenarioDir, scenarioName+"-template.yaml")
	confPath := filepath.Join(scenarioDir, scenarioName+"-"+variantName+".conf")
	outputPath := filepath.Join(scenarioDir, scenarioName+"-"+variantName+".yaml")

	if err := ProcessTemplate(templatePath, confPath, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}
