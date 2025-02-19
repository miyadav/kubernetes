/*
Copyright 2019 The Kubernetes Authors.

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

package cloud

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// CloudProviderNodeDeleter defines an interface for deleting a node from the cloud provider.
type CloudProviderNodeDeleter interface {
	DeleteNode(ctx context.Context, node *v1.Node) error
}

// deleteNodeOnCloudProvider deletes a node using the cloud provider interface.
func deleteNodeOnCloudProvider(ctx context.Context, node *v1.Node) error {
	provider := framework.TestContext.CloudConfig.Provider
	if provider == nil {
		framework.Failf("Cloud provider is not configured. New providers must implement DeleteNode")
		return fmt.Errorf("cloud provider is not configured")
	}

	// Check if provider supports DeleteNode(ctx, node) (new cloud providers)
	if deleter, ok := provider.(interface {
		DeleteNode(context.Context, *v1.Node) error
	}); ok {
		return deleter.DeleteNode(ctx, node)
	}

	// Fallback for AWS/GKE implementations using DeleteNode(node) (legacy behavior)
	if legacyDeleter, ok := provider.(interface {
		DeleteNode(*v1.Node) error
	}); ok {
		return legacyDeleter.DeleteNode(node)
	}

	framework.Failf("Cloud provider does not support DeleteNode method")
	return fmt.Errorf("cloud provider does not support DeleteNode method")
}

var _ = SIGDescribe(feature.CloudProvider, framework.WithDisruptive(), "Nodes", func() {
	f := framework.NewDefaultFramework("cloudprovider")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged
	var c clientset.Interface

	ginkgo.BeforeEach(func() {
		c = f.ClientSet
	})

	ginkgo.It("should be deleted on API server if it doesn't exist in the cloud provider", func(ctx context.Context) {
		ginkgo.By("deleting a node on the cloud provider")

		nodeToDelete, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		origNodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, c)
		if err != nil {
			framework.Logf("Unexpected error occurred: %v", err)
		}
		framework.ExpectNoErrorWithOffset(0, err)
		framework.Logf("Original number of ready nodes: %d", len(origNodes.Items))

		// Delete the node using the cloud provider-specific implementation
		err = deleteNodeOnCloudProvider(ctx, nodeToDelete)
		if err != nil {
			framework.Failf("failed to delete node %q, err: %q", nodeToDelete.Name, err)
		}

		// Verify the node is removed from Kubernetes API server
		newNodes, err := e2enode.CheckReady(ctx, c, len(origNodes.Items)-1, 5*time.Minute)
		framework.ExpectNoError(err)
		gomega.Expect(newNodes).To(gomega.HaveLen(len(origNodes.Items) - 1))

		_, err = c.CoreV1().Nodes().Get(ctx, nodeToDelete.Name, metav1.GetOptions{})
		if err == nil {
			framework.Failf("node %q still exists when it should be deleted", nodeToDelete.Name)
		} else if !apierrors.IsNotFound(err) {
			framework.Failf("failed to get node %q err: %q", nodeToDelete.Name, err)
		}
	})
})
