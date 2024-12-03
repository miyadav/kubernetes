package mockprovider

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

// MockCloudProvider is a mock implementation of ProviderInterface.

type MockCloudProvider struct{}

// NewMockProvider returns a new instance of MockCloudProvider.
func NewMockProvider(config interface{}) *MockCloudProvider {
	return &MockCloudProvider{}
}

// CleanupServiceResources mocks the cleanup of cloud resources.
func (m *MockCloudProvider) CleanupServiceResources(ctx context.Context, client kubernetes.Interface, namespace, resourceName, resourceType string) error {
	// Add mock cleanup logic here, if needed.
	return nil
}
