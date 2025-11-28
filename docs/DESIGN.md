# RHTAS Configuration Test Suite - Design Document

## Project Goals

This project provides a Ginkgo-based test suite for installing, verifying, and testing Red Hat Trusted Artifact Signer (RHTAS) configurations on OpenShift clusters.

### Core Functionality

1. **Install RHTAS Components**: Install Securesign custom resources and their components based on YAML configuration files
2. **Configuration Management**: Support updating configuration values using dot-notation paths (e.g., `spec.fulcio.config.OIDCIssuers.Issuer=value`)
3. **Verification**: Verify that RHTAS components are installed and running correctly
4. **Sign/Verify Operations**: Perform simple image signing and verification using cosign
5. **Multi-Configuration Support**: Run tests on multiple configuration files organized in subfolders

## Technologies

- **Go**: Primary programming language (Go 1.21+)
- **Ginkgo v2**: BDD testing framework for test suite structure
- **Gomega**: Matcher library for assertions
- **Kubernetes client-go**: Go client library for interacting with OpenShift/Kubernetes API
- **controller-runtime**: For working with custom resources (Securesign CRD)
- **Viper**: Configuration management (environment variables)
- **logrus**: Structured logging
- **cosign**: CLI tool for container image signing and verification
- **YAML**: Configuration file format (gopkg.in/yaml.v3)

## Project Structure

```
config-examples/
├── README.md
├── docs/                        # Design and analysis documents
│   ├── DESIGN.md
│   ├── ANALYSIS.md
│   ├── OPERATOR_TEST_ANALYSIS.md
│   └── E2E_TEST_ANALYSIS.md
├── go.mod
├── go.sum
├── scenarios/                   # Test scenarios organized by subfolder
│   ├── basic/
│   │   └── rhtas-basic.yaml     # Basic configuration example
│   └── advanced/
│       └── rhtas-advanced.yaml
├── pkg/
│   ├── api/                     # Configuration constants and environment variables
│   │   └── values.go
│   ├── config/                  # Configuration loading and manipulation
│   │   └── config.go
│   ├── clients/                 # CLI tool wrappers
│   │   ├── cli.go              # Base CLI abstraction
│   │   └── cosign.go          # Cosign client wrapper
│   ├── kubernetes/              # Kubernetes client and helpers
│   │   └── client.go
│   ├── installer/               # RHTAS installation logic
│   │   └── installer.go
│   └── verifier/                # Component verification logic
│       ├── verifier.go          # Main verification functions
│       ├── securesign.go        # Securesign CR verification
│       ├── fulcio.go            # Fulcio component verification
│       ├── rekor.go             # Rekor component verification
│       ├── tsa.go               # TSA component verification
│       ├── tuf.go               # TUF component verification
│       ├── trillian.go          # Trillian component verification
│       ├── ctlog.go             # CTlog component verification
│       └── condition.go         # Condition checking utilities
├── test/                        # Test suites (note: 'test' not 'tests')
│   ├── rhtas/                  # Main RHTAS test suite
│   │   ├── rhtas_suite_test.go
│   │   └── rhtas_test.go
│   └── support/                # Shared test utilities
│       └── test_support.go
├── Makefile                    # Build and test automation
└── tas-env-variables.sh        # Environment variable setup script
```

## Key Components

### 1. API/Configuration (`pkg/api`)

- **Environment Variables**: Constants for all configuration keys
- **Viper Integration**: Centralized configuration management with defaults
- **Required Parameters**: FulcioURL, RekorURL, TufURL, OidcIssuerURL (mandatory), TsaURL (optional)
- **OIDC Authentication**: OidcToken (optional), OidcUser, OidcPassword, OidcUserDomain, OidcRealm (with defaults)
- **Image Setup**: ManualImageSetup, TargetImageName

See `.cursor/rules` for implementation patterns.

### 2. Configuration Management (`pkg/config`)

- **LoadConfig**: Load YAML configuration files from filesystem
- **UpdateConfig**: Update configuration values using dot-notation paths
  - Example: `spec.fulcio.config.OIDCIssuers.Issuer=https://new-issuer.com`
  - Supports nested maps and arrays
- **FindConfigFiles**: Discover YAML files in directories/subfolders
- **ToYAML**: Convert config back to YAML for applying to cluster

### 3. Kubernetes Client (`pkg/kubernetes`)

- Singleton pattern with `sync.Once`
- Combines `controller-runtime` client with `kubernetes.Interface`
- Registers Securesign CRD scheme
- Helper methods for creating resources from YAML

### 4. CLI Tool Abstraction (`pkg/clients`)

- Base CLI wrapper with setup strategy
- Cosign client: Simple wrapper using environment variables
- Logging integration with logrus

### 5. Installer (`pkg/installer`)

- Creates/applies Securesign CR to OpenShift cluster
- Handles namespace creation if needed
- Supports idempotent operations

### 6. Verifier (`pkg/verifier`)

- Component-specific `Get()` and `Verify()` functions
- `VerifyAllComponents()` verifies all components in dependency order
- Checks CR status conditions and deployment/pod status

