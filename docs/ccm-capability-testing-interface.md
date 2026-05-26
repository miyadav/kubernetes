# CCM Capability-Based Testing Interface

## The Problem

Today, Kubernetes e2e tests decide which cloud-specific tests to run using hardcoded provider names:

```go
e2eskipper.SkipUnlessProviderIs("aws", "gce")
```

This creates three problems:

1. **New providers must patch k/k.** If you build `cloud-provider-acme`, you must submit a PR to `kubernetes/kubernetes` adding `"acme"` to every relevant `SkipUnlessProviderIs` call before your tests can run.

2. **No programmatic discovery.** There is no way for a test to ask "does this cloud support load balancers?" — it can only ask "is this AWS or GCE?"

3. **No enforcement.** A provider can claim support for anything without ever proving the methods actually work.

## The Solution: Capability Discovery

Replace identity checks ("are you AWS?") with capability checks ("do you support load balancers?"):

```go
// BEFORE — identity-based
e2eskipper.SkipUnlessProviderIs("aws", "gce")

// AFTER — capability-based
e2eskipper.SkipUnlessCloudCapability(cloudprovidertesting.CapNodeDeletion)
```

A cloud provider declares what it supports by filling a map of capabilities. Tests query the map. No provider names anywhere in test logic.

---

## How It Works — The Full Picture

There are five pieces. Each is simple on its own.

```
 +--------------------------+
 |  1. Capability Constants |   What can be declared (LoadBalancer, Zones, ...)
 +--------------------------+
              |
              v
 +--------------------------+
 |  2. TestCapabilities     |   The interface: Has(cap) bool, ProviderName() string
 |     + MapCapabilities    |   The concrete implementation: a name + a map
 +--------------------------+
              |
              v
 +--------------------------+
 |  3. DeriveFromCloud()    |   Auto-fills the map by calling cloud.LoadBalancer(), etc.
 +--------------------------+
              |
              v
 +--------------------------+
 |  4. Global Registry      |   RegisterCapabilities("aws", caps) / GetCapabilities("aws")
 +--------------------------+
              |
              v
 +--------------------------+
 |  5. E2E Skip Helpers     |   SkipUnlessCloudCapability(cap) — reads the registry
 +--------------------------+
```

---

## Piece 1: Capability Constants

**File:** `staging/src/k8s.io/cloud-provider/testing/capabilities.go`

A `Capability` is just a string. There are two kinds:

**Core capabilities** — auto-derivable from `cloudprovider.Interface`:

| Constant | What it means |
|---|---|
| `CapLoadBalancer` | Provider implements `LoadBalancer()` |
| `CapInstances` | Provider implements `Instances()` |
| `CapInstancesV2` | Provider implements `InstancesV2()` |
| `CapZones` | Provider implements `Zones()` |
| `CapRoutes` | Provider implements `Routes()` |
| `CapClusters` | Provider implements `Clusters()` |

**Sub-capabilities** — provider must opt-in explicitly:

| Constant | What it means |
|---|---|
| `CapNodeDeletion` | Provider can delete nodes from the cloud |
| `CapSSHAccess` | Nodes are SSH-accessible |
| `CapInternalLoadBalancer` | Supports internal (non-internet-facing) LBs |
| `CapVolumeProvisioning` | Can dynamically provision volumes |
| `CapNodeResize` | Supports changing node instance types |
| `CapTopologyLabels` | Sets topology-related node labels |

**Extensibility:** Since `Capability` is a string, external providers can define their own without touching upstream:

```go
myCustomCap := cloudprovidertesting.Capability("acme/gpu-scheduling")
```

---

## Piece 2: The Interface and Its Implementation

**File:** `staging/src/k8s.io/cloud-provider/testing/capabilities.go` (interface)
**File:** `staging/src/k8s.io/cloud-provider/testing/default_capabilities.go` (implementation)

The interface has exactly two methods:

```go
type TestCapabilities interface {
    Has(cap Capability) bool       // "Do you support this?"
    ProviderName() string          // "Who are you?"
}
```

The concrete implementation is a struct with a name and a map:

```go
type MapCapabilities struct {
    Name string                    // e.g. "aws"
    Caps map[Capability]bool       // e.g. {LoadBalancer: true, Clusters: false}
}
```

`Has()` returns `true` if the key exists and its value is `true`. A missing key returns `false`.

