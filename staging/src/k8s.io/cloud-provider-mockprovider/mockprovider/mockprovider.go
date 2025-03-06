package mockprovider

import (
	"context"
	"errors"

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

type MockInstances struct{}

func (m *MockInstances) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	return nil, nil
}

func (m *MockInstances) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	return nil, nil
}

func (m *MockInstances) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return string(nodeName), nil
}

func (m *MockInstances) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	return "mock-instance-type", nil
}

func (m *MockInstances) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	return "mock-instance-type", nil
}

func (m *MockInstances) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return nil
}

func (m *MockInstances) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

func (m *MockInstances) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	return true, nil
}

func (m *MockInstances) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	return false, nil
}

type MockInstancesV2 struct{}

func (m *MockInstancesV2) InstanceExists(ctx context.Context, node *v1.Node) (bool, error) {
	return true, nil
}

func (m *MockInstancesV2) InstanceShutdown(ctx context.Context, node *v1.Node) (bool, error) {
	return false, nil
}

func (m *MockInstancesV2) InstanceMetadata(ctx context.Context, node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
	return &cloudprovider.InstanceMetadata{
		ProviderID:    "mocker://" + node.Name,
		InstanceType:  "mock-instance-type",
		NodeAddresses: []v1.NodeAddress{},
		Zone:          "mock-zone",
		Region:        "mock-region",
	}, nil
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
