# External Cloud Provider Testing Framework

This directory contains a standardized testing framework for external cloud-controller-manager implementations. It enables cloud providers to test their cloud controller implementations against a common set of test cases.

## Overview

The cloud-controller-manager is responsible for three main functions:

1. **Node Controller**: Initializing nodes with cloud-specific labels, detecting node deletions, and managing node lifecycle
2. **Route Controller**: Configuring routes in the cloud so containers on different nodes can communicate
3. **Service Controller**: Provisioning load balancers for services of type LoadBalancer

This framework provides test interfaces and test suites for each of these controllers.

See: https://kubernetes.io/docs/concepts/architecture/cloud-controller/#functions-of-the-ccm

## How to Use

### For Cloud Providers

Cloud providers should implement the `TestInterface` in their own repository and register it with this framework. The interface uses a capability-reporting pattern where each method returns `(implemented bool, interface)` to indicate which features are supported.

#### Step 1: Implement the TestInterface

In your cloud provider repository (e.g., `k8s.io/cloud-provider-aws`), create a test driver:

```go
package e2e

import (
    "context"
    v1 "k8s.io/api/core/v1"
    "k8s.io/kubernetes/test/e2e/cloud/external"
)

type awsTestDriver struct {
    // Your cloud-specific fields
}

// Implement external.TestInterface
func (d *awsTestDriver) NodeController() (bool, external.TestNodeControllerInterface) {
    return true, &awsNodeController{driver: d}
}

func (d *awsTestDriver) RouteController() (bool, external.TestRouteControllerInterface) {
    return true, &awsRouteController{driver: d}
}

func (d *awsTestDriver) ServiceController() (bool, external.TestServiceControllerInterface) {
    return true, &awsServiceController{driver: d}
}
```

#### Step 2: Implement the Controller Interfaces

Implement each controller interface you support:

```go
type awsNodeController struct {
    driver *awsTestDriver
}

func (c *awsNodeController) NodeExists(ctx context.Context, node *v1.Node) (bool, error) {
    // Call your cloud provider API to check if the node exists
}

func (c *awsNodeController) NodeShutdown(ctx context.Context, node *v1.Node) (bool, error) {
    // Call your cloud provider API to check if the node is shutdown
}

func (c *awsNodeController) NodeMetadata(ctx context.Context, node *v1.Node) (*external.NodeMetadata, error) {
    // Return node metadata from your cloud provider
}

func (c *awsNodeController) DeleteNode(ctx context.Context, node *v1.Node) error {
    // Delete the node from your cloud provider
}
```

#### Step 3: Register Your Test Driver

Register your test driver in an init function:

```go
func init() {
    err := external.RegisterTestDriver(&awsTestDriver{})
    if err != nil {
        panic(err)
    }
}
```

#### Step 4: Import Your Driver in the Test Binary

Add an import for your test driver package in `test/e2e/cloud/external/imports.go`:

```go
import _ "k8s.io/cloud-provider-aws/test/e2e"
```

### For Tests Not Yet Implemented

If your cloud provider doesn't support a particular controller, simply return `false` for the implemented flag:

```go
func (d *myDriver) RouteController() (bool, external.TestRouteControllerInterface) {
    // We don't support routes yet
    return false, nil
}
```

The test suite will automatically skip tests for unimplemented features.

## Running the Tests

To run the tests for your cloud provider:

```bash
# Build the e2e test binary with your driver imported
make WHAT=test/e2e/e2e.test

# Run the external cloud provider tests
ginkgo -focus="cloud-provider-external" ./test/e2e/e2e.test -- \
    --provider=<your-provider> \
    --kubeconfig=$KUBECONFIG
```

You can also run specific controller tests:

```bash
# Run only node controller tests
ginkgo -focus="cloud-provider-external.*Node Controller" ./test/e2e/e2e.test

# Run only route controller tests
ginkgo -focus="cloud-provider-external.*Route Controller" ./test/e2e/e2e.test

# Run only service controller tests
ginkgo -focus="cloud-provider-external.*Service Controller" ./test/e2e/e2e.test
```

## Test Coverage

### Node Controller Tests

- ✓ Correctly report that a node exists
- ✓ Correctly report node metadata (provider ID, instance type, addresses)
- ✓ Detect when a node is deleted from the cloud
- ✓ Correctly detect node shutdown state
- ✓ Report node as non-existent after deletion

### Route Controller Tests

- ✓ List routes
- ✓ Create and delete routes
- ✓ Create routes with correct target node information
- ✓ Handle blackhole routes

### Service Controller Tests

- ✓ Create a load balancer for a service
- ✓ Update a load balancer when nodes change
- ✓ Delete a load balancer when requested
- ✓ Handle multiple ports on a load balancer

## Compatibility with Existing Tests

This framework is designed to work alongside the existing cloud provider test infrastructure. The `TestNodeControllerInterface.DeleteNode` method signature is compatible with the existing `framework.ProviderInterface.DeleteNode`, allowing tests to use both the new and old patterns.

## Example Implementations

For reference implementations, see:
- `k8s.io/cloud-provider-aws/test/e2e` (when available)
- `k8s.io/cloud-provider-azure/test/e2e` (when available)
- `k8s.io/cloud-provider-gcp/test/e2e` (when available)

## Contributing

When adding new test cases:

1. Follow the existing test patterns in `nodes.go`, `routes.go`, and `loadbalancer.go`
2. Use the capability-reporting pattern to allow cloud providers to skip unsupported features
3. Add appropriate cleanup in defer blocks to ensure test resources are cleaned up
4. Update this README with new test coverage

## References

- [Cloud Controller Manager Concepts](https://kubernetes.io/docs/concepts/architecture/cloud-controller/)
- [Kubernetes AI Guidance for Pull Requests](https://www.kubernetes.dev/docs/guide/pull-requests/#ai-guidance)
- [Cloud Provider Interface](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/cloud-provider/cloud.go)
