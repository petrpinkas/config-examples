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
		scenarioDir := filepath.Join(projectRoot, "scenarios", "basic")
		
		// Generate config file from template first
		runtimeCtx := &RuntimeContext{
			Namespace:    "test-namespace",
			InstanceName: "securesign-sample",
		}
		configPath, err := ProcessTemplateFromPaths(scenarioDir, "rhtas-basic", "default", runtimeCtx)
		Expect(err).NotTo(HaveOccurred())
		
		config, err := LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())
		Expect(config.Data).NotTo(BeNil())
		Expect(config.GetKind()).To(Equal("Securesign"))
		Expect(config.GetAPIVersion()).To(Equal("rhtas.redhat.com/v1alpha1"))
		Expect(config.GetName()).NotTo(BeEmpty())
		Expect(config.GetNamespace()).NotTo(BeEmpty())
		// Verify spec exists
		if spec, ok := config.Data["spec"].(map[string]interface{}); ok {
			Expect(spec).NotTo(BeNil())
		} else {
			Fail("spec should be a map")
		}
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
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		_, err = tmpFile.WriteString("invalid: yaml: content: [")
		Expect(err).NotTo(HaveOccurred())
		_ = tmpFile.Close()

		_, err = LoadConfig(tmpFile.Name())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to parse YAML"))
	})
})

var _ = Describe("Config Updates", func() {
	var config *Config

	BeforeEach(func() {
		projectRoot := getProjectRoot()
		scenarioDir := filepath.Join(projectRoot, "scenarios", "basic")
		
		// Generate config file from template first
		runtimeCtx := &RuntimeContext{
			Namespace:    "test-namespace",
			InstanceName: "securesign-sample",
		}
		configPath, err := ProcessTemplateFromPaths(scenarioDir, "rhtas-basic", "default", runtimeCtx)
		Expect(err).NotTo(HaveOccurred())
		
		config, err = LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should update metadata.name", func() {
		err := UpdateConfig(config, "metadata.name=my-securesign")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.GetName()).To(Equal("my-securesign"))
	})

	It("should update metadata.namespace", func() {
		err := UpdateConfig(config, "metadata.namespace=my-namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.GetNamespace()).To(Equal("my-namespace"))
	})

	It("should update metadata.labels", func() {
		err := UpdateConfig(config, "metadata.labels.test=value")
		Expect(err).NotTo(HaveOccurred())

		metadata, ok := config.Data["metadata"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		labels, ok := metadata["labels"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(labels["test"]).To(Equal("value"))
	})

	It("should update nested spec values", func() {
		err := UpdateConfig(config, "spec.fulcio.certificate.commonName=fulcio.example.com")
		Expect(err).NotTo(HaveOccurred())

		spec, ok := config.Data["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		fulcio, ok := spec["fulcio"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		certificate, ok := fulcio["certificate"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(certificate["commonName"]).To(Equal("fulcio.example.com"))
	})

	It("should create nested maps when they don't exist", func() {
		err := UpdateConfig(config, "spec.newcomponent.setting=value")
		Expect(err).NotTo(HaveOccurred())

		spec, ok := config.Data["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		newComponent, ok := spec["newcomponent"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(newComponent["setting"]).To(Equal("value"))
	})

	It("should return error for invalid path format", func() {
		err := UpdateConfig(config, "invalid-path")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid path=value format"))
	})

	It("should update any top-level field", func() {
		err := UpdateConfig(config, "kind=MyCustomKind")
		Expect(err).NotTo(HaveOccurred())
		Expect(config.GetKind()).To(Equal("MyCustomKind"))
	})
})

var _ = Describe("Config to YAML", func() {
	It("should convert config back to YAML", func() {
		projectRoot := getProjectRoot()
		scenarioDir := filepath.Join(projectRoot, "scenarios", "basic")
		
		// Generate config file from template first
		runtimeCtx := &RuntimeContext{
			Namespace:    "test-namespace",
			InstanceName: "securesign-sample",
		}
		configPath, err := ProcessTemplateFromPaths(scenarioDir, "rhtas-basic", "default", runtimeCtx)
		Expect(err).NotTo(HaveOccurred())
		
		config, err := LoadConfig(configPath)
		Expect(err).NotTo(HaveOccurred())

		yamlData, err := config.ToYAML()
		Expect(err).NotTo(HaveOccurred())
		Expect(yamlData).NotTo(BeEmpty())
		Expect(string(yamlData)).To(ContainSubstring("apiVersion"))
		Expect(string(yamlData)).To(ContainSubstring("kind: Securesign"))
	})

	It("should handle generic Kubernetes resources", func() {
		// Create a minimal generic config
		config := &Config{
			Data: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test-config",
					"namespace": "default",
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		}

		yamlData, err := config.ToYAML()
		Expect(err).NotTo(HaveOccurred())
		Expect(yamlData).NotTo(BeEmpty())
		Expect(string(yamlData)).To(ContainSubstring("kind: ConfigMap"))
		Expect(string(yamlData)).To(ContainSubstring("name: test-config"))
	})
})

var _ = Describe("Find Config Files", func() {
	It("should find YAML files in directory", func() {
		projectRoot := getProjectRoot()
		dir := filepath.Join(projectRoot, "scenarios", "basic")
		
		// Generate config file from template first
		runtimeCtx := &RuntimeContext{
			Namespace:    "test-namespace",
			InstanceName: "securesign-sample",
		}
		_, err := ProcessTemplateFromPaths(dir, "rhtas-basic", "default", runtimeCtx)
		Expect(err).NotTo(HaveOccurred())
		
		files, err := FindConfigFiles(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).NotTo(BeEmpty())
		Expect(files).To(ContainElement(ContainSubstring("rhtas-basic-default.yaml")))
	})

	It("should find both .yaml and .yml files", func() {
		// Create a temporary directory with both extensions
		tmpDir, err := os.MkdirTemp("", "test-configs-*")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = os.RemoveAll(tmpDir) }()

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
		defer func() { _ = os.RemoveAll(tmpDir) }()

		files, err := FindConfigFiles(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeEmpty())
	})

	It("should return error for non-existent directory", func() {
		_, err := FindConfigFiles("non-existent-dir")
		Expect(err).To(HaveOccurred())
	})
})
