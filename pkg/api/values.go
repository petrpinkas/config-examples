package api

import "github.com/spf13/viper"

// Environment variable constants for RHTAS configuration
const (
	// Required URLs
	FulcioURL     = "SIGSTORE_FULCIO_URL"
	RekorURL      = "SIGSTORE_REKOR_URL"
	TufURL        = "TUF_URL"
	OidcIssuerURL = "SIGSTORE_OIDC_ISSUER"
	TsaURL        = "TSA_URL" // Optional, for TSA tests

	// OIDC Authentication
	OidcToken      = "OIDC_TOKEN"        // Optional, retrieved if not set
	OidcUser       = "OIDC_USER"         // Default: "jdoe"
	OidcPassword   = "OIDC_PASSWORD"     // Default: "secure"
	OidcUserDomain = "OIDC_USER_DOMAIN"  // Default: "redhat.com"
	OidcRealm      = "KEYCLOAK_REALM"     // Default: "trusted-artifact-signer"
	OidcClientID   = "OIDC_CLIENT_ID"    // Default: "trusted-artifact-signer"

	// Image Setup
	ManualImageSetup = "MANUAL_IMAGE_SETUP" // Default: "false"
	TargetImageName  = "TARGET_IMAGE_NAME"   // Required if ManualImageSetup=true
)

// Values holds the Viper instance for configuration management
var Values *viper.Viper

func init() {
	Values = viper.New()

	// Set default values
	Values.SetDefault(OidcRealm, "trusted-artifact-signer")
	Values.SetDefault(OidcUser, "jdoe")
	Values.SetDefault(OidcPassword, "secure")
	Values.SetDefault(OidcUserDomain, "redhat.com")
	Values.SetDefault(OidcClientID, "trusted-artifact-signer")
	Values.SetDefault(ManualImageSetup, "false")

	// Automatically read from environment variables
	Values.AutomaticEnv()
}

// GetValueFor retrieves a configuration value by key
func GetValueFor(key string) string {
	return Values.GetString(key)
}

