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
