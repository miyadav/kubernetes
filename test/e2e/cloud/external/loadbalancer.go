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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = SIGDescribe(feature.CloudProvider, "Service Controller", func() {
	f := framework.NewDefaultFramework("service-controller")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	var (
		c                  clientset.Interface
		testDriver         TestInterface
		serviceController  TestServiceControllerInterface
		implemented        bool
		clusterName        string
		defaultServiceName = "test-loadbalancer"
	)

	ginkgo.BeforeEach(func(ctx context.Context) {
		c = f.ClientSet

		// Get the test driver from the cloud provider
		testDriver = GetTestDriver()
		if testDriver == nil {
			e2eskipper.Skipf("External cloud provider test driver not configured")
		}

		// Check if service controller is implemented
		implemented, serviceController = testDriver.ServiceController()
		if !implemented {
			e2eskipper.Skipf("Service controller not implemented by cloud provider")
		}

		clusterName = serviceController.GetClusterName()
	})

	ginkgo.It("should create a load balancer for a service", func(ctx context.Context) {
		serviceName := defaultServiceName
		ns := f.Namespace.Name

		ginkgo.By("Creating a LoadBalancer service")
		service := createLoadBalancerService(ctx, c, ns, serviceName)

		ginkgo.By("Getting nodes to use as backends")
		nodes, err := e2enode.GetReadySchedulableNodes(ctx, c)
		framework.ExpectNoError(err)
		gomega.Expect(nodes.Items).ToNot(gomega.BeEmpty(), "Need at least one node")

		var nodePointers []*v1.Node
		for i := range nodes.Items {
			nodePointers = append(nodePointers, &nodes.Items[i])
		}

		ginkgo.By("Ensuring the load balancer is created")
		lbStatus, err := serviceController.EnsureLoadBalancer(ctx, clusterName, service, nodePointers)
		framework.ExpectNoError(err)
		gomega.Expect(lbStatus).ToNot(gomega.BeNil())
		gomega.Expect(lbStatus.Ingress).ToNot(gomega.BeEmpty(), "Load balancer should have at least one ingress point")

		framework.Logf("Load balancer created with ingress: %+v", lbStatus.Ingress)

		ginkgo.By("Verifying the load balancer exists")
		status, exists, err := serviceController.GetLoadBalancer(ctx, clusterName, service)
		framework.ExpectNoError(err)
		gomega.Expect(exists).To(gomega.BeTrue(), "Load balancer should exist")
		gomega.Expect(status).ToNot(gomega.BeNil())

		ginkgo.By("Cleaning up the load balancer")
		defer func() {
			err := serviceController.EnsureLoadBalancerDeleted(ctx, clusterName, service)
			if err != nil {
				framework.Logf("Failed to cleanup load balancer: %v", err)
			}
		}()
	})

	ginkgo.It("should update a load balancer when nodes change", func(ctx context.Context) {
		serviceName := fmt.Sprintf("%s-update", defaultServiceName)
		ns := f.Namespace.Name

		ginkgo.By("Creating a LoadBalancer service")
		service := createLoadBalancerService(ctx, c, ns, serviceName)

		ginkgo.By("Getting all nodes")
		allNodes, err := e2enode.GetReadySchedulableNodes(ctx, c)
		framework.ExpectNoError(err)
		gomega.Expect(len(allNodes.Items)).To(gomega.BeNumerically(">=", 1), "Need at least one node")

		var allNodePointers []*v1.Node
		for i := range allNodes.Items {
			allNodePointers = append(allNodePointers, &allNodes.Items[i])
		}

		ginkgo.By("Creating load balancer with all nodes")
		_, err = serviceController.EnsureLoadBalancer(ctx, clusterName, service, allNodePointers)
		framework.ExpectNoError(err)

		ginkgo.By("Updating load balancer with subset of nodes")
		// Use only the first node
		err = serviceController.UpdateLoadBalancer(ctx, clusterName, service, allNodePointers[:1])
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the load balancer still exists after update")
		_, exists, err := serviceController.GetLoadBalancer(ctx, clusterName, service)
		framework.ExpectNoError(err)
		gomega.Expect(exists).To(gomega.BeTrue(), "Load balancer should still exist after update")

		ginkgo.By("Cleaning up the load balancer")
		defer func() {
			err := serviceController.EnsureLoadBalancerDeleted(ctx, clusterName, service)
			if err != nil {
				framework.Logf("Failed to cleanup load balancer: %v", err)
			}
		}()
	})

	ginkgo.It("should delete a load balancer when requested", func(ctx context.Context) {
		serviceName := fmt.Sprintf("%s-delete", defaultServiceName)
		ns := f.Namespace.Name

		ginkgo.By("Creating a LoadBalancer service")
		service := createLoadBalancerService(ctx, c, ns, serviceName)

		ginkgo.By("Getting nodes")
		nodes, err := e2enode.GetReadySchedulableNodes(ctx, c)
		framework.ExpectNoError(err)

		var nodePointers []*v1.Node
		for i := range nodes.Items {
			nodePointers = append(nodePointers, &nodes.Items[i])
		}

		ginkgo.By("Creating the load balancer")
		_, err = serviceController.EnsureLoadBalancer(ctx, clusterName, service, nodePointers)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the load balancer exists")
		_, exists, err := serviceController.GetLoadBalancer(ctx, clusterName, service)
		framework.ExpectNoError(err)
		gomega.Expect(exists).To(gomega.BeTrue(), "Load balancer should exist")

		ginkgo.By("Deleting the load balancer")
		err = serviceController.EnsureLoadBalancerDeleted(ctx, clusterName, service)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the load balancer is deleted")
		err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
			_, exists, err := serviceController.GetLoadBalancer(ctx, clusterName, service)
			if err != nil {
				return false, err
			}
			return !exists, nil
		})
		framework.ExpectNoError(err, "Load balancer should be deleted")
	})

	ginkgo.It("should handle multiple ports on a load balancer", func(ctx context.Context) {
		serviceName := fmt.Sprintf("%s-multiport", defaultServiceName)
		ns := f.Namespace.Name

		ginkgo.By("Creating a LoadBalancer service with multiple ports")
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: ns,
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt32(8080),
						Protocol:   v1.ProtocolTCP,
					},
					{
						Name:       "https",
						Port:       443,
						TargetPort: intstr.FromInt32(8443),
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"app": "test",
				},
			},
		}
		service, err := c.CoreV1().Services(ns).Create(ctx, service, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		ginkgo.By("Getting nodes")
		nodes, err := e2enode.GetReadySchedulableNodes(ctx, c)
		framework.ExpectNoError(err)

		var nodePointers []*v1.Node
		for i := range nodes.Items {
			nodePointers = append(nodePointers, &nodes.Items[i])
		}

		ginkgo.By("Creating the load balancer")
		lbStatus, err := serviceController.EnsureLoadBalancer(ctx, clusterName, service, nodePointers)
		framework.ExpectNoError(err)
		gomega.Expect(lbStatus).ToNot(gomega.BeNil())
		gomega.Expect(lbStatus.Ingress).ToNot(gomega.BeEmpty())

		ginkgo.By("Cleaning up the load balancer")
		defer func() {
			err := serviceController.EnsureLoadBalancerDeleted(ctx, clusterName, service)
			if err != nil {
				framework.Logf("Failed to cleanup load balancer: %v", err)
			}
		}()
	})
})

// createLoadBalancerService creates a basic LoadBalancer type service
func createLoadBalancerService(ctx context.Context, c clientset.Interface, ns, name string) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt32(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test",
			},
		},
	}

	service, err := c.CoreV1().Services(ns).Create(ctx, service, metav1.CreateOptions{})
	framework.ExpectNoError(err)
	return service
}
