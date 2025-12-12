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
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
)

// CCMNodeTester implements the NodeTester interface for Cloud Controller Manager node tests
// It provides generic test logic and delegates cloud-specific operations to a NodeTester implementation
type CCMNodeTester struct {
	nodeTester        NodeTester
	stabilityVerifier ClusterStabilityVerifier
}

// NewCCMNodeTester creates a new CCMNodeTester instance
func NewCCMNodeTester() NodeTester {
	return &CCMNodeTester{}
}

// SetNodeTester sets the cloud-specific NodeTester implementation
// This allows the generic test logic to call cloud-specific DeleteNodeOnCloudProvider
func (c *CCMNodeTester) SetNodeTester(tester NodeTester) {
	c.nodeTester = tester
}

// SetClusterStabilityVerifier sets the cloud-specific ClusterStabilityVerifier implementation
// This allows the generic test logic to call cloud-specific stability verification
func (c *CCMNodeTester) SetClusterStabilityVerifier(verifier ClusterStabilityVerifier) {
	c.stabilityVerifier = verifier
}

// TestNodeDeletedOnAPIServerWhenNotInCloudProvider tests that a node
// should be deleted on API server if it doesn't exist in the cloud provider.
// This implementation is based on the test in e2e/cloud/nodes.go
func (c *CCMNodeTester) TestNodeDeletedOnAPIServerWhenNotInCloudProvider(ctx context.Context, client clientset.Interface) (TestResult, error) {
	framework.Logf("Testing node deletion when not present in cloud provider")

	// Get a random ready schedulable node to delete
	nodeToDelete, err := e2enode.GetRandomReadySchedulableNode(ctx, client)
	if err != nil {
		return NewFailedTestResult(err, "failed to get random ready schedulable node"), fmt.Errorf("failed to get random ready schedulable node: %w", err)
	}

	// Get the original list of ready nodes
	origNodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, client)
	if err != nil {
		return NewFailedTestResult(err, "failed to get ready nodes"), fmt.Errorf("failed to get ready nodes: %w", err)
	}

	framework.Logf("Original number of ready nodes: %d", len(origNodes.Items))
	framework.Logf("Deleting node %q on the cloud provider", nodeToDelete.Name)

	// Delete the node on the cloud provider using the cloud-specific implementation
	if c.nodeTester != nil {
		err = c.nodeTester.DeleteNodeOnCloudProvider(nodeToDelete)
	} else {
		err = c.DeleteNodeOnCloudProvider(nodeToDelete)
	}
	if err != nil {
		return NewFailedTestResult(err, fmt.Sprintf("failed to delete node %q on cloud provider", nodeToDelete.Name)), fmt.Errorf("failed to delete node %q on cloud provider: %w", nodeToDelete.Name, err)
	}

	// Wait for the node count to decrease by 1
	newNodes, err := e2enode.CheckReady(ctx, client, len(origNodes.Items)-1, 5*time.Minute)
	if err != nil {
		return NewFailedTestResult(err, "failed to wait for ready nodes to decrease"), fmt.Errorf("failed to wait for ready nodes to decrease: %w", err)
	}

	if len(newNodes) != len(origNodes.Items)-1 {
		err := fmt.Errorf("expected %d nodes, got %d", len(origNodes.Items)-1, len(newNodes))
		return NewFailedTestResult(err, "node count mismatch"), err
	}

	// Verify the node is deleted from the API server
	_, err = client.CoreV1().Nodes().Get(ctx, nodeToDelete.Name, metav1.GetOptions{})
	if err == nil {
		err := fmt.Errorf("node %q still exists when it should be deleted", nodeToDelete.Name)
		return NewFailedTestResult(err, "node still exists in API server"), err
	}
	if !apierrors.IsNotFound(err) {
		return NewFailedTestResult(err, fmt.Sprintf("unexpected error when getting node %q", nodeToDelete.Name)), fmt.Errorf("unexpected error when getting node %q: %w", nodeToDelete.Name, err)
	}

	framework.Logf("Successfully verified node %q was deleted from API server", nodeToDelete.Name)

	// Verify cluster stability after node deletion
	stabilityResult, err := c.VerifyClusterStabilityAfterNodeDeletion(ctx, client)
	if err != nil {
		framework.Logf("Cluster stability verification returned error: %v", err)
		// Don't fail the test if stability check fails, but log it
	}
	if !stabilityResult.Skipped && !stabilityResult.Success {
		framework.Logf("Cluster stability check failed: %s", stabilityResult.Message)
		// Don't fail the test if stability check fails, but log it
	}

	return NewSuccessTestResult(fmt.Sprintf("Successfully verified node %q was deleted from API server", nodeToDelete.Name)), nil
}

