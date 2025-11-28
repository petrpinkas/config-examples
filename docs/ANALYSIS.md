# Analysis of Similar Projects

This document summarizes patterns and conventions found in similar RHTAS/Securesign projects, particularly `sigstore-e2e` and `quickstarts`.

# Analysis of Similar Projects

This document summarizes key findings from analyzing similar RHTAS/Securesign projects (`sigstore-e2e` and `quickstarts`).

**Note**: For implementation patterns and code examples, see `.cursor/rules`.

## Key Findings

### Project Structure
- Uses `pkg/` for reusable packages
- Test suites in `test/` directory (not `tests/`)
- Shared utilities in `test/support/` or `test/testsupport/`

### Configuration Management
- Uses Viper for environment variable management
- Constants defined for all config keys
- Default values set in `init()` function
- `CheckMandatoryAPIConfigValues()` for validation

### Kubernetes Client
- Singleton pattern with `sync.Once`
- Combines `controller-runtime` client with `kubernetes.Interface`
- Registers all required schemes (core K8s, CRDs, OpenShift APIs)

### CLI Tool Integration
- Base `cli` struct with setup strategy
- Simple wrapper pattern (no complex abstractions)
- Logging integration with logrus

### Test Patterns
- Ginkgo v2 with `Ordered` tests
- `BeforeAll()` for setup, `DeferCleanup()` for cleanup
- Uses `Eventually()` with context for async operations

### Makefile Patterns
- Environment variable loading from script
- Common targets: `all`, `env`, `build`, `test`, `lint`

### YAML Configuration
- Uses kustomize with envsubst for placeholders
- Base configs with patches applied via kustomization

## Differences for Our Project

1. **No cleanup** - Keep installations after tests
2. **YAML config files** - Load from filesystem, not just environment
3. **Dot-notation updates** - Update nested YAML values before installation
4. **Subfolder structure** - Run tests per subfolder scenario
5. **Multiple configs** - Support multiple YAML files in a directory

