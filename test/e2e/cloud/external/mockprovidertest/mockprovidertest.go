package mockprovidertest

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider"
	. "k8s.io/kubernetes/staging/src/k8s.io/cloud-provider-mockprovider/mockprovider"
	"k8s.io/kubernetes/test/e2e/cloud"
)

type MockCloudProvider struct {
	MockCloud
}

func (m *MockCloudProvider) Name() string {
	return m.ProviderName()
}

func (m *MockCloudProvider) DeleteNodes() (func(nodes []string) error, bool) {
	return func(nodes []string) error {
		return nil
	}, true
}

func (m *MockCloudProvider) CreateInstance() (func(name string, config cloud.InstanceConfig) error, bool) {
	return func(name string, config cloud.InstanceConfig) error {
		return nil
	}, true
}

func (m *MockCloudProvider) LoadBalancer() (func(lbName string, action string) error, bool) {
	return nil, false
}

func (m *MockCloudProvider) UpgradeMaster() (func() error, bool) {
	return func() error {
		return nil
	}, true
}

func (m *MockCloudProvider) UpgradeCluster() (func() error, bool) {
	return func() error {
		return nil
	}, true
}

func (m *MockCloudProvider) DowngradeCluster() (func() error, bool) {
	return func() error {
		return nil
	}, false
}

func (m *MockCloudProvider) RebootNodes() (func(nodes []string) error, bool) {
	return func(nodes []string) error {
		return nil
	}, true
}

func (m *MockCloudProvider) VerifyPodCount() (func() error, bool) {
	return func() error {
		return nil
	}, true
}

func (m *MockCloudProvider) ServiceTests() (func() error, bool) {
	return func() error {
		return nil
	}, true
}

func (m *MockCloudProvider) InstanceMetadata() (func(node *v1.Node) (*cloudprovider.InstanceMetadata, error), bool) {
	return func(node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
		instanceMetadata, err := (&MockInstancesV2{}).InstanceMetadata(context.TODO(), node)
		if err != nil {
			return nil, err
		}
		return &cloudprovider.InstanceMetadata{
			ProviderID:    instanceMetadata.ProviderID,
			InstanceType:  instanceMetadata.InstanceType,
			NodeAddresses: instanceMetadata.NodeAddresses,
			Zone:          instanceMetadata.Zone,
			Region:        instanceMetadata.Region,
		}, nil
	}, true
}

var _ = Describe("Cloud Provider CCM Tests", func() {
	var provider *MockCloudProvider
	nodes := []string{"node1", "node2"}

	BeforeEach(func() {
		provider = &MockCloudProvider{}
	})

	It("should delete nodes successfully", func() {
		Expect(func() { cloud.TestDeleteNodes(provider, nodes) }).ShouldNot(Panic())
	})

	It("should upgrade master successfully", func() {
		Expect(func() { cloud.TestUpgradeMaster(provider) }).ShouldNot(Panic())
	})

	It("should upgrade cluster successfully", func() {
		Expect(func() { cloud.TestUpgradeCluster(provider) }).ShouldNot(Panic())
	})

	It("should downgrade cluster successfully", func() {
		Expect(func() { cloud.TestDowngradeCluster(provider) }).ShouldNot(Panic())
	})

	It("should reboot nodes successfully", func() {
		Expect(func() { cloud.TestRebootNodes(provider, nodes) }).ShouldNot(Panic())
	})

	It("should verify pod count successfully", func() {
		Expect(func() { cloud.TestVerifyPodCount(provider) }).ShouldNot(Panic())
	})

	It("should validate InstanceMetadata values", func() {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		instanceMetadataFunc, supported := provider.InstanceMetadata()
		Expect(supported).To(BeTrue())

		metadata, err := instanceMetadataFunc(node)
		Expect(err).NotTo(HaveOccurred())
		//Test fails here due to metadata comparision node being like mocker://...
		Expect(metadata.ProviderID).To(Equal("mock://test-node"))
		Expect(metadata.InstanceType).To(Equal("mock-instance-type"))
		Expect(metadata.Zone).To(Equal("mock-zone"))
		Expect(metadata.Region).To(Equal("mock-region"))
	})
})
