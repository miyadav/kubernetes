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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMInstancesV2Tester implements the InstancesV2Tester interface for Cloud Controller Manager InstancesV2 tests.
// It provides generic test logic that handles all Kubernetes API operations and delegates
// cloud-specific verification to InstanceV2Verifier.
type CCMInstancesV2Tester struct {
	verifier InstanceV2Verifier
}

// NewCCMInstancesV2Tester creates a new CCMInstancesV2Tester instance.
func NewCCMInstancesV2Tester() InstancesV2Tester {
	return &CCMInstancesV2Tester{}
}

// SetInstanceV2Verifier sets the cloud-specific InstanceV2Verifier implementation.
func (c *CCMInstancesV2Tester) SetInstanceV2Verifier(verifier InstanceV2Verifier) {
	c.verifier = verifier
}

// TestInstanceExists tests the InstanceExists functionality.
// This test verifies that the cloud provider can check if an instance exists for a given node.
// It handles all Kubernetes API operations (listing nodes) and delegates cloud-specific
// verification to InstanceV2Verifier.
func (c *CCMInstancesV2Tester) TestInstanceExists(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	if c.verifier == nil {
		return NewSkippedTestResult("skipped - InstanceV2Verifier is not configured"), fmt.Errorf("InstanceV2Verifier is not configured")
	}

	// Get list of nodes
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list nodes"), fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return NewFailedTestResult(fmt.Errorf("no nodes available"), "no nodes available for testing"), fmt.Errorf("no nodes available for testing")
	}

	// Test each node's existence in the cloud provider
	for _, node := range nodes.Items {
		providerID := node.Spec.ProviderID
		if providerID == "" {
			framework.Logf("Skipping node %s: no providerID", node.Name)
			continue
		}

		// Check if instance exists in cloud provider
		exists, err := c.verifier.VerifyInstanceExists(ctx, &node)
		if err != nil {
			return NewFailedTestResult(err, fmt.Sprintf("failed to check instance existence for node %s", node.Name)), fmt.Errorf("failed to check instance existence for node %s: %w", node.Name, err)
		}

		if !exists {
			return NewFailedTestResult(fmt.Errorf("instance does not exist"), fmt.Sprintf("instance for node %s does not exist in cloud provider but node exists in Kubernetes", node.Name)), fmt.Errorf("instance for node %s does not exist in cloud provider but node exists in Kubernetes", node.Name)
		}

		framework.Logf("Verified instance exists for node %s", node.Name)
	}

	framework.Logf("Successfully verified all %d nodes exist in cloud provider", len(nodes.Items))
	return NewSuccessTestResult(fmt.Sprintf("Successfully verified all %d nodes exist in cloud provider", len(nodes.Items))), nil
}

// TestInstanceShutdown tests the InstanceShutdown functionality.
// This test verifies that the cloud provider can check if an instance is shutdown for a given node.
// It handles all Kubernetes API operations (listing nodes) and delegates cloud-specific
// verification to InstanceV2Verifier.
func (c *CCMInstancesV2Tester) TestInstanceShutdown(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	if c.verifier == nil {
		return NewSkippedTestResult("skipped - InstanceV2Verifier is not configured"), fmt.Errorf("InstanceV2Verifier is not configured")
	}

	// Get list of nodes
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list nodes"), fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return NewFailedTestResult(fmt.Errorf("no nodes available"), "no nodes available for testing"), fmt.Errorf("no nodes available for testing")
	}

	// Verify running nodes are not reported as shutdown
	for _, node := range nodes.Items {
		providerID := node.Spec.ProviderID
		if providerID == "" {
			framework.Logf("Skipping node %s: no providerID", node.Name)
			continue
		}

		// Check instance shutdown state
		shutdown, err := c.verifier.VerifyInstanceShutdown(ctx, &node)
		if err != nil {
			return NewFailedTestResult(err, fmt.Sprintf("failed to check shutdown state for node %s", node.Name)), fmt.Errorf("failed to check shutdown state for node %s: %w", node.Name, err)
		}

		// For nodes that are Ready, they should not be shutdown
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				isReady = true
				break
			}
		}

		if isReady && shutdown {
			return NewFailedTestResult(fmt.Errorf("node is ready but instance is shutdown"), fmt.Sprintf("node %s is Ready but instance is reported as shutdown", node.Name)), fmt.Errorf("node %s is Ready but instance is reported as shutdown", node.Name)
		}

		framework.Logf("Verified shutdown state for node %s: shutdown=%v, ready=%v", node.Name, shutdown, isReady)
	}

	framework.Logf("Successfully verified shutdown state for all nodes")
	return NewSuccessTestResult("Successfully verified shutdown state for all nodes"), nil
}

