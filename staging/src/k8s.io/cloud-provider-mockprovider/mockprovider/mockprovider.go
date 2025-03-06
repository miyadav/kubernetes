package mockprovider

import (
	"context"
	"errors"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/cloud-provider"
)

type MockCloud struct{}

func (m *MockCloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

func (m *MockCloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) { return nil, false }

//func (m *MockCloud) Instances() (cloudprovider.Instances, bool) { return &MockInstances{}, true }

func (m *MockCloud) InstancesV2() (cloudprovider.InstancesV2, bool) { return &MockInstancesV2{}, true }

func (m *MockCloud) Zones() (cloudprovider.Zones, bool) { return nil, false }

func (m *MockCloud) Clusters() (cloudprovider.Clusters, bool) { return nil, false }

func (m *MockCloud) Routes() (cloudprovider.Routes, bool) { return nil, false }

func (m *MockCloud) ProviderName() string { return "mock" }

func (m *MockCloud) HasClusterID() bool { return false }

// MockInstances simulates cloud provider instances
type MockInstances struct {
	mu          sync.Mutex
	shutdownMap map[string]bool
}

// NewMockInstances initializes MockInstances
func NewMockInstances() *MockInstances {
	return &MockInstances{
		shutdownMap: make(map[string]bool),
	}
}

// NodeAddresses returns a mock node's addresses
func (m *MockInstances) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	return []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		{Type: v1.NodeExternalIP, Address: "34.45.56.67"},
	}, nil
}

// NodeAddressesByProviderID returns mock node addresses by provider ID
func (m *MockInstances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	if !strings.HasPrefix(providerID, "mock://") {
		return nil, errors.New("invalid provider ID format")
	}
	return m.NodeAddresses(ctx, types.NodeName(strings.TrimPrefix(providerID, "mock://")))
}

// InstanceID returns the instance ID, assuming nodeName is the ID
func (m *MockInstances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return "mock://" + string(nodeName), nil
}

// InstanceType returns a static mock instance type
func (m *MockInstances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	return "mock-instance-type", nil
}

// InstanceTypeByProviderID returns instance type based on provider ID
func (m *MockInstances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	if !strings.HasPrefix(providerID, "mock://") {
		return "", errors.New("invalid provider ID format")
	}
	return "mock-instance-type", nil
}

// AddSSHKeyToAllInstances simulates adding an SSH key
func (m *MockInstances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	if user == "" || len(keyData) == 0 {
		return errors.New("invalid SSH key data")
	}
	return nil
}

// CurrentNodeName returns the current node's name
func (m *MockInstances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// InstanceExistsByProviderID checks if an instance exists
func (m *MockInstances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	if !strings.HasPrefix(providerID, "mock://") {
		return false, errors.New("invalid provider ID format")
	}
	return true, nil
}

// InstanceShutdownByProviderID simulates instance shutdown state tracking
func (m *MockInstances) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !strings.HasPrefix(providerID, "mock://") {
		return false, errors.New("invalid provider ID format")
	}

	return m.shutdownMap[providerID], nil
}

// SimulateShutdown allows setting a mock instance shutdown state
func (m *MockInstances) SimulateShutdown(providerID string, shutdown bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownMap[providerID] = shutdown
}

// MockInstancesV2 simulates cloud provider instances (v2 interface)
type MockInstancesV2 struct {
	mu          sync.Mutex
	shutdownMap map[string]bool
}

// NewMockInstancesV2 initializes MockInstancesV2
func NewMockInstancesV2() *MockInstancesV2 {
	return &MockInstancesV2{
		shutdownMap: make(map[string]bool),
	}
}

// InstanceExists checks if an instance exists in the mock provider
func (m *MockInstancesV2) InstanceExists(ctx context.Context, node *v1.Node) (bool, error) {
	if node == nil || node.Name == "" {
		return false, errors.New("invalid node")
	}
	return true, nil
}

// InstanceShutdown checks if the instance is in a shutdown state
func (m *MockInstancesV2) InstanceShutdown(ctx context.Context, node *v1.Node) (bool, error) {
	if node == nil || node.Name == "" {
		return false, errors.New("invalid node")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	providerID := "mocker://" + node.Name
	return m.shutdownMap[providerID], nil
}

// InstanceMetadata returns mock metadata for an instance
func (m *MockInstancesV2) InstanceMetadata(ctx context.Context, node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
	if node == nil || node.Name == "" {
		return nil, errors.New("invalid node")
	}

	return &cloudprovider.InstanceMetadata{
		ProviderID:   "mocker://" + node.Name,
		InstanceType: "mock-instance-type",
		NodeAddresses: []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.1.200"},
			{Type: v1.NodeExternalIP, Address: "45.67.89.123"},
		},
		Zone:   "mock-zone",
		Region: "mock-region",
	}, nil
}

// SimulateShutdown allows setting a mock instance shutdown state
func (m *MockInstancesV2) SimulateShutdown(node *v1.Node, shutdown bool) {
	if node == nil || node.Name == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	providerID := "mocker://" + node.Name
	m.shutdownMap[providerID] = shutdown
}

type MockInformerUser struct{}

func (m *MockInformerUser) SetInformers(informerFactory informers.SharedInformerFactory) {}

type MockClientBuilder struct{}

func (m *MockClientBuilder) Config(name string) (*rest.Config, error) {
	return nil, errors.New("not implemented")
}

func (m *MockClientBuilder) ConfigOrDie(name string) *rest.Config {
	panic("not implemented")
}

func (m *MockClientBuilder) Client(name string) (kubernetes.Interface, error) {
	return nil, errors.New("not implemented")
}

func (m *MockClientBuilder) ClientOrDie(name string) kubernetes.Interface {
	panic("not implemented")
}