---

## Piece 3: DeriveFromCloud — Automatic Capability Detection

**File:** `staging/src/k8s.io/cloud-provider/testing/default_capabilities.go`

Every `cloudprovider.Interface` method already returns `(implementation, bool)`:

```go
// This is the existing pattern in cloud.go:
LoadBalancer() (cloudprovider.LoadBalancer, bool)
Instances()    (cloudprovider.Instances, bool)
// ... etc.
```

`DeriveFromCloud` calls each of these and reads the bool:

```go
func DeriveFromCloud(cloud cloudprovider.Interface) *MapCapabilities {
    caps := &MapCapabilities{
        Name: cloud.ProviderName(),
        Caps: make(map[Capability]bool),
    }
    if _, ok := cloud.LoadBalancer(); ok {
        caps.Caps[CapLoadBalancer] = true
    }
    if _, ok := cloud.Instances(); ok {
        caps.Caps[CapInstances] = true
    }
    // ... same for InstancesV2, Zones, Routes, Clusters
    return caps
}
```

This means **every existing cloud provider gets correct core capabilities for free** — no code changes needed. Providers only add code to declare sub-capabilities.

**Typical usage:**

```go
caps := cloudprovidertesting.DeriveFromCloud(myCloud)   // auto-fills core caps
caps.Caps[cloudprovidertesting.CapNodeDeletion] = true   // opt-in to sub-caps
caps.Caps[cloudprovidertesting.CapSSHAccess] = true
```

---

## Piece 4: Global Registry

**File:** `staging/src/k8s.io/cloud-provider/testing/registry.go`

A process-global map, protected by a mutex. Same pattern as `framework.RegisterProvider`:

```go
var capabilities = make(map[string]TestCapabilities)   // guarded by sync.RWMutex

func RegisterCapabilities(providerName string, caps TestCapabilities)   // write
func GetCapabilities(providerName string) TestCapabilities              // read (nil if not found)
```

Registration happens in `init()` functions — the same `init()` that already registers the provider:

```go
// test/e2e/framework/providers/aws/aws.go
func init() {
    framework.RegisterProvider("aws", newProvider)
    cloudprovidertesting.RegisterCapabilities("aws", &cloudprovidertesting.MapCapabilities{
        Name: "aws",
        Caps: map[cloudprovidertesting.Capability]bool{
            cloudprovidertesting.CapLoadBalancer:         true,
            cloudprovidertesting.CapInstances:            true,
            cloudprovidertesting.CapInstancesV2:          true,
            cloudprovidertesting.CapZones:                true,
            cloudprovidertesting.CapRoutes:               true,
            cloudprovidertesting.CapClusters:             false,
            cloudprovidertesting.CapNodeDeletion:         true,
            cloudprovidertesting.CapSSHAccess:            true,
            cloudprovidertesting.CapInternalLoadBalancer: true,
            cloudprovidertesting.CapVolumeProvisioning:   true,
            cloudprovidertesting.CapNodeResize:           false,
            cloudprovidertesting.CapTopologyLabels:       true,
        },
    })
}
```

### How init() fires

The chain is already wired:

```
test/e2e/providers.go          (blank-imports provider packages)
    └── import _ ".../providers/aws"
            └── aws/aws.go init()
                    ├── framework.RegisterProvider("aws", ...)    // existing
                    └── cloudprovidertesting.RegisterCapabilities("aws", ...)  // NEW
```

When the e2e test binary starts, Go's init mechanism fires these automatically. By the time any test function runs, the registry is populated.

---

## Piece 5: E2E Skip Helpers

**File:** `test/e2e/framework/skipper/cloud_capabilities.go`

Two new functions:

```go
func SkipUnlessCloudCapability(cap Capability) {
    caps := cloudprovidertesting.GetCapabilities(framework.TestContext.Provider)
    if caps == nil {
        return   // <-- KEY: no-op when nothing is registered = backward compatible
    }
    if !caps.Has(cap) {
        skip("provider %q does not support %q", ...)
    }
}

func SkipIfCloudCapability(cap Capability) {
    // Inverse: skip if the provider DOES support the capability
}
```

### Backward Compatibility — The Critical Detail

`SkipUnlessCloudCapability` returns immediately (does nothing) when no capabilities are registered for the current provider. This means:

