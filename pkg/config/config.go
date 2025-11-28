package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SecuresignConfig represents the Securesign custom resource configuration
type SecuresignConfig struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   Metadata               `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// Metadata represents the metadata section of the config
type Metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// LoadConfig loads a YAML configuration file
func LoadConfig(filePath string) (*SecuresignConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SecuresignConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// UpdateConfig updates a configuration value using dot-notation path
// Example: spec.fulcio.config.OIDCIssuers.Issuer=value
func UpdateConfig(config *SecuresignConfig, pathValue string) error {
	parts := strings.SplitN(pathValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid path=value format: %s", pathValue)
	}

	path := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	pathParts := strings.Split(path, ".")
	if len(pathParts) < 2 {
		return fmt.Errorf("path must have at least 2 parts (e.g., spec.fulcio.config): %s", path)
	}

	// Handle metadata updates
	if pathParts[0] == "metadata" {
		return updateMetadata(&config.Metadata, pathParts[1:], value)
	}

	// Handle spec updates
	if pathParts[0] == "spec" {
		return updateNestedMap(config.Spec, pathParts[1:], value)
	}

	return fmt.Errorf("unsupported top-level path: %s (only 'metadata' and 'spec' are supported)", pathParts[0])
}

func updateMetadata(metadata *Metadata, path []string, value string) error {
	if len(path) == 0 {
		return fmt.Errorf("metadata path cannot be empty")
	}

	switch path[0] {
	case "name":
		if len(path) > 1 {
			return fmt.Errorf("metadata.name does not support nested paths")
		}
		metadata.Name = value
	case "namespace":
		if len(path) > 1 {
			return fmt.Errorf("metadata.namespace does not support nested paths")
		}
		metadata.Namespace = value
	case "labels":
		if metadata.Labels == nil {
			metadata.Labels = make(map[string]string)
		}
		if len(path) == 2 {
			metadata.Labels[path[1]] = value
		} else {
			return fmt.Errorf("metadata.labels requires key: metadata.labels.key=value")
		}
	case "annotations":
		if metadata.Annotations == nil {
			metadata.Annotations = make(map[string]string)
		}
		if len(path) == 2 {
			metadata.Annotations[path[1]] = value
		} else {
			return fmt.Errorf("metadata.annotations requires key: metadata.annotations.key=value")
		}
	default:
		return fmt.Errorf("unsupported metadata field: %s", path[0])
	}

	return nil
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
		m[key] = make(map[string]interface{})
		nextMap = m[key].(map[string]interface{})
	}

	return updateNestedMap(nextMap, path[1:], value)
}

// ToYAML converts the config back to YAML
func (c *SecuresignConfig) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
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

