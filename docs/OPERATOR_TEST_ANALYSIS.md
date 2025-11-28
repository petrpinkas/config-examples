# Operator Test Patterns Analysis

This document summarizes key test patterns from `secure-sign-operator` project.

**Note**: For implementation patterns and code examples, see `.cursor/rules`.

## Test Structure

### Unit Tests (`api/v1alpha1/`, `internal/controller/`)

- **Location**: Co-located with source code or in `internal/` packages
- **Framework**: Ginkgo v2 with Gomega matchers
- **Test Environment**: Uses `envtest` (fake K8s API server) for unit tests
- **Build Tags**: No special tags (run by default)

### Integration/E2E Tests (`test/e2e/`)

- **Location**: Separate `test/e2e/` directory
- **Framework**: Ginkgo v2 with Gomega matchers
- **Build Tags**: Uses `//go:build integration` tag
- **Real Cluster**: Tests run against real OpenShift/Kubernetes cluster

## Key Patterns

### 1. Test Suite Setup

Unit tests use `envtest` (fake K8s API server). E2E tests use real cluster with `SetDefaultEventuallyTimeout(3 * time.Minute)` and `EnforceDefaultTimeoutsWhenUsingContexts()`.

See `.cursor/rules` for detailed setup patterns.

### 2. Client Creation Pattern

Register all required schemes (core K8s, CRDs, OpenShift APIs), use `config.GetConfig()` for automatic kubeconfig detection.

### 3. Component Verification Pattern

Each component has `Get()` and `Verify()` functions. `Get()` retrieves CR instance (returns `nil` if not found), `Verify()` uses `Eventually()` to wait for readiness, checks both CR status conditions AND deployment status.

### 4. Condition Checking

Check CR status conditions (Ready condition) and deployment/pod status separately using label selectors.

### 5. Test Resource Creation Pattern

Builder pattern with Options functions for composable configuration (not used in our project - we load from YAML files).

### 6. All Components Verification

Single function to verify all components in dependency order: Trillian → Fulcio → TSA → Rekor → CTlog → TUF → Securesign.

### 7. Cosign Sign/Verify Testing

Get component URLs from status, wait for certificate chain, use OIDC token, initialize TUF, sign with all URLs, verify with identity/issuer regex patterns.

### 8. Test Structure

Uses `Ordered` tests, multiple `BeforeAll` blocks, `SpecContext` for context handling, `Eventually()` for async operations.

See `.cursor/rules` for detailed code examples.

## Key Takeaways

- Component verification helpers with `Get()` and `Verify()` functions
- All components verification in dependency order
- Cosign integration pattern (get URLs, initialize TUF, sign/verify)
- Test structure with `Ordered` tests and `SpecContext`

## Differences for Our Project

1. **No Cleanup**: We keep installations (no `AfterEach` cleanup)
2. **YAML File Loading**: Load configs from filesystem, not programmatic creation
3. **Dot-notation Updates**: Update YAML values before installation
4. **Subfolder Structure**: Run tests per subfolder scenario
5. **Multiple Configs**: Support multiple YAML files in a directory

See `.cursor/rules` for implementation patterns and guidance.

