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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestInterface provides a cloud-agnostic testing interface for cloud providers
// to implement in their repositories. This enables standardized testing across
// different cloud provider implementations for the cloud-controller-manager.
//
// Based on the cloud-controller-manager functions described at:
// https://kubernetes.io/docs/concepts/architecture/cloud-controller/#functions-of-the-ccm
//
// Cloud providers should implement this interface to return true for capabilities
// they support, along with the corresponding test interface implementation.
type TestInterface interface {
	// NodeController returns the node controller test interface if implemented.
	// The node controller is responsible for:
	// - Initializing nodes with cloud-specific zone/region labels
	// - Initializing nodes with cloud-specific instance details
	// - Verifying node network addresses
	// - Detecting when nodes have been deleted from the cloud
	NodeController() (implemented bool, iface TestNodeControllerInterface)

	// RouteController returns the route controller test interface if implemented.
	// The route controller is responsible for configuring routes in the cloud
	// so that containers on different nodes can communicate with each other.
	RouteController() (implemented bool, iface TestRouteControllerInterface)

	// ServiceController returns the service controller test interface if implemented.
	// The service controller is responsible for provisioning load balancers for
	// services of type LoadBalancer.
	ServiceController() (implemented bool, iface TestServiceControllerInterface)
}

// TestNodeControllerInterface provides testing operations for the node controller.
// This interface allows testing node lifecycle management, metadata, and addressing.
type TestNodeControllerInterface interface {
	// NodeExists checks if a node exists in the cloud provider.
	NodeExists(ctx context.Context, node *v1.Node) (bool, error)

	// NodeShutdown checks if a node is shutdown in the cloud provider.
	NodeShutdown(ctx context.Context, node *v1.Node) (bool, error)

	// NodeMetadata returns metadata about the node from the cloud provider.
	// This includes provider ID, instance type, addresses, and topology information.
	NodeMetadata(ctx context.Context, node *v1.Node) (*NodeMetadata, error)

	// DeleteNode deletes the node from the cloud provider.
	// This is used to test that the node controller properly removes nodes
	// from Kubernetes when they no longer exist in the cloud.
	// Note: This should be compatible with framework.ProviderInterface.DeleteNode
	DeleteNode(ctx context.Context, node *v1.Node) error
}

// TestRouteControllerInterface provides testing operations for the route controller.
// This interface allows testing route creation, deletion, and listing.
type TestRouteControllerInterface interface {
	// ListRoutes lists all managed routes for the cluster.
	ListRoutes(ctx context.Context, clusterName string) ([]*Route, error)

	// CreateRoute creates a new route in the cloud provider.
	CreateRoute(ctx context.Context, clusterName string, nameHint string, route *Route) error

	// DeleteRoute deletes a route from the cloud provider.
	DeleteRoute(ctx context.Context, clusterName string, route *Route) error

	// GetClusterName returns the cluster name to use for route operations.
	GetClusterName() string
}

// TestServiceControllerInterface provides testing operations for the service controller.
// This interface allows testing load balancer provisioning, updates, and deletion.
type TestServiceControllerInterface interface {
	// GetLoadBalancer checks if a load balancer exists for the service and returns its status.
	GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*LoadBalancerStatus, bool, error)

	// EnsureLoadBalancer ensures a load balancer exists for the service.
	// It creates a new load balancer if one doesn't exist, or updates the existing one.
	EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*LoadBalancerStatus, error)

	// UpdateLoadBalancer updates the hosts under the specified load balancer.
	UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error

	// EnsureLoadBalancerDeleted ensures the load balancer for the service is deleted.
	EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error

	// GetClusterName returns the cluster name to use for load balancer operations.
	GetClusterName() string
}

// NodeMetadata contains metadata about a node from the cloud provider.
// This corresponds to the InstanceMetadata returned by the InstancesV2 interface.
type NodeMetadata struct {
	// ProviderID is the cloud provider's unique identifier for the node.
	// Format: <provider-name>://<instance-id>
	ProviderID string

	// InstanceType is the instance type/size (e.g., "t3.medium", "n1-standard-1").
	InstanceType string

	// NodeAddresses contains the node's addresses (internal IP, external IP, hostname).
	NodeAddresses []v1.NodeAddress

	// Zone is the availability zone the node is in.
	Zone string

	// Region is the region the node is in.
	Region string

	// AdditionalLabels contains additional cloud-specific labels to apply to the node.
	AdditionalLabels map[string]string
}

// Route represents a network route in the cloud provider.
// This corresponds to the Route struct in the cloudprovider.Routes interface.
type Route struct {
	// Name is the name of the route in the cloud provider.
	Name string

	// TargetNode is the node name that is the target of this route.
	TargetNode types.NodeName

	// TargetNodeAddresses are the IP addresses of the target node.
	TargetNodeAddresses []v1.NodeAddress

	// DestinationCIDR is the CIDR block this route applies to.
	DestinationCIDR string

	// Blackhole indicates if this is a blackhole route.
	Blackhole bool
}

// LoadBalancerStatus contains the status of a load balancer.
// This corresponds to the v1.LoadBalancerStatus that the LoadBalancer interface returns.
type LoadBalancerStatus struct {
	// Ingress contains the list of ingress points for the load balancer.
	Ingress []LoadBalancerIngress
}

// LoadBalancerIngress represents an ingress point for a load balancer.
type LoadBalancerIngress struct {
	// IP is the IP address of the load balancer ingress point.
	IP string

	// Hostname is the hostname of the load balancer ingress point.
	Hostname string

	// Ports specifies the port configurations for the load balancer.
	Ports []PortStatus
}

// PortStatus represents the status of a load balancer port.
type PortStatus struct {
	// Port is the port number.
	Port int32

	// Protocol is the protocol (TCP, UDP, etc.).
	Protocol v1.Protocol
}
