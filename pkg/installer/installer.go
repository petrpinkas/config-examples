package installer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/petrpinkas/config-examples/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// InstallConfig installs a configuration to the cluster
// It supports both single and multi-document YAML files (separated by ---)
// If filePath is provided, it reads directly from the file to preserve multi-document structure
// Otherwise, it uses cfg.ToYAML() which only contains the first document
func InstallConfig(ctx context.Context, cli client.Client, cfg *config.Config, filePath string) error {
	var yamlData []byte
	var err error

	// If filePath is provided, read directly from file to preserve multi-document structure
	if filePath != "" {
		yamlData, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Fallback to ToYAML() which only contains the first document
		yamlData, err = cfg.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert config to YAML: %w", err)
		}
	}

	// Split multi-document YAML into individual documents
	documents := splitYAMLDocuments(string(yamlData))

	// Install each document as a separate resource
	for i, doc := range documents {
		if strings.TrimSpace(doc) == "" {
			continue // Skip empty documents
		}

		// Unmarshal YAML into unstructured object
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			return fmt.Errorf("failed to unmarshal YAML document %d: %w", i+1, err)
		}

		// Apply the object (Create or Update)
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GroupVersionKind())
		err = cli.Get(ctx, client.ObjectKey{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		}, existing)

		if errors.IsNotFound(err) {
			// Create new resource
			if err := cli.Create(ctx, obj); err != nil {
				return fmt.Errorf("failed to create resource %d (%s/%s): %w", i+1, obj.GetKind(), obj.GetName(), err)
			}
		} else if err == nil {
			// Update existing resource
			obj.SetResourceVersion(existing.GetResourceVersion())
			if err := cli.Update(ctx, obj); err != nil {
				return fmt.Errorf("failed to update resource %d (%s/%s): %w", i+1, obj.GetKind(), obj.GetName(), err)
			}
		} else {
			return fmt.Errorf("failed to check if resource %d exists: %w", i+1, err)
		}
	}

	return nil
}

// splitYAMLDocuments splits a multi-document YAML string into individual documents
// Documents are separated by "---" on a line by itself
func splitYAMLDocuments(content string) []string {
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
