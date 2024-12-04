package mockprovider

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider"
)

// MockProvider is a mock implementation of the cloudprovider interface.
type MockProvider struct{}

// NewMockProvider creates a new instance of MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// Initialize is called during the initialization of the cloud provider.
func (mp *MockProvider) Initialize(ctx context.Context, controllerManagerName string, cloudConfig []byte) error {
	// Initialization logic for mock provider
	fmt.Println("MockProvider initialized.")
	return nil
}

// GetCloudProviderName returns the name of the cloud provider.
func (mp *MockProvider) GetCloudProviderName() string {
	return "MockProvider"
}

// GetInstances returns a mock instance list.
func (mp *MockProvider) GetInstances() cloudprovider.Instances {
	return nil
}

// GetZones returns a mock zone list.
func (mp *MockProvider) GetZones() cloudprovider.Zones {
	return nil
}

// GetRegions returns a mock region list.
func (mp *MockProvider) GetRegions() cloudprovider.Regions {
	return nil
}

// CreateCluster mocks the creation of a cluster.
func (mp *MockProvider) CreateCluster(clusterName string) error {
	fmt.Printf("MockProvider: Cluster '%s' created.\n", clusterName)
	return nil
}

// DeleteCluster mocks the deletion of a cluster.
func (mp *MockProvider) DeleteCluster(clusterName string) error {
	fmt.Printf("MockProvider: Cluster '%s' deleted.\n", clusterName)
	return nil
}

// Other cloud provider functions can be mocked similarly...
