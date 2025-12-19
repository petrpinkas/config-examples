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
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Initialize tests for all discovered scenarios at package level
// This creates parametrized tests where each scenario variant is a parameter
// Supports multiple folder structures in scenarios/ (e.g., scenarios/rhtas/, scenarios/tuf/, etc.)
// and multiple variants per scenario (e.g., "base", "nomonitoring")
func init() {
	scenariosDir := filepath.Join("..", "..", "scenarios")
	scenarioVariants, err := support.DiscoverAllScenarios(scenariosDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to discover scenarios: %v", err))
	}
	if len(scenarioVariants) == 0 {
		panic("No scenarios found")
	}

	// Log found template files
	support.LogFoundTemplates(scenarioVariants, scenariosDir)

	// Create parametrized tests for each scenario variant
	for _, sv := range scenarioVariants {
		testScenario(sv.FolderName, sv.ScenarioName, sv.VariantName)
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
	resourceKind     string
	resourceGVK      schema.GroupVersionKind
	dryRun           bool
}

// setupScenario performs all setup steps for a scenario variant
// folderName: Top-level folder name (e.g., "rhtas", "tuf")
// scenarioName: Scenario name within the folder (e.g., "default", "simple")
// variantName: Variant name (e.g., "base", "nomonitoring")
func setupScenario(ctx SpecContext, folderName, scenarioName, variantName string) *scenarioTestContext {
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
	scenariosDir := filepath.Join("..", "..", "scenarios", folderName)
	var err error
	testCtx.configPath, err = config.ProcessScenarioTemplate(
		scenarioName,
		scenariosDir,
		testCtx.namespace.Name,
		"securesign-sample",
		variantName,
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to process template")
	fmt.Printf("Processing scenario: %s (%s) in namespace: %s\n", scenarioName, testCtx.configPath, testCtx.namespace.Name)

	// Load configuration
	testCtx.securesignConfig, err = config.LoadConfig(testCtx.configPath)
	Expect(err).NotTo(HaveOccurred())
	testCtx.securesignName = testCtx.securesignConfig.GetName()
	testCtx.resourceKind = testCtx.securesignConfig.GetKind()

	// Extract GVK from config for generic verification
	group, version, kind := testCtx.securesignConfig.GetGroupVersionKind()
	testCtx.resourceGVK = schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}

	if testCtx.dryRun {
		fmt.Printf("DRY RUN: Would install %s: %s in namespace: %s\n", testCtx.resourceKind, testCtx.securesignName, testCtx.namespace.Name)
	} else {
		fmt.Printf("Installing %s: %s in namespace: %s\n", testCtx.resourceKind, testCtx.securesignName, testCtx.namespace.Name)

		// Install the configuration (works generically for any Kubernetes resource)
		err = installer.InstallConfig(ctx, testCtx.k8sClient, testCtx.securesignConfig)
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("%s CR created, waiting for installation...\n", testCtx.resourceKind)

		// Register cleanup: Delete resource first, then namespace
		DeferCleanup(func(ctx SpecContext) {
			// Delete resource using GVK from config
			obj := verifier.Get(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName, testCtx.resourceGVK)
			if obj != nil {
				fmt.Printf("Deleting %s CR: %s/%s\n", testCtx.resourceKind, testCtx.namespace.Name, testCtx.securesignName)
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
// folderName: Top-level folder name (e.g., "rhtas", "ctlog")
// scenarioName: Scenario name within the folder (e.g., "basic", "simple")
func testScenario(folderName, scenarioName, variantName string) {
	// Use folder name as prefix for YAML filename (e.g., "rhtas-default-base", "tuf-simple-nomonitoring")
	yamlFileName := fmt.Sprintf("%s-%s-%s-scenario.yaml", folderName, scenarioName, variantName)
	scenarioPath := fmt.Sprintf("%s/%s", folderName, yamlFileName)
	Describe(fmt.Sprintf("Scenario %s", scenarioPath), Ordered, func() {
		var testCtx *scenarioTestContext

		BeforeAll(func(ctx SpecContext) {
			testCtx = setupScenario(ctx, folderName, scenarioName, variantName)
		})

		Describe("Config Loading", func() {
			It("should load the configuration file", func() {
				Expect(testCtx.securesignConfig).NotTo(BeNil())
				Expect(testCtx.securesignConfig.Data).NotTo(BeNil())
			})

			It("should have correct resource type", func() {
				// Verify kind and apiVersion are present (values depend on the scenario)
				Expect(testCtx.securesignConfig.GetKind()).NotTo(BeEmpty())
				Expect(testCtx.securesignConfig.GetAPIVersion()).NotTo(BeEmpty())
			})

			It("should have metadata", func() {
				// Verify name and namespace are present (name may vary by scenario)
				Expect(testCtx.securesignConfig.GetName()).NotTo(BeEmpty())
				Expect(testCtx.securesignConfig.GetNamespace()).To(Equal(testCtx.namespace.Name))
			})

			It("should have spec section", func() {
				// Verify spec exists (may be empty for some resources)
				spec, ok := testCtx.securesignConfig.Data["spec"]
				Expect(ok).To(BeTrue())
				Expect(spec).NotTo(BeNil())
			})
		})

		Describe("Resource Installation", func() {
			It("should install CR successfully", func(ctx SpecContext) {
				if testCtx.dryRun {
					fmt.Printf("DRY RUN: Skipping CR verification (would check: %s/%s)\n", testCtx.namespace.Name, testCtx.securesignName)
					return
				}
				// Verify the CR exists using GVK from config
				obj := verifier.Get(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName, testCtx.resourceGVK)
				Expect(obj).NotTo(BeNil())
				fmt.Printf("%s CR found: %s/%s\n", testCtx.resourceKind, testCtx.namespace.Name, testCtx.securesignName)
			})

			It("should wait for resource to be ready", func(ctx SpecContext) {
				if testCtx.dryRun {
					fmt.Printf("DRY RUN: Skipping readiness verification (would wait for: %s/%s)\n", testCtx.namespace.Name, testCtx.securesignName)
					return
				}
				fmt.Printf("Waiting for %s %s/%s to be ready...\n", testCtx.resourceKind, testCtx.namespace.Name, testCtx.securesignName)
				verifier.Verify(ctx, testCtx.k8sClient, testCtx.namespace.Name, testCtx.securesignName, testCtx.resourceGVK)
				fmt.Printf("%s %s/%s is ready!\n", testCtx.resourceKind, testCtx.namespace.Name, testCtx.securesignName)
			})
		})
	})
}
