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

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMInstancesV2Tester implements the InstancesV2Tester interface for Cloud Controller Manager InstancesV2 tests.
// It provides generic test logic and delegates cloud-specific operations to the cloud provider interface.
type CCMInstancesV2Tester struct {
	// Cloud provider interface can be accessed through framework.TestContext.CloudConfig.Provider
	// The actual cloudprovider.Interface is not directly accessible, so cloud providers
	// implementing this should provide their own implementation that accesses the cloud provider.
}

// NewCCMInstancesV2Tester creates a new CCMInstancesV2Tester instance.
func NewCCMInstancesV2Tester() InstancesV2Tester {
	return &CCMInstancesV2Tester{}
}

// TestInstanceExists tests the InstanceExists functionality.
// This test verifies that the cloud provider can check if an instance exists for a given node.
func (c *CCMInstancesV2Tester) TestInstanceExists(ctx context.Context, node *v1.Node) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.InstancesV2.InstanceExists
	return NewSkippedTestResult("TestInstanceExists not yet implemented - cloud providers should implement this"), nil
}

// TestInstanceShutdown tests the InstanceShutdown functionality.
// This test verifies that the cloud provider can check if an instance is shutdown for a given node.
func (c *CCMInstancesV2Tester) TestInstanceShutdown(ctx context.Context, node *v1.Node) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.InstancesV2.InstanceShutdown
	return NewSkippedTestResult("TestInstanceShutdown not yet implemented - cloud providers should implement this"), nil
}

// TestInstanceMetadata tests the InstanceMetadata functionality.
// This test verifies that the cloud provider can retrieve instance metadata for a given node.
func (c *CCMInstancesV2Tester) TestInstanceMetadata(ctx context.Context, node *v1.Node) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.InstancesV2.InstanceMetadata
	return NewSkippedTestResult("TestInstanceMetadata not yet implemented - cloud providers should implement this"), nil
}

