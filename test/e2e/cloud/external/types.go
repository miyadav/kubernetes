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
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
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

// LoadBalancerVerifier defines cloud-specific operations for load balancer verification.
// Cloud providers should implement this interface to provide cloud-specific verification logic.
type LoadBalancerVerifier interface {
	// VerifyLoadBalancerExists checks if a load balancer exists in the cloud provider
	// by its hostname or IP address. Returns true if the load balancer exists and is active.
	VerifyLoadBalancerExists(ctx context.Context, hostnameOrIP string) (bool, error)
}

// LoadBalancerTester defines the interface for load balancer cloud provider tests.
// The methods accept clientset.Interface to handle all Kubernetes API operations,
// while cloud-specific verification is delegated to LoadBalancerVerifier.
type LoadBalancerTester interface {
	TestGetLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestGetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) (TestResult, error)
	TestEnsureLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestUpdateLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestEnsureLoadBalancerDeleted(ctx context.Context, client clientset.Interface) (TestResult, error)
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

// InstanceV2Verifier defines cloud-specific operations for instance verification.
// Cloud providers should implement this interface to provide cloud-specific verification logic.
type InstanceV2Verifier interface {
	// VerifyInstanceExists checks if an instance exists in the cloud provider for the given node.
	VerifyInstanceExists(ctx context.Context, node *v1.Node) (bool, error)
	// VerifyInstanceShutdown checks if an instance is shutdown in the cloud provider for the given node.
	VerifyInstanceShutdown(ctx context.Context, node *v1.Node) (bool, error)
	// GetInstanceMetadata retrieves instance metadata from the cloud provider for the given node.
	GetInstanceMetadata(ctx context.Context, node *v1.Node) (map[string]interface{}, error)
}

// InstancesV2Tester defines the interface for InstancesV2 cloud provider tests.
// The methods accept clientset.Interface to handle all Kubernetes API operations,
// while cloud-specific verification is delegated to InstanceV2Verifier.
type InstancesV2Tester interface {
	TestInstanceExists(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestInstanceShutdown(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestInstanceMetadata(ctx context.Context, client clientset.Interface) (TestResult, error)
}

// RoutesTester defines the interface for routes cloud provider tests.
type RoutesTester interface {
	TestListRoutes(ctx context.Context, clusterName string) (TestResult, error)
	TestCreateRoute(ctx context.Context, clusterName string, nameHint string, route *cloudprovider.Route) (TestResult, error)
	TestDeleteRoute(ctx context.Context, clusterName string, route *cloudprovider.Route) (TestResult, error)
}

// ZoneVerifier defines cloud-specific operations for zone verification.
// Cloud providers should implement this interface to provide cloud-specific verification logic.
type ZoneVerifier interface {
	// GetZoneByProviderID retrieves the zone from the cloud provider using the provider ID.
	GetZoneByProviderID(ctx context.Context, providerID string) (string, error)
	// GetZoneByInstanceID retrieves the zone from the cloud provider using the instance ID.
	GetZoneByInstanceID(ctx context.Context, instanceID string) (string, error)
	// GetAvailableZones returns the list of available zones in the region.
	GetAvailableZones(ctx context.Context) ([]string, error)
}

// ZonesTester defines the interface for zones cloud provider tests.
// The methods accept clientset.Interface to handle all Kubernetes API operations,
// while cloud-specific verification is delegated to ZoneVerifier.
type ZonesTester interface {
	TestGetZone(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestGetZoneByProviderID(ctx context.Context, client clientset.Interface) (TestResult, error)
	TestGetZoneByNodeName(ctx context.Context, client clientset.Interface) (TestResult, error)
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

// validateCloudProviderConfigured validates that a cloud provider is configured
// by checking nodes for cloud provider indicators. It checks:
// 1. Cloud provider annotation (cloudprovider.kubernetes.io/provider-name)
// 2. ProviderID format (aws://, gce://, azure://, etc.)
// 3. Node labels (topology.kubernetes.io/region)
// Returns an error if no cloud provider indicators are found on any node, nil otherwise.
func validateCloudProviderConfigured(ctx context.Context, client clientset.Interface) error {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found")
	}

	// Common cloud provider ProviderID prefixes
	cloudProviderPrefixes := []string{
		"aws://",
		"gce://",
		"azure://",
		"vsphere://",
		"openstack://",
		"cloudstack://",
		"ovirt://",
		"photon://",
		"alicloud://",
		"tencent://",
		"huawei://",
		"baiducloud://",
		"ibmcloud://",
		"kubemark://",
		"external://",
	}

	// Check if at least one node has cloud provider indicators
	for _, node := range nodes.Items {
		// Check 1: Cloud provider annotation (if present, use it)
		if providerName, ok := node.Annotations["cloudprovider.kubernetes.io/provider-name"]; ok {
			if providerName != "" {
				return nil // Node has cloud provider configured
			}
		}

		// Check 2: ProviderID format
		providerID := node.Spec.ProviderID
		if providerID != "" {
			for _, prefix := range cloudProviderPrefixes {
				if strings.HasPrefix(providerID, prefix) {
					return nil // ProviderID indicates cloud provider
				}
			}
		}

		// Check 3: Node labels (fallback)
		if _, ok := node.Labels["topology.kubernetes.io/region"]; ok {
			return nil // Region label suggests cloud provider
		}
	}

	// If we get here, no nodes had cloud provider indicators
	return fmt.Errorf("cloud provider is not configured (no annotation, invalid ProviderID format, or missing region label found on any node)")
}
