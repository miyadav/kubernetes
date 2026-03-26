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
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = SIGDescribe(feature.CloudProvider, framework.WithDisruptive(), "Node Controller", func() {
	f := framework.NewDefaultFramework("node-controller")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	var (
		c              clientset.Interface
		testDriver     TestInterface
		nodeController TestNodeControllerInterface
		implemented    bool
	)

	ginkgo.BeforeEach(func(ctx context.Context) {
		c = f.ClientSet

		// Get the test driver from the cloud provider
		testDriver = GetTestDriver()
		if testDriver == nil {
			e2eskipper.Skipf("External cloud provider test driver not configured")
		}

		// Check if node controller is implemented
		implemented, nodeController = testDriver.NodeController()
		if !implemented {
			e2eskipper.Skipf("Node controller not implemented by cloud provider")
		}
	})

	ginkgo.It("should correctly report that a node exists", func(ctx context.Context) {
		ginkgo.By("Getting a ready node")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the node exists in the cloud provider")
		exists, err := nodeController.NodeExists(ctx, node)
		framework.ExpectNoError(err)
		gomega.Expect(exists).To(gomega.BeTrue(), "Node should exist in cloud provider")
	})

	ginkgo.It("should correctly report node metadata", func(ctx context.Context) {
		ginkgo.By("Getting a ready node")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Getting node metadata from cloud provider")
		metadata, err := nodeController.NodeMetadata(ctx, node)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying metadata is populated")
		gomega.Expect(metadata).ToNot(gomega.BeNil())
		gomega.Expect(metadata.ProviderID).ToNot(gomega.BeEmpty(), "Provider ID should be set")
		gomega.Expect(metadata.InstanceType).ToNot(gomega.BeEmpty(), "Instance type should be set")
		gomega.Expect(metadata.NodeAddresses).ToNot(gomega.BeEmpty(), "Node addresses should be set")

		ginkgo.By("Verifying topology information")
		if node.Labels != nil {
			if _, hasZone := node.Labels[v1.LabelTopologyZone]; hasZone {
				gomega.Expect(metadata.Zone).ToNot(gomega.BeEmpty(), "Zone should be set when node has zone label")
			}
			if _, hasRegion := node.Labels[v1.LabelTopologyRegion]; hasRegion {
				gomega.Expect(metadata.Region).ToNot(gomega.BeEmpty(), "Region should be set when node has region label")
			}
		}
	})

	ginkgo.It("should be deleted on API server if it doesn't exist in the cloud provider", func(ctx context.Context) {
		ginkgo.By("Getting a ready node to delete")
		nodeToDelete, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Recording the original number of ready nodes")
		origNodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, c)
		framework.ExpectNoError(err)
		framework.Logf("Original number of ready nodes: %d", len(origNodes.Items))

		ginkgo.By("Deleting the node from the cloud provider")
		err = nodeController.DeleteNode(ctx, nodeToDelete)
		if err != nil {
			framework.Failf("failed to delete node %q, err: %q", nodeToDelete.Name, err)
		}

		ginkgo.By("Waiting for the node count to decrease")
		newNodes, err := e2enode.CheckReady(ctx, c, len(origNodes.Items)-1, 5*time.Minute)
		framework.ExpectNoError(err)
		gomega.Expect(newNodes).To(gomega.HaveLen(len(origNodes.Items) - 1))

		ginkgo.By("Verifying the node is deleted from the API server")
		_, err = c.CoreV1().Nodes().Get(ctx, nodeToDelete.Name, metav1.GetOptions{})
		if err == nil {
			framework.Failf("node %q still exists when it should be deleted", nodeToDelete.Name)
		} else if !apierrors.IsNotFound(err) {
			framework.Failf("failed to get node %q err: %q", nodeToDelete.Name, err)
		}
	})

	ginkgo.It("should correctly detect node shutdown state", func(ctx context.Context) {
		ginkgo.By("Getting a ready node")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Checking if the node is shutdown")
		isShutdown, err := nodeController.NodeShutdown(ctx, node)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying a running node is not reported as shutdown")
		gomega.Expect(isShutdown).To(gomega.BeFalse(), "A running node should not be shutdown")
	})

	ginkgo.It("should report node as non-existent after deletion", func(ctx context.Context) {
		ginkgo.By("Getting a ready node to delete")
		nodeToDelete, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the node exists before deletion")
		exists, err := nodeController.NodeExists(ctx, nodeToDelete)
		framework.ExpectNoError(err)
		gomega.Expect(exists).To(gomega.BeTrue(), "Node should exist before deletion")

		ginkgo.By("Deleting the node from the cloud provider")
		err = nodeController.DeleteNode(ctx, nodeToDelete)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the node no longer exists in the cloud provider")
		// Note: There might be a delay before the cloud provider reports the node as deleted
		gomega.Eventually(ctx, func(ctx context.Context) bool {
			exists, err := nodeController.NodeExists(ctx, nodeToDelete)
			if err != nil {
				framework.Logf("Error checking node existence: %v", err)
				return false
			}
			return !exists
		}, 2*time.Minute, 5*time.Second).Should(gomega.BeTrue(), "Node should eventually not exist in cloud provider")
	})
})