// TestInstanceMetadata tests the InstanceMetadata functionality.
// This test verifies that the cloud provider can retrieve instance metadata for a given node.
// It handles all Kubernetes API operations (listing nodes, verifying labels) and delegates
// cloud-specific metadata retrieval to InstanceV2Verifier.
func (c *CCMInstancesV2Tester) TestInstanceMetadata(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	if c.verifier == nil {
		return NewSkippedTestResult("skipped - InstanceV2Verifier is not configured"), fmt.Errorf("InstanceV2Verifier is not configured")
	}

	// Get list of nodes
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list nodes"), fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return NewFailedTestResult(fmt.Errorf("no nodes available"), "no nodes available for testing"), fmt.Errorf("no nodes available for testing")
	}

	for _, node := range nodes.Items {
		providerID := node.Spec.ProviderID
		if providerID == "" {
			framework.Logf("Skipping node %s: no providerID", node.Name)
			continue
		}

		// Get instance metadata from cloud provider
		metadata, err := c.verifier.GetInstanceMetadata(ctx, &node)
		if err != nil {
			return NewFailedTestResult(err, fmt.Sprintf("failed to get instance metadata for node %s", node.Name)), fmt.Errorf("failed to get instance metadata for node %s: %w", node.Name, err)
		}

		// Verify instance type matches node label (if available)
		if instanceType, ok := metadata["instanceType"].(string); ok {
			if nodeInstanceType, hasLabel := node.Labels["node.kubernetes.io/instance-type"]; hasLabel {
				if instanceType != nodeInstanceType {
					return NewFailedTestResult(fmt.Errorf("instance type mismatch"), fmt.Sprintf("instance type mismatch for node %s: cloud=%s, label=%s", node.Name, instanceType, nodeInstanceType)), fmt.Errorf("instance type mismatch for node %s: cloud=%s, label=%s", node.Name, instanceType, nodeInstanceType)
				}
				framework.Logf("Verified instance type for node %s: %s", node.Name, instanceType)
			}
		}

		// Verify zone matches node label (if available)
		if zone, ok := metadata["zone"].(string); ok {
			if nodeZone, hasLabel := node.Labels["topology.kubernetes.io/zone"]; hasLabel {
				if zone != nodeZone {
					return NewFailedTestResult(fmt.Errorf("zone mismatch"), fmt.Sprintf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, zone, nodeZone)), fmt.Errorf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, zone, nodeZone)
				}
				framework.Logf("Verified zone for node %s: %s", node.Name, zone)
			}
		}

		// Verify region matches node label (if available)
		if region, ok := metadata["region"].(string); ok {
			if nodeRegion, hasLabel := node.Labels["topology.kubernetes.io/region"]; hasLabel {
				if region != nodeRegion {
					return NewFailedTestResult(fmt.Errorf("region mismatch"), fmt.Sprintf("region mismatch for node %s: cloud=%s, label=%s", node.Name, region, nodeRegion)), fmt.Errorf("region mismatch for node %s: cloud=%s, label=%s", node.Name, region, nodeRegion)
				}
				framework.Logf("Verified region for node %s: %s", node.Name, region)
			}
		}

		// Verify at least one node address exists
		if len(node.Status.Addresses) == 0 {
			return NewFailedTestResult(fmt.Errorf("no addresses"), fmt.Sprintf("node %s has no addresses", node.Name)), fmt.Errorf("node %s has no addresses", node.Name)
		}

		framework.Logf("Verified metadata for node %s", node.Name)
	}

	framework.Logf("Successfully verified instance metadata for all nodes")
	return NewSuccessTestResult("Successfully verified instance metadata for all nodes"), nil
}
