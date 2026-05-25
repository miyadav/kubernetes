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
	cloudprovider "k8s.io/cloud-provider"
)

// MapCapabilities is a map-based implementation of TestCapabilities.
// Providers can construct one directly or use DeriveFromCloud as a
// starting point and then set sub-capabilities.
type MapCapabilities struct {
	Name string
	Caps map[Capability]bool
}

// Has returns true if the capability is present and set to true.
func (m *MapCapabilities) Has(cap Capability) bool {
	return m.Caps[cap]
}

// ProviderName returns the cloud provider name.
func (m *MapCapabilities) ProviderName() string {
	return m.Name
}

// DeriveFromCloud introspects a cloudprovider.Interface and returns a
// MapCapabilities with the core capabilities set according to whether
// each sub-interface is reported as implemented.
//
// Providers should call this as a starting point and then override
// sub-capabilities as needed:
//
//	caps := cloudprovidertesting.DeriveFromCloud(myCloud)
//	caps.Caps[cloudprovidertesting.CapSSHAccess] = true
//	caps.Caps[cloudprovidertesting.CapNodeDeletion] = true
func DeriveFromCloud(cloud cloudprovider.Interface) *MapCapabilities {
	caps := &MapCapabilities{
		Name: cloud.ProviderName(),
		Caps: make(map[Capability]bool),
	}

	if _, ok := cloud.LoadBalancer(); ok {
		caps.Caps[CapLoadBalancer] = true
	}
	if _, ok := cloud.Instances(); ok {
		caps.Caps[CapInstances] = true
	}
	if _, ok := cloud.InstancesV2(); ok {
		caps.Caps[CapInstancesV2] = true
	}
	if _, ok := cloud.Zones(); ok {
		caps.Caps[CapZones] = true
	}
	if _, ok := cloud.Routes(); ok {
		caps.Caps[CapRoutes] = true
	}
	if _, ok := cloud.Clusters(); ok {
		caps.Caps[CapClusters] = true
	}

	return caps
}

// Verify MapCapabilities implements TestCapabilities.
var _ TestCapabilities = (*MapCapabilities)(nil)