- **Providers that haven't adopted yet:** Tests fall through to the existing `SkipUnlessProviderIs` check. Nothing changes.
- **Providers that have adopted:** Both checks run. Since both must pass, behavior is identical.
- **Future (Phase 2):** Once all providers register, `SkipUnlessProviderIs` calls can be removed.

**Phase 1 (current) — dual-guard pattern:**

```go
ginkgo.BeforeEach(func() {
    e2eskipper.SkipUnlessProviderIs("aws", "gce")                          // old guard
    e2eskipper.SkipUnlessCloudCapability(cloudprovidertesting.CapNodeDeletion)  // new guard
})
```

---

## Enforcement: How Do We Know Providers Are Honest?

Declaring capabilities is cheap — a provider could claim `CapLoadBalancer: true` without implementing it. Three layers prevent this:

### Layer 1: Declaration (What the provider says)

The map itself. This is the provider's claim. It is necessary but not sufficient.

### Layer 2: Contract Validation (Do claims match the interface?)

**File:** `staging/src/k8s.io/cloud-provider/testing/contract.go`

`ValidateCapabilities` cross-checks declared capabilities against what `cloudprovider.Interface` actually reports:

```go
func ValidateCapabilities(t ContractT, cloud cloudprovider.Interface, declared TestCapabilities)
```

It catches two classes of errors:

- **Over-declaration:** Capabilities say "LoadBalancer: true" but `cloud.LoadBalancer()` returns `(nil, false)`.
- **Under-declaration:** `cloud.LoadBalancer()` returns `(impl, true)` but capabilities say "LoadBalancer: false".

It also checks that `cloud.ProviderName()` matches `declared.ProviderName()`.

**How it works internally:**

```go
var coreCapMapping = []struct {
    cap        Capability
    methodName string
    check      func(cloudprovider.Interface) bool
}{
    {CapLoadBalancer, "LoadBalancer()", func(c cloudprovider.Interface) bool {
        _, ok := c.LoadBalancer(); return ok
    }},
    // ... same for Instances, InstancesV2, Zones, Routes, Clusters
}
```

For each core capability, it calls the real method and compares the bool result with what was declared.

**Provider usage — runs without a cluster:**

```go
func TestCapabilitiesContract(t *testing.T) {
    myCloud := newMyCloud(cfg)
    caps := cloudprovidertesting.DeriveFromCloud(myCloud)
    caps.Caps[cloudprovidertesting.CapNodeDeletion] = true
    cloudprovidertesting.ValidateCapabilities(t, myCloud, caps)
}
```

### Layer 3: Conformance Suite (Do the methods actually work?)

**File:** `staging/src/k8s.io/cloud-provider/testing/conformance.go`

`RunConformanceSuite` goes beyond checking booleans — it calls every method on every declared sub-interface and verifies:

- Methods don't return `cloudprovider.NotImplemented`
- Methods return reasonable responses (non-empty names, non-nil metadata, etc.)

```go
func RunConformanceSuite(t *testing.T, cfg ConformanceConfig)
```

The suite auto-derives capabilities and only runs tests for supported ones:

```
RunConformanceSuite
  ├── ProviderName        (always)
  ├── LoadBalancer/        (if CapLoadBalancer)
  │   ├── GetLoadBalancerName
  │   └── GetLoadBalancer
  ├── Instances/           (if CapInstances)
  │   ├── CurrentNodeName
  │   ├── NodeAddresses
  │   ├── InstanceID
  │   ├── InstanceType
  │   └── InstanceExistsByProviderID
  ├── InstancesV2/         (if CapInstancesV2)
  │   ├── InstanceExists
  │   ├── InstanceShutdown
  │   └── InstanceMetadata
  ├── Zones/               (if CapZones)
  │   ├── GetZone
  │   ├── GetZoneByNodeName
  │   └── GetZoneByProviderID
  ├── Routes/              (if CapRoutes)
  │   └── ListRoutes
  └── Clusters/            (if CapClusters)
      └── ListClusters
```

**Provider usage — also runs without a cluster:**

```go
func TestConformance(t *testing.T) {
    myCloud := newMyCloud(cfg)
    cloudprovidertesting.RunConformanceSuite(t, cloudprovidertesting.ConformanceConfig{
        Cloud:       myCloud,
        ClusterName: "test-cluster",
        NodeName:    "test-node-1",
        ProviderID:  "acme://i-1234",
        Node:        &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node-1"}},
    })
}
```

