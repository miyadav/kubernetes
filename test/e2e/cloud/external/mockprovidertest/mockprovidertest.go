package mockprovidertest

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/fake"
	"k8s.io/kubernetes/test/e2e/cloud"
)

type MockCloudProvider struct {
	fake.Cloud
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
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("instance name cannot be empty")
		}
		// Simulating providerID creation
		providerID := fmt.Sprintf("fake://%s", name)
		fmt.Printf("Instance %s created with providerID: %s\n", name, providerID)
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
		// Use the fake.Cloud's InstanceMetadata method
		metadata, err := m.Cloud.InstanceMetadata(context.TODO(), node)
		if err != nil {
			return nil, err
		}
		return metadata, nil
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
		Expect(metadata.ProviderID).To(Equal("fake://test-node"))
		Expect(metadata.InstanceType).To(Equal("fake-instance-type"))
		Expect(metadata.Zone).To(Equal("us-central1-b")) // default in fake.Cloud
		Expect(metadata.Region).To(Equal("us-central1")) // default in fake.Cloud
	})

	It("should create an instance and validate providerID format", func() {
		Expect(func() { cloud.TestCreateInstance(provider) }).ShouldNot(Panic())
	})
})
