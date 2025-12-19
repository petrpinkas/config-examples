package support

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverScenarios finds all scenario directories in the scenarios folder
// It looks for directories containing a template file matching the pattern: rhtas-{scenario}-template.yaml
// scenariosDir should be the path to the scenarios directory (e.g., "../../scenarios")
func DiscoverScenarios(scenariosDir string) ([]string, error) {
	var scenarios []string

	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains a template file
			scenarioName := entry.Name()
			templatePattern := fmt.Sprintf("rhtas-%s-template.yaml", scenarioName)
			templatePath := filepath.Join(scenariosDir, scenarioName, templatePattern)

			if _, err := os.Stat(templatePath); err == nil {
				scenarios = append(scenarios, scenarioName)
			}
		}
	}

	return scenarios, nil
}

// LogFoundTemplates logs the list of found template files for discovered scenarios
// scenariosDir: Base directory containing scenario directories (e.g., "../../scenarios")
// scenarios: List of scenario names discovered
// variantName: Variant name for the configuration (e.g., "default")
func LogFoundTemplates(scenariosDir string, scenarios []string, variantName string) {
	fmt.Printf("Found %d scenario(s):\n", len(scenarios))
	for _, scenarioName := range scenarios {
		baseName := fmt.Sprintf("rhtas-%s", scenarioName)
		templateFile := baseName + "-template.yaml"
		confFile := baseName + "-" + variantName + ".conf"
		outputFile := baseName + "-" + variantName + ".yaml"
		
		// Show relative path from project root (scenarios is in active folder)
		scenarioPath := filepath.Join("scenarios", scenarioName)
		fmt.Printf("  %s %s + %s -> %s\n", scenarioPath, templateFile, confFile, outputFile)
	}
}

