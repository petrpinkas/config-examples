package rhtas

import (
	"fmt"
	"path/filepath"

	"github.com/petrpinkas/config-examples/pkg/config"
	"github.com/petrpinkas/config-examples/pkg/kubernetes"
	"github.com/petrpinkas/config-examples/test/support"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Basic Scenario", Ordered, func() {
	var configPath string
	var scenarioName string
	var k8sClient client.Client
	var namespace *v1.Namespace

	BeforeAll(func(ctx SpecContext) {
		var err error
		k8sClient, err = kubernetes.GetClient()
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())
	})

	BeforeAll(func(ctx SpecContext) {
		namespace = support.CreateTestNamespace(ctx, k8sClient)
		DeferCleanup(func(ctx SpecContext) {
			GinkgoWriter.Printf("Deleting test namespace: %s\n", namespace.Name)
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
		})
	})

	BeforeAll(func() {
		scenarioName = "basic"
		// Path to the basic scenario config file
		configPath = filepath.Join("..", "..", "scenarios", scenarioName, "rhtas-basic.yaml")
		fmt.Printf("Processing scenario: %s (%s) in namespace: %s\n", scenarioName, configPath, namespace.Name)
	})

	Describe("Config Loading", func() {
		It("should load the basic configuration file", func() {
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Data).NotTo(BeNil())
		})

		It("should have correct resource type", func() {
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.GetKind()).To(Equal("Securesign"))
			Expect(cfg.GetAPIVersion()).To(Equal("rhtas.redhat.com/v1alpha1"))
		})

		It("should have metadata", func() {
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.GetName()).To(Equal("securesign-sample"))
			Expect(cfg.GetNamespace()).To(Equal("openshift-operators"))
		})

		It("should have spec section", func() {
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			spec, ok := cfg.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(spec).NotTo(BeNil())
		})

		It("should have fulcio configuration in spec", func() {
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			spec, ok := cfg.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())

			fulcio, ok := spec["fulcio"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(fulcio).NotTo(BeNil())
		})
	})
})
