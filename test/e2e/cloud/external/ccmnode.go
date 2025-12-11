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
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
)

// CCMNodeTester implements the NodeTester interface for Cloud Controller Manager node tests
// It provides generic test logic and delegates cloud-specific operations to a NodeTester implementation
type CCMNodeTester struct {
	nodeTester NodeTester
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
	return NewSuccessTestResult(fmt.Sprintf("Successfully verified node %q was deleted from API server", nodeToDelete.Name)), nil
}

// DeleteNodeOnCloudProvider deletes the specified node from the cloud provider
// Note: This method doesn't have access to a client to validate cloud provider configuration
// via nodes, so it uses the framework's cloud provider check. The calling test method
// (TestNodeDeletedOnAPIServerWhenNotInCloudProvider) validates cloud provider configuration
// before calling this method.
func (c *CCMNodeTester) DeleteNodeOnCloudProvider(node *v1.Node) error {
	if framework.TestContext.CloudConfig.Provider == nil {
		return fmt.Errorf("cloud provider is not configured")
	}
	return framework.TestContext.CloudConfig.Provider.DeleteNode(node)
}