A provider with everything disabled passes too — it just runs the `ProviderName` test:

```go
// Provider that supports nothing extra
cloud := &fake.Cloud{
    Provider:             "limited",
    DisableLoadBalancers: true,
    DisableRoutes:        true,
    DisableClusters:      true,
    DisableInstances:     true,
    DisableZones:         true,
}
RunConformanceSuite(t, ConformanceConfig{Cloud: cloud})
// Result: only ProviderName test runs. PASS.
```

---

## The Quality Gate for New Providers

The three layers form a progression:

```
  Provider registers capabilities               (compile time)
       |
       v
  ValidateCapabilities catches mismatches        (go test, no cluster)
       |
       v
  RunConformanceSuite verifies methods work       (go test, no cluster)
       |
       v
  E2E tests use SkipUnlessCloudCapability         (runtime, real cluster)
```

If a provider says `CapLoadBalancer: true` but `LoadBalancer()` returns `false`, the contract test fails — in CI, before any cluster is involved.

If a provider's `LoadBalancer()` returns `true` but `GetLoadBalancer()` returns `NotImplemented`, the conformance suite fails.

Only after both layers pass does the provider reach real e2e tests.

---

## How External Cloud Providers Adopt

External providers (`cloud-provider-aws`, `cloud-provider-azure`, etc.) already depend on `k8s.io/cloud-provider`. The testing package lives inside that module, so adoption requires zero new dependencies.

### Step 1: Register capabilities

Create a file (e.g., `tests/e2e/capabilities.go`):

```go
package e2e

import cloudprovidertesting "k8s.io/cloud-provider/testing"

func init() {
    cloudprovidertesting.RegisterCapabilities("aws", &cloudprovidertesting.MapCapabilities{
        Name: "aws",
        Caps: map[cloudprovidertesting.Capability]bool{
            cloudprovidertesting.CapLoadBalancer:         true,
            cloudprovidertesting.CapInstances:            true,
            cloudprovidertesting.CapInstancesV2:          true,
            cloudprovidertesting.CapZones:                true,
            cloudprovidertesting.CapRoutes:               true,
            cloudprovidertesting.CapClusters:             false,
            cloudprovidertesting.CapNodeDeletion:         true,
            cloudprovidertesting.CapSSHAccess:            true,
            cloudprovidertesting.CapInternalLoadBalancer: true,
            cloudprovidertesting.CapVolumeProvisioning:   true,
            cloudprovidertesting.CapNodeResize:           false,
            cloudprovidertesting.CapTopologyLabels:       true,
        },
    })
}
```

### Step 2: Use capability checks in tests

```go
// BEFORE
var _ = Describe("[cloud-provider-aws-e2e] load balancers", func() {
    // test assumes LB support exists

// AFTER
var _ = Describe("[cloud-provider-aws-e2e] load balancers", func() {
    f := framework.NewDefaultFramework("cloud-provider-aws")
    BeforeEach(func() {
        e2eskipper.SkipUnlessCloudCapability(cloudprovidertesting.CapLoadBalancer)
    })
```

### Step 3: Add contract and conformance tests

```go
func TestCapabilitiesContract(t *testing.T) {
    awsCloud := newAWSCloud(cfg)
    caps := cloudprovidertesting.DeriveFromCloud(awsCloud)
    caps.Caps[cloudprovidertesting.CapNodeDeletion] = true
    // ... add sub-capabilities
    cloudprovidertesting.ValidateCapabilities(t, awsCloud, caps)
}

func TestConformance(t *testing.T) {
    awsCloud := newAWSCloud(cfg)
    cloudprovidertesting.RunConformanceSuite(t, cloudprovidertesting.ConformanceConfig{
        Cloud:       awsCloud,
        ClusterName: "test-cluster",
        NodeName:    "test-node",
        ProviderID:  "aws://i-1234",
        Node:        &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}},
    })
}
```

### Step 4: Run it

```bash
# Contract + conformance — no cluster needed
go test ./tests/e2e/... -run TestCapabilitiesContract
go test ./tests/e2e/... -run TestConformance

# Full e2e — needs a running cluster
make test-e2e
```

---

## File Inventory

### New files in `k8s.io/cloud-provider/testing/`

