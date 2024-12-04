package external

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
)


var _ = SIGDescribe(feature.CloudProvider, framework.WithDisruptive(),"CloudControllerManager End-to-End Tests", func() {
	    f := framework.NewDefaultFramework("cloudprovider")
    f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged
    var c clientset.Interface

    ginkgo.BeforeEach(func() {
        // Only supported in AWS/GCE/GKE because those are the cloud providers
        // where E2E tests are currently running.
        e2eskipper.SkipUnlessProviderIs("aws", "gce", "gke","mock")
        c = f.ClientSet
    })
	var provider CloudProviderInterface
	var capabilities []string

	// Load the cloud provider capabilities
	providerMap, err := loadCloudProviderCapabilities()
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed to load cloud provider capabilities: %v", err))
	}

	// Iterate over cloud providers and run the corresponding tests
	for providerName, providerCapabilities := range providerMap {
		ginkgo.Context(fmt.Sprintf("Provider: %s", providerName), func() {

			// Instantiate the provider based on the providerName
			switch providerName {
			case "mock":
				provider = mockprovider.NewMockProvider()
			// Add more cases for AWS, GCP, etc. as needed
			default:
				ginkgo.Fail(fmt.Sprintf("Unknown provider: %s", providerName))
				return
			}

			// Initialize provider before running tests
			ginkgo.BeforeEach(func() {
				var err error
				err = provider.Initialize(context.Background())
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				capabilities = providerCapabilities.Capabilities
			})

			// Test node deletion capability if supported
			if contains(capabilities, "delete_node") {
				ginkgo.It("should delete a node", func() {
					err := provider.DeleteNode("test-node")
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				})
			}

			// Test cluster creation capability if supported
			if contains(capabilities, "create_cluster") {
				ginkgo.It("should create a cluster", func() {
					err := provider.CreateCluster("test-cluster")
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				})
			}

			// Test cluster deletion capability if supported
			if contains(capabilities, "delete_cluster") {
				ginkgo.It("should delete a cluster", func() {
					err := provider.DeleteCluster("test-cluster")
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				})
			})

		})
	}
})

// Helper function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Load cloud provider capabilities from a YAML file
func loadCloudProviderCapabilities() (map[string]CloudProviderCapabilities, error) {
	var capabilities CloudProviderCapabilities
	file, err := ioutil.ReadFile("test/e2e/cloud/external/cloud_provider_capabilities.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read capabilities file: %v", err)
	}

	err = yaml.Unmarshal(file, &capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities file: %v", err)
	}

	providerMap := make(map[string]CloudProviderCapabilities)
	for _, provider := range capabilities.Providers {
		providerMap[provider.Name] = provider
	}

	return providerMap, nil
}
