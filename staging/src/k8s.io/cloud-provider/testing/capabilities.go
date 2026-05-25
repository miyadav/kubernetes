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

// Capability represents a named test capability of a cloud provider.
// Cloud providers declare which capabilities they support, and e2e tests
// use these declarations to skip tests for unsupported features rather than
// relying on hardcoded provider name checks.
//
// External cloud providers can define custom capabilities using the
// provider-namespaced convention: Capability("mycloud/custom-feature").
type Capability string

const (
	// Core capabilities auto-derivable from cloudprovider.Interface methods.

	CapLoadBalancer Capability = "LoadBalancer"
	CapInstances    Capability = "Instances"
	CapInstancesV2  Capability = "InstancesV2"
	CapZones        Capability = "Zones"
	CapRoutes       Capability = "Routes"
	CapClusters     Capability = "Clusters"

	// Sub-capabilities requiring explicit provider opt-in.

	CapNodeDeletion         Capability = "NodeDeletion"
	CapSSHAccess            Capability = "SSHAccess"
	CapInternalLoadBalancer Capability = "InternalLoadBalancer"
	CapVolumeProvisioning   Capability = "VolumeProvisioning"
	CapNodeResize           Capability = "NodeResize"
	CapTopologyLabels       Capability = "TopologyLabels"
)

// TestCapabilities declares what a cloud provider supports for e2e testing.
// Implementations return true from Has() for supported capabilities and
// false for unsupported ones.
type TestCapabilities interface {
	// Has returns true if the provider supports the given capability.
	Has(cap Capability) bool
	// ProviderName returns the cloud provider name (e.g., "aws", "gce").
	ProviderName() string
}
