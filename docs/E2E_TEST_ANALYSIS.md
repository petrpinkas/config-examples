# E2E Test Analysis - sigstore-e2e Project

This document summarizes key patterns from `sigstore-e2e` project, focusing on simple cosign sign/verify tests.

**Note**: For implementation patterns and code examples, see `.cursor/rules`.

## Key Findings

### 1. Simple Cosign Sign/Verify Test Structure

**Key Points**:
- Uses environment variables for URLs (set via `tas-env-variables.sh`)
- `cosign initialize` is called first (uses TUF_URL from env)
- Sign uses `--identity-token` flag
- Verify uses `--certificate-identity-regexp` and `--certificate-oidc-issuer-regexp`
- No explicit `--fulcio-url` or `--rekor-url` flags needed (uses env vars)

See `.cursor/rules` for code examples.

### 2. Required Environment Variables

**Mandatory**: FulcioURL, RekorURL, TufURL, OidcIssuerURL  
**Optional**: TsaURL, OidcToken  
**Defaults**: OidcUser="jdoe", OidcPassword="secure", OidcUserDomain="redhat.com", OidcRealm="trusted-artifact-signer", ManualImageSetup="false"

See `.cursor/rules` for implementation details.

### 3. Environment Variable Setup Script (`tas-env-variables.sh`)

The script auto-discovers URLs from the cluster:

```bash
# Auto-discover OIDC issuer
if [ -z "$OIDC_ISSUER_URL" ]; then
  export OIDC_ISSUER_URL=https://$(oc get route keycloak -n keycloak-system | tail -n 1 | awk '{print $2}')/auth/realms/trusted-artifact-signer
fi

# Auto-discover TUF URL
if [ -z "$TUF_URL" ]; then
  export TUF_URL=$(oc get tuf -o jsonpath='{.items[0].status.url}')
fi

# Auto-discover Fulcio URL
if [ -z "$FULCIO_URL" ]; then
  export COSIGN_FULCIO_URL=$(oc get fulcio -o jsonpath='{.items[0].status.url}')
fi

# Auto-discover Rekor URL
if [ -z "$REKOR_URL" ]; then
  export COSIGN_REKOR_URL=$(oc get rekor -o jsonpath='{.items[0].status.url}')
fi

# Auto-discover TSA URL
if [ -z "$TSA_URL" ]; then
  export TSA_URL=$(oc get timestampauthorities -o jsonpath='{.items[0].status.url}')/api/v1/timestamp
fi

# Set cosign environment variables
export COSIGN_MIRROR=$TUF_URL
export COSIGN_ROOT=$TUF_URL/root.json
export COSIGN_OIDC_CLIENT_ID=${OIDC_CLIENT_ID:-trusted-artifact-signer}
export COSIGN_OIDC_ISSUER=$OIDC_ISSUER_URL
export COSIGN_CERTIFICATE_OIDC_ISSUER=$OIDC_ISSUER_URL
export COSIGN_YES="true"

# Also export SIGSTORE_* variants
export SIGSTORE_FULCIO_URL=$COSIGN_FULCIO_URL
export SIGSTORE_OIDC_ISSUER=$COSIGN_OIDC_ISSUER
export SIGSTORE_REKOR_URL=$COSIGN_REKOR_URL
```

**Key Points**:
- Uses `oc get` commands to discover component URLs from CR status
- Sets both `COSIGN_*` and `SIGSTORE_*` environment variables
- Sets `COSIGN_MIRROR` and `COSIGN_ROOT` for TUF initialization
- Sets `COSIGN_YES="true"` for non-interactive mode

### 4. Cosign Client Implementation

Very simple wrapper around base CLI with `Command()`, `CommandOutput()`, and `Setup()` methods.

### 5. OIDC Token Retrieval

Checks `OIDC_TOKEN` env var first, falls back to HTTP request with username/password grant type.

### 6. Image Preparation

Default: Generate random image (`ttl.sh/<uuid>:5m`) and push. Alternative: Use `MANUAL_IMAGE_SETUP=true` with `TARGET_IMAGE_NAME`.

### 7. TSA Test Pattern (Optional)

Adds `--timestamp-server-url` to sign, downloads cert chain from `/certchain`, uses `--timestamp-certificate-chain` in verify.

See `.cursor/rules` for code examples.

## Simplified Test Flow for Our Project

1. Setup: Check mandatory config values, initialize cosign client, prepare test image
2. Initialize TUF: `cosign initialize` (uses env vars)
3. Sign Image: `cosign sign -y --identity-token=<token> <image>` (uses env vars)
4. Verify Image: `cosign verify --certificate-identity-regexp <pattern> --certificate-oidc-issuer-regexp <pattern> <image>`

Optional TSA: Add `--timestamp-server-url` to sign, download cert chain, add `--timestamp-certificate-chain` to verify.

See `.cursor/rules` for detailed implementation.

## Key Takeaways

- Environment variables are central (set by `tas-env-variables.sh` or auto-discovered)
- Simple cosign client wrapper (no complex abstractions)
- OIDC token can be provided via env or retrieved via HTTP
- Image handling: default to random images, support manual setup
- Minimal flags in cosign commands (uses env vars)
- TSA support is optional

## Differences for Our Project

1. **We install RHTAS**: They assume it's already installed
2. **We get URLs from CR status**: Can use Kubernetes client instead of `oc get`
3. **We verify installation first**: Need to verify components before testing
4. **We keep installations**: No cleanup needed
5. **YAML configs**: We load from files, they use programmatic creation

See `.cursor/rules` for implementation patterns.

