# RHTAS Configuration Test Suite

A Ginkgo-based test suite for installing, verifying, and testing Red Hat Trusted Artifact Signer (RHTAS) configurations on OpenShift clusters.

## Running Tests

### Using Ginkgo CLI

Run all tests:
```bash
go run github.com/onsi/ginkgo/v2/ginkgo -v test/...
```

Run specific scenario (using focus):
```bash
go run github.com/onsi/ginkgo/v2/ginkgo --focus "Basic Scenario" -v test/...
```

Run specific test group:
```bash
go run github.com/onsi/ginkgo/v2/ginkgo --focus "Config Loading" -v test/...
```

Run from a specific test directory:
```bash
cd test/rhtas
go run github.com/onsi/ginkgo/v2/ginkgo -v
```

**Note**: If you have `ginkgo` installed in your PATH, you can use it directly:
```bash
ginkgo -v test/...
```

### Using Go Test

Run all tests:
```bash
go test ./test/rhtas/... -v
```

Run specific test:
```bash
go test ./test/rhtas/... -v -ginkgo.focus "Basic Scenario"
```

Run with labels (if you add labels to tests):
```bash
go test ./test/rhtas/... -v -ginkgo.label-filter "scenario=basic"
```

### Common Ginkgo Flags

- `-v` or `--verbose`: Verbose output
- `--focus <regex>`: Run tests matching the regex
- `--skip <regex>`: Skip tests matching the regex
- `--label-filter <expression>`: Filter by labels (e.g., `scenario=basic`)
- `--until-it-fails`: Keep running until a test fails
- `--repeat <n>`: Run tests n times
- `--randomize-all`: Randomize test execution order
- `--seed <n>`: Set random seed for reproducible runs

## Project Structure

- `pkg/` - Reusable packages (api, config, clients, kubernetes, installer, verifier)
- `test/rhtas/` - Main RHTAS test suite
- `scenarios/` - Test scenarios organized by subfolder (e.g., `scenarios/basic/`)

## Scenarios

Tests are organized by scenario folders. Each scenario folder contains YAML configuration files that define the RHTAS setup to test.

- `scenarios/basic/` - Basic RHTAS configuration
