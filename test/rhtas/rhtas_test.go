package rhtas

import (
	"fmt"
	"path/filepath"

	"github.com/petrpinkas/config-examples/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Scenario", Ordered, func() {
	var configPath string
	var scenarioName string

	BeforeAll(func() {
		scenarioName = "basic"
		// Path to the basic scenario config file
		configPath = filepath.Join("..", "..", "scenarios", scenarioName, "rhtas-basic.yaml")
		fmt.Printf("Processing scenario: %s (%s)\n", scenarioName, configPath)
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
