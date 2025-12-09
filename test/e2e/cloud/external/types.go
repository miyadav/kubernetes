/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package external

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider"
	clientset "k8s.io/client-go/kubernetes"
)

// TestResult provides rich information about test execution results.
// This struct allows test methods to return detailed information beyond
// a simple error, enabling better test reporting and handling of skipped tests.
type TestResult struct {
	// Success indicates whether the test passed successfully.
	// If true, the test completed without errors.
	Success bool

	// Error contains the error if the test failed.
	// This should be nil if Success is true or if Skipped is true.
	Error error

	// Skipped indicates whether the test was skipped.
	// A test should be marked as skipped if the underlying feature
	// is not implemented by the cloud provider.
	Skipped bool

	// Message provides a human-readable description of the test result.
	// This can be used for logging and test reporting.
	Message string

	// Details contains additional test-specific information.
	// This map can be used to store arbitrary key-value pairs
	// that provide context about the test execution.
	Details map[string]interface{}
}

// NewTestResult creates a new TestResult with the specified success status.
func NewTestResult(success bool, message string) TestResult {
	return TestResult{
		Success: success,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewSkippedTestResult creates a TestResult indicating the test was skipped.
func NewSkippedTestResult(reason string) TestResult {
	return TestResult{
		Skipped: true,
		Message: reason,
		Details: make(map[string]interface{}),
	}
}

// NewFailedTestResult creates a TestResult indicating the test failed.
func NewFailedTestResult(err error, message string) TestResult {
	return TestResult{
		Success: false,
		Error:   err,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewSuccessTestResult creates a TestResult indicating the test passed.
func NewSuccessTestResult(message string) TestResult {
	return TestResult{
		Success: true,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NodeTester defines the interface for node cloud provider tests.
// Based on the tests in e2e/cloud/nodes.go
type NodeTester interface {
	// TestNodeDeletedOnAPIServerWhenNotInCloudProvider tests that a node
	// should be deleted on API server if it doesn't exist in the cloud provider.
	// Returns TestResult for rich information and error for compatibility.
	// If the feature is not implemented, TestResult.Skipped should be true.
	TestNodeDeletedOnAPIServerWhenNotInCloudProvider(ctx context.Context, c clientset.Interface) (TestResult, error)

	// DeleteNodeOnCloudProvider deletes the specified node from the cloud provider.
	// This is a helper method used by test implementations to perform
	// cloud-specific node deletion operations.
	DeleteNodeOnCloudProvider(node *v1.Node) error
}

// LoadBalancerTester defines the interface for load balancer cloud provider tests.
type LoadBalancerTester interface {
	TestGetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (TestResult, error)
	TestGetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) (TestResult, error)
	TestEnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (TestResult, error)
	TestUpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (TestResult, error)
	TestEnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) (TestResult, error)
}

// InstancesTester defines the interface for instances cloud provider tests.
type InstancesTester interface {
	TestNodeAddresses(ctx context.Context, nodeName types.NodeName) (TestResult, error)
	TestNodeAddressesByProviderID(ctx context.Context, providerID string) (TestResult, error)
	TestInstanceID(ctx context.Context, nodeName types.NodeName) (TestResult, error)
	TestInstanceType(ctx context.Context, nodeName types.NodeName) (TestResult, error)
	TestInstanceTypeByProviderID(ctx context.Context, providerID string) (TestResult, error)
	TestAddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) (TestResult, error)
	TestCurrentNodeName(ctx context.Context, hostname string) (TestResult, error)
	TestInstanceExistsByProviderID(ctx context.Context, providerID string) (TestResult, error)
	TestInstanceShutdownByProviderID(ctx context.Context, providerID string) (TestResult, error)
}

// InstancesV2Tester defines the interface for InstancesV2 cloud provider tests.
type InstancesV2Tester interface {
	TestInstanceExists(ctx context.Context, node *v1.Node) (TestResult, error)
	TestInstanceShutdown(ctx context.Context, node *v1.Node) (TestResult, error)
	TestInstanceMetadata(ctx context.Context, node *v1.Node) (TestResult, error)
}

// RoutesTester defines the interface for routes cloud provider tests.
type RoutesTester interface {
	TestListRoutes(ctx context.Context, clusterName string) (TestResult, error)
	TestCreateRoute(ctx context.Context, clusterName string, nameHint string, route *cloudprovider.Route) (TestResult, error)
	TestDeleteRoute(ctx context.Context, clusterName string, route *cloudprovider.Route) (TestResult, error)
}

// ZonesTester defines the interface for zones cloud provider tests.
type ZonesTester interface {
	TestGetZone(ctx context.Context) (TestResult, error)
	TestGetZoneByProviderID(ctx context.Context, providerID string) (TestResult, error)
	TestGetZoneByNodeName(ctx context.Context, nodeName types.NodeName) (TestResult, error)
}

// ClustersTester defines the interface for clusters cloud provider tests.
type ClustersTester interface {
	TestListClusters(ctx context.Context) (TestResult, error)
	TestMaster(ctx context.Context, clusterName string) (TestResult, error)
}

// Tester is the main interface for cloud provider testing.
// It mirrors the cloudprovider.Interface pattern, allowing cloud providers
// to implement test interfaces for their specific capabilities.
type Tester interface {
	LoadBalancerTester() (LoadBalancerTester, bool)
	InstancesTester() (InstancesTester, bool)
	InstancesV2Tester() (InstancesV2Tester, bool)
	RoutesTester() (RoutesTester, bool)
	ZonesTester() (ZonesTester, bool)
	ClustersTester() (ClustersTester, bool)
	NodeTester() (NodeTester, bool)
	ProviderName() string
}