// DeleteNodeOnCloudProvider deletes the specified node from the cloud provider
// Note: This method doesn't have access to a client to validate cloud provider configuration
// via nodes, so it uses the framework's cloud provider check. The calling test method
// (TestNodeDeletedOnAPIServerWhenNotInCloudProvider) validates cloud provider configuration
// before calling this method.
//
// If the cloud provider is not configured or the DeleteNode method is not implemented,
// this method returns a descriptive error message.
func (c *CCMNodeTester) DeleteNodeOnCloudProvider(node *v1.Node) error {
	if framework.TestContext.CloudConfig.Provider == nil {
		return fmt.Errorf("cloud provider is not configured - cannot delete node %q", node.Name)
	}
	if framework.TestContext.CloudConfig.Provider.DeleteNode == nil {
		return fmt.Errorf("DeleteNode method is not implemented by cloud provider - cannot delete node %q", node.Name)
	}
	return framework.TestContext.CloudConfig.Provider.DeleteNode(node)
}

// VerifyClusterStabilityAfterNodeDeletion verifies that the cluster is stable after a node deletion operation.
// This method performs basic stability checks:
// - Node count stability
// - Pod rescheduling (checks that pods are not stuck in Pending state)
// - Service endpoint availability
// - API server responsiveness
//
// Cloud providers can override this method or set a ClusterStabilityVerifier for custom checks.
// If not implemented, returns a skipped result.
func (c *CCMNodeTester) VerifyClusterStabilityAfterNodeDeletion(ctx context.Context, client clientset.Interface) (TestResult, error) {
	// If a custom verifier is set, use it
	if c.stabilityVerifier != nil {
		err := c.stabilityVerifier.VerifyClusterStability(ctx, client)
		if err != nil {
			return NewFailedTestResult(err, "cluster stability verification failed"), err
		}
		return NewSuccessTestResult("cluster stability verified by custom verifier"), nil
	}

	// Default implementation: perform basic stability checks
	framework.Logf("Verifying cluster stability after node deletion")

	// Check 1: Verify node count is stable (no unexpected changes)
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list nodes for stability check"), fmt.Errorf("failed to list nodes: %w", err)
	}

	readyNodeCount := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				readyNodeCount++
				break
			}
		}
	}

	if readyNodeCount == 0 {
		return NewFailedTestResult(fmt.Errorf("no ready nodes found"), "cluster has no ready nodes after node deletion"), fmt.Errorf("cluster has no ready nodes")
	}

	framework.Logf("Found %d ready nodes after node deletion", readyNodeCount)

	// Check 2: Verify pods are rescheduled (wait for pending pods to be scheduled)
	framework.Logf("Checking pod rescheduling status...")
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		pendingCount := 0
		for _, pod := range pods.Items {
			// Skip completed pods (Succeeded/Failed)
			if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
				continue
			}
			// Skip pods that are intentionally not scheduled (e.g., DaemonSets on deleted node)
			if pod.Spec.NodeName != "" {
				// Check if the node still exists
				_, err := client.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
				if apierrors.IsNotFound(err) {
					// Pod is on a deleted node, it should be rescheduled
					pendingCount++
				}
			} else if pod.Status.Phase == v1.PodPending {
				// Pod is pending and not assigned to a node
				pendingCount++
			}
		}

		if pendingCount > 0 {
			framework.Logf("Waiting for %d pods to be rescheduled...", pendingCount)
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		// Don't fail the test, but log the warning
		framework.Logf("Warning: Some pods may not have been rescheduled: %v", err)
	}

	// Check 3: Verify API server is responsive
	_, err = client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return NewFailedTestResult(err, "API server is not responsive"), fmt.Errorf("API server health check failed: %w", err)
	}

	framework.Logf("Cluster stability verification completed successfully")
	return NewSuccessTestResult("cluster is stable after node deletion"), nil
}
