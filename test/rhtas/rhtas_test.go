package rhtas

import (
	"path/filepath"

	"github.com/petrpinkas/config-examples/pkg/config"
	"github.com/sirupsen/logrus"

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

		logrus.WithFields(logrus.Fields{
			"scenario": scenarioName,
			"config":   configPath,
		}).Info("Starting scenario test")
	})

	Describe("Config Loading", func() {
		It("should load the basic configuration file", func() {
			logrus.WithField("scenario", scenarioName).Info("Loading configuration file")
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Data).NotTo(BeNil())
			logrus.WithFields(logrus.Fields{
				"scenario": scenarioName,
				"kind":     cfg.GetKind(),
				"name":     cfg.GetName(),
			}).Info("Configuration loaded successfully")
		})

		It("should have correct resource type", func() {
			logrus.WithField("scenario", scenarioName).Info("Verifying resource type")
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.GetKind()).To(Equal("Securesign"))
			Expect(cfg.GetAPIVersion()).To(Equal("rhtas.redhat.com/v1alpha1"))
			logrus.WithFields(logrus.Fields{
				"scenario":   scenarioName,
				"kind":       cfg.GetKind(),
				"apiVersion": cfg.GetAPIVersion(),
			}).Info("Resource type verified")
		})

		It("should have metadata", func() {
			logrus.WithField("scenario", scenarioName).Info("Verifying metadata")
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.GetName()).To(Equal("securesign-sample"))
			Expect(cfg.GetNamespace()).To(Equal("openshift-operators"))
			logrus.WithFields(logrus.Fields{
				"scenario":  scenarioName,
				"name":      cfg.GetName(),
				"namespace": cfg.GetNamespace(),
			}).Info("Metadata verified")
		})

		It("should have spec section", func() {
			logrus.WithField("scenario", scenarioName).Info("Verifying spec section")
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			spec, ok := cfg.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(spec).NotTo(BeNil())
			logrus.WithField("scenario", scenarioName).Info("Spec section verified")
		})

		It("should have fulcio configuration in spec", func() {
			logrus.WithField("scenario", scenarioName).Info("Verifying Fulcio configuration")
			cfg, err := config.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())

			spec, ok := cfg.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())

			fulcio, ok := spec["fulcio"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(fulcio).NotTo(BeNil())
			logrus.WithField("scenario", scenarioName).Info("Fulcio configuration verified")
		})
	})
})
