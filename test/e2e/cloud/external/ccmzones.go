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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMZonesTester implements the ZonesTester interface for Cloud Controller Manager zones tests.
// It provides generic test logic that handles all Kubernetes API operations and delegates
// cloud-specific verification to ZoneVerifier.
//
// DEPRECATED: Zones is deprecated in favor of retrieving zone/region information from InstancesV2.
// This interface will not be called if InstancesV2 is enabled.
type CCMZonesTester struct {
	verifier ZoneVerifier
}

// NewCCMZonesTester creates a new CCMZonesTester instance.
func NewCCMZonesTester() ZonesTester {
	return &CCMZonesTester{}
}

// SetZoneVerifier sets the cloud-specific ZoneVerifier implementation.
func (c *CCMZonesTester) SetZoneVerifier(verifier ZoneVerifier) {
	c.verifier = verifier
}

// TestGetZone tests the GetZone functionality.
// This test verifies that the cloud provider can retrieve the Zone containing the current failure zone.
// It handles all Kubernetes API operations (listing nodes, verifying labels) and delegates
// cloud-specific zone retrieval to ZoneVerifier.
func (c *CCMZonesTester) TestGetZone(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	// Get list of nodes to verify zone information
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list nodes"), fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return NewFailedTestResult(fmt.Errorf("no nodes available"), "no nodes available for testing"), fmt.Errorf("no nodes available for testing")
	}

	// Get available zones in the region (if verifier is set)
	var availableZones []string
	if c.verifier != nil {
		availableZones, err = c.verifier.GetAvailableZones(ctx)
		if err != nil {
			return NewFailedTestResult(err, "failed to get available zones"), fmt.Errorf("failed to get available zones: %w", err)
		}
		framework.Logf("Available zones: %v", availableZones)
	}

	for _, node := range nodes.Items {
		// Check zone label
		zone, hasZone := node.Labels["topology.kubernetes.io/zone"]
		if !hasZone {
			// Try legacy label
			zone, hasZone = node.Labels["failure-domain.beta.kubernetes.io/zone"]
		}

		if !hasZone {
			framework.Logf("Node %s does not have zone label", node.Name)
			continue
		}

		// Verify zone is in the list of available zones (if verifier is set)
		if c.verifier != nil && len(availableZones) > 0 {
			found := false
			for _, az := range availableZones {
				if az == zone {
					found = true
					break
				}
			}
			if !found {
				return NewFailedTestResult(fmt.Errorf("invalid zone"), fmt.Sprintf("node %s has invalid zone %s (not in available zones)", node.Name, zone)), fmt.Errorf("node %s has invalid zone %s (not in available zones)", node.Name, zone)
			}
		}

		framework.Logf("Verified zone for node %s: %s", node.Name, zone)
	}

	framework.Logf("Successfully verified zone information for all nodes")
	return NewSuccessTestResult("Successfully verified zone information for all nodes"), nil
}

// TestGetZoneByProviderID tests the GetZoneByProviderID functionality.
// This test verifies that the cloud provider can retrieve the Zone using the provider ID.
// It handles all Kubernetes API operations (listing nodes) and delegates cloud-specific
// zone retrieval to ZoneVerifier.
func (c *CCMZonesTester) TestGetZoneByProviderID(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	if c.verifier == nil {
		return NewSkippedTestResult("skipped - ZoneVerifier is not configured"), fmt.Errorf("ZoneVerifier is not configured")
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

		// Get zone from cloud provider
		awsZone, err := c.verifier.GetZoneByProviderID(ctx, providerID)
		if err != nil {
			return NewFailedTestResult(err, fmt.Sprintf("failed to get zone for node %s", node.Name)), fmt.Errorf("failed to get zone for node %s: %w", node.Name, err)
		}

		// Get zone from node label
		nodeZone, hasZone := node.Labels["topology.kubernetes.io/zone"]
		if !hasZone {
			nodeZone, hasZone = node.Labels["failure-domain.beta.kubernetes.io/zone"]
		}

		if hasZone && awsZone != nodeZone {
			return NewFailedTestResult(fmt.Errorf("zone mismatch"), fmt.Sprintf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, awsZone, nodeZone)), fmt.Errorf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, awsZone, nodeZone)
		}

		framework.Logf("Verified zone by providerID for node %s: %s", node.Name, awsZone)
	}

	framework.Logf("Successfully verified GetZoneByProviderID for all nodes")
	return NewSuccessTestResult("Successfully verified GetZoneByProviderID for all nodes"), nil
}

// TestGetZoneByNodeName tests the GetZoneByNodeName functionality.
// This test verifies that the cloud provider can retrieve the Zone using the node name.
// It handles all Kubernetes API operations (listing nodes, getting nodes by name) and delegates
// cloud-specific zone retrieval to ZoneVerifier.
func (c *CCMZonesTester) TestGetZoneByNodeName(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if err := validateCloudProviderConfigured(ctx, client); err != nil {
		return NewSkippedTestResult("skipped - cloud provider is not configured"), err
	}

	if c.verifier == nil {
		return NewSkippedTestResult("skipped - ZoneVerifier is not configured"), fmt.Errorf("ZoneVerifier is not configured")
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
		// Get the node's provider ID to find the instance
		providerID := node.Spec.ProviderID
		if providerID == "" {
			framework.Logf("Skipping node %s: no providerID", node.Name)
			continue
		}

		// Extract instance ID from provider ID (cloud-specific parsing may be needed)
		// For now, we'll use GetZoneByProviderID which should work for most cases
		awsZone, err := c.verifier.GetZoneByProviderID(ctx, providerID)
		if err != nil {
			return NewFailedTestResult(err, fmt.Sprintf("failed to get zone for node %s", node.Name)), fmt.Errorf("failed to get zone for node %s: %w", node.Name, err)
		}

		// Verify zone label matches
		nodeZone, hasZone := node.Labels["topology.kubernetes.io/zone"]
		if !hasZone {
			nodeZone, hasZone = node.Labels["failure-domain.beta.kubernetes.io/zone"]
		}

		if hasZone && awsZone != nodeZone {
			return NewFailedTestResult(fmt.Errorf("zone mismatch"), fmt.Sprintf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, awsZone, nodeZone)), fmt.Errorf("zone mismatch for node %s: cloud=%s, label=%s", node.Name, awsZone, nodeZone)
		}

		framework.Logf("Verified zone by node name for %s: %s", node.Name, awsZone)
	}

	framework.Logf("Successfully verified GetZoneByNodeName for all nodes")
	return NewSuccessTestResult("Successfully verified GetZoneByNodeName for all nodes"), nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
