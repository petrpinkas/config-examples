package installer

import (
	"context"
	"fmt"

	"github.com/petrpinkas/config-examples/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// InstallConfig installs a Securesign configuration to the cluster
func InstallConfig(ctx context.Context, cli client.Client, cfg *config.Config) error {
	yamlData, err := cfg.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to convert config to YAML: %w", err)
	}

	// Unmarshal YAML into unstructured object
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(yamlData, &obj.Object); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
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
			return fmt.Errorf("failed to create resource: %w", err)
		}
	} else if err == nil {
		// Update existing resource
		obj.SetResourceVersion(existing.GetResourceVersion())
		if err := cli.Update(ctx, obj); err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check if resource exists: %w", err)
	}

	return nil
}
