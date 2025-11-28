package config

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Template tests are included in TestConfig suite

var _ = Describe("Template Processing", func() {
	var tmpDir string
	var templatePath string
	var confPath string
	var outputPath string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "template-test-*")
		Expect(err).NotTo(HaveOccurred())

		templatePath = filepath.Join(tmpDir, "test-template.yaml")
		confPath = filepath.Join(tmpDir, "test-default.conf")
		outputPath = filepath.Join(tmpDir, "test-default.yaml")
	})

	AfterEach(func() {
		if tmpDir != "" {
			_ = os.RemoveAll(tmpDir)
		}
	})

	Describe("LoadConfFile", func() {
		It("should load conf file with key=value pairs", func() {
			confContent := `Issuer=https://keycloak.example.com/auth/realms/rhtas
IssuerURL=https://keycloak.example.com/auth/realms/rhtas
`
			err := os.WriteFile(confPath, []byte(confContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			conf, err := LoadConfFile(confPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(conf).To(HaveLen(2))
			Expect(conf["Issuer"]).To(Equal("https://keycloak.example.com/auth/realms/rhtas"))
			Expect(conf["IssuerURL"]).To(Equal("https://keycloak.example.com/auth/realms/rhtas"))
		})

		It("should skip empty lines and comments", func() {
			confContent := `# This is a comment
Issuer=https://keycloak.example.com

IssuerURL=https://keycloak.example.com
`
			err := os.WriteFile(confPath, []byte(confContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			conf, err := LoadConfFile(confPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(conf).To(HaveLen(2))
		})

		It("should return error for invalid format", func() {
			confContent := `invalid-line-without-equals`
			err := os.WriteFile(confPath, []byte(confContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			_, err = LoadConfFile(confPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid format"))
		})
	})

	Describe("ProcessTemplate", func() {
		It("should replace placeholder values with conf values", func() {
			// Create template file
			templateContent := `kind: Securesign
spec:
  fulcio:
    config:
      OIDCIssuers:
        - Issuer: 'https://your-oidc-issuer-url'
          IssuerURL: 'https://your-oidc-issuer-url'
`
			err := os.WriteFile(templatePath, []byte(templateContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Create conf file
			confContent := `Issuer=https://keycloak.example.com/auth/realms/rhtas
IssuerURL=https://keycloak.example.com/auth/realms/rhtas
`
			err = os.WriteFile(confPath, []byte(confContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Process template
			err = ProcessTemplate(templatePath, confPath, outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify output file exists
			_, err = os.Stat(outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Load and verify the output
			outputConfig, err := LoadConfig(outputPath)
			Expect(err).NotTo(HaveOccurred())

			spec, ok := outputConfig.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			fulcio, ok := spec["fulcio"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			config, ok := fulcio["config"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			oidcIssuers, ok := config["OIDCIssuers"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(oidcIssuers).To(HaveLen(1))

			issuerMap, ok := oidcIssuers[0].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(issuerMap["Issuer"]).To(Equal("https://keycloak.example.com/auth/realms/rhtas"))
			Expect(issuerMap["IssuerURL"]).To(Equal("https://keycloak.example.com/auth/realms/rhtas"))
		})
	})

	Describe("ProcessTemplateFromPaths", func() {
		It("should process template using scenario and variant names", func() {
			projectRoot := getProjectRoot()
			scenarioDir := filepath.Join(projectRoot, "scenarios", "basic")
			baseName := "rhtas-basic"
			variantName := "default"

			outputPath, err := ProcessTemplateFromPaths(scenarioDir, baseName, variantName)
			Expect(err).NotTo(HaveOccurred())
			Expect(outputPath).To(ContainSubstring("rhtas-basic-default.yaml"))

			// Verify the file was created
			_, err = os.Stat(outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Load and verify values were replaced
			outputConfig, err := LoadConfig(outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Check that placeholder was replaced
			spec, ok := outputConfig.Data["spec"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			fulcio, ok := spec["fulcio"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			config, ok := fulcio["config"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			oidcIssuers, ok := config["OIDCIssuers"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(oidcIssuers).To(HaveLen(1))

			issuerMap, ok := oidcIssuers[0].(map[string]interface{})
			Expect(ok).To(BeTrue())
			// Verify placeholder was replaced (not the placeholder value)
			Expect(issuerMap["Issuer"]).NotTo(Equal("'https://your-oidc-issuer-url'"))
			Expect(issuerMap["Issuer"]).NotTo(BeEmpty())
			Expect(issuerMap["IssuerURL"]).NotTo(Equal("'https://your-oidc-issuer-url'"))
			Expect(issuerMap["IssuerURL"]).NotTo(BeEmpty())
		})
	})
})

