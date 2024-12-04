package external

import (
	"context"
)

// CloudProviderInterface defines the necessary methods that a cloud provider's CCM must implement
type CloudProviderInterface interface {
	// Initialize is called to set up the provider for testing
	Initialize(ctx context.Context) error

	// DeleteNode deletes a node from the cloud provider
	DeleteNode(nodeName string) error

	// CreateCluster creates a cluster in the cloud provider
	CreateCluster(clusterName string) error

	// DeleteCluster deletes a cluster from the cloud provider
	DeleteCluster(clusterName string) error

	// GetInstances gets the instances from the cloud provider
	GetInstances() ([]string, error)

	// GetZones gets the available zones
	GetZones() ([]string, error)

	// GetRegions gets the available regions
	GetRegions() ([]string, error)
}
