package config

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Start from the current test file location
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Package Suite")
}

var _ = Describe("Config Loading", func() {
	It("should load YAML configuration file", func() {
		projectRoot := getProjectRoot()
		configPath := filepath.Join(projectRoot, "scenarios", "basic", "rhtas-basic.yaml")
		config, err := LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())
		Expect(config.Kind).To(Equal("Securesign"))
		Expect(config.APIVersion).To(Equal("rhtas.redhat.com/v1alpha1"))
		Expect(config.Metadata.Name).NotTo(BeEmpty())
		Expect(config.Spec).NotTo(BeNil())
	})

	It("should return error for non-existent file", func() {
		_, err := LoadConfig("non-existent.yaml")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to read config file"))
	})

	It("should return error for invalid YAML", func() {
		// Create a temporary invalid YAML file
		tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("invalid: yaml: content: [")
		Expect(err).NotTo(HaveOccurred())
		tmpFile.Close()

		_, err = LoadConfig(tmpFile.Name())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to parse YAML"))
	})
})

var _ = Describe("Config Updates", func() {
	var config *SecuresignConfig

	BeforeEach(func() {
		projectRoot := getProjectRoot()
		configPath := filepath.Join(projectRoot, "scenarios", "basic", "rhtas-basic.yaml")
		var err error
		config, err = LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should update metadata.name", func() {
		err := UpdateConfig(config, "metadata.name=my-securesign")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Metadata.Name).To(Equal("my-securesign"))
	})

	It("should update metadata.namespace", func() {
		err := UpdateConfig(config, "metadata.namespace=my-namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Metadata.Namespace).To(Equal("my-namespace"))
	})

	It("should update metadata.labels", func() {
		err := UpdateConfig(config, "metadata.labels.test=value")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Metadata.Labels["test"]).To(Equal("value"))
	})

	It("should update nested spec values", func() {
		err := UpdateConfig(config, "spec.fulcio.certificate.commonName=fulcio.example.com")
		Expect(err).NotTo(HaveOccurred())

		fulcio, ok := config.Spec["fulcio"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		certificate, ok := fulcio["certificate"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(certificate["commonName"]).To(Equal("fulcio.example.com"))
	})

	It("should create nested maps when they don't exist", func() {
		err := UpdateConfig(config, "spec.newcomponent.setting=value")
		Expect(err).NotTo(HaveOccurred())

		newComponent, ok := config.Spec["newcomponent"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(newComponent["setting"]).To(Equal("value"))
	})

	It("should return error for invalid path format", func() {
		err := UpdateConfig(config, "invalid-path")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid path=value format"))
	})

	It("should return error for unsupported top-level path", func() {
		err := UpdateConfig(config, "unsupported.field=value")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported top-level path"))
	})
})

var _ = Describe("Config to YAML", func() {
	It("should convert config back to YAML", func() {
		projectRoot := getProjectRoot()
		configPath := filepath.Join(projectRoot, "scenarios", "basic", "rhtas-basic.yaml")
		config, err := LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())

		yamlData, err := config.ToYAML()
		Expect(err).NotTo(HaveOccurred())
		Expect(yamlData).NotTo(BeEmpty())
		Expect(string(yamlData)).To(ContainSubstring("apiVersion"))
		Expect(string(yamlData)).To(ContainSubstring("kind: Securesign"))
	})
})

var _ = Describe("Find Config Files", func() {
	It("should find YAML files in directory", func() {
		projectRoot := getProjectRoot()
		dir := filepath.Join(projectRoot, "scenarios", "basic")
		files, err := FindConfigFiles(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).NotTo(BeEmpty())
		Expect(files).To(ContainElement(ContainSubstring("rhtas-basic.yaml")))
	})

	It("should find both .yaml and .yml files", func() {
		// Create a temporary directory with both extensions
		tmpDir, err := os.MkdirTemp("", "test-configs-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		// Create .yaml file
		yamlFile := filepath.Join(tmpDir, "test.yaml")
		err = os.WriteFile(yamlFile, []byte("test: value"), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Create .yml file
		ymlFile := filepath.Join(tmpDir, "test.yml")
		err = os.WriteFile(ymlFile, []byte("test: value"), 0644)
		Expect(err).NotTo(HaveOccurred())

		files, err := FindConfigFiles(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(2))
	})

	It("should return empty slice for empty directory", func() {
		tmpDir, err := os.MkdirTemp("", "empty-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tmpDir)

		files, err := FindConfigFiles(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeEmpty())
	})

	It("should return error for non-existent directory", func() {
		_, err := FindConfigFiles("non-existent-dir")
		Expect(err).To(HaveOccurred())
	})
})