### 7. Test Support (`test/support`)

- TestContext, timeouts, config validation
- OIDC token retrieval, image preparation
- Kubernetes client creation, URL auto-discovery

### 8. Test Suite (`test/rhtas/rhtas_suite_test.go`)

- Ginkgo suite with Ordered tests
- Discovers config files, loads, updates, installs, verifies
- Keeps installations after tests (no cleanup)

See `.cursor/rules` for detailed implementation patterns.

## Usage

### Environment Setup

First, set up environment variables. The `tas-env-variables.sh` script can auto-discover URLs from the cluster:

```bash
# Source environment variables script (auto-discovers from cluster)
source tas-env-variables.sh

# Or set manually
export SIGSTORE_FULCIO_URL=https://fulcio-server-rhtas-simple.apps.example.com
export SIGSTORE_REKOR_URL=https://rekor-server-rhtas-simple.apps.example.com
export SIGSTORE_OIDC_ISSUER=https://keycloak-keycloak-system.apps.example.com/auth/realms/trusted-artifact-signer
export TUF_URL=https://tuf-rhtas-simple.apps.example.com
export TSA_URL=https://tsa-server-rhtas-simple.apps.example.com

# OIDC authentication (optional, defaults provided)
export OIDC_USER=jdoe
export OIDC_PASSWORD=secure
export OIDC_USER_DOMAIN=redhat.com
export KEYCLOAK_REALM=trusted-artifact-signer

# Image setup (optional)
export MANUAL_IMAGE_SETUP=false  # or true to use TARGET_IMAGE_NAME
export TARGET_IMAGE_NAME=ttl.sh/my-image:5m  # if MANUAL_IMAGE_SETUP=true
```

**Auto-Discovery**: The script uses `oc get` commands to discover URLs from CR status:
- `oc get fulcio -o jsonpath='{.items[0].status.url}'`
- `oc get rekor -o jsonpath='{.items[0].status.url}'`
- `oc get tuf -o jsonpath='{.items[0].status.url}'`
- `oc get timestampauthorities -o jsonpath='{.items[0].status.url}'`

**Alternative**: Our tests can use Kubernetes client to discover URLs programmatically after installation.

### Running Tests

```bash
# Using Makefile (recommended)
make test                    # Run all tests with environment loaded
make env test               # Load env and run tests

# Using Ginkgo directly
ginkgo test/rhtas/ -- --scenario=scenarios/basic

# Run tests for root level config
ginkgo test/rhtas/ -- --scenario=.

# Run with placeholder updates
ginkgo test/rhtas/ -- --scenario=scenarios/basic \
  --update=spec.fulcio.config.OIDCIssuers.Issuer=https://new-issuer.com \
  --update=metadata.name=my-securesign-instance

# Using go test
go test -v ./test/rhtas/... --ginkgo.v
```

### Configuration Placeholders

Update configuration values before installation using dot-notation paths:

```bash
--update=spec.fulcio.config.OIDCIssuers.Issuer=https://keycloak.example.com/auth/realms/rhtas
--update=spec.fulcio.certificate.commonName=fulcio.example.com
--update=metadata.namespace=my-namespace
--update=spec.tsa.signer.certificateChain.rootCA.commonName=tsa-root.example.com
```

### Makefile Targets

```bash
make all          # Build, load env, and run tests
make env          # Generate .env file from tas-env-variables.sh
make build        # Build Go code
make test         # Run tests (loads .env if present)
make lint         # Run linters (golangci-lint)
```

## Workflow

1. **Discovery**: Find all YAML config files in specified subfolder
2. **Load**: Load each configuration file
3. **Update**: Apply placeholder updates (if provided)
4. **Install**: Create Securesign CR in OpenShift cluster
5. **Verify**: Wait for and verify components are ready
6. **Test**: Optionally perform sign/verify operations
7. **Keep**: Leave installation in place (no cleanup)

## Implementation Patterns

See `.cursor/rules` for detailed implementation patterns and code examples. Key patterns include:
- Configuration management with Viper
- Kubernetes client singleton pattern
- CLI tool abstraction
- Component verification pattern
- Test structure with Ginkgo v2

## Assumptions

- OpenShift cluster is accessible via `oc` or kubeconfig (auto-detected)
- RHTAS operator is already installed in the cluster
- User has appropriate permissions to create Securesign CRs
- cosign CLI can be downloaded/setup automatically or is in PATH
- Tests run against a test/development cluster (installations are kept intentionally)
- Environment variables can be set via script or manually

## Code Examples

See `.cursor/rules` for detailed code examples and patterns.

## Future Enhancements

- Support for multiple placeholder files
- Parallel test execution for independent configurations
- Health check endpoints verification
- TUF root key management
- Integration with CI/CD pipelines
- Support for downloading CLI tools from cluster console
- Enhanced logging and reporting
- Test result artifacts collection
- Component-specific verification helpers (similar to operator tests)
- Support for verifying component secrets and configmaps
- Enhanced error messages with component status details

