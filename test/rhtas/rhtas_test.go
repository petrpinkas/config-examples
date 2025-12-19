package rhtas

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/petrpinkas/config-examples/pkg/config"
	"github.com/petrpinkas/config-examples/pkg/installer"
	"github.com/petrpinkas/config-examples/pkg/kubernetes"
	"github.com/petrpinkas/config-examples/pkg/verifier"
	"github.com/petrpinkas/config-examples/test/support"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// discoverScenarios finds all scenario directories in the scenarios folder
func discoverScenarios() ([]string, error) {
	scenariosDir := filepath.Join("..", "..", "scenarios")
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

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

// testScenario creates a test for a specific scenario
func testScenario(scenarioName string) {
	Describe(fmt.Sprintf("%s Scenario", capitalizeFirst(scenarioName)), Ordered, func() {
		var configPath string
		var k8sClient client.Client
		var namespace *v1.Namespace
		var securesignConfig *config.Config
		var securesignName string

		BeforeAll(func(ctx SpecContext) {
			var err error
			k8sClient, err = kubernetes.GetClient()
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient).NotTo(BeNil())
		})

		BeforeAll(func(ctx SpecContext) {
			namespace = support.CreateTestNamespace(ctx, k8sClient)
		})

		BeforeAll(func() {
			scenarioDir := filepath.Join("..", "..", "scenarios", scenarioName)
			baseName := fmt.Sprintf("rhtas-%s", scenarioName)
			variantName := "default"

			// Create runtime context with standard placeholders
			runtimeCtx := &config.RuntimeContext{
				Namespace:    namespace.Name,
				InstanceName: "securesign-sample",
			}

			// Process template with conf file to generate the final YAML
			var err error
			configPath, err = config.ProcessTemplateFromPaths(scenarioDir, baseName, variantName, runtimeCtx)
			Expect(err).NotTo(HaveOccurred(), "Failed to process template")
			fmt.Printf("Processing scenario: %s (%s) in namespace: %s\n", scenarioName, configPath, namespace.Name)
		})

		BeforeAll(func(ctx SpecContext) {
			var err error
			securesignConfig, err = config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			// Namespace and instance name are already set via placeholders during template processing
			securesignName = securesignConfig.GetName()
			fmt.Printf("Installing Securesign: %s in namespace: %s\n", securesignName, namespace.Name)
		})

		BeforeAll(func(ctx SpecContext) {
			// Install the Securesign configuration
			err := installer.InstallConfig(ctx, k8sClient, securesignConfig)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("Securesign CR created, waiting for installation...\n")

			// Register cleanup: Delete Securesign CR first, then namespace
			DeferCleanup(func(ctx SpecContext) {
				// Delete Securesign CR
				obj := verifier.Get(ctx, k8sClient, namespace.Name, securesignName)
				if obj != nil {
					fmt.Printf("Deleting Securesign CR: %s/%s\n", namespace.Name, securesignName)
					Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
				}

				// Delete namespace
				fmt.Printf("Deleting test namespace: %s\n", namespace.Name)
				Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
			})
		})

		Describe("Config Loading", func() {
			It("should load the configuration file", func() {
				Expect(securesignConfig).NotTo(BeNil())
				Expect(securesignConfig.Data).NotTo(BeNil())
			})

			It("should have correct resource type", func() {
				Expect(securesignConfig.GetKind()).To(Equal("Securesign"))
				Expect(securesignConfig.GetAPIVersion()).To(Equal("rhtas.redhat.com/v1alpha1"))
			})

			It("should have metadata", func() {
				Expect(securesignConfig.GetName()).To(Equal("securesign-sample"))
				Expect(securesignConfig.GetNamespace()).To(Equal(namespace.Name))
			})

			It("should have spec section", func() {
				spec, ok := securesignConfig.Data["spec"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(spec).NotTo(BeNil())
			})

			It("should have fulcio configuration in spec", func() {
				spec, ok := securesignConfig.Data["spec"].(map[string]interface{})
				Expect(ok).To(BeTrue())

				fulcio, ok := spec["fulcio"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(fulcio).NotTo(BeNil())
			})
		})

		Describe("Securesign Installation", func() {
			It("should install Securesign CR successfully", func(ctx SpecContext) {
				// Verify the CR exists
				obj := verifier.Get(ctx, k8sClient, namespace.Name, securesignName)
				Expect(obj).NotTo(BeNil())
				fmt.Printf("Securesign CR found: %s/%s\n", namespace.Name, securesignName)
			})

			It("should wait for Securesign to be ready", func(ctx SpecContext) {
				fmt.Printf("Waiting for Securesign %s/%s to be ready...\n", namespace.Name, securesignName)
				verifier.Verify(ctx, k8sClient, namespace.Name, securesignName)
				fmt.Printf("Securesign %s/%s is ready!\n", namespace.Name, securesignName)
			})
		})
	})
}

// Initialize tests for all discovered scenarios at package level
func init() {
	scenarios, err := discoverScenarios()
	if err != nil {
		panic(fmt.Sprintf("Failed to discover scenarios: %v", err))
	}
	if len(scenarios) == 0 {
		panic("No scenarios found")
	}

	for _, scenarioName := range scenarios {
		testScenario(scenarioName)
	}
}
