package rhtas

import (
	"fmt"
	"path/filepath"

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

// Initialize tests for all discovered scenarios at package level
// This creates parametrized tests where each scenario is a parameter
func init() {
	scenariosDir := filepath.Join("..", "..", "scenarios")
	scenarios, err := support.DiscoverScenarios(scenariosDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to discover scenarios: %v", err))
	}
	if len(scenarios) == 0 {
		panic("No scenarios found")
	}

	// Log found template files
	support.LogFoundTemplates(scenariosDir, scenarios, "default")

	// Create parametrized tests for each scenario
	for _, scenarioName := range scenarios {
		testScenario(scenarioName)
	}
}

// scenarioTestContext holds the test context for a scenario
type scenarioTestContext struct {
	scenarioName     string
	configPath       string
	k8sClient        client.Client
	namespace        *v1.Namespace
	securesignConfig *config.Config
	securesignName   string
	dryRun           bool
}

// setupScenario performs all setup steps for a scenario
func setupScenario(ctx SpecContext, scenarioName string) *scenarioTestContext {
	testCtx := &scenarioTestContext{
		scenarioName: scenarioName,
		dryRun:       support.IsDryRun(),
	}

	if testCtx.dryRun {
		fmt.Printf("DRY RUN MODE: Skipping OpenShift operations for scenario: %s\n", scenarioName)
		// Create a mock namespace name for dry run
		testCtx.namespace = &v1.Namespace{}
		testCtx.namespace.Name = fmt.Sprintf("dry-run-namespace-%s", scenarioName)
	} else {
		// Initialize Kubernetes client
		var err error
		testCtx.k8sClient, err = kubernetes.GetClient()
		Expect(err).NotTo(HaveOccurred())
		Expect(testCtx.k8sClient).NotTo(BeNil())

		// Create namespace
		testCtx.namespace = support.CreateTestNamespace(ctx, testCtx.k8sClient)
	}

	// Process template with conf file to generate the final YAML
	scenariosDir := filepath.Join("..", "..", "scenarios")
	var err error
	testCtx.configPath, err = config.ProcessScenarioTemplate(
		scenarioName,
		scenariosDir,
		testCtx.namespace.Name,
		"securesign-sample",
		"default",
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to process template")
	fmt.Printf("Processing scenario: %s (%s) in namespace: %s\n", scenarioName, testCtx.configPath, testCtx.namespace.Name)

	// Load configuration
	testCtx.securesignConfig, err = config.LoadConfig(testCtx.configPath)
	Expect(err).NotTo(HaveOccurred())
	testCtx.securesignName = testCtx.securesignConfig.GetName()

	if testCtx.dryRun {
		fmt.Printf("DRY RUN: Would install Securesign: %s in namespace: %s\n", testCtx.securesignName, testCtx.namespace.Name)
	} else {
		fmt.Printf("Installing Securesign: %s in namespace: %s\n", testCtx.securesignName, testCtx.namespace.Name)

		// Install the Securesign configuration
		err = installer.InstallConfig(ctx, testCtx.k8sClient, testCtx.securesignConfig)
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Securesign CR created, waiting for installation...\n")

		// Register cleanup: Delete Securesign CR first, then namespace
		DeferCleanup(func(ctx SpecContext) {
			// Delete Securesign CR
			obj := verifier.Get(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName)
			if obj != nil {
				fmt.Printf("Deleting Securesign CR: %s/%s\n", testCtx.namespace.Name, testCtx.securesignName)
				Expect(testCtx.k8sClient.Delete(ctx, obj)).To(Succeed())
			}

			// Delete namespace
			fmt.Printf("Deleting test namespace: %s\n", testCtx.namespace.Name)
			Expect(testCtx.k8sClient.Delete(ctx, testCtx.namespace)).To(Succeed())
		})
	}

	return testCtx
}

// testScenario creates a test for a specific scenario using parametrized approach
func testScenario(scenarioName string) {
	yamlFileName := fmt.Sprintf("rhtas-%s-default-scenario.yaml", scenarioName)
	Describe(fmt.Sprintf("Scenario %s", yamlFileName), Ordered, func() {
		var testCtx *scenarioTestContext

		BeforeAll(func(ctx SpecContext) {
			testCtx = setupScenario(ctx, scenarioName)
		})

		Describe("Config Loading", func() {
			It("should load the configuration file", func() {
				Expect(testCtx.securesignConfig).NotTo(BeNil())
				Expect(testCtx.securesignConfig.Data).NotTo(BeNil())
			})

			It("should have correct resource type", func() {
				Expect(testCtx.securesignConfig.GetKind()).To(Equal("Securesign"))
				Expect(testCtx.securesignConfig.GetAPIVersion()).To(Equal("rhtas.redhat.com/v1alpha1"))
			})

			It("should have metadata", func() {
				Expect(testCtx.securesignConfig.GetName()).To(Equal("securesign-sample"))
				Expect(testCtx.securesignConfig.GetNamespace()).To(Equal(testCtx.namespace.Name))
			})

			It("should have spec section", func() {
				spec, ok := testCtx.securesignConfig.Data["spec"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(spec).NotTo(BeNil())
			})

			It("should have fulcio configuration in spec", func() {
				spec, ok := testCtx.securesignConfig.Data["spec"].(map[string]interface{})
				Expect(ok).To(BeTrue())

				fulcio, ok := spec["fulcio"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(fulcio).NotTo(BeNil())
			})
		})

		Describe("Securesign Installation", func() {
			It("should install Securesign CR successfully", func(ctx SpecContext) {
				if testCtx.dryRun {
					fmt.Printf("DRY RUN: Skipping CR verification (would check: %s/%s)\n", testCtx.namespace.Name, testCtx.securesignName)
					return
				}
				// Verify the CR exists
				obj := verifier.Get(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName)
				Expect(obj).NotTo(BeNil())
				fmt.Printf("Securesign CR found: %s/%s\n", testCtx.namespace.Name, testCtx.securesignName)
			})

			It("should wait for Securesign to be ready", func(ctx SpecContext) {
				if testCtx.dryRun {
					fmt.Printf("DRY RUN: Skipping readiness verification (would wait for: %s/%s)\n", testCtx.namespace.Name, testCtx.securesignName)
					return
				}
				fmt.Printf("Waiting for Securesign %s/%s to be ready...\n", testCtx.namespace.Name, testCtx.securesignName)
				verifier.Verify(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName)
				fmt.Printf("Securesign %s/%s is ready!\n", testCtx.namespace.Name, testCtx.securesignName)
			})
		})
	})
}
