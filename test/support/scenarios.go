package support

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverScenarios finds all scenario directories in a specific folder
// It looks for directories containing a template file matching the pattern: {prefix}-{scenario}-template.yaml
// folderPath should be the path to a folder containing scenario directories (e.g., "../../scenarios/rhtas")
// prefix is the prefix used in template filenames (e.g., "rhtas", "ctlog")
func DiscoverScenarios(folderPath, prefix string) ([]string, error) {
	var scenarios []string

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains a template file
			scenarioName := entry.Name()
			templatePattern := fmt.Sprintf("%s-%s-template.yaml", prefix, scenarioName)
			templatePath := filepath.Join(folderPath, scenarioName, templatePattern)

			if _, err := os.Stat(templatePath); err == nil {
				scenarios = append(scenarios, scenarioName)
			}
		}
	}

	return scenarios, nil
}

// DiscoverVariants finds all variant names for a scenario by looking for {prefix}-{scenario}-{variant}.conf files
// scenarioPath: Path to the scenario directory (e.g., "../../scenarios/rhtas/default")
// prefix: Prefix used in filenames (e.g., "rhtas", "ctlog")
// scenarioName: Name of the scenario (e.g., "default", "simple")
// Returns a list of variant names found
func DiscoverVariants(scenarioPath, prefix, scenarioName string) ([]string, error) {
	var variants []string
	basePattern := fmt.Sprintf("%s-%s-", prefix, scenarioName)
	confSuffix := ".conf"

	entries, err := os.ReadDir(scenarioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := entry.Name()
			// Check if it's a conf file matching the pattern: {prefix}-{scenario}-{variant}.conf
			if strings.HasPrefix(fileName, basePattern) && strings.HasSuffix(fileName, confSuffix) {
				// Extract variant name: remove prefix and suffix
				variant := strings.TrimPrefix(fileName, basePattern)
				variant = strings.TrimSuffix(variant, confSuffix)
				if variant != "" {
					variants = append(variants, variant)
				}
			}
		}
	}

	return variants, nil
}

// ScenarioVariant represents a scenario with its variant
type ScenarioVariant struct {
	FolderName   string
	ScenarioName string
	VariantName  string
}

// DiscoverAllScenarios finds all scenarios and their variants across all folder structures
// It discovers top-level folders (e.g., "rhtas", "tuf") and then finds scenarios and variants within each
// scenariosDir should be the path to the scenarios directory (e.g., "../../scenarios")
// Returns a list of ScenarioVariant structs
func DiscoverAllScenarios(scenariosDir string) ([]ScenarioVariant, error) {
	var scenarioVariants []ScenarioVariant

	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folderName := entry.Name()
			folderPath := filepath.Join(scenariosDir, folderName)

			// Use folder name as prefix for template files (e.g., "rhtas" -> "rhtas-*.yaml", "tuf" -> "tuf-*.yaml")
			prefix := folderName

			// Discover scenarios within this folder
			scenarios, err := DiscoverScenarios(folderPath, prefix)
			if err != nil {
				// Skip folders that don't contain scenarios
				continue
			}

			// For each scenario, discover its variants
			for _, scenarioName := range scenarios {
				scenarioPath := filepath.Join(folderPath, scenarioName)
				variants, err := DiscoverVariants(scenarioPath, prefix, scenarioName)
				if err != nil {
					// Skip scenarios where we can't read variants
					continue
				}

				// If no variants found, skip this scenario
				if len(variants) == 0 {
					continue
				}

				// Add each variant as a separate scenario variant
				for _, variantName := range variants {
					scenarioVariants = append(scenarioVariants, ScenarioVariant{
						FolderName:   folderName,
						ScenarioName: scenarioName,
						VariantName:  variantName,
					})
				}
			}
		}
	}

	return scenarioVariants, nil
}

// LogFoundTemplates logs the list of found template files for discovered scenarios and variants
// scenarioVariants: List of scenario variants discovered
// scenariosDir: Base directory containing scenario directories (e.g., "../../scenarios")
func LogFoundTemplates(scenarioVariants []ScenarioVariant, scenariosDir string) {
	fmt.Printf("Found %d scenario variant(s):\n", len(scenarioVariants))

	for _, sv := range scenarioVariants {
		// Use folder name as prefix (e.g., "rhtas", "ctlog", "tuf")
		baseName := fmt.Sprintf("%s-%s", sv.FolderName, sv.ScenarioName)
		templateFile := baseName + "-template.yaml"
		confFile := baseName + "-" + sv.VariantName + ".conf"
		outputFile := baseName + "-" + sv.VariantName + "-scenario.yaml"

		// Show relative path from project root (normalize scenariosDir to remove ../..)
		scenarioPath := filepath.Join(scenariosDir, sv.FolderName, sv.ScenarioName)
		// Clean the path to show relative path from project root
		scenarioPathClean := filepath.Clean(scenarioPath)
		// Extract relative path (remove any leading ../..)
		scenarioPathRel := scenarioPathClean
		if strings.HasPrefix(scenarioPathClean, "..") {
			// Count how many ../.. and remove them
			parts := strings.Split(scenarioPathClean, string(filepath.Separator))
			var relParts []string
			for i, part := range parts {
				if part == ".." {
					continue
				}
				relParts = parts[i:]
				break
			}
			scenarioPathRel = filepath.Join(relParts...)
		}
		fmt.Printf("  %s %s + %s -> %s\n", scenarioPathRel, templateFile, confFile, outputFile)
	}
}
