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
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMTester implements the main Tester interface for Cloud Controller Manager tests.
// It aggregates all test interfaces and provides a unified interface for cloud provider testing.
// This implementation follows the same pattern as cloudprovider.Interface, where each
// method returns (InterfaceType, bool) to indicate feature support.
type CCMTester struct {
	// Cloud provider interface - accessed through framework.TestContext.CloudConfig.Provider
	// Individual test implementations can access the cloud provider through this
	loadBalancerTester LoadBalancerTester
	instancesTester    InstancesTester
	instancesV2Tester  InstancesV2Tester
	routesTester       RoutesTester
	zonesTester        ZonesTester
	clustersTester     ClustersTester
	nodeTester         NodeTester
}

// NewCCMTester creates a new CCMTester instance.
// This initializes all test interfaces that are supported by the cloud provider.
func NewCCMTester() Tester {
	tester := &CCMTester{}

	// Initialize test interfaces based on cloud provider capabilities
	// The cloud provider interface is accessed through framework.TestContext.CloudConfig.Provider
	// We check which interfaces are supported and initialize the corresponding testers

	// Check if cloud provider is configured
	if framework.TestContext.CloudConfig.Provider == nil {
		return tester
	}

	// Note: The actual cloud provider interface (cloudprovider.Interface) is not directly
	// accessible through the test framework. Cloud providers implementing this tester
	// should set the individual testers using the Set methods below, or implement
	// their own logic to determine which interfaces are supported.

	return tester
}

// SetLoadBalancerTester sets the load balancer tester implementation.
func (c *CCMTester) SetLoadBalancerTester(tester LoadBalancerTester) {
	c.loadBalancerTester = tester
}

// SetInstancesTester sets the instances tester implementation.
func (c *CCMTester) SetInstancesTester(tester InstancesTester) {
	c.instancesTester = tester
}

// SetInstancesV2Tester sets the InstancesV2 tester implementation.
func (c *CCMTester) SetInstancesV2Tester(tester InstancesV2Tester) {
	c.instancesV2Tester = tester
}

// SetRoutesTester sets the routes tester implementation.
func (c *CCMTester) SetRoutesTester(tester RoutesTester) {
	c.routesTester = tester
}

// SetZonesTester sets the zones tester implementation.
func (c *CCMTester) SetZonesTester(tester ZonesTester) {
	c.zonesTester = tester
}

// SetClustersTester sets the clusters tester implementation.
func (c *CCMTester) SetClustersTester(tester ClustersTester) {
	c.clustersTester = tester
}

// SetNodeTester sets the node tester implementation.
func (c *CCMTester) SetNodeTester(tester NodeTester) {
	c.nodeTester = tester
}

// LoadBalancerTester returns the load balancer test interface.
// The boolean return indicates whether load balancer testing is supported.
func (c *CCMTester) LoadBalancerTester() (LoadBalancerTester, bool) {
	return c.loadBalancerTester, c.loadBalancerTester != nil
}

// InstancesTester returns the instances test interface.
// The boolean return indicates whether instances testing is supported.
func (c *CCMTester) InstancesTester() (InstancesTester, bool) {
	return c.instancesTester, c.instancesTester != nil
}

// InstancesV2Tester returns the InstancesV2 test interface.
// The boolean return indicates whether InstancesV2 testing is supported.
func (c *CCMTester) InstancesV2Tester() (InstancesV2Tester, bool) {
	return c.instancesV2Tester, c.instancesV2Tester != nil
}

// RoutesTester returns the routes test interface.
// The boolean return indicates whether routes testing is supported.
func (c *CCMTester) RoutesTester() (RoutesTester, bool) {
	return c.routesTester, c.routesTester != nil
}

// ZonesTester returns the zones test interface.
// The boolean return indicates whether zones testing is supported.
func (c *CCMTester) ZonesTester() (ZonesTester, bool) {
	return c.zonesTester, c.zonesTester != nil
}

// ClustersTester returns the clusters test interface.
// The boolean return indicates whether clusters testing is supported.
func (c *CCMTester) ClustersTester() (ClustersTester, bool) {
	return c.clustersTester, c.clustersTester != nil
}

// NodeTester returns the node test interface.
// The boolean return indicates whether node testing is supported.
func (c *CCMTester) NodeTester() (NodeTester, bool) {
	return c.nodeTester, c.nodeTester != nil
}

// ProviderName returns the cloud provider identifier.
// This should match the value returned by cloudprovider.Interface.ProviderName().
func (c *CCMTester) ProviderName() string {
	if framework.TestContext.CloudConfig.Provider == nil {
		return "unknown"
	}
	// The actual provider name should be determined from the cloud provider interface
	// For now, we use the framework's provider name
	return framework.TestContext.Provider
}

