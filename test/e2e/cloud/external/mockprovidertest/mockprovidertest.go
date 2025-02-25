package mockprovidertest

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
})
