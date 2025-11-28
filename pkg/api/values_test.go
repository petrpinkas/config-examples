package api

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Package Suite")
}

var _ = Describe("Configuration Values", func() {
	BeforeEach(func() {
		// Reset environment for each test
		os.Clearenv()
		// Reinitialize Viper to pick up cleared env
		Values = viper.New()
		Values.SetDefault(OidcRealm, "trusted-artifact-signer")
		Values.SetDefault(OidcUser, "jdoe")
		Values.SetDefault(OidcPassword, "secure")
		Values.SetDefault(OidcUserDomain, "redhat.com")
		Values.SetDefault(OidcClientID, "trusted-artifact-signer")
		Values.SetDefault(ManualImageSetup, "false")
		Values.AutomaticEnv()
	})

	It("should return default values when environment variables are not set", func() {
		Expect(GetValueFor(OidcRealm)).To(Equal("trusted-artifact-signer"))
		Expect(GetValueFor(OidcUser)).To(Equal("jdoe"))
		Expect(GetValueFor(OidcPassword)).To(Equal("secure"))
		Expect(GetValueFor(OidcUserDomain)).To(Equal("redhat.com"))
		Expect(GetValueFor(OidcClientID)).To(Equal("trusted-artifact-signer"))
		Expect(GetValueFor(ManualImageSetup)).To(Equal("false"))
	})

	It("should return environment variable values when set", func() {
		os.Setenv("OIDC_USER", "testuser")
		os.Setenv("SIGSTORE_FULCIO_URL", "https://fulcio.example.com")
		os.Setenv("TUF_URL", "https://tuf.example.com")

		Expect(GetValueFor(OidcUser)).To(Equal("testuser"))
		Expect(GetValueFor(FulcioURL)).To(Equal("https://fulcio.example.com"))
		Expect(GetValueFor(TufURL)).To(Equal("https://tuf.example.com"))
	})

	It("should return empty string for unset non-default values", func() {
		Expect(GetValueFor(FulcioURL)).To(BeEmpty())
		Expect(GetValueFor(RekorURL)).To(BeEmpty())
		Expect(GetValueFor(OidcIssuerURL)).To(BeEmpty())
	})
})

