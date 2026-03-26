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
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = SIGDescribe(feature.CloudProvider, "Route Controller", func() {
	f := framework.NewDefaultFramework("route-controller")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	var (
		c               clientset.Interface
		testDriver      TestInterface
		routeController TestRouteControllerInterface
		implemented     bool
		clusterName     string
	)

	ginkgo.BeforeEach(func(ctx context.Context) {
		c = f.ClientSet

		// Get the test driver from the cloud provider
		testDriver = GetTestDriver()
		if testDriver == nil {
			e2eskipper.Skipf("External cloud provider test driver not configured")
		}

		// Check if route controller is implemented
		implemented, routeController = testDriver.RouteController()
		if !implemented {
			e2eskipper.Skipf("Route controller not implemented by cloud provider")
		}

		clusterName = routeController.GetClusterName()
	})

	ginkgo.It("should be able to list routes", func(ctx context.Context) {
		ginkgo.By("Listing routes from the cloud provider")
		routes, err := routeController.ListRoutes(ctx, clusterName)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying routes are returned")
		gomega.Expect(routes).ToNot(gomega.BeNil())
		framework.Logf("Found %d routes in cluster %s", len(routes), clusterName)
	})

	ginkgo.It("should be able to create and delete a route", func(ctx context.Context) {
		ginkgo.By("Getting a node to use as the route target")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Creating a test route")
		testRoute := &Route{
			TargetNode:      types.NodeName(node.Name),
			DestinationCIDR: "10.240.1.0/24",
		}
		nameHint := fmt.Sprintf("test-route-%s", node.Name)
		err = routeController.CreateRoute(ctx, clusterName, nameHint, testRoute)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the route appears in the route list")
		routes, err := routeController.ListRoutes(ctx, clusterName)
		framework.ExpectNoError(err)
		found := false
		var createdRoute *Route
		for _, route := range routes {
			if route.DestinationCIDR == testRoute.DestinationCIDR {
				found = true
				createdRoute = route
				break
			}
		}
		gomega.Expect(found).To(gomega.BeTrue(), "Created route should appear in route list")

		ginkgo.By("Deleting the test route")
		err = routeController.DeleteRoute(ctx, clusterName, createdRoute)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the route is removed from the route list")
		routes, err = routeController.ListRoutes(ctx, clusterName)
		framework.ExpectNoError(err)
		found = false
		for _, route := range routes {
			if route.DestinationCIDR == testRoute.DestinationCIDR {
				found = true
				break
			}
		}
		gomega.Expect(found).To(gomega.BeFalse(), "Deleted route should not appear in route list")
	})

	ginkgo.It("should create routes with correct target node information", func(ctx context.Context) {
		ginkgo.By("Getting a node to use as the route target")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Creating a test route with node addresses")
		testRoute := &Route{
			TargetNode:          types.NodeName(node.Name),
			TargetNodeAddresses: node.Status.Addresses,
			DestinationCIDR:     "10.240.2.0/24",
		}
		nameHint := fmt.Sprintf("test-route-addr-%s", node.Name)
		err = routeController.CreateRoute(ctx, clusterName, nameHint, testRoute)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the route was created")
		routes, err := routeController.ListRoutes(ctx, clusterName)
		framework.ExpectNoError(err)
		found := false
		var createdRoute *Route
		for _, route := range routes {
			if route.DestinationCIDR == testRoute.DestinationCIDR {
				found = true
				createdRoute = route
				framework.Logf("Created route: %+v", route)
				break
			}
		}
		gomega.Expect(found).To(gomega.BeTrue(), "Created route should appear in route list")

		ginkgo.By("Cleaning up the test route")
		err = routeController.DeleteRoute(ctx, clusterName, createdRoute)
		framework.ExpectNoError(err)
	})

	ginkgo.It("should handle blackhole routes", func(ctx context.Context) {
		ginkgo.By("Getting a node for the blackhole route")
		node, err := e2enode.GetRandomReadySchedulableNode(ctx, c)
		framework.ExpectNoError(err)

		ginkgo.By("Creating a blackhole route")
		testRoute := &Route{
			TargetNode:      types.NodeName(node.Name),
			DestinationCIDR: "10.240.3.0/24",
			Blackhole:       true,
		}
		nameHint := fmt.Sprintf("test-route-blackhole-%s", node.Name)
		err = routeController.CreateRoute(ctx, clusterName, nameHint, testRoute)
		framework.ExpectNoError(err)

		ginkgo.By("Verifying the blackhole route was created")
		routes, err := routeController.ListRoutes(ctx, clusterName)
		framework.ExpectNoError(err)
		found := false
		var createdRoute *Route
		for _, route := range routes {
			if route.DestinationCIDR == testRoute.DestinationCIDR {
				found = true
				createdRoute = route
				gomega.Expect(route.Blackhole).To(gomega.BeTrue(), "Route should be marked as blackhole")
				break
			}
		}
		gomega.Expect(found).To(gomega.BeTrue(), "Created blackhole route should appear in route list")

		ginkgo.By("Cleaning up the blackhole route")
		err = routeController.DeleteRoute(ctx, clusterName, createdRoute)
		framework.ExpectNoError(err)
	})
})
