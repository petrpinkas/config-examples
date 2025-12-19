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

// DiscoverAllScenarios finds all scenarios across all folder structures in the scenarios directory
// It discovers top-level folders (e.g., "rhtas", "operator") and then finds scenarios within each
// scenariosDir should be the path to the scenarios directory (e.g., "../../scenarios")
// Returns a map where keys are folder names and values are lists of scenario names
// Example: {"rhtas": ["basic", "simple"], "operator": ["default"]}
func DiscoverAllScenarios(scenariosDir string) (map[string][]string, error) {
	folderScenarios := make(map[string][]string)

	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folderName := entry.Name()
			folderPath := filepath.Join(scenariosDir, folderName)

			// Use folder name as prefix for template files (e.g., "rhtas" -> "rhtas-*.yaml", "ctlog" -> "ctlog-*.yaml")
			prefix := folderName

			// Discover scenarios within this folder
			scenarios, err := DiscoverScenarios(folderPath, prefix)
			if err != nil {
				// Skip folders that don't contain scenarios
				continue
			}

			if len(scenarios) > 0 {
				folderScenarios[folderName] = scenarios
			}
		}
	}

	return folderScenarios, nil
}

// LogFoundTemplates logs the list of found template files for discovered scenarios
// folderScenarios: Map of folder names to scenario names (e.g., {"rhtas": ["basic", "simple"]})
// scenariosDir: Base directory containing scenario directories (e.g., "../../scenarios")
// variantName: Variant name for the configuration (e.g., "default")
func LogFoundTemplates(folderScenarios map[string][]string, scenariosDir string, variantName string) {
	totalScenarios := 0
	for _, scenarios := range folderScenarios {
		totalScenarios += len(scenarios)
	}
	fmt.Printf("Found %d scenario(s) across %d folder(s):\n", totalScenarios, len(folderScenarios))

	for folderName, scenarios := range folderScenarios {
		for _, scenarioName := range scenarios {
			// Use folder name as prefix (e.g., "rhtas", "ctlog")
			baseName := fmt.Sprintf("%s-%s", folderName, scenarioName)
			templateFile := baseName + "-template.yaml"
			confFile := baseName + "-" + variantName + ".conf"
			outputFile := baseName + "-" + variantName + "-scenario.yaml"

			// Show relative path from project root (normalize scenariosDir to remove ../..)
			scenarioPath := filepath.Join(scenariosDir, folderName, scenarioName)
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
}
