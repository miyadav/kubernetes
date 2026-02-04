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

// TestInterface provides a cloud-agnostic testing interface for cloud providers
// to implement in their repositories. This enables standardized testing across
// different cloud provider implementations.
type TestInterface interface {

	// ---------- Node / VM ----------
	NodeLifecycle() (implemented bool, iface TestNodeLifecycleInterface)

	// ---------- Load Balancer ----------
	LoadBalancer() (implemented bool, iface TestLoadBalancerInterface)

	// ---------- Networking ----------
	Routes() (implemented bool, iface TestRoutesInterface)

	// ---------- Topology ----------
	Topology() (implemented bool, iface TestTopologyInterface)

	// ---------- Cluster ----------
	Cluster() (implemented bool, iface TestClusterInterface)
}

// TestNodeLifecycleInterface provides node/VM lifecycle testing operations
type TestNodeLifecycleInterface interface {
	// Exists checks if the node exists in the cloud
	Exists() bool

	// IsShutdown checks if the node is shutdown
	IsShutdown() bool

	// Details returns node information
	Details() NodeDetails

	// Addresses returns node addresses
	Addresses() []NodeAddress
}

// TestLoadBalancerInterface provides load balancer testing operations
type TestLoadBalancerInterface interface {
	// Create creates a load balancer
	Create() LoadBalancer

	// Update updates a load balancer
	Update() LoadBalancer

	// Get fetches load balancer state
	Get() LoadBalancer

	// Delete deletes a load balancer
	Delete() bool
}

// TestRoutesInterface provides network routing testing operations
type TestRoutesInterface interface {
	// List lists routes
	List() []Route

	// Create creates a route
	Create() Route

	// Delete deletes a route
	Delete() bool
}

// TestTopologyInterface provides topology testing operations
type TestTopologyInterface interface {
	// GetZone provides zone/region info
	GetZone() Zone
}

// TestClusterInterface provides cluster-level testing operations
type TestClusterInterface interface {
	// List lists clusters
	List() []string

	// GetEndpoint returns a cluster endpoint
	GetEndpoint() string
}

// NodeDetails contains information about a node/VM
type NodeDetails struct {
	// InstanceID is the cloud provider's unique identifier for the instance
	InstanceID string

	// InstanceType is the machine type/size
	InstanceType string

	// ProviderID is the full provider ID (e.g., "aws:///us-east-1a/i-1234567890abcdef0")
	ProviderID string

	// Metadata contains additional cloud-specific node information
	Metadata map[string]string
}

// NodeAddress represents an address associated with a node
type NodeAddress struct {
	// Type is the type of address (e.g., InternalIP, ExternalIP, Hostname)
	Type string

	// Address is the actual address value
	Address string
}

// LoadBalancer contains information about a load balancer
type LoadBalancer struct {
	// Name is the load balancer name
	Name string

	// IPAddress is the load balancer's IP address
	IPAddress string

	// Hostname is the load balancer's hostname/DNS name
	Hostname string

	// Ports are the ports the load balancer listens on
	Ports []int32

	// Metadata contains additional cloud-specific load balancer information
	Metadata map[string]string
}

// Route represents a network route
type Route struct {
	// Name is the route name
	Name string

	// DestinationCIDR is the destination CIDR block
	DestinationCIDR string

	// TargetNode is the target node name
	TargetNode string

	// Metadata contains additional cloud-specific route information
	Metadata map[string]string
}

// Zone contains zone and region information
type Zone struct {
	// Region is the cloud region
	Region string

	// Zone is the availability zone
	Zone string

	// FailureDomain is the failure domain identifier
	FailureDomain string
}