| File | Purpose |
|---|---|
| `capabilities.go` | `Capability` type, 12 constants, `TestCapabilities` interface |
| `default_capabilities.go` | `MapCapabilities` struct, `DeriveFromCloud()` |
| `registry.go` | Global `RegisterCapabilities` / `GetCapabilities` |
| `contract.go` | `ValidateCapabilities` — declaration vs. reality check |
| `conformance.go` | `RunConformanceSuite` — method-level verification |
| `capabilities_test.go` | 9 unit tests for capabilities, derive, and registry |
| `contract_test.go` | 7 tests for contract validation and conformance suite |

### New file in `test/e2e/framework/skipper/`

| File | Purpose |
|---|---|
| `cloud_capabilities.go` | `SkipUnlessCloudCapability`, `SkipIfCloudCapability` |

### Modified files in `test/e2e/framework/providers/`

| File | Change |
|---|---|
| `providers/aws/aws.go` | Added `RegisterCapabilities("aws", ...)` in `init()` |
| `providers/gce/gce.go` | Added `RegisterCapabilities("gce", ...)` in `init()` |

### Modified file in `test/e2e/cloud/`

| File | Change |
|---|---|
| `cloud/nodes.go` | Added `SkipUnlessCloudCapability(CapNodeDeletion)` alongside existing `SkipUnlessProviderIs` |

---

## Design Patterns Used

### 1. Capability Discovery Map

Same pattern used by the Kubernetes storage test framework (`test/e2e/storage/framework/testdriver.go:168-260`). A `type Capability string` with a `map[Capability]bool`. Simple, extensible, well-understood in the Kubernetes codebase.

### 2. Interface Introspection

`DeriveFromCloud` leverages the existing `(impl, bool)` return pattern of `cloudprovider.Interface`. Every method like `LoadBalancer()` already returns a bool indicating support. We just read those bools into a map.

### 3. Global Registry with init()

Same pattern as `framework.RegisterProvider` and `plugins.RegisterCloudProvider`. A `map[string]T` protected by `sync.RWMutex`, populated by `init()` functions that fire when the test binary starts. Well-established Go idiom.

### 4. Graceful Degradation

`SkipUnlessCloudCapability` is a no-op when no capabilities are registered. This means adoption is incremental — providers can adopt at their own pace, and existing tests never break.

---

## What NOT to Migrate

Some tests are genuinely tied to specific cloud infrastructure:

- GCE cluster upgrade tests (use `gcloud` CLI directly)
- GKE node pool management tests
- AWS-specific API tests (NLB annotations, EBS CSI)

These should keep `SkipUnlessProviderIs`. Only tests gated by **generalizable capabilities** (node deletion, load balancing, topology labels) should transition.

---

## Test Results

All 16 tests pass:

```
$ go test -v ./staging/src/k8s.io/cloud-provider/testing/...

=== RUN   TestDeriveFromCloud_AllEnabled           PASS
=== RUN   TestDeriveFromCloud_DisabledFeatures     PASS
=== RUN   TestDeriveFromCloud_InstancesV2Enabled   PASS
=== RUN   TestDeriveFromCloud_AWSLike              PASS
=== RUN   TestDeriveFromCloud_DefaultProviderName  PASS
=== RUN   TestMapCapabilities_SubCapabilities      PASS
=== RUN   TestMapCapabilities_CustomCapability     PASS
=== RUN   TestRegistry                             PASS
=== RUN   TestRegistry_ReplaceExisting             PASS
=== RUN   TestValidateCapabilities_Consistent      PASS
=== RUN   TestValidateCapabilities_OverDeclared    PASS
=== RUN   TestValidateCapabilities_UnderDeclared   PASS
=== RUN   TestValidateCapabilities_ProviderNameMismatch  PASS
=== RUN   TestValidateCapabilities_AWSLike         PASS
=== RUN   TestConformanceSuite_FakeCloud           PASS  (22 subtests)
=== RUN   TestConformanceSuite_LimitedCloud        PASS  (1 subtest)

ok  k8s.io/cloud-provider/testing
```

---

## Summary

The interface is two methods. The rest is a map of capabilities. Cloud providers import the package, fill the map in an `init()` function, and their tests start getting skipped or run based on what they actually support — not based on who they are.

No existing tests break. No existing providers need changes to keep working. Adoption is opt-in and incremental. Contract validation and conformance suites enforce honesty, all without requiring a running cluster.
