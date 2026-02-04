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

type TestNodeLifecycleInterface interface {
	Exists() bool
	IsShutdown() bool
	Details() NodeDetails
	Addresses() []NodeAddress
}

type TestLoadBalancerInterface interface {
	Create() LoadBalancer
	Update() LoadBalancer
	Get() LoadBalancer
	Delete() LoadBalancer
}

// TestInterface provides a cloud-agnostic testing interface for cloud providers
// to implement in their repositories. This enables standardized testing across
// different cloud provider implementations.
type TestInterface interface {

	// ---------- Node / VM ----------
	NodeLifecycle() (implemented bool, iface TestNodeLifecycleInterface)

	// ---------- Load Balancer ----------
	LoadBalancer() (implemented bool, iface TestLoadBalancerInterface)

	// ---------- Networking ----------

	// Can the cloud list routes?
	ListRoutes() (supported bool, routes []Route)

	// Can the cloud create a route?
	CreateRoute() (supported bool, route Route)

	// Can the cloud delete a route?
	DeleteRoute() (supported bool, deleted bool)

	// ---------- Topology ----------

	// Can the cloud provide zone/region info?
	GetZone() (supported bool, zone Zone)

	// ---------- Cluster ----------

	// Can the cloud list clusters?
	ListClusters() (supported bool, clusters []string)

	// Can the cloud return a cluster endpoint?
	GetClusterEndpoint() (supported bool, endpoint string)
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
