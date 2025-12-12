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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMInstancesTester implements the InstancesTester interface for Cloud Controller Manager instances tests.
// It provides generic test logic and delegates cloud-specific operations to the cloud provider interface.
type CCMInstancesTester struct {
	// Cloud provider interface can be accessed through framework.TestContext.CloudConfig.Provider
	// The actual cloudprovider.Interface is not directly accessible, so cloud providers
	// implementing this should provide their own implementation that accesses the cloud provider.
}

// NewCCMInstancesTester creates a new CCMInstancesTester instance.
func NewCCMInstancesTester() InstancesTester {
	return &CCMInstancesTester{}
}

// TestNodeAddresses tests the NodeAddresses functionality.
// This test verifies that the cloud provider can retrieve node addresses for a given node name.
func (c *CCMInstancesTester) TestNodeAddresses(ctx context.Context, nodeName types.NodeName) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.NodeAddresses
	return NewSkippedTestResult("skipped - TestNodeAddresses not implemented"), nil
}

// TestNodeAddressesByProviderID tests the NodeAddressesByProviderID functionality.
// This test verifies that the cloud provider can retrieve node addresses using the provider ID.
func (c *CCMInstancesTester) TestNodeAddressesByProviderID(ctx context.Context, providerID string) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.NodeAddressesByProviderID
	return NewSkippedTestResult("skipped - TestNodeAddressesByProviderID not implemented"), nil
}

// TestInstanceID tests the InstanceID functionality.
// This test verifies that the cloud provider can retrieve the instance ID for a given node name.
func (c *CCMInstancesTester) TestInstanceID(ctx context.Context, nodeName types.NodeName) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.InstanceID
	return NewSkippedTestResult("skipped - TestInstanceID not implemented"), nil
}

// TestInstanceType tests the InstanceType functionality.
// This test verifies that the cloud provider can retrieve the instance type for a given node name.
func (c *CCMInstancesTester) TestInstanceType(ctx context.Context, nodeName types.NodeName) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.InstanceType
	return NewSkippedTestResult("skipped - TestInstanceType not implemented"), nil
}

// TestInstanceTypeByProviderID tests the InstanceTypeByProviderID functionality.
// This test verifies that the cloud provider can retrieve the instance type using the provider ID.
func (c *CCMInstancesTester) TestInstanceTypeByProviderID(ctx context.Context, providerID string) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.InstanceTypeByProviderID
	return NewSkippedTestResult("skipped - TestInstanceTypeByProviderID not implemented"), nil
}

// TestAddSSHKeyToAllInstances tests the AddSSHKeyToAllInstances functionality.
// This test verifies that the cloud provider can add an SSH public key to all instances.
func (c *CCMInstancesTester) TestAddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.AddSSHKeyToAllInstances
	return NewSkippedTestResult("skipped - TestAddSSHKeyToAllInstances not implemented"), nil
}

// TestCurrentNodeName tests the CurrentNodeName functionality.
// This test verifies that the cloud provider can determine the current node name from the hostname.
func (c *CCMInstancesTester) TestCurrentNodeName(ctx context.Context, hostname string) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.CurrentNodeName
	return NewSkippedTestResult("skipped - TestCurrentNodeName not implemented"), nil
}

// TestInstanceExistsByProviderID tests the InstanceExistsByProviderID functionality.
// This test verifies that the cloud provider can check if an instance exists using the provider ID.
func (c *CCMInstancesTester) TestInstanceExistsByProviderID(ctx context.Context, providerID string) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.InstanceExistsByProviderID
	return NewSkippedTestResult("skipped - TestInstanceExistsByProviderID not implemented"), nil
}

// TestInstanceShutdownByProviderID tests the InstanceShutdownByProviderID functionality.
// This test verifies that the cloud provider can check if an instance is shutdown using the provider ID.
func (c *CCMInstancesTester) TestInstanceShutdownByProviderID(ctx context.Context, providerID string) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.Instances.InstanceShutdownByProviderID
	return NewSkippedTestResult("skipped - TestInstanceShutdownByProviderID not implemented"), nil
}
