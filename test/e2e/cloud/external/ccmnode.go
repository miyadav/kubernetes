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
	cloudprovidertest "k8s.io/cloud-provider/test"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
)

// CCMNodeTester implements the NodeTester interface for Cloud Controller Manager node tests
type CCMNodeTester struct{}

// NewCCMNodeTester creates a new CCMNodeTester instance
func NewCCMNodeTester() cloudprovidertest.NodeTester {
	return &CCMNodeTester{}
}

// TestNodeDeletedOnAPIServerWhenNotInCloudProvider tests that a node
// should be deleted on API server if it doesn't exist in the cloud provider.
// This implementation is based on the test in e2e/cloud/nodes.go
func (c *CCMNodeTester) TestNodeDeletedOnAPIServerWhenNotInCloudProvider(ctx context.Context, client clientset.Interface) error {
	framework.Logf("Testing node deletion when not present in cloud provider")

	// Get a random ready schedulable node to delete
	nodeToDelete, err := e2enode.GetRandomReadySchedulableNode(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get random ready schedulable node: %w", err)
	}

	// Get the original list of ready nodes
	origNodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get ready nodes: %w", err)
	}

	framework.Logf("Original number of ready nodes: %d", len(origNodes.Items))
	framework.Logf("Deleting node %q on the cloud provider", nodeToDelete.Name)

	// Delete the node on the cloud provider
	err = c.DeleteNodeOnCloudProvider(nodeToDelete)
	if err != nil {
		return fmt.Errorf("failed to delete node %q on cloud provider: %w", nodeToDelete.Name, err)
	}

	// Wait for the node count to decrease by 1
	newNodes, err := e2enode.CheckReady(ctx, client, len(origNodes.Items)-1, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to wait for ready nodes to decrease: %w", err)
	}

	if len(newNodes) != len(origNodes.Items)-1 {
		return fmt.Errorf("expected %d nodes, got %d", len(origNodes.Items)-1, len(newNodes))
	}

	// Verify the node is deleted from the API server
	_, err = client.CoreV1().Nodes().Get(ctx, nodeToDelete.Name, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("node %q still exists when it should be deleted", nodeToDelete.Name)
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("unexpected error when getting node %q: %w", nodeToDelete.Name, err)
	}

	framework.Logf("Successfully verified node %q was deleted from API server", nodeToDelete.Name)
	return nil
}

// DeleteNodeOnCloudProvider deletes the specified node from the cloud provider
func (c *CCMNodeTester) DeleteNodeOnCloudProvider(node *v1.Node) error {
	if framework.TestContext.CloudConfig.Provider == nil {
		return fmt.Errorf("cloud provider is not configured")
	}
	return framework.TestContext.CloudConfig.Provider.DeleteNode(node)
}
