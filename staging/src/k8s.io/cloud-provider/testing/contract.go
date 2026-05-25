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

package testing

import (
	"fmt"
	"strings"

	cloudprovider "k8s.io/cloud-provider"
)

// ContractT is the subset of testing.TB needed by ValidateCapabilities.
// Both *testing.T and *testing.B satisfy this interface.
type ContractT interface {
	Helper()
	Errorf(format string, args ...interface{})
}

// coreCapMapping maps core capabilities to the cloudprovider.Interface method
// names that report them. Used by ValidateCapabilities to produce clear messages.
var coreCapMapping = []struct {
	cap        Capability
	methodName string
	check      func(cloudprovider.Interface) bool
}{
	{CapLoadBalancer, "LoadBalancer()", func(c cloudprovider.Interface) bool { _, ok := c.LoadBalancer(); return ok }},
	{CapInstances, "Instances()", func(c cloudprovider.Interface) bool { _, ok := c.Instances(); return ok }},
	{CapInstancesV2, "InstancesV2()", func(c cloudprovider.Interface) bool { _, ok := c.InstancesV2(); return ok }},
	{CapZones, "Zones()", func(c cloudprovider.Interface) bool { _, ok := c.Zones(); return ok }},
	{CapRoutes, "Routes()", func(c cloudprovider.Interface) bool { _, ok := c.Routes(); return ok }},
	{CapClusters, "Clusters()", func(c cloudprovider.Interface) bool { _, ok := c.Clusters(); return ok }},
}

// ValidateCapabilities checks that declared TestCapabilities are consistent
// with the actual cloudprovider.Interface implementation.
//
// For core capabilities (LoadBalancer, Instances, InstancesV2, Zones, Routes,
// Clusters), it calls the corresponding method on the cloud provider and
// compares the bool return with what the capabilities declare.
//
// It catches two classes of errors:
//   - Over-declaration: capabilities say supported, but cloud.Interface says no
//   - Under-declaration: cloud.Interface says supported, but capabilities say no
//
// Cloud providers should call this in their own test suites:
//
//	func TestCapabilitiesContract(t *testing.T) {
//	    myCloud := newMyCloud(cfg)
//	    caps := cloudprovidertesting.DeriveFromCloud(myCloud)
//	    caps.Caps[cloudprovidertesting.CapNodeDeletion] = true
//	    cloudprovidertesting.ValidateCapabilities(t, myCloud, caps)
//	}
func ValidateCapabilities(t ContractT, cloud cloudprovider.Interface, declared TestCapabilities) {
	t.Helper()

	if cloud.ProviderName() != declared.ProviderName() {
		t.Errorf("provider name mismatch: cloud.ProviderName() = %q, declared.ProviderName() = %q",
			cloud.ProviderName(), declared.ProviderName())
	}

	var errors []string
	for _, m := range coreCapMapping {
		cloudSays := m.check(cloud)
		declaredSays := declared.Has(m.cap)

		if declaredSays && !cloudSays {
			errors = append(errors, fmt.Sprintf(
				"over-declared: capability %q is declared as supported, but %s returns false",
				m.cap, m.methodName))
		}
		if cloudSays && !declaredSays {
			errors = append(errors, fmt.Sprintf(
				"under-declared: %s returns true, but capability %q is declared as unsupported",
				m.methodName, m.cap))
		}
	}

	if len(errors) > 0 {
		t.Errorf("capability contract violations for provider %q:\n  %s",
			declared.ProviderName(), strings.Join(errors, "\n  "))
	}
}

