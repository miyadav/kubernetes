package mockprovider

import (
	"context"
	"fmt"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (mp *MockProvider) Initialize(ctx context.Context) error {
	fmt.Println("MockProvider initialized.")
	return nil
}

func (mp *MockProvider) DeleteNode(nodeName string) error {
	fmt.Printf("MockProvider: Node '%s' deleted.\n", nodeName)
	return nil
}

func (mp *MockProvider) CreateCluster(clusterName string) error {
	fmt.Printf("MockProvider: Cluster '%s' created.\n", clusterName)
	return nil
}

func (mp *MockProvider) DeleteCluster(clusterName string) error {
	fmt.Printf("MockProvider: Cluster '%s' deleted.\n", clusterName)
	return nil
}

func (mp *MockProvider) GetInstances() ([]string, error) {
	return []string{"instance1", "instance2"}, nil
}

func (mp *MockProvider) GetZones() ([]string, error) {
	return []string{"zone1", "zone2"}, nil
}

func (mp *MockProvider) GetRegions() ([]string, error) {
	return []string{"region1", "region2"}, nil
}
